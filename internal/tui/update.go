package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"

	"github.com/crper/tqrx/internal/core"
	"github.com/crper/tqrx/internal/render"
)

// Init 实现 tea.Model 接口。
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		func() tea.Msg {
			return tea.RequestBackgroundColor()
		},
	)
}

// Update 实现 tea.Model 接口，并处理键盘输入、预览刷新和保存动作。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		if m.themeMode != uiThemeAuto || msg.IsDark() == m.theme.dark {
			return m, nil
		}
		m.applyTheme(resolveUITheme(m.themeMode, msg.IsDark()))
		return m, nil

	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil

	case previewTickMsg:
		if msg.id != m.pendingPreviewID {
			return m, nil
		}
		if strings.TrimSpace(m.content.Value()) == "" {
			m.applyPreview(preparedPreview{status: statusModel{Kind: statusReady, Message: "Ready"}})
			return m, nil
		}
		return m, m.preparePreviewCmd(msg.id)

	case previewReadyMsg:
		if msg.id != m.pendingPreviewID {
			return m, nil
		}
		if msg.err != nil {
			status := statusModel{Kind: statusError, Symbol: "!", Message: humanizeError(msg.err)}
			m.applyPreview(preparedPreview{status: status, footerStatus: status})
			return m, nil
		}
		m.applyPreview(preparedPreview{
			prepared:            msg.prepared,
			levelModules:        msg.levelModules,
			levelModulesContent: msg.levelModulesContent,
			status:              statusModel{Kind: statusReady, Message: "Live"},
			clearErrorFooter:    true,
		})
		return m, nil

	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			return m, nil
		}
		return m.handleMouseClick(msg.X, msg.Y)

	case tea.MouseWheelMsg:
		if m.layoutRects().preview.contains(msg.X, msg.Y) {
			m.focus = focusPreview
			focusCmd := m.applyFocus()
			var scrollCmd tea.Cmd
			m.preview, scrollCmd = m.preview.Update(msg)
			return m, tea.Batch(focusCmd, scrollCmd)
		}
		if m.focus == focusPreview {
			var cmd tea.Cmd
			m.preview, cmd = m.preview.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.PasteMsg:
		if next, cmd, handled := m.updateEditableFocus(msg); handled {
			return next, cmd
		}
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case matchesKey(msg, m.keys.Quit):
			return m, tea.Quit
		case matchesKey(msg, m.keys.Save):
			return m.saveCurrent()
		case matchesKey(msg, m.keys.Reset):
			return m.resetToDefaults()
		case matchesKey(msg, m.keys.ToggleTheme):
			m.cycleThemeMode()
			return m, nil
		case matchesKey(msg, m.keys.NextFocus):
			m.focus = nextFocus(m.focus)
			return m, m.applyFocus()
		case matchesKey(msg, m.keys.PrevFocus):
			m.focus = prevFocus(m.focus)
			return m, m.applyFocus()
		}

		if next, cmd, handled := m.updateEditableFocus(msg); handled {
			return next, cmd
		}

		switch m.focus {
		case focusFormat:
			if applyFormatCycle(msg, &m.format) {
				m.syncDerivedOutput()
				return m.afterValueChange(true, false, nil)
			}
			return m, nil

		case focusLevel:
			if applyLevelCycle(msg, &m.level) {
				return m.afterValueChange(true, false, nil)
			}
			return m, nil

		case focusPreview:
			var cmd tea.Cmd
			m.preview, cmd = m.preview.Update(msg)
			return m, cmd

		case focusSave:
			if msg.Code == tea.KeyEnter {
				return m.saveCurrent()
			}
		}
	}

	return m, nil
}

func (m *Model) resize(width, height int) {
	if width > 0 {
		m.width = width
	}
	if height > 0 {
		m.height = height
	}

	m.help.SetWidth(m.width)
	if m.isNarrow() {
		m.resizeNarrow()
	} else {
		m.resizeWide()
	}
	m.syncPreviewViewport()
}

func (m *Model) resizeNarrow() {
	panelWidth := max(36, m.width-4)
	contentInner := max(24, panelWidth-panelFrameWidth)
	m.content.SetWidth(contentInner)
	m.content.SetHeight(clamp(m.height/8, 4, 6))
	m.size.SetWidth(contentInner)
	m.output.SetWidth(contentInner)

	m.previewMaxWidth = max(30, panelWidth-panelFrameWidth-1)
	m.previewMaxHeight = clamp(m.height*3/5-4, 14, 22)
}

func (m *Model) resizeWide() {
	leftWidth := m.wideLeftWidth()
	leftInner := max(24, leftWidth-panelFrameWidth)
	m.content.SetWidth(leftInner)
	m.content.SetHeight(clamp(m.height/6, 5, 7))
	m.size.SetWidth(leftInner)
	m.output.SetWidth(leftInner)

	rightWidth := max(38, m.width-leftWidth-3)
	m.previewMaxWidth = max(38, rightWidth-panelFrameWidth-2)
	m.previewMaxHeight = max(16, m.height-7)
}

func (m Model) currentRequest() (core.NormalizedRequest, error) {
	return core.Normalize(core.Request{
		Content:    m.content.Value(),
		Format:     string(m.format),
		Size:       m.size.Value(),
		OutputPath: m.output.Value(),
		Level:      string(m.level),
		Source:     core.SourceTUI,
	})
}

// prepareCurrent 把“读取当前控件值”和“准备渲染产物”绑成同一步，保证保存和
// 预览永远基于同一份标准化请求。
func (m Model) prepareCurrent() (core.NormalizedRequest, *render.Prepared, error) {
	req, err := m.currentRequest()
	if err != nil {
		return core.NormalizedRequest{}, nil, err
	}

	prepared, err := m.engine.Prepare(req)
	if err != nil {
		return req, nil, err
	}
	return req, prepared, nil
}

// schedulePreview 通过自增 id 给每次预览刷新打标签。旧 tick 即使晚到，也会
// 因为 id 不匹配而被直接丢弃。
//
//	输入变化
//	    |
//	    v
//	schedulePreview --debounce--> previewTickMsg
//	                                    |
//	                                    v
//	                            preparePreviewCmd
//	                                    |
//	                                    v
//	                             previewReadyMsg
//	                                    |
//	                                    v
//	                               applyPreview
func (m *Model) schedulePreview() tea.Cmd {
	m.pendingPreviewID++
	m.previewStatus = statusModel{Kind: statusWaiting, Symbol: "…", Message: "Updating"}
	id := m.pendingPreviewID
	return tea.Tick(m.debounce, func(time.Time) tea.Msg {
		return previewTickMsg{id: id}
	})
}

// preparePreviewCmd 在命令执行时重新读取 model 快照，避免把昂贵准备逻辑放
// 在 Update 同步路径里阻塞输入。
func (m Model) preparePreviewCmd(id int) tea.Cmd {
	return func() tea.Msg {
		req, prepared, err := m.prepareCurrent()
		if err != nil {
			return previewReadyMsg{id: id, prepared: prepared, err: err}
		}
		return previewReadyMsg{
			id:                  id,
			prepared:            prepared,
			levelModules:        m.levelModulesForContent(req.Content),
			levelModulesContent: req.Content,
		}
	}
}

// saveCurrent 复用当前已准备好的二维码产物，避免保存和预览走出两条行为不
// 一致的路径。
func (m Model) saveCurrent() (Model, tea.Cmd) {
	req, prepared, err := m.prepareCurrent()
	if err != nil {
		m.setPathError(err)
		return m, nil
	}
	if err := prepared.WriteToPath(req.OutputPath); err != nil {
		m.setPathError(err)
		return m, nil
	}

	message := fmt.Sprintf("Saved to %s", req.OutputPath)
	m.pathStatus = statusModel{Kind: statusSuccess, Symbol: "✓", Message: message}
	m.footerStatus = statusModel{Kind: statusSuccess, Symbol: "✓", Message: message}
	m.applyPreview(preparedPreview{
		prepared: prepared,
		status:   statusModel{Kind: statusSuccess, Symbol: "✓", Message: "Synced"},
	})
	return m, nil
}

func (m Model) isNarrow() bool {
	return m.width > 0 && m.width < narrowWidthThreshold
}

func (m *Model) syncDerivedOutput() {
	if !m.outputDerived {
		return
	}
	m.output.SetValue(core.DefaultOutputPath(m.format))
}

type preparedPreview struct {
	prepared            *render.Prepared
	levelModules        map[core.Level]int
	levelModulesContent string
	status              statusModel
	footerStatus        statusModel
	clearErrorFooter    bool
}

// applyPreview 负责把后台准备好的预览结果原子地灌回 model，并维护那些和
// 预览结果绑定的衍生状态。
func (m *Model) applyPreview(next preparedPreview) {
	m.prepared = next.prepared
	if next.prepared == nil {
		m.levelModules = nil
		m.levelModulesContent = ""
		m.contentWarning = core.WarningNone
	} else if len(next.levelModules) > 0 {
		m.levelModules = cloneLevelModules(next.levelModules)
		m.levelModulesContent = next.levelModulesContent
		m.contentWarning = core.CheckContentLength(next.levelModulesContent)
	}
	m.previewStatus = next.status
	if next.footerStatus.Message != "" {
		m.footerStatus = next.footerStatus
	} else if next.clearErrorFooter && m.footerStatus.Kind == statusError {
		m.footerStatus = statusModel{}
	}
	m.syncPreviewViewport()
}

// updateEditableFocus 把可编辑控件的输入更新路由集中起来，避免主 Update
// 分支不断膨胀。
func (m Model) updateEditableFocus(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch m.focus {
	case focusContent:
		return m.updateContent(msg)
	case focusSize:
		return m.updateSize(msg)
	case focusOutput:
		return m.updateOutput(msg)
	default:
		return m, nil, false
	}
}

func (m Model) updateContent(msg tea.Msg) (Model, tea.Cmd, bool) {
	before := m.content.Value()
	var cmd tea.Cmd
	m.content, cmd = m.content.Update(msg)
	return m.afterEditableUpdate(m.content.Value() != before, false, cmd)
}

func (m Model) updateSize(msg tea.Msg) (Model, tea.Cmd, bool) {
	before := m.size.Value()
	var cmd tea.Cmd
	m.size, cmd = m.size.Update(msg)
	return m.afterEditableUpdate(m.size.Value() != before, false, cmd)
}

func (m Model) updateOutput(msg tea.Msg) (Model, tea.Cmd, bool) {
	before := m.output.Value()
	var cmd tea.Cmd
	m.output, cmd = m.output.Update(msg)
	return m.afterEditableUpdate(m.output.Value() != before, true, cmd)
}

// afterEditableUpdate 把“值是否变化”和“是否需要触发后续副作用”的判断收口
// 在一起，让各个输入框更新逻辑保持对称。
func (m Model) afterEditableUpdate(changed, explicitOutput bool, cmd tea.Cmd) (Model, tea.Cmd, bool) {
	if !changed {
		return m, cmd, true
	}
	next, nextCmd := m.afterValueChange(true, explicitOutput, cmd)
	return next, nextCmd, true
}

func (m Model) afterValueChange(changed, explicitOutput bool, cmd tea.Cmd) (Model, tea.Cmd) {
	if !changed {
		return m, cmd
	}
	if explicitOutput {
		m.outputDerived = false
	}
	m.clearTransientStatuses()
	return m, tea.Batch(cmd, m.schedulePreview())
}

func (m *Model) syncPreviewViewport() {
	width := max(24, m.previewMaxWidth-m.styles.previewCanvas.GetHorizontalFrameSize())
	height := max(5, m.previewMaxHeight-m.styles.previewCanvas.GetVerticalFrameSize())
	m.preview.SetWidth(width)
	m.preview.SetHeight(height)
	offset := m.preview.YOffset()
	if m.prepared != nil {
		m.previewText = m.prepared.PreviewFit(width, height)
		m.previewProto = "Matrix"
	} else {
		m.previewText = ""
		m.previewProto = ""
	}
	m.preview.SetContent(m.previewDocument(width, height))
	maxOffset := max(0, m.preview.TotalLineCount()-m.preview.Height())
	m.preview.SetYOffset(min(offset, maxOffset))
}

func (m *Model) clearTransientStatuses() {
	m.pathStatus = statusModel{}
	m.footerStatus = statusModel{}
}

func (m *Model) setPathError(err error) {
	status := statusModel{Kind: statusError, Symbol: "!", Message: humanizeError(err)}
	m.pathStatus = status
	m.footerStatus = status
}

// collectLevelModules 预先计算不同纠错等级在当前内容下需要的模块数，用于
// 在预览区即时给出“推荐降低等级/扩大终端”的提示。
func collectLevelModules(content string) map[core.Level]int {
	result := make(map[core.Level]int, len(levelOrder))
	for _, level := range levelOrder {
		modules, err := render.RequiredModules(content, level)
		if err != nil {
			continue
		}
		result[level] = modules
	}
	return result
}

// levelModulesForContent 使用一层按内容命中的浅缓存，避免用户只改主题或焦
// 点时重复计算同一组模块信息。
func (m Model) levelModulesForContent(content string) map[core.Level]int {
	if content == m.levelModulesContent && len(m.levelModules) > 0 {
		return cloneLevelModules(m.levelModules)
	}
	return collectLevelModules(content)
}

// cloneLevelModules 避免 map 直接共享给调用方后被意外修改。
func cloneLevelModules(input map[core.Level]int) map[core.Level]int {
	if len(input) == 0 {
		return nil
	}
	out := make(map[core.Level]int, len(input))
	for level, modules := range input {
		out[level] = modules
	}
	return out
}

func (m *Model) applyFocus() tea.Cmd {
	m.content.Blur()
	m.size.Blur()
	m.output.Blur()

	switch m.focus {
	case focusContent:
		return m.content.Focus()
	case focusSize:
		return m.size.Focus()
	case focusOutput:
		return m.output.Focus()
	default:
		return nil
	}
}

func (m Model) resetToDefaults() (Model, tea.Cmd) {
	m.content.SetValue("")
	m.format = core.FormatPNG
	m.level = core.LevelMedium
	m.size.SetValue("256")
	m.output.SetValue(core.DefaultOutputPath(core.FormatPNG))
	m.outputDerived = true
	m.prepared = nil
	m.previewText = ""
	m.previewProto = ""
	m.levelModules = nil
	m.levelModulesContent = ""
	m.previewStatus = statusModel{Kind: statusReady, Message: "Ready"}
	m.pathStatus = statusModel{}
	m.footerStatus = statusModel{}
	m.focus = focusContent

	return m, tea.Batch(m.applyFocus(), m.schedulePreview())
}
