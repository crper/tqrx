package core

import "testing"

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
	if !AsUserError(err, &userErr) {
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
	if !AsUserError(err, &userErr) {
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
	if !AsUserError(err, &userErr) {
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
	if !AsUserError(err, &userErr) {
		t.Fatalf("error = %T, want *UserError", err)
	}
	if userErr.Kind != ErrorInvalidSize {
		t.Fatalf("Kind = %q, want %q", userErr.Kind, ErrorInvalidSize)
	}
}
