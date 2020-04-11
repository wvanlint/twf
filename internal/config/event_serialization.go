package config

import (
	"fmt"
	"unicode/utf8"

	term "github.com/wvanlint/twf/internal/terminal"
)

var strToEventM map[string]*term.Event
var eventHashKeyToStrM map[string]string

func init() {
	strToEventM = map[string]*term.Event{
		"ctrl-space":        &term.Event{Symbol: term.CtrlSpace},
		"tab":               &term.Event{Symbol: term.Tab},
		"enter":             &term.Event{Symbol: term.Enter},
		"esc":               &term.Event{Symbol: term.Escape},
		"ctrl-backslash":    &term.Event{Symbol: term.CtrlBackslash},
		"ctrl-rightbracket": &term.Event{Symbol: term.CtrlRightBracket},
		"ctrl-caret":        &term.Event{Symbol: term.CtrlCaret},
		"ctrl-slash":        &term.Event{Symbol: term.CtrlSlash},
		"up":                &term.Event{Symbol: term.Up},
		"down":              &term.Event{Symbol: term.Down},
		"left":              &term.Event{Symbol: term.Left},
		"right":             &term.Event{Symbol: term.Right},
		"home":              &term.Event{Symbol: term.Home},
		"end":               &term.Event{Symbol: term.End},
		"shift-up":          &term.Event{Symbol: term.SUp},
		"shift-down":        &term.Event{Symbol: term.SDown},
		"shift-left":        &term.Event{Symbol: term.SLeft},
		"shift-right":       &term.Event{Symbol: term.SRight},
		"del":               &term.Event{Symbol: term.Del},
	}
	for i := 0; i < 26; i++ {
		strToEventM["ctrl-"+string('a'+i)] = &term.Event{Symbol: term.EventSymbol(term.CtrlA + i)}
	}
	eventHashKeyToStrM = make(map[string]string)
	for str, event := range strToEventM {
		eventHashKeyToStrM[event.HashKey()] = str
	}
}

func eventHashKeyToString(key string) string {
	s, ok := eventHashKeyToStrM[key]
	if ok {
		return s
	}
	r, size := utf8.DecodeRune([]byte(key))
	if r != utf8.RuneError && size == len(key) {
		return key
	}
	return ""
}

func parseEvent(s string) (*term.Event, error) {
	event, ok := strToEventM[s]
	if ok {
		return event, nil
	}
	r, size := utf8.DecodeRune([]byte(s))
	if r != utf8.RuneError && size == len(s) {
		return &term.Event{Symbol: term.Rune, Value: r}, nil
	}
	return &term.Event{}, fmt.Errorf("Can't parse event: %s", s)
}
