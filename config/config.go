package config

import (
	"flag"
	"fmt"
	"strings"

	term "github.com/wvanlint/twf/terminal"
)

type TwfConfig struct {
	Preview     PreviewConfig
	TreeView    TreeViewConfig
	Terminal    term.TerminalConfig
	Graphics    GraphicsMapping
	Keybindings Keybindings
}

type PreviewConfig struct {
	Enabled        bool
	PreviewCommand string
}

type TreeViewConfig struct {
	FindCommand string
}

type GraphicsMapping map[string]*term.Graphics

func NewGraphicsMapping() GraphicsMapping {
	return make(map[string]*term.Graphics)
}

func (m GraphicsMapping) String() string {
	pairStrs := []string{}
	for key, graphics := range m {
		pairStrs = append(pairStrs, fmt.Sprint(key, "::", graphicsToString(graphics)))
	}
	return strings.Join(pairStrs, ",")
}

func (m GraphicsMapping) Set(s string) error {
	parts := strings.Split(s, ",")
	for _, part := range parts {
		pair := strings.Split(part, "::")
		if len(pair) != 2 {
			return fmt.Errorf("Unexpected graphics configuration string: %s", s)
		}
		g, err := parseGraphics(pair[1])
		if err != nil {
			return err
		}
		m[pair[1]] = g
	}
	return nil
}

func defaultGraphicsMapping() GraphicsMapping {
	return map[string]*term.Graphics{
		"tree:dir": &term.Graphics{
			FgColor: term.Color3Bit{Value: 4, Bright: true},
			Bold:    true,
		},
		"tree:cursor": &term.Graphics{
			Reverse: true,
		},
	}
}

type Keybindings map[string][]string

func NewKeybindings() Keybindings {
	return make(map[string][]string)
}

func (ks Keybindings) String() string {
	bindingStrs := []string{}
	for hash, cmds := range ks {
		bindingStrs = append(
			bindingStrs,
			fmt.Sprint(
				eventHashKeyToString(hash),
				"::",
				strings.Join(cmds, ";"),
			),
		)
	}
	return strings.Join(bindingStrs, ",")
}

func (ks Keybindings) Set(s string) error {
	bindingStrs := strings.Split(s, ",")
	for _, bindingStr := range bindingStrs {
		pair := strings.Split(bindingStr, "::")
		if len(pair) != 2 {
			return fmt.Errorf("Unexpected keybinding string: %s", bindingStr)
		}
		event, err := parseEvent(pair[0])
		if err != nil {
			return err
		}
		ks[event.HashKey()] = strings.Split(pair[1], ";")
	}
	return nil
}

func defaultKeybindings() Keybindings {
	return map[string][]string{
		(&term.Event{term.Rune, 'j'}).HashKey():     []string{"tree:next"},
		(&term.Event{term.Rune, 'k'}).HashKey():     []string{"tree:prev"},
		(&term.Event{term.Rune, 'o'}).HashKey():     []string{"tree:toggle"},
		(&term.Event{term.Rune, 'O'}).HashKey():     []string{"tree:toggleAll"},
		(&term.Event{term.Rune, '/'}).HashKey():     []string{"tree:findExternal"},
		(&term.Event{term.Rune, 'q'}).HashKey():     []string{"quit"},
		(&term.Event{Symbol: term.Enter}).HashKey(): []string{"tree:selectPath", "quit"},
	}
}

func GetConfig() *TwfConfig {
	config := TwfConfig{}
	flag.StringVar(
		&config.Preview.PreviewCommand,
		"previewCmd",
		"cat {}",
		"Command to create preview of a file.",
	)
	flag.BoolVar(
		&config.Preview.Enabled,
		"preview",
		true,
		"Enable/disable previews.",
	)
	flag.StringVar(
		&config.TreeView.FindCommand,
		"findCmd",
		"fzf",
		"External command to select a path to locate.",
	)
	flag.Float64Var(
		&config.Terminal.Height,
		"height",
		1.0,
		"Proportion of the vertical space to take up.",
	)
	config.Keybindings = defaultKeybindings()
	flag.Var(
		config.Keybindings,
		"bind",
		"Keybindings for command sequences",
	)
	config.Graphics = defaultGraphicsMapping()
	flag.Var(
		config.Graphics,
		"graphics",
		"Graphics per type of text span.",
	)
	flag.Parse()
	return &config
}
