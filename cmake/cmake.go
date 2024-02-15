package cmake

import (
	"fmt"
	"os"
	"strings"

	"github.com/jurgen-kluft/ccode/denv"
	"github.com/jurgen-kluft/ccode/items"
	"github.com/jurgen-kluft/ccode/vars"

	"path/filepath"
)

// AddProjectVariables adds variables from the Project information
//
//	Example for 'xhash' project with 'cbase' as a dependency:
//	- xhash:GUID
//	- xhash:ROOT_DIR
//	- xhash:INCLUDE_DIRS
func addProjectVariables(p *denv.Project, isdep bool, v vars.Variables, r vars.Replacer) {

	p.MergeVars(v)
	p.ReplaceVars(v, r)

	v.AddVar(p.Name+":GUID", p.GUID)
	v.AddVar(p.Name+":ROOT_DIR", denv.Path(p.PackagePath))

	path, _ := filepath.Rel(p.ProjectPath, p.PackagePath)

	switch p.Type {
	case denv.StaticLibrary:
		v.AddVar(p.Name+":TYPE", "Static Library")
	case denv.SharedLibrary:
		v.AddVar(p.Name+":TYPE", "Shared Library")
	case denv.Executable:
		v.AddVar(p.Name+":TYPE", "Executable")
	}

	if isdep {
		v.AddVar(fmt.Sprintf("%s:SOURCE_DIR", p.Name), denv.Path("..\\"+p.Name+"\\"+p.SrcPath))
	} else {
		v.AddVar(fmt.Sprintf("%s:SOURCE_DIR", p.Name), denv.Path(p.SrcPath))
	}

	for _, config := range p.Platform.Configs {
		config.Defines.Delimiter = " "
		v.AddVar(p.Name+":"+config.Name+":DEFINES", config.Defines.String())
	}

	for _, config := range p.Platform.Configs {
		includes := config.IncludeDirs.Prefix(path, items.PathPrefixer)
		includes = includes.Prefix(path, denv.PathFixer)
		includes.Delimiter = ","
		includes.Quote = ``
		v.AddVar(fmt.Sprintf("%s:INCLUDE_DIRS", p.Name), includes.String())
	}
}

func collectProjectDefinesFromDependencies(prj *denv.Project) {
	for _, prjConfig := range prj.Platform.Configs {
		for _, dep := range prj.Dependencies {
			// Get the config from the dependency
			depConfig := dep.GetConfig(prjConfig.Name)
			// Merge the defines
			prjConfig.Defines = prjConfig.Defines.Merge(depConfig.Defines)
		}
	}
}

// setupProjectPaths will set correct paths for the main and dependency packages
// Note: This currently assumes that the dependency packages are in the vendor
//
//	folder relative to the main package.
//
// All project and workspace files will be written in the root of the main package
func setupProjectPaths(prj *denv.Project, deps []*denv.Project) {
	prj.PackagePath, _ = os.Getwd()
	prj.ProjectPath, _ = os.Getwd()
	fmt.Println("PACKAGE:" + prj.Name + " -  packagePath=" + prj.PackagePath + ", projectpath=" + prj.ProjectPath)
	for _, dep := range deps {
		//dep.PackagePath = filepath.Join(prj.PackagePath, "vendor", denv.Path(dep.PackageURL))
		dep.PackagePath = denv.Path(filepath.Join(prj.PackagePath, "..", dep.Name))
		dep.ProjectPath = prj.ProjectPath
		fmt.Println("DEPENDENCY:" + dep.Name + " -  packagePath=" + dep.PackagePath + ", projectpath=" + dep.ProjectPath)
	}
}

type strStack []string

func (s strStack) Empty() bool    { return len(s) == 0 }
func (s strStack) Peek() string   { return s[len(s)-1] }
func (s *strStack) Push(i string) { (*s) = append((*s), i) }
func (s *strStack) Pop() string {
	d := (*s)[len(*s)-1]
	(*s) = (*s)[:len(*s)-1]
	return d
}

// GenerateBuildFiles will generate the `CMakeLists.txt` file to be used by the Make Build System
func GenerateBuildFiles(pkg *denv.Package) error {
	mainprj := pkg.GetMainApp()
	mainapp := true
	if mainprj == nil {
		mainapp = false
		mainprj = pkg.GetUnittest()
	}
	if mainprj == nil {
		return fmt.Errorf("this package has no main app or main test")
	}

	// And dependency projects (dependency tree)
	depmap := map[string]*denv.Project{}
	depmap[mainprj.Name] = mainprj
	depstack := &strStack{mainprj.Name}
	for !depstack.Empty() {
		prjname := depstack.Pop()
		prj := depmap[prjname]
		for _, dep := range prj.Dependencies {
			if _, ok := depmap[dep.Name]; !ok {
				depstack.Push(dep.Name)
				depmap[dep.Name] = dep
			}
		}
	}
	delete(depmap, mainprj.Name)

	dependencies := []*denv.Project{}
	for _, dep := range depmap {
		dependencies = append(dependencies, dep)
	}

	setupProjectPaths(mainprj, dependencies)

	variables := vars.NewVars()
	replacer := vars.NewReplacer()

	// Main project
	projects := []*denv.Project{mainprj}
	projects = append(projects, dependencies...)

	// Register project variables
	for _, prj := range projects {
		isdep := prj.PackageURL != mainprj.PackageURL
		addProjectVariables(prj, isdep, variables, replacer)
	}

	// For each project collect defines from dependencies
	for _, prj := range projects {
		collectProjectDefinesFromDependencies(prj)
	}

	variables.Print()

	// Glob all the source and header files for every project
	for _, prj := range projects {
		fmt.Println("GLOBBING: " + prj.Name + " : " + prj.PackagePath + " : ignore(" + strings.Join(prj.Platform.FilePatternsToIgnore, ", ") + ")")
		prj.SrcFiles.GlobFiles(prj.PackagePath, prj.Platform.FilePatternsToIgnore)
		prj.HdrFiles.GlobFiles(prj.PackagePath, prj.Platform.FilePatternsToIgnore)

		// Convert the list of source files to a string by delimiting with double quotes and joining them with a comma
		relpath, _ := filepath.Rel(prj.ProjectPath, prj.PackagePath)
		src_files := ""
		for i, src := range prj.SrcFiles.Files {
			if i > 0 {
				src_files += ","
			}
			src_files += "../../" + filepath.Join(relpath, src)
		}

		// Register the list of source files as a variable for the project
		variables.AddVar(prj.Name+":SOURCE_FILES", src_files)
	}

	makefile := &denv.ProjectTextWriter{}
	os.MkdirAll("target/cmake", os.ModePerm)
	makefilepath := filepath.Join(mainprj.ProjectPath, "target/cmake/CMakeLists.txt")
	if makefile.Open(makefilepath) != nil {
		fmt.Printf("Error opening file '%s'", makefilepath)
		return fmt.Errorf("error opening file '%s'", makefilepath)
	}

	makefile.WriteLn(`cmake_minimum_required(VERSION 3.23)`)
	makefile.WriteLn(`project(${Name} LANGUAGES CXX)`)
	makefile.WriteLn(`set(CMAKE_CXX_STANDARD 14)`)
	makefile.WriteLn(``)

	for _, cfg := range mainprj.Platform.Configs {

		makefile.WriteLn(`macro(config_` + cfg.Name + `)`)
		makefile.WriteLn(`+message(STATUS "Configuring for ` + cfg.Name + `")`)
		makefile.WriteLn(``)

		makefile.WriteLn(`+set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/target/lib/` + cfg.Name + `)`)
		makefile.WriteLn(`+set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/target/lib/` + cfg.Name + `)`)
		makefile.WriteLn(`+set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/target/bin/` + cfg.Name + `)`)
		makefile.WriteLn(``)

		// register the library and executable targets
		for _, dep := range dependencies {
			switch dep.Type {
			case denv.StaticLibrary:
				makefile.WriteLn(`+add_library(` + dep.Name + `_library STATIC)`)
			case denv.SharedLibrary:
				makefile.WriteLn(`+add_library(` + dep.Name + `_library SHARED)`)
			case denv.Executable:
				makefile.WriteLn(`+add_executable(` + dep.Name + `_program)`)
			}
		}

		if mainprj.Type == denv.Executable {
			makefile.WriteLn(`+add_executable(` + mainprj.Name + `_program)`)
		}
		makefile.WriteLn(``)

		for _, dep := range dependencies {
			dependency := make([]string, 0)

			// use the name of this dependency appended with _SOURCES
			dependency = append(dependency, `+# ${Name}_library`)
			makefile.WriteLn(``)

			dependency = append(dependency, `+# set source files`)
			dependency = append(dependency, `+target_sources(${Name}_library PRIVATE`)
			dependency_sources := strings.Split(variables.GetVar(dep.Name+":SOURCE_FILES"), ",")
			for _, source := range dependency_sources {
				dependency = append(dependency, `++`+source)
			}
			dependency = append(dependency, `+)`)

			// compiler definitions
			dependency = append(dependency, `+# set compiler definitions`)
			dependency = append(dependency, `+target_compile_definitions(${Name}_library PUBLIC ${`+dep.Name+`:`+cfg.Name+`:DEFINES})`)

			// register all the include directories of this dependency
			dependency = append(dependency, `+# set include directories`)
			includeDirectories := strings.Split(variables.GetVar(dep.Name+":INCLUDE_DIRS"), ",")
			for _, includeDirectory := range includeDirectories {
				dependency = append(dependency, `+target_include_directories(${Name}_library PUBLIC ../../`+includeDirectory+`)`)
			}
			for _, ddep := range dep.Dependencies {
				includeDirectories := strings.Split(variables.GetVar(ddep.Name+":INCLUDE_DIRS"), ",")
				for _, includeDirectory := range includeDirectories {
					dependency = append(dependency, `+target_include_directories(${Name}_library PUBLIC ../../`+includeDirectory+`)`)
				}
			}

			dependency = append(dependency, ``)

			replacer.ReplaceInLines("${Name}", dep.Name, dependency)
			variables.ReplaceInLines(replacer, dependency)
			makefile.WriteLns(dependency)
		}

		makefile.WriteLn(`+set(CMAKE_CXX_FLAGS "-g ${CMAKE_CXX_FLAGS}")`)
		makefile.WriteLn(`+set(CMAKE_BUILD_TYPE ` + cfg.Config + `)`)

		program := []string{}

		// use the name of this dependency appended with _SOURCES
		program = append(program, `+# set source files`)
		program = append(program, `+target_sources(${Name}_program PRIVATE`)
		dependency_sources := strings.Split(variables.GetVar(mainprj.Name+":SOURCE_FILES"), ",")
		for _, source := range dependency_sources {
			program = append(program, `++`+source)
		}
		program = append(program, `+)`)

		// compiler definitions
		program = append(program, `+# set compiler definitions`)
		program = append(program, `+target_compile_definitions(${Name}_program PUBLIC ${`+mainprj.Name+`:`+cfg.Name+`:DEFINES})`)

		// register all the include directories of this program
		program = append(program, `+# set include directories`)
		includeDirectories := strings.Split(variables.GetVar(mainprj.Name+":INCLUDE_DIRS"), ",")
		for _, includeDirectory := range includeDirectories {
			program = append(program, `+target_include_directories(${Name}_program PUBLIC ../../`+includeDirectory+`)`)
		}

		for _, dep := range mainprj.Dependencies {
			includeDirectories := strings.Split(variables.GetVar(dep.Name+":INCLUDE_DIRS"), ",")
			for _, includeDirectory := range includeDirectories {
				program = append(program, `+target_include_directories(${Name}_program PUBLIC ../../`+includeDirectory+`)`)
			}
		}

		// target_link_libraries
		link_libraries := "+target_link_libraries(${Name}_program"
		for _, dep := range dependencies {
			link_libraries = link_libraries + ` ` + dep.Name + `_library`
		}
		link_libraries = link_libraries + `)`
		program = append(program, link_libraries)

		// if the platform is Mac also write out the Frameworks we are using
		// if mainprj.Platform.OS == "darwin" {
		// 	program = append(program, `Frameworks = {`)
		// 	program = append(program, `+{ "Cocoa" },`)
		// 	program = append(program, `+{ "Metal" },`)
		// 	program = append(program, `+{ "OpenGL" },`)
		// 	program = append(program, `+{ "IOKit" },`)
		// 	program = append(program, `+{ "CoreVideo" },`)
		// 	program = append(program, `+{ "QuartzCore" },`)
		// 	program = append(program, `},`)
		// }

		configitems := map[string]items.List{
			"INCLUDE_DIRS": items.NewList(variables.GetVar(mainprj.Name+":INCLUDE_DIRS"), ",", ""),
		}

		for configitem, defaults := range configitems {
			varkeystr := fmt.Sprintf("${%s}", configitem)
			varlist := defaults.Copy()

			for _, depDep := range mainprj.Dependencies {
				varkey := fmt.Sprintf("%s:%s", depDep.Name, configitem)
				varitem := variables.GetVar(varkey)
				if len(varitem) > 0 {
					varlist = varlist.Add(varitem)
				}
			}
			varset := items.ListToSet(varlist)
			replacer.InsertInLines(varkeystr, varset.String(), "", program)
			replacer.ReplaceInLines(varkeystr, "", program)

		}

		depends := items.NewList("", ",", "")
		for _, dep := range dependencies {
			depends = depends.Add(dep.Name + "_library")
		}
		replacer.ReplaceInLines("${DEPENDS}", depends.String(), program)
		replacer.ReplaceInLines("${SOURCE_DIR}", "${"+mainprj.Name+":SOURCE_DIR}", program)

		if mainapp {
			replacer.ReplaceInLines("${Main}", "app", program)
		} else {
			replacer.ReplaceInLines("${Main}", "unittest", program)
		}

		replacer.ReplaceInLines("${Name}", mainprj.Name, program)
		variables.ReplaceInLines(replacer, program)
		makefile.WriteLns(program)

		makefile.WriteLn(`endmacro()`)
		makefile.WriteLn(``)
	}

	// for _, cfg := range mainprj.Platform.Configs {
	// }
	makefile.WriteLn(`if (CMAKE_BUILD_TYPE STREQUAL "DEBUG")`)
	makefile.WriteLn(`+config_DevDebugStatic()`)
	makefile.WriteLn(`elseif (CMAKE_BUILD_TYPE STREQUAL "RELEASE")`)
	makefile.WriteLn(`+config_DevReleaseStatic()`)
	makefile.WriteLn(`endif ()`)

	return nil
}

// IsCMake checks if IDE is requesting a cmake build file
func IsCMake(DEV string, OS string, ARCH string) bool {
	return strings.ToLower(DEV) == "cmake"
}
