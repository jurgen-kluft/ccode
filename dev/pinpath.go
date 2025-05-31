package dev

import (
	"path/filepath"

	utils "github.com/jurgen-kluft/ccode/utils"
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
	return utils.PathGetRelativeTo(fp.String(), root)
}
