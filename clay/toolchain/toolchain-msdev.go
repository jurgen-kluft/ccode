package toolchain

import (
	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
	corepkg "github.com/jurgen-kluft/ccode/core"
	"github.com/jurgen-kluft/ccode/denv"
)

type WinMsdev struct {
	Name string
	Vars *corepkg.Vars
}

func (ms *WinMsdev) NewCompiler(config denv.BuildConfig, target denv.BuildTarget) Compiler {
	return nil
}

func (ms *WinMsdev) NewArchiver(a ArchiverType, config denv.BuildConfig, target denv.BuildTarget) Archiver {
	return nil
}

func (ms *WinMsdev) NewLinker(config denv.BuildConfig, target denv.BuildTarget) Linker {
	return nil
}

func (ms *WinMsdev) NewBurner(config denv.BuildConfig, target denv.BuildTarget) Burner {
	return nil
}

func (ms *WinMsdev) NewDependencyTracker(dirpath string) deptrackr.FileTrackr {
	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Toolchain for Visual Studio on Windows

func NewWinMsdev(arch string, product string) (t *WinMsdev, err error) {
	return &WinMsdev{
		Name: "msdev",
		Vars: corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
	}, nil
}
