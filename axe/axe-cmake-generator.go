package axe

import (
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/denv"
)

type CMakeGenerator struct {
	Workspace  *Workspace
	VcxProjCpu string
}

func NewCMakeGenerator(ws *Workspace) *CMakeGenerator {
	g := &CMakeGenerator{
		Workspace: ws,
	}

	return g
}

func (g *CMakeGenerator) Generate() {

	makefile := NewLineWriter(IndentModeSpaces)

	mainprj := g.Workspace.StartupProject
	dependencies := g.Workspace.StartupProject.Dependencies

	makefile.WriteLine(`cmake_minimum_required(VERSION 3.23)`)
	makefile.WriteILine(``, `project(`, mainprj.Name, ` LANGUAGES CXX)`)
	makefile.WriteLine(`set(CMAKE_CXX_STANDARD 14)`)
	makefile.WriteLine(``)

	for _, cfg := range g.Workspace.Configs.Values {

		makefile.WriteILine(``, `macro(config_`, g.asIdentifierString(cfg.Type.String()), `)`)
		makefile.WriteILine(`+`, `message(STATUS "Configuring for `, cfg.Type.String(), `")`)
		makefile.WriteLine(``)

		makefile.WriteILine(`+`, `set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY lib/`, cfg.Type.String(), `)`)
		makefile.WriteILine(`+`, `set(CMAKE_LIBRARY_OUTPUT_DIRECTORY lib/`, cfg.Type.String(), `)`)
		makefile.WriteILine(`+`, `set(CMAKE_RUNTIME_OUTPUT_DIRECTORY bin/`, cfg.Type.String(), `)`)
		makefile.WriteLine(``)

		// register the library and executable targets

		for _, dep := range dependencies.Values {
			switch dep.Type {
			case ProjectTypeCLib, ProjectTypeCppLib:
				makefile.WriteILine(`+`, `add_library(`, dep.Name, `_library STATIC)`)
			case ProjectTypeCDll, ProjectTypeCppDll:
				makefile.WriteILine(`+`, `add_library(`, dep.Name, `_library SHARED)`)
			case ProjectTypeCExe, ProjectTypeCppExe:
				makefile.WriteILine(`+`, `add_executable(`, dep.Name, `_program)`)
			}
		}

		if mainprj.Type.IsExecutable() {
			makefile.WriteILine(`+`, `add_executable(`, mainprj.Name, `_program)`)
		}
		makefile.WriteLine(``)

		for _, dep := range dependencies.Values {

			makefile.WriteILine(`+`, `# `, dep.Name, `_library`)
			makefile.WriteLine(``)
			{
				// source files

				makefile.WriteILine(`+`, `# set source files`)
				makefile.WriteILine(`+`, `target_sources(`, dep.Name, `_library PRIVATE`)
				{
					for _, src := range dep.FileEntries.Values {
						path := dep.FileEntries.GetRelativePath(src, dep.Workspace.GenerateAbsPath)
						makefile.WriteILine("++", path)
					}
				}
				makefile.WriteILine(`+`, `)`)

				// compiler definitions

				makefile.WriteILine(`+`, `# set compiler definitions`)
				makefile.WriteILine(`+`, `target_compile_definitions(`, dep.Name, `_library PUBLIC`)
				{
					dcfg, _ := dep.Configs.Get(cfg.Type)
					for _, def := range dcfg.CppDefines.FinalDict.Values {
						escapedDef := g.escapeString(def)
						makefile.WriteILine("++", escapedDef)
					}
				}
				makefile.WriteILine(`+`, `)`)

				// include directories

				makefile.WriteILine("+", `target_include_directories(`, dep.Name, `_library PUBLIC`)
				{
					{
						dcfg, _ := dep.Configs.Get(cfg.Type)
						for _, inc := range dcfg.IncludeDirs.FinalDict.Values {
							path := PathGetRel(filepath.Join(dep.ProjectAbsPath, inc), g.Workspace.GenerateAbsPath)
							makefile.WriteILine("++", path)
						}
					}
					for _, depdep := range dep.Dependencies.Values {
						dcfg, _ := depdep.Configs.Get(cfg.Type)
						for _, inc := range dcfg.IncludeDirs.FinalDict.Values {
							path := PathGetRel(filepath.Join(depdep.ProjectAbsPath, inc), g.Workspace.GenerateAbsPath)
							makefile.WriteILine(`++`, path)
						}
					}

				}
				makefile.WriteILine(`+`, `)`)
			}
			makefile.WriteILine(``)
		}

		makefile.WriteILine(`+`, `set(CMAKE_CXX_FLAGS "-g ${CMAKE_CXX_FLAGS}")`)
		makefile.WriteILine(`+`, `set(CMAKE_BUILD_TYPE `, cfg.Type.String(), `)`)

		// use the name of this dependency appended with _SOURCES
		makefile.WriteILine(`+`, `# set source files`)
		makefile.WriteILine(`+`, `target_sources(`, mainprj.Name, `_program PRIVATE`)
		{
			for _, src := range mainprj.FileEntries.Values {
				path := mainprj.FileEntries.GetRelativePath(src, g.Workspace.GenerateAbsPath)
				makefile.WriteILine("++", path)
			}
		}
		makefile.WriteILine(`+`, `)`)

		// compiler definitions
		makefile.WriteILine(`+`, `# set compiler definitions`)

		makefile.WriteILine(`+`, `target_compile_definitions(`, mainprj.Name, `_program PUBLIC `)
		{
			for _, def := range cfg.CppDefines.FinalDict.Values {
				escapedDef := g.escapeString(def)
				makefile.WriteILine("++", escapedDef)
			}
		}
		makefile.WriteILine(`+`, `)`)

		makefile.WriteILine(`+`, `# set include directories`)
		{
			for _, inc := range cfg.IncludeDirs.FinalDict.Values {
				path := PathGetRel(filepath.Join(mainprj.ProjectAbsPath, inc), g.Workspace.GenerateAbsPath)
				makefile.WriteILine(`+`, `target_include_directories(`, mainprj.Name, `_program PUBLIC `, path, `)`)
			}

			for _, dep := range dependencies.Values {
				dcfg, _ := dep.Configs.Get(cfg.Type)
				for _, inc := range dcfg.IncludeDirs.FinalDict.Values {
					path := PathGetRel(filepath.Join(dep.ProjectAbsPath, inc), g.Workspace.GenerateAbsPath)
					makefile.WriteILine(`+`, `target_include_directories(`, mainprj.Name, `_program PUBLIC `, path, `)`)
				}
			}
		}

		// link libraries

		makefile.WriteILine(`+`, `# link libraries`)
		makefile.WriteILine(`+`, `target_link_libraries(`, mainprj.Name, `_program PRIVATE`)
		for _, dep := range mainprj.Dependencies.Values {
			makefile.WriteILine(`+++`, dep.Name, `_library`)
		}

		// platform specific libraries

		if denv.OS == denv.OS_MAC {
			makefile.WriteILine(`+++`, `"-framework Foundation"`)
			makefile.WriteILine(`+++`, `"-framework Cocoa"`)
			makefile.WriteILine(`+++`, `"-framework Metal"`)
			makefile.WriteILine(`+++`, `"-framework OpenGL"`)
			makefile.WriteILine(`+++`, `"-framework IOKit"`)
			makefile.WriteILine(`+++`, `"-framework CoreVideo"`)
			makefile.WriteILine(`+++`, `"-framework QuartzCore"`)
		}
		makefile.WriteILine(`+`, `)`)

		makefile.WriteLine(`endmacro()`)
		makefile.WriteLine(``)
	}

	for _, cfg := range g.Workspace.Configs.Values {

		if cfg.Type.IsDebug() {
			makefile.WriteLine(`if (CMAKE_BUILD_TYPE STREQUAL "DEBUG")`)
			makefile.WriteILine(`+`, `config_`, g.asIdentifierString(cfg.Type.String()), `()`)
			makefile.WriteLine(`endif ()`)
		} else if cfg.Type.IsRelease() {
			makefile.WriteLine(`if (CMAKE_BUILD_TYPE STREQUAL "RELEASE")`)
			makefile.WriteILine(`+`, `config_`, g.asIdentifierString(cfg.Type.String()), `()`)
			makefile.WriteLine(`endif ()`)
		}
	}

	makefilepath := filepath.Join(g.Workspace.GenerateAbsPath, "CMakeLists.txt")
	makefile.WriteToFile(makefilepath)
}

func (g *CMakeGenerator) asIdentifierString(s string) string {
	s = strings.Replace(s, "+", "_", -1)
	s = strings.Replace(s, "-", "_", -1)
	return s
}

func (g *CMakeGenerator) escapeString(s string) string {
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "\"", "\\\"", -1)
	s = strings.Replace(s, "-", "_", -1)
	return s
}
