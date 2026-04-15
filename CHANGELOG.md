# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-04-16

### Added

- `-m` flag: print QR code directly in the terminal without generating a file
- `-m` combined with `-o`: print to terminal and save file in one command
- `prepareRequest` helper to share Normalize+Prepare logic between CLI paths
- Comprehensive test coverage for `-m` flag (7 new test cases)

### Changed

- Restore `bitmapModules` helper to keep module count formula as single source of truth
- Pre-compute `formatLabels` and `levelLabels` as package-level variables to avoid per-render allocations
- Extract `renderBadgeWithStyles` to unify status badge rendering in TUI
- Share `basePreviewCanvas` and `boldStyle` in `newUIStyles` to reduce style duplication
- Remove negative-index checks in `previewGridModule` (callers guarantee non-negative indices)
- Add doc comments to unexported helpers (`prepareRequest`, `renderToTerminal`, `editPanelParts`, `renderBadgeWithStyles`, `previewGridModule`)
- Replace custom `scripts/check-docs.sh` with `revive` linter (`exported` and `package-comments` rules) in `.golangci.yml`
- Remove `scripts/check-docs.sh` and all references (lefthook, CI, docs)
- Restructure README: `README.md` as English primary, add `README.zh-Hans.md` for Chinese
- Delete `README.en.md` (replaced by new `README.md`)
- Add `CHANGELOG.md` and `CHANGELOG.zh-Hans.md` (v0.1.0–v0.3.0)
- Remove `AGENT_LEARNINGS.md` (sandbox-only notes)

## [0.3.0] - 2026-04-06

### Added

- Lefthook-based local Git hooks (`pre-commit` auto-format, `pre-push` validation)
- GitHub issue templates and PR template

### Changed

- Drop stale previews during TUI update state instead of keeping old frames
- Simplify preview state ownership and related helpers
- Switch GitHub Actions to Node 24 LTS runtime
- Sync project docs with updated release and contribution flow

### Fixed

- Align TUI tests with updated state boundaries

## [0.2.0] - 2026-04-04

### Added

- Content length warnings: `content long` (> 500 chars) and `content very long` (> 1000 chars)
- `Ctrl+R` shortcut to reset all TUI settings to defaults
- Enhanced golangci-lint configuration

### Changed

- SVG rendering now pre-allocates buffer for better performance
- Share edit panel layout state between rendering and hit geometry
- Simplify shared QR and TUI helpers to reduce code duplication
- Centralize TUI option and status constants

## [0.1.0] - 2026-03-27

### Added

- Initial release of `tqrx`
- CLI: generate PNG/SVG QR codes from text or stdin
- `-m` flag: print QR code directly in terminal
- TUI: interactive workbench with live preview, format/size/level controls
- Shared render path: preview and export use the same bitmap source
- Scan-oriented preview hints (`mods X/Y`, `suggest M for scan`)
- Theme support: `AUTO` / `LIGHT` / `DARK` via `TQRX_THEME` env var
- GoReleaser CI/CD with Homebrew Cask target

[Unreleased]: https://github.com/crper/tqrx/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/crper/tqrx/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/crper/tqrx/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/crper/tqrx/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/crper/tqrx/releases/tag/v0.1.0
