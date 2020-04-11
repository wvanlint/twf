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
	//	"time"

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

type Callbacks struct {
	ChangeCursor func(string)
	Prev         func()
	Next         func()
	Up           func()
	Open         func()
	Close        func()
	OpenAll      func()
	CloseAll     func()
	Toggle       func()
	ToggleAll    func()
	Quit         func()
}

type Terminal struct {
	originalTermios sys.Termios
	winSize         sys.Winsize
	renderedPaths   []string
	cursorLine      int
	callbacks       Callbacks
	in              *os.File
	out             *os.File
}

func InitTerm(callbacks Callbacks) (*Terminal, error) {
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
		callbacks:       callbacks,
		in:              os.NewFile(uintptr(inFd), "/dev/tty"),
		out:             os.NewFile(uintptr(outFd), "/dev/tty"),
	}

	return &term, term.init()
}

func (t *Terminal) init() error {
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

func (t *Terminal) Close() {
	t.out.WriteString(escape + altbuf + high)
	t.out.WriteString(escape + altbuf + low)
	t.out.WriteString(escape + cursor + high)
	sys.IoctlSetTermios(1, sys.TIOCSETA, &t.originalTermios)
	t.in.Close()
	t.out.Close()
}

func (t *Terminal) fetchWinSize() error {
	winSize, err := sys.IoctlGetWinsize(1, sys.TIOCGWINSZ)
	if err != nil {
		return err
	}
	t.winSize = *winSize
	return nil
}

func (t *Terminal) StartLoop(views []View, stop chan bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Terminal: %v, stacktrace: %s", r, string(debug.Stack()))
		}
	}()

	intSigs := make(chan os.Signal, 1)
	signal.Notify(intSigs, sys.SIGINT, sys.SIGTERM)

	winChSig := make(chan os.Signal, 1)
	signal.Notify(winChSig, sys.SIGWINCH)

	eventSig := make(chan []byte, 1)
	go func() {
		for {
			input := make([]byte, 128)
			_, err := t.in.Read(input)
			if err == nil {
				eventSig <- input
			}
		}
	}()

	stopLoop := false
	err = t.fetchWinSize()
	if err != nil {
		return err
	}
	t.render(views)
	for {
		select {
		case <-intSigs:
			stopLoop = true
		case <-stop:
			stopLoop = true
		case <-winChSig:
			zap.L().Debug("Received window change.")
			t.fetchWinSize()
			t.render(views)
			zap.L().Debug("Rerendered.")
		case event := <-eventSig:
			t.ProcessCommand(event)
			t.render(views)
		}
		if stopLoop {
			break
		}
	}
	return err
}

func (t *Terminal) ProcessCommand(input []byte) {
	switch input[0] {
	case 'j':
		t.callbacks.Next()
	case 'k':
		t.callbacks.Prev()
	case 'q':
		t.callbacks.Quit()
	case 27:
		t.callbacks.Quit()
	case 'o':
		t.callbacks.Toggle()
	case 'O':
		t.callbacks.ToggleAll()
	case 'p':
		t.callbacks.Up()
	case 'P':
		t.callbacks.Up()
		t.callbacks.Close()
	case '/':
		tempF, _ := ioutil.TempFile("", "twf_")
		fzf := exec.Command("bash", "-c", "fzf > "+tempF.Name())
		fzf.Stdin = t.in
		fzf.Stdout = t.out
		fzf.Stderr = t.out
		t.Close()
		fzf.Run()
		t.init()
		content, _ := ioutil.ReadAll(tempF)
		tempF.Close()
		os.Remove(tempF.Name())
		t.callbacks.ChangeCursor(strings.TrimSpace(string(content)))
	}
}
