package ccode_gen

import (
	_ "embed"
)

// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------

type EspMakeGenerator struct {
	Workspace     *Workspace
	Verbose       bool
	TargetAbsPath string
	Libraries     []*Project
	Product       *Project
}

func NewEspMakeGenerator(ws *Workspace, verbose bool) *EspMakeGenerator {
	g := &EspMakeGenerator{
		Workspace: ws,
		Verbose:   verbose,
	}
	g.TargetAbsPath = ws.GenerateAbsPath

	// Add the libraries
	for _, p := range ws.ProjectList.Values {
		if p.TypeIsLib() || p.TypeIsDll() {
			g.Libraries = append(g.Libraries, p)
		} else if p.TypeIsExe() {
			g.Product = p
		}
	}

	return g
}

func (g *EspMakeGenerator) Generate() error {

	// Use Clay to setup a project and populate it with the libraries
	// and the product.
	// We could 'compile' and copy a Go binary to the build directory
	// which the user could use to 'compile'the project.
	// We can also provide other usefull utilities in that binary, like
	// listing all the available libraries, available boards, possible
	// flash sizes, etc.

	// Copy clay code to "build/clay"
	// Generate a go file which creates the project and populates it with
	// the libraries and source files.

	// Output Example:
	//
	// // !!Project!!
	//
	// func CreateProject(buildPath string) *clay.Project {
	// 	   prjName := "test_project"
	// 	   prjVersion := "0.1.0"
	// 	   prj := clay.NewProject(prjName, prjVersion, buildPath)
	// 	   AddLibraries(prj)
	// 	   return prj
	// }
	//
	// func AddLibraries(prj *clay.Project) {
	// 		name := "test_lib"
	// 		library := clay.NewCppLibrary(name, "0.1.0", name, name+".a")
	//
	// 		// Include directories
	// 		library.IncludeDirs.Add("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/ccode/clay/app/clay/test_lib/include", false)
	//
	// 		// Define macros
	// 		library.Defines.Add("TARGET_DEBUG")
	// 		library.Defines.Add("TARGET_ESP32")
	//
	// 		// Source files of chash
	// 		library.AddSourceFile("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/ccode/clay/app/clay/test_lib/src/test.cpp", "test.cpp", true)
	// 		// etc..
	// 		prj.Executable.AddLibrary(library)
	// }

	return nil
}
