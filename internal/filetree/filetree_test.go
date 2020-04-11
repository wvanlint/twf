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
	children, err := tree.Children()
	assert.Nil(t, err)
	childrenNames := []string{}
	for _, childTree := range children {
		childrenNames = append(childrenNames, childTree.Name())
	}
	assert.ElementsMatch(t, []string{"dir", "a"}, childrenNames)
}

func TestFindPath(t *testing.T) {
	root, err := InitFileTree("testdata")
	assert.Nil(t, err)

	node, err := root.FindPath("dir/b")
	assert.Nil(t, err)
	assert.Equal(t, "b", node.Name())

	node, err = root.FindPath(".")
	assert.Nil(t, err)
	assert.Equal(t, root, node)

	wd, err := os.Getwd()
	assert.Nil(t, err)
	node, err = root.FindPath(filepath.Join(wd, "testdata/dir/b"))
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
	err = root.Traverse(func(node *FileTree) {
		names = append(names, node.Name())
	})
	assert.ElementsMatch(t, []string{"testdata", "dir", "a", "b"}, names)
}

func TestByTypeAndName(t *testing.T) {
	nodes := []*FileTree{}
	root, err := InitFileTree("testdata")
	assert.Nil(t, err)
	err = root.Traverse(func(node *FileTree) {
		nodes = append(nodes, node)
	})
	sort.Slice(nodes, ByTypeAndName(nodes))
	names := []string{}
	for _, node := range nodes {
		names = append(names, node.Name())
	}
	assert.Equal(t, []string{"dir", "testdata", "a", "b"}, names)
}
