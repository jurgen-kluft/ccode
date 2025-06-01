package denv

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	dev "github.com/jurgen-kluft/ccode/dev"
	utils "github.com/jurgen-kluft/ccode/utils"
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

	appDir := filepath.Join(g.TargetAbsPath, "clay-app")
	utils.MakeDir(appDir)

	log.Printf("Generating clay project files in '%s' for target %s", utils.PathGetRelativeTo(appDir, currentDir), g.BuildTarget)

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
		log.Printf("You can now use the clay command in the build directory")
		log.Printf("    %s", utils.PathGetRelativeTo(g.TargetAbsPath, currentDir))
		log.Printf("Execute 'cd %s' to change to the build directory", utils.PathGetRelativeTo(g.TargetAbsPath, currentDir))
		log.Printf("In there, run './clay help' for more information.")
	}
	return nil
}

func (g *ClayGenerator) generateMain(out *utils.LineWriter) {
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
	out.WriteLine()
}

func (g *ClayGenerator) generateProjectFile(out *utils.LineWriter) {
	out.WriteLine("// --------------------------------------------------------------------")
	out.WriteLine("// ---------------------- GENERATED -----------------------------------")
	out.WriteLine("// --------------------------------------------------------------------")
	out.WriteLine("package main")
	out.WriteLine()
	out.WriteLine("import \"github.com/jurgen-kluft/ccode/clay\"")
	out.WriteLine()
	out.WriteLine("func CreateProjects(arch, buildPath string) []*clay.Project {")
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
			depVersionInfo := utils.NewGitVersionInfo(prj.ProjectAbsPath)
			prj.Version = depVersionInfo.Commit

			//			projectBaseDir := prj.ProjectAbsPath

			for _, prjCfg := range prj.Resolved.Configs.Values {
				configName := prjCfg.BuildConfig.AsString()

				out.WriteILine("", "{")
				out.WriteILine("+", "// Project Index = "+strconv.Itoa(len(projectToIndex)))
				out.WriteILine("+", "configName := ", `"`, configName, `"`)
				out.WriteILine("+", "projectName := ", `"`, prj.Name, `"`)
				out.WriteILine("+", `projectConfig := clay.NewConfig("`+os+`", arch, configName)`)
				//				out.WriteILine("+", `projectBaseDir := "`, projectBaseDir, `"`)
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
