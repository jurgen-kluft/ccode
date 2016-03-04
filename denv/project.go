package denv

import (
	"fmt"
	"path"
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
	StaticLibrary ProjectType = iota // .lib, .a
	// DynamicLibrary is a library that can be dynamically linked with, like a .DLL
	DynamicLibrary ProjectType = iota // .dll
	// Executable is an application that can be run
	Executable ProjectType = iota // .exe
)

// Project is a structure that holds all the information that defines a project in an IDE
type Project struct {
	Name         string
	Type         ProjectType
	Author       string
	GUID         string
	Path         string
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
}

// SetupDefaultCppProject returns a default C++ project
// Example:
//              SetupDefaultCppProject("xbase", "github.com\\jurgen-kluft")
//
func SetupDefaultCppProject(name string, url string) *Project {
	project := &Project{Name: name}
	project.GUID = uid.GetGUID(project.Name)
	project.Path = path.Join(url, project.Name)
	project.Language = "C++"
	project.Type = StaticLibrary

	fmt.Println(project.Path)

	project.SrcFiles = &Files{GlobPaths: []string{"source\\main\\^**\\*.cpp"}, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: []string{"source\\main\\include\\^**\\*.h"}}

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
		ppath := path.Join(dirpath, pp[0])
		f.Files, _ = glob.GlobFiles(ppath, pp[1])
	}

	// Generate the virtual files

}
