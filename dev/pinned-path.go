package dev

import (
	"path/filepath"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

type PinnedPath struct {
	Root string // Root directory
	Base string // Base directory
	Sub  string // Sub directory
}

func (fp PinnedPath) String() string {
	return filepath.Join(fp.Root, fp.Base, fp.Sub)
}

func (fp PinnedPath) RelativeTo(root string) string {
	return corepkg.PathGetRelativeTo(fp.String(), root)
}
