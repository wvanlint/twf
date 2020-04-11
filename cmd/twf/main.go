package main

import (
	"fmt"
	"strings"

	"github.com/wvanlint/twf/internal/config"
	"github.com/wvanlint/twf/internal/filetree"
	"github.com/wvanlint/twf/internal/state"
	"github.com/wvanlint/twf/internal/terminal"
	"github.com/wvanlint/twf/internal/views"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	config := config.GetConfig()

	if config.LogLevel != "" {
		var level zapcore.Level
		if err := level.UnmarshalText([]byte(config.LogLevel)); err != nil {
			panic(err)
		}
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
	} else {
		zap.ReplaceGlobals(zap.NewNop())
	}

	zap.L().Info("Starting twf.")

	tree, err := filetree.InitFileTree(config.Dir)
	if err != nil {
		panic(err)
	}
	state := state.State{
		Root:       tree,
		Cursor:     tree.AbsPath,
		Expansions: map[string]bool{tree.AbsPath: true},
	}
	views := []terminal.View{
		views.NewTreeView(config, &state),
		views.NewPreviewView(config, &state),
		views.NewStatusView(config, &state),
	}

	t, err := terminal.OpenTerm(&config.Terminal)
	if err != nil {
		panic(err)
	}
	err = t.StartLoop(config.Keybindings, views)
	t.Close()
	if err != nil {
		panic(err)
	}

	if len(state.Selection) > 0 {
		fmt.Println(strings.Join(state.Selection, "\n"))
	}
	zap.L().Info("Stopping twf.")
}
