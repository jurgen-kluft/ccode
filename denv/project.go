package denv

import (
	"fmt"
	"os"
	"path"

	"github.com/jurgen-kluft/xcode/glob"
	"github.com/jurgen-kluft/xcode/uid"
)

type Config struct {
	Name        string
	IncludeDirs []string
	LibraryDirs []string
	LinkWith    []string
}

type Files struct {
	GlobPaths    []string
	VirtualPaths []string
	Files        []string
}

type Project struct {
	Name         string
	Author       string
	GUID         string
	Path         string
	Language     string
	Platforms    []string
	HdrFiles     *Files
	SrcFiles     *Files
	Configs      map[string]Config
	Dependencies []*Project
}

func GetDefaultPlatforms() []string {
	return []string{"Win32", "x64"}
}

var DefaultDefines = []string{
	"TARGET_DEV_DEBUG;_DEBUG;",
	"TARGET_DEV_RELEASE;NDEBUG;",
	"TARGET_TEST_DEBUG;_DEBUG;",
	"TARGET_TEST_RELEASE;NDEBUG;",
}

var SupportedPlatforms = []string{
	"Win32",
	"x64",
}

// $(Configuration)_$(Platform)
var DefaultConfigs = []Config{
	{Name: "DevDebugStatic", IncludeDirs: []string{"source\\main\\include"}, LibraryDirs: []string{"target\\$(Configuration)_$(Platform)_$(ToolSet)"}, LinkWith: []string{"$(Name)_$(Configuration)_$(Platform)_$(ToolSet).lib"}},
	{Name: "DevReleaseStatic", IncludeDirs: []string{"source\\main\\include"}, LibraryDirs: []string{"target\\$(Configuration)_$(Platform)_$(ToolSet)"}, LinkWith: []string{"$(Name)_$(Configuration)_$(Platform)_$(ToolSet).lib"}},
	{Name: "TestDebugStatic", IncludeDirs: []string{"source\\main\\include"}, LibraryDirs: []string{"target\\$(Configuration)_$(Platform)_$(ToolSet)"}, LinkWith: []string{"$(Name)_$(Configuration)_$(Platform)_$(ToolSet).lib"}},
	{Name: "TestReleaseStatic", IncludeDirs: []string{"source\\main\\include"}, LibraryDirs: []string{"target\\$(Configuration)_$(Platform)_$(ToolSet)"}, LinkWith: []string{"$(Name)_$(Configuration)_$(Platform)_$(ToolSet).lib"}},
}

//var DefaultConfigs = []string{
//	"DevDebugStatic",
//	"DevReleaseStatic",
//	"TestDebugStatic",
//	"TestReleaseStatic",
//}

func GetDefaultConfigs() map[string]Config {
	configs := make(map[string]Config)
	for _, config := range DefaultConfigs {
		configs[config.Name] = config
	}
	return configs
}

// SetupDefaultCppProject returns a default C++ project
// Example:
//              SetupDefaultCppProject("xbase", "github.com\\jurgen-kluft")
//
func SetupDefaultCppProject(name string, url string) *Project {
	gopath := os.Getenv("GOPATH")

	project := &Project{Name: name}
	project.GUID = uid.GetGUID(project.Name)
	project.Path = path.Join(gopath, "src", url, project.Name)
	project.Language = "C++"

	fmt.Println(project.Path)

	project.SrcFiles = &Files{GlobPaths: []string{"source\\main\\^cpp\\**\\*.cpp"}}
	project.SrcFiles.GlobFiles(project.Path)

	project.HdrFiles = &Files{GlobPaths: []string{"source\\main\\include\\^**\\*.h"}}
	project.HdrFiles.GlobFiles(project.Path)

	project.Platforms = GetDefaultPlatforms()
	project.Configs = GetDefaultConfigs()
	project.Dependencies = []*Project{}
	return project
}

func (files *Files) GlobFiles(path string) {
	// Glob all the on-disk files
	files.Files, _ = glob.GlobFiles(path, files.GlobPaths)

	// Generate the virtual files

}
