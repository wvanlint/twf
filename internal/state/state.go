package state

import (
	"github.com/wvanlint/twf/internal/filetree"
)

type State struct {
	Root      *filetree.FileTree
	Cursor    *filetree.FileTree
	Selection []*filetree.FileTree
}

func (s *State) LocatePath(path string) error {
	node, err := s.Root.FindPath(path)
	if err != nil {
		return err
	}
	s.Cursor = node
	for node.Parent() != nil {
		node = node.Parent()
		err = node.Expand()
		if err != nil {
			return err
		}
	}
	return nil
}
