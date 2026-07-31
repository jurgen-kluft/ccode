package clay

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/gide/denv"
	"github.com/jurgen-kluft/gide/msdev"
)

type GenerateOptions struct {
	Dev            string
	Arch           string
	Build          string
	StartupProject string
	OutputDir      string
}

func ParseGenerateOptions(args []string) (GenerateOptions, error) {
	options := GenerateOptions{}
	flags := flag.NewFlagSet("generate", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.Dev, "dev", "", "IDE or build system to generate")
	flags.StringVar(&options.Arch, "arch", "x64", "Target architecture")
	flags.StringVar(&options.Build, "build", "", "Build configuration filter")
	flags.StringVar(&options.StartupProject, "p", "", "Startup project")
	flags.StringVar(&options.OutputDir, "output", "", "Generation output directory")
	if err := flags.Parse(args); err != nil {
		return GenerateOptions{}, err
	}
	options.Dev = strings.ToLower(options.Dev)
	options.Arch = strings.ToLower(options.Arch)
	options.Build = strings.ToLower(options.Build)
	if options.Dev == "" {
		return GenerateOptions{}, fmt.Errorf("generate requires --dev")
	}
	if options.Dev != "vs2022" {
		return GenerateOptions{}, fmt.Errorf("unsupported generator %q", options.Dev)
	}
	switch options.Arch {
	case "x86", "x64", "arm32", "arm64":
	default:
		return GenerateOptions{}, fmt.Errorf("unsupported Visual Studio architecture %q", options.Arch)
	}
	return options, nil
}

func (a *App) Generate(args []string) error {
	options, err := ParseGenerateOptions(args)
	if err != nil {
		return err
	}

	buildTarget := denv.GetBuildTargetFromOsArch("windows", options.Arch)

	var buildConfigs []denv.BuildConfig
	if options.Build != "" {
		buildConfig := denv.BuildConfigFromString(options.Build)
		if buildConfig.Build == denv.BuildNone {
			return fmt.Errorf("unsupported build configuration %q", options.Build)
		}
		buildConfigs = []denv.BuildConfig{buildConfig}
	}

	outputDir := options.OutputDir
	if outputDir == "" {
		outputDir = filepath.Join(a.Pkg.Path(), a.Pkg.RepoName, "target", options.Dev)
	}

	return msdev.GeneratePackage(a.Pkg, msdev.WorkspaceOptions{
		OutputDir:      outputDir,
		BuildTarget:    buildTarget,
		BuildConfigs:   buildConfigs,
		StartupProject: options.StartupProject,
	})
}
