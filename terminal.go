package main

import (
	sys "golang.org/x/sys/unix"
	"os"
	"sort"
	"strings"
)

const (
	escape    = "\x1b["
	high      = "h"
	low       = "l"
	altbuf    = "?1049"
	cursor    = "?25"
	termclear = "2J"
	clearline = "2K"
	jump      = "H"

	bold    = "1"
	no      = "2"
	faint   = "2"
	fg      = "3"
	br_fg   = "9"
	bg      = "4"
	br_bg   = "10"
	with    = ";"
	plain   = ""
	black   = "0"
	red     = "1"
	green   = "2"
	yellow  = "3"
	blue    = "4"
	magenta = "5"
	cyan    = "6"
	white   = "7"
	style   = "m"
)

type Terminal struct {
	originalTermios sys.Termios
}

func InitTerm() (*Terminal, error) {
	termios, err := sys.IoctlGetTermios(1, sys.TIOCGETA)
	if err != nil {
		return nil, err
	}
	term := Terminal{originalTermios: *termios}

	termios.Lflag = termios.Lflag & (^uint64(sys.ECHO) & ^uint64(sys.ICANON))
	sys.IoctlSetTermios(1, sys.TIOCSETA, termios)
	os.Stdout.WriteString(escape + altbuf + high)
	os.Stdout.WriteString(escape + cursor + low)

	return &term, nil
}

func DefaultLayout() string {
	return escape + plain + style + escape + br_fg + white + with + br_bg + black + style
}

func (t *Terminal) Render(state *AppState) {
	lines := []string{}
	type Item struct {
		tree  *Tree
		depth int
	}
	stack := []Item{Item{state.Root, 0}}
	for len(stack) > 0 {
		var item Item
		item, stack = stack[len(stack)-1], stack[:len(stack)-1]
		line := strings.Builder{}
		line.WriteString(strings.Repeat("  ", item.depth))
		fgColor := white
		bgColor := black
		effect := no + faint
		if item.tree.IsDir() {
			fgColor = blue
			effect = bold
		}
		if len(lines) == state.CursorLine {
			fgColor, bgColor = bgColor, fgColor
		}
		line.WriteString(escape + effect + with + br_fg + fgColor + with + br_bg + bgColor + style)
		line.WriteString(item.tree.info.Name())
		line.WriteString(DefaultLayout())
		lines = append(lines, line.String())

		if item.tree == state.Root {
			item.tree.MaybeExpand()
			children := append(item.tree.Children[:0:0], item.tree.Children...)
			sort.Slice(children, func(i, j int) bool {
				if children[i].IsDir() != children[j].IsDir() {
					return children[i].IsDir()
				} else {
					return children[i].info.Name() < children[j].info.Name()
				}
			})
			for i := len(children) - 1; i >= 0; i-- {
				stack = append(stack, Item{children[i], item.depth + 1})
			}
		}
	}
	os.Stdout.WriteString(escape + termclear)
	os.Stdout.WriteString(escape + jump)
	os.Stdout.WriteString(strings.Join(lines, "\n"))
}

func (t *Terminal) Close() {
	os.Stdout.WriteString(escape + altbuf + high)
	os.Stdout.WriteString(escape + altbuf + low)
	os.Stdout.WriteString(escape + cursor + high)
	sys.IoctlSetTermios(1, sys.TIOCSETA, &t.originalTermios)
}
