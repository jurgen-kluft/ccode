package denv

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/glob"
	"github.com/jurgen-kluft/ccode/items"
	"github.com/jurgen-kluft/ccode/uid"
	"github.com/jurgen-kluft/ccode/vars"
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
// Also it will ignore files that contain certain incoming patterns
func (f *Files) GlobFiles(dirpath string, file_patterns_to_ignore []string) {
	// Glob all the on-disk files
	for _, g := range f.GlobPaths {
		pp := strings.Split(g, "^")
		ppath := filepath.Join(dirpath, pp[0])

		globbedfiles, _ := glob.GlobFiles(ppath, pp[1])
		for _, file := range globbedfiles {
			globbedfile := filepath.Join(pp[0], file)
			ignored := 0
			for _, ignore := range file_patterns_to_ignore {
				if strings.Contains(globbedfile, ignore) {
					ignored++
				}
			}
			if ignored == 0 {
				f.Files = append(f.Files, globbedfile)
			}
		}
	}
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

type CustomFiles struct {
	Type  string // e.g. "ResourceCompile"
	Files *Files
}

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
	Platform     *Platform
	SrcPath      string
	HdrFiles     *Files
	SrcFiles     *Files
	LibraryFiles items.List
	CustomFiles  []*CustomFiles
	Dependencies []*Project
	Vars         vars.Variables
}

// HasPlatform returns true if the project is configured for that platform
func (prj *Project) HasPlatform(platformname string) bool {
	return prj.Platform.Name == platformname
}

// HasConfig will return true if platform @platformname has a configuration with name @configname
func (prj *Project) HasConfig(platformname, configname string) bool {
	if prj.Platform.Name == platformname {
		if prj.Platform.HasConfig(configname) == false {
			return false
		}
	}
	return true
}

// GetConfig will return the configuration of platform @platformname with name @configname
func (prj *Project) GetConfig(configname string) *Config {
	return prj.Platform.GetConfig(configname)
}

// AddDefine adds a define
func (prj *Project) AddDefine(define string) {
	prj.Platform.AddDefine(define)
}

// AddDefine adds a library or libraries
func (prj *Project) AddLibrary(library string) {
	prj.LibraryFiles = prj.LibraryFiles.Add(library)
}

// AddVar adds a variable to this project
func (prj *Project) AddVar(name, value string) {
	prj.Vars.AddVar(name, value)
}

// MergeVars merges  any variable that exists in objects of Project
func (prj *Project) MergeVars(v vars.Variables) {

	// Merge in the project level variables
	prjmerger := func(key, value string, vv vars.Variables) {
		vv.AddVar(prj.Name+":"+key, value)
	}
	vars.MergeVars(v, prj.Vars, prjmerger)

	// Merge in the project\platform\config variables
	for _, config := range prj.Platform.Configs {
		pcmerger := func(key, value string, vv vars.Variables) {
			vv.AddVar(fmt.Sprintf("%s:%s[%s][%s]", prj.Name, key, prj.Platform.Name, config.Name), value)
		}
		vars.MergeVars(v, config.Vars, pcmerger)
	}
}

// ReplaceVars replaces any variable that exists in members of Project
func (prj *Project) ReplaceVars(v vars.Variables, r vars.Replacer) {
	v.AddVar("${Name}", prj.Name)
	prj.Platform.ReplaceVars(v, r)
	v.DelVar("${Name}")
}

var defaultMainSourcePaths = []string{Path("source\\main\\^**\\*.cpp"), Path("source\\main\\^**\\*.c")}
var defaultTestSourcePaths = []string{Path("source\\test\\^**\\*.cpp"), Path("source\\main\\^**\\*.c")}
var defaultMainIncludePaths = []string{Path("source\\main\\include\\^**\\*.h"), Path("source\\main\\include\\^**\\*.hpp"), Path("source\\main\\include\\^**\\*.inl")}
var defaultTestIncludePaths = []string{Path("source\\test\\include\\^**\\*.h"), Path("source\\main\\include\\^**\\*.h")}

var defaultCocoaMainSourcePaths = []string{Path("source\\main\\^**\\*.m"), Path("source\\main\\^**\\*.mm")}

// SetupDefaultCppLibProject returns a default C++ project
// Example:
//              SetupDefaultCppLibProject("cbase", "github.com\\jurgen-kluft")
//
func SetupDefaultCppLibProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.GUID = uid.GetGUID(project.Name)
	project.PackageURL = URL
	project.Language = CppLanguageToken
	project.Type = StaticLibrary

	project.SrcPath = Path("source\\main\\cpp")
	project.SrcFiles = &Files{GlobPaths: defaultMainSourcePaths, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: defaultMainIncludePaths, VirtualPaths: []string{}, Files: []string{}}
	project.LibraryFiles = items.NewList("", ";", "")

	if OS == "darwin" {
		project.SrcFiles.GlobPaths = append(project.SrcFiles.GlobPaths, defaultCocoaMainSourcePaths...)
	}

	project.Platform = GetDefaultPlatform()
	project.Dependencies = []*Project{}
	project.Vars = vars.NewVars()

	project.AddVar("EXCEPTIONS", "false")
	project.AddVar("COMPILE_AS", "CompileAsCpp")

	return project
}

// SetupDefaultCppTestProject returns a default C++ project
// Example:
//              SetupDefaultCppTestProject("cbase", "github.com\\jurgen-kluft")
//
func SetupDefaultCppTestProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.GUID = uid.GetGUID(project.Name)
	project.PackageURL = URL
	project.Language = CppLanguageToken
	project.Type = Executable

	project.SrcPath = Path("source\\test\\cpp")
	project.SrcFiles = &Files{GlobPaths: defaultTestSourcePaths, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: defaultTestIncludePaths, VirtualPaths: []string{}, Files: []string{}}
	project.LibraryFiles = items.NewList("", ";", "")

	if OS == "darwin" {
		project.SrcFiles.GlobPaths = append(project.SrcFiles.GlobPaths, defaultCocoaMainSourcePaths...)
	}

	project.Platform = GetDefaultPlatform()
	project.Dependencies = []*Project{}
	project.Vars = vars.NewVars()

	project.AddVar("EXCEPTIONS", "Sync")
	project.AddVar("COMPILE_AS", "CompileAsCpp")

	project.Platform.AddIncludeDir(Path("source\\test\\include"))
	return project
}

// SetupDefaultCppAppProject returns a default C++ project
// Example:
//              SetupDefaultCppAppProject("cbase", "github.com\\jurgen-kluft")
//
func SetupDefaultCppAppProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.GUID = uid.GetGUID(project.Name)
	project.PackageURL = URL
	project.Language = CppLanguageToken
	project.Type = Executable

	project.SrcPath = Path("source\\main\\cpp")
	project.SrcFiles = &Files{GlobPaths: defaultMainSourcePaths, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: defaultMainIncludePaths, VirtualPaths: []string{}, Files: []string{}}
	project.LibraryFiles = items.NewList("", ";", "")

	if OS == "darwin" {
		project.SrcFiles.GlobPaths = append(project.SrcFiles.GlobPaths, defaultCocoaMainSourcePaths...)
	}

	project.Platform = GetDefaultPlatform()
	project.Dependencies = []*Project{}
	project.Vars = vars.NewVars()

	project.AddVar("EXCEPTIONS", "false")
	project.AddVar("COMPILE_AS", "CompileAsCpp")

	return project
}
