# MarkdownDiff

**Semantic diff for Markdown documents** — understands document structure, not just lines.

Regular `diff` is noisy for Markdown: reformatting a paragraph shows as every line changed. MarkdownDiff parses documents into semantic blocks (headings, paragraphs, lists, code blocks, tables, blockquotes) and compares them structurally, providing meaningful change summaries with word-level diffs.

## Features

- **Semantic Block Parsing** — understands headings, paragraphs, lists, code blocks, tables, blockquotes, horizontal rules
- **LCS-based Alignment** — optimal block alignment using Longest Common Subsequence algorithm
- **Modification Detection** — identifies similar blocks as modifications rather than separate add/remove
- **Word-level Diffs** — shows exactly which words changed within modified blocks
- **Similarity Scoring** — Jaccard index-based similarity for each block and overall document
- **Multiple Output Formats** — unified, side-by-side, JSON, and summary views
- **Color Support** — ANSI-colored terminal output (configurable)
- **Stdin Support** — pipe input from other commands

## Installation

```bash
go install github.com/EdgarOrtegaRamirez/markdowndiff/cmd/markdowndiff@latest
```

Or build from source:

```bash
git clone https://github.com/EdgarOrtegaRamirez/markdowndiff.git
cd markdowndiff
go build -o markdowndiff ./cmd/markdowndiff/
```

## Usage

### Compare Two Files

```bash
markdowndiff diff old.md new.md
```

### Compare with Stdin

```bash
cat old.md | markdowndiff diff - new.md
```

### Output Formats

```bash
# Unified diff (default)
markdowndiff diff old.md new.md

# Side-by-side view
markdowndiff diff --format side-by-side old.md new.md

# JSON output (for programmatic use)
markdowndiff diff --format json old.md new.md

# High-level summary
markdowndiff diff --format summary old.md new.md
```

### Options

```
--format, -f <format>   Output format: unified, side-by-side, json, summary
--no-color              Disable colored output
--no-stats              Hide statistics header
--width, -w <width>     Column width for side-by-side (default: 80)
--help, -h              Show help
```

### Parse Command

View the block structure of a markdown file:

```bash
markdowndiff parse document.md
markdowndiff parse --types document.md  # Show block indices
```

### Compare Command

Quick comparison with summary output:

```bash
markdowndiff compare old.md new.md
```

## Example Output

### Unified Format

```
--- Old (24 blocks)  +++ New (32 blocks)  [8+, 0-, 5~, 77% match]
~ [heading] "Project Title" → "Project Title (Updated)"
    Project
    Title
  + (Updated)
~ [paragraph] "This is the introduction pa..." → "This is the introduction pa..."
    This is the introduction paragraph for the project. It describes what the project
  - does.
  + does and why it matters.
```

### Summary Format

```
=== Document Diff Summary ===

Similarity: 76.8%
Old blocks: 24  |  New blocks: 32
Added: 8  |  Removed: 0  |  Modified: 5  |  Equal: 19

Modified blocks:
  ~ [heading] 33% changed
  ~ [paragraph] 21% changed
  ~ [list] 29% changed
```

## Architecture

```
markdowndiff/
├── cmd/markdowndiff/     # CLI entry point
│   └── main.go
├── pkg/
│   ├── parser/           # Markdown parser (block-level)
│   │   ├── parser.go     # Block parser with type detection
│   │   └── parser_test.go
│   ├── diff/             # Semantic diff engine
│   │   ├── diff.go       # LCS alignment, similarity, word diffs
│   │   └── diff_test.go
│   └── output/           # Output formatters
│       ├── output.go     # Unified, side-by-side, JSON, summary
│       └── output_test.go
├── tests/                # Integration tests
│   └── integration_test.go
├── .github/workflows/    # CI configuration
│   └── ci.yml
├── AGENTS.md
├── LICENSE (MIT)
└── README.md
```

### How It Works

1. **Parse** — Markdown text is parsed into a sequence of semantic blocks (heading, paragraph, list, code block, table, blockquote, etc.)
2. **Align** — The LCS algorithm finds the optimal alignment between two block sequences, maximizing equal matches
3. **Detect** — Adjacent removed/added blocks with similarity > 30% are identified as modifications
4. **Diff** — Word-level LCS computes exact changes within modified blocks
5. **Render** — Results are formatted in the chosen output format

## When to Use This

- **Documentation reviews** — see meaningful changes, not line noise
- **Changelog generation** — identify what actually changed between versions
- **Wiki diffs** — compare markdown wiki pages semantically
- **README updates** — track documentation evolution
- **CI/CD pipelines** — detect breaking changes in documentation

## License

MIT
