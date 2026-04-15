package tui

import (
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/crper/tqrx/internal/core"
	"github.com/crper/tqrx/internal/render"
)

const (
	// 这些布局常量集中描述终端工作台的“骨架”，这样改视觉密度时不用到处找
	// 魔法数字。
	narrowWidthThreshold = 96
	appPaddingX          = 1
	panelPaddingX        = 1
	panelPaddingY        = 0
	panelFrameWidth      = panelPaddingX*2 + 2
	panelFrameHeight     = panelPaddingY*2 + 2
	settingsLabelWidth   = 8
	previewActionIndent  = 2
)

// focusTarget 表示当前键盘焦点停留在哪个可交互区域。
type focusTarget int

const (
	focusContent focusTarget = iota
	focusFormat
	focusSize
	focusLevel
	focusOutput
	focusPreview
	focusSave
)

var (
	// formatChoices 列出 TUI 中可切换的导出格式，同时作为 formatLabels
	// 和 applyFormatCycle 的数据源。
	formatChoices = [...]core.Format{core.FormatPNG, core.FormatSVG}
)

// rect 是鼠标命中测试和布局计算共享的最小矩形单元。
type rect struct {
	x int
	y int
	w int
	h int
}

func (r rect) contains(x, y int) bool {
	return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
}

// layoutRects 汇总一次渲染周期里会参与命中测试的矩形区域。
type layoutRects struct {
	content     rect
	controls    rect
	controlRows controlRects
	preview     rect
	saveButton  rect
}

// controlRects 记录设置区每一行的位置，以及可点击 chip 的精确命中范围。
type controlRects struct {
	format      rect
	size        rect
	level       rect
	output      rect
	formatChips []formatChipRect
	levelChips  []levelChipRect
}

// formatChipRect 把格式值和对应的可点击区域绑在一起，方便鼠标事件直接命中。
type formatChipRect struct {
	rect   rect
	format core.Format
}

// levelChipRect 让纠错等级的视觉 chip 和实际业务值保持同源。
type levelChipRect struct {
	rect  rect
	level core.Level
}

// statusModel 描述预览区、路径状态和页脚共用的轻量状态文案。
type statusKind string

const (
	statusReady   statusKind = "ready"
	statusWaiting statusKind = "waiting"
	statusError   statusKind = "error"
	statusSuccess statusKind = "success"
)

type statusModel struct {
	Kind    statusKind
	Symbol  string
	Message string
}

// previewTickMsg 和 previewReadyMsg 组成一个带防抖的异步预览刷新管线：
//
//	输入变化 -> previewTickMsg -> preparePreviewCmd -> previewReadyMsg -> applyPreview
//
// 这样既能避免每个按键都重算二维码，也能丢弃过期结果。
type previewTickMsg struct {
	id int
}

type previewReadyMsg struct {
	id       int
	prepared *render.Prepared
	err      error
}

// keyMap 集中管理工作台可见的快捷键以及 help 区展示文案。
type keyMap struct {
	NextFocus   key.Binding
	PrevFocus   key.Binding
	Cycle       key.Binding
	Save        key.Binding
	Reset       key.Binding
	ToggleTheme key.Binding
	Quit        key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		NextFocus: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		PrevFocus: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev"),
		),
		Cycle: key.NewBinding(
			key.WithKeys("left", "right", "enter", " "),
			key.WithHelp("←/→", "cycle"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Reset: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "reset"),
		),
		ToggleTheme: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "theme"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextFocus, k.PrevFocus, k.Cycle, k.Save, k.Reset, k.ToggleTheme, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.NextFocus, k.PrevFocus, k.Cycle, k.Save, k.Reset, k.ToggleTheme, k.Quit}}
}

// uiStyles 把不同视觉语义拆成明确字段，避免渲染层直接拼样式细节。
type uiStyles struct {
	app                lipgloss.Style
	header             lipgloss.Style
	brand              lipgloss.Style
	subtitle           lipgloss.Style
	headerChip         lipgloss.Style
	panel              lipgloss.Style
	panelFocused       lipgloss.Style
	field              lipgloss.Style
	panelTitle         lipgloss.Style
	panelTitleFocused  lipgloss.Style
	label              lipgloss.Style
	labelFocused       lipgloss.Style
	muted              lipgloss.Style
	note               lipgloss.Style
	path               lipgloss.Style
	meta               lipgloss.Style
	metaValue          lipgloss.Style
	chip               lipgloss.Style
	chipSelected       lipgloss.Style
	chipSelectedActive lipgloss.Style
	saveButton         lipgloss.Style
	saveButtonFocused  lipgloss.Style
	previewCanvas      lipgloss.Style
	previewCanvasFocus lipgloss.Style
	statusReady        lipgloss.Style
	statusWaiting      lipgloss.Style
	statusError        lipgloss.Style
	statusSuccess      lipgloss.Style
	footer             lipgloss.Style
}

// surfaceSpec 描述一个“带背景的表面”该如何安放内容。textarea、textinput
// 和空白预览都走同一套表面渲染逻辑。
type surfaceSpec struct {
	content    string
	width      int
	height     int
	background lipgloss.Style
}

// Model 是交互式二维码工作台对应的 Bubble Tea 状态容器。
type Model struct {
	engine *render.Engine
	styles uiStyles
	theme  uiTheme
	keys   keyMap
	help   help.Model

	content textarea.Model
	size    textinput.Model
	output  textinput.Model
	preview viewport.Model

	format core.Format
	level  core.Level

	focus  focusTarget
	width  int
	height int

	outputDerived bool
	themeMode     uiThemeMode

	prepared            *render.Prepared
	previewText         string
	levelModules        map[core.Level]int
	levelModulesContent string

	previewStatus statusModel
	saveStatus    statusModel

	pendingPreviewID int
	debounce         time.Duration
	previewMaxWidth  int
	previewMaxHeight int
}

// NewModel 构建交互式二维码工作台的默认 Bubble Tea model。
func NewModel(engine *render.Engine) Model {
	themeMode := resolveUIThemeMode()
	theme := resolveUITheme(themeMode, detectDarkBackground())
	styles := newUIStyles(theme)
	keys := defaultKeyMap()

	content := textarea.New()
	content.Placeholder = "Type text or paste a link."
	content.Prompt = "│ "
	content.EndOfBufferCharacter = ' '
	content.ShowLineNumbers = false
	content.SetWidth(32)
	content.SetHeight(8)

	size := textinput.New()
	size.Placeholder = "256"
	size.SetValue("256")
	size.SetWidth(20)

	output := textinput.New()
	output.Placeholder = core.DefaultOutputPath(core.FormatPNG)
	output.SetValue(core.DefaultOutputPath(core.FormatPNG))
	output.SetWidth(32)

	preview := viewport.New(
		viewport.WithWidth(40),
		viewport.WithHeight(16),
	)
	preview.MouseWheelEnabled = true
	preview.SoftWrap = false
	preview.SetHorizontalStep(4)

	helper := help.New()
	helper.ShortSeparator = " • "
	helper.Styles = newHelpStyles(theme)
	helper.SetWidth(120)

	model := Model{
		engine:        engine,
		styles:        styles,
		theme:         theme,
		keys:          keys,
		help:          helper,
		content:       content,
		size:          size,
		output:        output,
		preview:       preview,
		format:        core.FormatPNG,
		level:         core.LevelMedium,
		focus:         focusContent,
		width:         120,
		height:        40,
		outputDerived: true,
		themeMode:     themeMode,
		previewStatus: statusModel{Kind: statusReady, Message: "Ready"},
		debounce:      120 * time.Millisecond,
	}

	model.applyComponentStyles()
	model.resize(model.width, model.height)
	model.applyFocus()
	return model
}

// Run 启动交互式终端界面，并阻塞直到退出。
func Run() error {
	program := tea.NewProgram(NewModel(render.NewEngine()))
	_, err := program.Run()
	return err
}
