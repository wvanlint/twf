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
		"ctrl-space":        {Symbol: term.CtrlSpace},
		"tab":               {Symbol: term.Tab},
		"enter":             {Symbol: term.Enter},
		"esc":               {Symbol: term.Escape},
		"ctrl-backslash":    {Symbol: term.CtrlBackslash},
		"ctrl-rightbracket": {Symbol: term.CtrlRightBracket},
		"ctrl-caret":        {Symbol: term.CtrlCaret},
		"ctrl-slash":        {Symbol: term.CtrlSlash},
		"up":                {Symbol: term.Up},
		"down":              {Symbol: term.Down},
		"left":              {Symbol: term.Left},
		"right":             {Symbol: term.Right},
		"home":              {Symbol: term.Home},
		"end":               {Symbol: term.End},
		"pgup":              {Symbol: term.PgUp},
		"pgdown":            {Symbol: term.PgDown},
		"shift-up":          {Symbol: term.ShiftUp},
		"shift-down":        {Symbol: term.ShiftDown},
		"shift-left":        {Symbol: term.ShiftLeft},
		"shift-right":       {Symbol: term.ShiftRight},
		"ctrl-up":           {Symbol: term.CtrlUp},
		"ctrl-down":         {Symbol: term.CtrlDown},
		"ctrl-left":         {Symbol: term.CtrlLeft},
		"ctrl-right":        {Symbol: term.CtrlRight},
		"del":               {Symbol: term.Del},
	}
	for i := 0; i < 26; i++ {
		strToEventM["ctrl-"+string(rune('a'+i))] = &term.Event{Symbol: term.EventSymbol(term.CtrlA + i)}
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
