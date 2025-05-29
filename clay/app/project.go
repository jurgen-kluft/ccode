// --------------------------------------------------------------------
// ---------------------- GENERATED -----------------------------------
// --------------------------------------------------------------------
package main

import "github.com/jurgen-kluft/ccode/clay"

func CreateProjects(buildPath string) []*clay.Project {
	projects := []*clay.Project{}

	prjName := "test_project"
	prjConfig := clay.NewConfig("macos", "arm64", "debug", "dev")
	prj := clay.NewProject(prjName, prjConfig, buildPath)
	clay.AddBuildInfoAsCppLibrary(prj)
	AddLibraries(prj)

	projects = append(projects, prj)

	return projects
}

func AddLibraries(prj *clay.Project) {
	{
		name := "test_lib"
		library := clay.NewLibrary(name, prj.Config)

		// Include directories
		library.IncludeDirs.Add("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/ccode/clay/app/clay/test_lib/include")
		// etc..

		// Define macros
		library.Defines.Add("TARGET_DEBUG")
		library.Defines.Add("TARGET_ESP32")
		// etc..

		// Source files
		library.AddSourceFile("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/ccode/clay/app/clay/test_lib/src/test.cpp", "test.cpp")
		// etc..

		prj.Executable.AddLibrary(library)
	}

}
