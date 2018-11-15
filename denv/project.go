package denv

import (
	"fmt"
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
	Platforms    PlatformSet
	SrcPath      string
	HdrFiles     *Files
	SrcFiles     *Files
	CustomFiles  []*CustomFiles
	Dependencies []*Project
	Vars         vars.Variables
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
	for _, platform := range prj.Platforms {
		for _, config := range platform.Configs {
			pcmerger := func(key, value string, vv vars.Variables) {
				vv.AddVar(fmt.Sprintf("%s:%s[%s][%s]", prj.Name, key, platform.Name, config.Name), value)
			}
			vars.MergeVars(v, config.Vars, pcmerger)
		}
	}
}

// ReplaceVars replaces any variable that exists in members of Project
func (prj *Project) ReplaceVars(v vars.Variables, r vars.Replacer) {
	v.AddVar("${Name}", prj.Name)
	prj.Platforms.ReplaceVars(v, r)
	v.DelVar("${Name}")
}

var defaultMainSourcePaths []string
var defaultTestSourcePaths []string
var defaultMainIncludePaths []string
var defaultTestIncludePaths []string

func initDefaultPaths() {
	defaultMainSourcePaths = []string{Path("source\\main\\^**\\*.cpp"), Path("source\\main\\^**\\*.c")}
	defaultTestSourcePaths = []string{Path("source\\test\\^**\\*.cpp")}
	defaultMainIncludePaths = []string{Path("source\\main\\include\\^**\\*.h"), Path("source\\main\\include\\^**\\*.hpp"), Path("source\\main\\include\\^**\\*.inl")}
	defaultTestIncludePaths = []string{Path("source\\test\\include\\^**\\*.h"), Path("source\\main\\include\\^**\\*.h")}
}

// SetupDefaultCppLibProject returns a default C++ project
// Example:
//              SetupDefaultCppLibProject("xbase", "github.com\\jurgen-kluft")
//
func SetupDefaultCppLibProject(name string, URL string) *Project {
	initDefaultPaths()
	project := &Project{Name: name}
	project.GUID = uid.GetGUID(project.Name)
	project.PackageURL = URL
	project.Language = CppLanguageToken
	project.Type = StaticLibrary

	project.SrcPath = Path("source\\main\\cpp")
	project.SrcFiles = &Files{GlobPaths: defaultMainSourcePaths, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: defaultMainIncludePaths, VirtualPaths: []string{}, Files: []string{}}

	project.Platforms = GetDefaultPlatforms()
	project.Dependencies = []*Project{}
	project.Vars = vars.NewVars()

	project.AddVar("EXCEPTIONS", "false")
	project.AddVar("COMPILE_AS", "CompileAsCpp")

	return project
}

// SetupDefaultCppTestProject returns a default C++ project
// Example:
//              SetupDefaultCppTestProject("xbase", "github.com\\jurgen-kluft")
//
func SetupDefaultCppTestProject(name string, URL string) *Project {
	initDefaultPaths()
	project := &Project{Name: name}
	project.GUID = uid.GetGUID(project.Name)
	project.PackageURL = URL
	project.Language = CppLanguageToken
	project.Type = Executable

	project.SrcPath = Path("source\\test\\cpp")
	project.SrcFiles = &Files{GlobPaths: defaultTestSourcePaths, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: defaultTestIncludePaths, VirtualPaths: []string{}, Files: []string{}}

	project.Platforms = GetDefaultPlatforms()
	project.Dependencies = []*Project{}
	project.Vars = vars.NewVars()

	project.AddVar("EXCEPTIONS", "Sync")
	project.AddVar("COMPILE_AS", "CompileAsCpp")

	project.Platforms.AddIncludeDir(Path("source\\test\\include"))
	return project
}

// SetupDefaultCppAppProject returns a default C++ project
// Example:
//              SetupDefaultCppAppProject("xbase", "github.com\\jurgen-kluft")
//
func SetupDefaultCppAppProject(name string, URL string) *Project {
	initDefaultPaths()
	project := &Project{Name: name}
	project.GUID = uid.GetGUID(project.Name)
	project.PackageURL = URL
	project.Language = CppLanguageToken
	project.Type = Executable

	project.SrcPath = Path("source\\main\\cpp")
	project.SrcFiles = &Files{GlobPaths: defaultMainSourcePaths, VirtualPaths: []string{}, Files: []string{}}
	project.HdrFiles = &Files{GlobPaths: defaultMainIncludePaths, VirtualPaths: []string{}, Files: []string{}}

	project.Platforms = GetDefaultPlatforms()
	project.Dependencies = []*Project{}
	project.Vars = vars.NewVars()

	project.AddVar("EXCEPTIONS", "false")
	project.AddVar("COMPILE_AS", "CompileAsCpp")

	return project
}
