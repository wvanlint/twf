package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	term "github.com/wvanlint/twf/internal/terminal"
)

var fromStrTo3Bit map[string]int
var from3BitToStr map[int]string

func init() {
	fromStrTo3Bit = map[string]int{
		"black":   0,
		"red":     1,
		"green":   2,
		"yellow":  3,
		"blue":    4,
		"magenta": 5,
		"cyan":    6,
		"white":   7,
	}
	from3BitToStr = make(map[int]string)
	for s, i := range fromStrTo3Bit {
		from3BitToStr[i] = s
	}
}

func parseColor3Bit(s string) (term.Color3Bit, error) {
	c := term.Color3Bit{}
	if strings.HasPrefix(s, "bright") {
		s = s[len("bright"):]
		c.Bright = true
	}
	if v, ok := fromStrTo3Bit[s]; ok {
		c.Value = v
		return c, nil
	} else {
		return c, errors.New("Could not interpret color string.")
	}
}

func color3BitToString(c term.Color3Bit) string {
	colorStr, _ := from3BitToStr[c.Value]
	if c.Bright {
		return "bright" + colorStr
	} else {
		return colorStr
	}
}

func parseColor8Bit(s string) (term.Color8Bit, error) {
	color := term.Color8Bit{}
	i, err := strconv.Atoi(s)
	if err != nil {
		return color, err
	}
	if i >= 0 && i < 256 {
		color.Value = i
		return color, nil
	} else {
		return color, errors.New("Color integer not in correct range.")
	}
}

func color8BitToString(c term.Color8Bit) string {
	return strconv.Itoa(c.Value)
}

func parseColor24Bit(s string) (term.Color24Bit, error) {
	color := term.Color24Bit{}
	rStr, gStr, bStr := s[0:2], s[2:4], s[4:6]
	r, err := strconv.ParseInt(rStr, 16, 8)
	if err != nil {
		return color, err
	}
	g, err := strconv.ParseInt(gStr, 16, 8)
	if err != nil {
		return color, err
	}
	b, err := strconv.ParseInt(bStr, 16, 8)
	if err != nil {
		return color, err
	}
	color.R, color.G, color.B = int(r), int(g), int(b)
	return color, nil
}

func color24BitToString(c term.Color24Bit) string {
	return fmt.Sprintf("%2x%2x%2x", c.R, c.G, c.B)
}

func parseColor(s string) (term.Color, error) {
	if color, err := parseColor3Bit(s); err == nil {
		return color, nil
	} else if color, err := parseColor8Bit(s); err == nil {
		return color, nil
	} else if color, err := parseColor24Bit(s); err == nil {
		return color, nil
	} else {
		return nil, fmt.Errorf("Could not parse color: %s", s)
	}
}

func colorToString(color term.Color) string {
	if c, ok := color.(term.Color3Bit); ok {
		return color3BitToString(c)
	} else if c, ok := color.(term.Color8Bit); ok {
		return color8BitToString(c)
	} else if c, ok := color.(term.Color24Bit); ok {
		return color24BitToString(c)
	} else {
		return ""
	}
}

func parseGraphics(s string) (*term.Graphics, error) {
	g := term.Graphics{}
	parts := strings.Split(s, ":")
	for _, part := range parts {
		switch {
		case part == "bold":
			g.Bold = true
		case part == "reverse":
			g.Reverse = true
		case strings.HasPrefix(part, "fg#"):
			color, err := parseColor(part[3:])
			if err != nil {
				return nil, err
			}
			g.FgColor = color
		case strings.HasPrefix(part, "bg#"):
			color, err := parseColor(part[3:])
			if err != nil {
				return nil, err
			}
			g.BgColor = color
		default:
			return nil, fmt.Errorf("Could not parse graphics: %s", s)
		}
	}
	return &g, nil
}

func graphicsToString(g *term.Graphics) string {
	parts := []string{}
	if g.Bold {
		parts = append(parts, "bold")
	}
	if g.Reverse {
		parts = append(parts, "reverse")
	}
	if g.FgColor != nil {
		parts = append(parts, fmt.Sprint("fg#", colorToString(g.FgColor)))
	}
	if g.BgColor != nil {
		parts = append(parts, fmt.Sprint("bg#", colorToString(g.BgColor)))
	}
	return strings.Join(parts, ":")
}
