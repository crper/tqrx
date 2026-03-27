package tui

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/crper/tqrx/internal/core"
	"github.com/crper/tqrx/internal/render"
)

func TestViewWideLayoutGolden(t *testing.T) {
	model := NewModel(render.NewEngine())
	next, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	got := plainView(next.(Model).View())
	assertGolden(t, "wide_empty.golden", got)
}

func TestViewNarrowLayoutGolden(t *testing.T) {
	model := NewModel(render.NewEngine())
	next, _ := model.Update(tea.WindowSizeMsg{Width: 72, Height: 40})
	got := plainView(next.(Model).View())
	assertGolden(t, "narrow_empty.golden", got)
}

func TestNewModelContentFocused(t *testing.T) {
	model := NewModel(render.NewEngine())
	if !model.content.Focused() {
		t.Fatal("content focused = false, want true")
	}
}

func TestEnterInContentAddsNewlineInsteadOfSaving(t *testing.T) {
	model := NewModel(render.NewEngine())
	model.content.SetValue("hello")
	model.focus = focusContent

	next, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Text: "\n"})
	got := next.(Model)

	if strings.Contains(got.pathStatus.Message, "Saved to") {
		t.Fatalf("path status = %q, want no save message", got.pathStatus.Message)
	}
	if got.focus != focusContent {
		t.Fatalf("focus = %v, want %v", got.focus, focusContent)
	}
}

func TestPasteMsgInContentUpdatesTextarea(t *testing.T) {
	model := NewModel(render.NewEngine())
	model.focus = focusContent
	model.applyFocus()

	next, _ := model.Update(tea.PasteMsg{Content: "第一行\n第二行"})
	got := next.(Model)

	if got.content.Value() != "第一行\n第二行" {
		t.Fatalf("content = %q, want pasted content", got.content.Value())
	}
	if got.previewStatus.Kind != "waiting" {
		t.Fatalf("preview status kind = %q, want waiting after paste", got.previewStatus.Kind)
	}
}

func TestPasteMsgInOutputDisablesDerivedPath(t *testing.T) {
	model := NewModel(render.NewEngine())
	model.output.SetValue("")
	model.focus = focusOutput
	model.applyFocus()

	next, _ := model.Update(tea.PasteMsg{Content: "/tmp/custom.png"})
	got := next.(Model)

	if got.output.Value() != "/tmp/custom.png" {
		t.Fatalf("output = %q, want pasted output path", got.output.Value())
	}
	if got.outputDerived {
		t.Fatal("outputDerived = true, want false after pasted output path")
	}
}

func TestFormatChangeSyncsDerivedOutputPath(t *testing.T) {
	model := NewModel(render.NewEngine())
	model.focus = focusFormat

	next, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	got := next.(Model)

	if got.format != core.FormatSVG {
		t.Fatalf("format = %q, want %q", got.format, core.FormatSVG)
	}
	if got.output.Value() != "./qrcode.svg" {
		t.Fatalf("output = %q, want ./qrcode.svg", got.output.Value())
	}
}

func TestNewModelUsesExplicitThemeOverride(t *testing.T) {
	t.Setenv("TQRX_THEME", "light")

	model := NewModel(render.NewEngine())
	if model.themeMode != uiThemeLight {
		t.Fatalf("themeMode = %q, want %q", model.themeMode, uiThemeLight)
	}
	if model.theme.dark {
		t.Fatal("theme.dark = true, want false for explicit light theme")
	}
}

func TestBackgroundColorMsgUpdatesThemeInAutoMode(t *testing.T) {
	model := NewModel(render.NewEngine())
	model.themeMode = uiThemeAuto
	model.applyTheme(resolveUITheme(uiThemeAuto, true))

	next, _ := model.Update(tea.BackgroundColorMsg{
		Color: color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
	})
	got := next.(Model)

	if got.theme.dark {
		t.Fatal("theme.dark = true, want false after light background update")
	}
	if got.theme.name != "light" {
		t.Fatalf("theme.name = %q, want light", got.theme.name)
	}
}

func TestLightThemeUsesReadableSurfaces(t *testing.T) {
	model := NewModel(render.NewEngine())
	model.themeMode = uiThemeLight
	model.applyTheme(resolveUITheme(uiThemeLight, false))

	if model.View().BackgroundColor != model.theme.appBg {
		t.Fatalf("view background = %#v, want %#v", model.View().BackgroundColor, model.theme.appBg)
	}
	if _, ok := model.styles.field.GetBackground().(lipgloss.NoColor); ok {
		t.Fatal("field background = lipgloss.NoColor, want explicit field surface")
	}
	if _, ok := model.styles.previewCanvas.GetBackground().(lipgloss.NoColor); ok {
		t.Fatal("previewCanvas background = lipgloss.NoColor, want explicit surface for light theme")
	}
	if _, ok := model.content.Styles().Focused.Base.GetBackground().(lipgloss.NoColor); ok {
		t.Fatal("textarea focused base background = lipgloss.NoColor, want explicit field surface")
	}
	if _, ok := model.size.Styles().Focused.Text.GetBackground().(lipgloss.NoColor); ok {
		t.Fatal("textinput focused text background = lipgloss.NoColor, want explicit field surface")
	}
	if _, ok := model.size.Styles().Focused.Prompt.GetBackground().(lipgloss.NoColor); ok {
		t.Fatal("textinput focused prompt background = lipgloss.NoColor, want explicit field surface")
	}
}

func TestRenderSurfaceWithEmptyContentMatchesBackgroundBlock(t *testing.T) {
	model := NewModel(render.NewEngine())
	spec := surfaceSpec{
		content:    "",
		width:      12,
		height:     3,
		background: model.styles.field,
	}

	got := renderSurface(spec)

	if lipgloss.Width(got) != spec.width {
		t.Fatalf("renderSurface(empty) width = %d, want %d", lipgloss.Width(got), spec.width)
	}
	if lipgloss.Height(got) != spec.height {
		t.Fatalf("renderSurface(empty) height = %d, want %d", lipgloss.Height(got), spec.height)
	}

	backgroundRow := spec.background.Render(strings.Repeat(" ", spec.width))
	if !strings.Contains(got, backgroundRow) {
		t.Fatalf("renderSurface(empty) = %q, want background row %q", got, backgroundRow)
	}
}

func TestTextareaFocusedCursorLineStaysTransparent(t *testing.T) {
	model := NewModel(render.NewEngine())
	styles := model.content.Styles()

	if _, ok := styles.Focused.CursorLine.GetBackground().(lipgloss.NoColor); !ok {
		t.Fatalf("Focused.CursorLine background = %#v, want lipgloss.NoColor", styles.Focused.CursorLine.GetBackground())
	}
	if styles.Cursor.Shape != tea.CursorBar {
		t.Fatalf("Cursor.Shape = %v, want %v", styles.Cursor.Shape, tea.CursorBar)
	}
}

func TestRenderTextareaSurfaceKeepsBackgroundOnEmptyRows(t *testing.T) {
	model := sizedModel(t, 120, 40)
	model.content.SetValue("")

	spec := model.textareaSurfaceSpec()
	surface := model.renderTextareaSurface()

	if lipgloss.Width(surface) != spec.width {
		t.Fatalf("renderTextareaSurface() width = %d, want %d", lipgloss.Width(surface), spec.width)
	}
	if lipgloss.Height(surface) != spec.height {
		t.Fatalf("renderTextareaSurface() height = %d, want %d", lipgloss.Height(surface), spec.height)
	}

	lines := strings.Split(ansi.Strip(surface), "\n")
	if len(lines) < 2 {
		t.Fatalf("renderTextareaSurface() line count = %d, want at least 2", len(lines))
	}
	if got := len([]rune(lines[1])); got != spec.width {
		t.Fatalf("renderTextareaSurface() second row width = %d, want %d", got, spec.width)
	}
}

func TestPreviewEmptyStateIsVerticallyCentered(t *testing.T) {
	model := NewModel(render.NewEngine())

	doc := ansi.Strip(model.previewDocument(64, 9))
	lines := strings.Split(strings.TrimRight(doc, "\n"), "\n")
	messageLine := -1
	for i, line := range lines {
		if strings.Contains(line, "Paste text or a link to render a QR.") {
			messageLine = i
			break
		}
	}

	if messageLine < 2 {
		t.Fatalf("empty-state message line = %d, want vertically centered content", messageLine)
	}
}

func TestCtrlTTogglesThemeMode(t *testing.T) {
	model := NewModel(render.NewEngine())
	model.themeMode = uiThemeAuto
	model.applyTheme(resolveUITheme(uiThemeAuto, true))

	next, _ := model.Update(tea.KeyPressMsg{
		Code: 't',
		Text: "t",
		Mod:  tea.ModCtrl,
	})
	got := next.(Model)

	if got.themeMode != uiThemeLight {
		t.Fatalf("themeMode = %q, want %q after first toggle", got.themeMode, uiThemeLight)
	}
	if got.theme.dark {
		t.Fatal("theme.dark = true, want false after switching to light mode")
	}

	next, _ = got.Update(tea.KeyPressMsg{
		Code: 't',
		Text: "t",
		Mod:  tea.ModCtrl,
	})
	got = next.(Model)
	if got.themeMode != uiThemeDark {
		t.Fatalf("themeMode = %q, want %q after second toggle", got.themeMode, uiThemeDark)
	}

	next, _ = got.Update(tea.KeyPressMsg{
		Code: 't',
		Text: "t",
		Mod:  tea.ModCtrl,
	})
	got = next.(Model)
	if got.themeMode != uiThemeAuto {
		t.Fatalf("themeMode = %q, want %q after third toggle", got.themeMode, uiThemeAuto)
	}
}

func TestHeaderShowsThemeModeChip(t *testing.T) {
	model := NewModel(render.NewEngine())
	model.themeMode = uiThemeLight
	model.applyTheme(resolveUITheme(uiThemeLight, false))

	view := plainView(model.View())
	if !strings.Contains(view, "[LIGHT]") {
		t.Fatalf("view = %q, want [LIGHT] theme chip", view)
	}
}

func TestStalePreviewResultIsIgnored(t *testing.T) {
	engine := render.NewEngine()
	model := NewModel(engine)
	model.pendingPreviewID = 2
	model.previewText = "newest"

	oldReq, err := core.Normalize(core.Request{
		Content: "old",
		Source:  core.SourceTUI,
	})
	if err != nil {
		t.Fatalf("Normalize(old) error = %v", err)
	}
	oldPrepared, err := engine.Prepare(oldReq)
	if err != nil {
		t.Fatalf("Prepare(old) error = %v", err)
	}

	next, _ := model.Update(previewReadyMsg{
		id:       1,
		prepared: oldPrepared,
	})
	got := next.(Model)

	if got.previewText != "newest" {
		t.Fatalf("preview = %q, want stale result ignored", got.previewText)
	}
}

func TestPreviewErrorUpdatesMetadataAndFooter(t *testing.T) {
	model := NewModel(render.NewEngine())
	model.content.SetValue("preview me")
	model.size.SetValue("100x200")
	model.pendingPreviewID = 1

	next, cmd := model.Update(previewTickMsg{id: 1})
	if cmd == nil {
		t.Fatal("previewTick cmd = nil, want follow-up previewReadyMsg")
	}

	msg := cmd()
	updated, _ := next.(Model).Update(msg)
	got := updated.(Model)

	if got.previewStatus.Kind != "error" {
		t.Fatalf("preview status kind = %q, want error", got.previewStatus.Kind)
	}
	if got.footerStatus.Kind != "error" {
		t.Fatalf("footer status kind = %q, want error", got.footerStatus.Kind)
	}
	if !strings.Contains(got.previewStatus.Message, "Size must be square") {
		t.Fatalf("preview status = %q, want size guidance", got.previewStatus.Message)
	}
	if !strings.Contains(got.footerMessage(), "Size must be square") {
		t.Fatalf("footer = %q, want size guidance", got.footerMessage())
	}
}

func TestPreparePreviewCmdReusesLevelModuleCacheForSameContent(t *testing.T) {
	model := NewModel(render.NewEngine())
	model.content.SetValue("cache me")
	model.levelModulesContent = "cache me"
	model.levelModules = map[core.Level]int{
		core.LevelLow:    91,
		core.LevelMedium: 92,
		core.LevelQuart:  93,
		core.LevelHigh:   94,
	}

	cmd := model.preparePreviewCmd(7)
	if cmd == nil {
		t.Fatal("preparePreviewCmd() = nil, want previewReadyMsg command")
	}

	msg, ok := cmd().(previewReadyMsg)
	if !ok {
		t.Fatalf("preparePreviewCmd() message = %T, want previewReadyMsg", cmd())
	}

	for level, want := range model.levelModules {
		if got := msg.levelModules[level]; got != want {
			t.Fatalf("preparePreviewCmd() %s modules = %d, want cached value %d", level, got, want)
		}
	}

	msg.levelModules[core.LevelLow] = 0
	if got := model.levelModules[core.LevelLow]; got != 91 {
		t.Fatalf("cached level modules mutated to %d, want original value 91", got)
	}
}

func TestPreviewErrorClearsStalePreview(t *testing.T) {
	engine := render.NewEngine()
	model := NewModel(engine)

	req, err := core.Normalize(core.Request{
		Content: "old preview",
		Source:  core.SourceTUI,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	prepared, err := engine.Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	model.prepared = prepared
	model.previewText = prepared.Preview()
	model.previewProto = "Matrix"
	model.syncPreviewViewport()
	model.pendingPreviewID = 3

	next, _ := model.Update(previewReadyMsg{
		id: 3,
		err: &core.UserError{
			Kind:    core.ErrorSizeTooSmall,
			Message: "size must be at least 29 for this content",
		},
	})
	got := next.(Model)

	if got.prepared != nil {
		t.Fatalf("prepared = %#v, want nil", got.prepared)
	}
	if got.previewText != "" {
		t.Fatalf("preview = %q, want empty preview after error", got.previewText)
	}
	if got.previewStatus.Kind != "error" {
		t.Fatalf("preview status kind = %q, want error", got.previewStatus.Kind)
	}
}

func TestWaitingStatusDoesNotChangePreviewPanelHeight(t *testing.T) {
	model := sizedModel(t, 120, 40)

	readyHeight := lipgloss.Height(model.renderPreviewPanel(69))

	model.previewStatus = statusModel{Kind: "waiting", Symbol: "…", Message: "Updating"}
	waitingHeight := lipgloss.Height(model.renderPreviewPanel(69))

	if waitingHeight != readyHeight {
		t.Fatalf("renderPreviewPanel() height = %d, want %d while waiting", waitingHeight, readyHeight)
	}
}

func TestPreviewMetaSharesSingleAdaptiveRailWhenWide(t *testing.T) {
	model := sizedModel(t, 120, 40)
	model.content.SetValue("meta rail")

	req, err := model.currentRequest()
	if err != nil {
		t.Fatalf("currentRequest() error = %v", err)
	}
	prepared, err := render.NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	model.prepared = prepared
	model.previewProto = "Matrix"
	model.syncPreviewViewport()

	meta := ansi.Strip(model.renderPreviewMeta(90))
	lines := strings.Split(meta, "\n")
	if len(lines) != 1 {
		t.Fatalf("renderPreviewMeta() line count = %d, want 1 for wide layout", len(lines))
	}
	if !strings.Contains(lines[0], "via Matrix") || !strings.Contains(lines[0], "Path ./qrcode.png") {
		t.Fatalf("renderPreviewMeta() = %q, want protocol and path on same line", lines[0])
	}
	if !strings.Contains(lines[0], "mods ") {
		t.Fatalf("renderPreviewMeta() = %q, want module/capacity summary", lines[0])
	}
}

func TestPreviewMetaWrapsLongWidePath(t *testing.T) {
	model := sizedModel(t, 120, 40)
	model.content.SetValue("meta rail")
	model.output.SetValue("/tmp/二维码导出/这是一个特别特别长的输出路径-用于验证预览元信息换行是否稳定并且和布局估算一致.png")
	model.outputDerived = false

	req, err := model.currentRequest()
	if err != nil {
		t.Fatalf("currentRequest() error = %v", err)
	}
	prepared, err := render.NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	model.prepared = prepared
	model.previewProto = "Matrix"
	model.syncPreviewViewport()

	meta := ansi.Strip(model.renderPreviewMeta(69))
	lines := strings.Split(meta, "\n")
	if len(lines) != 2 {
		t.Fatalf("renderPreviewMeta() line count = %d, want 2 for wrapped wide path", len(lines))
	}
	if !strings.Contains(lines[0], "via Matrix") {
		t.Fatalf("renderPreviewMeta() first line = %q, want protocol info", lines[0])
	}
	if !strings.Contains(lines[1], "Path ") {
		t.Fatalf("renderPreviewMeta() second line = %q, want wrapped path line", lines[1])
	}
}

func TestPreviewMetaSuggestsLowerLevelWhenOverCapacity(t *testing.T) {
	model := sizedModel(t, 120, 40)
	model.level = core.LevelQuart
	model.content.SetValue("fasdkfjasdlkfasdkfldajsfklajflkfaksd\nlfjadsklfjadsklfjadsklf\njadsklfjadsklfjadsklfjadsklfja\ndsfasdfhkjhkjjhkjhjhk")

	req, err := model.currentRequest()
	if err != nil {
		t.Fatalf("currentRequest() error = %v", err)
	}
	prepared, err := render.NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	model.prepared = prepared
	model.previewProto = "Matrix"
	model.levelModules = collectLevelModules(req.Content)
	model.preview.SetWidth(56)
	model.preview.SetHeight(30)

	meta := ansi.Strip(model.renderPreviewMeta(120))
	if !strings.Contains(meta, "mods ") {
		t.Fatalf("renderPreviewMeta() = %q, want module/capacity summary", meta)
	}
	if !strings.Contains(meta, "suggest M for scan") {
		t.Fatalf("renderPreviewMeta() = %q, want level suggestion for scan", meta)
	}
}

func TestPreviewViewportAutoFitsWithoutScrolling(t *testing.T) {
	model := sizedModel(t, 120, 40)
	model.content.SetValue("https://example.com/a/preview/that/should/fit")
	model.level = core.LevelHigh

	req, err := model.currentRequest()
	if err != nil {
		t.Fatalf("currentRequest() error = %v", err)
	}
	prepared, err := render.NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	model.prepared = prepared
	model.previewProto = "Matrix"
	model.syncPreviewViewport()

	if got := lipgloss.Width(model.previewText); got > model.preview.Width() {
		t.Fatalf("preview width = %d, want <= %d", got, model.preview.Width())
	}
	if got := lipgloss.Height(model.previewText); got > model.preview.Height() {
		t.Fatalf("preview height = %d, want <= %d", got, model.preview.Height())
	}
	if model.previewScrollable() {
		t.Fatal("previewScrollable() = true, want false after auto-fit")
	}
}

func TestPreviewFocusHintWarnsWhenPreviewIsCondensed(t *testing.T) {
	model := sizedModel(t, 120, 40)
	model.content.SetValue("发的上课了放假快乐 sd 卡放假啦 sd 卡房间为埃及人 weakly 放假 ADSL 客服啊圣诞快乐发阿斯蒂芬看阿斯蒂芬阿斯蒂芬跨时代开放啦是的副卡就是的罚款了是的积分啊上看到了放假啊圣诞快乐发阿斯蒂芬克拉斯都发开始了地方啊圣诞快乐发阿斯蒂芬开了撒旦法是短发收到了客服阿斯蒂芬集卡老师的飞机 adslkffafffffffffffffffffjjlkje2kjeio2u09u 批发第三方届奥斯卡放假啊放假啊酸辣粉卡时间发快手发阿斯蒂芬撒旦法撒旦法是")
	model.level = core.LevelHigh

	req, err := model.currentRequest()
	if err != nil {
		t.Fatalf("currentRequest() error = %v", err)
	}
	prepared, err := render.NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	model.prepared = prepared
	model.previewProto = "Matrix"
	model.syncPreviewViewport()

	if !model.previewCondensed() {
		t.Fatal("previewCondensed() = false, want true for high-density preview")
	}
	if got := model.previewFocusHint(); got != "native preview exceeds viewport; enlarge terminal" {
		t.Fatalf("previewFocusHint() = %q, want condensed warning", got)
	}
}

func TestViewRendersMultiplePreviewLines(t *testing.T) {
	engine := render.NewEngine()
	model := NewModel(engine)
	model.content.SetValue("multi line preview")

	req, err := model.currentRequest()
	if err != nil {
		t.Fatalf("currentRequest() error = %v", err)
	}
	prepared, err := engine.Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	model.prepared = prepared
	model.previewProto = "Matrix"
	model.syncPreviewViewport()

	var previewLines []string
	for _, line := range strings.Split(model.previewText, "\n") {
		if strings.TrimSpace(line) != "" {
			previewLines = append(previewLines, line)
		}
	}
	if len(previewLines) < 2 {
		t.Fatalf("preview lines = %v, want at least 2 non-empty lines", previewLines)
	}

	next, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	view := plainView(next.(Model).View())
	if !strings.Contains(view, previewLines[0]) || !strings.Contains(view, previewLines[1]) {
		t.Fatalf("view = %q, want multiple preview lines rendered", view)
	}
}

func TestFocusedSaveWritesFileAndAnchorsStatus(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "saved.png")

	engine := render.NewEngine()
	model := NewModel(engine)
	model.output.SetValue(target)
	model.outputDerived = false
	model.content.SetValue("save me")
	model.focus = focusSave

	req, err := model.currentRequest()
	if err != nil {
		t.Fatalf("currentRequest() error = %v", err)
	}
	prepared, err := engine.Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	model.prepared = prepared
	model.previewText = prepared.Preview()
	model.syncPreviewViewport()

	next, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := next.(Model)

	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected %s to exist: %v", target, err)
	}
	if !strings.Contains(got.pathStatus.Message, "Saved to") {
		t.Fatalf("path status = %q, want save confirmation", got.pathStatus.Message)
	}
	if !strings.Contains(got.footerStatus.Message, "Saved to") {
		t.Fatalf("footer status = %q, want save confirmation", got.footerStatus.Message)
	}
	view := plainView(got.View())
	if !strings.Contains(view, "✓ Saved to "+target) {
		t.Fatalf("view = %q, want path-anchored save status", view)
	}
}

func TestCtrlSSavesFromContentFocus(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "save-shortcut.png")

	model := NewModel(render.NewEngine())
	model.output.SetValue(target)
	model.outputDerived = false
	model.content.SetValue("shortcut save")
	model.focus = focusContent
	model.applyFocus()

	next, _ := model.Update(tea.KeyPressMsg{
		Code: 's',
		Text: "s",
		Mod:  tea.ModCtrl,
	})
	got := next.(Model)

	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected %s to exist: %v", target, err)
	}
	if got.pathStatus.Kind != "success" {
		t.Fatalf("path status kind = %q, want success", got.pathStatus.Kind)
	}
}

func TestMouseClickControlsPanelFocusesControls(t *testing.T) {
	model := sizedModel(t, 120, 40)
	rects := model.layoutRects()

	next, _ := model.Update(mouseClickAt(rects.controls.x+1, rects.controls.y))
	got := next.(Model)

	if got.focus != focusFormat {
		t.Fatalf("focus = %v, want %v", got.focus, focusFormat)
	}
}

func TestMouseClickControlRowsFocusesMatchingField(t *testing.T) {
	tests := []struct {
		name string
		row  func(layoutRects) rect
		want focusTarget
	}{
		{
			name: "size",
			row: func(rects layoutRects) rect {
				return rects.controlRows.size
			},
			want: focusSize,
		},
		{
			name: "level",
			row: func(rects layoutRects) rect {
				return rects.controlRows.level
			},
			want: focusLevel,
		},
		{
			name: "output",
			row: func(rects layoutRects) rect {
				return rects.controlRows.output
			},
			want: focusOutput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := sizedModel(t, 120, 40)
			rects := model.layoutRects()

			next, _ := model.Update(mouseClick(tt.row(rects)))
			got := next.(Model)

			if got.focus != tt.want {
				t.Fatalf("focus = %v, want %v", got.focus, tt.want)
			}
		})
	}
}

func TestMouseClickFormatChipChangesFormat(t *testing.T) {
	model := sizedModel(t, 120, 40)
	rects := model.layoutRects()
	target := formatChipByValue(t, rects.controlRows.formatChips, core.FormatSVG)

	next, _ := model.Update(mouseClick(target.rect))
	got := next.(Model)

	if got.focus != focusFormat {
		t.Fatalf("focus = %v, want %v", got.focus, focusFormat)
	}
	if got.format != core.FormatSVG {
		t.Fatalf("format = %q, want %q", got.format, core.FormatSVG)
	}
	if got.output.Value() != "./qrcode.svg" {
		t.Fatalf("output = %q, want ./qrcode.svg", got.output.Value())
	}
}

func TestMouseClickLevelChipChangesLevel(t *testing.T) {
	model := sizedModel(t, 120, 40)
	rects := model.layoutRects()
	target := levelChipByValue(t, rects.controlRows.levelChips, core.LevelHigh)

	next, _ := model.Update(mouseClick(target.rect))
	got := next.(Model)

	if got.focus != focusLevel {
		t.Fatalf("focus = %v, want %v", got.focus, focusLevel)
	}
	if got.level != core.LevelHigh {
		t.Fatalf("level = %q, want %q", got.level, core.LevelHigh)
	}
}

func TestMouseClickPreviewSaveButtonFocusesSave(t *testing.T) {
	model := sizedModel(t, 120, 40)
	rects := model.layoutRects()

	next, _ := model.Update(mouseClick(rects.saveButton))
	got := next.(Model)

	if got.focus != focusSave {
		t.Fatalf("focus = %v, want %v", got.focus, focusSave)
	}
}

func TestMouseWheelOverPreviewFocusesAndScrolls(t *testing.T) {
	model := sizedModel(t, 120, 40)
	model.focus = focusContent
	model.preview.SetWidth(32)
	model.preview.SetHeight(8)
	model.preview.SetContent(strings.TrimSuffix(strings.Repeat("scroll me\n", 64), "\n"))
	rects := model.layoutRects()

	next, _ := model.Update(tea.MouseWheelMsg{
		X:      rects.preview.x + 1,
		Y:      rects.preview.y + 1,
		Button: tea.MouseWheelDown,
	})
	got := next.(Model)

	if got.focus != focusPreview {
		t.Fatalf("focus = %v, want %v", got.focus, focusPreview)
	}
	if got.preview.YOffset() == 0 {
		t.Fatalf("preview YOffset = %d, want scroll after mouse wheel", got.preview.YOffset())
	}
}

func TestLayoutPlanMatchesRenderedPanels(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		adjust func(*Model)
	}{
		{
			name:   "wide ready",
			width:  120,
			height: 40,
		},
		{
			name:   "narrow ready",
			width:  72,
			height: 40,
		},
		{
			name:   "wide status and error",
			width:  120,
			height: 40,
			adjust: func(model *Model) {
				model.pathStatus = statusModel{Kind: "success", Symbol: "✓", Message: "Saved to ./qrcode.png"}
				model.previewStatus = statusModel{Kind: "error", Symbol: "!", Message: "Can't write to this path."}
			},
		},
		{
			name:   "wide long unicode path",
			width:  120,
			height: 40,
			adjust: func(model *Model) {
				model.output.SetValue("/tmp/二维码导出/这是一个特别特别长的输出路径-用于验证预览元信息换行是否稳定并且和布局估算一致.png")
				model.outputDerived = false
				model.previewProto = "Matrix"
			},
		},
		{
			name:   "wide scan summary wraps",
			width:  120,
			height: 40,
			adjust: func(model *Model) {
				model.level = core.LevelQuart
				model.content.SetValue("fasdkfjasdlkfasdkfldajsfklajflkfaksd\nlfjadsklfjadsklfjadsklf\njadsklfjadsklfjadsklfjadsklfja\ndsfasdfhkjhkjjhkjhjhk")

				req, err := model.currentRequest()
				if err != nil {
					t.Fatalf("currentRequest() error = %v", err)
				}
				prepared, err := render.NewEngine().Prepare(req)
				if err != nil {
					t.Fatalf("Prepare() error = %v", err)
				}

				model.prepared = prepared
				model.previewProto = "Matrix"
				model.levelModules = collectLevelModules(req.Content)
				model.syncPreviewViewport()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := sizedModel(t, tt.width, tt.height)
			if tt.adjust != nil {
				tt.adjust(&model)
			}

			plan := model.planLayout()

			if got, want := plan.editPanel.h, lipgloss.Height(model.renderEditPanel(plan.editWidth)); got != want {
				t.Fatalf("planLayout().editPanel.h = %d, want %d", got, want)
			}
			if got, want := plan.previewPanel.h, lipgloss.Height(model.renderPreviewPanel(plan.previewWidth)); got != want {
				t.Fatalf("planLayout().previewPanel.h = %d, want %d", got, want)
			}
			if got, want := plan.rects.saveButton.y, plan.previewPanel.y+plan.previewPanel.h; got != want {
				t.Fatalf("planLayout().saveButton.y = %d, want %d", got, want)
			}
		})
	}
}

func TestViewEnablesAltScreenAndMouseMode(t *testing.T) {
	model := NewModel(render.NewEngine())
	view := model.View()

	if !view.AltScreen {
		t.Fatal("AltScreen = false, want true")
	}
	if view.MouseMode != tea.MouseModeCellMotion {
		t.Fatalf("MouseMode = %v, want %v", view.MouseMode, tea.MouseModeCellMotion)
	}
	if view.WindowTitle != "tqrx" {
		t.Fatalf("WindowTitle = %q, want tqrx", view.WindowTitle)
	}
	if view.BackgroundColor != model.theme.appBg {
		t.Fatalf("BackgroundColor = %#v, want %#v", view.BackgroundColor, model.theme.appBg)
	}
}

func sizedModel(t *testing.T, width, height int) Model {
	t.Helper()

	next, _ := NewModel(render.NewEngine()).Update(tea.WindowSizeMsg{Width: width, Height: height})
	return next.(Model)
}

func mouseClick(r rect) tea.MouseClickMsg {
	return mouseClickAt(r.x+max(0, (r.w-1)/2), r.y+max(0, (r.h-1)/2))
}

func mouseClickAt(x, y int) tea.MouseClickMsg {
	return tea.MouseClickMsg{
		X:      x,
		Y:      y,
		Button: tea.MouseLeft,
	}
}

func formatChipByValue(t *testing.T, chips []formatChipRect, want core.Format) formatChipRect {
	t.Helper()

	for _, chip := range chips {
		if chip.format == want {
			return chip
		}
	}
	t.Fatalf("format chip = %q, want existing chip", want)
	return formatChipRect{}
}

func levelChipByValue(t *testing.T, chips []levelChipRect, want core.Level) levelChipRect {
	t.Helper()

	for _, chip := range chips {
		if chip.level == want {
			return chip
		}
	}
	t.Fatalf("level chip = %q, want existing chip", want)
	return levelChipRect{}
}

func plainView(view tea.View) string {
	lines := strings.Split(strings.TrimRight(ansi.Strip(view.Content), "\n"), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()

	want, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", name, err)
	}
	wantText := strings.TrimRight(string(want), "\n")
	gotText := strings.TrimRight(got, "\n")
	if gotText != wantText {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, gotText, wantText)
	}
}
