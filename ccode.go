package ccode

import (
	base "github.com/jurgen-kluft/ccode/ccode-base"
	"github.com/jurgen-kluft/ccode/denv"
)

// Init will initialize ccode before anything else is run
func Init() bool {
	return base.Init()
}

// Generate is the main function that requires 'arguments' to then generate
// workspace and project files for a specified IDE.
func Generate(pkg *denv.Package) error {
	return base.Generate(pkg)
}

func GenerateGitIgnore() {
	base.GenerateGitIgnore()
}

func GenerateTestMainCpp() {
	base.GenerateTestMainCpp()
}

func GenerateEmbedded() {
	base.GenerateEmbedded()
}

func GenerateClangFormat() {
	base.GenerateClangFormat()
}

func GenerateFiles() {
	base.GenerateFiles()
}

func GenerateCppEnums(inputFile string, outputFile string) error {
	return base.GenerateCppEnums(inputFile, outputFile)
}
