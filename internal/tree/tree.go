package tree

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// Node represents a file or directory in the walked tree.
type Node struct {
	Path     string // absolute path
	Name     string // base name
	IsDir    bool
	Size     int64
	IsBinary bool // files only — directories always false
	TooLarge bool // files only — exceeded MaxFileSize threshold
	Children []*Node
	Parent   *Node
	Expanded bool // UI state; default false (root forced to true by walker)
}

// Options controls Walk behavior.
type Options struct {
	ExcludeDirs      []string // basenames to skip entirely, e.g. "node_modules"
	RespectGitignore bool
	MaxFileSize      int64 // files larger are marked TooLarge (still included)
}

// Walk builds a tree rooted at root. Errors from individual entries (e.g. permission
// denied) are skipped silently; only failures to read the root return an error.
func Walk(root string, opts Options) (*Node, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, err
	}
	excludeSet := make(map[string]struct{}, len(opts.ExcludeDirs))
	for _, d := range opts.ExcludeDirs {
		excludeSet[d] = struct{}{}
	}
	rootNode := &Node{
		Path:     absRoot,
		Name:     filepath.Base(absRoot),
		IsDir:    info.IsDir(),
		Expanded: true,
	}
	if !info.IsDir() {
		rootNode.Size = info.Size()
		rootNode.IsBinary = detectBinary(absRoot)
		rootNode.TooLarge = opts.MaxFileSize > 0 && info.Size() > opts.MaxFileSize
		return rootNode, nil
	}
	walkDir(rootNode, absRoot, excludeSet, opts, nil)
	return rootNode, nil
}

// ignoreFrame is one level of a .gitignore stack, rooted at Dir.
type ignoreFrame struct {
	Dir     string
	Matcher *gitignore.GitIgnore
}

func walkDir(parent *Node, dir string, excludeSet map[string]struct{}, opts Options, stack []ignoreFrame) {
	if opts.RespectGitignore {
		if m := loadGitignore(dir); m != nil {
			stack = append(stack, ignoreFrame{Dir: dir, Matcher: m})
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.IsDir() != b.IsDir() {
			return a.IsDir() // directories first
		}
		return strings.ToLower(a.Name()) < strings.ToLower(b.Name())
	})
	for _, e := range entries {
		name := e.Name()
		full := filepath.Join(dir, name)
		if _, skip := excludeSet[name]; skip {
			continue
		}
		if opts.RespectGitignore && matchedByStack(stack, full, e.IsDir()) {
			continue
		}
		if e.IsDir() {
			child := &Node{Path: full, Name: name, IsDir: true, Parent: parent}
			walkDir(child, full, excludeSet, opts, stack)
			// Skip empty directories after filtering — they add noise.
			if len(child.Children) == 0 {
				continue
			}
			parent.Children = append(parent.Children, child)
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		node := &Node{
			Path:     full,
			Name:     name,
			Size:     info.Size(),
			Parent:   parent,
			IsBinary: detectBinary(full),
		}
		if opts.MaxFileSize > 0 && info.Size() > opts.MaxFileSize {
			node.TooLarge = true
		}
		parent.Children = append(parent.Children, node)
	}
}

func loadGitignore(dir string) *gitignore.GitIgnore {
	path := filepath.Join(dir, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(data), "\n")
	m := gitignore.CompileIgnoreLines(lines...)
	return m
}

func matchedByStack(stack []ignoreFrame, path string, isDir bool) bool {
	for _, f := range stack {
		rel, err := filepath.Rel(f.Dir, path)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		probe := rel
		if isDir {
			probe = rel + "/"
		}
		if f.Matcher.MatchesPath(probe) {
			return true
		}
	}
	return false
}

// detectBinary returns true if the first 512 bytes of the file contain a null byte.
// A read error is treated as "not binary" to avoid silently hiding files.
func detectBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	var buf [512]byte
	n, err := f.Read(buf[:])
	if err != nil && err != io.EOF {
		return false
	}
	return bytes.IndexByte(buf[:n], 0) >= 0
}

// Flatten returns the visible nodes in depth-first order, honoring Expanded state
// on directories. The root itself is included.
func Flatten(root *Node) []*Node {
	var out []*Node
	var visit func(n *Node)
	visit = func(n *Node) {
		out = append(out, n)
		if !n.IsDir || !n.Expanded {
			return
		}
		for _, c := range n.Children {
			visit(c)
		}
	}
	visit(root)
	return out
}

// Files returns all non-directory descendants of n (including n itself if it's a file).
func Files(n *Node) []*Node {
	var out []*Node
	var visit func(n *Node)
	visit = func(n *Node) {
		if !n.IsDir {
			out = append(out, n)
			return
		}
		for _, c := range n.Children {
			visit(c)
		}
	}
	visit(n)
	return out
}
