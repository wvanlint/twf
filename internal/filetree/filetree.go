package filetree

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type FileTree struct {
	AbsPath        string
	info           os.FileInfo
	targetInfo     os.FileInfo
	parent         *FileTree
	children       []*FileTree
	childrenByName map[string]*FileTree
	expanded       bool
}

func InitFileTree(p string) (*FileTree, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(abs)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	tree := &FileTree{
		AbsPath: abs,
		info:    info,
	}
	return tree, nil
}

func (t *FileTree) Name() string {
	return t.info.Name()
}

func (t *FileTree) IsDir() bool {
	if t.targetInfo != nil {
		return t.targetInfo.IsDir()
	} else {
		return t.info.IsDir()
	}
}

func (t *FileTree) Parent() *FileTree {
	return t.parent
}

func (t *FileTree) Expanded() bool {
	return t.expanded
}

func (t *FileTree) Expand() {
	t.expanded = true
}

func (t *FileTree) Collapse() {
	t.expanded = false
}

func (t *FileTree) maybeLoadChildren() error {
	if t.children != nil {
		return nil
	}
	t.children = []*FileTree{}
	t.childrenByName = map[string]*FileTree{}
	if !t.IsDir() {
		return nil
	}
	f, err := os.Open(t.AbsPath)
	if err != nil {
		return err
	}
	defer f.Close()
	contents, err := f.Readdir(0)
	if err != nil {
		return err
	}
	for _, content := range contents {
		childFileTree := &FileTree{
			AbsPath: filepath.Join(t.AbsPath, content.Name()),
			info:    content,
			parent:  t,
		}
		if content.Mode()&os.ModeSymlink != 0 {
			if targetInfo, err := os.Stat(childFileTree.AbsPath); err == nil {
				childFileTree.targetInfo = targetInfo
			} else {
				return err
			}
		}
		t.children = append(t.children, childFileTree)
		t.childrenByName[childFileTree.Name()] = childFileTree
	}
	return nil
}

func (t *FileTree) Children(order Order) ([]*FileTree, error) {
	if err := t.maybeLoadChildren(); err != nil {
		return nil, err
	}
	children := append(t.children[:0:0], t.children...)
	if order != nil {
		sort.Slice(children, order(children))
	}
	return children, nil
}

type PathNotFound struct {
	Path string
}

func (e PathNotFound) Error() string {
	return fmt.Sprint("Path not found: ", e.Path)
}

func (t *FileTree) FindPath(origPath string) (*FileTree, error) {
	var err error
	path := filepath.Clean(origPath)
	if filepath.IsAbs(path) {
		path, err = filepath.Rel(t.AbsPath, path)
		if err != nil {
			return nil, PathNotFound{origPath}
		}
	}
	parts := []string{}
	for base := filepath.Base(path); base != "."; base = filepath.Base(path) {
		path = filepath.Dir(path)
		parts = append([]string{base}, parts...)
	}
	currentNode := t
	var ok bool
	for _, part := range parts {
		err := currentNode.maybeLoadChildren()
		if err != nil {
			return nil, err
		}
		currentNode, ok = currentNode.childrenByName[part]
		if !ok {
			return nil, PathNotFound{origPath}
		}
	}
	return currentNode, nil
}

func (t *FileTree) Traverse(visibleOnly bool, order Order, f func(*FileTree, int)) error {
	type treeWithDepth struct {
		tree  *FileTree
		depth int
	}
	stack := []treeWithDepth{treeWithDepth{t, 0}}
	for len(stack) > 0 {
		var current treeWithDepth
		current, stack = stack[len(stack)-1], stack[:len(stack)-1]
		f(current.tree, current.depth)

		if !visibleOnly || current.tree.Expanded() {
			children, err := current.tree.Children(order)
			if err != nil {
				return err
			}
			for i := len(children) - 1; i >= 0; i-- {
				stack = append(stack, treeWithDepth{children[i], current.depth + 1})
			}
		}
	}
	return nil
}

func (t *FileTree) Prev(visibleOnly bool, order Order) (*FileTree, error) {
	if t.Parent() == nil {
		return nil, nil
	}
	siblings, err := t.Parent().Children(order)
	if err != nil {
		return nil, err
	}
	if t == siblings[0] {
		return t.Parent(), nil
	}
	var prevSibling *FileTree
	for i, sibling := range siblings {
		if sibling == t {
			prevSibling = siblings[i-1]
		}
	}
	node := prevSibling
	for {
		if !node.Expanded() && visibleOnly {
			return node, nil
		}
		children, err := node.Children(order)
		if err != nil {
			return nil, err
		}
		if len(children) == 0 {
			return node, nil
		} else {
			node = children[len(children)-1]
		}
	}
	return nil, nil
}

func (t *FileTree) Next(visibleOnly bool, order Order) (*FileTree, error) {
	if t.Expanded() || !visibleOnly {
		children, err := t.Children(order)
		if err != nil {
			return nil, err
		}
		if len(children) > 0 {
			return children[0], nil
		}
	}
	node := t
	for node.Parent() != nil {
		siblings, err := node.Parent().Children(order)
		if err != nil {
			return nil, err
		}
		for i, sibling := range siblings {
			if sibling == node && i < len(siblings)-1 {
				return siblings[i+1], nil
			}
		}
		node = node.Parent()
	}
	return nil, nil
}

type Order func([]*FileTree) func(i, j int) bool

func ByTypeAndName(children []*FileTree) func(i, j int) bool {
	return func(i, j int) bool {
		if children[i].IsDir() != children[j].IsDir() {
			return children[i].IsDir()
		} else {
			return children[i].Name() < children[j].Name()
		}
	}
}
