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
	// var Project *clay.Project = &clay.Project{
	//     Name:       "ccode_gen",
	//     Version:    "0.1.0",
	//     BuildPath:  BuildPath,
	//     Executable: clay.NewExecutable("ccode_gen", "0.1.0", BuildPath),
	// }

	return nil
}
