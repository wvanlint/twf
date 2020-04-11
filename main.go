package main

import (
	"github.com/wvanlint/twf/terminal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger, err := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Encoding:    "console",
		OutputPaths: []string{"/tmp/twf.log"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			LevelKey:    "level",
			TimeKey:     "time",
			EncodeLevel: zapcore.CapitalColorLevelEncoder,
			EncodeTime:  zapcore.RFC3339TimeEncoder,
		},
	}.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)
	zap.L().Info("Starting twf.")

	tree, err := InitTreeFromWd()
	if err != nil {
		panic(err)
	}
	state := AppState{
		Root:       tree,
		Cursor:     tree.Path,
		Expansions: map[string]bool{tree.Path: true},
	}
	treeView := TreeView{state: &state}
	views := []terminal.View{
		&treeView,
		&PreviewView{state: &state},
		&StatusView{state: &state},
	}
	bindings := map[string]string{
		(&terminal.Event{terminal.Rune, 'j'}).HashKey(): "tree:next",
		(&terminal.Event{terminal.Rune, 'k'}).HashKey(): "tree:prev",
		(&terminal.Event{terminal.Rune, 'o'}).HashKey(): "tree:toggle",
		(&terminal.Event{terminal.Rune, 'O'}).HashKey(): "tree:toggleAll",
		(&terminal.Event{terminal.Rune, '/'}).HashKey(): "tree:findExternal",
		(&terminal.Event{terminal.Rune, 'q'}).HashKey(): "quit",
	}

	t, err := terminal.OpenTerm()
	if err != nil {
		panic(err)
	}
	defer t.Close()
	err = t.StartLoop(bindings, views)
	if err != nil {
		panic(err)
	}
	zap.L().Info("Stopping twf.")
}
