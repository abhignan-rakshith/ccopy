package tree

import (
	"os"
	"path/filepath"
	"testing"
)

// buildFS creates a temp directory tree from a map of relative paths to contents.
// A path ending in "/" is a directory; otherwise it's a file with the given bytes.
func buildFS(t *testing.T, layout map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range layout {
		full := filepath.Join(root, rel)
		if len(rel) > 0 && rel[len(rel)-1] == '/' {
			if err := os.MkdirAll(full, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func findChild(n *Node, name string) *Node {
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestWalkBasic(t *testing.T) {
	root := buildFS(t, map[string]string{
		"a.txt":       "alpha",
		"b.txt":       "bravo",
		"sub/c.txt":   "charlie",
		"sub/d.md":    "delta",
		"empty/":      "",
		".git/x":      "gitjunk",
		"node_modules/foo/bar.js": "junk",
	})

	n, err := Walk(root, Options{
		ExcludeDirs:      []string{".git", "node_modules"},
		RespectGitignore: false,
		MaxFileSize:      0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !n.IsDir {
		t.Fatal("root should be dir")
	}
	if findChild(n, ".git") != nil {
		t.Error(".git should be excluded")
	}
	if findChild(n, "node_modules") != nil {
		t.Error("node_modules should be excluded")
	}
	if findChild(n, "empty") != nil {
		t.Error("empty dirs should be pruned")
	}
	if findChild(n, "a.txt") == nil || findChild(n, "b.txt") == nil {
		t.Error("missing top-level files")
	}
	sub := findChild(n, "sub")
	if sub == nil || len(sub.Children) != 2 {
		t.Fatalf("sub children = %v", sub)
	}
	// Dirs come before files in sort.
	if n.Children[0].Name != "sub" {
		t.Errorf("expected sub first, got %s", n.Children[0].Name)
	}
}

func TestWalkGitignore(t *testing.T) {
	root := buildFS(t, map[string]string{
		".gitignore":       "ignored.txt\nbuild/\n",
		"keep.txt":         "k",
		"ignored.txt":      "i",
		"build/stuff.txt":  "s",
		"sub/.gitignore":   "*.log\n",
		"sub/app.txt":      "a",
		"sub/debug.log":    "d",
	})
	n, err := Walk(root, Options{RespectGitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	if findChild(n, "ignored.txt") != nil {
		t.Error("ignored.txt should be excluded")
	}
	if findChild(n, "build") != nil {
		t.Error("build/ should be excluded")
	}
	sub := findChild(n, "sub")
	if sub == nil {
		t.Fatal("sub missing")
	}
	if findChild(sub, "debug.log") != nil {
		t.Error("debug.log should be excluded by nested .gitignore")
	}
	if findChild(sub, "app.txt") == nil {
		t.Error("app.txt should be present")
	}
}

func TestBinaryAndLarge(t *testing.T) {
	root := buildFS(t, map[string]string{
		"text.txt": "hi",              // 2 bytes
		"bin.dat":  "ab\x00cd",
		"big.txt":  "xxxxxxxxxx",      // 10 bytes
	})
	n, err := Walk(root, Options{MaxFileSize: 5})
	if err != nil {
		t.Fatal(err)
	}
	bin := findChild(n, "bin.dat")
	if bin == nil || !bin.IsBinary {
		t.Errorf("bin.dat should be flagged binary, got %+v", bin)
	}
	big := findChild(n, "big.txt")
	if big == nil || !big.TooLarge {
		t.Errorf("big.txt should be flagged TooLarge, got %+v", big)
	}
	txt := findChild(n, "text.txt")
	if txt == nil || txt.IsBinary || txt.TooLarge {
		t.Errorf("text.txt flags wrong: %+v", txt)
	}
}

func TestFlatten(t *testing.T) {
	root := buildFS(t, map[string]string{
		"a.txt":     "a",
		"sub/b.txt": "b",
	})
	n, _ := Walk(root, Options{})
	sub := findChild(n, "sub")
	sub.Expanded = false
	flat := Flatten(n)
	// root + sub + a.txt
	if len(flat) != 3 {
		t.Fatalf("flatten count = %d, want 3: %v", len(flat), names(flat))
	}
	sub.Expanded = true
	flat = Flatten(n)
	if len(flat) != 4 {
		t.Fatalf("flatten expanded count = %d, want 4: %v", len(flat), names(flat))
	}
}

func TestFiles(t *testing.T) {
	root := buildFS(t, map[string]string{
		"a.txt":         "a",
		"sub/b.txt":     "b",
		"sub/x/c.txt":   "c",
	})
	n, _ := Walk(root, Options{})
	files := Files(n)
	if len(files) != 3 {
		t.Fatalf("files = %d: %v", len(files), names(files))
	}
}

func names(ns []*Node) []string {
	out := make([]string, len(ns))
	for i, n := range ns {
		out[i] = n.Name
	}
	return out
}
