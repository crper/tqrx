package cli

import (
	"bytes"
	"errors"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liyue201/goqr"

	"github.com/crper/tqrx/internal/core"
)

func TestRunGeneratesDefaultPNGFromArgument(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	var stdout bytes.Buffer
	runner := NewRunner()

	if err := runner.Run([]string{"https://example.com"}, strings.NewReader(""), &stdout, &stdout); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	target := filepath.Join(dir, "qrcode.png")
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected %s to exist: %v", target, err)
	}
	if !strings.Contains(stdout.String(), "Saved to ./qrcode.png") {
		t.Fatalf("stdout = %q, want save confirmation", stdout.String())
	}
}

func TestRunFallsBackToStdinWhenNoArgument(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	var stdout bytes.Buffer
	runner := NewRunner()

	if err := runner.Run(nil, strings.NewReader("from pipe\n"), &stdout, &stdout); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	target := filepath.Join(dir, "qrcode.png")
	got := decodePNGPayload(t, target)
	if got != "from pipe" {
		t.Fatalf("payload = %q, want %q", got, "from pipe")
	}
}

func TestRunInfersFormatFromOutputExtension(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "custom.svg")

	var stdout bytes.Buffer
	runner := NewRunner()

	if err := runner.Run([]string{"https://example.com", "-o", target}, strings.NewReader(""), &stdout, &stdout); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", target, err)
	}
	if !strings.Contains(string(data), "<svg") {
		t.Fatalf("output %q = %q, want SVG data", target, string(data))
	}
}

func TestRunArgumentWinsOverStdin(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	var stdout bytes.Buffer
	runner := NewRunner()

	if err := runner.Run([]string{"from arg"}, strings.NewReader("from pipe\n"), &stdout, &stdout); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	target := filepath.Join(dir, "qrcode.png")
	got := decodePNGPayload(t, target)
	if got != "from arg" {
		t.Fatalf("payload = %q, want %q", got, "from arg")
	}
}

func TestRunRejectsMismatchedFormatAndOutputExtension(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "custom.svg")

	runner := NewRunner()
	err := runner.Run([]string{"https://example.com", "-f", "png", "-o", target}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Run() error = nil, want mismatch error")
	}

	var userErr *core.UserError
	if !errors.As(err, &userErr) {
		t.Fatalf("Run() error = %T, want *core.UserError", err)
	}
	if userErr.Kind != core.ErrorFormatMismatch {
		t.Fatalf("Run() error kind = %q, want %q", userErr.Kind, core.ErrorFormatMismatch)
	}
}

func TestRunRejectsUnsupportedOutputExtension(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "custom.txt")

	runner := NewRunner()
	err := runner.Run([]string{"https://example.com", "-o", target}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Run() error = nil, want extension error")
	}

	var userErr *core.UserError
	if !errors.As(err, &userErr) {
		t.Fatalf("Run() error = %T, want *core.UserError", err)
	}
	if userErr.Kind != core.ErrorInvalidOutputExtension {
		t.Fatalf("Run() error kind = %q, want %q", userErr.Kind, core.ErrorInvalidOutputExtension)
	}
}

func TestRunCanEncodeLiteralTUIWithDoubleDash(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	var stdout bytes.Buffer
	runner := NewRunner()

	if err := runner.Run([]string{"--", "tui"}, strings.NewReader(""), &stdout, &stdout); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	target := filepath.Join(dir, "qrcode.png")
	got := decodePNGPayload(t, target)
	if got != "tui" {
		t.Fatalf("payload = %q, want %q", got, "tui")
	}
}

func TestRunCanEncodeLiteralTUIWithFlagsBeforeDoubleDash(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "literal.png")

	var stdout bytes.Buffer
	runner := NewRunner()

	err := runner.Run(
		[]string{"-o", target, "-l", "H", "--", "tui"},
		strings.NewReader(""),
		&stdout,
		&stdout,
	)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got := decodePNGPayload(t, target)
	if got != "tui" {
		t.Fatalf("payload = %q, want %q", got, "tui")
	}
}

func TestRunRoutesTUICommand(t *testing.T) {
	var launched bool
	runner := NewRunner()
	runner.LaunchTUI = func() error {
		launched = true
		return nil
	}

	if err := runner.Run([]string{"tui"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !launched {
		t.Fatal("expected LaunchTUI to be called")
	}
}

func TestRunMessageFlagPrintsPreviewToTerminal(t *testing.T) {
	var stdout bytes.Buffer
	runner := NewRunner()

	if err := runner.Run([]string{"-m", "hello"}, strings.NewReader(""), &stdout, &stdout); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "█") && !strings.Contains(output, "▀") && !strings.Contains(output, "▄") {
		t.Fatalf("stdout = %q, want half-block QR preview", output)
	}
	if strings.Contains(output, "Saved to") {
		t.Fatalf("stdout = %q, want no file save confirmation", output)
	}
}

func TestRunMessageFlagDoesNotCreateFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	var stdout bytes.Buffer
	runner := NewRunner()

	if err := runner.Run([]string{"-m", "hello"}, strings.NewReader(""), &stdout, &stdout); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	target := filepath.Join(dir, "qrcode.png")
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected %s not to exist, but it does", target)
	}
}

func TestRunMessageFlagWithLevel(t *testing.T) {
	var stdout bytes.Buffer
	runner := NewRunner()

	if err := runner.Run([]string{"-m", "hello", "-l", "H"}, strings.NewReader(""), &stdout, &stdout); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "█") && !strings.Contains(output, "▀") && !strings.Contains(output, "▄") {
		t.Fatalf("stdout = %q, want half-block QR preview", output)
	}
}

func TestRunMessageFlagRejectsEmptyContent(t *testing.T) {
	runner := NewRunner()

	err := runner.Run([]string{"-m", ""}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Run() error = nil, want empty content error")
	}

	var userErr *core.UserError
	if !errors.As(err, &userErr) {
		t.Fatalf("Run() error = %T, want *core.UserError", err)
	}
	if userErr.Kind != core.ErrorEmptyContent {
		t.Fatalf("Run() error kind = %q, want %q", userErr.Kind, core.ErrorEmptyContent)
	}
}

func TestRunMessageFlagWithOutputSavesFileAndPrintsPreview(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "qr.png")

	var stdout bytes.Buffer
	runner := NewRunner()

	if err := runner.Run([]string{"-m", "hello", "-o", target}, strings.NewReader(""), &stdout, &stdout); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "█") && !strings.Contains(output, "▀") && !strings.Contains(output, "▄") {
		t.Fatalf("stdout = %q, want half-block QR preview", output)
	}
	if !strings.Contains(output, "Saved to") {
		t.Fatalf("stdout = %q, want save confirmation", output)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected %s to exist: %v", target, err)
	}
}

func TestRunMessageFlagWithOutputAndFormatSavesCorrectFormat(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "qr.svg")

	var stdout bytes.Buffer
	runner := NewRunner()

	if err := runner.Run([]string{"-m", "hello", "-o", target, "-f", "svg"}, strings.NewReader(""), &stdout, &stdout); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", target, err)
	}
	if !strings.Contains(string(data), "<svg") {
		t.Fatalf("output %q = %q, want SVG data", target, string(data))
	}
}

func TestRunMessageFlagTakesPriorityOverPositionalArg(t *testing.T) {
	var stdout bytes.Buffer
	runner := NewRunner()

	if err := runner.Run([]string{"-m", "from flag", "from arg"}, strings.NewReader(""), &stdout, &stdout); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := stdout.String()
	if strings.Contains(output, "Saved to") {
		t.Fatalf("stdout = %q, want terminal preview not file save", output)
	}
}

func TestRootHelpShowsMessageFlag(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--help")
	cmd.Dir = filepath.Join("..", "..")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run . --help error = %v\noutput:\n%s", err, output)
	}

	text := string(output)
	if !strings.Contains(text, "-m") {
		t.Fatalf("help output = %q, want -m flag", text)
	}
}

func TestRunRejectsRemovedGenerateSubcommand(t *testing.T) {
	runner := NewRunner()

	err := runner.Run([]string{"generate", "hello"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Run() error = nil, want parse error")
	}
}

func TestTUISubcommandHelpDoesNotRequireTTY(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "tui", "--help")
	cmd.Dir = filepath.Join("..", "..")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run . tui --help error = %v\noutput:\n%s", err, output)
	}

	text := string(output)
	if !strings.Contains(text, "Usage: tqrx tui") {
		t.Fatalf("help output = %q, want tui usage", text)
	}
	if !strings.Contains(text, "Open the interactive terminal UI.") {
		t.Fatalf("help output = %q, want tui description", text)
	}
	if strings.Contains(text, "could not open a new TTY") {
		t.Fatalf("help output = %q, want no TTY failure", text)
	}
}

func TestRootHelpShowsDirectGenerateUsage(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--help")
	cmd.Dir = filepath.Join("..", "..")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run . --help error = %v\noutput:\n%s", err, output)
	}

	text := string(output)
	if !strings.Contains(text, "Usage: tqrx [<content>] [flags]") {
		t.Fatalf("help output = %q, want root usage", text)
	}
	if !strings.Contains(text, "tui  Open the interactive terminal UI.") {
		t.Fatalf("help output = %q, want tui note", text)
	}
	if strings.Contains(text, "Usage: tqrx generate") {
		t.Fatalf("help output = %q, want no generate subcommand", text)
	}
}

func decodePNGPayload(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
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
	return string(qrCodes[0].Payload)
}
