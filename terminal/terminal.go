package terminal

import (
	"fmt"
	sys "golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
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
}

func InitTerm(callbacks Callbacks) (*Terminal, error) {
	termios, err := sys.IoctlGetTermios(1, sys.TIOCGETA)
	if err != nil {
		return nil, err
	}
	term := Terminal{originalTermios: *termios, callbacks: callbacks}

	return &term, term.init()
}

func (t *Terminal) init() error {
	termios := t.originalTermios
	termios.Iflag &^= uint64(sys.IGNCR) | uint64(sys.INLCR) | uint64(sys.ICRNL)
	termios.Lflag &^= uint64(sys.ECHO) | uint64(sys.ICANON)
	if err := sys.IoctlSetTermios(1, sys.TIOCSETA, &termios); err != nil {
		return err
	}

	os.Stdout.WriteString(escape + altbuf + high)
	os.Stdout.WriteString(escape + cursor + low)
	return nil
}

func (t *Terminal) render(view View) {
	lines, _ := view.Render()
	strLines := []string{}
	for _, line := range lines {
		strLines = append(strLines, line.Text())
	}

	os.Stdout.WriteString(escape + termclear)
	os.Stdout.WriteString(escape + jump)
	os.Stdout.WriteString(strings.Join(strLines, "\n"))
}

func (t *Terminal) Close() {
	os.Stdout.WriteString(escape + altbuf + high)
	os.Stdout.WriteString(escape + altbuf + low)
	os.Stdout.WriteString(escape + cursor + high)
	sys.IoctlSetTermios(1, sys.TIOCSETA, &t.originalTermios)
}

func (t *Terminal) fetchWinSize() error {
	winSize, err := sys.IoctlGetWinsize(1, sys.TIOCGWINSZ)
	if err != nil {
		return err
	}
	t.winSize = *winSize
	return nil
}

func (t *Terminal) StartLoop(view View, stop chan bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Terminal: %v, stacktrace: %s", r, string(debug.Stack()))
		}
	}()

	intSigs := make(chan os.Signal, 1)
	signal.Notify(intSigs, sys.SIGINT, sys.SIGTERM)

	winChSig := make(chan os.Signal, 1)
	signal.Notify(winChSig, sys.SIGWINCH)

	stopLoop := false
	//err = t.fetchWinSize()
	//if err != nil {
	//	return err
	//}
	t.render(view)
	for {
		select {
		case <-intSigs:
			stopLoop = true
		case <-stop:
			stopLoop = true
		case <-winChSig:
			//t.fetchWinSize()
			t.render(view)
		default:
			t.ReadCommand()
			t.render(view)
		}
		if stopLoop {
			break
		}
	}
	return err
}

func (t *Terminal) GetPreview(cmdTemplate string, path string) (string, error) {
	cmd := strings.ReplaceAll(cmdTemplate, "{}", path)
	var output strings.Builder
	preview := exec.Command("bash", "-c", cmd)
	preview.Stdout = &output
	err := preview.Run()
	return output.String(), err
}

func (t *Terminal) ReadCommand() {
	input := make([]byte, 128)
	_, err := sys.Read(1, input)
	if err == nil {
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
			fzf.Stdin = os.Stdin
			fzf.Stdout = os.Stdout
			fzf.Stderr = os.Stderr
			t.Close()
			fzf.Run()
			t.init()
			content, _ := ioutil.ReadAll(tempF)
			tempF.Close()
			os.Remove(tempF.Name())
			t.callbacks.ChangeCursor(strings.TrimSpace(string(content)))
		}
	}
}
