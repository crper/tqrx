# 更新日志

本项目的所有重要变更都会记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/) 规范。

## [Unreleased]

## [0.4.0] - 2026-04-16

### 新增

- `-m` 标志：在终端直接打印二维码，无需生成文件
- `-m` 搭配 `-o`：终端打印的同时保存文件
- `prepareRequest` 辅助函数，在 CLI 路径间共享 Normalize+Prepare 逻辑
- `-m` 标志完整测试覆盖（7 个新增测试用例）

### 变更

- 恢复 `bitmapModules` 辅助函数，保持模块计数公式为唯一真相来源
- 将 `formatLabels` 和 `levelLabels` 预计算为包级变量，避免每次渲染分配临时切片
- 提取 `renderBadgeWithStyles` 统一 TUI 状态徽章渲染
- 在 `newUIStyles` 中共享 `basePreviewCanvas` 和 `boldStyle`，减少样式重复
- 移除 `previewGridModule` 的负索引检查（调用方保证非负索引）
- 为未导出的辅助函数补充文档注释（`prepareRequest`、`renderToTerminal`、`editPanelParts`、`renderBadgeWithStyles`、`previewGridModule`）
- 用 `revive` linter（`exported` 和 `package-comments` 规则）替代自定义 `scripts/check-docs.sh`
- 移除 `scripts/check-docs.sh` 及所有引用（lefthook、CI、文档）
- 重构 README：`README.md` 作为英文主文档，新增 `README.zh-Hans.md` 中文版
- 删除 `README.en.md`（由新 `README.md` 替代）
- 新增 `CHANGELOG.md` 和 `CHANGELOG.zh-Hans.md`（v0.1.0–v0.3.0）
- 移除 `AGENT_LEARNINGS.md`（沙箱专用笔记）

## [0.3.0] - 2026-04-06

### 新增

- 基于 Lefthook 的本地 Git 钩子（`pre-commit` 自动格式化，`pre-push` 验证）
- GitHub Issue 模板和 PR 模板

### 变更

- TUI 更新状态期间丢弃旧预览，而非保留旧画面
- 简化预览状态所有权及相关辅助函数
- GitHub Actions 切换到 Node 24 LTS 运行时
- 同步项目文档与更新后的发布和贡献流程

### 修复

- 对齐 TUI 测试与更新后的状态边界

## [0.2.0] - 2026-04-04

### 新增

- 内容长度警告：`content long`（> 500 字符）和 `content very long`（> 1000 字符）
- `Ctrl+R` 快捷键：重置所有 TUI 设置为默认值
- 增强 golangci-lint 配置

### 变更

- SVG 渲染预分配缓冲区，提升性能
- 编辑面板布局状态在渲染和命中几何之间共享
- 简化共享 QR 和 TUI 辅助函数，减少代码重复
- 集中管理 TUI 选项和状态常量

## [0.1.0] - 2026-03-27

### 新增

- `tqrx` 初始发布
- CLI：从文本或 stdin 生成 PNG/SVG 二维码
- `-m` 标志：在终端直接打印二维码
- TUI：交互式工作台，支持实时预览、格式/尺寸/纠错等级控制
- 共享渲染路径：预览和导出使用同一份位图源
- 面向扫码的预览提示（`mods X/Y`、`suggest M for scan`）
- 主题支持：通过 `TQRX_THEME` 环境变量切换 `AUTO` / `LIGHT` / `DARK`
- GoReleaser CI/CD，支持 Homebrew Cask 分发

[Unreleased]: https://github.com/crper/tqrx/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/crper/tqrx/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/crper/tqrx/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/crper/tqrx/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/crper/tqrx/releases/tag/v0.1.0
