package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"golang.org/x/term"

	"github.com/crper/tqrx/internal/core"
	"github.com/crper/tqrx/internal/render"
	"github.com/crper/tqrx/internal/tui"
)

// Runner 负责顶层命令分发，让 main.go 保持为纯装配代码。
type Runner struct {
	// LaunchTUI 通过注入的方式提供，便于测试交互式路径。
	LaunchTUI func() error
}

// rootCLI 描述直接生成二维码时暴露给用户的 flags。
type rootCLI struct {
	Content string `arg:"" optional:"" help:"Text content to encode."`
	Output  string `short:"o" help:"Output path for the generated file."`
	Format  string `short:"f" help:"Output format (png or svg)."`
	Size    string `short:"s" help:"Output size (for example 256 or 256x256)."`
	Level   string `short:"l" help:"Error correction level (L, M, Q, H)."`
}

// tuiCLI 只承担 `tqrx tui` 这个保留子命令的匹配职责。
type tuiCLI struct{}

// NewRunner 创建同时支持 CLI 和 TUI 工作流的默认入口。
func NewRunner() *Runner {
	return &Runner{
		LaunchTUI: tui.Run,
	}
}

// Run 应用产品定义的输入优先级：
// 显式参数优先，其次是管道 stdin，最后才进入空输入校验。
func (r *Runner) Run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	parsed, command, err := parseArgs(args, stdout, stderr)
	if err != nil {
		return err
	}
	if command == "tui" {
		return r.LaunchTUI()
	}

	content := parsed.Content
	source := core.SourceCLIArg
	if strings.TrimSpace(content) == "" {
		content, err = readInput(stdin)
		if err != nil {
			return err
		}
		source = core.SourceStdin
	}

	req, err := core.Normalize(core.Request{
		Content:    content,
		Format:     parsed.Format,
		Size:       parsed.Size,
		OutputPath: parsed.Output,
		Level:      parsed.Level,
		Source:     source,
	})
	if err != nil {
		return err
	}

	engine := render.NewEngine()
	prepared, err := engine.Prepare(req)
	if err != nil {
		return err
	}
	if err := prepared.WriteToPath(req.OutputPath); err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "Saved to %s\n", req.OutputPath)
	return err
}

// parseArgs 先识别保留的 tui 入口，再分别交给对应的 Kong parser。
// 这样既保留 `tqrx tui`，又允许根命令直接接收位置参数内容。
func parseArgs(args []string, stdout, stderr io.Writer) (rootCLI, string, error) {
	if len(args) > 0 && args[0] == "tui" {
		return parseTUIArgs(args[1:], stdout, stderr)
	}
	return parseGenerateArgs(args, stdout, stderr)
}

func parseGenerateArgs(args []string, stdout, stderr io.Writer) (rootCLI, string, error) {
	var cli rootCLI
	parser, err := kong.New(
		&cli,
		kong.Name("tqrx"),
		kong.Description("Generate QR codes from text or stdin.\n\nSpecial command:\n  tui  Open the interactive terminal UI."),
		kong.UsageOnError(),
		kong.Writers(stdout, stderr),
	)
	if err != nil {
		return rootCLI{}, "", err
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		return rootCLI{}, "", err
	}
	return cli, ctx.Command(), nil
}

// parseTUIArgs 维持一个极小的子命令 parser，这样错误提示和 `--help`
// 输出仍由 Kong 统一生成。
func parseTUIArgs(args []string, stdout, stderr io.Writer) (rootCLI, string, error) {
	var cli tuiCLI
	parser, err := kong.New(
		&cli,
		kong.Name("tqrx tui"),
		kong.Description("Open the interactive terminal UI."),
		kong.UsageOnError(),
		kong.Writers(stdout, stderr),
	)
	if err != nil {
		return rootCLI{}, "", err
	}

	if _, err := parser.Parse(args); err != nil {
		return rootCLI{}, "", err
	}
	return rootCLI{}, "tui", nil
}

// readInput 把交互式终端视为“没有 stdin”，这样手动调用时会和 TUI
// 走同一条空内容校验路径。
func readInput(stdin io.Reader) (string, error) {
	if file, ok := stdin.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		return "", nil
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
