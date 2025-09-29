// --------------------------------------------------------------------
// ---------------------- GENERATED -----------------------------------
// --------------------------------------------------------------------
package main

import (
	"fmt"
	"runtime"

	"github.com/jurgen-kluft/ccode/clay"
	"github.com/jurgen-kluft/ccode/dev"
)

func CreateProjects() []*clay.Project {
	arch := runtime.GOARCH
	projectName := "test_project"
	buildTargetStr := fmt.Sprintf("%s(%s)", "macos", arch)
	projectTarget := dev.BuildTargetFromString(buildTargetStr)
	projectConfig := clay.NewConfig(projectTarget, "debug-dev")
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
