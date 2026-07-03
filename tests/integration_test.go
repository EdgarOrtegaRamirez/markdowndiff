package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "markdowndiff")
	// Find the project root (parent of tests/)
	projectRoot := filepath.Join("..")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/markdowndiff/")
	cmd.Dir = projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return binary
}

func TestVersion(t *testing.T) {
	binary := buildBinary(t)
	out, err := exec.Command(binary, "version").CombinedOutput()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if !strings.Contains(string(out), "markdowndiff v") {
		t.Errorf("expected version output, got: %s", out)
	}
}

func TestHelp(t *testing.T) {
	binary := buildBinary(t)
	out, err := exec.Command(binary, "help").CombinedOutput()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected usage info, got: %s", out)
	}
}

func TestDiffTwoFiles(t *testing.T) {
	binary := buildBinary(t)

	oldFile := filepath.Join(t.TempDir(), "old.md")
	newFile := filepath.Join(t.TempDir(), "new.md")

	os.WriteFile(oldFile, []byte("# Title\n\nOld content."), 0644)
	os.WriteFile(newFile, []byte("# Title\n\nNew content."), 0644)

	out, err := exec.Command(binary, "diff", oldFile, newFile).CombinedOutput()
	if err != nil {
		t.Fatalf("diff command failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "modified") && !strings.Contains(string(out), "+") && !strings.Contains(string(out), "-") {
		t.Errorf("expected diff output, got: %s", out)
	}
}

func TestDiffStdin(t *testing.T) {
	binary := buildBinary(t)

	newFile := filepath.Join(t.TempDir(), "new.md")
	os.WriteFile(newFile, []byte("# New"), 0644)

	cmd := exec.Command(binary, "diff", newFile)
	cmd.Stdin = strings.NewReader("# Old")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("diff stdin command failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Error("expected output from stdin diff")
	}
}

func TestDiffJSONFormat(t *testing.T) {
	binary := buildBinary(t)

	oldFile := filepath.Join(t.TempDir(), "old.md")
	newFile := filepath.Join(t.TempDir(), "new.md")

	os.WriteFile(oldFile, []byte("# A"), 0644)
	os.WriteFile(newFile, []byte("# B"), 0644)

	out, err := exec.Command(binary, "diff", "--format", "json", oldFile, newFile).CombinedOutput()
	if err != nil {
		t.Fatalf("diff json command failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), `"stats"`) {
		t.Errorf("expected JSON output with stats, got: %s", out)
	}
}

func TestDiffSummaryFormat(t *testing.T) {
	binary := buildBinary(t)

	oldFile := filepath.Join(t.TempDir(), "old.md")
	newFile := filepath.Join(t.TempDir(), "new.md")

	os.WriteFile(oldFile, []byte("# A\n\nB"), 0644)
	os.WriteFile(newFile, []byte("# A\n\nB\n\nC"), 0644)

	out, err := exec.Command(binary, "diff", "--format", "summary", oldFile, newFile).CombinedOutput()
	if err != nil {
		t.Fatalf("diff summary command failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Similarity") {
		t.Errorf("expected summary output, got: %s", out)
	}
}

func TestParseCommand(t *testing.T) {
	binary := buildBinary(t)

	mdFile := filepath.Join(t.TempDir(), "test.md")
	os.WriteFile(mdFile, []byte("# Heading\n\nParagraph.\n\n- List"), 0644)

	out, err := exec.Command(binary, "parse", mdFile).CombinedOutput()
	if err != nil {
		t.Fatalf("parse command failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "[heading]") {
		t.Errorf("expected block type in output, got: %s", out)
	}
	if !strings.Contains(string(out), "Total blocks:") {
		t.Errorf("expected total blocks count, got: %s", out)
	}
}

func TestCompareCommand(t *testing.T) {
	binary := buildBinary(t)

	oldFile := filepath.Join(t.TempDir(), "old.md")
	newFile := filepath.Join(t.TempDir(), "new.md")

	os.WriteFile(oldFile, []byte("# A"), 0644)
	os.WriteFile(newFile, []byte("# B"), 0644)

	out, err := exec.Command(binary, "compare", oldFile, newFile).CombinedOutput()
	if err != nil {
		t.Fatalf("compare command failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Similarity") {
		t.Errorf("expected similarity in compare output, got: %s", out)
	}
}

func TestDiffNoColor(t *testing.T) {
	binary := buildBinary(t)

	oldFile := filepath.Join(t.TempDir(), "old.md")
	newFile := filepath.Join(t.TempDir(), "new.md")

	os.WriteFile(oldFile, []byte("# A"), 0644)
	os.WriteFile(newFile, []byte("# B"), 0644)

	out, err := exec.Command(binary, "diff", "--no-color", oldFile, newFile).CombinedOutput()
	if err != nil {
		t.Fatalf("diff no-color command failed: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "\033[") {
		t.Error("expected no ANSI codes with --no-color")
	}
}

func TestDiffIdenticalFiles(t *testing.T) {
	binary := buildBinary(t)

	oldFile := filepath.Join(t.TempDir(), "old.md")
	newFile := filepath.Join(t.TempDir(), "new.md")

	content := "# Same\n\nContent."
	os.WriteFile(oldFile, []byte(content), 0644)
	os.WriteFile(newFile, []byte(content), 0644)

	out, err := exec.Command(binary, "diff", oldFile, newFile).CombinedOutput()
	if err != nil {
		t.Fatalf("diff identical command failed: %v\n%s", err, out)
	}
	// Should show 100% match
	if !strings.Contains(string(out), "100%") {
		t.Errorf("expected 100%% match for identical files, got: %s", out)
	}
}

func TestUnknownCommand(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "unknown")
	err := cmd.Run()
	if err == nil {
		t.Error("expected error for unknown command")
	}
}
