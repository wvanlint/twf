package terminal

import (
	"fmt"
	"io"
	"unicode/utf8"

	"go.uber.org/zap"
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
	CtrlI
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
	Escape
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

	ShiftUp
	ShiftDown
	ShiftLeft
	ShiftRight

	CtrlUp
	CtrlDown
	CtrlLeft
	CtrlRight

	Rune
)

const (
	Tab   = CtrlI
	Enter = CtrlM
	Del   = 127
)

type EventSymbol int

type Event struct {
	Symbol EventSymbol
	Value  rune
}

func (e *Event) HashKey() string {
	if e.Symbol == Rune {
		return string(e.Value)
	} else {
		return fmt.Sprint("#", e.Symbol)
	}
}

func hasPrefix(s []byte, prefix []byte) bool {
	if (s == nil) != (prefix == nil) {
		return false
	}

	if len(s) < len(prefix) {
		return false
	}

	for i := range prefix {
		if s[i] != prefix[i] {
			return false
		}
	}

	return true
}

func readEvents(r io.Reader, out chan Event, next chan bool) {
	for {
		in := make([]byte, 128)
		n, err := r.Read(in)
		if err != nil {
			return
		}
		in = in[:n]
		zap.L().Sugar().Debug("Input bytes: ", in)

		for len(in) > 0 {
			if in[0] == Escape {
				// Literal escape or escape sequences.
				switch {
				case in[0] == Escape && (len(in) == 1 || in[1] == Escape):
					out <- Event{Symbol: Escape}
					in = in[1:]
				case hasPrefix(in, []byte{27, 91, 65}):
					out <- Event{Symbol: Up}
					in = in[3:]
				case hasPrefix(in, []byte{27, 91, 66}):
					out <- Event{Symbol: Down}
					in = in[3:]
				case hasPrefix(in, []byte{27, 91, 67}):
					out <- Event{Symbol: Right}
					in = in[3:]
				case hasPrefix(in, []byte{27, 91, 68}):
					out <- Event{Symbol: Left}
					in = in[3:]
				case hasPrefix(in, []byte{27, 91, 49, 126}):
					out <- Event{Symbol: Home}
					in = in[4:]
				case hasPrefix(in, []byte{27, 91, 52, 126}):
					out <- Event{Symbol: End}
					in = in[4:]
				case hasPrefix(in, []byte{27, 91, 49, 59, 50, 65}):
					out <- Event{Symbol: ShiftUp}
					in = in[6:]
				case hasPrefix(in, []byte{27, 91, 49, 59, 50, 66}):
					out <- Event{Symbol: ShiftDown}
					in = in[6:]
				case hasPrefix(in, []byte{27, 91, 49, 59, 50, 67}):
					out <- Event{Symbol: ShiftRight}
					in = in[6:]
				case hasPrefix(in, []byte{27, 91, 49, 59, 50, 68}):
					out <- Event{Symbol: ShiftLeft}
					in = in[6:]
				case hasPrefix(in, []byte{27, 91, 49, 59, 53, 65}):
					out <- Event{Symbol: CtrlUp}
					in = in[6:]
				case hasPrefix(in, []byte{27, 91, 49, 59, 53, 66}):
					out <- Event{Symbol: CtrlDown}
					in = in[6:]
				case hasPrefix(in, []byte{27, 91, 49, 59, 53, 67}):
					out <- Event{Symbol: CtrlRight}
					in = in[6:]
				case hasPrefix(in, []byte{27, 91, 49, 59, 53, 68}):
					out <- Event{Symbol: CtrlLeft}
					in = in[6:]
				default:
					// Unhandled entries.
					in = in[0:0]
				}
			} else {
				r, rSize := utf8.DecodeRune(in)
				switch {
				case in[0] <= 31:
					out <- Event{Symbol: EventSymbol(in[0])}
					in = in[1:]
				case in[0] == Del:
					out <- Event{Symbol: Del}
					in = in[1:]
				case r != utf8.RuneError:
					out <- Event{Symbol: Rune, Value: r}
					in = in[rSize:]
				default:
					// Unhandled entries.
					in = in[0:0]
				}
			}
			in = in[0:0] // Don't queue commands.
		}
		<-next
	}
}
