package terminal

import (
	"fmt"
	"strings"
)

const (
	csi = "\x1b["

	enableAltBuf  = csi + "?1049h"
	disableAltBuf = csi + "?1049l"
	showCursor    = csi + "?25h"
	hideCursor    = csi + "?25l"
	enableWrap    = csi + "?7h"
	disableWrap   = csi + "?7l"

	eraseDisplayEnd = csi + "0J"
	eraseDisplayAll = csi + "2J"
	eraseLineAll    = csi + "2K"

	resetGraphics = csi + "m"

	bold      = "1"
	faint     = "2"
	reverse   = "7"
	nobold    = "21"
	nofaint   = "22"
	noreverse = "27"
)

func cursorUp(args ...int) string {
	i := 1
	if len(args) > 0 {
		i = args[0]
	}
	if i == 0 {
		return ""
	}
	return fmt.Sprint(csi, i, "A")
}

func cursorDown(args ...int) string {
	i := 1
	if len(args) > 0 {
		i = args[0]
	}
	if i == 0 {
		return ""
	}
	return fmt.Sprint(csi, i, "B")
}

func cursorForward(args ...int) string {
	i := 1
	if len(args) > 0 {
		i = args[0]
	}
	if i == 0 {
		return ""
	}
	return fmt.Sprint(csi, i, "C")
}

func cursorBack(args ...int) string {
	i := 1
	if len(args) > 0 {
		i = args[0]
	}
	if i == 0 {
		return ""
	}
	return fmt.Sprint(csi, i, "D")
}

func cursorPosition(row int, column int) string {
	return fmt.Sprint(csi, row, ";", column, "H")
}

type Color interface {
	FgCode() string
	BgCode() string
}

type Color3Bit struct {
	Value  int
	Bright bool
}

type Color8Bit struct {
	Value int
}

type Color24Bit struct {
	R, G, B int
}

func (c Color3Bit) FgCode() string {
	if c.Bright {
		return fmt.Sprint("9", c.Value)
	} else {
		return fmt.Sprint("3", c.Value)
	}
}

func (c Color3Bit) BgCode() string {
	if c.Bright {
		return fmt.Sprint("10", c.Value)
	} else {
		return fmt.Sprint("4", c.Value)
	}
}

func (c Color8Bit) FgCode() string {
	return fmt.Sprint("38;5;", c.Value)
}

func (c Color8Bit) BgCode() string {
	return fmt.Sprint("48;5;", c.Value)
}

func (c Color24Bit) FgCode() string {
	return fmt.Sprint("38;2;", c.R, ";", c.G, ";", c.B)
}

func (c Color24Bit) BgCode() string {
	return fmt.Sprint("48;2;", c.R, ";", c.G, ";", c.B)
}

type Graphics struct {
	FgColor Color
	BgColor Color

	Bold    bool
	Reverse bool
}

func (g *Graphics) ToEscapeCode() string {
	codes := []string{}
	if g.Bold {
		codes = append(codes, bold)
	}
	if g.Reverse {
		codes = append(codes, reverse)
	}
	if g.FgColor != nil {
		codes = append(codes, g.FgColor.FgCode())
	}
	if g.BgColor != nil {
		codes = append(codes, g.BgColor.BgCode())
	}

	return csi + strings.Join(codes, ";") + "m"
}

func (g *Graphics) Merge(other *Graphics) {
	if g.FgColor == nil {
		g.FgColor = other.FgColor
	}
	if g.BgColor == nil {
		g.BgColor = other.BgColor
	}
	g.Bold = g.Bold || other.Bold
	g.Reverse = g.Reverse || other.Reverse
}
