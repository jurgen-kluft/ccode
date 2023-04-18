package ccode

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/jurgen-kluft/ccode/cli"
	"github.com/jurgen-kluft/ccode/denv"
	"github.com/jurgen-kluft/ccode/embedded"
	"github.com/jurgen-kluft/ccode/tundra"
	"github.com/jurgen-kluft/ccode/vs"
)

// Init will initialize ccode before anything else is run
func Init() error {
	// Parse command-line
	app := cli.NewApp()
	app.Name = "ccode, a tool to generate C/C++ workspace and project files"
	app.Usage = "ccode --DEV=VS2022 --OS=Windows --ARCH=amd64"

	denv.DEV = ""
	denv.OS = runtime.GOOS
	denv.ARCH = runtime.GOARCH

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "DEV",
			Usage:       "The build system to generate projects for (VS2022, TUNDRA))",
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
			if denv.OS == "darwin" {
				denv.DEV = "TUNDRA"
			} else if denv.OS == "linux" {
				denv.DEV = "TUNDRA"
			} else {
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
		return vs.GenerateVisualStudioSolutionAndProjects(vs.GetVisualStudio(denv.DEV), pkg)
	} else if tundra.IsTundra(denv.DEV, denv.OS, denv.ARCH) {
		return tundra.GenerateTundraBuildFile(pkg)
	}
	return fmt.Errorf("Unknown DEV '%s'", denv.DEV)
}

func GenerateFiles() {
	embedded.WriteClangFormat(false)
	embedded.WriteGitIgnore(false)
	embedded.WriteTestMainCxx(true)
}
