package main

import (
	"fmt"
	"os"
	"path"
)

type Tree struct {
	path       string
	info       os.FileInfo
	targetInfo os.FileInfo
	Children   []*Tree
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
		info: info,
		path: ".",
	}
	return tree, nil
}

func (t *Tree) IsDir() bool {
	if t.targetInfo != nil {
		return t.targetInfo.IsDir()
	} else {
		return t.info.IsDir()
	}
}

func (t *Tree) MaybeExpand() error {
	if !t.IsDir() || t.Children != nil {
		return nil
	}
	f, err := os.Open(t.path)
	if err != nil {
		return err
	}
	contents, err := f.Readdir(0)
	if err != nil {
		return err
	}
	for _, content := range contents {
		childTree := &Tree{
			path: path.Join(t.path, content.Name()),
			info: content,
		}
		if content.Mode()&os.ModeSymlink != 0 {
			if targetInfo, err := os.Stat(childTree.path); err != nil {
				childTree.targetInfo = targetInfo
			}
		}
		t.Children = append(t.Children, childTree)
	}
	return nil
}

func (t *Tree) Print(indent string) {
	fmt.Println(indent + t.info.Name())
	err := t.MaybeExpand()
	if err != nil {
		panic(err)
	}
	for _, child := range t.Children {
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
