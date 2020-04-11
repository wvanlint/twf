package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	term "github.com/wvanlint/twf/internal/terminal"
)

func TestEventSerializationSpecialKey(t *testing.T) {
	ev, err := parseEvent("enter")
	assert.Nil(t, err)
	ev2, err := parseEvent(eventHashKeyToString(ev.HashKey()))
	assert.Nil(t, err)
	assert.Equal(
		t,
		&term.Event{Symbol: term.Enter},
		ev,
	)
	assert.Equal(t, ev, ev2)
}

func TestEventSerializationAscii(t *testing.T) {
	ev, err := parseEvent("b")
	assert.Nil(t, err)
	ev2, err := parseEvent(eventHashKeyToString(ev.HashKey()))
	assert.Nil(t, err)
	assert.Equal(
		t,
		&term.Event{Symbol: term.Rune, Value: 'b'},
		ev,
	)
	assert.Equal(t, ev, ev2)
}

func TestEventSerializationUnicode(t *testing.T) {
	ev, err := parseEvent("😊")
	assert.Nil(t, err)
	ev2, err := parseEvent(eventHashKeyToString(ev.HashKey()))
	assert.Nil(t, err)
	assert.Equal(
		t,
		&term.Event{Symbol: term.Rune, Value: '😊'},
		ev,
	)
	assert.Equal(t, ev, ev2)
}

func TestEventSerializationCtrlKey(t *testing.T) {
	ev, err := parseEvent("ctrl-l")
	assert.Nil(t, err)
	ev2, err := parseEvent(eventHashKeyToString(ev.HashKey()))
	assert.Nil(t, err)
	assert.Equal(
		t,
		&term.Event{Symbol: term.CtrlL},
		ev,
	)
	assert.Equal(t, ev, ev2)
}
