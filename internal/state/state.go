package state

import (
	"github.com/wvanlint/twf/internal/filetree"
)

type State struct {
	Root      *filetree.FileTree
	Cursor    *filetree.FileTree
	Selection []*filetree.FileTree
}
