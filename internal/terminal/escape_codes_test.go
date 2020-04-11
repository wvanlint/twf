package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphicsToEscapeCode(t *testing.T) {
	assert.Contains(
		t,
		[]string{
			"\x1b[1;32m",
			"\x1b[32;1m",
		},
		(&Graphics{
			Bold: true,
			FgColor: Color3Bit{
				Value: 2,
			},
		}).ToEscapeCode(),
	)
	assert.Contains(
		t,
		[]string{
			"\x1b[7;104m",
			"\x1b[104;7m",
		},
		(&Graphics{
			Reverse: true,
			BgColor: Color3Bit{
				Value:  4,
				Bright: true,
			},
		}).ToEscapeCode(),
	)

	assert.Contains(
		t,
		[]string{
			"\x1b[38;5;240m",
		},
		(&Graphics{
			FgColor: Color8Bit{
				Value: 240,
			},
		}).ToEscapeCode(),
	)
	assert.Contains(
		t,
		[]string{
			"\x1b[48;5;25m",
		},
		(&Graphics{
			BgColor: Color8Bit{
				Value: 25,
			},
		}).ToEscapeCode(),
	)

	assert.Contains(
		t,
		[]string{
			"\x1b[38;2;5;6;7m",
		},
		(&Graphics{
			FgColor: Color24Bit{
				R: 5,
				G: 6,
				B: 7,
			},
		}).ToEscapeCode(),
	)
}
