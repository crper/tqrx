package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// Source 表示内容来源，用于让标准化过程区分 CLI 参数、stdin 和交互式
// 编辑这几种输入路径。
type Source string

const (
	// SourceCLIArg 表示内容来自根命令的位置参数。
	SourceCLIArg Source = "cli_arg"
	// SourceStdin 表示内容来自管道标准输入。
	SourceStdin Source = "stdin"
	// SourceTUI 表示内容来自交互式编辑器。
	SourceTUI Source = "tui"
)

// Format 表示支持的导出格式。
type Format string

const (
	// FormatPNG 表示导出为位图文件。
	FormatPNG Format = "png"
	// FormatSVG 表示导出为矢量文件。
	FormatSVG Format = "svg"
)

// Level 表示面向用户暴露的二维码纠错等级。
type Level string

const (
	// LevelLow 表示最低纠错等级。
	LevelLow Level = "L"
	// LevelMedium 表示默认纠错等级。
	LevelMedium Level = "M"
	// LevelQuart 表示四分位纠错等级。
	LevelQuart Level = "Q"
	// LevelHigh 表示最高纠错等级。
	LevelHigh Level = "H"
)

// ErrorKind 对校验失败进行分类，让 CLI 和 TUI 可以基于稳定的机器可读原
// 因分支处理，而不是匹配原始字符串。
type ErrorKind string

const (
	// ErrorEmptyContent 表示没有提供可用的二维码内容。
	ErrorEmptyContent ErrorKind = "empty_content"
	// ErrorInvalidFormat 表示请求的文件格式不受支持。
	ErrorInvalidFormat ErrorKind = "invalid_format"
	// ErrorInvalidOutputExtension 表示输出路径使用了不受支持的后缀。
	ErrorInvalidOutputExtension ErrorKind = "invalid_output_extension"
	// ErrorFormatMismatch 表示输出路径扩展名和所选文件格式冲突。
	ErrorFormatMismatch ErrorKind = "format_mismatch"
	// ErrorInvalidSize 表示请求的输出尺寸不合法。
	ErrorInvalidSize ErrorKind = "invalid_size"
	// ErrorSizeTooSmall 表示请求的尺寸过小，无法容纳二维码的全部模块。
	ErrorSizeTooSmall ErrorKind = "size_too_small"
	// ErrorInvalidLevel 表示请求的纠错等级不合法。
	ErrorInvalidLevel ErrorKind = "invalid_level"
)

// UserError 封装用户可以自行修复的问题，包含可读消息和可选的底层原因。
type UserError struct {
	Kind    ErrorKind
	Message string
	Cause   error
}

// Error 返回面向用户的可读消息，并在存在底层原因时一并保留。
func (e *UserError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

// Unwrap 让调用方可以对 UserError 使用 errors.Is 和 errors.As。
func (e *UserError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// AsUserError 是一个小工具函数，用于提取面向用户的校验错误。
func AsUserError(err error, target **UserError) bool {
	return errors.As(err, target)
}

// Request 表示在默认值填充和校验之前，从 CLI 标志或 TUI 控件收集到的原始
// 输入。
type Request struct {
	// Content 是待编码的原始内容。
	Content string
	// Format 是原始输出格式字符串，例如 "png" 或 "svg"。
	Format string
	// Size 是原始正方形尺寸字符串，例如 "256" 或 "256x256"。
	Size string
	// OutputPath 是填充默认值之前的目标文件路径。
	OutputPath string
	// Level 是原始纠错等级字符串。
	Level string
	// Source 记录请求来源。
	Source Source
}

// NormalizedRequest 表示预览和导出共用的规范化请求结构，确保所有入口行为
// 一致。
type NormalizedRequest struct {
	// Content 是标准化后的待编码内容。
	Content string
	// Format 是校验后的输出格式。
	Format Format
	// Size 是校验后的正方形输出尺寸，单位为像素。
	Size int
	// OutputPath 是最终目标文件路径。
	OutputPath string
	// Level 是校验后的二维码纠错等级。
	Level Level
	// Source 记录请求来源。
	Source Source
}

// Normalize 会填充默认值、裁剪特定来源的噪声，并把原始字符串转换为其余模
// 块使用的稳定请求结构。
func Normalize(req Request) (NormalizedRequest, error) {
	content := normalizeContent(req.Content, req.Source)
	if strings.TrimSpace(content) == "" {
		return NormalizedRequest{}, &UserError{
			Kind:    ErrorEmptyContent,
			Message: "content is required",
		}
	}

	format, formatSet, err := parseFormat(req.Format)
	if err != nil {
		return NormalizedRequest{}, err
	}

	inferred, inferredFromPath := formatFromPath(req.OutputPath)
	if ext := outputPathExt(req.OutputPath); ext != "" && !inferredFromPath {
		return NormalizedRequest{}, &UserError{
			Kind:    ErrorInvalidOutputExtension,
			Message: fmt.Sprintf("output path extension %q must be .png or .svg", ext),
		}
	}
	if !formatSet {
		if inferredFromPath {
			format = inferred
		} else {
			format = FormatPNG
		}
	}
	if inferredFromPath && inferred != format {
		return NormalizedRequest{}, &UserError{
			Kind: ErrorFormatMismatch,
			Message: fmt.Sprintf(
				"output path extension %q must match format %q",
				filepath.Ext(req.OutputPath),
				format,
			),
		}
	}

	size, err := normalizeSize(req.Size)
	if err != nil {
		return NormalizedRequest{}, err
	}

	level, err := normalizeLevel(req.Level)
	if err != nil {
		return NormalizedRequest{}, err
	}

	outputPath := req.OutputPath
	if outputPath == "" {
		outputPath = "./qrcode." + string(format)
	}

	return NormalizedRequest{
		Content:    content,
		Format:     format,
		Size:       size,
		OutputPath: outputPath,
		Level:      level,
		Source:     req.Source,
	}, nil
}

// normalizeContent 只会为 stdin 去掉末尾换行，从而保留 CLI 参数和 TUI
// 编辑器里有意输入的换行。
func normalizeContent(content string, source Source) string {
	if source == SourceStdin {
		return strings.TrimRight(content, "\r\n")
	}
	return content
}

// parseFormat 用一套解析逻辑对齐 CLI 和 TUI 的输出规则，同时保留在未显
// 式指定格式时由 Normalize 根据输出路径推导格式的能力。
func parseFormat(raw string) (Format, bool, error) {
	if raw == "" {
		return "", false, nil
	}

	format := Format(strings.ToLower(strings.TrimSpace(raw)))
	switch format {
	case FormatPNG, FormatSVG:
		return format, true, nil
	default:
		return "", false, &UserError{
			Kind:    ErrorInvalidFormat,
			Message: "format must be png or svg",
		}
	}
}

func formatFromPath(path string) (Format, bool) {
	switch outputPathExt(path) {
	case ".png":
		return FormatPNG, true
	case ".svg":
		return FormatSVG, true
	default:
		return "", false
	}
}

// outputPathExt 会先裁掉用户输入里常见的首尾空白，再统一转成小写后缀，避
// 免 CLI 和 TUI 在扩展名判断上出现细微分叉。
func outputPathExt(path string) string {
	return strings.ToLower(filepath.Ext(strings.TrimSpace(path)))
}

// normalizeSize 接受 "256" 或 "256x256" 两种形式，并刻意拒绝矩形尺寸，
// 因为当前渲染器只暴露正方形输出。
func normalizeSize(raw string) (int, error) {
	if raw == "" {
		return 256, nil
	}

	parts := strings.Split(strings.ToLower(strings.TrimSpace(raw)), "x")
	switch len(parts) {
	case 1:
		size, err := strconv.Atoi(parts[0])
		if err != nil || size <= 0 {
			return 0, &UserError{
				Kind:    ErrorInvalidSize,
				Message: "size must be a positive integer or square dimension",
				Cause:   err,
			}
		}
		return size, nil
	case 2:
		width, errW := strconv.Atoi(parts[0])
		height, errH := strconv.Atoi(parts[1])
		if errW != nil || errH != nil || width <= 0 || height <= 0 || width != height {
			return 0, &UserError{
				Kind:    ErrorInvalidSize,
				Message: "size must be square, like 256 or 256x256",
			}
		}
		return width, nil
	default:
		return 0, &UserError{
			Kind:    ErrorInvalidSize,
			Message: "size must be square, like 256 or 256x256",
		}
	}
}

// normalizeLevel 解析二维码纠错等级，并让内部表示尽量贴近用户在界面里看
// 到的值。
func normalizeLevel(raw string) (Level, error) {
	if raw == "" {
		return LevelMedium, nil
	}

	level := Level(strings.ToUpper(strings.TrimSpace(raw)))
	switch level {
	case LevelLow, LevelMedium, LevelQuart, LevelHigh:
		return level, nil
	default:
		return "", &UserError{
			Kind:    ErrorInvalidLevel,
			Message: "level must be one of L, M, Q, H",
		}
	}
}
