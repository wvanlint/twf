package filetree

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileTree struct {
	AbsPath        string
	info           os.FileInfo
	targetInfo     os.FileInfo
	parent         *FileTree
	children       []*FileTree
	childrenByName map[string]*FileTree
}

func InitFileTree(p string) (*FileTree, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(p)
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

func (t *FileTree) Children() ([]*FileTree, error) {
	if err := t.maybeLoadChildren(); err != nil {
		return nil, err
	}
	return append(t.children[:0:0], t.children...), nil
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
		currentNode.maybeLoadChildren()
		currentNode, ok = currentNode.childrenByName[part]
		if !ok {
			return nil, PathNotFound{origPath}
		}
	}
	return currentNode, nil
}

func (t *FileTree) Traverse(f func(*FileTree)) error {
	queue := []*FileTree{t}
	var item *FileTree
	for len(queue) > 0 {
		item, queue = queue[0], queue[1:]
		f(item)
		children, err := item.Children()
		if err != nil {
			return err
		}
		queue = append(queue, children...)
	}
	return nil
}

func ByTypeAndName(children []*FileTree) func(i, j int) bool {
	return func(i, j int) bool {
		if children[i].IsDir() != children[j].IsDir() {
			return children[i].IsDir()
		} else {
			return children[i].Name() < children[j].Name()
		}
	}
}
