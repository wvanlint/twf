package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	term "github.com/wvanlint/twf/internal/terminal"
)

func TestParseGraphics(t *testing.T) {
	g, err := parseGraphics("bold:fg#blue")
	assert.Nil(t, err)
	g2, err := parseGraphics(graphicsToString(g))
	assert.Nil(t, err)
	assert.Equal(
		t,
		&term.Graphics{
			Bold:    true,
			FgColor: term.Color3Bit{Value: 4, Bright: false},
		},
		g,
	)
	assert.Equal(t, g, g2)

	g, err = parseGraphics("bg#56:reverse")
	assert.Nil(t, err)
	g2, err = parseGraphics(graphicsToString(g))
	assert.Nil(t, err)
	assert.Equal(
		t,
		&term.Graphics{
			Reverse: true,
			BgColor: term.Color8Bit{Value: 56},
		},
		g,
	)
	assert.Equal(t, g, g2)

	g, err = parseGraphics("fg#6a6a6a")
	assert.Nil(t, err)
	g2, err = parseGraphics(graphicsToString(g))
	assert.Nil(t, err)
	assert.Equal(
		t,
		&term.Graphics{
			FgColor: term.Color24Bit{R: 106, G: 106, B: 106},
		},
		g,
	)
	assert.Equal(t, g, g2)
}
