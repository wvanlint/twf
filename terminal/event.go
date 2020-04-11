package terminal

import (
	"fmt"
	"io"
	"unicode/utf8"
)

const (
	// Ascii control characters.
	CtrlSpace = iota
	CtrlA
	CtrlB
	CtrlC
	CtrlD
	CtrlE
	CtrlF
	CtrlG
	CtrlH
	CtrlJ
	CtrlK
	CtrlL
	CtrlM
	CtrlN
	CtrlO
	CtrlP
	CtrlQ
	CtrlR
	CtrlS
	CtrlT
	CtrlU
	CtrlV
	CtrlW
	CtrlX
	CtrlY
	CtrlZ
	ESC
	CtrlBackslash
	CtrlRightBracket
	CtrlCaret
	CtrlSlash

	Up
	Down
	Left
	Right
	Home
	End

	SUp
	SDown
	SLeft
	SRight

	Rune
)

const (
	Tab = CtrlL
	Del = 127
)

type EventSymbol int

type Event struct {
	Symbol EventSymbol
	Value  rune
}

func (e *Event) HashKey() string {
	return fmt.Sprint(e.Symbol, "#", e.Value)
}

func cmp(a []byte, b []byte) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func sendEventsLoop(r io.Reader, out chan Event) {
	for {
		in := make([]byte, 128)
		n, err := r.Read(in)
		if err != nil {
			continue
		}
		in = in[:n]
		switch {
		case len(in) == 1 && in[0] <= 31:
			out <- Event{Symbol: EventSymbol(in[0])}
		case len(in) == 1 && in[0] == Del:
			out <- Event{Symbol: Del}
		case cmp(in, []byte{27, 91, 65}):
			out <- Event{Symbol: Up}
		case cmp(in, []byte{27, 91, 66}):
			out <- Event{Symbol: Down}
		case cmp(in, []byte{27, 91, 67}):
			out <- Event{Symbol: Right}
		case cmp(in, []byte{27, 91, 68}):
			out <- Event{Symbol: Left}
		case cmp(in, []byte{27, 91, 49, 126}):
			out <- Event{Symbol: Home}
		case cmp(in, []byte{27, 91, 52, 126}):
			out <- Event{Symbol: End}
		default:
			r, size := utf8.DecodeRune(in)
			if r != utf8.RuneError && size == len(in) {
				out <- Event{Symbol: Rune, Value: r}
			}
		}
	}
}
