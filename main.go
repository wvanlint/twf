package main

import (
	"github.com/wvanlint/twf/terminal"
)

func main() {
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

	stop := make(chan bool, 1)
	t, err := terminal.InitTerm(terminal.Callbacks{
		ChangeCursor: state.ChangeCursor,
		Prev:         func() { state.ChangeCursor(treeView.GetPrevPath()) },
		Next:         func() { state.ChangeCursor(treeView.GetNextPath()) },
		Open:         func() { state.SetExpansion(state.Cursor, true) },
		Close:        func() { state.SetExpansion(state.Cursor, false) },
		Toggle:       func() { state.ToggleExpansion(state.Cursor) },
		ToggleAll:    func() { state.ToggleExpansionAll(state.Cursor) },
		OpenAll:      func() { state.SetExpansionAll(state.Cursor, true) },
		CloseAll:     func() { state.SetExpansionAll(state.Cursor, false) },
		Up:           func() { state.MoveCursorToParent() },
		Quit:         func() { stop <- true },
	})
	if err != nil {
		panic(err)
	}
	defer t.Close()
	err = t.StartLoop(&treeView, stop)
	if err != nil {
		panic(err)
	}
}
