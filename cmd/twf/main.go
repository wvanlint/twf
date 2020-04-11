package main

import (
	"fmt"
	"os"
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

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	tree, err := filetree.InitFileTree(wd)
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

	fmt.Println(strings.Join(state.Selection, "\n"))
	zap.L().Info("Stopping twf.")
}
