package toolchain

import (
	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
	"github.com/jurgen-kluft/ccode/denv"
)

type Environment interface {
	NewCompiler(config denv.BuildConfig, target denv.BuildTarget) Compiler
	NewArchiver(a ArchiverType, config denv.BuildConfig, target denv.BuildTarget) Archiver
	NewLinker(config denv.BuildConfig, target denv.BuildTarget) Linker
	NewFileCommander(config denv.BuildConfig, target denv.BuildTarget) FileCommander
	NewBurner(config denv.BuildConfig, target denv.BuildTarget) Burner
	NewDependencyTracker(buildPath string) deptrackr.FileTrackr
}
