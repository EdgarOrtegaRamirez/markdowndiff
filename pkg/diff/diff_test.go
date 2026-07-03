package diff

import (
	"testing"
)

func TestDiffIdentical(t *testing.T) {
	text := "# Hello\n\nWorld"
	result := Diff(text, text)

	if result.Stats.Equal == 0 {
		t.Error("expected some equal blocks for identical documents")
	}
	if result.Stats.Added > 0 || result.Stats.Removed > 0 {
		t.Error("expected no additions or removals for identical documents")
	}
	if result.Stats.Similarity != 1.0 {
		t.Errorf("expected similarity 1.0, got %f", result.Stats.Similarity)
	}
}

func TestDiffEmpty(t *testing.T) {
	result := Diff("", "")
	if result.Stats.Similarity != 1.0 {
		t.Errorf("expected similarity 1.0 for empty docs, got %f", result.Stats.Similarity)
	}
}

func TestDiffAddedBlock(t *testing.T) {
	old := "# Title"
	new := "# Title\n\nNew paragraph"

	result := Diff(old, new)

	added := 0
	for _, d := range result.Blocks {
		if d.Change == ChangeAdded {
			added++
		}
	}
	if added == 0 {
		t.Error("expected at least 1 added block")
	}
}

func TestDiffRemovedBlock(t *testing.T) {
	old := "# Title\n\nOld paragraph"
	new := "# Title"

	result := Diff(old, new)

	removed := 0
	for _, d := range result.Blocks {
		if d.Change == ChangeRemoved {
			removed++
		}
	}
	if removed == 0 {
		t.Error("expected at least 1 removed block")
	}
}

func TestDiffModifiedBlock(t *testing.T) {
	old := "# Title"
	new := "# Updated Title"

	result := Diff(old, new)

	modified := 0
	for _, d := range result.Blocks {
		if d.Change == ChangeModified {
			modified++
			if d.SimScore <= 0 || d.SimScore > 1.0 {
				t.Errorf("expected similarity score between 0 and 1, got %f", d.SimScore)
			}
			if len(d.WordDiffs) == 0 {
				t.Error("expected word diffs for modified block")
			}
		}
	}
	if modified == 0 {
		t.Error("expected at least 1 modified block")
	}
}

func TestDiffMultipleChanges(t *testing.T) {
	old := `# Title

First paragraph.

## Section

Old content.`
	new := `# Updated Title

First paragraph updated.

## New Section

New content added.

### Subsection

More content.`

	result := Diff(old, new)

	// The LCS algorithm may identify changes as modifications rather than
	// separate adds/removes when blocks have similarity > 0.3
	totalChanges := result.Stats.Added + result.Stats.Removed + result.Stats.Modified
	if totalChanges == 0 {
		t.Error("expected some changes (additions, removals, or modifications)")
	}
	if result.Stats.Similarity < 0.3 {
		t.Errorf("expected similarity > 0.3, got %f", result.Stats.Similarity)
	}
}

func TestDiffWordLevelChanges(t *testing.T) {
	old := "The quick brown fox"
	new := "The fast brown fox"

	result := Diff(old, new)

	for _, d := range result.Blocks {
		if d.Change == ChangeModified {
			hasAdd := false
			hasRemove := false
			for _, wd := range d.WordDiffs {
				if wd.Change == ChangeAdded {
					hasAdd = true
				}
				if wd.Change == ChangeRemoved {
					hasRemove = true
				}
			}
			if !hasAdd || !hasRemove {
				t.Error("expected both additions and removals in word diff")
			}
			return
		}
	}
	t.Error("expected a modified block")
}

func TestDiffStats(t *testing.T) {
	old := "# A\n\nB"
	new := "# A\n\nB\n\nC"

	result := Diff(old, new)

	if result.Stats.TotalOld == 0 {
		t.Error("expected non-zero TotalOld")
	}
	if result.Stats.TotalNew == 0 {
		t.Error("expected non-zero TotalNew")
	}
}

func TestDiffJSONOutput(t *testing.T) {
	result := Diff("# Hello", "# Hello World")

	if result.OldText != "# Hello" {
		t.Error("expected OldText to be preserved")
	}
	if result.NewText != "# Hello World" {
		t.Error("expected NewText to be preserved")
	}
}

func TestDiffBlockTypes(t *testing.T) {
	old := "# Heading\n\nParagraph."
	new := "# Heading\n\nUpdated paragraph."

	result := Diff(old, new)

	for _, d := range result.Blocks {
		if d.Change == ChangeModified {
			if d.OldBlock == nil || d.NewBlock == nil {
				t.Error("expected both OldBlock and NewBlock for modifications")
			}
		}
	}
}

func TestSimilarity(t *testing.T) {
	tests := []struct {
		a, b string
		min  float64
	}{
		{"hello", "hello", 1.0},
		{"hello", "", 0.0},
		{"hello", "world", 0.0},
		{"the quick brown fox", "the fast brown fox", 0.5},
	}

	for _, tt := range tests {
		sim := similarity(tt.a, tt.b)
		if sim < tt.min {
			t.Errorf("similarity(%q, %q) = %f, expected >= %f", tt.a, tt.b, sim, tt.min)
		}
	}
}

func TestTokenize(t *testing.T) {
	tokens := tokenize("Hello, World!")
	if len(tokens) != 2 {
		t.Errorf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0] != "hello" {
		t.Errorf("expected 'hello', got '%s'", tokens[0])
	}
	if tokens[1] != "world" {
		t.Errorf("expected 'world', got '%s'", tokens[1])
	}
}

func TestDiffComplexMarkdown(t *testing.T) {
	old := `# Project

## Features

- Feature A
- Feature B

## Install

` + "```bash" + `
go install
` + "```" + `

> Note: important`

	new := `# Project (v2)

## Features

- Feature A
- Feature B
- Feature C

## Install

` + "```bash" + `
go install v2
` + "```" + `

## Docs

Read the docs.

> Note: very important`

	result := Diff(old, new)

	if result.Stats.Similarity < 0.3 {
		t.Errorf("expected similarity > 0.3, got %f", result.Stats.Similarity)
	}
	// Changes may be detected as modifications or additions depending on similarity
	totalChanges := result.Stats.Added + result.Stats.Removed + result.Stats.Modified
	if totalChanges == 0 {
		t.Error("expected some changes")
	}
}

func TestChangeTypeString(t *testing.T) {
	tests := []struct {
		ct   ChangeType
		want string
	}{
		{ChangeEqual, "equal"},
		{ChangeAdded, "added"},
		{ChangeRemoved, "removed"},
		{ChangeModified, "modified"},
		{ChangeType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.ct.String(); got != tt.want {
			t.Errorf("ChangeType(%d).String() = %q, want %q", tt.ct, got, tt.want)
		}
	}
}

func TestChangeTypeSymbol(t *testing.T) {
	tests := []struct {
		ct   ChangeType
		want string
	}{
		{ChangeEqual, " "},
		{ChangeAdded, "+"},
		{ChangeRemoved, "-"},
		{ChangeModified, "~"},
		{ChangeType(99), "?"},
	}

	for _, tt := range tests {
		if got := tt.ct.Symbol(); got != tt.want {
			t.Errorf("ChangeType(%d).Symbol() = %q, want %q", tt.ct, got, tt.want)
		}
	}
}
