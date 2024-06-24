package denv

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
	Language     string
	Platform     *Platform
	Dependencies []*Project
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

// SetupDefaultCppLibProject returns a default C++ project
// Example:
//
//	SetupDefaultCppLibProject("cbase", "github.com/jurgen-kluft")
func SetupDefaultCppLibProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.PackageURL = Path(URL)
	project.Language = CppLanguageToken
	project.Type = StaticLibrary

	project.Platform = GetDefaultPlatform()
	project.Dependencies = []*Project{}

	return project
}

// SetupDefaultCppTestProject returns a default C++ project
// Example:
//
//	SetupDefaultCppTestProject("cbase", "github.com\\jurgen-kluft")
func SetupDefaultCppTestProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.PackageURL = Path(URL)
	project.Language = CppLanguageToken
	project.Type = Executable

	project.Platform = GetDefaultPlatform()
	project.Dependencies = []*Project{}

	return project
}

// SetupDefaultCppAppProject returns a default C++ project
// Example:
//
//	SetupDefaultCppAppProject("cbase", "github.com\\jurgen-kluft")
func SetupDefaultCppAppProject(name string, URL string) *Project {
	project := &Project{Name: name}
	project.PackageURL = Path(URL)
	project.Language = CppLanguageToken
	project.Type = Executable

	project.Platform = GetDefaultPlatform()
	project.Dependencies = []*Project{}

	return project
}
