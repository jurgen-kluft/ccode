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
	Executable ProjectType = 3 // .exe, .app
)

// Project is a structure that holds all the information that defines a project in an IDE
type Project struct {
	Name         string
	Type         ProjectType
	ProjectPath  string
	PackagePath  string
	PackageURL   string
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
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}
	project.Type = StaticLibrary

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

	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}

	project.Type = Executable

	project.Configs = append(project.Configs, NewDebugConfig())
	project.Configs = append(project.Configs, NewReleaseConfig())
	project.Dependencies = []*Project{}

	return project
}

// SetupDefaultCppAppProject returns a default C++ project
// Example:
//
//	SetupDefaultCppAppProject("cbase", "github.com\\jurgen-kluft")
func SetupDefaultCppAppProject(name string, URL string) *Project {
	project := &Project{Name: name}
	if os.PathSeparator == '\\' {
		project.PackageURL = strings.Replace(URL, "/", "\\", -1)
	} else {
		project.PackageURL = strings.Replace(URL, "\\", "/", -1)
	}
	project.Type = Executable

	project.Configs = append(project.Configs, NewDebugConfig())
	project.Configs = append(project.Configs, NewReleaseConfig())
	project.Dependencies = []*Project{}

	return project
}
