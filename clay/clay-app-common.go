package clay

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
	"github.com/jurgen-kluft/ccode/denv"
)

const (
	BuildInfoFilenameWithoutExt = "buildinfo"
)

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

func GetBuildDirname(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) string {
	return buildTarget.Os().String() + "-" + buildTarget.Arch().String() + "-" + buildConfig.String()
}

func GetBuildPath(subdir string) string {
	var buildPath string
	if len(clayConfig.TargetBoard) > 0 {
		buildPath = filepath.Join("build", subdir, clayConfig.TargetBoard)
	} else {
		buildPath = filepath.Join("build", subdir)
	}
	return buildPath
}

func GetNativeOs() string {
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

func GetNativeArch() string {
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

func BuildTargetFromString(s string) denv.BuildTarget {
	return denv.BuildTargetFromString(s)
}

func ParseProjectNameAndConfig() (string, denv.BuildConfig, denv.BuildTarget) {
	flag.StringVar(&clayConfig.ProjectName, "p", "", "Name of the project")
	flag.StringVar(&clayConfig.TargetOs, "os", "", "Target OS (windows, darwin, linux, arduino)")
	flag.StringVar(&clayConfig.TargetBuild, "build", "", "Format 'build' or 'build-variant', e.g. debug, debug-dev, release-dev, debug-dev-test)")
	flag.StringVar(&clayConfig.TargetArch, "arch", "", "Cpu Architecture (amd64, x64, arm64, esp32, esp8266)")
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
		if len(clayConfig.TargetArch) != 0 && clayConfig.TargetArch != "esp32" && clayConfig.TargetArch != "esp8266" {
			clayConfig.TargetArch = "esp32"
		}
	}

	if len(clayConfig.TargetOs) == 0 {
		clayConfig.TargetOs = loadedConfig.TargetOs
	} else {
		// If the target OS was specified on the command line, clear the board and arch
		switch clayConfig.TargetOs {
		case "native":
			clayConfig.TargetBoard = ""
			clayConfig.TargetArch = GetNativeArch()
			clayConfig.TargetOs = GetNativeOs()
		case "windows":
			clayConfig.TargetBoard = ""
			clayConfig.TargetArch = "x64"
		case "darwin":
			clayConfig.TargetBoard = ""
			clayConfig.TargetArch = GetNativeArch()
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

	buildConfig := denv.BuildConfigFromString(clayConfig.TargetBuild)
	buildTargetStr := fmt.Sprintf("%s(%s)", clayConfig.TargetOs, clayConfig.TargetArch)
	buildTarget := denv.BuildTargetFromString(buildTargetStr)
	return clayConfig.ProjectName, buildConfig, buildTarget
}

func Build(projectName string, buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) (err error) {
	// Note: We should be running this from the "target/{build target}" directory
	// Create the build directory
	buildPath := GetBuildPath(GetBuildDirname(buildConfig, buildTarget))
	os.MkdirAll(buildPath+"/", os.ModePerm)

	prjs, err := CreateProjects(buildTarget, buildConfig)
	if err != nil {
		return err
	}

	for _, prj := range prjs {
		prj.SetToolchain(buildConfig, buildTarget, buildPath)
	}

	var noMatchConfigs int
	var outOfDate int
	for _, prj := range prjs {
		if prj.CanBuildFor(buildConfig, buildTarget) {
			// if prj.IsExecutable {
			// 	AddBuildInfoAsCppLibrary(prj, buildConfig)
			// }
			if outOfDate, err = prj.Build(buildConfig, buildTarget, buildPath); err != nil {
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

func Clean(projectName string, buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) error {
	buildPath := GetBuildPath(GetBuildDirname(buildConfig, buildTarget))

	prjs, err := CreateProjects(buildTarget, buildConfig)
	if err != nil {
		return err
	}

	for _, prj := range prjs {
		for _, cfg := range prj.Config {
			if cfg.BuildConfig.IsEqual(buildConfig) {
				fmt.Println(prj.DevProject.Name, buildConfig.String())

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
	}

	return nil
}

func ListLibraries(buildTarget denv.BuildTarget) error {
	buildConfig := denv.NewDebugDevConfig()
	prjs, err := CreateProjects(buildTarget, buildConfig)
	if err != nil {
		return err
	}

	configs := make([]string, 0, 16)
	nameToIndex := make(map[string]int)
	for _, prj := range prjs {
		if idx, ok := nameToIndex[prj.DevProject.Name]; !ok {
			idx = len(configs)
			nameToIndex[prj.DevProject.Name] = idx
			configs = append(configs, buildConfig.String())
		} else {
			configs[idx] += ", " + buildConfig.String()
		}
	}

	for _, prj := range prjs {
		if i, ok := nameToIndex[prj.DevProject.Name]; ok {
			corepkg.LogInfo("Project: %s", prj.DevProject.Name)
			corepkg.LogInfo("  Configs: %s", configs[i])
			if len(prj.Dependencies) > 0 {
				corepkg.LogInfo("  Libraries:")
				for _, dep := range prj.Dependencies {
					corepkg.LogInfo("  - %s", dep.DevProject.Name)
				}
			}
			corepkg.LogInfo()

			// Remove the entry from the map to avoid duplicates
			delete(nameToIndex, prj.DevProject.Name)
		}
	}

	return nil
}

// CreateProjects creates projects using the package.json file, it will create
// projects that match the build target and build configuration.
func CreateProjects(buildTarget denv.BuildTarget, buildConfig denv.BuildConfig) ([]*Project, error) {
	projects := make([]*Project, 0, 4)

	pkg, err := denv.LoadPackageFromJson("../package.json")
	if err != nil {
		return nil, err
	}

	// First create the projects without dependencies
	projectNameToIndex := make(map[string]int)

	for _, devPrj := range pkg.GetMainLib() {
		if !devPrj.HasMatchingConfigForTarget(buildConfig, buildTarget) {
			continue
		}
		project := NewProjectFromDevProject(devPrj, devPrj.Configs)
		projectNameToIndex[devPrj.Name] = len(projects)
		projects = append(projects, project)
	}

	for _, devPrj := range pkg.GetMainApp() {
		if !devPrj.HasMatchingConfigForTarget(buildConfig, buildTarget) {
			continue
		}
		project := NewProjectFromDevProject(devPrj, devPrj.Configs)
		projectNameToIndex[devPrj.Name] = len(projects)
		projects = append(projects, project)
	}

	for _, devPrj := range pkg.GetTestLib() {
		if !devPrj.HasMatchingConfigForTarget(buildConfig, buildTarget) {
			continue
		}
		project := NewProjectFromDevProject(devPrj, devPrj.Configs)
		projectNameToIndex[devPrj.Name] = len(projects)
		projects = append(projects, project)
	}

	for _, devPrj := range pkg.GetUnittest() {
		if !devPrj.HasMatchingConfigForTarget(buildConfig, buildTarget) {
			continue
		}
		project := NewProjectFromDevProject(devPrj, devPrj.Configs)
		projectNameToIndex[devPrj.Name] = len(projects)
		projects = append(projects, project)
	}

	// Now fix all project dependencies
	for i := 0; i < len(projects); i++ {
		prj := projects[i]

		// Add dependencies as libraries
		for _, depPrj := range prj.DevProject.Dependencies.Values {
			if idx, ok := projectNameToIndex[depPrj.Name]; ok {
				prj.Dependencies = append(prj.Dependencies, projects[idx])
			} else {
				// A dependency project should have matching config + target
				if !depPrj.HasMatchingConfigForTarget(buildConfig, buildTarget) {
					corepkg.LogWarnf("Dependency project %s does not have a matching config for target %s and build %s, skipping", depPrj.Name, buildTarget.String(), buildConfig.String())
					continue
				}
				depProject := NewProjectFromDevProject(depPrj, depPrj.Configs)
				projectNameToIndex[depPrj.Name] = len(projects)
				projects = append(projects, depProject)
				prj.AddLibrary(depProject)
			}
		}
	}

	// Glob all source files for each project
	exclusionFilter := denv.NewExclusionFilter(buildTarget)
	for _, prj := range projects {
		prj.GlobSourceFiles(exclusionFilter.IsExcluded)
	}

	return projects, nil
}
