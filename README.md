# tqrx

> 终端优先的二维码生成器，兼顾一条命令出图和可交互预览。

[![CI](https://github.com/crper/tqrx/actions/workflows/checks.yml/badge.svg?branch=main)](https://github.com/crper/tqrx/actions/workflows/checks.yml)
![Go](https://img.shields.io/badge/Go-1.26.1-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-black.svg)

[English](./README.en.md) · [Design](./DESIGN.md) · [Contributing](./CONTRIBUTING.md)

`tqrx` 是一个小而直接的 Go 工具：

- `CLI` 适合快速生成 `PNG` / `SVG`
- `TUI` 适合边改内容边看预览再保存
- 预览和导出共用同一份二维码渲染结果，减少“界面看着对，导出不一致”的情况

```text
TQRX  live qr workbench                      [PNG/M] [AUTO] [Ready]

┌──────────────────────────────┐┌────────────────────────────────────────────┐
│ [ Edit ]                     ││ [ Preview ]                                │
│ Compose                      ││ PNG • M • 256px  mods 37/64               │
│ │ https://example.com        ││ Path ./qrcode.png                          │
│ │                            ││                                            │
│ Settings                     ││                 QR PREVIEW                 │
│ Format  [PNG] [SVG]          ││                                            │
│ Size    > 256                │└────────────────────────────────────────────┘
│ Level   [L] [M] [Q] [H]      │  [Save QR]
│ Output  > ./qrcode.png       │
└──────────────────────────────┘
```

如果想先看一遍交互流程，可以直接看下面的演示视频：

https://github.com/user-attachments/assets/286c8d1c-1db2-4a64-a2c5-1a5c01895825

## Why

- 默认命令就能出图：`tqrx "https://example.com"`
- 保留交互工作台：`tqrx tui`
- 明确的输出规则：格式、后缀、尺寸、纠错等级都有统一校验
- 扫码导向的预览提示：`mods X/Y` 和 `suggest M for scan`

## Install

当前最稳的方式是本地构建：

```bash
go build -o tqrx .
```

首个 release 发布后，可走 Homebrew Cask：

```bash
brew install --cask crper/tap/tqrx
```

## Quick Start

```bash
# 默认导出 PNG
./tqrx "https://example.com"

# 从 stdin 读取
printf 'from-pipe\n' | ./tqrx

# 导出 SVG
./tqrx "hello svg" -f svg -s 256 -o hello.svg -l H

# 打开交互式工作台
./tqrx tui
```

命令帮助：

```bash
./tqrx --help
./tqrx tui --help
```

说明：

- 直接把位置参数当作待编码内容
- `tui` 是保留子命令名；要编码字面量 `tui` 时请用 `./tqrx -- tui`
- 如果还要传 `-f` / `-o` / `-s` / `-l`，请把这些 flag 放在 `--` 前面，例如 `./tqrx -f svg -o out.svg -- tui`
- 默认输出为 `./qrcode.png`

## TUI

常用按键：

- `Tab` / `Shift+Tab` 切换焦点
- `Ctrl+S` 保存
- `Ctrl+T` 切换 `AUTO / LIGHT / DARK`
- `Enter` 在内容区输入换行，在 `Save` 上执行保存
- 支持鼠标点击切焦点，也支持终端粘贴事件

预览规则：

- 预览固定使用高对比黑白画布
- `mods X/Y` 表示当前二维码模块尺寸与预览容量
- 终端过小会提示 `native preview exceeds viewport; enlarge terminal`
- 纠错等级过高导致预览过密时，会给出 `suggest M for scan` 这类建议

环境变量：

```bash
TQRX_THEME=auto|light|dark
```

## Development

```bash
go test ./...
go vet ./...
go build ./...
bash scripts/check-docs.sh
```

TUI 开发可配合 `air`：

```bash
go install github.com/air-verse/air@latest
air
```

## Release

- 推送 `v*` tag 会触发 [`.github/workflows/release.yml`](./.github/workflows/release.yml)
- 产物由 [`.goreleaser.yml`](./.goreleaser.yml) 构建
- 当前目标平台为 `darwin/linux/windows` + `amd64/arm64`
- Homebrew Cask 目标仓库为 `crper/homebrew-tap`

## Docs

- [README.en.md](./README.en.md)
- [DESIGN.md](./DESIGN.md)
- [CONTRIBUTING.md](./CONTRIBUTING.md)

## License

[MIT](./LICENSE) · Copyright (c) 2026 crper
