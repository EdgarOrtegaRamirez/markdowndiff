package parser

import (
	"testing"
)

func TestParseHeadings(t *testing.T) {
	input := `# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6`
	blocks := Parse(input)

	headings := 0
	for _, b := range blocks {
		if b.Type == BlockHeading {
			headings++
		}
	}
	if headings != 6 {
		t.Errorf("expected 6 headings, got %d", headings)
	}
}

func TestParseParagraphs(t *testing.T) {
	input := `This is a paragraph.

This is another paragraph.`
	blocks := Parse(input)

	paragraphs := 0
	for _, b := range blocks {
		if b.Type == BlockParagraph {
			paragraphs++
		}
	}
	if paragraphs != 2 {
		t.Errorf("expected 2 paragraphs, got %d", paragraphs)
	}
}

func TestParseUnorderedList(t *testing.T) {
	input := `- Item one
- Item two
- Item three`
	blocks := Parse(input)

	lists := 0
	for _, b := range blocks {
		if b.Type == BlockList {
			lists++
			if len(b.Items) != 3 {
				t.Errorf("expected 3 items, got %d", len(b.Items))
			}
		}
	}
	if lists != 1 {
		t.Errorf("expected 1 list, got %d", lists)
	}
}

func TestParseOrderedList(t *testing.T) {
	input := `1. First item
2. Second item
3. Third item`
	blocks := Parse(input)

	ordered := 0
	for _, b := range blocks {
		if b.Type == BlockOrderedList {
			ordered++
			if len(b.Items) != 3 {
				t.Errorf("expected 3 items, got %d", len(b.Items))
			}
			if !b.IsOrdered {
				t.Error("expected ordered list")
			}
		}
	}
	if ordered != 1 {
		t.Errorf("expected 1 ordered list, got %d", ordered)
	}
}

func TestParseCodeBlock(t *testing.T) {
	input := "```go\npackage main\n\nfunc main() {}\n```"
	blocks := Parse(input)

	codeBlocks := 0
	for _, b := range blocks {
		if b.Type == BlockCodeBlock {
			codeBlocks++
			if b.Lang != "go" {
				t.Errorf("expected lang 'go', got '%s'", b.Lang)
			}
		}
	}
	if codeBlocks != 1 {
		t.Errorf("expected 1 code block, got %d", codeBlocks)
	}
}

func TestParseTable(t *testing.T) {
	input := `| Name | Age |
|------|-----|
| John | 30  |
| Jane | 25  |`
	blocks := Parse(input)

	tables := 0
	for _, b := range blocks {
		if b.Type == BlockTable {
			tables++
			if len(b.Lines) != 4 {
				t.Errorf("expected 4 table lines, got %d", len(b.Lines))
			}
		}
	}
	if tables != 1 {
		t.Errorf("expected 1 table, got %d", tables)
	}
}

func TestParseBlockquote(t *testing.T) {
	input := `> This is a quote.
> It spans multiple lines.`
	blocks := Parse(input)

	blockquotes := 0
	for _, b := range blocks {
		if b.Type == BlockBlockquote {
			blockquotes++
			if b.Content == "" {
				t.Error("expected non-empty blockquote content")
			}
		}
	}
	if blockquotes != 1 {
		t.Errorf("expected 1 blockquote, got %d", blockquotes)
	}
}

func TestParseHorizontalRule(t *testing.T) {
	input := `---
***
___`
	blocks := Parse(input)

	rules := 0
	for _, b := range blocks {
		if b.Type == BlockHorizontalRule {
			rules++
		}
	}
	if rules != 3 {
		t.Errorf("expected 3 horizontal rules, got %d", rules)
	}
}

func TestParseBlankLines(t *testing.T) {
	input := `Paragraph 1

Paragraph 2

Paragraph 3`
	blocks := Parse(input)

	blanks := 0
	for _, b := range blocks {
		if b.Type == BlockBlank {
			blanks++
		}
	}
	if blanks != 2 {
		t.Errorf("expected 2 blank lines, got %d", blanks)
	}
}

func TestParseComplexDocument(t *testing.T) {
	input := `# Title

This is a paragraph.

## Section

- Item 1
- Item 2

` + "```" + `
code here
` + "```" + `

> A quote

---

End.`
	blocks := Parse(input)

	if len(blocks) < 5 {
		t.Errorf("expected at least 5 blocks, got %d", len(blocks))
	}

	types := make(map[BlockType]bool)
	for _, b := range blocks {
		types[b.Type] = true
	}

	expectedTypes := []BlockType{BlockHeading, BlockParagraph, BlockList, BlockCodeBlock, BlockBlockquote, BlockHorizontalRule, BlockBlank}
	for _, et := range expectedTypes {
		if !types[et] {
			t.Errorf("expected block type %s not found", et)
		}
	}
}

func TestBlockToText(t *testing.T) {
	heading := Block{Type: BlockHeading, Content: "Hello"}
	if BlockToText(heading) != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", BlockToText(heading))
	}

	para := Block{Type: BlockParagraph, Content: "World"}
	if BlockToText(para) != "World" {
		t.Errorf("expected 'World', got '%s'", BlockToText(para))
	}

	list := Block{Type: BlockList, Items: []string{"a", "b", "c"}}
	if BlockToText(list) != "a\nb\nc" {
		t.Errorf("expected 'a\\nb\\nc', got '%s'", BlockToText(list))
	}
}

func TestBlockTypeString(t *testing.T) {
	tests := []struct {
		bt   BlockType
		want string
	}{
		{BlockHeading, "heading"},
		{BlockParagraph, "paragraph"},
		{BlockList, "list"},
		{BlockOrderedList, "ordered_list"},
		{BlockCodeBlock, "code_block"},
		{BlockTable, "table"},
		{BlockBlockquote, "blockquote"},
		{BlockHorizontalRule, "horizontal_rule"},
		{BlockHTML, "html"},
		{BlockBlank, "blank"},
		{BlockType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.bt.String(); got != tt.want {
			t.Errorf("BlockType(%d).String() = %q, want %q", tt.bt, got, tt.want)
		}
	}
}

func TestEmptyInput(t *testing.T) {
	blocks := Parse("")
	if len(blocks) != 1 {
		t.Errorf("expected 1 block (blank) for empty input, got %d", len(blocks))
	}
	if len(blocks) > 0 && blocks[0].Type != BlockBlank {
		t.Errorf("expected blank block type, got %s", blocks[0].Type)
	}
}

func TestWhitespaceOnly(t *testing.T) {
	blocks := Parse("   \n  \n   ")
	// Should have blank blocks
	if len(blocks) == 0 {
		t.Error("expected at least 1 block for whitespace-only input")
	}
}
