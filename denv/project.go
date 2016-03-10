package denv

import (
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/xcode/glob"
	"github.com/jurgen-kluft/xcode/uid"
	"github.com/jurgen-kluft/xcode/vars"
)

// Files helps to collect source and header files as well as virtual files as they
// can be presented in an IDE
type Files struct {
	GlobPaths    []string
	VirtualPaths []string
	Files        []string
}

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
	ProjectPath  string
	PackagePath  string
	PackageURL   string
	Name         string
	Type         ProjectType
	Author       string
	GUID         string
	Language     string
	Platforms    []string
	HdrFiles     *Files
	SrcFiles     *Files
	Configs      ConfigSet
	Dependencies []*Project
}

// HasPlatform returns true if the project is configured for that platform
func (prj *Project) HasPlatform(platformname string) bool {
	for _, platform := range prj.Platforms {
		if platformname == platform {
			return true
		}
	}
	return false
}

// HasConfig returns true if the project has that configuration
func (prj *Project) HasConfig(configname string) bool {
	return prj.Configs.HasConfig(configname)
}

// SupportedPlatforms returns a list of platforms that are supported by xcode
var SupportedPlatforms = []string{
	"Win32",
	"x64",
}

// ReplaceVars replaces any variable that exists in members of Project
func (prj *Project) ReplaceVars(v vars.Variables, r vars.Replacer) {
	v.AddVar("${Name}", prj.Name)
	for _, config := range prj.Configs {
		config.ReplaceVars(v, r)
	}
	v.DelVar("${Name}")
}

// SetupDefaultCppLibProject returns a default C++ project
// Example:
//              SetupDefaultCppLibProject("xbase", "github.com\\jurgen-kluft")
//
func SetupDefaultCppLibProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.GUID = uid.GetGUID(project.Name)
	project.PackageURL = URL
	project.Language = "C++"
	project.Type = StaticLibrary

	project.SrcFiles = &Files{GlobPaths: []string{Path("source\\main\\^**\\*.cpp")}, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: []string{Path("source\\main\\include\\^**\\*.h")}}

	project.Platforms = SupportedPlatforms
	project.Configs = GetDefaultConfigs()
	project.Dependencies = []*Project{}
	return project
}

// SetupDefaultCppTestProject returns a default C++ project
// Example:
//              SetupDefaultCppTestProject("xbase", "github.com\\jurgen-kluft")
//
func SetupDefaultCppTestProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.GUID = uid.GetGUID(project.Name)
	project.PackageURL = URL
	project.Language = "C++"
	project.Type = Executable

	project.SrcFiles = &Files{GlobPaths: []string{Path("source\\test\\^**\\*.cpp")}, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: []string{Path("source\\main\\include\\^**\\*.h"), Path("source\\test\\include\\^**\\*.h")}}

	project.Platforms = SupportedPlatforms
	project.Configs = GetDefaultConfigs()
	project.Dependencies = []*Project{}
	return project
}

// SetupDefaultCppAppProject returns a default C++ project
// Example:
//              SetupDefaultCppAppProject("xbase", "github.com\\jurgen-kluft")
//
func SetupDefaultCppAppProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.GUID = uid.GetGUID(project.Name)
	project.PackageURL = URL
	project.Language = "C++"
	project.Type = Executable

	project.SrcFiles = &Files{GlobPaths: []string{Path("source\\main\\^**\\*.cpp")}, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: []string{Path("source\\main\\include\\^**\\*.h")}}

	project.Platforms = SupportedPlatforms
	project.Configs = GetDefaultConfigs()
	project.Dependencies = []*Project{}
	return project
}

// GlobFiles will collect files that can be found in @dirpath that matches
// any of the Files.GlobPaths into Files.Files
func (f *Files) GlobFiles(dirpath string) {
	// Glob all the on-disk files
	for _, g := range f.GlobPaths {
		pp := strings.Split(g, "^")
		ppath := filepath.Join(dirpath, pp[0])
		f.Files, _ = glob.GlobFiles(ppath, pp[1])
		for i, file := range f.Files {
			f.Files[i] = filepath.Join(pp[0], file)
		}
	}

	// Generate the virtual files

}
