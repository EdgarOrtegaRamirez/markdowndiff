// Package diff provides semantic diffing for markdown documents.
// It compares two lists of blocks and produces a structured diff result.
package diff

import (
	"strings"

	"github.com/EdgarOrtegaRamirez/markdowndiff/pkg/parser"
)

// ChangeType represents the type of change for a block.
type ChangeType int

const (
	ChangeEqual ChangeType = iota
	ChangeAdded
	ChangeRemoved
	ChangeModified
)

// String returns a human-readable label for the change type.
func (ct ChangeType) String() string {
	switch ct {
	case ChangeEqual:
		return "equal"
	case ChangeAdded:
		return "added"
	case ChangeRemoved:
		return "removed"
	case ChangeModified:
		return "modified"
	default:
		return "unknown"
	}
}

// Symbol returns the diff symbol for the change type.
func (ct ChangeType) Symbol() string {
	switch ct {
	case ChangeAdded:
		return "+"
	case ChangeRemoved:
		return "-"
	case ChangeModified:
		return "~"
	case ChangeEqual:
		return " "
	default:
		return "?"
	}
}

// BlockDiff represents the diff result for a single block.
type BlockDiff struct {
	Change    ChangeType
	OldBlock  *parser.Block
	NewBlock  *parser.Block
	OldIndex  int
	NewIndex  int
	SimScore  float64 // Similarity score (0.0 to 1.0)
	WordDiffs []WordDiff
}

// WordDiff represents word-level changes within a modified block.
type WordDiff struct {
	Change ChangeType
	Text   string
}

// DiffResult holds the complete diff result between two documents.
type DiffResult struct {
	Blocks  []BlockDiff
	Stats   DiffStats
	OldText string
	NewText string
}

// DiffStats provides summary statistics for the diff.
type DiffStats struct {
	TotalOld   int
	TotalNew   int
	Added      int
	Removed    int
	Modified   int
	Equal      int
	Similarity float64 // Overall document similarity (0.0 to 1.0)
}

// Diff compares two markdown documents and returns structured diff results.
func Diff(oldText, newText string) *DiffResult {
	oldBlocks := parser.Parse(oldText)
	newBlocks := parser.Parse(newText)

	blockDiffs := computeLCS(oldBlocks, newBlocks)
	stats := computeStats(blockDiffs)

	return &DiffResult{
		Blocks:  blockDiffs,
		Stats:   stats,
		OldText: oldText,
		NewText: newText,
	}
}

// computeLCS uses the Longest Common Subsequence algorithm to find the
// optimal alignment between two block sequences.
func computeLCS(oldBlocks, newBlocks []parser.Block) []BlockDiff {
	n := len(oldBlocks)
	m := len(newBlocks)

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if blocksEqual(oldBlocks[i-1], newBlocks[j-1]) {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	var diffs []BlockDiff
	i, j := n, m
	var stack []BlockDiff

	for i > 0 || j > 0 {
		if i > 0 && j > 0 && blocksEqual(oldBlocks[i-1], newBlocks[j-1]) {
			oldBlock := oldBlocks[i-1]
			newBlock := newBlocks[j-1]
			stack = append(stack, BlockDiff{
				Change:   ChangeEqual,
				OldBlock: &oldBlock,
				NewBlock: &newBlock,
				OldIndex: i - 1,
				NewIndex: j - 1,
				SimScore: 1.0,
			})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			newBlock := newBlocks[j-1]
			stack = append(stack, BlockDiff{
				Change:   ChangeAdded,
				NewBlock: &newBlock,
				NewIndex: j - 1,
				SimScore: 0.0,
			})
			j--
		} else if i > 0 {
			oldBlock := oldBlocks[i-1]
			stack = append(stack, BlockDiff{
				Change:   ChangeRemoved,
				OldBlock: &oldBlock,
				OldIndex: i - 1,
				SimScore: 0.0,
			})
			i--
		}
	}

	for k := len(stack) - 1; k >= 0; k-- {
		diffs = append(diffs, stack[k])
	}

	diffs = detectModifications(diffs)

	return diffs
}

// blocksEqual checks if two blocks have the same content.
func blocksEqual(a, b parser.Block) bool {
	if a.Type != b.Type {
		return false
	}
	if a.Level != b.Level {
		return false
	}
	return parser.BlockToText(a) == parser.BlockToText(b)
}

// detectModifications identifies adjacent removed/added blocks that are
// actually modifications (high similarity).
func detectModifications(diffs []BlockDiff) []BlockDiff {
	var result []BlockDiff
	i := 0

	for i < len(diffs) {
		if diffs[i].Change == ChangeRemoved {
			j := i + 1
			foundMod := false
			for j < len(diffs) && diffs[j].Change == ChangeAdded {
				oldText := safeBlockText(diffs[i].OldBlock)
				newText := safeBlockText(diffs[j].NewBlock)
				sim := similarity(oldText, newText)
				if sim > 0.3 {
					wordDiffs := computeWordDiff(oldText, newText)
					result = append(result, BlockDiff{
						Change:    ChangeModified,
						OldBlock:  diffs[i].OldBlock,
						NewBlock:  diffs[j].NewBlock,
						OldIndex:  diffs[i].OldIndex,
						NewIndex:  diffs[j].NewIndex,
						SimScore:  sim,
						WordDiffs: wordDiffs,
					})
					i = j + 1
					foundMod = true
					break
				}
				j++
			}
			if !foundMod {
				result = append(result, diffs[i])
				i++
			}
		} else {
			result = append(result, diffs[i])
			i++
		}
	}

	return result
}

// safeBlockText returns the text content of a block, handling nil pointers safely.
func safeBlockText(b *parser.Block) string {
	if b == nil {
		return ""
	}
	return parser.BlockToText(*b)
}

// computeStats calculates diff statistics.
func computeStats(diffs []BlockDiff) DiffStats {
	stats := DiffStats{}

	for _, d := range diffs {
		switch d.Change {
		case ChangeEqual:
			stats.Equal++
			stats.TotalOld++
			stats.TotalNew++
		case ChangeAdded:
			stats.Added++
			stats.TotalNew++
		case ChangeRemoved:
			stats.Removed++
			stats.TotalOld++
		case ChangeModified:
			stats.Modified++
			stats.TotalOld++
			stats.TotalNew++
		}
	}

	total := stats.TotalOld + stats.TotalNew
	if total > 0 {
		equalParts := float64(stats.Equal*2+stats.Modified) / float64(total)
		stats.Similarity = equalParts
	}

	return stats
}

// similarity computes the similarity between two strings using the
// Jaccard index on word-level tokens.
func similarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}

	tokensA := tokenize(a)
	tokensB := tokenize(b)

	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0.0
	}

	setA := make(map[string]bool)
	for _, t := range tokensA {
		setA[t] = true
	}

	setB := make(map[string]bool)
	for _, t := range tokensB {
		setB[t] = true
	}

	intersection := 0
	for t := range setA {
		if setB[t] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// tokenize splits text into lowercase word tokens.
func tokenize(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	var tokens []string
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?\"'()[]{}")
		if w != "" {
			tokens = append(tokens, w)
		}
	}
	return tokens
}

// computeWordDiff computes word-level diff between two strings using LCS.
func computeWordDiff(oldText, newText string) []WordDiff {
	oldWords := strings.Fields(oldText)
	newWords := strings.Fields(newText)

	n := len(oldWords)
	m := len(newWords)

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if strings.EqualFold(oldWords[i-1], newWords[j-1]) {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	var result []WordDiff
	i, j := n, m
	var stack []WordDiff

	for i > 0 || j > 0 {
		if i > 0 && j > 0 && strings.EqualFold(oldWords[i-1], newWords[j-1]) {
			stack = append(stack, WordDiff{
				Change: ChangeEqual,
				Text:   oldWords[i-1],
			})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			stack = append(stack, WordDiff{
				Change: ChangeAdded,
				Text:   newWords[j-1],
			})
			j--
		} else if i > 0 {
			stack = append(stack, WordDiff{
				Change: ChangeRemoved,
				Text:   oldWords[i-1],
			})
			i--
		}
	}

	for k := len(stack) - 1; k >= 0; k-- {
		result = append(result, stack[k])
	}

	return result
}
