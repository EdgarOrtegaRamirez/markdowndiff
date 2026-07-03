// Package output provides formatters for displaying diff results.
package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/markdowndiff/pkg/diff"
	"github.com/EdgarOrtegaRamirez/markdowndiff/pkg/parser"
)

// FormatType represents the output format.
type FormatType string

const (
	FormatUnified    FormatType = "unified"
	FormatSideBySide FormatType = "side-by-side"
	FormatJSON       FormatType = "json"
	FormatSummary    FormatType = "summary"
)

// Formatter outputs diff results in various formats.
type Formatter struct {
	Format    FormatType
	Colorize  bool
	Width     int
	ShowStats bool
}

// NewFormatter creates a new Formatter with default settings.
func NewFormatter(format FormatType) *Formatter {
	return &Formatter{
		Format:    format,
		Colorize:  true,
		Width:     80,
		ShowStats: true,
	}
}

// Render formats a diff result into a string.
func (f *Formatter) Render(result *diff.DiffResult) string {
	switch f.Format {
	case FormatJSON:
		return f.renderJSON(result)
	case FormatSummary:
		return f.renderSummary(result)
	case FormatSideBySide:
		return f.renderSideBySide(result)
	default:
		return f.renderUnified(result)
	}
}

// renderUnified renders the diff in unified format (like git diff).
func (f *Formatter) renderUnified(result *diff.DiffResult) string {
	var sb strings.Builder

	if f.ShowStats {
		sb.WriteString(f.renderStatsHeader(result))
		sb.WriteString("\n")
	}

	for _, bd := range result.Blocks {
		switch bd.Change {
		case diff.ChangeEqual:
			// Equal blocks are not shown in unified view
			continue
		case diff.ChangeAdded:
			sb.WriteString(f.colorizeAdded("+" + formatBlockForDiff(bd.NewBlock)))
			sb.WriteString("\n")
		case diff.ChangeRemoved:
			sb.WriteString(f.colorizeRemoved("-" + formatBlockForDiff(bd.OldBlock)))
			sb.WriteString("\n")
		case diff.ChangeModified:
			sb.WriteString(f.colorizeModified("~ " + blockHeader(bd.OldBlock, bd.NewBlock)))
			sb.WriteString("\n")
			// Show word-level diffs
			for _, wd := range bd.WordDiffs {
				switch wd.Change {
				case diff.ChangeAdded:
					sb.WriteString(f.colorizeAdded("  + " + wd.Text))
					sb.WriteString("\n")
				case diff.ChangeRemoved:
					sb.WriteString(f.colorizeRemoved("  - " + wd.Text))
					sb.WriteString("\n")
				case diff.ChangeEqual:
					sb.WriteString("    " + wd.Text)
					sb.WriteString("\n")
				}
			}
		}
	}

	return sb.String()
}

// renderSideBySide renders the diff in side-by-side format.
func (f *Formatter) renderSideBySide(result *diff.DiffResult) string {
	var sb strings.Builder
	halfWidth := (f.Width - 3) / 2

	if f.ShowStats {
		sb.WriteString(f.renderStatsHeader(result))
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("%-*s | %s\n", halfWidth, "OLD", "NEW"))
	sb.WriteString(strings.Repeat("-", halfWidth) + "-+-" + strings.Repeat("-", halfWidth) + "\n")

	for _, bd := range result.Blocks {
		switch bd.Change {
		case diff.ChangeEqual:
			text := truncate(blockText(bd.OldBlock), halfWidth)
			sb.WriteString(fmt.Sprintf("%-*s | %s\n", halfWidth, text, text))
		case diff.ChangeAdded:
			sb.WriteString(fmt.Sprintf("%-*s | %s\n", halfWidth, "", f.colorizeAdded(truncate(blockText(bd.NewBlock), halfWidth))))
		case diff.ChangeRemoved:
			sb.WriteString(fmt.Sprintf("%-*s | %s\n", halfWidth, f.colorizeRemoved(truncate(blockText(bd.OldBlock), halfWidth)), ""))
		case diff.ChangeModified:
			oldText := truncate(blockText(bd.OldBlock), halfWidth)
			newText := truncate(blockText(bd.NewBlock), halfWidth)
			sb.WriteString(fmt.Sprintf("%-*s | %s\n", halfWidth, f.colorizeRemoved(oldText), f.colorizeAdded(newText)))
		}
	}

	return sb.String()
}

// renderJSON renders the diff result as formatted JSON.
func (f *Formatter) renderJSON(result *diff.DiffResult) string {
	type jsonBlock struct {
		Type    string  `json:"type"`
		Change  string  `json:"change"`
		OldText *string `json:"old_text,omitempty"`
		NewText *string `json:"new_text,omitempty"`
		Score   float64 `json:"similarity_score"`
	}

	var blocks []jsonBlock
	for _, bd := range result.Blocks {
		jb := jsonBlock{
			Change: bd.Change.String(),
			Score:  bd.SimScore,
		}

		if bd.OldBlock != nil {
			jb.Type = bd.OldBlock.Type.String()
			t := blockText(bd.OldBlock)
			jb.OldText = &t
		}
		if bd.NewBlock != nil {
			if jb.Type == "" {
				jb.Type = bd.NewBlock.Type.String()
			}
			t := blockText(bd.NewBlock)
			jb.NewText = &t
		}

		blocks = append(blocks, jb)
	}

	type jsonOutput struct {
		Stats  diff.DiffStats `json:"stats"`
		Blocks []jsonBlock    `json:"blocks"`
	}

	output := jsonOutput{
		Stats:  result.Stats,
		Blocks: blocks,
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	return string(data)
}

// renderSummary renders a high-level summary of the diff.
func (f *Formatter) renderSummary(result *diff.DiffResult) string {
	var sb strings.Builder

	sb.WriteString("=== Document Diff Summary ===\n\n")

	stats := result.Stats
	sb.WriteString(fmt.Sprintf("Similarity: %.1f%%\n", stats.Similarity*100))
	sb.WriteString(fmt.Sprintf("Old blocks: %d  |  New blocks: %d\n", stats.TotalOld, stats.TotalNew))
	sb.WriteString(fmt.Sprintf("Added: %d  |  Removed: %d  |  Modified: %d  |  Equal: %d\n\n",
		stats.Added, stats.Removed, stats.Modified, stats.Equal))

	// List changes by type
	if stats.Added > 0 {
		sb.WriteString("Added blocks:\n")
		for _, bd := range result.Blocks {
			if bd.Change == diff.ChangeAdded {
				sb.WriteString(f.colorizeAdded(fmt.Sprintf("  + [%s] %s\n", bd.NewBlock.Type, truncate(blockText(bd.NewBlock), 60))))
			}
		}
		sb.WriteString("\n")
	}

	if stats.Removed > 0 {
		sb.WriteString("Removed blocks:\n")
		for _, bd := range result.Blocks {
			if bd.Change == diff.ChangeRemoved {
				sb.WriteString(f.colorizeRemoved(fmt.Sprintf("  - [%s] %s\n", bd.OldBlock.Type, truncate(blockText(bd.OldBlock), 60))))
			}
		}
		sb.WriteString("\n")
	}

	if stats.Modified > 0 {
		sb.WriteString("Modified blocks:\n")
		for _, bd := range result.Blocks {
			if bd.Change == diff.ChangeModified {
				sb.WriteString(f.colorizeModified(fmt.Sprintf("  ~ [%s] %.0f%% changed\n", bd.OldBlock.Type, (1-bd.SimScore)*100)))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderStatsHeader returns a compact stats header line.
func (f *Formatter) renderStatsHeader(result *diff.DiffResult) string {
	s := result.Stats
	return fmt.Sprintf("--- Old (%d blocks)  +++ New (%d blocks)  [%d+, %d-, %d~, %.0f%% match]",
		s.TotalOld, s.TotalNew, s.Added, s.Removed, s.Modified, s.Similarity*100)
}

// formatBlockForDiff returns a single-line representation of a block.
func formatBlockForDiff(b *parser.Block) string {
	if b == nil {
		return ""
	}
	text := blockText(b)
	return fmt.Sprintf("[%s] %s", b.Type, truncate(text, 72))
}

// blockText returns the text content of a block.
func blockText(b *parser.Block) string {
	if b == nil {
		return ""
	}
	return parser.BlockToText(*b)
}

// blockHeader returns a header for a modified block.
func blockHeader(old, new *parser.Block) string {
	oldText := truncate(blockText(old), 30)
	newText := truncate(blockText(new), 30)
	return fmt.Sprintf("[%s] %q → %q", old.Type, oldText, newText)
}

// truncate shortens text to maxWidth, appending "..." if truncated.
func truncate(text string, maxWidth int) string {
	text = strings.ReplaceAll(text, "\n", " ")
	if len(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return text[:maxWidth]
	}
	return text[:maxWidth-3] + "..."
}

// ANSI color helpers

func (f *Formatter) colorizeAdded(text string) string {
	if !f.Colorize {
		return text
	}
	return "\033[32m" + text + "\033[0m"
}

func (f *Formatter) colorizeRemoved(text string) string {
	if !f.Colorize {
		return text
	}
	return "\033[31m" + text + "\033[0m"
}

func (f *Formatter) colorizeModified(text string) string {
	if !f.Colorize {
		return text
	}
	return "\033[33m" + text + "\033[0m"
}
