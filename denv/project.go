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

func (f *Files) AddGlobPath(dirpath string) {
	f.GlobPaths = append(f.GlobPaths, dirpath)
}

// GlobFiles will collect files that can be found in @dirpath that matches
// any of the Files.GlobPaths into Files.Files
func (f *Files) GlobFiles(dirpath string) {
	// Glob all the on-disk files
	for _, g := range f.GlobPaths {
		pp := strings.Split(g, "^")
		ppath := filepath.Join(dirpath, pp[0])

		globbedfiles, _ := glob.GlobFiles(ppath, pp[1])
		for _, file := range globbedfiles {
			globbedfile := filepath.Join(pp[0], file)
			f.Files = append(f.Files, globbedfile)
		}
	}

	// Generate the virtual files

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

const (
	// CppLanguageToken is the language token for C++
	CppLanguageToken string = "C++"
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
	Platforms    PlatformSet
	HdrFiles     *Files
	SrcFiles     *Files
	Dependencies []*Project
}

// HasPlatform returns true if the project is configured for that platform
func (prj *Project) HasPlatform(platformname string) bool {
	return prj.Platforms.HasPlatform(platformname)
}

// HasConfig will return true if platform @platformname has a configuration with name @configname
func (prj *Project) HasConfig(platformname, configname string) bool {
	for _, platform := range prj.Platforms {
		if platform.Name == platformname {
			if platform.HasConfig(configname) == false {
				return false
			}
		}
	}
	return true
}

// GetConfig will return the configuration of platform @platformname with name @configname
func (prj *Project) GetConfig(platformname, configname string) (*Config, bool) {
	for _, platform := range prj.Platforms {
		if platform.Name == platformname {
			return platform.GetConfig(configname)
		}
	}
	return nil, false
}

// ReplaceVars replaces any variable that exists in members of Project
func (prj *Project) ReplaceVars(v vars.Variables, r vars.Replacer) {
	v.AddVar("${Name}", prj.Name)
	prj.Platforms.ReplaceVars(v, r)
	v.DelVar("${Name}")
}

var defaultMainSourcePaths = []string{Path("source\\main\\^**\\*.cpp"), Path("source\\main\\^**\\*.c")}
var defaultTestSourcePaths = []string{Path("source\\test\\^**\\*.cpp")}
var defaultMainIncludePaths = []string{Path("source\\main\\include\\^**\\*.h"), Path("source\\main\\include\\^**\\*.hpp"), Path("source\\main\\include\\^**\\*.inl")}
var defaultTestIncludePaths = []string{Path("source\\test\\include\\^**\\*.h"), Path("source\\main\\include\\^**\\*.h")}

// SetupDefaultCppLibProject returns a default C++ project
// Example:
//              SetupDefaultCppLibProject("xbase", "github.com\\jurgen-kluft")
//
func SetupDefaultCppLibProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.GUID = uid.GetGUID(project.Name)
	project.PackageURL = URL
	project.Language = CppLanguageToken
	project.Type = StaticLibrary

	project.SrcFiles = &Files{GlobPaths: defaultMainSourcePaths, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: defaultMainIncludePaths, VirtualPaths: []string{}, Files: []string{}}

	project.Platforms = GetDefaultPlatforms()
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
	project.Language = CppLanguageToken
	project.Type = Executable

	project.SrcFiles = &Files{GlobPaths: defaultTestSourcePaths, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: defaultTestIncludePaths, VirtualPaths: []string{}, Files: []string{}}

	project.Platforms = GetDefaultPlatforms()
	project.Dependencies = []*Project{}

	project.Platforms.AddIncludeDir(Path("source\\test\\include"))
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
	project.Language = CppLanguageToken
	project.Type = Executable

	project.SrcFiles = &Files{GlobPaths: defaultMainSourcePaths, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: defaultMainIncludePaths, VirtualPaths: []string{}, Files: []string{}}

	project.Platforms = GetDefaultPlatforms()
	project.Dependencies = []*Project{}
	return project
}

