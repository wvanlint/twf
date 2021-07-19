package state

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wvanlint/twf/internal/filetree"
)

func TestAutoExpand(t *testing.T) {
	for curDepth := -1; curDepth < 3; curDepth++ {
		tree, err := filetree.InitFileTree("../filetree/testdata")
		assert.Nil(t, err)
		state := &State{Root: tree}
		err = state.AutoExpand(curDepth, nil)
		assert.Nil(t, err)
		state.Root.Traverse(false, nil, func(node *filetree.FileTree, nodeDepth int) error {
			if !node.IsDir() {
				return nil
			}
			assert.Equal(t, curDepth == -1 || nodeDepth <= curDepth-1, node.Expanded())
			return nil
		})
	}
}

func TestAutoExpandRegex(t *testing.T) {
	re := regexp.MustCompile("dir1")
	for curDepth := -1; curDepth < 3; curDepth++ {
		tree, err := filetree.InitFileTree("../filetree/testdata")
		assert.Nil(t, err)
		state := &State{Root: tree}
		err = state.AutoExpand(curDepth, re)
		assert.Nil(t, err)
		state.Root.Traverse(false, nil, func(node *filetree.FileTree, nodeDepth int) error {
			if !node.IsDir() {
				return nil
			}
			if node.Name() == "dir1" {
				assert.False(t, node.Expanded())
				return nil
			}
			assert.Equal(t, curDepth == -1 || nodeDepth <= curDepth-1, node.Expanded())
			return nil
		})
	}

	re = regexp.MustCompile(".*")
	for curDepth := -1; curDepth < 3; curDepth++ {
		tree, err := filetree.InitFileTree("../filetree/testdata")
		assert.Nil(t, err)
		state := &State{Root: tree}
		err = state.AutoExpand(curDepth, re)
		assert.Nil(t, err)
		state.Root.Traverse(false, nil, func(node *filetree.FileTree, nodeDepth int) error {
			assert.Equal(t, curDepth != 0 && nodeDepth == 0, node.Expanded())
			return nil
		})
	}
}
