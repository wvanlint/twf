package main

import (
	"sort"
	"strings"

	term "github.com/wvanlint/twf/terminal"
)

type TreeView struct {
	state         *AppState
	renderedPaths []string
	cursorLine    int
}

func (tv *TreeView) Dims() (float32, float32) {
	return 1.0, 1.0
}

func (tv *TreeView) HasBorder() bool {
	return false
}

func (tv *TreeView) renderNode(node *Tree, indentation int, selected bool) term.Line {
	line := term.NewLine(&term.Graphics{})
	line.Append(strings.Repeat("  ", indentation), nil)

	graphics := term.Graphics{}
	if node.IsDir() {
		graphics.FgColor, _ = term.ColorFromString("brightblue")
		graphics.Bold = true
	}
	if selected {
		graphics.Reverse = true
	}
	line.Append(node.info.Name(), &graphics)
	return line
}

func (tv *TreeView) Render() ([]term.Line, bool) {
	lines := []term.Line{}
	tv.renderedPaths = []string{}
	type Item struct {
		tree  *Tree
		depth int
	}
	stack := []Item{Item{tv.state.Root, 0}}
	for len(stack) > 0 {
		var item Item
		item, stack = stack[len(stack)-1], stack[:len(stack)-1]
		line := tv.renderNode(item.tree, item.depth, item.tree.Path == tv.state.Cursor)
		tv.renderedPaths = append(tv.renderedPaths, item.tree.Path)
		lines = append(lines, line)

		if item.tree.Path == tv.state.Cursor {
			tv.cursorLine = len(lines) - 1
		}

		if value, _ := tv.state.Expansions[item.tree.Path]; value {
			children := item.tree.Children()
			sort.Slice(children, ByTypeAndName(children))
			for i := len(children) - 1; i >= 0; i-- {
				stack = append(stack, Item{children[i], item.depth + 1})
			}
		}
	}
	return lines, true
}

func (tv *TreeView) GetNextPath() string {
	if tv.cursorLine == len(tv.renderedPaths)-1 {
		return tv.renderedPaths[len(tv.renderedPaths)-1]
	} else {
		return tv.renderedPaths[tv.cursorLine+1]
	}
}

func (tv *TreeView) GetPrevPath() string {
	if tv.cursorLine == 0 {
		return tv.renderedPaths[0]
	} else {
		return tv.renderedPaths[tv.cursorLine-1]
	}
}
