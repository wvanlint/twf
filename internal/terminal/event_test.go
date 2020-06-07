package terminal

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testHelperReadEvent(
	t *testing.T,
	input string,
	expectedEvents []Event,
) {
	buffer := new(bytes.Buffer)
	out := make(chan Event)
	next := make(chan bool)

	buffer.WriteString(input)
	go readEvents(buffer, out, next)

	for _, expected := range expectedEvents {
		select {
		case ev := <-out:
			assert.Equal(t, expected, ev)
		case <-time.After(1 * time.Second):
			assert.Fail(t, "Timeout")
		}
	}
}

func TestReadTabEvent(t *testing.T) {
	testHelperReadEvent(
		t,
		"\x09",
		[]Event{
			Event{Symbol: Tab},
		},
	)
}

func TestReadAlphanumericEvent(t *testing.T) {
	testHelperReadEvent(
		t,
		"a",
		[]Event{
			Event{Symbol: Rune, Value: 'a'},
		},
	)
}

func TestReadUnicodeEvent(t *testing.T) {
	testHelperReadEvent(
		t,
		"😊",
		[]Event{
			Event{Symbol: Rune, Value: '😊'},
		},
	)
}

func TestReadEscapeCodeEvent(t *testing.T) {
	testHelperReadEvent(
		t,
		"\x1b[1~",
		[]Event{
			Event{Symbol: Home},
		},
	)
}

func TestReadUnknownEscapeCode(t *testing.T) {
	testHelperReadEvent(
		t,
		"\x1b[malformed",
		[]Event{},
	)
}
