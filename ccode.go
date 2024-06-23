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

var ccode_dev = "tundra"
var ccode_os = runtime.GOOS
var ccode_arch = runtime.GOARCH

// Init will initialize ccode before anything else is run
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

	if denv.GetDevEnum(ccode_dev) == denv.INVALID {
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

	fmt.Printf("finished generating %s for os:%s, arch:%s\n", ccode_dev, ccode_os, ccode_arch)

	return true
}

// Generate is the main function that requires 'arguments' to then generate
// workspace and project files for a specified IDE.
func Generate(pkg *denv.Package) error {
	generator := axe.NewAxeGenerator(ccode_dev, ccode_os, ccode_arch)
	if generator.IsVisualStudio() {
		return generator.GenerateMsDev(pkg)
	} else if generator.IsTundra() {
		return generator.GenerateTundra(pkg)
	} else if generator.IsMake() {
		return generator.GenerateMake(pkg)
	} else if generator.IsCMake() {
		return generator.GenerateCMake(pkg)
	} else if generator.IsXCode() {
		return generator.GenerateXcode(pkg)
	}
	return fmt.Errorf("unknown dev '%s', should be one of tundra, cmake, xcode or vs2022", ccode_dev)
}

// DEV is an enumeration for all possible IDE's that are supported
type GenerateFile int

// All development environment
const (
	CLANGFORMAT GenerateFile = 0x20000
	GITIGNORE   GenerateFile = 0x40000
	MAINTEST    GenerateFile = 0x80000
	EMBEDDED    GenerateFile = 0x100000
	ALL         GenerateFile = CLANGFORMAT | GITIGNORE | MAINTEST | EMBEDDED
	INVALID     GenerateFile = 0x0
)

func GenerateSpecificFiles(files GenerateFile) {
	if files&CLANGFORMAT == CLANGFORMAT {
		embedded.WriteClangFormat(false)
	}
	if files&GITIGNORE == GITIGNORE {
		embedded.WriteGitIgnore(false)
	}
	if files&MAINTEST == MAINTEST {
		embedded.WriteTestMainCpp(true)
	}
	if files&EMBEDDED == EMBEDDED {
		embedded.WriteEmbedded()
	}
}

func GenerateFiles() {
	GenerateSpecificFiles(ALL)
}

func GenerateCppEnums(inputFile string, outputFile string) error {
	return embedded.GenerateCppEnums(inputFile, outputFile)
}
