# AGENTS.md

## Project: MarkdownDiff

A semantic diff tool for Markdown documents that understands document structure.

## Building

```bash
go build -o markdowndiff ./cmd/markdowndiff/
```

## Testing

```bash
go test ./...
```

Run specific package tests:
```bash
go test ./pkg/parser/ -v
go test ./pkg/diff/ -v
go test ./pkg/output/ -v
go test ./tests/ -v
```

## Project Structure

- `cmd/markdowndiff/` — CLI entry point
- `pkg/parser/` — Markdown parser (block-level)
- `pkg/diff/` — Semantic diff engine (LCS, similarity, word diffs)
- `pkg/output/` — Output formatters (unified, side-by-side, JSON, summary)
- `tests/` — Integration tests

## Key Algorithms

- **LCS (Longest Common Subsequence)** — Used for both block alignment and word-level diffs
- **Jaccard Similarity** — Measures similarity between blocks using word token sets
- **Modification Detection** — Adjacent remove+add pairs with similarity > 30% are merged into modifications

## Adding New Block Types

1. Add the type constant to `pkg/parser/parser.go` (BlockType enum)
2. Add parsing logic in the `Parse` function
3. Update `BlockToText` for text extraction
4. Add tests in `pkg/parser/parser_test.go`

## Code Style

- Use Go standard formatting (`gofmt`)
- Export only necessary types and functions
- Handle nil pointers explicitly
- Add tests for new functionality
