package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/crper/tqrx/internal/core"
)

// View 实现 tea.Model 接口，并根据终端宽度渲染窄布局或宽布局。
func (m Model) View() tea.View {
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		m.renderBody(),
		m.renderFooter(),
	)

	view := tea.NewView(m.styles.app.Render(body))
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	view.WindowTitle = "tqrx"
	view.BackgroundColor = m.theme.appBg
	view.ForegroundColor = m.theme.text
	return view
}

func (m Model) renderHeader() string {
	left := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.brand.Render("TQRX"),
		"  ",
		m.styles.subtitle.Render("live qr workbench"),
	)

	right := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.headerChip.Render("["+strings.ToUpper(string(m.format))+"/"+string(m.level)+"]"),
		" ",
		m.styles.headerChip.Render("["+strings.ToUpper(string(m.themeMode))+"]"),
		" ",
		m.renderHeaderStatusBadge(m.previewStatus),
	)

	if m.isNarrow() {
		if lipgloss.Width(left)+2+lipgloss.Width(right) <= m.width-2 {
			return m.styles.header.Render(lipgloss.JoinHorizontal(lipgloss.Left, left, "  ", right))
		}
		return m.styles.header.Render(lipgloss.JoinVertical(lipgloss.Left, left, right))
	}

	leftWidth := max(0, m.width-lipgloss.Width(right)-4)
	return m.styles.header.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(leftWidth).Render(left),
			right,
		),
	)
}

func (m Model) renderBody() string {
	plan := m.planLayout()
	if m.isNarrow() {
		return m.renderNarrowBody(plan)
	}
	return m.renderWideBody(plan)
}

func (m Model) renderWideBody(plan layoutPlan) string {
	left := lipgloss.NewStyle().Width(plan.editWidth).Render(m.renderEditPanel(plan.editWidth))
	right := m.renderPreviewColumn(plan.previewWidth)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) renderNarrowBody(plan layoutPlan) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderEditPanel(plan.editWidth),
		"",
		m.renderPreviewColumn(plan.previewWidth),
	)
}

// editPanelParts 汇总编辑面板一次渲染所需的全部片段，避免 renderEditPanel
// 和 editPanelParts 之间重复计算焦点状态。
type editPanelParts struct {
	editFocused   bool
	title         string
	composeLabel  string
	textarea      string
	settingsLabel string
	formatRow     string
	sizeRow       string
	levelRow      string
	outputRow     string
	status        string
}

func (m Model) editPanelParts() editPanelParts {
	editFocused := panelHasFocus(m.focus, focusContent, focusFormat, focusSize, focusLevel, focusOutput)
	return editPanelParts{
		editFocused:   editFocused,
		title:         m.renderPanelTitle("Edit", editFocused),
		composeLabel:  m.renderSectionLabel("Compose", m.focus == focusContent),
		textarea:      m.renderTextareaSurface(),
		settingsLabel: m.renderSectionLabel("Settings", editFocused),
		formatRow:     m.renderSettingChipRow("Format", formatLabels, strings.ToLower(string(m.format)), m.focus == focusFormat),
		sizeRow:       m.renderSettingInputRow("Size", m.renderTextInputSurface(m.size), m.focus == focusSize),
		levelRow:      m.renderSettingChipRow("Level", levelLabels, string(m.level), m.focus == focusLevel),
		outputRow:     m.renderSettingInputRow("Output", m.renderTextInputSurface(m.output), m.focus == focusOutput),
		status:        m.renderInlineStatus("Status", m.saveStatus, false),
	}
}

func (m Model) renderEditPanel(width int) string {
	parts := m.editPanelParts()
	lines := []string{
		parts.title,
		parts.composeLabel,
		parts.textarea,
		"",
		parts.settingsLabel,
		parts.formatRow,
		parts.sizeRow,
		parts.levelRow,
		parts.outputRow,
	}
	if m.saveStatus.Message != "" {
		lines = append(lines, "", parts.status)
	}
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return m.panelStyle(parts.editFocused).Width(width).Render(body)
}

func (m Model) renderPreviewPanel(width int) string {
	metaRows := []string{m.renderPreviewMeta(width)}
	if shouldShowPreviewInlineStatus(m.previewStatus) {
		metaRows = append(metaRows, m.renderInlineStatus("State", m.previewStatus, false))
	}

	canvas := m.renderPreviewCanvas(width)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderPanelTitle("Preview", panelHasFocus(m.focus, focusPreview, focusSave)),
		strings.Join(metaRows, "\n"),
		canvas,
	)

	return m.panelStyle(panelHasFocus(m.focus, focusPreview, focusSave)).
		Width(width).
		Render(body)
}

func (m Model) renderPreviewColumn(width int) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderPreviewPanel(width),
		m.renderPreviewActions(width),
	)
}

func (m Model) renderPreviewCanvas(width int) string {
	innerWidth := max(0, width-panelFrameWidth)

	canvasStyle := m.styles.previewCanvas
	if m.focus == focusPreview {
		canvasStyle = m.styles.previewCanvasFocus
	}

	content := m.preview.View()
	paper := canvasStyle.Render(content)
	return lipgloss.PlaceHorizontal(innerWidth, lipgloss.Center, paper)
}

func (m Model) renderPreviewMeta(width int) string {
	innerWidth := max(0, width-panelFrameWidth)

	meta := m.previewMetaContent()
	infoParts := make([]string, 0, len(meta.infoParts))
	for i, part := range meta.infoParts {
		if i == 0 {
			infoParts = append(infoParts, m.styles.metaValue.Render(part))
			continue
		}
		infoParts = append(infoParts, m.styles.meta.Render(part))
	}
	info := strings.Join(infoParts, "  ")
	path := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.meta.Render("Path "),
		m.styles.path.MaxWidth(max(0, innerWidth-5)).Render(meta.path),
	)

	if lipgloss.Width(info)+3+lipgloss.Width(path) <= innerWidth {
		gap := max(2, innerWidth-lipgloss.Width(info)-lipgloss.Width(path))
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			info,
			strings.Repeat(" ", gap),
			path,
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		info,
		path,
	)
}

func (m Model) renderFooter() string {
	message := m.footerMessage()
	if message == "" {
		return ""
	}
	return m.styles.footer.Render(message)
}

func (m Model) previewDocument(width, height int) string {
	whitespace := lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(m.theme.canvasBg))

	if m.previewText == "" {
		line := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Background(m.theme.canvasBg)
		title := lipgloss.NewStyle().
			Inherit(line).
			Foreground(m.theme.emptyTitle).
			Bold(true)
		note := lipgloss.NewStyle().
			Inherit(line).
			Foreground(m.theme.emptyNote)
		blank := line.Render("")

		if m.previewStatus.Kind == statusWaiting && strings.TrimSpace(m.content.Value()) != "" {
			block := lipgloss.JoinVertical(
				lipgloss.Left,
				title.Render("Updating preview..."),
				blank,
				note.Render("Keep typing or wait a moment."),
			)
			return lipgloss.PlaceVertical(height, lipgloss.Center, block, whitespace)
		}

		block := lipgloss.JoinVertical(
			lipgloss.Left,
			title.Render("Paste text or a link to render a QR."),
			blank,
			note.Render("Example: https://example.com"),
		)
		return lipgloss.PlaceVertical(height, lipgloss.Center, block, whitespace)
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, m.previewText, whitespace)
}

func (m Model) focusedLabel(focused bool) lipgloss.Style {
	if focused {
		return m.styles.labelFocused
	}
	return m.styles.label
}

func (m Model) renderSectionLabel(title string, focused bool) string {
	return m.focusedLabel(focused).Render(title)
}

func (m Model) renderTextareaSurface() string {
	return renderSurface(m.textareaSurfaceSpec())
}

func (m Model) renderTextInputSurface(input textinput.Model) string {
	return renderSurface(m.textInputSurfaceSpec(input))
}

func (m Model) textareaSurfaceSpec() surfaceSpec {
	return surfaceSpec{
		content:    m.content.View(),
		width:      lipgloss.Width(m.content.Prompt) + m.content.Width(),
		height:     m.content.Height(),
		background: m.styles.field,
	}
}

func (m Model) textInputSurfaceSpec(input textinput.Model) surfaceSpec {
	raw := input.View()
	return surfaceSpec{
		content:    raw,
		width:      max(lipgloss.Width(raw), lipgloss.Width(input.Prompt)+input.Width()+1),
		height:     1,
		background: m.styles.field,
	}
}

func renderSurface(spec surfaceSpec) string {
	whitespace := lipgloss.WithWhitespaceStyle(spec.background)
	lines := strings.Split(trimSurfaceOverlay(spec.content), "\n")
	for i, line := range lines {
		lines[i] = lipgloss.PlaceHorizontal(spec.width, lipgloss.Left, line, whitespace)
	}
	block := strings.Join(lines, "\n")
	return lipgloss.PlaceVertical(spec.height, lipgloss.Top, block, whitespace)
}

// trimSurfaceOverlay 会裁掉 Bubble 组件为了光标和 overlay 留下的尾部空格，
// 让背景面板自己决定剩余空间怎么铺满。
func trimSurfaceOverlay(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}

func (m Model) footerMessage() string {
	if m.saveStatus.Message != "" {
		return statusText(m.saveStatus)
	}
	if m.previewStatus.Kind == statusError && m.previewStatus.Message != "" {
		return statusText(m.previewStatus)
	}
	helper := m.help
	helper.SetWidth(m.width)
	return helper.View(m.keys)
}

// panelStyle 只在边框强调上区分聚焦态，避免不同 panel 的视觉语义分叉。
func (m Model) panelStyle(focused bool) lipgloss.Style {
	if focused {
		return m.styles.panelFocused
	}
	return m.styles.panel
}

func (m Model) renderPanelTitle(title string, focused bool) string {
	style := m.styles.panelTitle
	if focused {
		style = m.styles.panelTitleFocused
	}
	return style.Render("[ " + title + " ]")
}

func (m Model) renderStatusBadge(status statusModel) string {
	return m.renderBadgeWithStyles(status, func(kind statusKind) lipgloss.Style {
		switch kind {
		case statusWaiting:
			return m.styles.statusWaiting
		case statusError:
			return m.styles.statusError
		case statusSuccess:
			return m.styles.statusSuccess
		default:
			return m.styles.statusReady
		}
	})
}

func (m Model) renderHeaderStatusBadge(status statusModel) string {
	return m.renderBadgeWithStyles(status, func(kind statusKind) lipgloss.Style {
		style := m.styles.headerChip
		switch kind {
		case statusWaiting:
			return style.Foreground(m.theme.warning).Bold(true)
		case statusError:
			return style.Foreground(m.theme.danger).Bold(true)
		case statusSuccess:
			return style.Foreground(m.theme.success).Bold(true)
		default:
			return style.Foreground(m.theme.muted)
		}
	})
}

// renderBadgeWithStyles 是 renderStatusBadge 和 renderHeaderStatusBadge 的
// 公共实现：通过 styleFor 闭包注入不同场景的样式查找逻辑，消除两处 badge
// 渲染中重复的文本拼接和 Render 调用。
func (m Model) renderBadgeWithStyles(status statusModel, styleFor func(statusKind) lipgloss.Style) string {
	text := statusText(status)
	return styleFor(status.Kind).Render("[" + text + "]")
}

func (m Model) renderSettingChipRow(label string, options []string, selected string, focused bool) string {
	chips := make([]string, 0, len(options))
	for _, option := range options {
		chips = append(chips, m.renderChoiceChip(option, strings.EqualFold(option, selected), focused))
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.focusedLabel(focused).Width(settingsLabelWidth).Render(label),
		strings.Join(chips, " "),
	)
}

func (m Model) renderSettingInputRow(label, field string, focused bool) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.focusedLabel(focused).Width(settingsLabelWidth).Render(label),
		field,
	)
}

func (m Model) renderInlineStatus(label string, status statusModel, focused bool) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.focusedLabel(focused).Render(label),
		m.renderStatusBadge(status),
	)
}

func (m Model) renderChoiceChip(option string, selected, focused bool) string {
	style := m.styles.chip
	if selected {
		style = m.styles.chipSelected
		if focused {
			style = m.styles.chipSelectedActive
		}
	}
	return style.Render("[" + strings.ToUpper(option) + "]")
}

func (m Model) renderSaveButton() string {
	if m.focus == focusSave {
		return m.styles.saveButtonFocused.Render("[Save QR]")
	}
	return m.styles.saveButton.Render("[Save QR]")
}

func (m Model) renderPreviewActions(width int) string {
	row := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.renderSaveButton(),
		"  ",
		m.styles.note.Render(m.previewFocusHint()),
	)

	return lipgloss.NewStyle().
		Width(max(0, width-previewActionIndent)).
		PaddingLeft(previewActionIndent).
		Render(row)
}

// previewFocusHint 用用户视角描述当前预览该怎么操作，而不是暴露内部状态名。
func (m Model) previewFocusHint() string {
	if m.previewCondensed() {
		return "native preview exceeds viewport; enlarge terminal"
	}
	if !m.previewScrollable() {
		return "auto-fit live preview"
	}
	if m.focus == focusPreview {
		return "wheel / arrows scroll"
	}
	return "click or tab to scroll"
}

func (m Model) previewScrollable() bool {
	return lipgloss.Height(m.previewText) > m.preview.Height() || lipgloss.Width(m.previewText) > m.preview.Width()
}

func (m Model) previewCondensed() bool {
	modules, ok := m.currentPreviewModules()
	if !ok {
		return false
	}
	capacity := min(m.preview.Width(), m.preview.Height()*2)
	if capacity <= 0 {
		return false
	}
	return modules > capacity
}

func (m Model) previewScanSummary() string {
	capacity := m.previewCapacityModules()
	if capacity <= 0 {
		return ""
	}

	current, ok := m.currentPreviewModules()
	if !ok {
		return ""
	}
	summary := fmt.Sprintf("mods %d/%d", current, capacity)
	if current <= capacity {
		return summary
	}
	level, ok := m.recommendedScanLevel(capacity)
	if !ok {
		return summary + " · enlarge terminal"
	}
	if level == m.level {
		return summary
	}
	return fmt.Sprintf("%s · suggest %s for scan", summary, level)
}

func (m Model) previewCapacityModules() int {
	return min(m.preview.Width(), m.preview.Height()*2)
}

func (m Model) currentPreviewModules() (int, bool) {
	switch {
	case m.prepared != nil:
		return m.prepared.PreviewModules(), true
	case len(m.levelModules) > 0:
		modules, ok := m.levelModules[m.level]
		return modules, ok
	default:
		return 0, false
	}
}

// recommendedScanLevel 会从高纠错往低纠错寻找“在当前终端还能完整显示”的最
// 高等级，这样建议结果更贴近真实扫码体验。
func (m Model) recommendedScanLevel(capacity int) (core.Level, bool) {
	if capacity <= 0 {
		return "", false
	}
	if len(m.levelModules) == 0 {
		return "", false
	}

	for _, level := range descendingLevelOrder {
		modules, ok := m.levelModules[level]
		if !ok {
			continue
		}
		if modules <= capacity {
			return level, true
		}
	}
	return "", false
}
