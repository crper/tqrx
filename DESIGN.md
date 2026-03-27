# DESIGN

## Implementation Stack

- `tqrx` 当前的交互式界面基于 `Charm v2` 技术栈实现：
  - `charm.land/bubbletea/v2`
  - `charm.land/bubbles/v2`
  - `charm.land/lipgloss/v2`
- 二维码渲染由 `internal/render` 统一准备共享位图
- TUI 预览统一使用按画布自动适配的块字符矩阵视图
- PNG / SVG 导出与 TUI 预览共用同一份二维码位图
- 预览区域使用 `viewport`
- 输入区域使用 `textarea` 与 `textinput`
- 键位提示使用 `help` / `key`
- 主题模式支持 `auto / light / dark`
  - `Ctrl+T` 在三个模式间循环
  - `TQRX_THEME` 可设定默认模式
- 设计目标不是“参数表单”，而是“安静、清晰、可持续使用的终端工作台”

## Product Shape

`tqrx` is a developer-first QR tool with two modes:

- CLI for the fastest possible "give me a QR now" workflow
- TUI for a focused editing workbench with live preview and export

The TUI should feel like a compact terminal workstation, not a parameter form and not a dashboard.

## Core Principles

1. Preview is the visual hero.
2. Content input is the interaction starting point.
3. Controls serve the workflow; they are not the interface's main event.
4. Calm beats flashy.
5. Trust beats delight theater.

## Layout Rules

### Wide terminal

- Use a two-pane layout.
- Left pane: content input and secondary controls.
- Right pane: large live preview and a lightweight metadata strip.
- Save remains visually attached to the preview column, but should sit outside the preview frame.
- Top bar stays light and secondary.
- Footer holds keyboard help and short status echoes.
- The preview region may be focusable for scrolling, but it must still read as a visual stage, not as a dense settings area.

### Narrow terminal

- Collapse to a single-column workflow.
- Order stays the same:
  1. Content input
  2. Secondary controls
  3. Output path
  4. Preview
  5. Save actions
  6. Footer help/status

## Hierarchy

```text
FIRST LOOK
==========
1. Branded top strip with current format / level / theme / state
2. Large QR preview stage
3. Focused content input
4. Secondary controls
5. Footer help
```

- The preview should have the most spatial emphasis.
- The content field gets the default cursor focus.
- The metadata strip should be visible but quiet.
- Preview metadata and output path should try to share one rail before wrapping into two lines.

## Surface and Border Rules

- Use strong outer framing for the overall screen and major regions.
- Use weak inner boundaries for controls, and at most one nested frame inside the preview panel.
- Avoid nested heavy boxes and ornamental card treatment.
- Create hierarchy through spacing, labels, and alignment before adding lines.

## Current ASCII Direction

```text
TQRX  live qr workbench                      [PNG/M] [AUTO] [Ready]

┌──────────────────────────────┐┌────────────────────────────────────────────┐
│ [ Edit ]                     ││ [ Preview ]                                │
│ Compose                      ││ PNG • M • 256px            Path ./qrcode.png│
│ │                            ││  ┌──────────────────────────────────────┐  │
│ │ https://example.com        ││  │                                      │  │
│ │                            ││  │              QR preview              │  │
│ │                            ││  │                                      │  │
│ Settings                     ││  └──────────────────────────────────────┘  │
│ Format  [PNG] [SVG]          └────────────────────────────────────────────┘
│ Size    > 256                  [Save QR]  auto-fit live preview
│ Level   [L] [M] [Q] [H]
│ Output  > ./qrcode.png
└──────────────────────────────┘
```

## Color Rules

- Use a calm neutral base with one vivid accent.
- Accent color should communicate active focus, current selection, and primary affordance.
- Error and success meaning must not rely on color alone.
- Colors reinforce meaning; they do not carry meaning by themselves.
- Dark mode may be more saturated than light mode, but the palette should stay coherent.

## Copy Tone

Status and helper copy should be:

- Short
- Clear
- Warm
- Unshowy

Good examples:

- `Type text or paste a link.`
- `Updating`
- `Saved to ./qrcode.png`
- `Can't write to this path.`

Avoid:

- Robotic system-log phrasing
- Cute brand voice
- Long instructional paragraphs

## Interaction Rules

- In the content field, `Enter` always inserts a newline.
- Saving is an explicit action from the Save control.
- `Ctrl+S` is allowed as a global save shortcut.
- `Ctrl+T` cycles theme mode.
- Focus order follows workflow:
  - content
  - format / size / level
  - output path
  - preview viewport
  - save actions
  - back to content
- Debounced preview updates should surface in the fixed top status strip so the preview stage does not jump while typing.

## State Design

### Empty

- Show a warm directional prompt.
- Include one small example.
- Never leave the preview area blank without explanation.

### Waiting

- Show a subtle `Updating` cue in the top status strip.
- Do not replace the whole preview with a loud spinner.
- Do not insert temporary rows that change preview panel height.

### Error

- Place the main error near the affected region.
- Repeat a short version in the footer if helpful.
- Always use text, with optional color/symbol reinforcement.

### Success

- Confirm both completion and destination.
- Anchor save success near the output path.
- Footer can repeat a short confirmation.

## Preview Rules

- Default preview should be a clean terminal matrix view derived from the shared QR bitmap.
- Do not default to a novelty ASCII-first presentation if it hurts clarity.
- Surface the active render path in quiet metadata; right now that is `via Matrix`.
- Show live preview density as `mods current/capacity` so scan risk is explicit.
- Keep the preview stage upscaled inside the panel, but never do lossy downscaling that changes QR modules.
- If the terminal grid is too small for the native matrix, expose that honestly with a `native preview exceeds viewport; enlarge terminal` hint.
- Preview contrast should stay pure black on white to preserve scanner robustness under different terminal themes.
- For high-density codes near viewport limits, prefer integer-module scaling over slight non-integer stretching.
- When over capacity, provide a lower error-correction recommendation for preview scanning (suggestion only, no silent level mutation).
- The preview is a strong main actor, but it should stay restrained:
  - no decorative drama
  - no gratuitous animation
  - no oversized chrome
- Scrolling is acceptable when native modules exceed the current viewport; it is an inspection fallback, not a replacement for a fully scannable frame.
- Empty-state preview copy should be centered, not pinned to the top-left corner.

## Accessibility Rules

- Status meaning must be understandable without color.
- Focus states must remain visible in low-contrast terminal themes.
- Keyboard navigation must fully support the TUI.
- Responsive degradation must preserve meaning and task order.

## Not in Scope

- Multi-page navigation
- Split ASCII + final preview by default
- Celebratory animation language
