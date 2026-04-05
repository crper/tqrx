package render

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liyue201/goqr"

	"github.com/crper/tqrx/internal/core"
)

func TestPreparePNGDecodesBackToPayload(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "https://example.com",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	engine := NewEngine()
	prepared, err := engine.Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	pngData, err := prepared.PNG()
	if err != nil {
		t.Fatalf("PNG() error = %v", err)
	}

	img, _, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("image.Decode() error = %v", err)
	}

	qrCodes, err := goqr.Recognize(img)
	if err != nil {
		t.Fatalf("goqr.Recognize() error = %v", err)
	}
	if len(qrCodes) != 1 {
		t.Fatalf("Recognize() len = %d, want 1", len(qrCodes))
	}
	if got := string(qrCodes[0].Payload); got != req.Content {
		t.Fatalf("Payload = %q, want %q", got, req.Content)
	}
}

func TestPreparePNGDecodesBackToPayloadAcrossLevels(t *testing.T) {
	levels := []core.Level{
		core.LevelLow,
		core.LevelMedium,
		core.LevelQuart,
		core.LevelHigh,
	}

	for _, level := range levels {
		t.Run(string(level), func(t *testing.T) {
			req, err := core.Normalize(core.Request{
				Content: "https://example.com/level-check?mode=" + strings.ToLower(string(level)),
				Level:   string(level),
				Size:    "256",
				Source:  core.SourceCLIArg,
			})
			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}

			prepared, err := NewEngine().Prepare(req)
			if err != nil {
				t.Fatalf("Prepare() error = %v", err)
			}

			pngData, err := prepared.PNG()
			if err != nil {
				t.Fatalf("PNG() error = %v", err)
			}

			img, _, err := image.Decode(bytes.NewReader(pngData))
			if err != nil {
				t.Fatalf("image.Decode() error = %v", err)
			}

			qrCodes, err := goqr.Recognize(img)
			if err != nil {
				t.Fatalf("goqr.Recognize() error = %v", err)
			}
			if len(qrCodes) != 1 {
				t.Fatalf("Recognize() len = %d, want 1", len(qrCodes))
			}
			if got := string(qrCodes[0].Payload); got != req.Content {
				t.Fatalf("Payload = %q, want %q", got, req.Content)
			}
		})
	}
}

func TestPrepareSVGUsesRequestedSize(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "https://example.com",
		Format:  "svg",
		Size:    "320x320",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	engine := NewEngine()
	prepared, err := engine.Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	svgData, err := prepared.SVG()
	if err != nil {
		t.Fatalf("SVG() error = %v", err)
	}

	svg := string(svgData)
	for _, want := range []string{
		`<svg`,
		`width="320"`,
		`height="320"`,
		`viewBox="0 0 320 320"`,
		`<rect`,
	} {
		if !strings.Contains(svg, want) {
			t.Fatalf("SVG missing %q", want)
		}
	}
}

func TestPreparePreviewReturnsBlockArt(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "preview",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	engine := NewEngine()
	prepared, err := engine.Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	preview := prepared.Preview()
	if preview == "" {
		t.Fatal("Preview() = empty string")
	}
	if !strings.ContainsAny(preview, "█▀▄") {
		t.Fatalf("Preview() = %q, want block characters", preview)
	}
	if strings.HasPrefix(preview, "\n") || strings.HasSuffix(preview, "\n") {
		t.Fatalf("Preview() = %q, want trimmed preview lines", preview)
	}
}

func TestPreparePreviewKeepsQuietZonePadding(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "https://example.com",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	prepared, err := NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	lines := strings.Split(prepared.Preview(), "\n")
	if len(lines) < 3 {
		t.Fatalf("Preview() line count = %d, want at least 3", len(lines))
	}
	if strings.TrimSpace(lines[0]) != "" {
		t.Fatalf("Preview() first line = %q, want quiet-zone padding", lines[0])
	}
	if strings.TrimSpace(lines[len(lines)-1]) != "" {
		t.Fatalf("Preview() last line = %q, want quiet-zone padding", lines[len(lines)-1])
	}

	wantWidth := len([]rune(lines[0]))
	for i, line := range lines {
		if got := len([]rune(line)); got != wantWidth {
			t.Fatalf("Preview() line %d width = %d, want %d", i, got, wantWidth)
		}
	}
}

func TestPreviewFitReturnsBasePreviewAtNativeSize(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "native preview",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	prepared, err := NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	modules := prepared.PreviewModules()
	if got := prepared.PreviewFit(modules, modules); got != prepared.Preview() {
		t.Fatalf("PreviewFit() = %q, want cached base preview %q", got, prepared.Preview())
	}
}

func TestPreviewFitKeepsNativeModulesWhenViewportTooSmall(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "https://example.com/very/long/path/for/a/denser/preview",
		Level:   "H",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	prepared, err := NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	modules := prepared.PreviewModules()
	preview := prepared.PreviewFit(24, 10)
	if preview == "" {
		t.Fatal("PreviewFit() = empty string")
	}
	lines := strings.Split(preview, "\n")
	if got := maxPreviewWidth(lines); got != modules {
		t.Fatalf("PreviewFit() width = %d, want native module width %d", got, modules)
	}
	if got := len(lines); got != (modules+1)/2 {
		t.Fatalf("PreviewFit() height = %d, want native char height %d", got, (modules+1)/2)
	}
	if got := maxPreviewWidth(lines); got <= 24 {
		t.Fatalf("PreviewFit() width = %d, want wider than viewport when preserving native modules", got)
	}
}

func TestPreviewFitUsesExtraViewportSpaceWhenViewportIsLarge(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "upscale",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	prepared, err := NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	base := prepared.Preview()
	fitted := prepared.PreviewFit(60, 24)

	if maxPreviewWidth(strings.Split(fitted, "\n")) <= maxPreviewWidth(strings.Split(base, "\n")) {
		t.Fatalf("PreviewFit() width = %d, want larger than base width %d", maxPreviewWidth(strings.Split(fitted, "\n")), maxPreviewWidth(strings.Split(base, "\n")))
	}
	if len(strings.Split(fitted, "\n")) <= len(strings.Split(base, "\n")) {
		t.Fatalf("PreviewFit() height = %d, want larger than base height %d", len(strings.Split(fitted, "\n")), len(strings.Split(base, "\n")))
	}
}

func TestPreviewFitKeepsSquareGridForDenseContent(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "发的上课了放假快乐 sd 卡放假啦 sd 卡房间为埃及人 weakly 放假 ADSL 客服啊圣诞快乐发阿斯蒂芬看阿斯蒂芬阿斯蒂芬跨时代开放啦是的副卡就是的罚款了是的积分啊上看到了放假啊圣诞快乐发阿斯蒂芬克拉斯都发开始了地方啊圣诞快乐发阿斯蒂芬开了撒旦法是短发收到了客服阿斯蒂芬集卡老师的飞机 adslkffafffffffffffffffffjjlkje2kjeio2u09u 批发第三方届奥斯卡放假啊放假啊酸辣粉卡时间发快手发阿斯蒂芬撒旦法撒旦法是",
		Level:   "H",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	prepared, err := NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	modules := prepared.PreviewModules()
	preview := prepared.PreviewFit(96, 28)
	lines := strings.Split(preview, "\n")
	if got := maxPreviewWidth(lines); got != modules {
		t.Fatalf("PreviewFit() width = %d, want native module width %d", got, modules)
	}
	if got := len(lines); got != (modules+1)/2 {
		t.Fatalf("PreviewFit() height = %d, want native char height %d", got, (modules+1)/2)
	}
	if got := maxPreviewWidth(lines); got <= 56 {
		t.Fatalf("PreviewFit() width = %d, want wider than old lossy 56-module cap", got)
	}
}

func TestPreviewFitKeepsUniformModulesForDenseNearCapacity(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "fasdkfjasdlkfasdkfldajsfklajflkfaksd\nlfjadsklfjadsklfjadsklf\njadsklfjadsklfjadsklfjadsklfja\ndsfasdfhkjhkjjhkjhjhk",
		Level:   "Q",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	prepared, err := NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	modules := prepared.PreviewModules()
	capacity := modules + 3
	preview := prepared.PreviewFit(capacity, 200)
	lines := strings.Split(preview, "\n")
	if got := maxPreviewWidth(lines); got != modules {
		t.Fatalf("PreviewFit() width = %d, want native width %d for dense near-capacity preview", got, modules)
	}
}

func TestPreviewFitSmallViewportStillDecodesPayload(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "https://example.com/preview/decode/check?payload=abcdefghijklmnopqrstuvwxyz0123456789",
		Level:   "H",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	prepared, err := NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	preview := prepared.PreviewFit(24, 10)
	if preview == "" {
		t.Fatal("PreviewFit() = empty string")
	}

	img := previewImage(preview, 4)
	qrCodes, err := goqr.Recognize(img)
	if err != nil {
		t.Fatalf("goqr.Recognize(preview) error = %v", err)
	}
	if len(qrCodes) != 1 {
		t.Fatalf("Recognize(preview) len = %d, want 1", len(qrCodes))
	}
	if got := string(qrCodes[0].Payload); got != req.Content {
		t.Fatalf("Preview payload = %q, want %q", got, req.Content)
	}
}

func TestPrepareRasterUsesRequestedSize(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "https://example.com",
		Size:    "320",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	prepared, err := NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	raster := prepared.Raster()
	if raster == nil {
		t.Fatal("Raster() = nil")
	}
	if got := raster.Bounds().Dx(); got != 320 {
		t.Fatalf("Raster() width = %d, want 320", got)
	}
	if got := raster.Bounds().Dy(); got != 320 {
		t.Fatalf("Raster() height = %d, want 320", got)
	}
}

func TestPrepareRejectsSizesThatCannotRenderEveryModule(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "hello",
		Size:    "1",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	_, err = NewEngine().Prepare(req)
	if err == nil {
		t.Fatal("Prepare() error = nil, want size validation error")
	}

	var userErr *core.UserError
	if !errors.As(err, &userErr) {
		t.Fatalf("Prepare() error = %T, want *core.UserError", err)
	}
	if userErr.Kind != core.ErrorSizeTooSmall {
		t.Fatalf("Prepare() error kind = %q, want %q", userErr.Kind, core.ErrorSizeTooSmall)
	}
}

func TestRequiredModulesMatchesPreparedPreviewModules(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "https://example.com/required-modules",
		Level:   "Q",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	prepared, err := NewEngine().Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	modules, err := RequiredModules(req.Content, req.Level)
	if err != nil {
		t.Fatalf("RequiredModules() error = %v", err)
	}
	if modules != prepared.PreviewModules() {
		t.Fatalf("RequiredModules() = %d, want %d", modules, prepared.PreviewModules())
	}
}

func TestEngineReusesPreparedMatrixForSameVisualRequest(t *testing.T) {
	base, err := core.Normalize(core.Request{
		Content: "same",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	otherPath := base
	otherPath.OutputPath = "/tmp/other.png"

	engine := NewEngine()
	first, err := engine.Prepare(base)
	if err != nil {
		t.Fatalf("Prepare(first) error = %v", err)
	}

	second, err := engine.Prepare(otherPath)
	if err != nil {
		t.Fatalf("Prepare(second) error = %v", err)
	}

	if first != second {
		t.Fatal("Prepare() returned different prepared values for same visual request")
	}
}

func TestWriteToPathUsesAtomicFileReplace(t *testing.T) {
	req, err := core.Normalize(core.Request{
		Content: "atomic",
		Source:  core.SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	dir := t.TempDir()
	target := filepath.Join(dir, "atomic.png")
	req.OutputPath = target

	engine := NewEngine()
	prepared, err := engine.Prepare(req)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	if err := prepared.WriteToPath(target); err != nil {
		t.Fatalf("WriteToPath() error = %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("written file is empty")
	}

	matches, err := filepath.Glob(filepath.Join(dir, ".tmp-qrcode-*"))
	if err != nil {
		t.Fatalf("filepath.Glob() error = %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary files left behind: %v", matches)
	}
}

func maxPreviewWidth(lines []string) int {
	result := 0
	for _, line := range lines {
		if got := len([]rune(line)); got > result {
			result = got
		}
	}
	return result
}

func previewImage(preview string, modulePixels int) image.Image {
	lines := strings.Split(preview, "\n")
	width := maxPreviewWidth(lines)
	if modulePixels <= 0 {
		modulePixels = 1
	}
	moduleHeight := len(lines) * 2

	img := image.NewRGBA(image.Rect(0, 0, width*modulePixels, moduleHeight*modulePixels))
	fill(img, color.White)

	for y, line := range lines {
		row := []rune(line)
		for x := 0; x < width; x++ {
			cell := ' '
			if x < len(row) {
				cell = row[x]
			}

			top, bottom := false, false
			switch cell {
			case '█':
				top, bottom = true, true
			case '▀':
				top = true
			case '▄':
				bottom = true
			}

			if top {
				paintPreviewModule(img, x, y*2, modulePixels)
			}
			if bottom {
				paintPreviewModule(img, x, y*2+1, modulePixels)
			}
		}
	}

	return img
}

func paintPreviewModule(img *image.RGBA, moduleX, moduleY, modulePixels int) {
	startX := moduleX * modulePixels
	startY := moduleY * modulePixels
	for y := startY; y < startY+modulePixels; y++ {
		for x := startX; x < startX+modulePixels; x++ {
			img.Set(x, y, color.Black)
		}
	}
}
