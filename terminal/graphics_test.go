package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineAppendRaw(t *testing.T) {
	line := NewLine(
		&Graphics{Reverse: true},
		2,
	)

	line.AppendRaw("\x1b[44mabc")
	assert.Equal(t, 2, line.Length())
	assert.Equal(t, "\x1b[44mab\x1b[m\x1b[7m", line.Text())
}
