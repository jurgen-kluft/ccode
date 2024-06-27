package denv

import (
	"os"
	"strings"
)

// ProjectType defines the type of project, like 'StaticLibrary'
type ProjectType int

const (
	// StaticLibrary is a library that can statically be linked with
	StaticLibrary ProjectType = 1 // .lib, .a
	// SharedLibrary is a library that can be dynamically linked with, like a .DLL
	SharedLibrary ProjectType = 2 // .dll
	// Executable is an application that can be run
	Executable ProjectType = 4 // .exe, .app
)

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

// SetupDefaultCppLibProject returns a default C++ project
// Example:
//
//	SetupDefaultCppLibProject("cbase", "github.com/jurgen-kluft")
func SetupDefaultCppLibProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.Type = StaticLibrary
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}

	project.IncludeDirs = []string{"source/main/include"}
	project.SourceDirs = []string{"source/main/cpp"}
	project.Configs = append(project.Configs, NewDebugConfig())
	project.Configs = append(project.Configs, NewReleaseConfig())
	project.Dependencies = []*Project{}

	return project
}

// SetupDefaultCppTestProject returns a default C++ project
// Example:
//
//	SetupDefaultCppTestProject("cbase", "github.com\\jurgen-kluft")
func SetupDefaultCppTestProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.Type = Executable
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}
	project.IncludeDirs = []string{"source/main/include", "source/test/include"}
	project.SourceDirs = []string{"source/test/cpp"}
	project.Configs = append(project.Configs, NewDebugConfig())
	project.Configs = append(project.Configs, NewReleaseConfig())
	project.Dependencies = []*Project{}

	return project
}

// SetupDefaultCppAppProject returns a default C++ project
// Example:
//
//	SetupDefaultCppAppProject("cmyapp", "github.com\\jurgen-kluft")
func SetupDefaultCppAppProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.Type = Executable
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}
	project.IncludeDirs = []string{"source/main/include", "source/app/include"}
	project.SourceDirs = []string{"source/app/cpp"}
	project.Configs = append(project.Configs, NewDebugConfig())
	project.Configs = append(project.Configs, NewReleaseConfig())
	project.Dependencies = []*Project{}

	return project
}
