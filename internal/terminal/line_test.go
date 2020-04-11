package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineAppend(t *testing.T) {
	line := NewLine(
		&Graphics{Reverse: true},
		2,
	)
	line.Append("b", &Graphics{Bold: true})
	assert.Equal(t, 1, line.Length())
	line.Append("la", &Graphics{})
	assert.Equal(t, 2, line.Length())
	assert.Equal(t, "\x1b[1mb\x1b[m\x1b[7m\x1b[ml\x1b[m\x1b[7m", line.Text())
}

func TestLineAppendRaw(t *testing.T) {
	line := NewLine(
		&Graphics{Reverse: true},
		2,
	)

	line.AppendRaw("\x1b[44mabc")
	assert.Equal(t, 2, line.Length())
	assert.Equal(t, "\x1b[44mab\x1b[m\x1b[7m", line.Text())
}

func TestLineUnicode(t *testing.T) {
	line := NewLine(
		&Graphics{Reverse: true},
		2,
	)

	line.Append("😊", nil)
	assert.Equal(t, 1, line.Length())
}
