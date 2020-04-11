package main

import (
	"fmt"
	"strings"

	"github.com/wvanlint/twf/config"
	"github.com/wvanlint/twf/terminal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	config := config.GetConfig()

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
	views := []terminal.View{
		NewTreeView(config, &state),
		NewPreviewView(config, &state),
		NewStatusView(config, &state),
	}

	t, err := terminal.OpenTerm(&config.Terminal)
	if err != nil {
		panic(err)
	}
	err = t.StartLoop(config.Keybindings, views)
	if err != nil {
		panic(err)
	}
	t.Close()

	fmt.Println(strings.Join(state.Selection, "\n"))
	zap.L().Info("Stopping twf.")
}
