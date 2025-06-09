package denv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jurgen-kluft/ccode/dev"
	"github.com/jurgen-kluft/ccode/foundation"
)

// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------

type ClayGenerator struct {
	Workspace     *Workspace
	Verbose       bool
	BuildTarget   dev.BuildTarget
	TargetAbsPath string
}

func NewClayGenerator(ws *Workspace, verbose bool) *ClayGenerator {
	g := &ClayGenerator{
		Workspace:     ws,
		Verbose:       verbose,
		BuildTarget:   ws.BuildTarget,
		TargetAbsPath: ws.GenerateAbsPath,
	}
	return g
}

func (g *ClayGenerator) Generate() error {
	// Current directory
	currentDir, _ := os.Getwd()

	appDir := g.TargetAbsPath
	foundation.MakeDir(appDir)

	foundation.LogPrintf("Generating clay project files in '%s' for target %s", foundation.PathGetRelativeTo(appDir, currentDir), g.BuildTarget)

	out := foundation.NewLineWriter(foundation.IndentModeSpaces)
	g.generateMain(out)
	g.generateProjectFile(out)

	// Write the generated file to the target path
	projectDotGoFilepath := filepath.Join(appDir, "clay.go")
	if err := out.WriteToFile(projectDotGoFilepath); err != nil {
		return foundation.LogErrorf(err, "Error writing file %s: %v", projectDotGoFilepath)
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
		foundation.LogPrintlnf("You can now use the clay command in the build directory")
		foundation.LogPrintlnf("    %s", foundation.PathGetRelativeTo(g.TargetAbsPath, currentDir))
		foundation.LogPrintlnf("Execute 'cd %s' to change to the build directory", foundation.PathGetRelativeTo(g.TargetAbsPath, currentDir))
		foundation.LogPrintlnf("In there, run './clay help' for more information.")
	}
	return nil
}

func (g *ClayGenerator) generateMain(out *foundation.LineWriter) {
	out.WriteLine("// --------------------------------------------------------------------")
	out.WriteLine("// ---------------------- GENERATED -----------------------------------")
	out.WriteLine("// --------------------------------------------------------------------")
	out.WriteLine("package main")
	out.WriteLine()
	out.WriteLine("import (")
	out.WriteILine("", "clay \"github.com/jurgen-kluft/ccode/clay\"")
	out.WriteLine(")")
	out.WriteLine()
	out.WriteLine("func main() {")
	out.WriteILine("", "clay.ClayAppCreateProjectsFunc = CreateProjects")
	if g.BuildTarget.OSAsString() == "arduino" {
		out.WriteILine("", "clay.ClayAppMainArduino()")
	} else {
		out.WriteILine("", "clay.ClayAppMainDesktop()")
	}
	out.WriteLine("}")
}

func (g *ClayGenerator) generateProjectFile(out *foundation.LineWriter) {
	out.WriteLine()
	out.WriteLine("func CreateProjects(arch string) []*clay.Project {")
	out.WriteILine("", "projects := []*clay.Project{}")

	// Here we create a clay.Project per ccode_gen.Project+Config:
	// clay.Project = ucore + debug
	// clay.Project = ucore + release
	// clay.Project = ublinky + debug
	// clay.Project = ublinky + release

	os := g.BuildTarget.OSAsString()

	projectToIndex := map[string]int{}
	for _, prj := range g.Workspace.ProjectList.Values {
		if prj.SupportedTargets.Contains(g.BuildTarget) {

			// Get the version info for this project
			depVersionInfo := foundation.NewGitVersionInfo(prj.ProjectAbsPath)
			prj.Version = depVersionInfo.Commit

			//			projectBaseDir := prj.ProjectAbsPath

			for _, prjCfg := range prj.Resolved.Configs.Values {
				configName := prjCfg.BuildConfig.AsString()

				out.WriteILine("", "{")
				out.WriteILine("+", "// Project Index = "+strconv.Itoa(len(projectToIndex)))
				out.WriteILine("+", "configName := ", `"`, configName, `"`)
				out.WriteILine("+", "projectName := ", `"`, prj.Name, `"`)
				out.WriteILine("+", `projectConfig := clay.NewConfig("`+os+`", arch, configName)`)
				if prj.BuildType.IsExecutable() {
					out.WriteILine("+", "project := clay.NewExecutableProject(projectName, projectConfig)")
				} else {
					out.WriteILine("+", "project := clay.NewLibraryProject(projectName, projectConfig)")
				}
				out.WriteLine()

				if len(prjCfg.IncludeDirs.Values) > 0 {
					out.WriteILine("+", "// Project Include directories")
					for _, inc := range prjCfg.IncludeDirs.Values {
						includePath := filepath.Join(inc.Root, inc.Base, inc.Sub)
						includePath = strings.Replace(includePath, "\\", "/", -1)
						out.WriteILine("+", "project.IncludeDirs.Add(", `"`, includePath, `")`)
					}
					out.WriteLine()
				}

				if len(prjCfg.CppDefines.Values) > 0 {
					out.WriteILine("+", "// Project Define macros")
					for _, def := range prjCfg.CppDefines.Values {
						escapedDef := strings.Replace(def, `"`, `\"`, -1)
						out.WriteILine("+", "project.Defines.Add(", `"`, escapedDef, `")`)
					}
					out.WriteILine("+", "project.Defines.AddMany(projectConfig.GetCppDefines()...)")
					out.WriteLine()
				}

				{
					out.WriteILine("+", "// Project Source files")
					for _, group := range prj.SrcFileGroups {
						for _, src := range group.Values {
							if src.Is_SourceFile() {
								path := filepath.Join(group.Path, src.Path)
								path = strings.Replace(path, "\\", "/", -1)
								out.WriteILine("+", "project.AddSourceFile(", `"`, path, `", "`, filepath.Base(path), `")`)
							}
						}
					}

					out.WriteLine()
				}

				projectToIndex[prj.Name+"-"+configName] = len(projectToIndex)
				out.WriteILine("+", "projects = append(projects, project)")
				out.WriteILine("", "}")
			}
		}
	}

	out.WriteILine("", "// Setup Project Dependencies")
	for _, prj := range g.Workspace.ProjectList.Values {
		if prj.SupportedTargets.Contains(g.BuildTarget) {
			for _, prjCfg := range prj.Resolved.Configs.Values {
				configName := prjCfg.BuildConfig.AsString()

				projectIndex, _ := projectToIndex[prj.Name+"-"+configName]

				if prj.Dependencies.Len() > 0 {
					out.WriteILine("", "{")
					out.WriteILine("+", "project := projects[", strconv.Itoa(projectIndex), "]")

					for _, depProject := range prj.Dependencies.Values {
						depIndex, _ := projectToIndex[depProject.Name+"-"+configName]
						out.WriteILine("+", "project.Dependencies = append(project.Dependencies, projects[", strconv.Itoa(depIndex), "])")
					}
					out.WriteILine("", "}")
				}
			}
		}
	}
	out.WriteILine("", "return projects")
	out.WriteLine("}")
	out.WriteLine()
}

// func (g *ClayGenerator) registerEsp32CoreLibrary() {

// 		// System Library is at ESP_ROOT+'cores/esp32', collect
// 		// all the C and Cpp source files in this directory and create a Library.
// 		sdkRoot := tc.Vars.GetOne("esp.sdk.path")
// 		coreLibPath := filepath.Join(sdkRoot, "cores/esp32/")

// 		coreCppLib := NewLibraryProject("core-cpp-"+targetMcu, p.Config)

// 		// Get all the .cpp files from the core library path
// 		coreCppLib.AddSourceFilesFrom(coreLibPath, OptionAddCppFiles|OptionAddCFiles|OptionAddRecursively)

// }
