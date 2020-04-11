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

	stop := make(chan bool, 1)
	t, err := terminal.InitTerm(terminal.Callbacks{
		ChangeCursor: state.ChangeCursor,
		Prev: func() {
			p := treeView.GetPrevPath()
			state.ChangeCursor(p)
			state.ChangeScroll(treeView.ScrollForPath(p))
		},
		Next: func() {
			p := treeView.GetNextPath()
			state.ChangeCursor(p)
			state.ChangeScroll(treeView.ScrollForPath(p))
		},
		Open:      func() { state.SetExpansion(state.Cursor, true) },
		Close:     func() { state.SetExpansion(state.Cursor, false) },
		Toggle:    func() { state.ToggleExpansion(state.Cursor) },
		ToggleAll: func() { state.ToggleExpansionAll(state.Cursor) },
		OpenAll:   func() { state.SetExpansionAll(state.Cursor, true) },
		CloseAll:  func() { state.SetExpansionAll(state.Cursor, false) },
		Up:        func() { state.MoveCursorToParent() },
		Quit:      func() { stop <- true },
	})
	if err != nil {
		panic(err)
	}
	defer t.Close()
	err = t.StartLoop(views, stop)
	if err != nil {
		panic(err)
	}
	zap.L().Info("Stopping twf.")
}
