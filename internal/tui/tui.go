package tui

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhignan-rakshith/ccopy/internal/util"
	"github.com/charmbracelet/lipgloss"

	"github.com/abhignan-rakshith/ccopy/internal/formatter"
	"github.com/abhignan-rakshith/ccopy/internal/selection"
	"github.com/abhignan-rakshith/ccopy/internal/tree"
)

type mode int

const (
	modeBrowse mode = iota
	modeFilter
	modeConfirmLarge
	modeHelp
)

// Result is returned to the caller after the program exits. Copied=false means
// the user quit without committing a selection.
type Result struct {
	Copied    bool
	Files     []string
	TotalSize int64
	Output    string
}

// Model is the bubbletea model.
type Model struct {
	root      *tree.Node
	visible   []*tree.Node // flattened visible rows (root + expanded descendants)
	cursor    int
	selection *selection.Set
	format    string
	result    *Result
	err       error

	width, height int

	mode        mode
	filter      string
	confirmNode *tree.Node
}

// NewModel constructs a Model ready to be run.
func NewModel(root *tree.Node, format string) *Model {
	m := &Model{
		root:      root,
		selection: selection.New(),
		format:    format,
		result:    &Result{},
	}
	m.rebuildVisible()
	return m
}

// Run executes the TUI. On clean exit (either quit or commit), it returns the
// Result; on error the returned error is non-nil.
func Run(root *tree.Node, format string) (*Result, error) {
	m := NewModel(root, format)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := final.(*Model)
	if fm.err != nil {
		return nil, fm.err
	}
	return fm.result, nil
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch m.mode {
		case modeFilter:
			return m.updateFilter(msg)
		case modeConfirmLarge:
			return m.updateConfirm(msg)
		case modeHelp:
			return m.updateHelp(msg)
		default:
			return m.updateBrowse(msg)
		}
	}
	return m, nil
}

func (m *Model) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.visible)-1 {
			m.cursor++
		}
	case "left", "h":
		n := m.current()
		if n != nil && n.IsDir && n.Expanded {
			n.Expanded = false
			m.rebuildVisible()
		} else if n != nil && n.Parent != nil {
			// jump to parent
			for i, v := range m.visible {
				if v == n.Parent {
					m.cursor = i
					break
				}
			}
		}
	case "right", "l":
		n := m.current()
		if n != nil && n.IsDir && !n.Expanded {
			n.Expanded = true
			m.rebuildVisible()
		}
	case " ":
		n := m.current()
		if n == nil {
			break
		}
		if n.IsDir {
			m.selection.ToggleDir(n)
			break
		}
		if n.IsBinary {
			break
		}
		if n.TooLarge && !m.selection.Has(n.Path) {
			m.confirmNode = n
			m.mode = modeConfirmLarge
			break
		}
		m.selection.ToggleFile(n)
	case "a":
		m.selection.SelectAll(m.root)
	case "n":
		m.selection.Clear()
	case "i":
		m.selection.Invert(m.root)
	case "/":
		m.mode = modeFilter
		m.filter = ""
	case "?":
		m.mode = modeHelp
	case "enter":
		return m.commit()
	}
	return m, nil
}

func (m *Model) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.filter = ""
		m.mode = modeBrowse
		m.rebuildVisible()
	case "enter":
		m.mode = modeBrowse
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.rebuildVisible()
		}
	default:
		if len(msg.Runes) > 0 {
			m.filter += string(msg.Runes)
			m.rebuildVisible()
		}
	}
	return m, nil
}

func (m *Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if m.confirmNode != nil {
			m.selection.ToggleFile(m.confirmNode)
		}
		fallthrough
	case "n", "N", "esc":
		m.confirmNode = nil
		m.mode = modeBrowse
	}
	return m, nil
}

func (m *Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc", "q", "enter":
		m.mode = modeBrowse
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// commit reads the selected files, formats them, copies to clipboard via the
// caller (we only produce Output; copy is main's responsibility so the TUI
// stays clipboard-agnostic and testable).
func (m *Model) commit() (tea.Model, tea.Cmd) {
	paths := m.selection.Paths()
	sort.Strings(paths)
	if len(paths) == 0 {
		return m, tea.Quit
	}
	var buf bytes.Buffer
	if err := formatter.Format(&buf, paths, m.format); err != nil {
		m.err = err
		return m, tea.Quit
	}
	var total int64
	for _, p := range paths {
		if fi, err := os.Stat(p); err == nil {
			total += fi.Size()
		}
	}
	m.result.Copied = true
	m.result.Files = paths
	m.result.TotalSize = total
	m.result.Output = buf.String()
	return m, tea.Quit
}

func (m *Model) current() *tree.Node {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return nil
	}
	return m.visible[m.cursor]
}

func (m *Model) rebuildVisible() {
	if m.filter == "" {
		m.visible = tree.Flatten(m.root)
	} else {
		// Filter mode: only show files whose path contains the substring (case-insensitive),
		// and their ancestor directories. Ancestors are force-expanded.
		needle := strings.ToLower(m.filter)
		keep := map[*tree.Node]bool{}
		var walk func(n *tree.Node) bool
		walk = func(n *tree.Node) bool {
			any := false
			if !n.IsDir {
				if strings.Contains(strings.ToLower(n.Path), needle) {
					keep[n] = true
					any = true
				}
				return any
			}
			for _, c := range n.Children {
				if walk(c) {
					any = true
				}
			}
			if any {
				keep[n] = true
				n.Expanded = true
			}
			return any
		}
		walk(m.root)
		var out []*tree.Node
		var emit func(n *tree.Node)
		emit = func(n *tree.Node) {
			if !keep[n] {
				return
			}
			out = append(out, n)
			if !n.IsDir || !n.Expanded {
				return
			}
			for _, c := range n.Children {
				emit(c)
			}
		}
		emit(m.root)
		m.visible = out
	}
	if m.cursor >= len(m.visible) {
		m.cursor = len(m.visible) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// --- View ---

var (
	styleHeader    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	styleCursor    = lipgloss.NewStyle().Reverse(true)
	styleDir       = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	styleBinary    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleLarge     = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	styleSelected  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	styleStatus    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleHint      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleFilterBar = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

func (m *Model) View() string {
	if m.mode == modeHelp {
		return m.viewHelp()
	}
	var b strings.Builder
	cwd, _ := os.Getwd()
	fmt.Fprintln(&b, styleHeader.Render("ccopy")+"  "+styleStatus.Render(cwd))

	// Determine viewport size for body.
	bodyLines := m.height - 3 // header + status + optional filter
	if bodyLines < 5 {
		bodyLines = 5
	}

	// Ensure cursor is visible with a simple window.
	start := 0
	if m.cursor >= bodyLines {
		start = m.cursor - bodyLines + 1
	}
	end := start + bodyLines
	if end > len(m.visible) {
		end = len(m.visible)
	}

	for i := start; i < end; i++ {
		n := m.visible[i]
		line := m.renderRow(n)
		if i == m.cursor {
			line = styleCursor.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Pad out to body height so status bar stays pinned.
	for i := end - start; i < bodyLines; i++ {
		b.WriteString("\n")
	}

	// Filter bar
	if m.mode == modeFilter {
		fmt.Fprintln(&b, styleFilterBar.Render("/"+m.filter+"▌"))
	}

	// Confirm modal (inline)
	if m.mode == modeConfirmLarge && m.confirmNode != nil {
		fmt.Fprintln(&b, styleLarge.Render(fmt.Sprintf(
			"File is %s (over limit). Select anyway? [y/N]",
			util.HumanSize(m.confirmNode.Size),
		)))
	}

	// Status
	total := int64(0)
	for _, p := range m.selection.Paths() {
		if fi, err := os.Stat(p); err == nil {
			total += fi.Size()
		}
	}
	hint := "space:select  enter:copy  /:filter  ?:help  q:quit"
	fmt.Fprintf(&b, "%s  %s",
		styleStatus.Render(fmt.Sprintf("%d files  %s", m.selection.Len(), util.HumanSize(total))),
		styleHint.Render(hint))

	return b.String()
}

func (m *Model) renderRow(n *tree.Node) string {
	depth := nodeDepth(n, m.root)
	indent := strings.Repeat("  ", depth)
	box := m.checkbox(n)
	icon := "📄"
	name := n.Name
	var size string
	if n.IsDir {
		icon = "📁"
		if n.Expanded {
			icon = "📂"
		}
		name = styleDir.Render(name)
	} else {
		size = "  " + styleStatus.Render(util.HumanSize(n.Size))
	}
	line := fmt.Sprintf("%s%s %s %s%s", indent, box, icon, name, size)
	switch {
	case !n.IsDir && n.IsBinary:
		return styleBinary.Render(line)
	case !n.IsDir && n.TooLarge:
		return styleLarge.Render(line)
	case !n.IsDir && m.selection.Has(n.Path):
		return styleSelected.Render(line)
	}
	return line
}

func (m *Model) checkbox(n *tree.Node) string {
	if n.IsDir {
		switch m.selection.DirState(n) {
		case selection.StateAll:
			return "[x]"
		case selection.StatePartial:
			return "[~]"
		default:
			return "[ ]"
		}
	}
	if n.IsBinary {
		return "[-]"
	}
	if m.selection.Has(n.Path) {
		return "[x]"
	}
	return "[ ]"
}

func (m *Model) viewHelp() string {
	var b strings.Builder
	fmt.Fprintln(&b, styleHeader.Render("ccopy — keybindings"))
	fmt.Fprintln(&b)
	lines := [][2]string{
		{"↑/↓ or j/k", "move cursor"},
		{"←/→ or h/l", "collapse / expand directory"},
		{"space", "toggle selection (dir = select all under it)"},
		{"a", "select all"},
		{"n", "clear selection"},
		{"i", "invert selection"},
		{"/", "filter tree (esc to cancel)"},
		{"enter", "copy to clipboard and exit"},
		{"q / esc", "quit without copying"},
		{"?", "toggle this help"},
	}
	for _, l := range lines {
		fmt.Fprintf(&b, "  %-14s %s\n", l[0], l[1])
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, styleHint.Render("press ? or esc to return"))
	return b.String()
}

func nodeDepth(n, root *tree.Node) int {
	d := 0
	for cur := n; cur != nil && cur != root; cur = cur.Parent {
		d++
	}
	return d
}



