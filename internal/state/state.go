package state

import (
	"github.com/wvanlint/twf/internal/filetree"
)

type State struct {
	Root       *filetree.FileTree
	Cursor     string
	Expansions map[string]bool
	Selection  []string
}

func (s *State) ChangeCursor(path string) {
	node, _ := s.Root.FindPath(path)
	s.Cursor = node.AbsPath
	for node != s.Root {
		node = node.Parent()
		s.SetExpansion(node.AbsPath, true)
	}
}

func (s *State) MoveCursorToParent() {
	node, _ := s.Root.FindPath(s.Cursor)
	parent := node.Parent()
	if parent != nil {
		s.Cursor = parent.AbsPath
	}
}

func (s *State) SetExpansion(path string, value bool) {
	if value {
		s.Expansions[path] = true
	} else {
		delete(s.Expansions, path)
	}
}

func (s *State) ToggleExpansion(path string) {
	value, _ := s.Expansions[path]
	s.SetExpansion(path, !value)
}

func (s *State) SetExpansionAll(path string, value bool) {
	tree, _ := s.Root.FindPath(path)
	tree.Traverse(func(node *filetree.FileTree) {
		s.SetExpansion(node.AbsPath, value)
	})
}

func (s *State) ToggleExpansionAll(path string) {
	value, _ := s.Expansions[path]
	s.SetExpansionAll(path, !value)
}

func (s *State) AddSelection(path string) {
	s.Selection = append(s.Selection, path)
}
