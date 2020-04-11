package main

import (
	"fmt"
	"strings"
)

type log struct {
	builder strings.Builder
}

func (l *log) Println(args ...interface{}) {
	fmt.Fprintln(&l.builder, args...)
}

var Log = log{strings.Builder{}}

func main() {
	defer func() { fmt.Println(Log.builder.String()) }()
	Log.Println("Starting up")
	tree, err := InitTreeFromWd()
	if err != nil {
		panic(err)
	}
	state := AppState{
		Root:       tree,
		Cursor:     tree.Path,
		Expansions: map[string]bool{tree.Path: true},
	}

	stop := make(chan bool, 1)
	t, err := InitTerm(Callbacks{
		ChangeCursor: state.ChangeCursor,
		Open:         func() { state.SetExpansion(state.Cursor, true) },
		Close:        func() { state.SetExpansion(state.Cursor, false) },
		Toggle:       func() { state.ToggleExpansion(state.Cursor) },
		ToggleAll:    func() { state.ToggleExpansionAll(state.Cursor) },
		OpenAll:      func() { state.SetExpansionAll(state.Cursor, true) },
		CloseAll:     func() { state.SetExpansionAll(state.Cursor, false) },
		Up:           func() { state.MoveCursorUp() },
		Quit:         func() { stop <- true },
	})
	if err != nil {
		panic(err)
	}
	defer t.Close()
	t.StartLoop(&state, stop)
}
