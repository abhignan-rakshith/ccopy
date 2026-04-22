package selection

import "github.com/abhignan-rakshith/ccopy/internal/tree"

// DirState is the tri-state of a directory node based on its descendant files.
type DirState int

const (
	StateNone    DirState = iota // no descendant files selected
	StatePartial                 // some descendant files selected
	StateAll                     // every selectable descendant file selected
)

// Set holds the absolute paths of selected files.
// Directories are never stored directly — selection of a directory is shorthand
// for selecting every file under it.
type Set struct {
	paths map[string]struct{}
}

func New() *Set { return &Set{paths: map[string]struct{}{}} }

func (s *Set) Has(path string) bool {
	_, ok := s.paths[path]
	return ok
}

func (s *Set) Add(path string)    { s.paths[path] = struct{}{} }
func (s *Set) Remove(path string) { delete(s.paths, path) }
func (s *Set) Len() int           { return len(s.paths) }

// Paths returns the selected paths. The order is unspecified; callers that need
// determinism should sort.
func (s *Set) Paths() []string {
	out := make([]string, 0, len(s.paths))
	for p := range s.paths {
		out = append(out, p)
	}
	return out
}

// ToggleFile toggles a single file. No-op on directories or unselectable files.
func (s *Set) ToggleFile(n *tree.Node) {
	if n.IsDir || !selectable(n) {
		return
	}
	if s.Has(n.Path) {
		s.Remove(n.Path)
	} else {
		s.Add(n.Path)
	}
}

// ToggleDir applies the spec rule: if the dir is fully selected, deselect all;
// otherwise (None or Partial) select all selectable descendants.
func (s *Set) ToggleDir(n *tree.Node) {
	if !n.IsDir {
		return
	}
	switch s.DirState(n) {
	case StateAll:
		for _, f := range tree.Files(n) {
			s.Remove(f.Path)
		}
	default:
		for _, f := range tree.Files(n) {
			if selectable(f) {
				s.Add(f.Path)
			}
		}
	}
}

// DirState computes the tri-state for a directory. Files that aren't selectable
// (binary) are ignored when computing "all" so a folder of text files next to
// a binary still reports StateAll when the text files are all selected.
func (s *Set) DirState(n *tree.Node) DirState {
	if !n.IsDir {
		if s.Has(n.Path) {
			return StateAll
		}
		return StateNone
	}
	total, selected := 0, 0
	for _, f := range tree.Files(n) {
		if !selectable(f) {
			continue
		}
		total++
		if s.Has(f.Path) {
			selected++
		}
	}
	if total == 0 {
		return StateNone
	}
	if selected == 0 {
		return StateNone
	}
	if selected == total {
		return StateAll
	}
	return StatePartial
}

// SelectAll selects every selectable file in the tree rooted at n.
func (s *Set) SelectAll(n *tree.Node) {
	for _, f := range tree.Files(n) {
		if selectable(f) {
			s.Add(f.Path)
		}
	}
}

// Clear deselects everything.
func (s *Set) Clear() {
	for k := range s.paths {
		delete(s.paths, k)
	}
}

// Invert flips selection for every selectable file under n.
func (s *Set) Invert(n *tree.Node) {
	for _, f := range tree.Files(n) {
		if !selectable(f) {
			continue
		}
		if s.Has(f.Path) {
			s.Remove(f.Path)
		} else {
			s.Add(f.Path)
		}
	}
}

// selectable returns true if a file node is eligible for selection.
// Binary files are not selectable; oversized files are (they go through a confirm flow in the TUI).
func selectable(n *tree.Node) bool {
	if n.IsDir {
		return false
	}
	return !n.IsBinary
}
