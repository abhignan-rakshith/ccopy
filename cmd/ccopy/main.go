package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/atotto/clipboard"

	"github.com/abhignan-rakshith/ccopy/internal/config"
	"github.com/abhignan-rakshith/ccopy/internal/formatter"
	"github.com/abhignan-rakshith/ccopy/internal/tree"
	"github.com/abhignan-rakshith/ccopy/internal/tui"
	"github.com/abhignan-rakshith/ccopy/internal/util"
)

var version = "0.1.0-dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	noGitignore := flag.Bool("no-gitignore", false, "ignore .gitignore files during walk")
	maxSize := flag.String("max-size", "", "override max file size (e.g. 5MB)")
	dryRun := flag.Bool("dry-run", false, "print output to stdout instead of copying to clipboard")
	format := flag.String("format", "", "output format: tail | markdown | xml")
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Printf("ccopy %s\n", version)
		return
	}

	if err := run(runArgs{
		Path:        flag.Arg(0),
		NoGitignore: *noGitignore,
		MaxSize:     *maxSize,
		DryRun:      *dryRun,
		Format:      *format,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "ccopy:", err)
		os.Exit(1)
	}
}

type runArgs struct {
	Path        string
	NoGitignore bool
	MaxSize     string
	DryRun      bool
	Format      string
}

func run(a runArgs) error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ccopy: warning:", err)
	}

	if a.NoGitignore {
		cfg.RespectGitignore = false
	}
	if a.MaxSize != "" {
		n, err := config.ParseSize(a.MaxSize)
		if err != nil {
			return fmt.Errorf("--max-size: %w", err)
		}
		cfg.MaxFileSize = n
	}
	if a.Format != "" {
		cfg.Format = a.Format
	}

	path := a.Path
	if path == "" {
		path = "."
	}

	root, err := tree.Walk(path, tree.Options{
		ExcludeDirs:      cfg.ExcludePatterns,
		RespectGitignore: cfg.RespectGitignore,
		MaxFileSize:      cfg.MaxFileSize,
	})
	if err != nil {
		return err
	}

	if a.DryRun {
		paths := allSelectableFiles(root)
		sort.Strings(paths)
		return formatter.Format(os.Stdout, paths, cfg.Format)
	}

	result, err := tui.Run(root, cfg.Format)
	if err != nil {
		return err
	}
	if !result.Copied {
		return nil
	}
	if err := clipboard.WriteAll(result.Output); err != nil {
		// Fall back to stdout if clipboard is unavailable (e.g. headless Linux).
		fmt.Fprintln(os.Stderr, "ccopy: clipboard unavailable, writing to stdout:", err)
		if _, werr := os.Stdout.WriteString(result.Output); werr != nil {
			return werr
		}
		return nil
	}
	fmt.Printf("✓ Copied %d files (%s) to clipboard\n", len(result.Files), util.HumanSize(result.TotalSize))
	return nil
}

// allSelectableFiles returns paths of every non-binary, in-tree file — used for --dry-run.
func allSelectableFiles(n *tree.Node) []string {
	var out []string
	for _, f := range tree.Files(n) {
		if f.IsBinary {
			continue
		}
		out = append(out, f.Path)
	}
	return out
}


func usage() {
	fmt.Fprintf(os.Stderr, `ccopy — interactive TUI file picker that copies concatenated file contents to the clipboard.

Usage:
  ccopy [flags] [path]

Flags:
  --format <tail|markdown|xml>   output format (default: tail)
  --no-gitignore                 don't respect .gitignore
  --max-size <size>              override max file size, e.g. 5MB
  --dry-run                      print to stdout instead of clipboard
  --version                      print version and exit
  --help                         print this help

Keybindings (in TUI):
  ↑/↓ or j/k   move cursor          space  toggle selection
  ←/→ or h/l   collapse / expand    a      select all
  /            filter tree          n      clear selection
  enter        copy and exit        i      invert selection
  q / esc      quit                 ?      help
`)
}
