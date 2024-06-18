package ccode

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/jurgen-kluft/ccode/cli"
	"github.com/jurgen-kluft/ccode/cmake"
	"github.com/jurgen-kluft/ccode/denv"
	"github.com/jurgen-kluft/ccode/embedded"
	"github.com/jurgen-kluft/ccode/ide"
	"github.com/jurgen-kluft/ccode/tundra"
	"github.com/jurgen-kluft/ccode/vs"
)

// Init will initialize ccode before anything else is run
func Init() error {
	// Parse command-line
	app := cli.NewApp()
	app.Name = "ccode, a tool to generate C/C++ workspace and project files"
	app.Usage = "ccode --DEV=VS2022 --OS=windows --ARCH=amd64"

	denv.DEV = ""
	denv.OS = runtime.GOOS
	denv.ARCH = runtime.GOARCH

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "DEV",
			Usage:       "The build system to generate projects for (VS2022, TUNDRA, CMAKE))",
			Destination: &denv.DEV,
		},
		cli.StringFlag{
			Name:        "OS",
			Usage:       "OS to include (windows, darwin, linux)",
			Destination: &denv.OS,
		},
		cli.StringFlag{
			Name:        "ARCH",
			Usage:       "Architecture to include (386, amd64)",
			Destination: &denv.ARCH,
		},
	}
	app.Action = func(c *cli.Context) {
		if denv.OS == "" {
			denv.OS = strings.ToLower(runtime.GOOS)
		}
		if denv.ARCH == "" {
			denv.ARCH = strings.ToLower(runtime.GOARCH)
		}
		if denv.DEV == "" {
			denv.DEV = "TUNDRA"
			if denv.OS == "windows" {
				denv.DEV = "VS2022"
			}
		}
		fmt.Printf("CCode (DEV:%s, OS:%s, ARCH:%s)\n", denv.DEV, denv.OS, denv.ARCH)
	}
	return app.Run(os.Args)
}

// Generate is the main function that requires 'arguments' to then generate
// workspace and project files for a specified IDE.
func Generate(pkg *denv.Package) error {
	if vs.IsVisualStudio(denv.DEV, denv.OS, denv.ARCH) {
		//return vs.GenerateBuildFiles(vs.GetVisualStudio(denv.DEV), pkg)
		generator := ide.NewAxeGenerator()
		return generator.GenerateMsDev(vs.GetVisualStudio(denv.DEV), pkg)
	} else if tundra.IsTundra(denv.DEV, denv.OS, denv.ARCH) {
		return tundra.GenerateBuildFiles(pkg)
	} else if cmake.IsCMake(denv.DEV, denv.OS, denv.ARCH) {
		return cmake.GenerateBuildFiles(pkg)
	}

	return fmt.Errorf("Unknown DEV '%s'", denv.DEV)
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
