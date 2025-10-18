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
	"github.com/jurgen-kluft/ccode/denv"
)

const (
	BuildInfoFilenameWithoutExt = "buildinfo"
)

type AppConfig struct {
	ProjectName string `json:"project,omitempty"`
	TargetOs    string `json:"os,omitempty"`
	TargetArch  string `json:"arch,omitempty"`
	TargetBuild string `json:"build,omitempty"`
	TargetBoard string `json:"board,omitempty"`
}

func (c *AppConfig) Equal(other *AppConfig) bool {
	return c.ProjectName == other.ProjectName &&
		c.TargetOs == other.TargetOs &&
		c.TargetArch == other.TargetArch &&
		c.TargetBuild == other.TargetBuild &&
		c.TargetBoard == other.TargetBoard
}

type App struct {
	Pkg         *denv.Package
	Config      *AppConfig
	BuildTarget denv.BuildTarget
	BuildConfig denv.BuildConfig
}

func NewApp(pkg *denv.Package) *App {
	return &App{Pkg: pkg, Config: &AppConfig{}}
}

func GetBuildDirname(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) string {
	return buildTarget.Os().String() + "-" + buildTarget.Arch().String() + "-" + buildConfig.String()
}

func (a *App) GetBuildPath(subdir string) string {
	var buildPath string
	if len(a.Config.TargetBoard) > 0 {
		buildPath = filepath.Join("build", subdir, a.Config.TargetBoard)
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

func ParseProjectNameAndConfig(app *App) {
	flag.StringVar(&app.Config.ProjectName, "p", "", "Name of the project")
	flag.StringVar(&app.Config.TargetOs, "os", "", "Target OS (windows, darwin, linux, arduino)")
	flag.StringVar(&app.Config.TargetBuild, "build", "", "Format 'build' or 'build-variant', e.g. debug, debug-dev, release-dev, debug-dev-test)")
	flag.StringVar(&app.Config.TargetArch, "arch", "", "Cpu Architecture (amd64, x64, arm64, esp32, esp8266)")
	flag.StringVar(&app.Config.TargetBoard, "board", "", "Board name (s3, c3, xiao-c3, ...)")
	flag.Parse()

	app.Config.TargetArch = strings.ToLower(app.Config.TargetArch)
	app.Config.TargetOs = strings.ToLower(app.Config.TargetOs)
	app.Config.TargetBuild = strings.ToLower(app.Config.TargetBuild)
	app.Config.TargetBoard = strings.ToLower(app.Config.TargetBoard)

	configFilePath := "clay.json"
	loadedConfig := &AppConfig{}
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

	if len(app.Config.ProjectName) == 0 {
		app.Config.ProjectName = loadedConfig.ProjectName
	}

	if len(app.Config.TargetBoard) > 0 {
		app.Config.TargetOs = "arduino"
		if len(app.Config.TargetArch) != 0 && app.Config.TargetArch != "esp32" && app.Config.TargetArch != "esp8266" {
			app.Config.TargetArch = "esp32"
		}
	}

	if len(app.Config.TargetOs) == 0 {
		app.Config.TargetOs = loadedConfig.TargetOs
	} else {
		// If the target OS was specified on the command line, clear the board and arch
		switch app.Config.TargetOs {
		case "native":
			app.Config.TargetBoard = ""
			app.Config.TargetArch = GetNativeArch()
			app.Config.TargetOs = GetNativeOs()
		case "windows":
			app.Config.TargetBoard = ""
			app.Config.TargetArch = "x64"
		case "darwin":
			app.Config.TargetBoard = ""
			app.Config.TargetArch = GetNativeArch()
		case "linux":
			app.Config.TargetBoard = ""
			app.Config.TargetArch = "amd64"
		case "arduino":
			if len(app.Config.TargetArch) == 0 {
				app.Config.TargetArch = "esp32"
			}
		}
	}
	if len(app.Config.TargetOs) == 0 {
		switch runtime.GOOS {
		case "windows":
			app.Config.TargetOs = "windows"
		case "darwin":
			app.Config.TargetOs = "darwin"
		default:
			app.Config.TargetOs = "linux"
		}
	}

	if app.Config.TargetOs == "arduino" {
		if len(app.Config.TargetBoard) == 0 {
			app.Config.TargetBoard = loadedConfig.TargetBoard
		}
	}

	if len(app.Config.TargetBuild) == 0 {
		app.Config.TargetBuild = loadedConfig.TargetBuild
	}
	if len(app.Config.TargetBuild) == 0 {
		app.Config.TargetBuild = "dev-release"
	}

	if len(app.Config.TargetArch) == 0 {
		app.Config.TargetArch = loadedConfig.TargetArch
	}
	if len(app.Config.TargetArch) == 0 {
		app.Config.TargetArch = runtime.GOARCH
		switch app.Config.TargetOs {
		case "arduino":
			app.Config.TargetArch = "esp32"
		case "darwin":
			app.Config.TargetArch = "arm64"
		case "windows":
			app.Config.TargetArch = "x64"
		case "linux":
			app.Config.TargetArch = "amd64"
		}
	}

	// If any of the config values were updated from the command line flags, write back the config file
	// Compare the loaded and current config, when they are different we need to update the file
	if app.Config.Equal(loadedConfig) == false {
		jsonContent, err := json.MarshalIndent(app.Config, "", "  ")
		if err != nil {
			corepkg.LogFatalf("Failed to marshal config file", configFilePath, err)
		}
		if err := os.WriteFile(configFilePath, jsonContent, 0644); err != nil {
			corepkg.LogFatalf("Failed to write config file", configFilePath, err)
		}
	}

	corepkg.LogInfof("Project: %s", app.Config.ProjectName)
	corepkg.LogInfof("Os: %s", app.Config.TargetOs)
	corepkg.LogInfof("Arch: %s", app.Config.TargetArch)
	corepkg.LogInfof("Build: %s", app.Config.TargetBuild)
	if len(app.Config.TargetBoard) > 0 {
		corepkg.LogInfof("Board: %s", app.Config.TargetBoard)
	}

	app.BuildConfig = denv.BuildConfigFromString(app.Config.TargetBuild)
	buildTargetStr := fmt.Sprintf("%s(%s)", app.Config.TargetOs, app.Config.TargetArch)
	app.BuildTarget = denv.BuildTargetFromString(buildTargetStr)
}

func (a *App) Build() (err error) {
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

	var numberOfProjects int
	var numberOfNoMatchConfigs int
	var outOfDate int

	// First build libraries
	for _, prj := range prjs {
		if prj.DevProject.BuildType.IsLibrary() && prj.CanBuildFor(a.BuildConfig, a.BuildTarget) {
			numberOfProjects++
			if outOfDate, err = prj.Build(a.BuildConfig, a.BuildTarget, buildPath); err != nil {
				return err
			}
		} else {
			numberOfNoMatchConfigs++
		}
	}

	// Then build applications
	for _, prj := range prjs {
		if prj.DevProject.BuildType.IsExecutable() && prj.CanBuildFor(a.BuildConfig, a.BuildTarget) {
			numberOfProjects++
			if outOfDate, err = prj.Build(a.BuildConfig, a.BuildTarget, buildPath); err != nil {
				return err
			}
		} else {
			numberOfNoMatchConfigs++
		}
	}

	if outOfDate == 0 && numberOfProjects > 0 {
		corepkg.LogInfo("Nothing to build, everything is up to date")
	} else if numberOfProjects == 0 {
		corepkg.LogError(fmt.Errorf("!"), "No matching project configurations found")
	}

	return err
}

func (a *App) Clean() error {
	buildPath := a.GetBuildPath(GetBuildDirname(a.BuildConfig, a.BuildTarget))
	prjs, err := a.CreateProjects(a.BuildTarget, a.BuildConfig)
	if err != nil {
		return err
	}

	for _, prj := range prjs {
		for _, cfg := range prj.Config {
			if cfg.BuildConfig.IsEqual(a.BuildConfig) {
				fmt.Println(prj.DevProject.Name, a.BuildConfig.String())

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

func (a *App) ListLibraries() error {
	prjs, err := a.CreateProjects(a.BuildTarget, a.BuildConfig)
	if err != nil {
		return err
	}

	configs := make([]string, 0, 16)
	nameToIndex := make(map[string]int)
	for _, prj := range prjs {
		if idx, ok := nameToIndex[prj.DevProject.Name]; !ok {
			idx = len(configs)
			nameToIndex[prj.DevProject.Name] = idx
			configs = append(configs, a.BuildConfig.String())
		} else {
			configs[idx] += ", " + a.BuildConfig.String()
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
func (a *App) CreateProjects(buildTarget denv.BuildTarget, buildConfig denv.BuildConfig) ([]*Project, error) {
	projects := make([]*Project, 0, 4)

	// First create the projects without dependencies
	projectNameToIndex := make(map[string]int)

	for _, devPrj := range a.Pkg.GetMainLib() {
		if !devPrj.HasMatchingConfigForTarget(buildConfig, buildTarget) {
			continue
		}
		project := NewProjectFromDevProject(devPrj, devPrj.Configs)
		projectNameToIndex[devPrj.Name] = len(projects)
		projects = append(projects, project)
	}

	for _, devPrj := range a.Pkg.GetMainApp() {
		if !devPrj.HasMatchingConfigForTarget(buildConfig, buildTarget) {
			continue
		}
		project := NewProjectFromDevProject(devPrj, devPrj.Configs)
		projectNameToIndex[devPrj.Name] = len(projects)
		projects = append(projects, project)
	}

	for _, devPrj := range a.Pkg.GetTestLib() {
		if !devPrj.HasMatchingConfigForTarget(buildConfig, buildTarget) {
			continue
		}
		project := NewProjectFromDevProject(devPrj, devPrj.Configs)
		projectNameToIndex[devPrj.Name] = len(projects)
		projects = append(projects, project)
	}

	for _, devPrj := range a.Pkg.GetUnittest() {
		if !devPrj.HasMatchingConfigForTarget(buildConfig, buildTarget) {
			continue
		}
		project := NewProjectFromDevProject(devPrj, devPrj.Configs)
		projectNameToIndex[devPrj.Name] = len(projects)
		projects = append(projects, project)
	}

	// Now fix all project dependencies
	prjDependencyList := denv.NewDevProjectList()
	for i := 0; i < len(projects); i++ {
		prj := projects[i]
		prjDependencyList.Reset()
		prj.DevProject.CollectProjectDependencies(prjDependencyList)
		// Add dependencies as libraries
		for _, depPrj := range prjDependencyList.Values {
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
				prj.Dependencies = append(prj.Dependencies, depProject)
			}
		}
	}

	// Glob all source files for each project
	exclusionFilter := denv.NewExclusionFilter(buildTarget)
	for _, prj := range projects {
		prj.GlobSourceFiles(exclusionFilter.IsExcluded)
	}

	// Collect all variables for each project

	return projects, nil
}

func (a *App) SetToolchain(p *Project, buildPath string) (err error) {
	if a.BuildTarget.Arduino() && a.BuildTarget.Esp32() {
		vars := corepkg.NewVars(corepkg.VarsFormatCurlyBraces)
		a.Pkg.GetVars(a.BuildTarget, a.BuildConfig, a.Config.TargetBoard, vars)
		p.Toolchain = toolchain.NewArduinoEsp32Toolchainv2(vars, p.DevProject.Name, p.GetBuildPath(buildPath))
	} else if a.BuildTarget.Arduino() && a.BuildTarget.Esp8266() {
		vars := corepkg.NewVars(corepkg.VarsFormatCurlyBraces)
		a.Pkg.GetVars(a.BuildTarget, a.BuildConfig, a.Config.TargetBoard, vars)
		p.Toolchain = toolchain.NewArduinoEsp8266Toolchain(vars, p.DevProject.Name, p.GetBuildPath(buildPath))
	} else if a.BuildTarget.Windows() {
		p.Toolchain, err = toolchain.NewWinMsdev(a.BuildTarget.Arch().String(), "Desktop")
	} else if a.BuildTarget.Mac() {
		vars := corepkg.NewVars(corepkg.VarsFormatCurlyBraces)
		a.Pkg.GetVars(a.BuildTarget, a.BuildConfig, a.Config.TargetBoard, vars)
		p.Toolchain = toolchain.NewDarwinClangv2(vars, p.DevProject.Name, p.GetBuildPath(buildPath))
	} else {
		err = corepkg.LogErrorf(os.ErrNotExist, "error, %s as a build target on %s is not supported", a.BuildTarget.Os().String(), runtime.GOOS)
	}
	return err
}
