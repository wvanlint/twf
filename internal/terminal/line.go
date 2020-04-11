package terminal

import (
	"regexp"
	"strings"
	"unicode/utf8"
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
	if graphics != nil {
		l.line.WriteString(graphics.ToEscapeCode())
	}
	for len(s) > 0 && l.length < l.maxLength {
		r, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		if r == utf8.RuneError {
			continue
		}
		l.length += 1
		l.line.WriteString(string(r))
	}
	l.line.WriteString(resetGraphics)
	l.line.WriteString(l.defaultGraphics.ToEscapeCode())
	return l
}

func (l *line) AppendRaw(s string) Line {
	matches := escapeRegex.FindAllStringIndex(s, -1)
	prevIndex := 0
	for i := 0; i < len(matches)+1; i++ {
		piece := ""
		if i < len(matches) {
			piece = s[prevIndex:matches[i][0]]
		} else {
			piece = s[prevIndex:]
		}
		for len(piece) > 0 && l.length < l.maxLength {
			r, size := utf8.DecodeRuneInString(piece)
			piece = piece[size:]
			if r == utf8.RuneError || r <= 0x1f || (r >= 0x7f && r <= 0x9f) {
				continue
			}
			l.length += 1
			l.line.WriteString(string(r))
		}
		if i < len(matches) {
			l.line.WriteString(s[matches[i][0]:matches[i][1]])
			prevIndex = matches[i][1]
		}
	}
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
