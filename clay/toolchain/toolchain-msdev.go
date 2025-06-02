package toolchain

import (
	"fmt"

	"github.com/jurgen-kluft/ccode/clay/toolchain/dpenc"
)

type Msdev struct {
	Name  string
	Vars  *Vars
	Tools map[string]string
}

// Compiler options
//      https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/compiler-options-listed-by-category.md
// Linker options
//      https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/linker-options.md

func NewWindowsMsdev() (*Msdev, error) {
	return nil, fmt.Errorf("Msdev is not implemented yet")
}

func (ms *Msdev) NewCompiler(config *Config) Compiler {
	return nil
}
func (ms *Msdev) NewArchiver(a ArchiverType, config *Config) Archiver {
	return nil
}
func (ms *Msdev) NewLinker(config *Config) Linker {
	return nil
}
func (t *Msdev) NewBurner(config *Config) Burner {
	return &EmptyBurner{}
}
func (t *Msdev) NewDependencyTracker(dirpath string) dpenc.FileTrackr {
	return nil
}
