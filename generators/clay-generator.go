package ide_generators

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
	"github.com/jurgen-kluft/ccode/denv"
)

// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------

type ClayGenerator struct {
	Workspace     *Workspace
	Verbose       bool
	BuildTarget   denv.BuildTarget
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
	corepkg.DirMake(appDir)

	corepkg.LogInfof("Generating clay project files in '%s' for target %s", corepkg.PathGetRelativeTo(appDir, currentDir), g.BuildTarget)

	out := corepkg.NewLineWriter(corepkg.IndentModeSpaces)
	g.generateMain(out)
	g.generateProjectFile(out)

	// Write the generated file to the target path
	projectDotGoFilepath := filepath.Join(appDir, "clay.go")
	if err := out.WriteToFile(projectDotGoFilepath); err != nil {
		return corepkg.LogErrorf(err, "Error writing file %s: %v", projectDotGoFilepath)
	}

	// Run 'go build -o clay clay' in the build directory to get the clay executable
	if goCmd, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("Go command not found in PATH")
	} else {
		execName := "clay"
		if runtime.GOOS == "windows" {
			// On Windows, we need to use 'go build -o clay.exe clay'
			execName += ".exe"
		}
		cmd := exec.Command(goCmd, "build", "-o", execName, appDir)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "GO111MODULE=off")
		cmd.Dir = g.TargetAbsPath
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("Error running go build: %v\nOutput: %s", err, out)
		}
		corepkg.LogInfo("You can now use the clay command in the build directory")
		corepkg.LogInfo("    ", corepkg.PathGetRelativeTo(g.TargetAbsPath, currentDir))
		corepkg.LogInfof("Execute 'cd %s' to change to the build directory", corepkg.PathGetRelativeTo(g.TargetAbsPath, currentDir))
		corepkg.LogInfo("In there, run './clay help' for more information.")
	}
	return nil
}

func (g *ClayGenerator) generateMain(out *corepkg.LineWriter) {
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
	if g.BuildTarget.Os().String() == "arduino" {
		out.WriteILine("", "clay.ClayAppMainArduino()")
	} else {
		out.WriteILine("", "clay.ClayAppMainDesktop()")
	}
	out.WriteLine("}")
}

func (g *ClayGenerator) generateProjectFile(out *corepkg.LineWriter) {
	count := 0
	for _, prj := range g.Workspace.ProjectList.Values {
		if prj.BuildTargets.HasOverlap(g.BuildTarget) {
			count += len(prj.Resolved.Configs.Values)
		}
	}

	projectIndexToId := make([]string, count)
	projectToIndex := make(map[string]int, count)

	index := 0
	for _, prj := range g.Workspace.ProjectList.Values {
		if prj.BuildTargets.HasOverlap(g.BuildTarget) {
			for _, prjCfg := range prj.Resolved.Configs.Values {
				configName := prjCfg.BuildConfig.String()
				projectId := strings.ReplaceAll(prj.Name+"_"+configName, "-", "_")
				projectIndexToId[index] = projectId
				projectToIndex[projectId] = index
				index++
			}
		}
	}

	/*
		TODO
			We write every project now with supported targets, but we already combine and list
			all include files that come from 'inherited' projects. What if instead we list
			the inherited projects as dependencies and list our own include dirs.
			For the source files, we also don't want to already glob them and list them here,
			instead based on the requested platform and config we should glob them at build time.
	*/

	out.WriteLine()
	out.WriteLine("// Setup Project Identifiers")
	out.WriteLine("const (")
	for i, projectId := range projectIndexToId {
		out.WriteILine("", projectId, "_id int = ", strconv.Itoa(i))
	}
	out.WriteLine(")")

	out.WriteLine()
	out.WriteLine("func CreateProjects() []*clay.Project {")
	out.WriteILine("", "projects := make([]*clay.Project, ", strconv.Itoa(count), ")")

	index = 0
	for _, prj := range g.Workspace.ProjectList.Values {
		if prj.BuildTargets.HasOverlap(g.BuildTarget) {

			//			projectBaseDir := prj.ProjectAbsPath

			for _, prjCfg := range prj.Resolved.Configs.Values {
				configName := prjCfg.BuildConfig.String()

				out.WriteILine("", "{")
				out.WriteILine("+", "configName := ", `"`, configName, `"`)
				out.WriteILine("+", "projectName := ", `"`, prj.Name, `"`)
				out.WriteILine("+", "supportedTargets := clay.BuildTargetFromString(", `"`, prj.BuildTargets.String(), `")`)
				out.WriteILine("+", `projectConfig := clay.NewConfig(supportedTargets, configName)`)
				if prj.BuildType.IsExecutable() {
					out.WriteILine("+", "project := clay.NewExecutableProject(projectName, projectConfig)")
				} else {
					out.WriteILine("+", "project := clay.NewLibraryProject(projectName, projectConfig)")
				}
				out.WriteLine()

				numIncludes := len(prjCfg.IncludeDirs.Values)
				if numIncludes > 0 {
					out.WriteILine("+", "// Project Include directories")
					out.WriteILine("+", "project.IncludeDirs = clay.NewIncludeMap(", strconv.Itoa(numIncludes), ")")
					for _, inc := range prjCfg.IncludeDirs.Values {
						includePath := filepath.Join(inc.Root, inc.Base, inc.Sub)
						includePath = strings.Replace(includePath, "\\", "/", -1)
						out.WriteILine("+", "project.IncludeDirs.Add(", `"`, includePath, `")`)
					}
				} else {
					out.WriteILine("+", "project.IncludeDirs = clay.NewIncludeMap(0)")
				}
				out.WriteLine()

				numDefines := len(prjCfg.CppDefines.Values)
				out.WriteILine("+", "// Project Define macros")
				out.WriteILine("+", "configDefines := projectConfig.GetCppDefines()")
				out.WriteILine("+", "project.Defines = clay.NewDefineMap(", strconv.Itoa(numDefines), " + len(configDefines))")
				for _, def := range prjCfg.CppDefines.Values {
					escapedDef := strings.Replace(def, `"`, `\"`, -1)
					out.WriteILine("+", "project.Defines.Add(", `"`, escapedDef, `")`)
				}
				out.WriteILine("+", "project.Defines.AddMany(configDefines...)")
				out.WriteLine()

				numSrcFiles := 0
				numPartitionFiles := 0

				for _, group := range prj.SrcFileGroups {
					for _, src := range group.Values {
						if src.Is_SourceFile() {
							numSrcFiles++
						} else if src.Is_PartitionFile() {
							numPartitionFiles++
						}
					}
				}

				if numSrcFiles > 0 {
					out.WriteILine("+", "// Project Source files")
					out.WriteILine("+", "project.SourceFiles = make([]clay.SourceFile, 0, ", strconv.Itoa(numSrcFiles), ")")
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
				if numPartitionFiles > 0 {
					out.WriteILine("+", "// Project Partition files")
					out.WriteILine("+", "project.PartitionFiles = make([]string, 0, ", strconv.Itoa(numPartitionFiles), ")")
					for _, group := range prj.SrcFileGroups {
						for _, src := range group.Values {
							if src.Is_PartitionFile() {
								path := filepath.Join(group.Path, src.Path)
								path = strings.Replace(path, "\\", "/", -1)
								out.WriteILine("+", "project.AddPartitionFile(", `"`, path, `")`)
							}
						}
					}

					out.WriteLine()
				}

				out.WriteILine("+", "projects[", projectIndexToId[index], "_id] = project")
				out.WriteILine("", "}")

				index++
			}
		}
	}

	out.WriteILine("", "// Setup Project Dependencies")
	for _, prj := range g.Workspace.ProjectList.Values {
		if prj.BuildTargets.HasOverlap(g.BuildTarget) {
			for _, prjCfg := range prj.Resolved.Configs.Values {
				if prj.Dependencies.Len() > 0 {
					configName := prjCfg.BuildConfig.String()
					projectId := strings.ReplaceAll(prj.Name+"_"+configName, "-", "_")

					out.WriteILine("", "{")
					out.WriteILine("+", "project := projects[", projectId, "_id]")
					out.WriteILine("+", `project.Dependencies = []*clay.Project{`)
					for _, depProject := range prj.Dependencies.Values {
						if depProject.BuildTargets.HasOverlap(g.BuildTarget) {
							depProjectId := strings.ReplaceAll(depProject.Name+"_"+configName, "-", "_")
							out.WriteILine("++", "projects[", depProjectId, "_id],")
						}
					}
					out.WriteILine("+", "}")
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
