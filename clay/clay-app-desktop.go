package clay

import (
	"os"

	"github.com/jurgen-kluft/ccode/foundation"
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
		version := foundation.NewVersionInfo()
		foundation.LogPrintf("Version: %s\n", version.Version)
	default:
		UsageDesktop()
	}

	if err != nil {
		foundation.LogPrintf("Error: %v\n", err)
		os.Exit(1)
	}
}

func UsageDesktop() {
	foundation.LogPrintln("Usage: clay [command] [options]")
	foundation.LogPrintln("Commands:")
	foundation.LogPrintln("  build-info -p <projectName> -c <projectConfig>")
	foundation.LogPrintln("  build -p <projectName> -c <projectConfig>")
	foundation.LogPrintln("  clean -p <projectName> -c <projectConfig>")
	foundation.LogPrintln("  list-libraries")
	foundation.LogPrintln("Options:")
	foundation.LogPrintln("  projectName       Project name (if more than one) ")
	foundation.LogPrintln("  projectConfig     Config name (debug, release, final, debug-dev, debug-test) ")
	foundation.LogPrintln("  --help            Show this help message")
	foundation.LogPrintln("  --version         Show version information")

	foundation.LogPrintln("Examples:")
	foundation.LogPrintln("  clay build-info (generates buildinfo.h and buildinfo.cpp for all projects and configs)")
	foundation.LogPrintln("  clay build-info -c debug  // generates buildinfo.h and buildinfo.cpp for debug-dev config")
	foundation.LogPrintln("  clay build                // builds the project for the release-dev config")
	foundation.LogPrintln("  clay build -c debug       // builds the project for the debug-dev config")
	foundation.LogPrintln("  clay clean -c debug       // cleans the project for the debug-dev config")
	foundation.LogPrintln("  clay list-libraries")
}

func BuildDesktop(projectName string, buildConfig *Config) error {
	// Note: We should be running this from the "target/{build target}" directory
	// Create the build directory
	buildPath := GetBuildPath(buildConfig.GetSubDir())
	os.MkdirAll(buildPath+"/", os.ModePerm)

	prjs := ClayAppCreateProjectsFunc(buildConfig.Target.ArchAsString())
	for _, prj := range prjs {
		prj.SetToolchain(buildConfig)
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
		foundation.LogPrintln("Nothing to build, everything is up to date...")
	}

	return nil
}

func BuildInfoDesktop(projectName string, buildConfig *Config) error {

	// TODO what should this do for just desktop applications?
	// Windows SDK version ?
	// Mac SDK version ?
	EspSdkPath := "/Users/obnosis5/sdk/arduino/esp32"
	buildPath := GetBuildPath(buildConfig.GetSubDir())

	prjs := ClayAppCreateProjectsFunc(buildConfig.Target.ArchAsString())
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
	foundation.LogPrintln("Ok, build info generated Ok")
	return nil
}
