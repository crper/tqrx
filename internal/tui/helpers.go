package tui

import (
	"errors"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/crper/tqrx/internal/core"
)

var (
	levelOrder = [...]core.Level{
		core.LevelLow,
		core.LevelMedium,
		core.LevelQuart,
		core.LevelHigh,
	}
	descendingLevelOrder = [...]core.Level{
		core.LevelHigh,
		core.LevelQuart,
		core.LevelMedium,
		core.LevelLow,
	}
	levelIndex = func() map[core.Level]int {
		m := make(map[core.Level]int, len(levelOrder))
		for i, level := range levelOrder {
			m[level] = i
		}
		return m
	}()
)

// formatLabels 从 formatChoices 派生小写标签，供 chip 行渲染使用。
// 作为包级变量预计算，避免每次渲染分配临时切片。
var formatLabels = func() []string {
	labels := make([]string, len(formatChoices))
	for i, f := range formatChoices {
		labels[i] = strings.ToLower(string(f))
	}
	return labels
}()

// levelLabels 从 levelOrder 派生字符串标签，供 chip 行渲染使用。
// 作为包级变量预计算，避免每次渲染分配临时切片。
var levelLabels = func() []string {
	labels := make([]string, len(levelOrder))
	for i, l := range levelOrder {
		labels[i] = string(l)
	}
	return labels
}()

// nextFocus / prevFocus 维持一个稳定的环形焦点顺序，让 tab 和 shift+tab
// 在所有交互区之间可预期地循环。
func nextFocus(current focusTarget) focusTarget {
	if current == focusSave {
		return focusContent
	}
	return current + 1
}

func prevFocus(current focusTarget) focusTarget {
	if current == focusContent {
		return focusSave
	}
	return current - 1
}

// nextThemeMode 把主题切换逻辑限制在 auto -> light -> dark 这条固定环上。
func nextThemeMode(current uiThemeMode) uiThemeMode {
	switch current {
	case uiThemeAuto:
		return uiThemeLight
	case uiThemeLight:
		return uiThemeDark
	default:
		return uiThemeAuto
	}
}

func isCycleKey(msg tea.KeyPressMsg) bool {
	switch msg.Code {
	case tea.KeyLeft, tea.KeyRight, tea.KeyEnter:
		return true
	default:
		return msg.Text == " "
	}
}

// applyFormatCycle 和 applyLevelCycle 把键盘“循环选择”规则收敛成独立函数，
// 避免 Update 里混入太多控件细节。
func applyFormatCycle(msg tea.KeyPressMsg, format *core.Format) bool {
	if !isCycleKey(msg) {
		return false
	}
	if *format == formatChoices[0] {
		*format = formatChoices[1]
		return true
	}
	*format = formatChoices[0]
	return true
}

func applyLevelCycle(msg tea.KeyPressMsg, level *core.Level) bool {
	if !isCycleKey(msg) {
		return false
	}

	current, ok := levelIndex[*level]
	if !ok {
		current = 0
	}

	if msg.Code == tea.KeyLeft {
		current = (current + len(levelOrder) - 1) % len(levelOrder)
	} else {
		current = (current + 1) % len(levelOrder)
	}
	*level = levelOrder[current]
	return true
}

func panelHasFocus(current focusTarget, targets ...focusTarget) bool {
	for _, target := range targets {
		if current == target {
			return true
		}
	}
	return false
}

func controlsFocus(current focusTarget) focusTarget {
	if panelHasFocus(current, focusFormat, focusSize, focusLevel, focusOutput) {
		return current
	}
	return focusFormat
}

func statusText(status statusModel) string {
	if status.Symbol == "" {
		return status.Message
	}
	return status.Symbol + " " + status.Message
}

func shouldShowPreviewInlineStatus(status statusModel) bool {
	if status.Message == "" {
		return false
	}
	return status.Kind == statusError
}

// humanizeError 把 core/render/os 层的错误统一翻译成适合终端界面展示的文
// 案，尽量避免把内部实现细节直接暴露给用户。
var userErrorMessages = map[core.ErrorKind]string{
	core.ErrorEmptyContent:           "Type text or paste a link.",
	core.ErrorInvalidFormat:          "Format must be png or svg.",
	core.ErrorInvalidOutputExtension: "Output path must end with .png or .svg.",
	core.ErrorFormatMismatch:         "Output extension must match the selected format.",
	core.ErrorInvalidSize:            "Size must be square, like 256 or 256x256.",
	core.ErrorInvalidLevel:           "Level must be one of L, M, Q, H.",
}

func humanizeError(err error) string {
	if err == nil {
		return ""
	}

	var userErr *core.UserError
	if errors.As(err, &userErr) {
		if msg, ok := userErrorMessages[userErr.Kind]; ok {
			return msg
		}
		return sentenceCase(userErr.Message)
	}

	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return "Can't write to this path."
	}

	return sentenceCase(err.Error())
}

// sentenceCase 负责把底层错误文本规范成适合 UI 直接展示的一句话。
func sentenceCase(message string) string {
	if message == "" {
		return ""
	}
	head := strings.ToUpper(message[:1])
	if strings.HasSuffix(message, ".") {
		return head + message[1:]
	}
	return head + message[1:] + "."
}

// matchesKey 同时兼容 Bubble Tea 的字符串表示和 keystroke 表示，减少不同
// 平台/终端组合下的快捷键判断差异。
func matchesKey(msg tea.KeyPressMsg, bindings ...key.Binding) bool {
	plain := msg.String()
	stroke := msg.Keystroke()
	for _, binding := range bindings {
		if !binding.Enabled() {
			continue
		}
		for _, candidate := range binding.Keys() {
			if candidate == plain || candidate == stroke {
				return true
			}
		}
	}
	return false
}

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
