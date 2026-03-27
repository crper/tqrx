# Repository Guidelines

## Project Structure & Module Organization

This repository is a small Go CLI/TUI application for generating QR codes.

- `main.go`: process entrypoint and CLI wiring
- `internal/cli`: command parsing and top-level runner behavior
- `internal/core`: request normalization, validation, and shared types
- `internal/render`: QR matrix rendering, preview generation, and file export
- `internal/tui`: Bubble Tea workbench internals, split by responsibility across model/theme/update/view/layout/helpers
- `internal/**/_test.go`: package-level tests
- `internal/tui/testdata`: golden files for TUI snapshots
- `.air.toml`: local hot-reload configuration for `air`
- `scripts/check-docs.sh`: lightweight documentation consistency check
- `.github/workflows` and `.goreleaser.yml`: CI and release automation

## Build, Test, and Development Commands

- `go build -o tqrx .`: build the local binary
- `./tqrx "https://example.com"`: generate a default PNG
- `./tqrx tui`: launch the interactive terminal UI
- `go test ./...`: run the full test suite
- `go test -race ./...`: verify concurrency safety
- `go vet ./...`: catch common Go issues
- `bash scripts/check-docs.sh`: verify docs stay aligned with behavior

Run the core validation commands (`go test ./...`, `go vet ./...`, `go build ./...`, `bash scripts/check-docs.sh`) before opening a PR if you change behavior. Run `go test -race ./...` when you touch concurrency-sensitive code.

## Coding Style & Naming Conventions

Use `gofmt` formatting with standard Go tabs and import grouping. Keep implementations simple, readable, and production-friendly; avoid adding new layers or abstractions for small features. Prefer clear package-local names such as `Runner`, `Model`, `Prepared`, and `Normalize`. Keep behavior centralized: request rules belong in `internal/core`, rendering rules in `internal/render`, and TUI state in `internal/tui`.

## Testing Guidelines

Write table-driven tests when several inputs share one code path. Prefer user-visible behavior and stable public contracts over implementation details. Name tests as `TestXxxBehavior`, keep failures in `got ... want ...` form, and update golden files in `internal/tui/testdata` only for intentional UI changes.

## Commit & Pull Request Guidelines

This repository already has a small commit history with both plain imperative subjects and lightweight Conventional Commit prefixes. Use short imperative subjects such as `fix preview error state` or `docs(readme): sync tui behavior`. For pull requests, include:

- a brief summary of the change
- linked issue or rationale
- validation results (`go test`, `go vet`, `go build`, docs check)
- updated README or design docs when CLI/TUI behavior changes
