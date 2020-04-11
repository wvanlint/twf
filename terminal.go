package main

import (
	sys "golang.org/x/sys/unix"
	"os"
	"os/signal"
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

type Callbacks struct {
	ChangeCursor    func(string)
	ToggleExpansion func(string)
	Quit            func()
}

type Terminal struct {
	originalTermios sys.Termios
	renderedPaths   []string
	cursorLine      int
	callbacks       Callbacks
}

func InitTerm(callbacks Callbacks) (*Terminal, error) {
	termios, err := sys.IoctlGetTermios(1, sys.TIOCGETA)
	if err != nil {
		return nil, err
	}
	term := Terminal{originalTermios: *termios, callbacks: callbacks}

	termios.Lflag = termios.Lflag & (^uint64(sys.ECHO) & ^uint64(sys.ICANON))
	sys.IoctlSetTermios(1, sys.TIOCSETA, termios)
	os.Stdout.WriteString(escape + altbuf + high)
	os.Stdout.WriteString(escape + cursor + low)

	return &term, nil
}

func DefaultLayout() string {
	return escape + plain + style + escape + br_fg + white + with + br_bg + black + style
}

func (t *Terminal) renderNode(node *Tree, indentation int, selected bool) string {
	line := strings.Builder{}
	line.WriteString(strings.Repeat("  ", indentation))
	fgColor := white
	bgColor := black
	effect := no + faint
	if node.IsDir() {
		fgColor = blue
		effect = bold
	}
	if selected {
		fgColor, bgColor = bgColor, fgColor
	}
	line.WriteString(escape + effect + with + br_fg + fgColor + with + br_bg + bgColor + style)
	line.WriteString(node.info.Name())
	line.WriteString(DefaultLayout())
	return line.String()
}

func (t *Terminal) Render(state *AppState) {
	t.renderedPaths = []string{}
	lines := []string{}
	type Item struct {
		tree  *Tree
		depth int
	}
	stack := []Item{Item{state.Root, 0}}
	for len(stack) > 0 {
		var item Item
		item, stack = stack[len(stack)-1], stack[:len(stack)-1]
		line := t.renderNode(item.tree, item.depth, item.tree.Path == state.Cursor)
		lines = append(lines, line)

		t.renderedPaths = append(t.renderedPaths, item.tree.Path)
		if item.tree.Path == state.Cursor {
			t.cursorLine = len(lines) - 1
		}

		if value, _ := state.Expansions[item.tree.Path]; value {
			children, _ := item.tree.Children()
			sort.Slice(children, ByTypeAndName(children))
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

func (t *Terminal) StartLoop(state *AppState, stop chan bool) {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	go func() {
		stopLoop := false
		t.Render(state)
		for {
			select {
			case <-sigs:
				stopLoop = true
			case <-stop:
				stopLoop = true
			default:
				t.ReadCommand(state)
				t.Render(state)
			}
			if stopLoop {
				break
			}
		}
		done <- true
	}()
	signal.Notify(sigs, sys.SIGINT, sys.SIGTERM)
	<-done
}

func (t *Terminal) ReadCommand(state *AppState) {
	input := make([]byte, 1)
	_, err := os.Stdout.Read(input)
	if err == nil {
		switch input[0] {
		case 'j':
			nextCursorLine := t.cursorLine + 1
			if nextCursorLine <= len(t.renderedPaths)-1 {
				t.callbacks.ChangeCursor(t.renderedPaths[nextCursorLine])
			}
		case 'k':
			prevCursorLine := t.cursorLine - 1
			if prevCursorLine >= 0 {
				t.callbacks.ChangeCursor(t.renderedPaths[prevCursorLine])
			}
		case 'q':
			t.callbacks.Quit()
		case 10:
			t.callbacks.ToggleExpansion(state.Cursor)
		}
	}
}
