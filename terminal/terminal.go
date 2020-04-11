package terminal

import (
	"fmt"
	sys "golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

const (
	escape         = "\x1b["
	high           = "h"
	low            = "l"
	altbuf         = "?1049"
	cursor         = "?25"
	termclear      = "2J"
	clearline      = "2K"
	cursorDown     = "B"
	cursorLeft     = "D"
	cursorPosition = "H"
)

type Terminal struct {
	originalTermios sys.Termios
	winSize         sys.Winsize
	renderedPaths   []string
	cursorLine      int
	in              *os.File
	out             *os.File
	loop            bool
}

func OpenTerm() (*Terminal, error) {
	termios, err := sys.IoctlGetTermios(1, sys.TIOCGETA)
	if err != nil {
		return nil, err
	}
	inFd, err := sys.Open("/dev/tty", sys.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	outFd, err := sys.Open("/dev/tty", sys.O_WRONLY, 0)
	term := Terminal{
		originalTermios: *termios,
		in:              os.NewFile(uintptr(inFd), "/dev/tty"),
		out:             os.NewFile(uintptr(outFd), "/dev/tty"),
	}

	return &term, term.initTerm()
}

func (t *Terminal) initTerm() error {
	termios := t.originalTermios
	termios.Iflag &^= uint64(sys.IGNCR) | uint64(sys.INLCR) | uint64(sys.ICRNL)
	termios.Lflag &^= uint64(sys.ECHO) | uint64(sys.ICANON)
	if err := sys.IoctlSetTermios(1, sys.TIOCSETA, &termios); err != nil {
		return err
	}

	t.out.WriteString(escape + altbuf + high)
	t.out.WriteString(escape + cursor + low)
	return nil
}

func (t *Terminal) revertTerm() {
	t.out.WriteString(escape + altbuf + high)
	t.out.WriteString(escape + altbuf + low)
	t.out.WriteString(escape + cursor + high)
	sys.IoctlSetTermios(1, sys.TIOCSETA, &t.originalTermios)
}

func (t *Terminal) Close() {
	t.revertTerm()
	t.in.Close()
	t.out.Close()
}

func moveTo(row int, col int) string {
	return escape + strconv.Itoa(row) + and + strconv.Itoa(col) + cursorPosition
}

func (t *Terminal) drawBorder(p Position) {
	if p.Rows < 2 || p.Cols < 2 {
		return
	}
	t.out.WriteString(moveTo(p.Top, p.Left))
	t.out.WriteString("┌" + strings.Repeat("─", p.Cols-2) + "┐")
	for i := 1; i < p.Rows-1; i++ {
		t.out.WriteString(moveTo(p.Top+i, p.Left+p.Cols-1))
		t.out.WriteString("│")
	}
	for i := 1; i < p.Rows-1; i++ {
		t.out.WriteString(moveTo(p.Top+i, p.Left))
		t.out.WriteString("│")
	}
	t.out.WriteString(moveTo(p.Top+p.Rows-1, p.Left))
	t.out.WriteString("└" + strings.Repeat("─", p.Cols-2) + "┘")
}

func (t *Terminal) render(views []View) {
	for _, view := range views {
		if !view.ShouldRender() {
			continue
		}

		p := view.Position(int(t.winSize.Row), int(t.winSize.Col))
		if view.HasBorder() {
			t.drawBorder(p)
			p = p.Shrink(1)
		}

		lines := view.Render(p)
		for row := 0; row < p.Rows; row++ {
			t.out.WriteString(moveTo(p.Top+row, p.Left))
			if row < len(lines) {
				t.out.WriteString(lines[row].Text())
				t.out.WriteString(strings.Repeat(" ", p.Cols-lines[row].Length()))
			} else {
				t.out.WriteString(strings.Repeat(" ", p.Cols))
			}
		}
	}
}

func (t *Terminal) fetchWinSize() error {
	winSize, err := sys.IoctlGetWinsize(1, sys.TIOCGWINSZ)
	if err != nil {
		return err
	}
	t.winSize = *winSize
	return nil
}

func (t *Terminal) StartLoop(bindings map[string]string, views []View) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Terminal: %v, stacktrace: %s", r, string(debug.Stack()))
		}
	}()

	intSigs := make(chan os.Signal, 1)
	signal.Notify(intSigs, sys.SIGINT, sys.SIGTERM)

	winChSig := make(chan os.Signal, 1)
	signal.Notify(winChSig, sys.SIGWINCH)

	events := make(chan Event, 1)
	go sendEventsLoop(t.in, events)

	err = t.fetchWinSize()
	if err != nil {
		return err
	}
	t.render(views)

	t.loop = true
	for {
		select {
		case <-intSigs:
			t.loop = false
		case <-winChSig:
			zap.L().Debug("Received window change.")
			t.fetchWinSize()
			t.render(views)
			zap.L().Debug("Rerendered.")
		case event := <-events:
			if cmdKey, ok := bindings[event.HashKey()]; ok {
				if cmd, ok := t.getCommands()[cmdKey]; ok {
					cmd(t)
				} else {
					for _, view := range views {
						if cmd, ok := view.GetCommands()[cmdKey]; ok {
							cmd(t)
							break
						}
					}
				}
				t.render(views)
			}
		}
		if !t.loop {
			break
		}
	}
	return err
}

func (t *Terminal) getCommands() map[string]Command {
	return map[string]Command{
		"quit": func(_ TerminalHelper, args ...interface{}) {
			t.loop = false
		},
	}
}

func (t *Terminal) ExecuteInTerminal(cmd string) (string, error) {
	tempF, err := ioutil.TempFile("", "twf_")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempF.Name())
	defer tempF.Close()

	fzf := exec.Command("bash", "-c", cmd+" > "+tempF.Name())
	fzf.Stdin = t.in
	fzf.Stdout = t.out
	fzf.Stderr = t.out
	t.revertTerm()
	defer t.initTerm()
	err = fzf.Run()
	if err != nil {
		return "", err
	}
	out, err := ioutil.ReadAll(tempF)
	return string(out), err
}
