package xcode

import (
	"fmt"
	"github.com/jurgen-kluft/xcode/cli"
	"github.com/jurgen-kluft/xcode/ide"
	"github.com/jurgen-kluft/xcode/visual_studio"
	"os"
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

// Version (based on semver)
type Version struct {
	Major uint32
	Minor uint32
	Patch uint32
}

// Package defines information of a C++ project
type Package struct {
	name     string
	guid     string
	author   string
	version  Version
	language string   // C++, C#
	targets  []string // Windows, Darwin
	configs  []Config
}

func Generate(version ide.Type, pkg Package) error {
	// Parse command-line
	app := cli.NewApp()
	app.Name = "xcode"
	app.Usage = "xcode --IDE=VS2015 --TARGET=Win64"
	app.Action = func(c *cli.Context) {
		println("boom! I say!")
	}

	var ide string
	var targets string
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "IDE",
			Value:       "VS2015",
			Usage:       "IDE to generate projects for",
			Destination: &ide,
		},
		cli.StringFlag{
			Name:        "TARGETS",
			Value:       "Win64",
			Usage:       "Targets to include (Win32, Win64, Darwin64)",
			Destination: &targets,
		},
	}
	app.Action = func(c *cli.Context) {
		generateProjects(ide, targets, pkg)
	}

	app.Run(os.Args)
}

func generateProjects(IDE string, targets string, pkg Package) error {
	if vs.IsVisualStudio(IDE) {
		return vs.Generate(vs.GetVisualStudio(IDE), "", targets, pkg)
	}
	return fmt.Errorf("Wrong visual studio version")
}
