package dev

import (
	"path/filepath"

	utils "github.com/jurgen-kluft/ccode/utils"
)

type PinFilepath struct {
	Path     PinPath
	Filename string // Filename (without extension)
}

func (fp PinFilepath) String() string {
	return filepath.Join(fp.Path.Root, fp.Path.Base, fp.Path.Sub, fp.Filename)
}

func (fp PinFilepath) RelativeTo(root string) string {
	return utils.PathGetRelativeTo(fp.String(), root)
}
