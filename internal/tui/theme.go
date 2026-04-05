package tui

import (
	"image/color"
	"os"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

type uiThemeMode string

const (
	// auto 会跟随终端背景，另外两个值用于显式覆盖。
	uiThemeAuto  uiThemeMode = "auto"
	uiThemeLight uiThemeMode = "light"
	uiThemeDark  uiThemeMode = "dark"
)

// uiTheme 汇总一套完整的终端配色语义，避免视图层直接操作原始颜色值。
type uiTheme struct {
	name          string
	dark          bool
	accent        color.Color
	accentStrong  color.Color
	appBg         color.Color
	fieldBg       color.Color
	canvasBg      color.Color
	text          color.Color
	textSoft      color.Color
	muted         color.Color
	border        color.Color
	canvasBorder  color.Color
	qrInk         color.Color
	success       color.Color
	warning       color.Color
	danger        color.Color
	placeholder   color.Color
	promptFocused color.Color
	promptBlurred color.Color
	emptyTitle    color.Color
	emptyNote     color.Color
}

// resolveUIThemeMode 只接受少量稳定环境变量值，未知输入统一回退到 auto。
func resolveUIThemeMode() uiThemeMode {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("TQRX_THEME"))) {
	case "", "auto":
		return uiThemeAuto
	case "light":
		return uiThemeLight
	case "dark":
		return uiThemeDark
	default:
		return uiThemeAuto
	}
}

// resolveUITheme 把“用户显式选择”和“终端背景探测”合成最终主题。
func resolveUITheme(mode uiThemeMode, hasDarkBackground bool) uiTheme {
	if mode == uiThemeLight {
		hasDarkBackground = false
	}
	if mode == uiThemeDark {
		hasDarkBackground = true
	}

	if hasDarkBackground {
		return uiTheme{
			name:          "dark",
			dark:          true,
			accent:        lipgloss.Color("#7AA2F7"),
			accentStrong:  lipgloss.Color("#7DCFFF"),
			appBg:         lipgloss.Color("#1A1B26"),
			fieldBg:       lipgloss.Color("#24283B"),
			canvasBg:      lipgloss.Color("#FFFFFF"),
			text:          lipgloss.Color("#C0CAF5"),
			textSoft:      lipgloss.Color("#A9B1D6"),
			muted:         lipgloss.Color("#565F89"),
			border:        lipgloss.Color("#414868"),
			canvasBorder:  lipgloss.Color("#7A88B8"),
			qrInk:         lipgloss.Color("#000000"),
			success:       lipgloss.Color("#9ECE6A"),
			warning:       lipgloss.Color("#E0AF68"),
			danger:        lipgloss.Color("#DB4B4B"),
			placeholder:   lipgloss.Color("#565F89"),
			promptFocused: lipgloss.Color("#7DCFFF"),
			promptBlurred: lipgloss.Color("#565F89"),
			emptyTitle:    lipgloss.Color("#3B4261"),
			emptyNote:     lipgloss.Color("#66719A"),
		}
	}

	return uiTheme{
		name:          "light",
		dark:          false,
		accent:        lipgloss.Color("#2E7DE9"),
		accentStrong:  lipgloss.Color("#2E7DE9"),
		appBg:         lipgloss.Color("#E9EDF5"),
		fieldBg:       lipgloss.Color("#DCE4F2"),
		canvasBg:      lipgloss.Color("#FFFFFF"),
		text:          lipgloss.Color("#2E3440"),
		textSoft:      lipgloss.Color("#4C566A"),
		muted:         lipgloss.Color("#6B7280"),
		border:        lipgloss.Color("#9AA5BC"),
		canvasBorder:  lipgloss.Color("#B5BED2"),
		qrInk:         lipgloss.Color("#000000"),
		success:       lipgloss.Color("#587539"),
		warning:       lipgloss.Color("#8C5E10"),
		danger:        lipgloss.Color("#B73A3A"),
		placeholder:   lipgloss.Color("#7B8395"),
		promptFocused: lipgloss.Color("#2E7DE9"),
		promptBlurred: lipgloss.Color("#7B8395"),
		emptyTitle:    lipgloss.Color("#3B4252"),
		emptyNote:     lipgloss.Color("#6B7280"),
	}
}

// detectDarkBackground 借助 lipgloss 的终端能力探测来推断默认主题。
func detectDarkBackground() bool {
	return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
}

// newUIStyles 把主题颜色翻译成一组可复用的组件样式，视图层只消费语义字段。
func newUIStyles(theme uiTheme) uiStyles {
	accentStrong := theme.accentStrong
	text := theme.text
	muted := theme.muted
	border := theme.border
	success := theme.success
	warning := theme.warning
	danger := theme.danger
	canvasBorder := theme.canvasBorder

	basePanel := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(border).
		Padding(panelPaddingY, panelPaddingX)

	return uiStyles{
		app: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(theme.text),
		header: lipgloss.NewStyle().MarginBottom(1),
		brand: lipgloss.NewStyle().
			Bold(true).
			Foreground(accentStrong),
		subtitle: lipgloss.NewStyle().
			Foreground(muted),
		headerChip: lipgloss.NewStyle().
			Foreground(muted),
		panel:        basePanel,
		panelFocused: basePanel.BorderForeground(accentStrong),
		field: lipgloss.NewStyle().
			Background(theme.fieldBg),
		panelTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(text),
		panelTitleFocused: lipgloss.NewStyle().
			Bold(true).
			Foreground(accentStrong),
		label: lipgloss.NewStyle().
			Foreground(theme.textSoft),
		labelFocused: lipgloss.NewStyle().
			Foreground(accentStrong).
			Bold(true),
		muted: lipgloss.NewStyle().
			Foreground(muted),
		note: lipgloss.NewStyle().
			Foreground(theme.textSoft),
		path: lipgloss.NewStyle().
			Foreground(text),
		meta: lipgloss.NewStyle().
			Foreground(muted),
		metaValue: lipgloss.NewStyle().
			Foreground(text),
		chip: lipgloss.NewStyle().
			Foreground(muted),
		chipSelected: lipgloss.NewStyle().
			Foreground(text).
			Bold(true),
		chipSelectedActive: lipgloss.NewStyle().
			Foreground(accentStrong).
			Bold(true),
		saveButton: lipgloss.NewStyle().
			Foreground(text).
			Bold(true),
		saveButtonFocused: lipgloss.NewStyle().
			Foreground(accentStrong).
			Bold(true),
		previewCanvas: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(canvasBorder).
			Background(theme.canvasBg).
			Foreground(theme.qrInk).
			Padding(0, 1),
		previewCanvasFocus: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(accentStrong).
			Background(theme.canvasBg).
			Foreground(theme.qrInk).
			Padding(0, 1),
		statusReady: lipgloss.NewStyle().
			Foreground(muted),
		statusWaiting: lipgloss.NewStyle().
			Foreground(warning).
			Bold(true),
		statusError: lipgloss.NewStyle().
			Foreground(danger).
			Bold(true),
		statusSuccess: lipgloss.NewStyle().
			Foreground(success).
			Bold(true),
		footer: lipgloss.NewStyle().MarginTop(0).Foreground(theme.textSoft),
	}
}

// newHelpStyles 单独维护 help 组件样式，避免它和主界面样式意外耦合。
func newHelpStyles(theme uiTheme) help.Styles {
	styles := help.DefaultStyles(theme.dark)
	styles.Ellipsis = lipgloss.NewStyle().Foreground(theme.warning).Bold(true)
	styles.ShortKey = lipgloss.NewStyle().Foreground(theme.accentStrong).Bold(true)
	styles.ShortDesc = lipgloss.NewStyle().Foreground(theme.textSoft)
	styles.ShortSeparator = lipgloss.NewStyle().Foreground(theme.accentStrong).Bold(true)
	styles.FullKey = lipgloss.NewStyle().Foreground(theme.accentStrong).Bold(true)
	styles.FullDesc = lipgloss.NewStyle().Foreground(theme.textSoft)
	styles.FullSeparator = lipgloss.NewStyle().Foreground(theme.accentStrong).Bold(true)
	return styles
}

// applyComponentStyles 统一刷新 textarea 和 textinput 的主题相关样式。
func (m *Model) applyComponentStyles() {
	ta := textarea.DefaultStyles(m.theme.dark)
	ta.Focused.Base = lipgloss.NewStyle().Background(m.theme.fieldBg)
	ta.Blurred.Base = lipgloss.NewStyle().Background(m.theme.fieldBg)
	ta.Focused.Text = lipgloss.NewStyle().Foreground(m.theme.text)
	ta.Blurred.Text = lipgloss.NewStyle().Foreground(m.theme.textSoft)
	ta.Focused.LineNumber = lipgloss.NewStyle().Foreground(m.theme.muted)
	ta.Blurred.LineNumber = lipgloss.NewStyle().Foreground(m.theme.muted)
	ta.Focused.CursorLine = lipgloss.NewStyle().Foreground(m.theme.text).UnsetBackground()
	ta.Blurred.CursorLine = lipgloss.NewStyle().Foreground(m.theme.textSoft).UnsetBackground()
	ta.Focused.CursorLineNumber = lipgloss.NewStyle().Foreground(m.theme.muted).UnsetBackground()
	ta.Blurred.CursorLineNumber = lipgloss.NewStyle().Foreground(m.theme.muted).UnsetBackground()
	ta.Focused.EndOfBuffer = lipgloss.NewStyle().Foreground(m.theme.promptBlurred)
	ta.Blurred.EndOfBuffer = lipgloss.NewStyle().Foreground(m.theme.promptBlurred)
	ta.Focused.Placeholder = lipgloss.NewStyle().Foreground(m.theme.placeholder)
	ta.Blurred.Placeholder = lipgloss.NewStyle().Foreground(m.theme.placeholder)
	ta.Focused.Prompt = lipgloss.NewStyle().Foreground(m.theme.promptFocused)
	ta.Blurred.Prompt = lipgloss.NewStyle().Foreground(m.theme.promptBlurred)
	ta.Cursor = textarea.CursorStyle{
		Color: m.theme.accentStrong,
		Shape: tea.CursorBar,
		Blink: true,
	}
	m.content.SetStyles(ta)

	ti := textinput.DefaultStyles(m.theme.dark)
	ti.Focused.Text = lipgloss.NewStyle().Foreground(m.theme.text).Background(m.theme.fieldBg)
	ti.Blurred.Text = lipgloss.NewStyle().Foreground(m.theme.textSoft).Background(m.theme.fieldBg)
	ti.Focused.Placeholder = lipgloss.NewStyle().Foreground(m.theme.placeholder).Background(m.theme.fieldBg)
	ti.Blurred.Placeholder = lipgloss.NewStyle().Foreground(m.theme.placeholder).Background(m.theme.fieldBg)
	ti.Focused.Prompt = lipgloss.NewStyle().Foreground(m.theme.promptFocused).Background(m.theme.fieldBg)
	ti.Blurred.Prompt = lipgloss.NewStyle().Foreground(m.theme.promptBlurred).Background(m.theme.fieldBg)
	ti.Focused.Suggestion = lipgloss.NewStyle().Foreground(m.theme.placeholder).Background(m.theme.fieldBg)
	ti.Blurred.Suggestion = lipgloss.NewStyle().Foreground(m.theme.placeholder).Background(m.theme.fieldBg)
	m.size.SetStyles(ti)
	m.output.SetStyles(ti)
}

// applyTheme 负责一次性刷新 model 内所有和主题相关的派生状态。
func (m *Model) applyTheme(theme uiTheme) {
	m.theme = theme
	m.styles = newUIStyles(theme)
	m.help.Styles = newHelpStyles(theme)
	m.applyComponentStyles()
	m.syncPreviewViewport()
}

// cycleThemeMode 只负责切换模式；真正的样式重算交给 applyTheme 收口。
func (m *Model) cycleThemeMode() {
	m.themeMode = nextThemeMode(m.themeMode)
	m.applyTheme(resolveUITheme(m.themeMode, detectDarkBackground()))
}
