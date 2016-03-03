package denv

import (
	"fmt"
	"path"
	"strings"

	"github.com/jurgen-kluft/xcode/glob"
	"github.com/jurgen-kluft/xcode/uid"
	"github.com/jurgen-kluft/xcode/vars"
)

// Config represents a project build configuration, like 'Debug' or 'Release'
type Config struct {
	Name         string
	IncludeDirs  []string
	LibraryDirs  []string
	LibraryFiles []string
	LibraryFile  string
}

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
	Configs      map[string]*Config
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
	for _, config := range prj.Configs {
		if configname == config.Name {
			return true
		}
	}
	return false
}

// DefaultDefines are defines for
var DefaultDefines = []string{
	"TARGET_DEV_DEBUG;_DEBUG;",
	"TARGET_DEV_RELEASE;NDEBUG;",
	"TARGET_TEST_DEBUG;_DEBUG;",
	"TARGET_TEST_RELEASE;NDEBUG;",
}

// SupportedPlatforms returns a list of platforms that are supported by xcode
var SupportedPlatforms = []string{
	"Win32",
	"x64",
}

// DefaultConfigs $(Configuration)_$(Platform)
var DefaultConfigs = []Config{
	{Name: "DevDebugStatic", IncludeDirs: []string{"source\\main\\include"}, LibraryDirs: []string{"target\\$(Configuration)_$(Platform)_$(ToolSet)"}, LibraryFiles: []string{}, LibraryFile: "${Name}_$(Configuration)_$(Platform)_$(ToolSet).lib"},
	{Name: "DevReleaseStatic", IncludeDirs: []string{"source\\main\\include"}, LibraryDirs: []string{"target\\$(Configuration)_$(Platform)_$(ToolSet)"}, LibraryFiles: []string{}, LibraryFile: "${Name}_$(Configuration)_$(Platform)_$(ToolSet).lib"},
	{Name: "TestDebugStatic", IncludeDirs: []string{"source\\main\\include"}, LibraryDirs: []string{"target\\$(Configuration)_$(Platform)_$(ToolSet)"}, LibraryFiles: []string{}, LibraryFile: "${Name}_$(Configuration)_$(Platform)_$(ToolSet).lib"},
	{Name: "TestReleaseStatic", IncludeDirs: []string{"source\\main\\include"}, LibraryDirs: []string{"target\\$(Configuration)_$(Platform)_$(ToolSet)"}, LibraryFiles: []string{}, LibraryFile: "${Name}_$(Configuration)_$(Platform)_$(ToolSet).lib"},
}

// CopyStringArray makes a copy of an array of strings
func CopyStringArray(strarray []string) []string {
	newstrarray := make([]string, len(strarray))
	for i, str := range strarray {
		newstrarray[i] = str
	}
	return newstrarray
}

// CopyConfig makes a deep copy of a Config
func CopyConfig(config Config) *Config {
	newconfig := &Config{Name: config.Name, IncludeDirs: []string{}, LibraryDirs: []string{}, LibraryFiles: []string{}, LibraryFile: ""}
	newconfig.IncludeDirs = CopyStringArray(config.IncludeDirs)
	newconfig.LibraryDirs = CopyStringArray(config.LibraryDirs)
	newconfig.LibraryFiles = CopyStringArray(config.LibraryFiles)
	newconfig.LibraryFile = config.LibraryFile
	return newconfig
}

// ReplaceVars replaces variables that are present in members of the Config
func (c *Config) ReplaceVars(v vars.Variables, r vars.Replacer) {
	c.LibraryFile = v.ReplaceInLine(r, c.LibraryFile)
	v.ReplaceInLines(r, c.IncludeDirs)
	v.ReplaceInLines(r, c.LibraryDirs)
	v.ReplaceInLines(r, c.LibraryFiles)
}

// ReplaceVars replaces any variable that exists in members of Project
func (prj *Project) ReplaceVars(v vars.Variables, r vars.Replacer) {
	v.AddVar("${Name}", prj.Name)
	for _, config := range prj.Configs {
		config.ReplaceVars(v, r)
	}
}

// GetDefaultConfigs returns a map of default configs
func GetDefaultConfigs() map[string]*Config {
	configs := make(map[string]*Config)
	for _, config := range DefaultConfigs {
		configs[config.Name] = CopyConfig(config)
	}
	return configs
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
