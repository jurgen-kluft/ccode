package denv

import (
	"os"
	"strings"
)

// ProjectType defines the type of project, like 'StaticLibrary'
type ProjectType int

const (
	StaticLibrary ProjectType = 1  // StaticLibrary is a library that can statically be linked with
	SharedLibrary ProjectType = 2  // SharedLibrary is a library that can be dynamically linked with, like a .DLL
	Executable    ProjectType = 4  // Executable is an application that can be executed
	UnitTest      ProjectType = 8  // The project is a UnitTest
	Application   ProjectType = 16 // The project is an Application
	Library       ProjectType = 32 // The project is a Library, static or shared/dynamic
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

func (pt ProjectType) IsExecutable() bool {
	return pt&Executable != 0
}

// Project is a structure that holds all the information that defines a project in an IDE
type Project struct {
	Name         string
	Type         ProjectType
	PackageURL   string
	IncludeDirs  []string
	SourceDirs   []string
	Configs      []*Config
	Dependencies []*Project
}

// AddDefine adds a define
func (prj *Project) AddDefine(define string) {
	for _, cfg := range prj.Configs {
		cfg.Defines = append(cfg.Defines, define)
	}
}

func (prj *Project) AddLibs(libs []*Lib) {
	for _, cfg := range prj.Configs {
		for _, lib := range libs {
			if lib.Configs.Contains(cfg.Configs) {
				cfg.Libs = append(cfg.Libs, lib)
			}
		}
	}
}

// SetupDefaultCppLibProject returns a default C++ library project, since such a project can be used by
// an application as well as an unittest we need to add the appropriate configurations.
// Example:
//
//	SetupDefaultCppLibProject("cbase", "github.com/jurgen-kluft")
func SetupDefaultCppLibProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.Type = StaticLibrary | Library
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}

	project.IncludeDirs = []string{"source/main/include"}
	project.SourceDirs = []string{"source/main/cpp"}
	project.Configs = append(project.Configs, NewConfig(ConfigTypeStaticLibrary|ConfigTypeDebug|ConfigTypeDevelopment|ConfigTypeLibrary|ConfigTypeStaticLibrary))
	project.Configs = append(project.Configs, NewConfig(ConfigTypeStaticLibrary|ConfigTypeRelease|ConfigTypeDevelopment|ConfigTypeLibrary|ConfigTypeStaticLibrary))
	project.Configs = append(project.Configs, NewConfig(ConfigTypeStaticLibrary|ConfigTypeDebug|ConfigTypeDevelopment|ConfigTypeLibrary|ConfigTypeStaticLibrary|ConfigTypeUnittest))
	project.Configs = append(project.Configs, NewConfig(ConfigTypeStaticLibrary|ConfigTypeRelease|ConfigTypeDevelopment|ConfigTypeLibrary|ConfigTypeStaticLibrary|ConfigTypeUnittest))
	project.Dependencies = []*Project{}

	return project
}

func SetupDefaultCppLibProjectWithLibs(name string, URL string, Libs []*Lib) *Project {
	prj := SetupDefaultCppLibProject(name, URL)
	prj.AddLibs(Libs)
	return prj
}

// SetupDefaultCppTestProject returns a default C++ project
// Example:
//
//	SetupDefaultCppTestProject("cbase", "github.com\\jurgen-kluft")
func SetupDefaultCppTestProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.Type = Executable | UnitTest
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}
	project.IncludeDirs = []string{"source/main/include", "source/test/include"}
	project.SourceDirs = []string{"source/test/cpp"}
	project.Configs = append(project.Configs, NewConfig(ConfigTypeDebug|ConfigTypeDevelopment|ConfigTypeUnittest|ConfigTypeExecutable))
	project.Configs = append(project.Configs, NewConfig(ConfigTypeRelease|ConfigTypeDevelopment|ConfigTypeUnittest|ConfigTypeExecutable))
	project.Dependencies = []*Project{}

	return project
}

// SetupDefaultCppAppProject returns a default C++ application project
// Example:
//
//	SetupDefaultCppAppProject("cmyapp", "github.com\\jurgen-kluft")
func SetupDefaultCppAppProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.Type = Executable | Application
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}
	project.IncludeDirs = []string{"source/main/include", "source/app/include"}
	project.SourceDirs = []string{"source/app/cpp"}
	project.Configs = append(project.Configs, NewConfig(ConfigTypeDebug|ConfigTypeDevelopment|ConfigTypeExecutable))
	project.Configs = append(project.Configs, NewConfig(ConfigTypeRelease|ConfigTypeDevelopment|ConfigTypeExecutable))
	project.Dependencies = []*Project{}

	return project
}
