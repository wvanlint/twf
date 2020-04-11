package terminal

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"strings"
	//	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"
	sys "golang.org/x/sys/unix"
)

type Terminal struct {
	config         *TerminalConfig
	originalState  terminal.State
	previousRender map[string]bool
	rows           int
	cols           int
	in             *os.File
	out            *os.File
	loop           bool
	currentRow     int
}

type TerminalConfig struct {
	Height float64
}

func OpenTerm(config *TerminalConfig) (*Terminal, error) {
	inFd, err := sys.Open("/dev/tty", sys.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	outFd, err := sys.Open("/dev/tty", sys.O_WRONLY, 0)
	term := Terminal{
		config:     config,
		in:         os.NewFile(uintptr(inFd), "/dev/tty"),
		out:        os.NewFile(uintptr(outFd), "/dev/tty"),
		currentRow: 1,
	}

	return &term, term.initTerm()
}

func (t *Terminal) initTerm() error {
	state, err := terminal.MakeRaw(int(t.out.Fd()))
	if err != nil {
		return err
	}
	t.originalState = *state
	if t.config.Height == 1.0 {
		t.out.WriteString(enableAltBuf)
		t.out.WriteString(cursorPosition(1, 1))
	}
	t.out.WriteString(disableWrap)
	t.out.WriteString(hideCursor)
	return nil
}

func (t *Terminal) revertTerm() {
	if t.config.Height == 1.0 {
		t.out.WriteString(enableAltBuf)
		t.out.WriteString(disableAltBuf)
	}
	t.previousRender = map[string]bool{}
	t.out.WriteString(enableWrap)
	t.out.WriteString(showCursor)
	terminal.Restore(int(t.out.Fd()), &t.originalState)
}

func (t *Terminal) Close() {
	t.out.WriteString(eraseDisplayEnd)
	t.revertTerm()
	t.in.Close()
	t.out.Close()
}

func (t *Terminal) position(row int, col int) string {
	out := &strings.Builder{}
	out.WriteString(cursorBack(t.cols))
	vertical := row - t.currentRow
	if vertical > 0 {
		out.WriteString(cursorDown(vertical))
	} else if vertical < 0 {
		out.WriteString(cursorUp(-vertical))
	}
	out.WriteString(cursorForward(col - 1))
	t.currentRow = row
	return out.String()
}

func (t *Terminal) border(p Position) string {
	if p.Rows < 2 || p.Cols < 2 {
		return ""
	}
	out := &strings.Builder{}
	out.WriteString(t.position(p.Top, p.Left))
	out.WriteString("┌" + strings.Repeat("─", p.Cols-2) + "┐")
	for i := 1; i < p.Rows-1; i++ {
		out.WriteString(t.position(p.Top+i, p.Left+p.Cols-1))
		out.WriteString("│")
	}
	for i := 1; i < p.Rows-1; i++ {
		out.WriteString(t.position(p.Top+i, p.Left))
		out.WriteString("│")
	}
	out.WriteString(t.position(p.Top+p.Rows-1, p.Left))
	out.WriteString("└" + strings.Repeat("─", p.Cols-2) + "┘")
	out.WriteString(t.position(1, 1))
	return out.String()
}

func (t *Terminal) render(views []View) {
	out := &strings.Builder{}
	newRender := map[string]bool{}
	for _, view := range views {
		if !view.ShouldRender() {
			continue
		}

		p := view.Position(t.rows, t.cols)
		if view.HasBorder() {
			s := t.border(p)
			if _, ok := t.previousRender[s]; !ok {
				out.WriteString(s)
			}
			newRender[s] = true
			p = p.Shrink(1)
		}

		lines := view.Render(p)
		for row := 0; row < p.Rows; row++ {
			lineRender := &strings.Builder{}
			lineRender.WriteString(t.position(p.Top+row, p.Left))
			if row < len(lines) {
				lineRender.WriteString(lines[row].Text())
				if p.Cols > lines[row].Length() {
					lineRender.WriteString(strings.Repeat(" ", p.Cols-lines[row].Length()))
				}
			} else {
				lineRender.WriteString(strings.Repeat(" ", p.Cols))
			}
			if p.Top+row < t.rows {
				lineRender.WriteString("\n")
				t.currentRow += 1
			}
			lineRender.WriteString(t.position(1, 1))
			if _, ok := t.previousRender[lineRender.String()]; !ok {
				out.WriteString(lineRender.String())
			}
			newRender[lineRender.String()] = true
		}
	}
	t.out.WriteString(out.String())
	t.previousRender = newRender
}

func (t *Terminal) fetchWinSize() error {
	width, height, err := terminal.GetSize(int(t.out.Fd()))
	if err != nil {
		return err
	}
	t.rows = int(float64(height) * t.config.Height)
	t.cols = width
	return nil
}

func (t *Terminal) StartLoop(bindings map[string][]string, views []View) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Terminal: %v, stacktrace: %s", r, string(debug.Stack()))
		}
	}()

	intSigs := make(chan os.Signal, 1)
	signal.Notify(intSigs, sys.SIGINT, sys.SIGTERM)

	winChSig := make(chan os.Signal, 1)
	signal.Notify(winChSig, sys.SIGWINCH)

	events := make(chan Event)
	nextEvents := make(chan bool)
	go readEvents(t.in, events, nextEvents)

	err = t.fetchWinSize()
	if err != nil {
		return err
	}
	t.render(views)

	t.loop = true
	for {
		select {
		case <-intSigs:
			zap.L().Debug("Received interrupt.")
			t.loop = false
		case <-winChSig:
			zap.L().Debug("Received window change.")
			t.fetchWinSize()
			t.render(views)
			zap.L().Debug("Rerendered.")
		case event := <-events:
			zap.L().Sugar().Debug("Event: ", event)
			cmdKeys, ok := bindings[event.HashKey()]
			zap.L().Sugar().Debug("Cmds: ", cmdKeys)
			if !ok {
				nextEvents <- true
				continue
			}
			for _, cmdKey := range cmdKeys {
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
			}
			nextEvents <- true
			t.render(views)
			//time.Sleep(10 * time.Millisecond)
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
