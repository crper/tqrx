# Contributing

参与贡献说明。当前只维护中文版本。

## 适合贡献什么

欢迎这些方向的贡献：

- Bug 修复
- CLI / TUI 交互优化
- README / CONTRIBUTING / DESIGN 等文档改进
- 测试补充
- 发布流程与开发体验优化

## 开始之前

建议先看这些文件：

- [README.md](./README.md)
- [README.en.md](./README.en.md)
- [DESIGN.md](./DESIGN.md)

如果你要改 TUI，先以 `DESIGN.md` 为准，不要一边写一边重新发明交互规则。

## 本地开发

本地构建：

```bash
go build -o tqrx .
```

TUI 开发热重载：

```bash
go install github.com/air-verse/air@latest
air
```

说明：

- 在仓库根目录启动时，`air` 会默认读取 [`.air.toml`](./.air.toml)
- 保存 `.go` / `go.mod` / `go.sum` 后，会自动重新构建并重启 `tqrx tui`
- 如命令不存在，先确认 `$(go env GOPATH)/bin` 在 `PATH` 里
- 当前 shell 还没配好 `PATH` 时，可以先直接运行 `$(go env GOPATH)/bin/air`
- 只有在不从仓库根目录启动，或要切换到别的配置文件时，才需要显式写 `-c`

常用验证：

```bash
go test ./...
go vet ./...
go build ./...
bash scripts/check-docs.sh
GOPROXY=https://goproxy.cn,direct go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --timeout=5m --disable-all --enable=errcheck,govet,staticcheck,ineffassign,unused
GOPROXY=https://goproxy.cn,direct go run github.com/goreleaser/goreleaser/v2@latest check
GOPROXY=https://goproxy.cn,direct go run github.com/goreleaser/goreleaser/v2@latest build --snapshot --clean
```

如果你在验证或维护 release 流程，记得：

- Homebrew Cask 目标 tap 为 `crper/homebrew-tap`
- 首次正式发布前，需要先创建该 tap 仓库
- 本仓库需要配置 `TAP_GITHUB_TOKEN`，供 GoReleaser 推送 cask 文件

## 提交前建议

提交前至少确认：

1. `go test ./...` 通过
2. `go vet ./...` 通过
3. `go build ./...` 通过
4. `bash scripts/check-docs.sh` 通过
5. 如果你改了 CLI / TUI 行为，`README.md`、`README.en.md`、`DESIGN.md` 一起更新

## 文档约定

- 中文主文档：
  - `README.md`
  - `CONTRIBUTING.md`
  - `DESIGN.md`
- 英文入口文档：
  - `README.en.md`
- `AGENTS.md` 是仓库本地的代理协作说明；只有当仓库事实、流程或约束变化时才更新

## 代码风格

- 优先简单、可维护、易读的实现
- 不为小功能做过度抽象
- 明确行为优先于“聪明写法”
- TUI 的状态、布局和提示文案尽量复用已有模型，不要复制分叉
- 当前 TUI 技术栈以 `Charm v2` 为准：
  - `charm.land/bubbletea/v2`
  - `charm.land/bubbles/v2`
  - `charm.land/lipgloss/v2`
- `internal/tui` 当前按 `model / theme / update / view / layout / helpers` 分文件维护
- 不要重新引入旧的 Charm v1 import path
- 当前终端预览基于共享二维码位图生成 `Matrix` 块字符视图；如改动预览规则，记得同步检查导出行为和 golden
- CLI 约定里 `tui` 是保留子命令名；需要编码这个字面量时，请用 `tqrx -- tui`
- 如果同时要传导出相关 flag，请把 flag 放在 `--` 前面，例如 `tqrx -f svg -o out.svg -- tui`

## Commit 与 PR

- 现有提交历史同时出现了简短祈使句和轻量 Conventional Commit 前缀，两种都可以
- 提交标题保持简短、明确，例如 `fix preview waiting layout shift` 或 `docs(readme): sync tui behavior`
- Pull Request 里请附上：
  - 变更摘要
  - 背景或原因
  - 验证结果
  - 如果改了 TUI，附上截图或终端输出

## License

提交到本仓库的代码默认按 [MIT License](./LICENSE) 发布。
