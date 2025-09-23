package clay

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain"
	corepkg "github.com/jurgen-kluft/ccode/core"
)

const (
	BuildInfoFilenameWithoutExt = "buildinfo"
)

var ClayAppCreateProjectsFunc func() []*Project

type ClayConfig struct {
	ProjectName string `json:"project,omitempty"`
	TargetOs    string `json:"os,omitempty"`
	TargetArch  string `json:"arch,omitempty"`
	TargetBuild string `json:"build,omitempty"`
	TargetBoard string `json:"board,omitempty"`
}

func (c ClayConfig) Equal(other ClayConfig) bool {
	return c.ProjectName == other.ProjectName &&
		c.TargetOs == other.TargetOs &&
		c.TargetArch == other.TargetArch &&
		c.TargetBuild == other.TargetBuild &&
		c.TargetBoard == other.TargetBoard
}

var clayConfig ClayConfig

func GetBuildPath(subdir string) string {
	var buildPath string
	if len(clayConfig.TargetBoard) > 0 {
		buildPath = filepath.Join("build", subdir, clayConfig.TargetBoard)
	} else {
		buildPath = filepath.Join("build", subdir)
	}
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
	flag.StringVar(&clayConfig.ProjectName, "p", "", "Name of the project")
	flag.StringVar(&clayConfig.TargetOs, "os", "", "Target OS (windows, darwin, linux, arduino)")
	flag.StringVar(&clayConfig.TargetBuild, "build", "", "Format 'build' or 'build-variant', e.g. debug, debug-dev, release-dev, debug-dev-test)")
	flag.StringVar(&clayConfig.TargetArch, "arch", "", "Cpu Architecture (amd64, x64, arm64, esp32)")
	flag.StringVar(&clayConfig.TargetBoard, "board", "", "Board name (s3, c3, xiao-c3, ...)")
	flag.Parse()

	clayConfig.TargetArch = strings.ToLower(clayConfig.TargetArch)
	clayConfig.TargetOs = strings.ToLower(clayConfig.TargetOs)
	clayConfig.TargetBuild = strings.ToLower(clayConfig.TargetBuild)
	clayConfig.TargetBoard = strings.ToLower(clayConfig.TargetBoard)

	configFilePath := "clay.json"
	loadedConfig := ClayConfig{}
	{
		if corepkg.FileExists(configFilePath) {
			data, err := os.ReadFile(configFilePath)
			if err != nil {
				corepkg.LogFatalf("Failed to read config file", configFilePath, err)
			}
			err = json.Unmarshal(data, &loadedConfig)
			if err != nil {
				corepkg.LogFatalf("Failed to parse config file", configFilePath, err)
			}
		}
	}

	if len(clayConfig.ProjectName) == 0 {
		clayConfig.ProjectName = loadedConfig.ProjectName
	}

	if len(clayConfig.TargetBoard) > 0 {
		clayConfig.TargetOs = "arduino"
		clayConfig.TargetArch = "esp32"
	}

	if len(clayConfig.TargetOs) == 0 {
		clayConfig.TargetOs = loadedConfig.TargetOs
	} else {
		// If the target OS was specified on the command line, clear the board and arch
		switch clayConfig.TargetOs {
		case "native":
			clayConfig.TargetBoard = ""
			clayConfig.TargetArch = GetDefaultArch()
			clayConfig.TargetOs = GetDefaultOs()
		case "windows":
			clayConfig.TargetBoard = ""
			clayConfig.TargetArch = "x64"
		case "darwin":
			clayConfig.TargetBoard = ""
			clayConfig.TargetArch = "arm64"
		case "linux":
			clayConfig.TargetBoard = ""
			clayConfig.TargetArch = "amd64"
		case "arduino":
			if len(clayConfig.TargetArch) == 0 {
				clayConfig.TargetArch = "esp32"
			}
		}
	}
	if len(clayConfig.TargetOs) == 0 {
		switch runtime.GOOS {
		case "windows":
			clayConfig.TargetOs = "windows"
		case "darwin":
			clayConfig.TargetOs = "darwin"
		default:
			clayConfig.TargetOs = "linux"
		}
	}

	if clayConfig.TargetOs == "arduino" {
		if len(clayConfig.TargetBoard) == 0 {
			clayConfig.TargetBoard = loadedConfig.TargetBoard
		}
	}

	if len(clayConfig.TargetBuild) == 0 {
		clayConfig.TargetBuild = loadedConfig.TargetBuild
	}
	if len(clayConfig.TargetBuild) == 0 {
		clayConfig.TargetBuild = "dev-release"
	}

	if len(clayConfig.TargetArch) == 0 {
		clayConfig.TargetArch = loadedConfig.TargetArch
	}
	if len(clayConfig.TargetArch) == 0 {
		clayConfig.TargetArch = runtime.GOARCH
		switch clayConfig.TargetOs {
		case "arduino":
			clayConfig.TargetArch = "esp32"
		case "darwin":
			clayConfig.TargetArch = "arm64"
		case "windows":
			clayConfig.TargetArch = "x64"
		case "linux":
			clayConfig.TargetArch = "amd64"
		}
	}

	// If any of the config values were updated from the command line flags, write back the config file
	// Compare the loaded and current config, when they are different we need to update the file
	if clayConfig.Equal(loadedConfig) == false {
		jsonContent, err := json.MarshalIndent(clayConfig, "", "  ")
		if err != nil {
			corepkg.LogFatalf("Failed to marshal config file", configFilePath, err)
		}
		if err := os.WriteFile(configFilePath, jsonContent, 0644); err != nil {
			corepkg.LogFatalf("Failed to write config file", configFilePath, err)
		}
	}

	corepkg.LogInfof("Project: %s", clayConfig.ProjectName)
	corepkg.LogInfof("Os: %s", clayConfig.TargetOs)
	corepkg.LogInfof("Arch: %s", clayConfig.TargetArch)
	corepkg.LogInfof("Build: %s", clayConfig.TargetBuild)
	if len(clayConfig.TargetBoard) > 0 {
		corepkg.LogInfof("Board: %s", clayConfig.TargetBoard)
	}

	return clayConfig.ProjectName, NewConfig(clayConfig.TargetOs, clayConfig.TargetArch, clayConfig.TargetBuild)
}

func Build(projectName string, targetConfig *Config) (err error) {
	// Note: We should be running this from the "target/{build target}" directory
	// Create the build directory
	buildPath := GetBuildPath(targetConfig.GetSubDir())
	os.MkdirAll(buildPath+"/", os.ModePerm)

	// Note: Here we are loading 'boards.txt', which is generated by 'clay generate-boards'
	var board *toolchain.Esp32Board
	if clayConfig.TargetOs == "arduino" && len(clayConfig.TargetBoard) > 0 {
		esp32Toolchain := NewEsp32Toolchain()
		err := LoadBoards(esp32Toolchain)
		if err != nil {
			return err
		}
		board = esp32Toolchain.GetBoardByName(clayConfig.TargetBoard)
		if board == nil {
			return fmt.Errorf("Board not found: %s", clayConfig.TargetBoard)
		}
	}

	prjs := ClayAppCreateProjectsFunc()
	for _, prj := range prjs {
		prj.SetToolchain(targetConfig, board)
	}

	var noMatchConfigs int
	var outOfDate int
	for _, prj := range prjs {
		if prj.Config.Matches(targetConfig) {
			if prj.IsExecutable {
				AddBuildInfoAsCppLibrary(prj, targetConfig)
			}
			if outOfDate, err = prj.Build(targetConfig, buildPath); err != nil {
				return err
			}
		} else {
			noMatchConfigs++
		}
	}
	if outOfDate == 0 && noMatchConfigs < len(prjs) {
		corepkg.LogInfo("Nothing to build, everything is up to date")
	} else if noMatchConfigs >= len(prjs) {
		corepkg.LogError(fmt.Errorf("!"), "No matching project configurations found")
	}
	return err
}

func Clean(projectName string, buildConfig *Config) error {
	buildPath := GetBuildPath(buildConfig.GetSubDir())

	prjs := ClayAppCreateProjectsFunc()
	for _, prj := range prjs {
		fmt.Println(prj.Name, prj.Config.ConfigString())
		if prj.Config.Matches(buildConfig) {

			projectBuildPath := prj.GetBuildPath(buildPath)
			corepkg.LogInfo("Clean " + projectBuildPath)

			if err := os.RemoveAll(projectBuildPath + "/"); err != nil {
				return corepkg.LogError(err, "Failed to remove build directory")
			}

			if err := os.MkdirAll(projectBuildPath+"/", os.ModePerm); err != nil {
				return corepkg.LogError(err, "Failed to create build directory")
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
			corepkg.LogInfo("Project: %s", prj.Name)
			corepkg.LogInfo("  Configs: %s", configs[i])
			if len(prj.Dependencies) > 0 {
				corepkg.LogInfo("  Libraries:")
				for _, dep := range prj.Dependencies {
					corepkg.LogInfo("  - %s", dep.Name)
				}
			}
			corepkg.LogInfo()

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
	if corepkg.FileExists(hdrFilepath) && corepkg.FileExists(srcFilepath) {
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
