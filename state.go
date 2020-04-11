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

func (s *AppState) MoveCursorUp() {
	parent := s.Root.GetPath(s.Cursor).Parent()
	if parent != nil {
		s.Cursor = parent.Path
	}
}

func (s *AppState) SetExpansion(path string, value bool) {
	if value {
		s.Expansions[path] = true
	} else {
		delete(s.Expansions, path)
	}
}

func (s *AppState) ToggleExpansion(path string) {
	value, _ := s.Expansions[path]
	s.SetExpansion(path, !value)
}

func (s *AppState) SetExpansionAll(path string, value bool) {
	tree := s.Root.GetPath(path)
	tree.Traverse(func(node *Tree) {
		s.SetExpansion(node.Path, value)
	})
}

func (s *AppState) ToggleExpansionAll(path string) {
	value, _ := s.Expansions[path]
	s.SetExpansionAll(path, !value)
}
