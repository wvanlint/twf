package main

type AppState struct {
	Root       *Tree
	Cursor     string
	ScrollTop  int
	Expansions map[string]bool
}

func (s *AppState) ChangeCursor(path string) {
	s.Cursor = path
}

func (s *AppState) ToggleExpansion(path string) {
	value, _ := s.Expansions[path]
	if value {
		delete(s.Expansions, path)
	} else {
		s.Expansions[path] = true
	}
}
