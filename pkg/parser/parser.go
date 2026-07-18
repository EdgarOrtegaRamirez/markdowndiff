// Package parser provides a Markdown parser that converts markdown text into
// a list of semantic blocks for structural comparison.
package parser

import (
	"strings"
)

// BlockType represents the type of a markdown block element.
type BlockType int

const (
	BlockHeading BlockType = iota
	BlockParagraph
	BlockList
	BlockOrderedList
	BlockCodeBlock
	BlockTable
	BlockBlockquote
	BlockHorizontalRule
	BlockHTML
	BlockBlank
)

// String returns the human-readable name of a block type.
func (bt BlockType) String() string {
	switch bt {
	case BlockHeading:
		return "heading"
	case BlockParagraph:
		return "paragraph"
	case BlockList:
		return "list"
	case BlockOrderedList:
		return "ordered_list"
	case BlockCodeBlock:
		return "code_block"
	case BlockTable:
		return "table"
	case BlockBlockquote:
		return "blockquote"
	case BlockHorizontalRule:
		return "horizontal_rule"
	case BlockHTML:
		return "html"
	case BlockBlank:
		return "blank"
	default:
		return "unknown"
	}
}

// Block represents a single semantic block in a markdown document.
type Block struct {
	Type      BlockType
	Level     int      // Heading level (1-6) or list nesting level
	Content   string   // Raw text content of the block
	Lines     []string // Individual lines in the block
	Lang      string   // Language for code blocks
	Items     []string // List items for list blocks
	IsOrdered bool     // Whether a list is ordered
}

// Parse converts markdown text into a list of semantic blocks.
func Parse(text string) []Block {
	lines := strings.Split(text, "\n")
	var blocks []Block
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Blank line
		if strings.TrimSpace(line) == "" {
			blocks = append(blocks, Block{
				Type:  BlockBlank,
				Lines: []string{line},
			})
			i++
			continue
		}

		// Heading
		if level, ok := isHeading(line); ok {
			blocks = append(blocks, Block{
				Type:    BlockHeading,
				Level:   level,
				Content: strings.TrimLeft(line, "# "),
				Lines:   []string{line},
			})
			i++
			continue
		}

		// Horizontal rule
		if isHorizontalRule(line) {
			blocks = append(blocks, Block{
				Type:  BlockHorizontalRule,
				Lines: []string{line},
			})
			i++
			continue
		}

		// Code block (fenced)
		if lang, ok := isFenceOpen(line); ok {
			var codeLines []string
			codeLines = append(codeLines, line)
			i++
			for i < len(lines) {
				codeLines = append(codeLines, lines[i])
				if isFenceClose(lines[i]) {
					i++
					break
				}
				i++
			}
			blocks = append(blocks, Block{
				Type:    BlockCodeBlock,
				Lang:    lang,
				Content: strings.Join(codeLines[1:len(codeLines)-1], "\n"),
				Lines:   codeLines,
			})
			continue
		}

		// Table (starts with |)
		if isTableStart(line) {
			var tableLines []string
			for i < len(lines) && isTableRow(lines[i]) {
				tableLines = append(tableLines, lines[i])
				i++
			}
			blocks = append(blocks, Block{
				Type:  BlockTable,
				Lines: tableLines,
			})
			continue
		}

		// Blockquote
		if isBlockquote(line) {
			var bqLines []string
			for i < len(lines) && (isBlockquote(lines[i]) || (strings.TrimSpace(lines[i]) != "" && !isSpecialLine(lines[i]))) {
				bqLines = append(bqLines, lines[i])
				i++
				if i < len(lines) && strings.TrimSpace(lines[i]) == "" {
					break
				}
			}
			content := stripBlockquotePrefix(bqLines)
			blocks = append(blocks, Block{
				Type:    BlockBlockquote,
				Content: content,
				Lines:   bqLines,
			})
			continue
		}

		// Unordered list
		if isUnorderedList(line) {
			var items []string
			var listLines []string
			for i < len(lines) && (isUnorderedList(lines[i]) || isListItemContinuation(lines[i])) {
				listLines = append(listLines, lines[i])
				if isUnorderedList(lines[i]) {
					items = append(items, strings.TrimLeft(lines[i], "- *+"))
				}
				i++
			}
			blocks = append(blocks, Block{
				Type:  BlockList,
				Items: items,
				Lines: listLines,
			})
			continue
		}

		// Ordered list
		if isOrderedList(line) {
			var items []string
			var listLines []string
			for i < len(lines) && (isOrderedList(lines[i]) || isListItemContinuation(lines[i])) {
				listLines = append(listLines, lines[i])
				if isOrderedList(lines[i]) {
					// Strip the number prefix
					content := lines[i]
					idx := strings.Index(content, ". ")
					if idx >= 0 {
						content = content[idx+2:]
					}
					items = append(items, content)
				}
				i++
			}
			blocks = append(blocks, Block{
				Type:      BlockOrderedList,
				Items:     items,
				IsOrdered: true,
				Lines:     listLines,
			})
			continue
		}

		// HTML block (simplified detection)
		if strings.HasPrefix(strings.TrimSpace(line), "<") && !strings.HasPrefix(strings.TrimSpace(line), "<!") {
			tag := extractHTMLTag(line)
			if tag != "" {
				var htmlLines []string
				htmlLines = append(htmlLines, line)
				closingTag := "</" + tag + ">"
				i++
				for i < len(lines) {
					htmlLines = append(htmlLines, lines[i])
					if strings.Contains(lines[i], closingTag) {
						i++
						break
					}
					i++
				}
				blocks = append(blocks, Block{
					Type:  BlockHTML,
					Lines: htmlLines,
				})
				continue
			}
		}

		// Paragraph (default)
		var paraLines []string
		for i < len(lines) && strings.TrimSpace(lines[i]) != "" && !isSpecialLine(lines[i]) {
			paraLines = append(paraLines, lines[i])
			i++
		}
		if len(paraLines) > 0 {
			blocks = append(blocks, Block{
				Type:    BlockParagraph,
				Content: strings.Join(paraLines, " "),
				Lines:   paraLines,
			})
		}
	}

	return blocks
}

// Helper functions

func isHeading(line string) (int, bool) {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) == 0 || trimmed[0] != '#' {
		return 0, false
	}
	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level > 0 && level <= 6 && level < len(trimmed) && trimmed[level] == ' ' {
		return level, true
	}
	return 0, false
}

func isHorizontalRule(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return false
	}
	char := trimmed[0]
	if char != '-' && char != '*' && char != '_' {
		return false
	}
	count := 0
	for _, c := range trimmed {
		if c == rune(char) {
			count++
		} else if c != ' ' {
			return false
		}
	}
	return count >= 3
}

func isFenceOpen(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return "", false
	}
	fence := trimmed[:3]
	if fence != "```" && fence != "~~~" {
		return "", false
	}
	lang := strings.TrimSpace(trimmed[3:])
	return lang, true
}

func isFenceClose(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "```" || trimmed == "~~~"
}

func isTableStart(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "|")
}

func isTableRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return false
	}
	// Check if it's a separator row (|---|---|)
	inner := strings.Trim(trimmed, "| ")
	if strings.Contains(inner, "---") {
		return true
	}
	return strings.HasPrefix(trimmed, "|")
}

func isBlockquote(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "> ")
}

func isUnorderedList(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ")
}

func isOrderedList(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 2 {
		return false
	}
	for i, c := range trimmed {
		if c >= '0' && c <= '9' {
			continue
		}
		if c == '.' && i > 0 && i < len(trimmed)-1 && trimmed[i+1] == ' ' {
			return true
		}
		return false
	}
	return false
}

func isListItemContinuation(line string) bool {
	trimmed := strings.TrimSpace(line)
	// Continuation lines start with spaces (indented)
	return len(line) > 0 && line[0] == ' ' && trimmed != ""
}

func isSpecialLine(line string) bool {
	// Check for block-level elements that should break paragraphs
	if _, ok := isHeading(line); ok {
		return true
	}
	if isHorizontalRule(line) {
		return true
	}
	if isTableStart(line) {
		return true
	}
	if isUnorderedList(line) || isOrderedList(line) {
		return true
	}
	if isBlockquote(line) {
		return true
	}
	if lang, ok := isFenceOpen(line); ok || lang != "" {
		return true
	}
	return false
}

func extractHTMLTag(line string) string {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "<") {
		return ""
	}
	end := strings.IndexAny(trimmed, " >")
	if end <= 1 {
		return ""
	}
	tag := trimmed[1:end]
	// Only block-level tags
	blockTags := map[string]bool{
		"div": true, "p": true, "table": true, "tr": true, "td": true,
		"th": true, "thead": true, "tbody": true, "ul": true, "ol": true,
		"li": true, "pre": true, "blockquote": true, "h1": true, "h2": true,
		"h3": true, "h4": true, "h5": true, "h6": true, "hr": true,
		"section": true, "article": true, "nav": true, "header": true,
		"footer": true, "aside": true, "main": true, "details": true,
		"summary": true,
	}
	if blockTags[tag] {
		return tag
	}
	return ""
}

func stripBlockquotePrefix(lines []string) string {
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "> ") {
			result = append(result, trimmed[2:])
		} else if strings.HasPrefix(trimmed, ">") {
			result = append(result, trimmed[1:])
		} else {
			result = append(result, trimmed)
		}
	}
	return strings.Join(result, "\n")
}

// BlockToText returns the plain text content of a block for comparison.
func BlockToText(b Block) string {
	switch b.Type {
	case BlockHeading:
		return b.Content
	case BlockParagraph:
		return b.Content
	case BlockList, BlockOrderedList:
		return strings.Join(b.Items, "\n")
	case BlockCodeBlock:
		return b.Content
	case BlockTable:
		return strings.Join(b.Lines, "\n")
	case BlockBlockquote:
		return b.Content
	default:
		return strings.Join(b.Lines, "\n")
	}
}
