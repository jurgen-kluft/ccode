package xcode

import (
	"fmt"
	"github.com/jurgen-kluft/xcode/cli"
	"github.com/jurgen-kluft/xcode/denv"
	"github.com/jurgen-kluft/xcode/vs"
	"os"
	"strings"
)

type Config struct {
	name      string // Debug, Release
	defines   string //
	includes  []string
	libraries []string
	linking   []string
}

var DefaultConfigs = []Config{
	{name: "DevDebug", defines: "TARGET_DEV_DEBUG;_DEBUG;", libraries: []string{""}, linking: []string{""}},
	{name: "DevRelease", defines: "TARGET_DEV_RELEASE;NDEBUG;"},
	{name: "DevFinal", defines: "TARGET_DEV_FINAL;NDEBUG;"},
	{name: "TestDebug", defines: "TARGET_TEST_DEBUG;_DEBUG;"},
	{name: "TestRelease", defines: "TARGET_TEST_RELEASE;NDEBUG;"},
}

func Generate(project denv.Project) error {
	// Parse command-line
	app := cli.NewApp()
	app.Name = "xcode"
	app.Usage = "xcode --IDE=VS2015 --TARGET=Win64"
	app.Action = func(c *cli.Context) {
		println("boom! I say!")
	}

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
		generateProjects(IDE, targets, project)
	}

	return app.Run(os.Args)
}

func ListToArray(list string, sep string) []string {
	return strings.Split(list, sep)
}

func generateProjects(IDE string, targets string, project denv.Project) error {
	if vs.IsVisualStudio(IDE) {
		return vs.Generate(vs.GetVisualStudio(IDE), "", ListToArray(targets, ","), project)
	}
	return fmt.Errorf("Wrong visual studio version")
}
