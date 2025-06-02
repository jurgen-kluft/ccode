package clay

import (
	"fmt"
	"log"
	"os"
	"time"

	utils "github.com/jurgen-kluft/ccode/utils"
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
		version := utils.NewVersionInfo()
		fmt.Printf("Version: %s\n", version.Version)
	default:
		UsageDesktop()
	}

	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func UsageDesktop() {
	fmt.Println("Usage: clay [command] [options]")
	fmt.Println("Commands:")
	fmt.Println("  build-info -p <projectName> -c <projectConfig>")
	fmt.Println("  build -p <projectName> -c <projectConfig>")
	fmt.Println("  clean -p <projectName> -c <projectConfig>")
	fmt.Println("  list-libraries")
	fmt.Println("Options:")
	fmt.Println("  projectName       Project name (if more than one) ")
	fmt.Println("  projectConfig     Config name (debug, release, final, debug-dev, debug-test) ")
	fmt.Println("  --help            Show this help message")
	fmt.Println("  --version         Show version information")

	fmt.Println("Examples:")
	fmt.Println("  clay build-info (generates buildinfo.h and buildinfo.cpp for all projects and configs)")
	fmt.Println("  clay build-info -c debug  // generates buildinfo.h and buildinfo.cpp for debug-dev config")
	fmt.Println("  clay build                // builds the project for the release-dev config")
	fmt.Println("  clay build -c debug       // builds the project for the debug-dev config")
	fmt.Println("  clay clean -c debug       // cleans the project for the debug-dev config")
	fmt.Println("  clay list-libraries")
}

func BuildDesktop(projectName string, buildConfig *Config) (err error) {
	// Note: We should be running this from the "target/{build target}" directory
	// Create the build directory
	buildPath := GetBuildPath(buildConfig.GetSubDir())
	os.MkdirAll(buildPath+"/", os.ModePerm)

	prjs := ClayAppCreateProjectsFunc(buildConfig.Target.ArchAsString())
	for _, prj := range prjs {
		prj.SetToolchain(buildConfig)
	}

	for _, prj := range prjs {
		if projectName == "" || projectName == prj.Name {
			if prj.Config.Matches(buildConfig) {
				log.Printf("Building project: %s, config: %s, arch: %s\n", prj.Name, prj.Config.String(), prj.Config.Target.ArchAsString())
				startTime := time.Now()
				{
					AddBuildInfoAsCppLibrary(prj, buildConfig)

					var state int
					if state, err = prj.Build(buildConfig, buildPath); err != nil {
						return err
					}
					if state == 0 { // 0 means project is up to date
						log.Printf("Building done ... %s is up to date (skipped build)\n", prj.Name)
					} else if state == 1 { // 1 means project was built successfully
						log.Printf("Building done ... (duration %s)\n", time.Since(startTime).Round(time.Second))
					}
					fmt.Println()
				}
			}
		}
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
	log.Println("Ok, build info generated Ok")
	return nil
}
