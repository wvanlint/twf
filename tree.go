package main

import (
	"fmt"
	"os"
	"path"
)

type Tree struct {
	Path       string
	info       os.FileInfo
	targetInfo os.FileInfo
	children   []*Tree
}

func InitTreeFromWd() (*Tree, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	fmt.Println(wd)
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

func (t *Tree) Children() ([]*Tree, error) {
	if t.children != nil {
		return t.children, nil
	}
	if !t.IsDir() {
		return []*Tree{}, nil
	}
	f, err := os.Open(t.Path)
	if err != nil {
		return nil, err
	}
	contents, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}
	for _, content := range contents {
		childTree := &Tree{
			Path: path.Join(t.Path, content.Name()),
			info: content,
		}
		if content.Mode()&os.ModeSymlink != 0 {
			if targetInfo, err := os.Stat(childTree.Path); err == nil {
				childTree.targetInfo = targetInfo
			} else {
				return nil, err
			}
		}
		t.children = append(t.children, childTree)
	}
	return append(t.children[:0:0], t.children...), nil
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
	children, err := t.Children()
	if err != nil {
		panic(err)
	}
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
