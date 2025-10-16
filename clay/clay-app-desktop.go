package clay

import (
	"os"

	corepkg "github.com/jurgen-kluft/ccode/core"
	"github.com/jurgen-kluft/ccode/denv"
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

func ClayAppMainDesktop(pkg *denv.Package) {
	// Consume the first argument as the command
	command := os.Args[1]
	os.Args = os.Args[1:]

	app := NewApp(pkg)

	// Parse command line arguments
	var err error
	switch command {
	case "build":
		ParseProjectNameAndConfig(app)
		err = app.BuildDesktop()
	case "build-info":
		ParseProjectNameAndConfig(app)
		err = app.BuildInfoDesktop()
	case "clean":
		ParseProjectNameAndConfig(app)
		err = app.Clean()
	case "list-libraries":
		err = app.ListLibraries()
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

func (a *App) BuildDesktop() error {
	// Note: We should be running this from the "target/{build target}" directory
	// Create the build directory
	buildPath := a.GetBuildPath(GetBuildDirname(a.BuildConfig, a.BuildTarget))
	os.MkdirAll(buildPath+"/", os.ModePerm)

	prjs, err := a.CreateProjects(a.BuildTarget, a.BuildConfig)
	if err != nil {
		return err
	}

	for _, prj := range prjs {
		a.SetToolchain(prj, buildPath)
	}

	var outOfDate int

	// Build the libraries first
	for _, prj := range prjs {
		if !prj.IsExecutable() && prj.CanBuildFor(a.BuildConfig, a.BuildTarget) {
			if ood, err := prj.Build(a.BuildConfig, a.BuildTarget, buildPath); err != nil {
				return err
			} else {
				outOfDate += ood
			}
		}
	}

	// Now build the executables
	for _, prj := range prjs {
		if a.Config.ProjectName == "" || a.Config.ProjectName == prj.DevProject.Name {
			if prj.IsExecutable() && prj.CanBuildFor(a.BuildConfig, a.BuildTarget) {
				//AddBuildInfoAsCppLibrary(prj, a.BuildConfig)
				if ood, err := prj.Build(a.BuildConfig, a.BuildTarget, buildPath); err != nil {
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

func (a *App) BuildInfoDesktop() error {

	// TODO what should this do for just desktop applications?
	// Windows SDK version ?
	// Mac SDK version ?

	sdkVersion := ""
	buildConfig := a.BuildConfig
	buildTarget := a.BuildTarget
	projectName := a.Config.ProjectName

	buildPath := a.GetBuildPath(GetBuildDirname(buildConfig, buildTarget))
	prjs, err := a.CreateProjects(buildTarget, buildConfig)
	if err != nil {
		return err
	}

	for _, prj := range prjs {
		if projectName == "" || projectName == prj.DevProject.Name {
			if prj.CanBuildFor(buildConfig, buildTarget) {
				appPath, _ := os.Getwd()
				if err := GenerateBuildInfo(prj.GetBuildPath(buildPath), appPath, sdkVersion, BuildInfoFilenameWithoutExt); err != nil {
					return err
				}
			}
		}
	}
	corepkg.LogInfo("Ok, build info generated Ok")
	return nil
}
