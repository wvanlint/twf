package views

import (
	"math"
	"sort"
	"strings"

	"github.com/wvanlint/twf/internal/config"
	"github.com/wvanlint/twf/internal/filetree"
	"github.com/wvanlint/twf/internal/state"
	term "github.com/wvanlint/twf/internal/terminal"
)

type treeView struct {
	config     *config.TwfConfig
	state      *state.State
	lineByPath map[string]int
	paths      []string
	cursorLine int
	rows       int
	scroll     int
}

func NewTreeView(config *config.TwfConfig, state *state.State) term.View {
	return &treeView{
		config: config,
		state:  state,
	}
}

func (v *treeView) Position(totalRows int, totalCols int) term.Position {
	cursorNode, _ := v.state.Root.FindPath(v.state.Cursor)
	if v.config.Preview.Enabled && !cursorNode.IsDir() {
		return term.Position{
			Top:  1,
			Left: 1,
			Rows: totalRows - 1,
			Cols: int(math.Ceil(float64(totalCols) / 2.0)),
		}
	} else {
		return term.Position{
			Top:  1,
			Left: 1,
			Rows: totalRows - 1,
			Cols: totalCols,
		}
	}
}

func (v *treeView) HasBorder() bool {
	return false
}

func (v *treeView) ShouldRender() bool {
	return true
}

func (v *treeView) renderNode(
	node *filetree.FileTree,
	indentation int,
	maxLength int,
) term.Line {
	line := term.NewLine(&term.Graphics{}, maxLength)
	line.Append(strings.Repeat("  ", indentation), nil)

	graphics := term.Graphics{}
	if node.IsDir() {
		if g, ok := v.config.Graphics["tree:dir"]; ok {
			graphics.Merge(g)
		}
	}
	if node.AbsPath == v.state.Cursor {
		if g, ok := v.config.Graphics["tree:cursor"]; ok {
			graphics.Merge(g)
		}
	}

	if node.IsDir() {
		if value, _ := v.state.Expansions[node.AbsPath]; value {
			line.Append("▼ ", &graphics)
		} else {
			line.Append("▶ ", &graphics)
		}
	}
	line.Append(node.Name(), &graphics)
	return line
}

func (v *treeView) Render(p term.Position) []term.Line {
	lines := []term.Line{}
	v.rows = p.Rows
	v.lineByPath = make(map[string]int)
	v.paths = []string{}
	type Item struct {
		tree  *filetree.FileTree
		depth int
	}
	stack := []Item{Item{v.state.Root, 0}}
	for len(stack) > 0 {
		var item Item
		item, stack = stack[len(stack)-1], stack[:len(stack)-1]
		line := v.renderNode(item.tree, item.depth, p.Cols)
		v.lineByPath[item.tree.AbsPath] = len(lines)
		v.paths = append(v.paths, item.tree.AbsPath)
		if item.tree.AbsPath == v.state.Cursor {
			v.cursorLine = len(lines)
		}
		lines = append(lines, line)

		if value, _ := v.state.Expansions[item.tree.AbsPath]; value {
			children, _ := item.tree.Children()
			sort.Slice(children, filetree.ByTypeAndName(children))
			for i := len(children) - 1; i >= 0; i-- {
				stack = append(stack, Item{children[i], item.depth + 1})
			}
		}
	}
	return lines[v.scroll:]
}

func (v *treeView) getNextPath() string {
	if v.cursorLine == len(v.paths)-1 {
		return v.paths[len(v.paths)-1]
	} else {
		return v.paths[v.cursorLine+1]
	}
}

func (v *treeView) getPrevPath() string {
	if v.cursorLine == 0 {
		return v.paths[0]
	} else {
		return v.paths[v.cursorLine-1]
	}
}

func (v *treeView) scrollForPath(path string) int {
	targetLine := v.lineByPath[path]
	if targetLine < v.scroll {
		return targetLine
	} else if targetLine >= v.scroll+v.rows {
		return targetLine - v.rows + 1
	} else {
		return v.scroll
	}
}

func (v *treeView) GetCommands() map[string]term.Command {
	return map[string]term.Command{
		"tree:prev":           v.prev,
		"tree:next":           v.next,
		"tree:open":           v.open,
		"tree:close":          v.close,
		"tree:toggle":         v.toggle,
		"tree:toggleAll":      v.toggleAll,
		"tree:openAll":        v.openAll,
		"tree:closeAll":       v.closeAll,
		"tree:parent":         v.parent,
		"tree:locateExternal": v.locateExternal,
		"tree:selectPath":     v.selectPath,
	}
}

func (v *treeView) selectPath(helper term.TerminalHelper, args ...interface{}) {
	v.state.AddSelection(v.state.Cursor)
}

func (v *treeView) prev(helper term.TerminalHelper, args ...interface{}) {
	p := v.getPrevPath()
	v.scroll = v.scrollForPath(p)
	v.state.ChangeCursor(p)
}

func (v *treeView) next(helper term.TerminalHelper, args ...interface{}) {
	p := v.getNextPath()
	v.scroll = v.scrollForPath(p)
	v.state.ChangeCursor(p)
}

func (v *treeView) open(helper term.TerminalHelper, args ...interface{}) {
	v.state.SetExpansion(v.state.Cursor, true)
}

func (v *treeView) close(helper term.TerminalHelper, args ...interface{}) {
	v.state.SetExpansion(v.state.Cursor, false)
}

func (v *treeView) toggle(helper term.TerminalHelper, args ...interface{}) {
	v.state.ToggleExpansion(v.state.Cursor)
}

func (v *treeView) toggleAll(helper term.TerminalHelper, args ...interface{}) {
	v.state.ToggleExpansionAll(v.state.Cursor)
}

func (v *treeView) openAll(helper term.TerminalHelper, args ...interface{}) {
	v.state.SetExpansionAll(v.state.Cursor, true)
}

func (v *treeView) closeAll(helper term.TerminalHelper, args ...interface{}) {
	v.state.SetExpansionAll(v.state.Cursor, false)
}

func (v *treeView) parent(helper term.TerminalHelper, args ...interface{}) {
	v.state.MoveCursorToParent()
}

func (v *treeView) locateExternal(helper term.TerminalHelper, args ...interface{}) {
	content, _ := helper.ExecuteInTerminal(v.config.TreeView.LocateCommand)
	v.state.ChangeCursor(strings.TrimSpace(content))
}
