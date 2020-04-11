package main

import (
	"math"
	"sort"
	"strings"

	term "github.com/wvanlint/twf/terminal"
)

type TreeView struct {
	state        *AppState
	pathsByIndex map[string]int
	paths        []string
	cursorLine   int
	rows         int
}

func (tv *TreeView) Position(totalRows int, totalCols int) term.Position {
	if tv.state.Root.FindPath(tv.state.Cursor).IsDir() {
		return term.Position{
			Top:  1,
			Left: 1,
			Rows: totalRows - 1,
			Cols: totalCols,
		}
	} else {
		return term.Position{
			Top:  1,
			Left: 1,
			Rows: totalRows - 1,
			Cols: int(math.Ceil(float64(totalCols) / 2.0)),
		}
	}
}

func (tv *TreeView) HasBorder() bool {
	return false
}

func (tv *TreeView) ShouldRender() bool {
	return true
}

func (tv *TreeView) renderNode(
	node *Tree,
	indentation int,
	maxLength int,
) term.Line {
	line := term.NewLine(&term.Graphics{}, maxLength)
	line.Append(strings.Repeat("  ", indentation), nil)

	graphics := term.Graphics{}
	if node.IsDir() {
		graphics.FgColor, _ = term.ColorFromString("brightblue")
		graphics.Bold = true
	}
	if node.Path == tv.state.Cursor {
		graphics.Reverse = true
	}
	line.Append(node.info.Name(), &graphics)
	return line
}

func (tv *TreeView) Render(p term.Position) []term.Line {
	lines := []term.Line{}
	tv.rows = p.Rows
	tv.pathsByIndex = make(map[string]int)
	tv.paths = []string{}
	type Item struct {
		tree  *Tree
		depth int
	}
	stack := []Item{Item{tv.state.Root, 0}}
	for len(stack) > 0 {
		var item Item
		item, stack = stack[len(stack)-1], stack[:len(stack)-1]
		line := tv.renderNode(item.tree, item.depth, p.Cols)
		tv.pathsByIndex[item.tree.Path] = len(lines)
		tv.paths = append(tv.paths, item.tree.Path)
		if item.tree.Path == tv.state.Cursor {
			tv.cursorLine = len(lines)
		}
		lines = append(lines, line)

		if value, _ := tv.state.Expansions[item.tree.Path]; value {
			children := item.tree.Children()
			sort.Slice(children, ByTypeAndName(children))
			for i := len(children) - 1; i >= 0; i-- {
				stack = append(stack, Item{children[i], item.depth + 1})
			}
		}
	}
	return lines[tv.state.Scroll:]
}

func (tv *TreeView) GetNextPath() string {
	if tv.cursorLine == len(tv.paths)-1 {
		return tv.paths[len(tv.paths)-1]
	} else {
		return tv.paths[tv.cursorLine+1]
	}
}

func (tv *TreeView) GetPrevPath() string {
	if tv.cursorLine == 0 {
		return tv.paths[0]
	} else {
		return tv.paths[tv.cursorLine-1]
	}
}

func (tv *TreeView) ScrollForPath(path string) int {
	targetLine := tv.pathsByIndex[path]
	if targetLine < tv.state.Scroll {
		return targetLine
	} else if targetLine >= tv.state.Scroll+tv.rows {
		return targetLine - tv.rows + 1
	} else {
		return tv.state.Scroll
	}
}
