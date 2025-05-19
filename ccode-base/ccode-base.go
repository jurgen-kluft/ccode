package ccode

import (
	"flag"
	"fmt"
	"runtime"
	"strings"

	ccode_gen "github.com/jurgen-kluft/ccode/ccode-gen"
	"github.com/jurgen-kluft/ccode/denv"
	"github.com/jurgen-kluft/ccode/embedded"
)

// Init will initialize ccode before anything else is run

// tundra, vs2022, make, cmake, xcode, espmake
var ccode_dev = "tundra"

// win32, win64, linux32, linux64, macos64
var ccode_os = runtime.GOOS

// x64, arm64, amd64, 386, esp32 / esp32c3 / esp32s3
var ccode_arch = runtime.GOARCH

// verbose
var ccode_verbose = false

func Init() bool {

	flag.StringVar(&ccode_dev, "dev", "tundra", "the build system to generate projects for (vs2022, tundra, make, cmake, xcode, espmake)")
	flag.BoolVar(&ccode_verbose, "verbose", false, "verbose output")
	flag.Parse()

    // Currently supported: esp32, esp32s3
	if strings.HasPrefix(ccode_dev, "esp32") {
        ccode_os = "arduino"
        ccode_arch = "invalid"
        if strings.EqualFold(ccode_dev, "esp32") {
            ccode_arch = "esp32"
        } else if strings.EqualFold(ccode_dev, "esp32s3") {
            ccode_arch = "esp32s3"
        }
	}

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

	if ccode_gen.DevEnumFromString(ccode_dev) == ccode_gen.DevInvalid {
		fmt.Println()
		fmt.Println("Error, wrong parameter for '-dev', '", ccode_dev, "' is not recognized")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("    -> Usage: ccode -dev=vs2022/vs2019/vs2015")
		fmt.Println("    -> Usage: ccode -dev=tundra")
		fmt.Println("    -> Usage: ccode -dev=make")
		fmt.Println("    -> Usage: ccode -dev=xcode")
		fmt.Println("    -> Usage: ccode -dev=espmake")
		return false
	}

	return true
}

// Generate is the main function that requires 'arguments' to then generate
// workspace and project files for a specified IDE.
func Generate(pkg *denv.Package) error {
	generator := ccode_gen.NewGenerator(ccode_dev, ccode_os, ccode_arch, ccode_verbose)
	return generator.Generate(pkg)
}

func GenerateGitIgnore() {
	embedded.WriteGitIgnore(false)
}

func GenerateTestMainCpp(ccore, cbase bool) {
	embedded.WriteTestMainCpp(ccore, cbase, true)
}

func GenerateEmbedded() {
	embedded.WriteEmbedded()
}

func GenerateClangFormat() {
	embedded.WriteClangFormat(false)
}

func GenerateFiles(pkg *denv.Package) {

	// Analyze the package to see if it has dependencies on:
	// - ccore
	// - cbase
	// If it only has a dependency on ccore, we should generate a TestMainCpp that is compatible with only ccore, if
	// however it is depending on cbase then we can use the TestMainCpp that is compatible with cbase.
	//
	// But, if there is no dependency on ccore or cbase, we should generate a TestMainCpp that can work without any
	// ccore or cbase functionality
	//
	has_ccore := pkg.HasDependencyOn("ccore")
	has_cbase := pkg.HasDependencyOn("cbase")

	GenerateGitIgnore()
	GenerateTestMainCpp(has_ccore, has_cbase)
	GenerateEmbedded()
	GenerateClangFormat()
}

func GenerateCppEnums(inputFile string, outputFile string) error {
	return embedded.GenerateCppEnums(inputFile, outputFile)
}
