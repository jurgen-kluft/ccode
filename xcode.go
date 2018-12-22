package xcode

import (
	"fmt"
	"os"
	"runtime"

	"github.com/jurgen-kluft/xcode/cli"
	"github.com/jurgen-kluft/xcode/denv"
	"github.com/jurgen-kluft/xcode/items"
	"github.com/jurgen-kluft/xcode/tundra"
	"github.com/jurgen-kluft/xcode/vs"
)

// Init will initialize xcode before anything else is run
func Init() error {
	// Parse command-line
	app := cli.NewApp()
	app.Name = "xcode"
	app.Usage = "xcode --DEV=VS2017 --OS=Windows --ARCH=amd64"

	DEV := ""
	OS := runtime.GOOS
	ARCH := runtime.GOARCH

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "DEV",
			Usage:       "The build system to generate projects for",
			Destination: &DEV,
		},
		cli.StringFlag{
			Name:        "OS",
			Usage:       "OS to include (windows, darwin)",
			Destination: &OS,
		},
		cli.StringFlag{
			Name:        "ARCH",
			Usage:       "Architecture to include (386, amd64)",
			Destination: &ARCH,
		},
	}
	app.Action = func(c *cli.Context) {
		if OS == "" {
			OS = runtime.GOOS
		}
		if ARCH == "" {
			ARCH = runtime.GOARCH
		}
		if DEV == "" {
			if OS == "darwin" {
				DEV = "TUNDRA"
			} else {
				DEV = "VS2017"
			}
		}
		fmt.Printf("xcode (DEV:%s, OS:%s, ARCH:%s)\n", DEV, OS, ARCH)
		denv.Init(DEV, OS, ARCH)
	}
	return app.Run(os.Args)
}

// Generate is the main function that requires 'arguments' to then generate
// workspace and project files for a specified IDE.
func Generate(pkg *denv.Package) error {
	if vs.IsVisualStudio(denv.XCodeDEV, denv.XCodeOS, denv.XCodeARCH) {
		return vs.GenerateVisualStudioSolutionAndProjects(vs.GetVisualStudio(denv.XCodeDEV), "", items.NewList(denv.XCodeOS, ",", "").Items, pkg)
	} else if tundra.IsTundra(denv.XCodeDEV, denv.XCodeOS, denv.XCodeARCH) {
		return tundra.GenerateTundraBuildFile(pkg)
	}
	return fmt.Errorf("Unknown DEV '%s'", denv.XCodeDEV)
}
