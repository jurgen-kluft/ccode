package xcode

import (
	"fmt"
	"os"

	"github.com/jurgen-kluft/xcode/cli"
	"github.com/jurgen-kluft/xcode/denv"
	"github.com/jurgen-kluft/xcode/items"
	"github.com/jurgen-kluft/xcode/tundra"
	"github.com/jurgen-kluft/xcode/vs"
)

// Generate is the main function that requires 'arguments' to then generate
// workspace and project files for a specified IDE.
func Generate(pkg *denv.Package) error {
	// Parse command-line
	app := cli.NewApp()
	app.Name = "xcode"
	app.Usage = "xcode --IDE=VS2015 --TARGET=Win64"

	var IDE string
	var targets string

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "IDE",
			Value:       "VS2015",
			Usage:       "IDE to generate projects for",
			Destination: &IDE,
		},
		cli.StringFlag{
			Name:        "TARGETS",
			Value:       "Win64",
			Usage:       "Targets to include (Win32, Win64, Darwin64)",
			Destination: &targets,
		},
	}
	app.Action = func(c *cli.Context) {
		err := generateProjects(IDE, targets, pkg)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return app.Run(os.Args)
}

func generateProjects(IDE string, targets string, pkg *denv.Package) error {
	if vs.IsVisualStudio(IDE, targets) {
		return vs.GenerateVisualStudioSolutionAndProjects(vs.GetVisualStudio(IDE), "", items.NewList(targets, ",", "").Items, pkg)
	} else if tundra.IsTundra(IDE, targets) {
		return tundra.GenerateTundraBuildFile(pkg)
	}
	return fmt.Errorf("Unknown IDE")
}
