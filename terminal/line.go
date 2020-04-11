package terminal

import (
	"regexp"
	"strings"
)

var escapeRegex *regexp.Regexp

func init() {
	escapeRegex = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")
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
		l.line.WriteString(graphics.ToEscapeCode())
	}
	l.line.WriteString(s)
	l.line.WriteString(resetGraphics)
	l.line.WriteString(l.defaultGraphics.ToEscapeCode())
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
	l.line.WriteString(resetGraphics)
	l.line.WriteString(l.defaultGraphics.ToEscapeCode())
	return l
}

func (l *line) Length() int {
	return l.length
}

func (l *line) Text() string {
	return l.line.String()
}
