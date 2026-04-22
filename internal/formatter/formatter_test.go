package formatter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFiles(t *testing.T, files map[string]string) []string {
	t.Helper()
	dir := t.TempDir()
	var paths []string
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p)
	}
	return paths
}

func TestFormatTail(t *testing.T) {
	paths := writeFiles(t, map[string]string{"a.txt": "alpha", "b.go": "package x"})
	// Order of map iteration isn't deterministic; test Format on one file at a time.
	var buf bytes.Buffer
	if err := Format(&buf, []string{paths[0]}, "tail"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "==> "+paths[0]+" <==\n") {
		t.Errorf("tail header missing: %q", out)
	}
}

func TestFormatMarkdown(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.go")
	if err := os.WriteFile(p, []byte("package x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Format(&buf, []string{p}, "markdown"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "## "+p+"\n") {
		t.Errorf("md heading missing: %q", out)
	}
	if !strings.Contains(out, "```go\n") {
		t.Errorf("md lang hint missing: %q", out)
	}
}

func TestFormatXML(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(p, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Format(&buf, []string{p}, "xml"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `<file path="`+p+`">`) {
		t.Errorf("xml open tag missing: %q", out)
	}
	if !strings.HasSuffix(strings.TrimRight(out, "\n"), "</file>") {
		t.Errorf("xml close tag missing: %q", out)
	}
}

func TestFormatReadError(t *testing.T) {
	var buf bytes.Buffer
	if err := Format(&buf, []string{"/nonexistent/xyz"}, "tail"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "<error:") {
		t.Errorf("expected error marker, got: %q", buf.String())
	}
}

func TestFormatUnknown(t *testing.T) {
	var buf bytes.Buffer
	if err := Format(&buf, nil, "bogus"); err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestLangHint(t *testing.T) {
	cases := map[string]string{
		"x.go": "go", "x.py": "python", "x.ts": "typescript",
		"x.unknown": "", "noext": "", "x.TOML": "toml",
	}
	for in, want := range cases {
		if got := langHint(in); got != want {
			t.Errorf("langHint(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMultipleFilesSeparator(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a")
	p2 := filepath.Join(dir, "b")
	os.WriteFile(p1, []byte("A"), 0o644)
	os.WriteFile(p2, []byte("B"), 0o644)
	var buf bytes.Buffer
	if err := Format(&buf, []string{p1, p2}, "tail"); err != nil {
		t.Fatal(err)
	}
	// Expect blank line between entries.
	if !strings.Contains(buf.String(), "A\n\n==> ") {
		t.Errorf("missing separator between entries: %q", buf.String())
	}
}
