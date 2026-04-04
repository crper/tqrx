package render

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/yeqown/go-qrcode/v2"

	"github.com/crper/tqrx/internal/core"
)

// quietZoneModules 定义预览和导出图像共用的留白模块数，保证视觉一致且易
// 于扫码。
const quietZoneModules = 4

// densePreviewModuleThreshold 表示在预览里需要优先保证模块等宽等高的密度阈值。
// 当二维码模块接近终端容量上限时，轻微非整数缩放也会增加扫码失败风险。
const densePreviewModuleThreshold = 49

// Engine 根据标准化请求准备可复用的渲染产物，并缓存最近一次结果，供预览
// 和保存流程复用。
type Engine struct {
	lastKey      string
	lastPrepared *Prepared
}

// Prepared 表示 PNG、SVG、终端预览和文件导出共用的渲染结果，让应用只有
// 一套视觉真相来源。
type Prepared struct {
	request core.NormalizedRequest
	bitmap  [][]bool
	preview string
	raster  image.Image
}

// NewEngine 返回一个带空单项缓存的渲染器。
func NewEngine() *Engine {
	return &Engine{}
}

// Prepare 只构建一次二维码位图，然后让所有输出模式都从同一份位图渲染。
// 外部库只负责生成矩阵，最终绘制保留在本包内，以避免预览和导出漂移。
func (e *Engine) Prepare(req core.NormalizedRequest) (*Prepared, error) {
	key := renderKey(req)
	if e.lastPrepared != nil && e.lastKey == key {
		return e.lastPrepared, nil
	}

	bitmap, err := bitmapFor(req.Content, req.Level)
	if err != nil {
		return nil, err
	}
	if err := validateSize(req, bitmap); err != nil {
		return nil, err
	}

	prepared := &Prepared{
		request: req,
		bitmap:  bitmap,
	}
	e.lastKey = key
	e.lastPrepared = prepared
	return prepared, nil
}

// PNG 把共享位图栅格化为指定尺寸的正方形图像。
func (p *Prepared) PNG() ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, p.Raster()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SVG 把共享位图渲染成清晰的矩形模块，确保矢量导出和 PNG 使用同一套几何
// 规则。
func (p *Prepared) SVG() ([]byte, error) {
	total := p.totalModules()
	activeModules := p.countActiveModules()
	estimatedSize := estimateSVGSize(p.request.Size, activeModules)

	var buf strings.Builder
	buf.Grow(estimatedSize)
	buf.WriteString(xml.Header)

	fmt.Fprintf(&buf,
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" shape-rendering="crispEdges">`,
		p.request.Size, p.request.Size, p.request.Size, p.request.Size)
	fmt.Fprintf(&buf, `<rect width="%d" height="%d" fill="#ffffff"/>`, p.request.Size, p.request.Size)

	for y := 0; y < total; y++ {
		startY, endY := band(y, p.request.Size, total)
		for x := 0; x < total; x++ {
			if !p.moduleActive(y, x) {
				continue
			}
			startX, endX := band(x, p.request.Size, total)
			fmt.Fprintf(&buf,
				`<rect x="%d" y="%d" width="%d" height="%d" fill="#000000"/>`,
				startX, startY, endX-startX, endY-startY)
		}
	}

	buf.WriteString(`</svg>`)
	return []byte(buf.String()), nil
}

func (p *Prepared) countActiveModules() int {
	if len(p.bitmap) == 0 {
		return 0
	}
	count := 0
	for _, row := range p.bitmap {
		for _, active := range row {
			if active {
				count++
			}
		}
	}
	return count
}

func (p *Prepared) moduleActive(y, x int) bool {
	srcY, srcX := y-quietZoneModules, x-quietZoneModules
	return srcY >= 0 && srcY < len(p.bitmap) &&
		srcX >= 0 && srcX < len(p.bitmap[srcY]) &&
		p.bitmap[srcY][srcX]
}

func estimateSVGSize(imageSize, activeModules int) int {
	const svgHeaderSize = 256
	const rectElementSize = 64
	return svgHeaderSize + activeModules*rectElementSize
}

// Preview 构建适合终端显示的块状预览，并缓存结果字符串，因为 TUI 可能反复
// 请求它。
func (p *Prepared) Preview() string {
	if p.preview != "" {
		return p.preview
	}

	total := p.totalModules()
	p.preview = p.renderHalfBlockPreview(total, total)
	return p.preview
}

// PreviewModules 返回包含 quiet zone 在内的预览模块边长，用于让上层判断
// 当前终端网格是否足够完整呈现二维码。
func (p *Prepared) PreviewModules() int {
	return p.totalModules()
}

// PreviewFit 会按当前终端预览画布放大二维码矩阵；当画布不足时保持原始模块
// 分辨率，不做有损压缩，避免预览看起来完整但实际不可扫码。
func (p *Prepared) PreviewFit(maxWidth, maxHeight int) string {
	if maxWidth <= 0 || maxHeight <= 0 {
		return ""
	}

	total := p.totalModules()
	capacity := max(1, min(maxWidth, maxHeight*2))
	var target int
	switch {
	case capacity < total:
		// 画布不够时保持原始模块，交给上层滚动，不做有损压缩。
		target = total
	default:
		scale := capacity / total
		switch {
		case scale >= 2:
			// 只用整数倍放大，保证模块几何一致。
			target = total * scale
		case total >= densePreviewModuleThreshold:
			// 高密度且只能 1x 时，避免非整数放大导致模块轻微形变。
			target = total
		default:
			// 低密度场景可适度拉满画布，提升可读性。
			target = capacity
		}
	}
	if target == total {
		return p.Preview()
	}
	return p.renderHalfBlockPreview(target, target)
}

// Raster 返回和最终导出一致的二维码栅格图像，供 PNG 导出与其他预览路径共用。
func (p *Prepared) Raster() image.Image {
	if p.raster != nil {
		return p.raster
	}

	img := image.NewRGBA(image.Rect(0, 0, p.request.Size, p.request.Size))
	fill(img, color.White)

	total := p.totalModules()
	for y := 0; y < total; y++ {
		startY, endY := band(y, p.request.Size, total)
		for x := 0; x < total; x++ {
			if !p.moduleActive(y, x) {
				continue
			}
			startX, endX := band(x, p.request.Size, total)
			for py := startY; py < endY; py++ {
				for px := startX; px < endX; px++ {
					img.Set(px, py, color.Black)
				}
			}
		}
	}

	p.raster = img
	return p.raster
}

// WriteToPath 会先写入临时文件，确保保存失败时不会留下半写入的图像文件。
func (p *Prepared) WriteToPath(path string) error {
	data, err := p.bytesForFormat(p.request.Format)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-qrcode-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, path)
}

// bytesForFormat 是最终的格式分发点。在此之前的所有流程都共享同一份规范化
// 请求和已准备好的位图。
func (p *Prepared) bytesForFormat(format core.Format) ([]byte, error) {
	switch format {
	case core.FormatSVG:
		return p.SVG()
	case core.FormatPNG:
		return p.PNG()
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (p *Prepared) previewModule(y, x int) bool {
	return p.moduleActive(y, x)
}

func (p *Prepared) renderHalfBlockPreview(targetWidth, targetModulesHeight int) string {
	grid := p.previewGrid(targetWidth, targetModulesHeight)
	charHeight := (targetModulesHeight + 1) / 2
	lines := make([]string, 0, charHeight)
	for y := 0; y < charHeight; y++ {
		var b strings.Builder
		for x := 0; x < targetWidth; x++ {
			top := previewGridModule(grid, x, y*2)
			bottom := previewGridModule(grid, x, y*2+1)
			switch {
			case top && bottom:
				b.WriteRune('█')
			case top:
				b.WriteRune('▀')
			case bottom:
				b.WriteRune('▄')
			default:
				b.WriteRune(' ')
			}
		}
		lines = append(lines, b.String())
	}

	return strings.Join(lines, "\n")
}

// previewGrid 会先把二维码模块映射到一个按目标尺寸放大的布尔网格，再由
// renderHalfBlockPreview 把两行模块折叠成一个终端字符。
func (p *Prepared) previewGrid(targetWidth, targetHeight int) [][]bool {
	grid := make([][]bool, targetHeight)
	for y := range grid {
		grid[y] = make([]bool, targetWidth)
	}

	total := p.totalModules()
	for srcY := 0; srcY < total; srcY++ {
		startY, endY := band(srcY, targetHeight, total)
		if startY == endY {
			continue
		}
		for srcX := 0; srcX < total; srcX++ {
			if !p.previewModule(srcY, srcX) {
				continue
			}
			startX, endX := band(srcX, targetWidth, total)
			if startX == endX {
				continue
			}
			for y := startY; y < endY; y++ {
				for x := startX; x < endX; x++ {
					grid[y][x] = true
				}
			}
		}
	}

	return grid
}

// previewGridModule 为越界访问提供统一的“空白模块”语义，避免预览绘制在边
// 缘判断上分散出很多 if。
func previewGridModule(grid [][]bool, x, y int) bool {
	if y < 0 || y >= len(grid) {
		return false
	}
	if x < 0 || x >= len(grid[y]) {
		return false
	}
	return grid[y][x]
}

// renderKey 会刻意忽略输出路径，因为把同一个二维码写到不同文件时仍应复用
// 同一份已准备好的渲染结果。
func renderKey(req core.NormalizedRequest) string {
	return strings.Join([]string{
		req.Content,
		string(req.Format),
		fmt.Sprintf("%d", req.Size),
		string(req.Level),
	}, "\x00")
}

// band 把模块索引映射为像素坐标。整数除法可以在尺寸不能整除时仍保持整张图
// 的总尺寸精确。
func band(index, size, total int) (int, int) {
	start := index * size / total
	end := (index + 1) * size / total
	return start, end
}

func (p *Prepared) totalModules() int {
	return bitmapModules(p.bitmap)
}

func bitmapModules(bitmap [][]bool) int {
	return len(bitmap) + quietZoneModules*2
}

// fill 会在绘制深色模块之前初始化 PNG 画布背景。
func fill(img *image.RGBA, c color.Color) {
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			img.Set(x, y, c)
		}
	}
}

// toECOption 把面向用户的纠错等级映射为二维码库使用的枚举值。
func toECOption(level core.Level) qrcode.EncodeOption {
	switch level {
	case core.LevelLow:
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionLow)
	case core.LevelQuart:
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionQuart)
	case core.LevelHigh:
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionHighest)
	default:
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium)
	}
}

// RequiredModules 返回指定内容和纠错等级在预览中的模块边长（包含 quiet zone）。
// 这个值可用于判断当前终端网格是否能完整显示并扫码。
func RequiredModules(content string, level core.Level) (int, error) {
	bitmap, err := bitmapFor(content, level)
	if err != nil {
		return 0, err
	}
	return bitmapModules(bitmap), nil
}

// validateSize 在真正开始导出前拒绝过小尺寸，避免生成视觉上存在但实际无
// 法扫码的图像。
func validateSize(req core.NormalizedRequest, bitmap [][]bool) error {
	minSize := bitmapModules(bitmap)
	if req.Size >= minSize {
		return nil
	}

	return &core.UserError{
		Kind:    core.ErrorSizeTooSmall,
		Message: fmt.Sprintf("size must be at least %d for this content", minSize),
	}
}

func bitmapFor(content string, level core.Level) ([][]bool, error) {
	qr, err := qrcode.NewWith(content, toECOption(level))
	if err != nil {
		return nil, err
	}

	capture := &captureWriter{}
	if err := qr.Save(capture); err != nil {
		return nil, err
	}
	return capture.bitmap, nil
}

// captureWriter 提取二维码库生成的位图，让应用的其余部分掌握实际渲染路径。
type captureWriter struct {
	bitmap [][]bool
}

// Write 只抓取底层库生成的布尔矩阵，把真正的绘制权留在本包里。
func (w *captureWriter) Write(mat qrcode.Matrix) error {
	w.bitmap = mat.Bitmap()
	return nil
}

// Close 满足底层 writer 接口；captureWriter 本身没有需要释放的资源。
func (w *captureWriter) Close() error {
	return nil
}
