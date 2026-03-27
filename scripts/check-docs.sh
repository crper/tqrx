#!/usr/bin/env bash
set -euo pipefail

# 这个脚本刻意把真正的 Go AST 检查逻辑内嵌在临时程序里：
# shell 只负责参数、包枚举和清理，语法判断交给 Go 标准库，避免用 grep
# 硬扫注释时出现大量误报。
if [[ "${1:-}" == "--help" ]]; then
  cat <<'EOF'
Usage: bash scripts/check-docs.sh [packages...]

Checks Go package comments and exported top-level declarations for missing
documentation comments.

Examples:
  bash scripts/check-docs.sh
  bash scripts/check-docs.sh ./internal/...
EOF
  exit 0
fi

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$ROOT"

# 默认检查整个模块；也允许调用方只传局部包模式来缩小范围。
patterns=("$@")
if [[ ${#patterns[@]} -eq 0 ]]; then
  patterns=(./...)
fi

# 临时 Go 程序负责：
# 1. 解析 package comment
# 2. 检查导出符号是否带 doc comment
# 3. 输出稳定、可读的诊断信息
tmp_go="$(mktemp "${TMPDIR:-/tmp}/check-docs-XXXX.go")"
trap 'rm -f "$tmp_go"' EXIT

cat >"$tmp_go" <<'EOF'
package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
)

type pkgInput struct {
	importPath string
	dir        string
	name       string
}

type issue struct {
	file string
	line int
	msg  string
}

func main() {
	inputs, err := readInputs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var issues []issue
	for _, input := range inputs {
		pkgIssues, err := checkPackage(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		issues = append(issues, pkgIssues...)
	}

	sort.Slice(issues, func(i, j int) bool {
		if issues[i].file != issues[j].file {
			return issues[i].file < issues[j].file
		}
		if issues[i].line != issues[j].line {
			return issues[i].line < issues[j].line
		}
		return issues[i].msg < issues[j].msg
	})

	if len(issues) == 0 {
		fmt.Println("OK: package comments and exported doc comments look complete.")
		return
	}

	for _, item := range issues {
		fmt.Printf("%s:%d: %s\n", item.file, item.line, item.msg)
	}
	os.Exit(1)
}

func readInputs() ([]pkgInput, error) {
	var inputs []pkgInput
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// shell 侧通过制表符把 import path / dir / package name 三元组喂进来。
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid go list input: %q", scanner.Text())
		}
		inputs = append(inputs, pkgInput{
			importPath: parts[0],
			dir:        parts[1],
			name:       parts[2],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return inputs, nil
}

func checkPackage(input pkgInput) ([]issue, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, input.dir, func(info os.FileInfo) bool {
		name := info.Name()
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse package %s: %w", input.importPath, err)
	}

	pkg := pkgs[input.name]
	if pkg == nil {
		for _, candidate := range pkgs {
			pkg = candidate
			break
		}
	}
	if pkg == nil {
		return nil, nil
	}

	var files []*ast.File
	for _, file := range pkg.Files {
		files = append(files, file)
	}
	sort.Slice(files, func(i, j int) bool {
		return fset.Position(files[i].Package).Filename < fset.Position(files[j].Package).Filename
	})

	var issues []issue
	packageDocs := 0
	packageDocFile := ""
	packageDocLine := 1
	for _, file := range files {
		filename := fset.Position(file.Package).Filename
		if file.Doc != nil && strings.TrimSpace(file.Doc.Text()) != "" {
			// Go 约定一个 package 只应该有一份 package comment；多了通常意味
			// 着注释漂移或误放在普通源码文件里。
			packageDocs++
			if packageDocFile == "" {
				packageDocFile = filename
				packageDocLine = fset.Position(file.Doc.Pos()).Line
			}
		}
	}
	switch {
	case packageDocs == 0:
		first := fset.Position(files[0].Package)
		issues = append(issues, issue{
			file: first.Filename,
			line: first.Line,
			msg:  "package comment is missing",
		})
	case packageDocs > 1:
		issues = append(issues, issue{
			file: packageDocFile,
			line: packageDocLine,
			msg:  "package has more than one package comment",
		})
	}

	for _, file := range files {
		filename := fset.Position(file.Package).Filename
		for _, decl := range file.Decls {
			switch node := decl.(type) {
			case *ast.FuncDecl:
				// 只有顶层导出函数或导出类型的方法才要求 doc comment。
				if ast.IsExported(node.Name.Name) && receiverNeedsDoc(node.Recv) && missingDoc(node.Doc) {
					issues = append(issues, issue{
						file: filename,
						line: fset.Position(node.Pos()).Line,
						msg:  fmt.Sprintf("exported function or method %s is missing a doc comment", node.Name.Name),
					})
				}
			case *ast.GenDecl:
				if node.Tok != token.TYPE && node.Tok != token.CONST && node.Tok != token.VAR {
					continue
				}
				for _, spec := range node.Specs {
					switch item := spec.(type) {
					case *ast.TypeSpec:
						if ast.IsExported(item.Name.Name) && missingDoc(item.Doc, node.Doc) {
							issues = append(issues, issue{
								file: filename,
								line: fset.Position(item.Pos()).Line,
								msg:  fmt.Sprintf("exported type %s is missing a doc comment", item.Name.Name),
							})
						}
					case *ast.ValueSpec:
						for _, name := range item.Names {
							if !ast.IsExported(name.Name) {
								continue
							}
							if missingDoc(item.Doc, node.Doc) {
								kind := strings.ToLower(node.Tok.String())
								issues = append(issues, issue{
									file: filename,
									line: fset.Position(name.Pos()).Line,
									msg:  fmt.Sprintf("exported %s %s is missing a doc comment", kind, name.Name),
								})
							}
						}
					}
				}
			}
		}
	}

	return issues, nil
}

func missingDoc(groups ...*ast.CommentGroup) bool {
	for _, group := range groups {
		if group != nil && strings.TrimSpace(group.Text()) != "" {
			return false
		}
	}
	return true
}

func receiverNeedsDoc(recv *ast.FieldList) bool {
	if recv == nil || len(recv.List) == 0 {
		return true
	}
	return exportedReceiverType(recv.List[0].Type)
}

func exportedReceiverType(expr ast.Expr) bool {
	switch node := expr.(type) {
	case *ast.Ident:
		return ast.IsExported(node.Name)
	case *ast.StarExpr:
		return exportedReceiverType(node.X)
	case *ast.IndexExpr:
		return exportedReceiverType(node.X)
	case *ast.IndexListExpr:
		return exportedReceiverType(node.X)
	case *ast.SelectorExpr:
		return ast.IsExported(node.Sel.Name)
	default:
		return false
	}
}
EOF

# go list 负责展开包模式；临时检查器只处理“某个具体目录里的某个包”。
go list -f '{{.ImportPath}}{{"\t"}}{{.Dir}}{{"\t"}}{{.Name}}' "${patterns[@]}" | go run "$tmp_go"
