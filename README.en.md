# tqrx

> A terminal-first QR generator for fast file output and live interactive preview.

[![CI](https://github.com/crper/tqrx/actions/workflows/checks.yml/badge.svg?branch=main)](https://github.com/crper/tqrx/actions/workflows/checks.yml)
![Go](https://img.shields.io/badge/Go-1.26.1-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-black.svg)

[中文说明](./README.md) · [Design](./DESIGN.md) · [Contributing](./CONTRIBUTING.md)

`tqrx` is a small Go tool with two clear workflows:

- `CLI` for generating `PNG` / `SVG` in one command
- `TUI` for editing content, checking preview, and saving in one session
- one shared render path for preview and export, so behavior stays consistent

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

If you want a quick look at the interactive flow, watch the demo below:

https://github.com/user-attachments/assets/286c8d1c-1db2-4a64-a2c5-1a5c01895825

## Why

- fast default path: `tqrx "https://example.com"`
- interactive path when you want to tune before export: `tqrx tui`
- stable validation around format, extension, size, and correction level
- scan-oriented preview hints such as `mods X/Y` and `suggest M for scan`

## Install

Today, local build is the most reliable path:

```bash
go build -o tqrx .
```

After the first tagged release, Homebrew Cask is ready for:

```bash
brew install --cask crper/tap/tqrx
```

## Quick Start

```bash
# default PNG export
./tqrx "https://example.com"

# read from stdin
printf 'from-pipe\n' | ./tqrx

# export SVG
./tqrx "hello svg" -f svg -s 256 -o hello.svg -l H

# open the interactive workbench
./tqrx tui
```

Help:

```bash
./tqrx --help
./tqrx tui --help
```

Notes:

- pass the text to encode as the root positional argument
- `tui` is a reserved subcommand name; to encode the literal text, use `./tqrx -- tui`
- if you also need `-f` / `-o` / `-s` / `-l`, place those flags before `--`, for example `./tqrx -f svg -o out.svg -- tui`
- the default output path is `./qrcode.png`

## TUI

Common controls:

- `Tab` / `Shift+Tab` move focus
- `Ctrl+S` saves
- `Ctrl+R` resets all settings to defaults
- `Ctrl+T` cycles `AUTO / LIGHT / DARK`
- `Enter` inserts a newline in content and saves when `Save` is focused
- mouse focus switching and terminal paste events are supported

Preview behavior:

- the canvas stays high-contrast black on white
- `mods X/Y` shows current module count versus preview capacity
- small terminals show `native preview exceeds viewport; enlarge terminal`
- dense previews can suggest a lower correction level, such as `suggest M for scan`
- long content (> 500 chars) shows `content long` warning, > 1000 chars shows `content very long`

Environment variable:

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

For TUI iteration with hot reload:

```bash
go install github.com/air-verse/air@latest
air
```

## Release

- pushing a `v*` tag triggers [`.github/workflows/release.yml`](./.github/workflows/release.yml)
- artifacts are built by [`.goreleaser.yml`](./.goreleaser.yml)
- current targets are `darwin/linux/windows` + `amd64/arm64`
- the Homebrew Cask target repository is `crper/homebrew-tap`

## Docs

- [README.md](./README.md)
- [DESIGN.md](./DESIGN.md)
- [CONTRIBUTING.md](./CONTRIBUTING.md)

## License

[MIT](./LICENSE) · Copyright (c) 2026 crper
