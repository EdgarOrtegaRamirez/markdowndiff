package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/markdowndiff/pkg/diff"
	"github.com/EdgarOrtegaRamirez/markdowndiff/pkg/output"
	"github.com/EdgarOrtegaRamirez/markdowndiff/pkg/parser"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "diff":
		cmdDiff(args)
	case "compare":
		cmdCompare(args)
	case "parse":
		cmdParse(args)
	case "version":
		fmt.Printf("markdowndiff v%s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`markdowndiff - Semantic diff for Markdown documents

Usage:
  markdowndiff <command> [options]

Commands:
  diff      Compare two markdown files (or stdin + file)
  compare   Compare two files and show a summary
  parse     Parse a markdown file and show its block structure
  version   Show version information
  help      Show this help message

Examples:
  # Compare two files
  markdowndiff diff old.md new.md

  # Compare with stdin
  cat old.md | markdowndiff diff - new.md

  # Side-by-side output
  markdowndiff diff --format side-by-side old.md new.md

  # JSON output
  markdowndiff diff --format json old.md new.md

  # Show block structure
  markdowndiff parse document.md

  # No color output
  markdowndiff diff --no-color old.md new.md

  # Custom width for side-by-side
  markdowndiff diff --format side-by-side --width 120 old.md new.md
`)
}

func cmdDiff(args []string) {
	format := output.FormatUnified
	colorize := true
	width := 80
	showStats := true
	var files []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format", "-f":
			if i+1 < len(args) {
				i++
				format = output.FormatType(args[i])
			}
		case "--no-color":
			colorize = false
		case "--no-stats":
			showStats = false
		case "--width", "-w":
			if i+1 < len(args) {
				i++
				fmt.Sscanf(args[i], "%d", &width)
			}
		case "--help", "-h":
			printDiffUsage()
			return
		default:
			files = append(files, args[i])
		}
	}

	var oldText, newText string
	var err error

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Error: at least one file required")
		os.Exit(1)
	} else if len(files) == 1 {
		// One file: compare with stdin
		oldText, err = readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
		newText, err = readFile(files[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", files[0], err)
			os.Exit(1)
		}
	} else {
		// Two files
		oldText, err = readFile(files[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", files[0], err)
			os.Exit(1)
		}
		newText, err = readFile(files[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", files[1], err)
			os.Exit(1)
		}
	}

	result := diff.Diff(oldText, newText)
	formatter := output.NewFormatter(format)
	formatter.Colorize = colorize
	formatter.Width = width
	formatter.ShowStats = showStats

	fmt.Print(formatter.Render(result))
}

func cmdCompare(args []string) {
	var files []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--help", "-h":
			printCompareUsage()
			return
		default:
			files = append(files, args[i])
		}
	}

	if len(files) != 2 {
		fmt.Fprintln(os.Stderr, "Error: exactly two files required")
		os.Exit(1)
	}

	oldText, err := readFile(files[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", files[0], err)
		os.Exit(1)
	}

	newText, err := readFile(files[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", files[1], err)
		os.Exit(1)
	}

	result := diff.Diff(oldText, newText)
	formatter := output.NewFormatter(output.FormatSummary)
	formatter.Colorize = true

	fmt.Print(formatter.Render(result))
}

func cmdParse(args []string) {
	showTypes := false
	var files []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--types", "-t":
			showTypes = true
		case "--help", "-h":
			printParseUsage()
			return
		default:
			files = append(files, args[i])
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Error: at least one file required")
		os.Exit(1)
	}

	for _, file := range files {
		text, err := readFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			os.Exit(1)
		}

		blocks := parser.Parse(text)

		if len(files) > 1 {
			fmt.Printf("--- %s (%d blocks) ---\n", file, len(blocks))
		}

		for i, b := range blocks {
			prefix := "  "
			if showTypes {
				prefix = fmt.Sprintf("%2d ", i)
			}
			fmt.Printf("%s[%s] %s\n", prefix, b.Type, truncateStr(parser.BlockToText(b), 72))
		}

		fmt.Printf("\nTotal blocks: %d\n", len(blocks))
		if len(files) > 1 {
			fmt.Println()
		}
	}
}

func printDiffUsage() {
	fmt.Print(`Usage: markdowndiff diff [options] <file1> [file2]

Compare two markdown files. If only one file is given, compares with stdin.

Options:
  --format, -f <format>   Output format: unified (default), side-by-side, json, summary
  --no-color              Disable colored output
  --no-stats              Hide statistics header
  --width, -w <width>     Column width for side-by-side (default: 80)
  --help, -h              Show this help
`)
}

func printCompareUsage() {
	fmt.Print(`Usage: markdowndiff compare <file1> <file2>

Compare two markdown files and show a high-level summary of changes.

Options:
  --help, -h    Show this help
`)
}

func printParseUsage() {
	fmt.Print(`Usage: markdowndiff parse [options] <file>

Parse a markdown file and display its block structure.

Options:
  --types, -t   Show block index numbers
  --help, -h    Show this help
`)
}

func readFile(path string) (string, error) {
	if path == "-" {
		return readStdin()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func truncateStr(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
