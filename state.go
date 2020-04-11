package main

type AppState struct {
	Root       *Tree
	CursorLine int
	ScrollTop  int
	Expansions map[*Tree]bool
}
