package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/crper/tqrx/internal/core"
)

type layoutPlan struct {
	headerHeight int
	editWidth    int
	previewWidth int
	editPanel    rect
	previewPanel rect
	rects        layoutRects
}

// previewMetaContent 是预览标题栏里“元信息 + 路径”的中间表示，方便先计算再
// 决定单行/双行排版。
type previewMetaContent struct {
	infoParts []string
	path      string
}

func (c previewMetaContent) infoText() string {
	return strings.Join(c.infoParts, "  ")
}

// wideLeftWidth 固定编辑区的大致可读宽度，避免宽屏下输入区被拉得过宽而难读。
func (m Model) wideLeftWidth() int {
	width := m.width * 3 / 10
	if width < 34 {
		return 34
	}
	if width > 42 {
		return 42
	}
	return width
}

// planLayout 是 TUI 布局的单一事实来源。
//
// 宽布局：
// [ Edit Panel ][ Preview Panel........ ]
//
// 窄布局：
// [ Edit Panel ]
// [ Preview... ]
// [ Save Row  ]
func (m Model) planLayout() layoutPlan {
	headerHeight := lipgloss.Height(m.renderHeader())
	left := appPaddingX
	top := headerHeight

	plan := layoutPlan{
		headerHeight: headerHeight,
	}

	if m.isNarrow() {
		width := max(36, m.width-4)
		plan.editWidth = width
		plan.previewWidth = width
		plan.editPanel = rect{
			x: left,
			y: top,
			w: width,
			h: lipgloss.Height(m.renderEditPanel(width)),
		}
		plan.previewPanel = rect{
			x: left,
			y: plan.editPanel.y + plan.editPanel.h + 1,
			w: width,
			h: lipgloss.Height(m.renderPreviewPanel(width)),
		}
		contentRect, controlsRect, rows := m.editRects(plan.editPanel)
		plan.rects = layoutRects{
			content:     contentRect,
			controls:    controlsRect,
			controlRows: rows,
			preview:     plan.previewPanel,
			saveButton:  m.previewSaveButtonRect(plan.previewPanel),
		}
		return plan
	}

	leftWidth := m.wideLeftWidth()
	previewWidth := max(38, m.width-leftWidth-3)
	plan.editWidth = leftWidth
	plan.previewWidth = previewWidth
	plan.editPanel = rect{
		x: left,
		y: top,
		w: leftWidth,
		h: lipgloss.Height(m.renderEditPanel(leftWidth)),
	}
	plan.previewPanel = rect{
		x: left + leftWidth,
		y: top,
		w: previewWidth,
		h: lipgloss.Height(m.renderPreviewPanel(previewWidth)),
	}

	contentRect, controlsRect, rows := m.editRects(plan.editPanel)
	plan.rects = layoutRects{
		content:     contentRect,
		controls:    controlsRect,
		controlRows: rows,
		preview:     plan.previewPanel,
		saveButton:  m.previewSaveButtonRect(plan.previewPanel),
	}
	return plan
}

func (m Model) layoutRects() layoutRects {
	return m.planLayout().rects
}

func (m Model) previewMetaLineCount(width int) int {
	innerWidth := max(0, width-panelFrameWidth)
	if innerWidth == 0 {
		return 2
	}

	meta := m.previewMetaContent()
	infoWidth := lipgloss.Width(meta.infoText())
	pathWidth := m.previewPathWidth(innerWidth, meta.path)
	if infoWidth+3+pathWidth <= innerWidth {
		return 1
	}
	return 2
}

func (m Model) previewMetaContent() previewMetaContent {
	parts := []string{
		fmt.Sprintf("%s • %s • %spx", strings.ToUpper(string(m.format)), m.level, m.size.Value()),
	}
	if m.previewProto != "" {
		parts = append(parts, "via "+m.previewProto)
	}
	if summary := m.previewScanSummary(); summary != "" {
		parts = append(parts, summary)
	}
	if warning := m.contentWarningMessage(); warning != "" {
		parts = append(parts, warning)
	}
	return previewMetaContent{
		infoParts: parts,
		path:      m.output.Value(),
	}
}

func (m Model) contentWarningMessage() string {
	switch m.contentWarning {
	case core.WarningCritical:
		return "content very long"
	case core.WarningHigh:
		return "content long"
	default:
		return ""
	}
}

func (m Model) previewPathWidth(innerWidth int, path string) int {
	if innerWidth <= 0 {
		return 0
	}

	const pathLabelWidth = 5
	valueWidth := min(lipgloss.Width(path), max(0, innerWidth-pathLabelWidth))
	return min(innerWidth, pathLabelWidth+valueWidth)
}

// editRects 会把编辑面板里“标题 / 内容 / 设置 / 状态”的视觉排版反推出鼠标
// 可命中的矩形区域，保证点击逻辑和最终渲染使用同一套几何结果。
func (m Model) editRects(panel rect) (rect, rect, controlRects) {
	innerX := panel.x + 1 + panelPaddingX
	innerY := panel.y + 1 + panelPaddingY
	innerWidth := max(0, panel.w-panelFrameWidth)
	parts := m.editPanelParts()

	y := innerY
	y += lipgloss.Height(parts.title)

	contentHeight := lipgloss.Height(parts.composeLabel) + lipgloss.Height(parts.textarea)
	contentRect := rect{
		x: innerX,
		y: y,
		w: innerWidth,
		h: contentHeight,
	}
	y += contentRect.h + 1

	settingsHeadingHeight := lipgloss.Height(parts.settingsLabel)
	y += settingsHeadingHeight

	formatRow := rect{
		x: innerX,
		y: y,
		w: innerWidth,
		h: lipgloss.Height(parts.formatRow),
	}
	y += formatRow.h

	sizeRow := rect{
		x: innerX,
		y: y,
		w: innerWidth,
		h: lipgloss.Height(parts.sizeRow),
	}
	y += sizeRow.h

	levelRow := rect{
		x: innerX,
		y: y,
		w: innerWidth,
		h: lipgloss.Height(parts.levelRow),
	}
	y += levelRow.h

	outputRow := rect{
		x: innerX,
		y: y,
		w: innerWidth,
		h: lipgloss.Height(parts.outputRow),
	}

	rows := controlRects{
		format:      formatRow,
		size:        sizeRow,
		level:       levelRow,
		output:      outputRow,
		formatChips: m.formatChipRects(formatRow),
		levelChips:  m.levelChipRects(levelRow),
	}

	controlsHeight := settingsHeadingHeight + formatRow.h + sizeRow.h + levelRow.h + outputRow.h
	if m.pathStatus.Message != "" {
		controlsHeight += 1 + lipgloss.Height(parts.status)
	}
	controlsRect := rect{
		x: innerX,
		y: contentRect.y + contentRect.h + 1,
		w: innerWidth,
		h: controlsHeight,
	}

	return contentRect, controlsRect, rows
}

func (m Model) previewSaveButtonRect(panel rect) rect {
	button := m.renderSaveButton()

	return rect{
		x: panel.x + previewActionIndent,
		y: panel.y + panel.h,
		w: lipgloss.Width(button),
		h: lipgloss.Height(button),
	}
}

func (m Model) formatChipRects(row rect) []formatChipRect {
	rects := make([]formatChipRect, 0, len(formatChoices))
	x := 0
	for _, format := range formatChoices {
		label := strings.ToLower(string(format))
		rendered := m.renderChoiceChip(label, m.format == format, m.focus == focusFormat)
		rects = append(rects, formatChipRect{
			rect:   m.rowChipRect(row, x, rendered),
			format: format,
		})
		x += lipgloss.Width(rendered) + 1
	}
	return rects
}

func (m Model) levelChipRects(row rect) []levelChipRect {
	rects := make([]levelChipRect, 0, len(levelOrder))
	x := 0
	for _, level := range levelOrder {
		rendered := m.renderChoiceChip(string(level), m.level == level, m.focus == focusLevel)
		rects = append(rects, levelChipRect{
			rect:  m.rowChipRect(row, x, rendered),
			level: level,
		})
		x += lipgloss.Width(rendered) + 1
	}
	return rects
}

func (m Model) rowChipRect(row rect, xOffset int, rendered string) rect {
	return rect{
		x: row.x + settingsLabelWidth + xOffset,
		y: row.y,
		w: lipgloss.Width(rendered),
		h: lipgloss.Height(rendered),
	}
}

// handleMouseClick 把鼠标位置映射为逻辑焦点或直接动作，保持鼠标和键盘操
// 作共享同一套状态流转。
func (m Model) handleMouseClick(x, y int) (Model, tea.Cmd) {
	rects := m.layoutRects()
	switch {
	case rects.saveButton.contains(x, y):
		m.focus = focusSave
		return m, m.applyFocus()
	case rects.content.contains(x, y):
		m.focus = focusContent
		return m, m.applyFocus()
	case rects.controls.contains(x, y):
		return m.handleControlClick(x, y, rects.controlRows)
	case rects.preview.contains(x, y):
		m.focus = focusPreview
		return m, m.applyFocus()
	default:
		return m, nil
	}
}

// handleControlClick 会优先处理命中 chip 的“直接切值”，否则仅移动焦点到对
// 应控件行。
func (m Model) handleControlClick(x, y int, rows controlRects) (Model, tea.Cmd) {
	for _, chip := range rows.formatChips {
		if !chip.rect.contains(x, y) {
			continue
		}
		m.focus = focusFormat
		focusCmd := m.applyFocus()
		if m.format == chip.format {
			return m, focusCmd
		}
		m.format = chip.format
		m.syncDerivedOutput()
		m.clearTransientStatuses()
		return m, tea.Batch(focusCmd, m.schedulePreview())
	}

	for _, chip := range rows.levelChips {
		if !chip.rect.contains(x, y) {
			continue
		}
		m.focus = focusLevel
		focusCmd := m.applyFocus()
		if m.level == chip.level {
			return m, focusCmd
		}
		m.level = chip.level
		m.clearTransientStatuses()
		return m, tea.Batch(focusCmd, m.schedulePreview())
	}

	switch {
	case rows.format.contains(x, y):
		m.focus = focusFormat
	case rows.size.contains(x, y):
		m.focus = focusSize
	case rows.level.contains(x, y):
		m.focus = focusLevel
	case rows.output.contains(x, y):
		m.focus = focusOutput
	default:
		m.focus = controlsFocus(m.focus)
	}
	return m, m.applyFocus()
}
