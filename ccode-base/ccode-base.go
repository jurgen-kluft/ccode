package ccode

import (
	"flag"
	"fmt"
	"runtime"
	"strings"

	"github.com/jurgen-kluft/ccode/axe"
	"github.com/jurgen-kluft/ccode/denv"
	"github.com/jurgen-kluft/ccode/embedded"
)

// Init will initialize ccode before anything else is run
var ccode_dev = "tundra"
var ccode_os = runtime.GOOS
var ccode_arch = runtime.GOARCH

func Init() bool {

	flag.StringVar(&ccode_dev, "dev", "tundra", "the build system to generate projects for (vs2022, tundra, make, cmake, xcode)")
	flag.Parse()

	if ccode_os == "" {
		ccode_os = strings.ToLower(runtime.GOOS)
	}
	if ccode_arch == "" {
		ccode_arch = strings.ToLower(runtime.GOARCH)
	}
	if ccode_dev == "" {
		ccode_dev = "tundra"
		if ccode_os == "windows" {
			ccode_dev = "vs2022"
		}
	}

	fmt.Println("ccode, a tool to generate C/C++ workspace and project files")

	if axe.GetDevEnum(ccode_dev) == axe.DevInvalid {
		fmt.Println()
		fmt.Println("Error, wrong parameter for '-dev', '", ccode_dev, "' is not recognized")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("    -> Usage: ccode -dev=vs2022/vs2019/vs2015")
		fmt.Println("    -> Usage: ccode -dev=tundra")
		fmt.Println("    -> Usage: ccode -dev=make")
		fmt.Println("    -> Usage: ccode -dev=xcode")
		return false
	}

	return true
}

// Generate is the main function that requires 'arguments' to then generate
// workspace and project files for a specified IDE.
func Generate(pkg *denv.Package) error {
	generator := axe.NewAxeGenerator(ccode_dev, ccode_os, ccode_arch)
	return generator.Generate(pkg)
}

func GenerateGitIgnore() {
	embedded.WriteGitIgnore(false)
}

func GenerateTestMainCpp() {
	embedded.WriteTestMainCpp(true)
}

func GenerateEmbedded() {
	embedded.WriteEmbedded()
}

func GenerateClangFormat() {
	embedded.WriteClangFormat(false)
}

func GenerateFiles() {
	GenerateGitIgnore()
	GenerateTestMainCpp()
	GenerateEmbedded()
	GenerateClangFormat()
}

func GenerateCppEnums(inputFile string, outputFile string) error {
	return embedded.GenerateCppEnums(inputFile, outputFile)
}
