package clay

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jurgen-kluft/ccode/foundation"
)

const (
	BuildInfoFilenameWithoutExt = "buildinfo"
)

var ClayAppCreateProjectsFunc func() []*Project

func GetBuildPath(subdir string) string {
	buildPath := filepath.Join("build", subdir)
	return buildPath
}

func GetDefaultOs() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	default:
		return runtime.GOOS
	}
}

func GetDefaultArch() string {
	switch runtime.GOARCH {
	case "386":
		return "x86"
	case "amd64":
		return "x64"
	case "arm64":
		return "arm64"
	case "arm":
		return "arm32"
	case "riscv64":
		return "riscv64"
	default:
		return runtime.GOARCH
	}
}

func ParseProjectNameAndConfig() (string, *Config) {
	var projectName string
	var targetOs string
	var targetArch string
	var targetBuild string
	flag.StringVar(&projectName, "p", "", "Name of the project")
	flag.StringVar(&targetOs, "os", GetDefaultOs(), "Target OS (windows, darwin, linux, arduino)")
	flag.StringVar(&targetBuild, "build", "release", "Format 'build' or 'build-variant', e.g. debug, debug-dev, release-dev, debug-dev-test)")
	flag.StringVar(&targetArch, "arch", GetDefaultArch(), "Cpu Architecture (amd64, x64, arm64, esp32, esp32s3)")
	flag.Parse()

	if runtime.GOOS == "windows" {
		targetOs = "windows"
	} else if runtime.GOOS == "darwin" {
		targetOs = "darwin"
	} else {
		targetOs = "linux"
	}

	if strings.HasPrefix(targetArch, "esp32") {
		targetOs = "arduino"
	}

	if targetArch == "" {
		targetArch = runtime.GOARCH
		if targetOs == "arduino" {
			targetArch = "esp32"
		} else if targetOs == "darwin" {
			targetArch = "arm64"
		} else if targetOs == "windows" {
			targetArch = "x64"
		} else if targetOs == "linux" {
			targetArch = "amd64"
		}
	}

	config := NewConfig(targetOs, targetArch, targetBuild)
	return projectName, config
}

func Build(projectName string, targetConfig *Config) (err error) {
	// Note: We should be running this from the "target/{build target}" directory
	// Create the build directory
	buildPath := GetBuildPath(targetConfig.GetSubDir())
	os.MkdirAll(buildPath+"/", os.ModePerm)

	prjs := ClayAppCreateProjectsFunc()
	for _, prj := range prjs {
		prj.SetToolchain(targetConfig)
	}

	var outOfDate int
	for _, prj := range prjs {
		if projectName == "" || projectName == prj.Name {
			if prj.Config.Matches(targetConfig) {
				if prj.IsExecutable {
					AddBuildInfoAsCppLibrary(prj, targetConfig)
				}
				if outOfDate, err = prj.Build(targetConfig, buildPath); err != nil {
					return err
				}
			}
		}
	}
	if outOfDate == 0 {
		foundation.LogPrintln("Nothing to build, everything is up to date")
	}
	return err
}

func Clean(projectName string, buildConfig *Config) error {
	prjs := ClayAppCreateProjectsFunc()
	for _, prj := range prjs {
		if projectName == "" || projectName == prj.Name {
			if prj.Config.Matches(buildConfig) {

				buildPath := prj.GetBuildPath(buildConfig.GetSubDir())

				// Note: We should be running this from the "target/esp" directory
				// Remove all folders and files from "build/"
				if err := os.RemoveAll(buildPath + "/"); err != nil {
					return foundation.LogError(err, "Failed to remove build directory")
				}

				if err := os.MkdirAll(buildPath+"/", os.ModePerm); err != nil {
					return foundation.LogError(err, "Failed to create build directory")
				}
			}
		}
	}

	return nil
}

func ListLibraries() error {
	prjs := ClayAppCreateProjectsFunc()

	configs := make([]string, 0, 16)
	nameToIndex := make(map[string]int)
	for _, prj := range prjs {
		if idx, ok := nameToIndex[prj.Name]; !ok {
			idx = len(configs)
			nameToIndex[prj.Name] = idx
			configs = append(configs, prj.Config.Config.AsString())
		} else {
			configs[idx] += ", " + prj.Config.Config.AsString()
		}
	}

	for _, prj := range prjs {
		if i, ok := nameToIndex[prj.Name]; ok {
			foundation.LogPrintf("Project: %s\n", prj.Name)
			foundation.LogPrintf("  Configs: %s\n", configs[i])
			if len(prj.Dependencies) > 0 {
				foundation.LogPrint("  Libraries:\n")
				for _, dep := range prj.Dependencies {
					foundation.LogPrintf("  - %s\n", dep.Name)
				}
			}
			foundation.LogPrintln()

			// Remove the entry from the map to avoid duplicates
			delete(nameToIndex, prj.Name)
		}
	}

	return nil
}

// AddBuildInfoAsCppLibrary checks if 'buildinfo.h' and 'buildinfo.cpp' exist,
// if so it creates a C++ library and adds it to the project
func AddBuildInfoAsCppLibrary(prj *Project, cfg *Config) {
	name := BuildInfoFilenameWithoutExt
	buildPath := prj.GetBuildPath(cfg.GetSubDir())
	hdrFilepath := filepath.Join(prj.GetBuildPath(buildPath), name+".h")
	srcFilepath := filepath.Join(prj.GetBuildPath(buildPath), name+".cpp")
	if foundation.FileExists(hdrFilepath) && foundation.FileExists(srcFilepath) {
		library := NewLibraryProject(name, prj.Config)

		library.Defines = NewDefineMap(1)
		library.IncludeDirs = NewIncludeMap(1)
		library.SourceFiles = make([]SourceFile, 0, 1)
		library.Dependencies = make([]*Project, 0, 1)

		library.IncludeDirs.Add(filepath.Dir(hdrFilepath))
		library.AddSourceFile(srcFilepath, filepath.Base(srcFilepath))
		prj.AddLibrary(library)
	}
}
