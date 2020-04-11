package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type Tree struct {
	Path           string
	info           os.FileInfo
	targetInfo     os.FileInfo
	parent         *Tree
	children       []*Tree
	childrenByName map[string]*Tree
}

func InitTreeFromWd() (*Tree, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	tree, err := InitTree(wd)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func InitTree(p string) (*Tree, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	tree := &Tree{
		Path: p,
		info: info,
	}
	return tree, nil
}

func (t *Tree) Name() string {
	return t.info.Name()
}

func (t *Tree) IsDir() bool {
	if t.targetInfo != nil {
		return t.targetInfo.IsDir()
	} else {
		return t.info.IsDir()
	}
}

func (t *Tree) Parent() *Tree {
	return t.parent
}

func (t *Tree) maybeLoadChildren() {
	if t.children != nil {
		return
	}
	t.children = []*Tree{}
	t.childrenByName = map[string]*Tree{}
	if !t.IsDir() {
		return
	}
	f, err := os.Open(t.Path)
	if err != nil {
		panic(err)
	}
	contents, err := f.Readdir(0)
	if err != nil {
		panic(err)
	}
	for _, content := range contents {
		childTree := &Tree{
			Path:   filepath.Join(t.Path, content.Name()),
			info:   content,
			parent: t,
		}
		if content.Mode()&os.ModeSymlink != 0 {
			if targetInfo, err := os.Stat(childTree.Path); err == nil {
				childTree.targetInfo = targetInfo
			} else {
				panic(err)
			}
		}
		t.children = append(t.children, childTree)
		t.childrenByName[childTree.Name()] = childTree
	}
}

func (t *Tree) Children() []*Tree {
	t.maybeLoadChildren()
	return append(t.children[:0:0], t.children...)
}

func (t *Tree) FindPath(path string) *Tree {
	var err error
	path = filepath.Clean(path)
	if path == t.Path {
		return t
	}
	if filepath.IsAbs(path) {
		path, err = filepath.Rel(t.Path, path)
		if err != nil {
			panic("Not found.")
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
		currentNode.maybeLoadChildren()
		currentNode, ok = currentNode.childrenByName[part]
		if !ok {
			panic("Not found: " + part)
		}
	}
	return currentNode
}

func (t *Tree) Traverse(f func(*Tree)) {
	queue := []*Tree{t}
	var item *Tree
	for len(queue) > 0 {
		item, queue = queue[0], queue[1:]
		f(item)
		queue = append(queue, item.Children()...)
	}
}

func ByTypeAndName(children []*Tree) func(i, j int) bool {
	return func(i, j int) bool {
		if children[i].IsDir() != children[j].IsDir() {
			return children[i].IsDir()
		} else {
			return children[i].Name() < children[j].Name()
		}
	}
}

func (t *Tree) Print(indent string) {
	fmt.Println(indent + t.info.Name())
	children := t.Children()
	for _, child := range children {
		child.Print(indent + "  ")
	}
}

func PrintWd() {
	tree, err := InitTreeFromWd()
	if err != nil {
		panic(err)
	}
	tree.Print("")
}
