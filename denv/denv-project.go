package denv

import (
	"os"
	"path/filepath"
	"strings"

	cutils "github.com/jurgen-kluft/ccode/cutils"
)

type ExternalSrcFiles struct {
	Path         string // Absolute path
	BuildTargets BuildTargets
	SrcFiles     []string
}

func NewExternalSrcFiles(path string) *ExternalSrcFiles {
	return &ExternalSrcFiles{
		Path:         path,
		BuildTargets: BuildTargets{},
		SrcFiles:     []string{},
	}
}

// DevProject is a structure that holds all the information that defines a project in an IDE
type DevProject struct {
	Name            string
	Type            DevConfigType
	Vars            map[string]string
	BuildTargets    []BuildTarget
	PackageURL      string
	Configs         []*DevConfig
	ExternalSources []*ExternalSrcFiles
	Dependencies    []*DevProject
}

func NewProject(name string) *DevProject {
	return &DevProject{
		Name:            name,
		Type:            0,
		Vars:            make(map[string]string),
		BuildTargets:    []BuildTarget{},
		PackageURL:      "",
		Configs:         []*DevConfig{},
		ExternalSources: []*ExternalSrcFiles{},
		Dependencies:    []*DevProject{},
	}
}

func (prj *DevProject) AddEnvironmentVariable(ev string) {
	// Environment variable should exist
	if value, ok := os.LookupEnv(ev); ok {
		prj.Vars[strings.ToLower(ev)] = value
	}
}

func (prj *DevProject) ResolveEnvironmentVariables(str string) string {
	// Replace all environment variables in the string
	// Variables can be nested, so we need to know when a replace was
	// done, and repeat the replace until no more replacements are done.
	for {
		// Check if there are any environment variables in the string
		start := strings.Index(str, "{")
		if start < 0 {
			break
		}
		end := strings.Index(str[start:], "}")
		if end < 0 {
			break
		}

		if value, ok := prj.Vars[strings.ToLower(str[start+1:end])]; ok {
			str = strings.ReplaceAll(str, str[start:end+1], value)
			end = start
			continue
		}
		break
	}
	return str
}

func (prj *DevProject) AddDependency(dep *DevProject) {
	if dep != nil {
		prj.Dependencies = append(prj.Dependencies, dep)
	}
}

func (prj *DevProject) AddDependencies(deps ...*DevProject) {
	for _, dep := range deps {
		if dep != nil {
			prj.Dependencies = append(prj.Dependencies, dep)
		}
	}
}

func (prj *DevProject) AddExternalInclude(include string) {
	include = prj.ResolveEnvironmentVariables(include)
	for _, cfg := range prj.Configs {
		cfg.ExternalIncludeDirs = append(cfg.ExternalIncludeDirs, include)
	}
}

func (prj *DevProject) externalSourcesFrom(path string) *ExternalSrcFiles {
	path = prj.ResolveEnvironmentVariables(path)

	handleDir := func(rootPath, relPath string) bool {
		return true
	}

	externalSources := []string{}
	handleFile := func(rootPath, relPath string) {
		isCpp := filepath.Ext(relPath) == ".cpp"
		isC := !isCpp && filepath.Ext(relPath) == ".c"
		if isCpp || isC {
			externalSources = append(externalSources, relPath)
		}
	}

	// Scan for .c and .cpp files in that directory, recursively
	cutils.AddFilesFrom(path, handleDir, handleFile)

	externalSrcFiles := NewExternalSrcFiles(path)
	externalSrcFiles.SrcFiles = externalSources
	return externalSrcFiles
}

func (prj *DevProject) AddExternalSourcesFromForArduino(path string) {
	externalSrcFiles := prj.externalSourcesFrom(path)
	externalSrcFiles.BuildTargets = []BuildTarget{BuildTargetArduinoEsp32}
    prj.ExternalSources = append(prj.ExternalSources, externalSrcFiles)
}

func (prj *DevProject) AddDefine(define string) {
	define = prj.ResolveEnvironmentVariables(define)
	for _, cfg := range prj.Configs {
		cfg.Defines.Add(define)
	}
}

func (prj *DevProject) AddLibs(libs []*DevLib) {
	for _, cfg := range prj.Configs {
		for _, lib := range libs {
			if lib.Configs.Contains(cfg.ConfigType) {
				cfg.Libs = append(cfg.Libs, lib)
			}
		}
	}
}

// Used by IncludeFixer
func (proj *DevProject) CollectLocalIncludeDirs() *DevValueSet {
	includeDirs := NewDevValueSet()
	for _, cfg := range proj.Configs {
		includeDirs.AddMany(cfg.LocalIncludeDirs...)
	}
	return includeDirs
}

func (proj *DevProject) CollectSourceDirs() *DevValueSet {
	sourceDirs := NewDevValueSet()
	for _, cfg := range proj.Configs {
		sourceDirs.AddMany(cfg.SourceDirs...)
	}
	return sourceDirs
}

func (proj *DevProject) CollectProjectDependencies() []*DevProject {

	// Traverse and collect all dependencies
	projectMap := map[string]int{}
	projectList := []*DevProject{}
	for _, dp := range proj.Dependencies {
		if _, ok := projectMap[dp.Name]; !ok {
			projectMap[dp.Name] = len(projectList)
			projectList = append(projectList, dp)
		}
	}

	projectIdx := 0
	for projectIdx < len(projectList) {
		cp := projectList[projectIdx]
		projectIdx++

		for _, dp := range cp.Dependencies {
			if _, ok := projectMap[dp.Name]; !ok {
				projectMap[dp.Name] = len(projectList)
				projectList = append(projectList, dp)
			}
		}
	}
	return projectList
}

// SetupDefaultCppLibProject returns a default C++ library project, since such a project can be used by
// an application as well as an unittest we need to add the appropriate configurations.
// Example:
//
//	SetupDefaultCppLibProject("cbase", "github.com/jurgen-kluft")
func SetupDefaultCppLibProject(name string, URL string, dir string, buildTarget BuildTarget) *DevProject {
	project := NewProject(name)
	project.Type = DevConfigTypeStaticLibrary
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}

	project.Configs = append(project.Configs, NewDevConfig(DevConfigTypeStaticLibrary|DevConfigTypeDebug|DevConfigTypeDevelopment|DevConfigTypeStaticLibrary))
	project.Configs = append(project.Configs, NewDevConfig(DevConfigTypeStaticLibrary|DevConfigTypeRelease|DevConfigTypeDevelopment|DevConfigTypeStaticLibrary))
	project.Configs = append(project.Configs, NewDevConfig(DevConfigTypeStaticLibrary|DevConfigTypeDebug|DevConfigTypeDevelopment|DevConfigTypeStaticLibrary|DevConfigTypeTest))
	project.Configs = append(project.Configs, NewDevConfig(DevConfigTypeStaticLibrary|DevConfigTypeRelease|DevConfigTypeDevelopment|DevConfigTypeStaticLibrary|DevConfigTypeTest))
	project.Dependencies = []*DevProject{}

	for _, cfg := range project.Configs {
		configureProjectBasicConfiguration(cfg)
		configureProjectLibConfiguration(cfg, dir)
	}

	return project
}

func SetupCppLibProject(name string, URL string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppLibProject(name, URL, "main", GetBuildTarget())
	project.BuildTargets = append(project.BuildTargets, BuildTargetsAll...)
	return project
}

func SetupCppLibProjectForDesktop(name string, URL string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppLibProject(name, URL, "main", GetBuildTarget())
	project.BuildTargets = append(project.BuildTargets, BuildTargetsDesktop...)
	return project
}

func SetupCppLibProjectForArduino(name string, URL string) *DevProject {
	// Arduino Esp32
	project := SetupDefaultCppLibProject(name, URL, "main", BuildTargetArduinoEsp32)
	project.BuildTargets = append(project.BuildTargets, BuildTargetsArduino...)
	return project
}

func SetupCppLibProjectWithLibs(name string, URL string, Libs []*DevLib) *DevProject {
	project := SetupDefaultCppLibProject(name, URL, "main", GetBuildTarget())
	project.AddLibs(Libs)
	project.BuildTargets = append(project.BuildTargets, BuildTargetsArduino...)
	return project
}

func SetupDefaultCppTestProject(name string, URL string, buildTarget BuildTarget) *DevProject {
	project := NewProject(name)
	project.Type = DevConfigTypeExecutable | DevConfigTypeTest
	project.BuildTargets = append(project.BuildTargets, BuildTargetsDesktop...)
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}
	project.Configs = append(project.Configs, NewDevConfig(DevConfigTypeDebug|DevConfigTypeDevelopment|DevConfigTypeTest|DevConfigTypeExecutable))
	project.Configs = append(project.Configs, NewDevConfig(DevConfigTypeRelease|DevConfigTypeDevelopment|DevConfigTypeTest|DevConfigTypeExecutable))
	project.Dependencies = []*DevProject{}

	for _, cfg := range project.Configs {
		configureProjectBasicConfiguration(cfg)
		configureProjectTestConfiguration(cfg)
	}

	return project
}

func SetupCppTestProject(name string, URL string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppTestProject(name, URL, GetBuildTarget())
	project.BuildTargets = append(project.BuildTargets, BuildTargetsDesktop...)
	return project
}

func SetupCppTestProjectForDesktop(name string, URL string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppTestProject(name, URL, GetBuildTarget())
	project.BuildTargets = append(project.BuildTargets, BuildTargetsDesktop...)
	return project
}

// SetupDefaultCppCliProject returns a default C++ command-line program project
// Example:
//
//	SetupDefaultCppCliProject("cmycli", "github.com\\jurgen-kluft")
func SetupDefaultCppCliProject(name string, URL string, buildTarget BuildTarget) *DevProject {
	project := NewProject(name)
	project.Type = DevConfigTypeExecutable
	project.BuildTargets = append(project.BuildTargets, BuildTargetsDesktop...)
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}
	project.Configs = append(project.Configs, NewDevConfig(DevConfigTypeDebug|DevConfigTypeDevelopment|DevConfigTypeExecutable))
	project.Configs = append(project.Configs, NewDevConfig(DevConfigTypeRelease|DevConfigTypeDevelopment|DevConfigTypeExecutable))
	project.Dependencies = []*DevProject{}

	for _, cfg := range project.Configs {
		configureProjectBasicConfiguration(cfg)
		configureProjectCliConfiguration(cfg)
	}

	return project
}

// SetupDefaultCppAppProject returns a default C++ application project
// Example:
//
//	SetupDefaultCppAppProject("cmyapp", "github.com\\jurgen-kluft")
func SetupDefaultCppAppProject(name string, URL string, buildTarget BuildTarget) *DevProject {
	project := NewProject(name)
	project.Type = DevConfigTypeExecutable
	project.BuildTargets = append(project.BuildTargets, BuildTargetsDesktop...)
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}
	project.Configs = append(project.Configs, NewDevConfig(DevConfigTypeDebug|DevConfigTypeDevelopment|DevConfigTypeExecutable))
	project.Configs = append(project.Configs, NewDevConfig(DevConfigTypeRelease|DevConfigTypeDevelopment|DevConfigTypeExecutable))
	project.Dependencies = []*DevProject{}

	for _, cfg := range project.Configs {
		configureProjectBasicConfiguration(cfg)
		configureProjectAppConfiguration(cfg)
	}

	return project
}

func SetupCppAppProject(name string, URL string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppAppProject(name, URL, GetBuildTarget())
	project.BuildTargets = append(project.BuildTargets, BuildTargetsAll...)
	return project
}

func SetupCppAppProjectForDesktop(name string, URL string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppAppProject(name, URL, GetBuildTarget())
	project.BuildTargets = append(project.BuildTargets, BuildTargetsDesktop...)
	return project
}

func SetupCppAppProjectForArduino(name string, URL string) *DevProject {
	// Arduino project
	project := SetupDefaultCppAppProject(name, URL, BuildTargetArduinoEsp32)
	project.BuildTargets = append(project.BuildTargets, BuildTargetsArduino...)
	return project
}

func configureProjectLibConfiguration(config *DevConfig, name string) {
	config.LocalIncludeDirs = append(config.LocalIncludeDirs, "source/"+name+"/include")
	config.SourceDirs = append(config.SourceDirs, "source/"+name+"/cpp")
}

func configureProjectProgramConfiguration(prg string, libs []string, config *DevConfig) {
	configureProjectLibConfiguration(config, prg)
	for _, lib := range libs {
		configureProjectLibConfiguration(config, lib)
	}
}

func configureProjectTestConfiguration(config *DevConfig) {
	configureProjectProgramConfiguration("test", []string{}, config)
}

func configureProjectCliConfiguration(config *DevConfig) {
	configureProjectProgramConfiguration("cli", []string{}, config)
}

func configureProjectAppConfiguration(config *DevConfig) {
	configureProjectProgramConfiguration("app", []string{}, config)
}

func configureProjectBasicConfiguration(config *DevConfig) {
	configType := config.ConfigType
	if configType.IsDebug() {
		config.Defines.AddMany("TARGET_DEBUG", "_DEBUG")
	} else if configType.IsRelease() {
		config.Defines.AddMany("TARGET_RELEASE", "NDEBUG")
	} else if configType.IsFinal() {
		config.Defines.AddMany("TARGET_FINAL", "NDEBUG")
	}

	if configType.IsProfile() {
		config.Defines.AddMany("TARGET_RELEASE", "TARGET_PROFILE", "NDEBUG")
	}

	if configType.IsTest() {
		config.Defines.AddMany("TARGET_TEST")
	}
}
