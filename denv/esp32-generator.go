package denv

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	dev "github.com/jurgen-kluft/ccode/dev"
	utils "github.com/jurgen-kluft/ccode/utils"
)

// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------

type Esp32Generator struct {
	Workspace     *Workspace
	Verbose       bool
	BuildTarget   dev.BuildTarget
	TargetAbsPath string
}

func NewEsp32Generator(ws *Workspace, verbose bool) *Esp32Generator {
	g := &Esp32Generator{
		Workspace:     ws,
		Verbose:       verbose,
		BuildTarget:   ws.BuildTarget,
		TargetAbsPath: ws.GenerateAbsPath,
	}
	return g
}

func (g *Esp32Generator) Generate() error {
	appDir := filepath.Join(g.TargetAbsPath, "clay-app")
	utils.MakeDir(appDir)

	log.Printf("Generating ESP32 project files in %s for target %s", appDir, g.BuildTarget)

	out := utils.NewLineWriter(utils.IndentModeSpaces)
	g.generateMain(out)
	appGoFilepath := filepath.Join(appDir, "main.go")
	if err := out.WriteToFile(appGoFilepath); err != nil {
		log.Printf("Error writing file %s: %v", appGoFilepath, err)
	}

	out = utils.NewLineWriter(utils.IndentModeSpaces)
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

func (g *Esp32Generator) generateMain(out *utils.LineWriter) {
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

func (g *Esp32Generator) generateProjectFile(out *utils.LineWriter) {
	out.WriteLine("// --------------------------------------------------------------------")
	out.WriteLine("// ---------------------- GENERATED -----------------------------------")
	out.WriteLine("// --------------------------------------------------------------------")
	out.WriteLine("package main")
	out.WriteLine()
	out.WriteLine("import \"github.com/jurgen-kluft/ccode/clay\"")
	out.WriteLine()
	out.WriteLine("func CreateProjects(buildPath string) []*clay.Project {")
	out.WriteILine("", "projects := []*clay.Project{}")

	cfgs := make([]*Config, 0, 16)
	deps := make([]*Project, 0, 16)

	// Here we create a clay.Project per ccode_gen.Project+Config:
	// clay.Project = ucore + debug
	// clay.Project = ucore + release
	// clay.Project = ublinky + debug
	// clay.Project = ublinky + release

	arch := g.BuildTarget.ArchAsString()

	for _, prj := range g.Workspace.ProjectList.Values {
		if prj.BuildType.IsExecutable() && prj.SupportedTargets.Contains(g.BuildTarget) {

			// Get the version info for this project
			depVersionInfo := utils.NewGitVersionInfo(prj.ProjectAbsPath)
			prj.Version = depVersionInfo.Commit

			for _, prjCfg := range prj.Resolved.Configs.Values {
				cfgName := prjCfg.BuildConfig.ConfigString()

				out.WriteILine("", "{")
				out.WriteILine("+", "prjName := ", `"`, prj.Name, `"`)
				out.WriteILine("+", "prjConfig := clay.NewConfig(\"arduino\", \"", arch, `", "`, prjCfg.BuildConfig.Build(), `", "`, prjCfg.BuildConfig.Variant(), `")`)
				out.WriteILine("+", "prj := clay.NewProject(prjName, prjConfig, buildPath)")

				out.WriteILine("+", "add_", prj.Name, "_", cfgName, "_library(prj)")
				cfgs = append(cfgs, prjCfg)
				deps = append(deps, prj)
				for _, dep := range prj.Dependencies.Values {
					depCfg := dep.Resolved.FindConfig(prjCfg.BuildConfig)
					cfgs = append(cfgs, depCfg)
					deps = append(deps, dep)
					out.WriteILine("+", "add_", dep.Name, "_", cfgName, "_library(prj)")
				}

				out.WriteILine("+", "projects = append(projects, prj)")
				out.WriteILine("", "}")
			}
		}
	}

	out.WriteILine("", "return projects")
	out.WriteLine("}")
	out.WriteLine()

	// Emit the 'library' projects, avoid duplicates
	duplicateTracking := make(map[string]bool)
	for i, p := range deps {
		if _, ok := duplicateTracking[p.Name+":"+cfgs[i].BuildConfig.ConfigString()]; !ok {
			g.generateLibrary(p, cfgs[i], out)
			duplicateTracking[p.Name+":"+cfgs[i].BuildConfig.ConfigString()] = true
			continue
		}
	}
}

func (g *Esp32Generator) generateLibrary(p *Project, cfg *Config, out *utils.LineWriter) {
	units := out

	{
		units.WriteLine("func add_", p.Name, "_", cfg.BuildConfig.ConfigString(), "_library(prj *clay.Project) {")
		name := p.Name
		units.WriteILine("", "name := \"", name, "\"")
		units.WriteILine("", "library := clay.NewLibrary(name, prj.Config)")
		units.WriteLine()

		units.WriteILine("", "// Include directories")
		for _, inc := range cfg.IncludeDirs.Values {
			includePath := filepath.Join(inc.Root, inc.Path)
			includePath = strings.Replace(includePath, "\\", "/", -1)
			units.WriteILine("", "library.IncludeDirs.Add(", `"`, includePath, `")`)
		}
		units.WriteLine()

		units.WriteILine("", "// Define macros")
		for _, def := range cfg.CppDefines.Values {
			escapedDef := strings.Replace(def, `"`, `\"`, -1)
			units.WriteILine("", "library.Defines.Add(", `"`, escapedDef, `")`)
		}
		units.WriteLine()

		units.WriteILine("", "// Source files")
		for _, src := range p.SrcFiles.Values {
			if src.Is_SourceFile() {
				path := filepath.Join(p.ProjectAbsPath, src.Path)
				path = strings.Replace(path, "\\", "/", -1)
				units.WriteILine("", "library.AddSourceFile(", `"`, path, `", "`, filepath.Base(path), `")`)
			}
		}
		for _, externalSource := range p.ExternalSrcFiles {
			for _, src := range externalSource.Values {
				if src.Is_SourceFile() {
					path := src.Path
					path = strings.Replace(path, "\\", "/", -1)
					units.WriteILine("", "library.AddSourceFile(", `"`, filepath.Join(externalSource.Path, path), `", "`, path, `")`)
				}
			}
		}
		units.WriteLine()

		units.WriteILine("", "// Add the library to the project")
		units.WriteILine("", "prj.Executable.AddLibrary(library)")
		units.WriteLine("}")
		units.WriteLine()
	}

	return
}
