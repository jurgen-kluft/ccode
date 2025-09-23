package clay

import (
	"os"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

// Clay App Desktop
//    <project>: name of a project (if you have more than one project)
//    <config>: debug, release (default), final
//
//    Commands:
//    - build -p <project> -c <config>
//    - build-info -p <project> -c <config>
//    - clean -p <project> -c <config>
//    - list-libraries

func ClayAppMainDesktop() {
	// Consume the first argument as the command
	command := os.Args[1]
	os.Args = os.Args[1:]

	// Parse command line arguments
	var err error
	switch command {
	case "build":
		err = BuildDesktop(ParseProjectNameAndConfig())
	case "build-info":
		err = BuildInfoDesktop(ParseProjectNameAndConfig())
	case "clean":
		err = Clean(ParseProjectNameAndConfig())
	case "list-libraries":
		err = ListLibraries()
	case "version":
		version := corepkg.NewVersionInfo()
		corepkg.LogInfof("Version: %s", version.Version)
	default:
		UsageDesktop()
	}

	if err != nil {
		corepkg.LogInfof("Error: %v", err)
		os.Exit(1)
	}
}

func UsageDesktop() {
	corepkg.LogInfo("Usage: clay [command] [options]")
	corepkg.LogInfo("Commands:")
	corepkg.LogInfo("  build-info -p <name> -build <config>")
	corepkg.LogInfo("  build -p <name> -build <config>")
	corepkg.LogInfo("  clean -p <name> -build <config>")
	corepkg.LogInfo("  list-libraries")
	corepkg.LogInfo("Options:")
	corepkg.LogInfo("  name       Project name (if more than one) ")
	corepkg.LogInfo("  config     Config name (debug, release, final, debug-dev, debug-test) ")
	corepkg.LogInfo("  --help     Show this help message")
	corepkg.LogInfo("  --version  Show version information")

	corepkg.LogInfo("Examples:")
	corepkg.LogInfo("  clay build-info (generates buildinfo.h and buildinfo.cpp for all projects and configs)")
	corepkg.LogInfo("  clay build-info -build debug  // generates buildinfo.h and buildinfo.cpp for debug-dev config")
	corepkg.LogInfo("  clay build                    // builds the project for the release-dev config")
	corepkg.LogInfo("  clay build -build debug       // builds the project for the debug-dev config")
	corepkg.LogInfo("  clay clean -build debug       // cleans the project for the debug-dev config")
	corepkg.LogInfo("  clay list-libraries")
}

func BuildDesktop(projectName string, buildConfig *Config) error {
	// Note: We should be running this from the "target/{build target}" directory
	// Create the build directory
	buildPath := GetBuildPath(buildConfig.GetSubDir())
	os.MkdirAll(buildPath+"/", os.ModePerm)

	prjs := ClayAppCreateProjectsFunc()
	for _, prj := range prjs {
		prj.SetToolchain(buildConfig, nil)
	}

	var outOfDate int

	// Build the libraries first
	for _, prj := range prjs {
		if !prj.IsExecutable && prj.Config.Matches(buildConfig) {
			if ood, err := prj.Build(buildConfig, buildPath); err != nil {
				return err
			} else {
				outOfDate += ood
			}
		}
	}

	// Now build the executables
	for _, prj := range prjs {
		if projectName == "" || projectName == prj.Name {
			if prj.IsExecutable && prj.Config.Matches(buildConfig) {
				AddBuildInfoAsCppLibrary(prj, buildConfig)
				if ood, err := prj.Build(buildConfig, buildPath); err != nil {
					return err
				} else {
					outOfDate += ood
				}
			}
		}
	}

	if outOfDate == 0 {
		corepkg.LogInfo("Nothing to build, everything is up to date...")
	}

	return nil
}

func BuildInfoDesktop(projectName string, buildConfig *Config) error {

	// TODO what should this do for just desktop applications?
	// Windows SDK version ?
	// Mac SDK version ?
	EspSdkPath := "/Users/obnosis5/sdk/arduino/esp32"
	buildPath := GetBuildPath(buildConfig.GetSubDir())

	prjs := ClayAppCreateProjectsFunc()
	for _, prj := range prjs {
		if projectName == "" || projectName == prj.Name {
			if prj.Config.Matches(buildConfig) {
				appPath, _ := os.Getwd()
				if err := GenerateBuildInfo(prj.GetBuildPath(buildPath), appPath, EspSdkPath, BuildInfoFilenameWithoutExt); err != nil {
					return err
				}
			}
		}
	}
	corepkg.LogInfo("Ok, build info generated Ok")
	return nil
}
