# Repository Guidelines

## Project Structure & Module Organization

This repository is a small Go CLI/TUI application for generating and previewing QR codes. Module path: `github.com/crper/tqrx`. Requires Go 1.26.1+.

- `main.go`: process entrypoint, delegates all behavior to `cli.Runner`
- `internal/cli`: command parsing and top-level runner behavior (includes `-m` terminal preview mode)
- `internal/core`: request normalization (`Normalize`), validation (`UserError`, `ErrorKind`), shared types (`Request`, `NormalizedRequest`, `Format`, `Level`, `Source`, `ContentWarning`), and content-length warnings
- `internal/render`: QR matrix rendering (`Engine` with single-item cache), half-block terminal preview (`Prepared.Preview`, `PreviewFit`), and PNG/SVG file export (`Prepared.WriteToPath`); preview and export share one bitmap source
- `internal/tui`: Bubble Tea v2 workbench internals, split by responsibility:
  - `model.go`: state container, types, key bindings, constructor
  - `theme.go`: theme resolution (auto/light/dark), color/style definitions
  - `update.go`: message handling, async preview pipeline with debounce
  - `view.go`: rendering logic for all panels
  - `layout.go`: layout measurement and hit-testing rectangles
  - `helpers.go`: shared small utilities
- `internal/**/doc.go`: package-level documentation comments
- `internal/**/_test.go`: package-level tests
- `internal/render/engine_benchmark_test.go`, `internal/tui/update_benchmark_test.go`: benchmark tests
- `internal/tui/testdata`: golden files for TUI snapshots
- `.air.toml`: local hot-reload configuration for `air`
- `.golangci.yml`: linter configuration (errcheck, govet, staticcheck, ineffassign, unused, bodyclose, misspell, revive with `exported` and `package-comments` rules; formatters: gofmt, goimports)
- `.github/workflows` and `.goreleaser.yml`: CI and release automation

## Build, Test, and Development Commands

- `go build -o tqrx .`: build the local binary
- `./tqrx "https://example.com"`: generate a default PNG
- `./tqrx -m "https://example.com"`: print QR code to terminal
- `./tqrx -m "hello" -o qr.png`: print to terminal and save file
- `./tqrx tui`: launch the interactive terminal UI
- `go test ./...`: run the full test suite
- `go test -race ./...`: verify concurrency safety
- `go vet ./...`: catch common Go issues
- `golangci-lint run`: run configured linters (errcheck, govet, staticcheck, revive, etc.)
- `lefthook install`: install local Git hooks for format-on-commit and validation-on-push

Run the core validation commands (`go test ./...`, `go vet ./...`, `go build ./...`) before opening a PR if you change behavior. Run `go test -race ./...` when you touch concurrency-sensitive code.

## Coding Style & Naming Conventions

Use `gofmt` / `goimports` formatting with standard Go tabs and import grouping (enforced by `.golangci.yml`). Keep implementations simple, readable, and production-friendly; avoid adding new layers or abstractions for small features. Prefer clear package-local names such as `Runner`, `Model`, `Prepared`, and `Normalize`. Keep behavior centralized: request rules belong in `internal/core`, rendering rules in `internal/render`, and TUI state in `internal/tui`. Every package must have a `doc.go` with a proper package comment; all exported symbols must have doc comments (enforced by `revive` linter rules `exported` and `package-comments`).

## Testing Guidelines

Write table-driven tests when several inputs share one code path. Prefer user-visible behavior and stable public contracts over implementation details. Name tests as `TestXxxBehavior`, keep failures in `got ... want ...` form, and update golden files in `internal/tui/testdata` only for intentional UI changes. Benchmark tests live alongside the code they measure (e.g., `engine_benchmark_test.go`, `update_benchmark_test.go`).

## Commit & Pull Request Guidelines

This repository already has a small commit history with both plain imperative subjects and lightweight Conventional Commit prefixes. Use short imperative subjects such as `fix preview error state` or `docs(readme): sync tui behavior`. For pull requests, include:

- a brief summary of the change
- linked issue or rationale
- validation results (`go test`, `go vet`, `go build`)
- updated README or design docs when CLI/TUI behavior changes


## Docs to Keep in Sync

- **`CHANGELOG.md`** — entry under `[Unreleased]` for every user-visible change
- **`CHANGELOG.zh-Hans.md`** — always update Chinese changelog when updating English version
- **`AGENTS.md`** — update if build commands, project structure, or conventions change
- **`README.md`** — update if features, install steps, or usage change
- **`README.zh-Hans.md`** — always update Chinese README when updating English version
