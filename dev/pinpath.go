package dev

import (
	"path/filepath"

	"github.com/jurgen-kluft/ccode/foundation"
)

type PinPath struct {
	Root string // Root directory
	Base string // Base directory
	Sub  string // Sub directory
}

func (fp PinPath) String() string {
	return filepath.Join(fp.Root, fp.Base, fp.Sub)
}

func (fp PinPath) RelativeTo(root string) string {
	return foundation.PathGetRelativeTo(fp.String(), root)
}
