package output

import (
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/markdowndiff/pkg/diff"
)

func TestRenderUnified(t *testing.T) {
	result := diff.Diff("# Old Title", "# New Title")
	formatter := NewFormatter(FormatUnified)
	formatter.Colorize = false

	output := formatter.Render(result)
	if !strings.Contains(output, "Old Title") && !strings.Contains(output, "New Title") {
		t.Error("expected diff output to contain title references")
	}
}

func TestRenderJSON(t *testing.T) {
	result := diff.Diff("# Hello", "# Hello World")
	formatter := NewFormatter(FormatJSON)
	formatter.Colorize = false

	output := formatter.Render(result)
	if !strings.Contains(output, `"stats"`) {
		t.Error("expected JSON output to contain stats")
	}
	if !strings.Contains(output, `"blocks"`) {
		t.Error("expected JSON output to contain blocks")
	}
}

func TestRenderSummary(t *testing.T) {
	result := diff.Diff("# A\n\nB", "# A\n\nB\n\nC")
	formatter := NewFormatter(FormatSummary)
	formatter.Colorize = false

	output := formatter.Render(result)
	if !strings.Contains(output, "Similarity") {
		t.Error("expected summary to contain similarity")
	}
}

func TestRenderSideBySide(t *testing.T) {
	result := diff.Diff("# Old", "# New")
	formatter := NewFormatter(FormatSideBySide)
	formatter.Colorize = false
	formatter.Width = 80

	output := formatter.Render(result)
	if !strings.Contains(output, "OLD") || !strings.Contains(output, "NEW") {
		t.Error("expected side-by-side to contain OLD and NEW headers")
	}
}

func TestNoColor(t *testing.T) {
	result := diff.Diff("# Hello", "# Hello World")
	formatter := NewFormatter(FormatUnified)
	formatter.Colorize = false

	output := formatter.Render(result)
	if strings.Contains(output, "\033[") {
		t.Error("expected no ANSI color codes when colorize is false")
	}
}

func TestWithColor(t *testing.T) {
	result := diff.Diff("# Hello", "# Hello World")
	formatter := NewFormatter(FormatUnified)
	formatter.Colorize = true

	output := formatter.Render(result)
	if !strings.Contains(output, "\033[") {
		t.Error("expected ANSI color codes when colorize is true")
	}
}

func TestNoStats(t *testing.T) {
	result := diff.Diff("# Hello", "# Hello World")
	formatter := NewFormatter(FormatUnified)
	formatter.ShowStats = false
	formatter.Colorize = false

	output := formatter.Render(result)
	if strings.Contains(output, "Old (") {
		t.Error("expected no stats header when ShowStats is false")
	}
}
