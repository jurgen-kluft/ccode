package ccode

import (
	"flag"
	"fmt"
	"runtime"
	"strings"

	"github.com/jurgen-kluft/ccode/denv"
	"github.com/jurgen-kluft/ccode/embedded"
)

// Init will initialize ccode before anything else is run

var (
	// tundra, vs2022, make, cmake, xcode, espmake
	dev = "tundra"

	// win32, win64, linux32, linux64, macos64
	os = runtime.GOOS

	// x64, arm64, amd64, 386, esp32 / esp32c3 / esp32s3
	arch = runtime.GOARCH

	// verbose
	verbose = false
)

func Init() bool {

	flag.StringVar(&dev, "dev", "tundra", "the build system to generate for (vs2022, tundra, make, cmake, xcode, esp32)")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()

	// Currently supported: esp32, esp32s3
	if strings.HasPrefix(dev, "esp32") {
		os = "arduino"
		arch = "esp32"
		if strings.EqualFold(dev, "esp32s3") {
			arch = "esp32s3"
		}
	}

	if os == "" {
		os = strings.ToLower(runtime.GOOS)
	}

	if arch == "" {
		arch = strings.ToLower(runtime.GOARCH)
	}

	if dev == "" {
		dev = "tundra"
		if os == "windows" {
			dev = "vs2022"
		}
	}

	fmt.Println("ccode, a tool to generate C/C++ workspace and project files")

	if denv.NewDevEnum(dev) == denv.DevInvalid {
		fmt.Println()
		fmt.Println("Error, wrong parameter for '-dev', '", dev, "' is not recognized")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("    -> Usage: go run cbase.go -dev=vs2022/vs2019/vs2015")
		fmt.Println("    -> Usage: go run cbase.go -dev=tundra")
		fmt.Println("    -> Usage: go run cbase.go -dev=make")
		fmt.Println("    -> Usage: go run cbase.go -dev=xcode")
		fmt.Println("    -> Usage: go run cbase.go -dev=esp32")
		return false
	}

	// Initialize the build target that will be used during Package, Project and Lib creation
	denv.SetBuildTarget(os, arch)

	return true
}

// Generate is the main function that requires 'arguments' to then generate
// workspace and project files for a specified IDE.
func Generate(pkg *denv.Package) error {
	buildTarget := denv.GetBuildTarget()
	generator := denv.NewGenerator(dev, buildTarget, verbose)
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
