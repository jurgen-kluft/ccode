package ccode_gen

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	ccode_utils "github.com/jurgen-kluft/ccode/ccode-utils"
	"github.com/jurgen-kluft/ccode/denv"
)

// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------

type Esp32Generator struct {
	Workspace     *Workspace
	Verbose       bool
	BuildTarget   denv.BuildTarget
	TargetAbsPath string
}

func NewEsp32Generator(ws *Workspace, verbose bool) *Esp32Generator {
	g := &Esp32Generator{
		Workspace:     ws,
		Verbose:       verbose,
		BuildTarget:   denv.BuildTargetArduinoEsp32,
		TargetAbsPath: ws.GenerateAbsPath,
	}
	return g
}

func (g *Esp32Generator) Generate() error {
	appDir := filepath.Join(g.TargetAbsPath, "clay-app")
	ccode_utils.MakeDir(appDir)

	out := ccode_utils.NewLineWriter(ccode_utils.IndentModeSpaces)
	g.generateMain(out)
	appGoFilepath := filepath.Join(appDir, "main.go")
	if err := out.WriteToFile(appGoFilepath); err != nil {
		log.Printf("Error writing file %s: %v", appGoFilepath, err)
	}

	out = ccode_utils.NewLineWriter(ccode_utils.IndentModeSpaces)
	g.generateProjectFile(out)

	// Write the generated file to the target path

	projectDotGoFilepath := filepath.Join(appDir, "project.go")
	if err := out.WriteToFile(projectDotGoFilepath); err != nil {
		log.Printf("Error writing file %s: %v", projectDotGoFilepath, err)
	}

	// Run 'go build -o clay clay' in the build directory to get the clay executable
	if goCmd, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("Go command not found in PATH")
	} else {
		cmd := exec.Command(goCmd, "build", "-o", "clay", appDir)
		cmd.Dir = g.TargetAbsPath
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("Error running go build: %v\nOutput: %s", err, out)
		}
		log.Printf("You can now use the clay command in the build directory: %s", g.TargetAbsPath)
		log.Printf("Run 'clay help' for more information.")
	}
	return nil
}

func (g *Esp32Generator) generateMain(out *ccode_utils.LineWriter) {
	out.WriteLine("package main")
	out.WriteLine()
	out.WriteLine("import (")
	out.WriteILine("", "clay \"github.com/jurgen-kluft/ccode/clay\"")
	out.WriteLine(")")
	out.WriteLine()
	out.WriteLine("func main() {")
	out.WriteILine("", "clay.ClayAppCreateProjectsFunc = CreateProjects")
	out.WriteILine("", "clay.ClayAppMain()")
	out.WriteLine("}")
	out.WriteLine()
}

func (g *Esp32Generator) generateProjectFile(out *ccode_utils.LineWriter) {
	out.WriteLine("// --------------------------------------------------------------------")
	out.WriteLine("// ---------------------- GENERATED -----------------------------------")
	out.WriteLine("// --------------------------------------------------------------------")
	out.WriteLine("package main")
	out.WriteLine()
	out.WriteLine("import \"github.com/jurgen-kluft/ccode/clay\"")
	out.WriteLine()
	out.WriteLine("func CreateProjects(buildPath string) []*clay.Project {")
	out.WriteILine("", "projects := []*clay.Project{}")

	// TODO multiple projects ?

	// Iterate over the 'executable' projects
	// Every project needs to 'add' its libraries
	for _, p := range g.Workspace.ProjectList.Values {
		if p.Type.IsExecutable() && p.SupportedTargets.Contains(g.BuildTarget) {
			out.WriteILine("", "{")
			out.WriteILine("+", "prjName := ", `"`, p.Name, `"`)
			out.WriteILine("+", "prjVersion := ", `"`, p.Version, `"`)
			out.WriteILine("+", "prj := clay.NewProject(prjName, prjVersion, buildPath)")

			out.WriteILine("+", "add__", p.Name, "__library(prj)")

			for _, dep := range p.Dependencies.Values {
				out.WriteILine("+", "add__", dep.Name, "__library(prj)")
			}

			out.WriteILine("+", "projects = append(projects, prj)")
			out.WriteILine("", "}")
		}
	}

	out.WriteILine("", "return projects")
	out.WriteLine("}")
	out.WriteLine()

	// Emit the 'library' projects
	for _, p := range g.Workspace.ProjectList.Values {
		if p.SupportedTargets.Contains(g.BuildTarget) {
			if p.Type.IsStaticLibrary() {
				g.generateLibrary(p, out)
			} else if p.Type.IsExecutable() {
				// An 'executable' project also has source files, so we also emit it as a library
				g.generateLibrary(p, out)
			}
		}
	}
}

func (g *Esp32Generator) generateLibrary(p *Project, out *ccode_utils.LineWriter) {
    units := out

    units.WriteLine("func add__", p.Name, "__library(prj *clay.Project) {")
    name := p.Name
    version := "0.1.0"
    units.WriteILine("", "name := \"", name, "\"")
    units.WriteILine("", "library := clay.NewCppLibrary(name, \"", version, "\", name, name+\".a\")")
    units.WriteLine()

    units.WriteILine("", "// Include directories")
    pcfg, _ := p.Resolved.Configs.Get(ConfigTypeRelease)
    for _, inc := range pcfg.IncludeDirs.Values {
        includePath := filepath.Join(inc.Root, inc.Path)
        includePath = strings.Replace(includePath, "\\", "/", -1)
        units.WriteILine("", "library.IncludeDirs.Add(", `"`, includePath, `", false)`)
    }
    units.WriteLine()

    units.WriteILine("", "// Define macros")
    for _, def := range pcfg.CppDefines.Vars.Values {
        escapedDef := strings.Replace(def, `"`, `\"`, -1)
        units.WriteILine("", "library.Defines.Add(", `"`, escapedDef, `")`)
    }
    units.WriteLine()

    units.WriteILine("", "// Source files")
    for _, src := range p.FileEntries.Values {
        if src.Is_SourceFile() {
            path := filepath.Join(p.ProjectAbsPath, src.Path)
            path = strings.Replace(path, "\\", "/", -1)
            units.WriteILine("", "library.AddSourceFile(", `"`, path, `", "`, filepath.Base(path), `", true)`)
        }
    }
    units.WriteLine()

    units.WriteILine("", "// Add the library to the project")
    units.WriteILine("", "prj.Executable.AddLibrary(library)")
    units.WriteLine( "}")
    units.WriteLine()

    return
}
