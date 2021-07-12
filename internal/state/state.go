package state

import (
	"path/filepath"
	"regexp"

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

func (s *State) AutoExpand(maxDepth int, ignore *regexp.Regexp) error {
	return s.Root.Traverse(false, nil, func(tree *filetree.FileTree, depth int) error {
		if maxDepth >= 0 && depth >= maxDepth {
			return nil
		}
		if tree == s.Root {
			return tree.Expand()
		}
		parent := tree.Parent()
		if parent != nil && !parent.Expanded() {
			return nil
		}

		if ignore != nil {
			rel, err := filepath.Rel(s.Root.AbsPath, tree.AbsPath)
			if err != nil {
				return err
			}
			if ignore.MatchString(rel) {
				return nil
			}
		}
		return tree.Expand()
	})
}
