package selection

import (
	"testing"

	"github.com/abhignan-rakshith/ccopy/internal/tree"
)

// helper builds a small synthetic tree without touching the filesystem.
func buildTree() *tree.Node {
	root := &tree.Node{Path: "/r", Name: "r", IsDir: true}
	a := &tree.Node{Path: "/r/a.txt", Name: "a.txt", Parent: root}
	b := &tree.Node{Path: "/r/b.txt", Name: "b.txt", Parent: root}
	sub := &tree.Node{Path: "/r/sub", Name: "sub", IsDir: true, Parent: root}
	c := &tree.Node{Path: "/r/sub/c.txt", Name: "c.txt", Parent: sub}
	d := &tree.Node{Path: "/r/sub/d.bin", Name: "d.bin", IsBinary: true, Parent: sub}
	sub.Children = []*tree.Node{c, d}
	root.Children = []*tree.Node{sub, a, b}
	return root
}

func TestToggleFileAndDir(t *testing.T) {
	root := buildTree()
	s := New()
	a := root.Children[1] // a.txt
	s.ToggleFile(a)
	if !s.Has(a.Path) {
		t.Fatal("a should be selected")
	}
	s.ToggleFile(a)
	if s.Has(a.Path) {
		t.Fatal("a should be deselected")
	}

	// Toggling a dir with none selected selects all selectable files under it.
	sub := root.Children[0]
	s.ToggleDir(sub)
	if !s.Has("/r/sub/c.txt") {
		t.Error("c.txt should be selected")
	}
	if s.Has("/r/sub/d.bin") {
		t.Error("binary should NOT be selected")
	}
	if s.DirState(sub) != StateAll {
		t.Errorf("sub state = %v, want All", s.DirState(sub))
	}
	// Toggling again from StateAll deselects.
	s.ToggleDir(sub)
	if s.Has("/r/sub/c.txt") {
		t.Error("c.txt should be deselected")
	}
	if s.DirState(sub) != StateNone {
		t.Errorf("sub state after deselect = %v", s.DirState(sub))
	}
}

func TestPartialDirToggleSelectsAll(t *testing.T) {
	root := buildTree()
	s := New()
	s.ToggleFile(root.Children[1]) // a.txt only
	// root is now Partial (1 of 3 selectable files).
	if s.DirState(root) != StatePartial {
		t.Fatalf("root state = %v, want Partial", s.DirState(root))
	}
	s.ToggleDir(root)
	if s.DirState(root) != StateAll {
		t.Errorf("partial→toggle should yield All, got %v", s.DirState(root))
	}
	if !s.Has("/r/b.txt") || !s.Has("/r/sub/c.txt") {
		t.Error("all text files should be selected")
	}
}

func TestSelectAllClearInvert(t *testing.T) {
	root := buildTree()
	s := New()
	s.SelectAll(root)
	if s.Len() != 3 {
		t.Errorf("SelectAll len = %d, want 3", s.Len())
	}
	s.Clear()
	if s.Len() != 0 {
		t.Error("Clear failed")
	}
	s.ToggleFile(root.Children[1]) // a selected
	s.Invert(root)
	// a should now be unselected, b + c selected
	if s.Has("/r/a.txt") {
		t.Error("a should be unselected after invert")
	}
	if !s.Has("/r/b.txt") || !s.Has("/r/sub/c.txt") {
		t.Error("b and c should be selected after invert")
	}
}

func TestBinaryIgnored(t *testing.T) {
	root := buildTree()
	s := New()
	d := root.Children[0].Children[1] // d.bin
	s.ToggleFile(d)
	if s.Has(d.Path) {
		t.Error("binary should not be selectable via ToggleFile")
	}
}

func TestDirStateNoneWhenNoFiles(t *testing.T) {
	empty := &tree.Node{Path: "/r", Name: "r", IsDir: true}
	s := New()
	if s.DirState(empty) != StateNone {
		t.Error("empty dir state should be None")
	}
}
