package formatter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Format writes entries to w using the named format ("tail", "markdown", "xml").
// Files are read fresh from disk; a read error is written in place of the content,
// prefixed with "<error: ...>", so a single unreadable file never aborts the run.
func Format(w io.Writer, paths []string, format string) error {
	fn, err := pick(format)
	if err != nil {
		return err
	}
	for i, p := range paths {
		content := readFile(p)
		if err := fn(w, p, content, i == 0); err != nil {
			return err
		}
	}
	return nil
}

type writeFn func(w io.Writer, path, content string, first bool) error

func pick(format string) (writeFn, error) {
	switch format {
	case "", "tail":
		return writeTail, nil
	case "markdown", "md":
		return writeMarkdown, nil
	case "xml":
		return writeXML, nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}

func writeTail(w io.Writer, path, content string, first bool) error {
	if !first {
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "==> %s <==\n%s\n", path, content)
	return err
}

func writeMarkdown(w io.Writer, path, content string, first bool) error {
	if !first {
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}
	lang := langHint(path)
	_, err := fmt.Fprintf(w, "## %s\n\n```%s\n%s\n```\n", path, lang, content)
	return err
}

func writeXML(w io.Writer, path, content string, first bool) error {
	if !first {
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "<file path=%q>\n%s\n</file>\n", path, content)
	return err
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("<error: %s>", err)
	}
	return string(data)
}

// langHint maps a file path to a markdown fenced-code language hint.
// Empty string is returned for unknown extensions (valid markdown — just no highlighting).
func langHint(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	switch ext {
	case "go":
		return "go"
	case "js", "mjs", "cjs":
		return "javascript"
	case "ts":
		return "typescript"
	case "tsx":
		return "tsx"
	case "jsx":
		return "jsx"
	case "py":
		return "python"
	case "rs":
		return "rust"
	case "rb":
		return "ruby"
	case "java":
		return "java"
	case "kt":
		return "kotlin"
	case "swift":
		return "swift"
	case "c", "h":
		return "c"
	case "cpp", "cc", "cxx", "hpp":
		return "cpp"
	case "cs":
		return "csharp"
	case "sh", "bash", "zsh":
		return "bash"
	case "yml", "yaml":
		return "yaml"
	case "json":
		return "json"
	case "toml":
		return "toml"
	case "md":
		return "markdown"
	case "html", "htm":
		return "html"
	case "css":
		return "css"
	case "sql":
		return "sql"
	case "xml":
		return "xml"
	case "dart":
		return "dart"
	}
	return ""
}
