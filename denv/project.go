package denv

import (
	"os"
	"strings"
)

// ProjectType defines the type of project, like 'StaticLibrary'
type ProjectType int

const (
	StaticLibrary ProjectType = 1   // StaticLibrary is a library that can statically be linked with
	SharedLibrary ProjectType = 2   // SharedLibrary is a library that can be dynamically linked with, like a .DLL
	Executable    ProjectType = 4   // Executable is an application that can be executed
	UnitTest      ProjectType = 8   // The project is a UnitTest
	Application   ProjectType = 16  // The project is an Application
	Library       ProjectType = 32  // The project is a Library, static or shared/dynamic
	CLanguage     ProjectType = 64  // The project is a C language project
	CppLanguage   ProjectType = 128 // The project is a C++ language project
	Headers       ProjectType = 256 // The project is a header only project
)

func (pt ProjectType) IsUnitTest() bool {
	return pt&UnitTest != 0
}

func (pt ProjectType) IsApplication() bool {
	return pt&Application != 0
}

func (pt ProjectType) IsLibrary() bool {
	return pt&Library != 0
}

func (pt ProjectType) IsStaticLibrary() bool {
	return pt&StaticLibrary != 0
}

func (pt ProjectType) IsSharedLibrary() bool {
	return pt&SharedLibrary != 0
}

func (pt ProjectType) IsExecutable() bool {
	return pt&Executable != 0
}

func (pt ProjectType) IsCLanguage() bool {
	return pt&CLanguage != 0
}

func (pt ProjectType) IsCppLanguage() bool {
	return pt&CppLanguage != 0
}

func (pt ProjectType) IsHeaders() bool {
	return pt&Headers != 0
}

// Project is a structure that holds all the information that defines a project in an IDE
type Project struct {
	Name         string
	Type         ProjectType
	BuildTargets []BuildTarget
	PackageURL   string
	Configs      []*Config
	Dependencies []*Project
}

func (prj *Project) AddDependency(dep *Project) {
	if dep != nil {
		prj.Dependencies = append(prj.Dependencies, dep)
	}
}

func (prj *Project) AddDependencies(deps ...*Project) {
	for _, dep := range deps {
		if dep != nil {
			prj.Dependencies = append(prj.Dependencies, dep)
		}
	}
}

// AddDefine adds a define
func (prj *Project) AddDefine(define string) {
	for _, cfg := range prj.Configs {
		cfg.Defines.Add(define)
	}
}

func (prj *Project) AddLibs(libs []*Lib) {
	for _, cfg := range prj.Configs {
		for _, lib := range libs {
			if lib.Configs.Contains(cfg.ConfigType) {
				cfg.Libs = append(cfg.Libs, lib)
			}
		}
	}
}

func (proj *Project) CollectIncludeDirs() *ValueSet {
	includeDirs := NewValueSet()
	for _, cfg := range proj.Configs {
		includeDirs.AddMany(cfg.IncludeDirs...)
	}
	return includeDirs
}

func (proj *Project) CollectSourceDirs() *ValueSet {
	sourceDirs := NewValueSet()
	for _, cfg := range proj.Configs {
		sourceDirs.AddMany(cfg.SourceDirs...)
	}
	return sourceDirs
}

func (proj *Project) CollectProjectDependencies() []*Project {

	// Traverse and collect all dependencies
	projectMap := map[string]int{}
	projectList := []*Project{}
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
func SetupDefaultCppLibProject(name string, URL string, dir string) *Project {
	project := &Project{Name: name}
	project.Type = StaticLibrary | Library | CppLanguage
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}

	project.Configs = append(project.Configs, NewConfig(ConfigTypeStaticLibrary|ConfigTypeDebug|ConfigTypeDevelopment|ConfigTypeLibrary|ConfigTypeStaticLibrary))
	project.Configs = append(project.Configs, NewConfig(ConfigTypeStaticLibrary|ConfigTypeRelease|ConfigTypeDevelopment|ConfigTypeLibrary|ConfigTypeStaticLibrary))
	project.Configs = append(project.Configs, NewConfig(ConfigTypeStaticLibrary|ConfigTypeDebug|ConfigTypeDevelopment|ConfigTypeLibrary|ConfigTypeStaticLibrary|ConfigTypeUnittest))
	project.Configs = append(project.Configs, NewConfig(ConfigTypeStaticLibrary|ConfigTypeRelease|ConfigTypeDevelopment|ConfigTypeLibrary|ConfigTypeStaticLibrary|ConfigTypeUnittest))
	project.Dependencies = []*Project{}

	for _, cfg := range project.Configs {
		configureProjectBasicConfiguration(cfg)
		configureProjectPlatformConfiguration(cfg)
		configureProjectLocalizedConfiguration(cfg)
		configureProjectLibConfiguration(cfg, dir)
	}

	return project
}

func SetupCppLibProject(name string, URL string) *Project {
	// All platforms
	project := SetupDefaultCppLibProject(name, URL, "main")
	project.BuildTargets = BuildTargetsAll
	return project
}

func SetupCppLibProjectForDesktop(name string, URL string) *Project {
	// Windows, Mac and Linux project
	project := SetupDefaultCppLibProject(name, URL, "main")
	project.BuildTargets = BuildTargetsDesktop
	return project
}

func SetupCppLibProjectForArduino(name string, URL string) *Project {
	// Arduino project
	project := SetupDefaultCppLibProject(name, URL, "main")
	project.BuildTargets = BuildTargetsArduino
	return project
}

func SetupCppLibProjectWithLibs(name string, URL string, Libs []*Lib) *Project {
	project := SetupDefaultCppLibProject(name, URL, "main")
	project.AddLibs(Libs)
	project.BuildTargets = BuildTargetsArduino
	return project
}

// SetupDefaultCppTestProject returns a default C++ project
// Example:
//
//	SetupDefaultCppTestProject("cbase", "github.com\\jurgen-kluft")
func SetupDefaultCppTestProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.Type = Executable | UnitTest | CppLanguage
	project.BuildTargets = BuildTargetsDesktop
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}
	project.Configs = append(project.Configs, NewConfig(ConfigTypeDebug|ConfigTypeDevelopment|ConfigTypeUnittest|ConfigTypeExecutable))
	project.Configs = append(project.Configs, NewConfig(ConfigTypeRelease|ConfigTypeDevelopment|ConfigTypeUnittest|ConfigTypeExecutable))
	project.Dependencies = []*Project{}

	for _, cfg := range project.Configs {
		configureProjectBasicConfiguration(cfg)
		configureProjectPlatformConfiguration(cfg)
		configureProjectLocalizedConfiguration(cfg)
		configureProjectTestConfiguration(cfg)
	}

	return project
}

func SetupCppTestProjectForDesktop(name string, URL string) *Project {
	// Windows, Mac and Linux project
	project := SetupDefaultCppTestProject(name, URL)
	project.BuildTargets = BuildTargetsDesktop
	return project
}

// SetupDefaultCppCliProject returns a default C++ command-line program project
// Example:
//
//	SetupDefaultCppCliProject("cmycli", "github.com\\jurgen-kluft")
func SetupDefaultCppCliProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.Type = Executable | Application | CppLanguage
	project.BuildTargets = BuildTargetsDesktop
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}
	project.Configs = append(project.Configs, NewConfig(ConfigTypeDebug|ConfigTypeDevelopment|ConfigTypeExecutable))
	project.Configs = append(project.Configs, NewConfig(ConfigTypeRelease|ConfigTypeDevelopment|ConfigTypeExecutable))
	project.Dependencies = []*Project{}

	for _, cfg := range project.Configs {
		configureProjectBasicConfiguration(cfg)
		configureProjectPlatformConfiguration(cfg)
		configureProjectLocalizedConfiguration(cfg)
		configureProjectCliConfiguration(cfg)
	}

	return project
}

// SetupDefaultCppAppProject returns a default C++ application project
// Example:
//
//	SetupDefaultCppAppProject("cmyapp", "github.com\\jurgen-kluft")
func SetupDefaultCppAppProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.Type = Executable | Application | CppLanguage
	project.BuildTargets = BuildTargetsDesktop
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}
	project.Configs = append(project.Configs, NewConfig(ConfigTypeDebug|ConfigTypeDevelopment|ConfigTypeExecutable))
	project.Configs = append(project.Configs, NewConfig(ConfigTypeRelease|ConfigTypeDevelopment|ConfigTypeExecutable))
	project.Dependencies = []*Project{}

	for _, cfg := range project.Configs {
		configureProjectBasicConfiguration(cfg)
		configureProjectPlatformConfiguration(cfg)
		configureProjectLocalizedConfiguration(cfg)
		configureProjectAppConfiguration(cfg)
	}

	return project
}

func SetupCppAppProject(name string, URL string) *Project {
	// All platforms
	project := SetupDefaultCppAppProject(name, URL)
	project.BuildTargets = BuildTargetsAll
	return project
}

func SetupCppAppProjectForDesktop(name string, URL string) *Project {
	// Windows, Mac and Linux project
	project := SetupDefaultCppAppProject(name, URL)
	project.BuildTargets = BuildTargetsDesktop
	return project
}

func SetupCppAppProjectForArduino(name string, URL string) *Project {
	// Arduino project
	project := SetupDefaultCppAppProject(name, URL)
	project.BuildTargets = BuildTargetsArduino
	return project
}

func configureProjectLibConfiguration(config *Config, name string) {
	config.IncludeDirs = append(config.IncludeDirs, "source/"+name+"/include")
	config.SourceDirs = append(config.SourceDirs, "source/"+name+"/cpp")
}

func configureProjectProgramConfiguration(prg string, libs []string, config *Config) {
	configureProjectLibConfiguration(config, prg)
	for _, lib := range libs {
		configureProjectLibConfiguration(config, lib)
	}
}

func configureProjectTestConfiguration(config *Config) {
	configureProjectProgramConfiguration("test", []string{}, config)
}

func configureProjectCliConfiguration(config *Config) {
	configureProjectProgramConfiguration("cli", []string{}, config)
}

func configureProjectAppConfiguration(config *Config) {
	configureProjectProgramConfiguration("app", []string{}, config)
}

func configureProjectBasicConfiguration(config *Config) {
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

	if configType.IsUnittest() {
		config.Defines.AddMany("TARGET_TEST")
	}
}

func configureProjectPlatformConfiguration(config *Config) {
	if IsWindows() {
		config.Defines.Add("TARGET_PC")
	} else if IsLinux() {
		config.Defines.Add("TARGET_LINUX")
	} else if IsMacOS() {
		config.Defines.Add("TARGET_MAC")
		config.LinkFlags.Add("-ObjC")

		// Foundation
		// Cocoa
		// Carbon
		// Metal
		// OpenGL
		// IOKit
		// AppKit
		// CoreVideo
		// QuartzCore

		// Add Cocoa Frameworks
		config.Libs = append(config.Libs, &Lib{Configs: ConfigTypeAll, Type: Framework, Files: []string{"Foundation", "Cocoa", "Carbon", "Metal", "OpenGL", "IOKit", "AppKit", "CoreVideo", "QuartzCore"}})
	}
}

func configureProjectLocalizedConfiguration(config *Config) {
	config.Defines.AddMany("_UNICODE", "UNICODE")
}
