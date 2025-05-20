// --------------------------------------------------------------------
// ---------------------- GENERATED -----------------------------------
// --------------------------------------------------------------------
package main

import "github.com/jurgen-kluft/ccode/clay"

func CreateProjects(buildPath string) []*clay.Project {
	projects := []*clay.Project{}

	prjName := "test_project"
	prjConfig := "release"
	prj := clay.NewProject(prjName, prjConfig, buildPath)
	clay.AddBuildInfoAsCppLibrary(prj)
	AddLibraries(prj)

	projects = append(projects, prj)

	return projects
}

func AddLibraries(prj *clay.Project) {
	{
		name := "test_lib"
		library := clay.NewCppLibrary(name, "0.1.0", name, name+".a")

		// Include directories
		library.IncludeDirs.Add("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/ccode/clay/app/clay/test_lib/include", false)
		// etc..

		// Define macros
		library.Defines.Add("TARGET_DEBUG")
		library.Defines.Add("TARGET_ESP32")
		// etc..

		// Source files
		library.AddSourceFile("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/ccode/clay/app/clay/test_lib/src/test.cpp", "test.cpp", true)
		// etc..

		prj.Executable.AddLibrary(library)
	}

}
