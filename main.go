package main

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

	stop := make(chan bool, 1)
	t, err := InitTerm(Callbacks{
		ChangeCursor:    state.ChangeCursor,
		ToggleExpansion: state.ToggleExpansion,
		Quit:            func() { stop <- true },
	})
	if err != nil {
		panic(err)
	}
	defer t.Close()
	t.StartLoop(&state, stop)
}
