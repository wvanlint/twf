package filetree

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	tree, err := InitFileTree("testdata")
	assert.Nil(t, err)
	assert.Equal(t, "testdata", tree.Name())
}

func TestIsDir(t *testing.T) {
	tree, err := InitFileTree("testdata")
	assert.Nil(t, err)
	assert.True(t, tree.IsDir())
}

func TestChildren(t *testing.T) {
	tree, err := InitFileTree("testdata")
	assert.Nil(t, err)
	children, err := tree.Children(nil)
	assert.Nil(t, err)
	childrenNames := []string{}
	for _, childTree := range children {
		childrenNames = append(childrenNames, childTree.Name())
	}
	assert.ElementsMatch(t, []string{"dir1", "dir2", "a"}, childrenNames)
}

func TestFindPath(t *testing.T) {
	root, err := InitFileTree("testdata")
	assert.Nil(t, err)

	node, err := root.FindPath("dir1/b")
	assert.Nil(t, err)
	assert.Equal(t, "b", node.Name())

	node, err = root.FindPath(".")
	assert.Nil(t, err)
	assert.Equal(t, root, node)

	wd, err := os.Getwd()
	assert.Nil(t, err)
	node, err = root.FindPath(filepath.Join(wd, "testdata/dir1/b"))
	assert.Nil(t, err)
	assert.Equal(t, "b", node.Name())

	node, err = root.FindPath(filepath.Join(wd, "testdata"))
	assert.Nil(t, err)
	assert.Equal(t, root, node)
}

func TestTraverse(t *testing.T) {
	names := []string{}
	root, err := InitFileTree("testdata")
	assert.Nil(t, err)
	err = root.Traverse(false, nil, func(node *FileTree, _ int) error {
		names = append(names, node.Name())
		return nil
	})
	assert.ElementsMatch(t, []string{"testdata", "dir1", "dir2", "a", "b", "c"}, names)
}

func TestByTypeAndName(t *testing.T) {
	nodes := []*FileTree{}
	root, err := InitFileTree("testdata")
	assert.Nil(t, err)
	err = root.Traverse(false, nil, func(node *FileTree, _ int) error {
		nodes = append(nodes, node)
		return nil
	})
	sort.Slice(nodes, ByTypeAndName(nodes))
	names := []string{}
	for _, node := range nodes {
		names = append(names, node.Name())
	}
	assert.Equal(t, []string{"dir1", "dir2", "testdata", "a", "b", "c"}, names)
}

func TestPrevNext(t *testing.T) {
	tree, err := InitFileTree("testdata")
	assert.Nil(t, err)
	names := []string{tree.Name()}
	for {
		next, err := tree.Next(false, ByTypeAndName)
		assert.Nil(t, err)
		if next == nil {
			break
		}
		tree = next
		names = append(names, tree.Name())
	}
	for {
		prev, err := tree.Prev(false, ByTypeAndName)
		assert.Nil(t, err)
		if prev == nil {
			break
		}
		tree = prev
		names = append(names, tree.Name())
	}
	assert.Equal(t, []string{"testdata", "dir1", "b", "dir2", "c", "a", "c", "dir2", "b", "dir1", "testdata"}, names)
}
