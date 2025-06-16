// --------------------------------------------------------------------
// ---------------------- GENERATED -----------------------------------
// --------------------------------------------------------------------
package main

import (
	"runtime"

	"github.com/jurgen-kluft/ccode/clay"
)

func CreateProjects() []*clay.Project {
	arch := runtime.GOARCH
	projectName := "test_project"
	projectConfig := clay.NewConfig("macos", arch, "debug-dev")
	project := clay.NewExecutableProject(projectName, projectConfig)
	clay.AddBuildInfoAsCppLibrary(project, projectConfig)
	AddLibraries(project)
	return []*clay.Project{project}
}

func AddLibraries(prj *clay.Project) {
	{
		name := "test_lib"
		library := clay.NewLibraryProject(name, prj.Config)

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

		prj.AddLibrary(library)
	}

}
