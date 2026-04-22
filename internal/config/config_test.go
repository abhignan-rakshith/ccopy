package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	c := Default()
	if c.MaxFileSize != 1<<20 {
		t.Errorf("MaxFileSize default = %d, want %d", c.MaxFileSize, 1<<20)
	}
	if !c.RespectGitignore {
		t.Error("RespectGitignore default should be true")
	}
	if c.Format != "tail" {
		t.Errorf("Format default = %q, want tail", c.Format)
	}
	if len(c.ExcludePatterns) == 0 {
		t.Error("ExcludePatterns should have defaults")
	}
}

func TestLoadFromMissing(t *testing.T) {
	c, err := LoadFrom(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatalf("LoadFrom missing: %v", err)
	}
	if c.Format != "tail" {
		t.Errorf("fallback to defaults failed: %+v", c)
	}
}

func TestLoadFromOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.toml")
	body := `
format = "xml"
max_file_size = 2048
respect_gitignore = false
exclude_patterns = ["foo", "bar"]
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if c.Format != "xml" {
		t.Errorf("Format = %q, want xml", c.Format)
	}
	if c.MaxFileSize != 2048 {
		t.Errorf("MaxFileSize = %d", c.MaxFileSize)
	}
	if c.RespectGitignore {
		t.Error("RespectGitignore should be false")
	}
	if len(c.ExcludePatterns) != 2 || c.ExcludePatterns[0] != "foo" {
		t.Errorf("ExcludePatterns = %v", c.ExcludePatterns)
	}
}

func TestLoadFromPartial(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.toml")
	if err := os.WriteFile(path, []byte(`format = "markdown"`), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.Format != "markdown" {
		t.Errorf("Format = %q", c.Format)
	}
	// Unspecified keys keep defaults.
	if c.MaxFileSize != 1<<20 {
		t.Errorf("MaxFileSize = %d, want default", c.MaxFileSize)
	}
	if !c.RespectGitignore {
		t.Error("RespectGitignore should stay default true")
	}
}

func TestParseSize(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		err  bool
	}{
		{"1024", 1024, false},
		{"1KB", 1024, false},
		{"2MB", 2 << 20, false},
		{"1GB", 1 << 30, false},
		{"512B", 512, false},
		{"  5 MB", 5 << 20, false},
		{"1.5MB", int64(1.5 * float64(1<<20)), false},
		{"", 0, true},
		{"-1", 0, true},
		{"abc", 0, true},
	}
	for _, tc := range cases {
		got, err := ParseSize(tc.in)
		if tc.err {
			if err == nil {
				t.Errorf("ParseSize(%q) expected error, got %d", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseSize(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseSize(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
