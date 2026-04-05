package core

import (
	"errors"
	"testing"
)

func TestNormalizeAppliesHumaneDefaults(t *testing.T) {
	req, err := Normalize(Request{
		Content: "https://example.com",
		Source:  SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	if req.Content != "https://example.com" {
		t.Fatalf("Content = %q, want exact input", req.Content)
	}
	if req.Format != FormatPNG {
		t.Fatalf("Format = %q, want %q", req.Format, FormatPNG)
	}
	if req.Size != 256 {
		t.Fatalf("Size = %d, want 256", req.Size)
	}
	if req.Level != LevelMedium {
		t.Fatalf("Level = %q, want %q", req.Level, LevelMedium)
	}
	if req.OutputPath != "./qrcode.png" {
		t.Fatalf("OutputPath = %q, want ./qrcode.png", req.OutputPath)
	}
}

func TestNormalizeDerivesOutputPathFromFormat(t *testing.T) {
	req, err := Normalize(Request{
		Content: "https://example.com",
		Format:  "svg",
		Source:  SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	if req.Format != FormatSVG {
		t.Fatalf("Format = %q, want %q", req.Format, FormatSVG)
	}
	if req.OutputPath != "./qrcode.svg" {
		t.Fatalf("OutputPath = %q, want ./qrcode.svg", req.OutputPath)
	}
}

func TestNormalizeInfersFormatFromOutputPath(t *testing.T) {
	req, err := Normalize(Request{
		Content:    "https://example.com",
		OutputPath: "/tmp/custom.SVG",
		Source:     SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	if req.Format != FormatSVG {
		t.Fatalf("Format = %q, want %q", req.Format, FormatSVG)
	}
	if req.OutputPath != "/tmp/custom.SVG" {
		t.Fatalf("OutputPath = %q, want explicit path preserved", req.OutputPath)
	}
}

func TestNormalizePreservesExplicitOutputPath(t *testing.T) {
	req, err := Normalize(Request{
		Content:    "https://example.com",
		Format:     "svg",
		OutputPath: "/tmp/custom.svg",
		Source:     SourceCLIArg,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	if req.OutputPath != "/tmp/custom.svg" {
		t.Fatalf("OutputPath = %q, want explicit path", req.OutputPath)
	}
}

func TestNormalizeRejectsMismatchedFormatAndOutputPath(t *testing.T) {
	_, err := Normalize(Request{
		Content:    "https://example.com",
		Format:     "png",
		OutputPath: "/tmp/custom.svg",
		Source:     SourceCLIArg,
	})
	if err == nil {
		t.Fatal("Normalize() error = nil, want mismatch error")
	}

	var userErr *UserError
	if !errors.As(err, &userErr) {
		t.Fatalf("error = %T, want *UserError", err)
	}
	if userErr.Kind != ErrorFormatMismatch {
		t.Fatalf("Kind = %q, want %q", userErr.Kind, ErrorFormatMismatch)
	}
}

func TestNormalizeRejectsUnsupportedOutputExtension(t *testing.T) {
	_, err := Normalize(Request{
		Content:    "https://example.com",
		OutputPath: "/tmp/custom.txt",
		Source:     SourceCLIArg,
	})
	if err == nil {
		t.Fatal("Normalize() error = nil, want extension error")
	}

	var userErr *UserError
	if !errors.As(err, &userErr) {
		t.Fatalf("error = %T, want *UserError", err)
	}
	if userErr.Kind != ErrorInvalidOutputExtension {
		t.Fatalf("Kind = %q, want %q", userErr.Kind, ErrorInvalidOutputExtension)
	}
}

func TestNormalizeTrimsTrailingNewlineFromStdin(t *testing.T) {
	req, err := Normalize(Request{
		Content: "hello from pipe\n",
		Source:  SourceStdin,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	if req.Content != "hello from pipe" {
		t.Fatalf("Content = %q, want trailing newline trimmed", req.Content)
	}
}

func TestNormalizePreservesMultilineContentForTUI(t *testing.T) {
	input := "hello\nworld"
	req, err := Normalize(Request{
		Content: input,
		Source:  SourceTUI,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	if req.Content != input {
		t.Fatalf("Content = %q, want %q", req.Content, input)
	}
}

func TestNormalizeRejectsWhitespaceOnlyContent(t *testing.T) {
	_, err := Normalize(Request{
		Content: "   \n\t  ",
		Source:  SourceCLIArg,
	})
	if err == nil {
		t.Fatal("Normalize() error = nil, want error")
	}

	var userErr *UserError
	if !errors.As(err, &userErr) {
		t.Fatalf("error = %T, want *UserError", err)
	}
	if userErr.Kind != ErrorEmptyContent {
		t.Fatalf("Kind = %q, want %q", userErr.Kind, ErrorEmptyContent)
	}
}

func TestNormalizeRejectsInvalidSize(t *testing.T) {
	_, err := Normalize(Request{
		Content: "https://example.com",
		Size:    "100x200",
		Source:  SourceCLIArg,
	})
	if err == nil {
		t.Fatal("Normalize() error = nil, want error")
	}

	var userErr *UserError
	if !errors.As(err, &userErr) {
		t.Fatalf("error = %T, want *UserError", err)
	}
	if userErr.Kind != ErrorInvalidSize {
		t.Fatalf("Kind = %q, want %q", userErr.Kind, ErrorInvalidSize)
	}
}

func TestUserError(t *testing.T) {
	t.Run("nil_error", func(t *testing.T) {
		var err *UserError
		if got := err.Error(); got != "" {
			t.Fatalf("Error() = %q, want empty string", got)
		}
	})

	t.Run("without_cause", func(t *testing.T) {
		err := &UserError{
			Kind:    ErrorEmptyContent,
			Message: "content is required",
		}
		if got := err.Error(); got != "content is required" {
			t.Fatalf("Error() = %q, want %q", got, "content is required")
		}
	})

	t.Run("with_cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &UserError{
			Kind:    ErrorInvalidSize,
			Message: "size is invalid",
			Cause:   cause,
		}
		want := "size is invalid: underlying error"
		if got := err.Error(); got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}

func TestUserErrorUnwrap(t *testing.T) {
	t.Run("nil_error", func(t *testing.T) {
		var err *UserError
		if got := err.Unwrap(); got != nil {
			t.Fatalf("Unwrap() = %v, want nil", got)
		}
	})

	t.Run("without_cause", func(t *testing.T) {
		err := &UserError{
			Kind:    ErrorEmptyContent,
			Message: "content is required",
		}
		if got := err.Unwrap(); got != nil {
			t.Fatalf("Unwrap() = %v, want nil", got)
		}
	})

	t.Run("with_cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &UserError{
			Kind:    ErrorInvalidSize,
			Message: "size is invalid",
			Cause:   cause,
		}
		unwrapped := err.Unwrap()
		if unwrapped != cause {
			t.Fatalf("Unwrap() = %v, want %v", unwrapped, cause)
		}
	})
}

func TestNormalizeSize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "empty_uses_default", input: "", want: 256, wantErr: false},
		{name: "single_number", input: "512", want: 512, wantErr: false},
		{name: "square_format", input: "256x256", want: 256, wantErr: false},
		{name: "square_format_different_values", input: "100x100", want: 100, wantErr: false},
		{name: "uppercase_x", input: "256X256", want: 256, wantErr: false},
		{name: "with_spaces", input: " 256 ", want: 256, wantErr: false},
		{name: "rejects_rectangle", input: "100x200", want: 0, wantErr: true},
		{name: "rejects_negative", input: "-100", want: 0, wantErr: true},
		{name: "rejects_zero", input: "0", want: 0, wantErr: true},
		{name: "rejects_invalid", input: "abc", want: 0, wantErr: true},
		{name: "rejects_too_many_parts", input: "100x200x300", want: 0, wantErr: true},
		{name: "rejects_negative_in_square", input: "-100x-100", want: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := Normalize(Request{
				Content: "test",
				Size:    tt.input,
				Source:  SourceCLIArg,
			})
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Normalize() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			if req.Size != tt.want {
				t.Fatalf("Size = %d, want %d", req.Size, tt.want)
			}
		})
	}
}

func TestNormalizeLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Level
		wantErr bool
	}{
		{name: "empty_uses_default", input: "", want: LevelMedium, wantErr: false},
		{name: "low", input: "L", want: LevelLow, wantErr: false},
		{name: "medium", input: "M", want: LevelMedium, wantErr: false},
		{name: "quartile", input: "Q", want: LevelQuart, wantErr: false},
		{name: "high", input: "H", want: LevelHigh, wantErr: false},
		{name: "lowercase", input: "l", want: LevelLow, wantErr: false},
		{name: "mixed_case", input: "m", want: LevelMedium, wantErr: false},
		{name: "with_spaces", input: " Q ", want: LevelQuart, wantErr: false},
		{name: "rejects_invalid", input: "X", want: "", wantErr: true},
		{name: "rejects_number", input: "1", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := Normalize(Request{
				Content: "test",
				Level:   tt.input,
				Source:  SourceCLIArg,
			})
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Normalize() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			if req.Level != tt.want {
				t.Fatalf("Level = %q, want %q", req.Level, tt.want)
			}
		})
	}
}

func TestFormatFromPath(t *testing.T) {
	tests := []struct {
		name       string
		outputPath string
		wantPNG    bool
		wantSVG    bool
	}{
		{name: "png_extension", outputPath: "file.png", wantPNG: true},
		{name: "svg_extension", outputPath: "file.svg", wantSVG: true},
		{name: "uppercase_png", outputPath: "file.PNG", wantPNG: true},
		{name: "uppercase_svg", outputPath: "file.SVG", wantSVG: true},
		{name: "no_extension", outputPath: "file", wantPNG: false, wantSVG: false},
		{name: "other_extension", outputPath: "file.txt", wantPNG: false, wantSVG: false},
		{name: "with_path_png", outputPath: "/path/to/file.png", wantPNG: true},
		{name: "with_path_svg", outputPath: "/path/to/file.svg", wantSVG: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := Normalize(Request{
				Content:    "test",
				OutputPath: tt.outputPath,
				Source:     SourceCLIArg,
			})
			if tt.wantPNG || tt.wantSVG {
				if err != nil {
					t.Fatalf("Normalize() error = %v", err)
				}
				if tt.wantPNG && req.Format != FormatPNG {
					t.Fatalf("Format = %q, want PNG", req.Format)
				}
				if tt.wantSVG && req.Format != FormatSVG {
					t.Fatalf("Format = %q, want SVG", req.Format)
				}
			}
		})
	}
}

func TestNormalizeFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Format
		wantErr bool
	}{
		{name: "png_lowercase", input: "png", want: FormatPNG, wantErr: false},
		{name: "png_uppercase", input: "PNG", want: FormatPNG, wantErr: false},
		{name: "svg_lowercase", input: "svg", want: FormatSVG, wantErr: false},
		{name: "svg_uppercase", input: "SVG", want: FormatSVG, wantErr: false},
		{name: "with_spaces", input: " png ", want: FormatPNG, wantErr: false},
		{name: "rejects_invalid", input: "gif", want: "", wantErr: true},
		{name: "rejects_jpeg", input: "jpeg", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := Normalize(Request{
				Content: "test",
				Format:  tt.input,
				Source:  SourceCLIArg,
			})
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Normalize() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			if req.Format != tt.want {
				t.Fatalf("Format = %q, want %q", req.Format, tt.want)
			}
		})
	}
}
