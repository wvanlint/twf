package terminal

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	bold      = "1"
	faint     = "2"
	reverse   = "7"
	nobold    = "21"
	nofaint   = "22"
	noreverse = "27"
	and       = ";"
	style     = "m"
)

var escapeRegex *regexp.Regexp

func init() {
	escapeRegex = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")
}

type Color interface {
	ToAnsiFg() string
	ToAnsiBg() string
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

func (c Color3Bit) ToAnsiFg() string {
	if c.Bright {
		return fmt.Sprintf("9%d", c.Value)
	} else {
		return fmt.Sprintf("3%d", c.Value)
	}
}

func (c Color3Bit) ToAnsiBg() string {
	if c.Bright {
		return fmt.Sprintf("10%d", c.Value)
	} else {
		return fmt.Sprintf("4%d", c.Value)
	}
}

type Graphics struct {
	FgColor Color
	BgColor Color

	Bold    bool
	Reverse bool
}

func (g *Graphics) ToAnsi(reset bool) string {
	changes := []string{}
	if g.Bold {
		changes = append(changes, bold)
	}
	if g.Reverse {
		changes = append(changes, reverse)
	}
	if g.FgColor != nil {
		changes = append(changes, g.FgColor.ToAnsiFg())
	}
	if g.BgColor != nil {
		changes = append(changes, g.BgColor.ToAnsiBg())
	}

	result := escape + strings.Join(changes, and) + style
	if reset {
		return escape + style + result
	} else {
		return result
	}
}

type Line interface {
	Append(string, *Graphics) Line
	AppendRaw(string) Line
	Length() int
	Text() string
}

type line struct {
	line            strings.Builder
	length          int
	maxLength       int
	defaultGraphics *Graphics
}

func NewLine(defaultGraphics *Graphics, maxLength int) Line {
	return &line{defaultGraphics: defaultGraphics, maxLength: maxLength}
}

func (l *line) Append(s string, graphics *Graphics) Line {
	if l.length >= l.maxLength {
		return l
	}
	if l.length+len(s) > l.maxLength {
		s = s[:l.maxLength-l.length]
	}
	l.length += len(s)
	if graphics != nil {
		l.line.WriteString(graphics.ToAnsi(false))
	}
	l.line.WriteString(s)
	l.line.WriteString(l.defaultGraphics.ToAnsi(true))
	return l
}

func (l *line) AppendRaw(s string) Line {
	matches := escapeRegex.FindAllStringIndex(s, -1)
	prevIndex := 0
	for i := 0; i < len(matches); i++ {
		piece := s[prevIndex:matches[i][0]]
		if l.length+len(piece) > l.maxLength {
			piece = piece[:l.maxLength-l.length]
		}
		l.length += len(piece)
		l.line.WriteString(piece)
		l.line.WriteString(s[matches[i][0]:matches[i][1]])
		prevIndex = matches[i][1]
	}
	piece := s[prevIndex:]
	if l.length+len(piece) > l.maxLength {
		piece = piece[:l.maxLength-l.length]
	}
	l.length += len(piece)
	l.line.WriteString(piece)
	l.line.WriteString(l.defaultGraphics.ToAnsi(true))
	return l
}

func (l *line) Length() int {
	return l.length
}

func (l *line) Text() string {
	return l.line.String()
}
