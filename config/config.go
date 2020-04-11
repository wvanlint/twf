package config

import (
	term "github.com/wvanlint/twf/terminal"
)

type TwfConfig struct {
	Preview  PreviewConfig
	TreeView TreeViewConfig
	Terminal term.TerminalConfig
}

type PreviewConfig struct {
	Enabled        bool
	PreviewCommand string
}

type TreeViewConfig struct {
	Graphics    GraphicsConfig
	FindCommand string
}

type GraphicsConfig map[string]term.Graphics

var defaultConfig = TwfConfig{
	TreeView: TreeViewConfig{
		Graphics: map[string]term.Graphics{
			"dir": term.Graphics{
				FgColor: term.Color3Bit{Value: 4, Bright: true},
				Bold:    true,
			},
			"cursor": term.Graphics{
				Reverse: true,
			},
		},
		FindCommand: "fzf",
	},
	Preview: PreviewConfig{
		Enabled:        true,
		PreviewCommand: "bat {}",
	},
	Terminal: term.TerminalConfig{
		Height: 0.2,
	},
}

func GetConfig() *TwfConfig {
	return &defaultConfig
}

//func color3BitFromString(s string) (Color, error) {
//	c := Color3Bit{}
//	if strings.HasPrefix(s, "bright") {
//		s = s[len("bright"):]
//		c.Bright = true
//	}
//	switch s {
//	case "black":
//		c.Value = 0
//	case "red":
//		c.Value = 1
//	case "green":
//		c.Value = 2
//	case "yellow":
//		c.Value = 3
//	case "blue":
//		c.Value = 4
//	case "magenta":
//		c.Value = 5
//	case "cyan":
//		c.Value = 6
//	case "white":
//		c.Value = 7
//	default:
//		return nil, errors.New("Could not interpret color string.")
//	}
//	return &c, nil
//}
//
//func ColorFromString(s string) (Color, error) {
//	return Color3BitFromString(s)
//}
