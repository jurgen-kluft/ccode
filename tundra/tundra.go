package tundra

import (
	"fmt"
	"os"
	"strings"

	"github.com/jurgen-kluft/xcode/denv"
	"github.com/jurgen-kluft/xcode/items"
	"github.com/jurgen-kluft/xcode/vars"

	"path/filepath"
)

// func fullReplaceVar(varname string, prjname string, platform string, config string, v vars.Variables, replacer func(name, value string)) bool {
// 	value, err := v.GetVar(fmt.Sprintf("%s:%s[%s][%s]", prjname, varname, platform, config))
// 	if err == nil {
// 		replacer(varname, value)
// 	} else {
// 		value, err = v.GetVar(fmt.Sprintf("%s:%s", prjname, varname))
// 		if err == nil {
// 			replacer(varname, value)
// 		} else {
// 			return false
// 		}
// 	}
// 	return true
// }

// func fullReplaceVarWithDefault(varname string, vardefaultvalue string, prjname string, platform string, config string, v vars.Variables, replacer func(name, value string)) {
// 	if !fullReplaceVar(varname, prjname, platform, config, v, replacer) {
// 		replacer(varname, vardefaultvalue)
// 	}
// }

// AddProjectVariables adds variables from the Project information
//   Example for 'xhash' project with 'xbase' as a dependency:
//   - xhash:GUID
//   - xhash:ROOT_DIR
//   - xhash:INCLUDE_DIRS
//
func addProjectVariables(p *denv.Project, isdep bool, v vars.Variables, r vars.Replacer) {

	p.MergeVars(v)
	p.ReplaceVars(v, r)

	v.AddVar(p.Name+":GUID", p.GUID)
	v.AddVar(p.Name+":ROOT_DIR", denv.Path(p.PackagePath))

	path, _ := filepath.Rel(p.ProjectPath, p.PackagePath)

	switch p.Type {
	case denv.StaticLibrary:
		v.AddVar(p.Name+":TYPE", "StaticLibrary")
	case denv.SharedLibrary:
		v.AddVar(p.Name+":TYPE", "SharedLibrary")
	case denv.Executable:
		v.AddVar(p.Name+":TYPE", "Program")
	}

	if isdep {
		v.AddVar(fmt.Sprintf("%s:SOURCE_DIR", p.Name), denv.Path("..\\"+p.Name+"\\"+p.SrcPath))
	} else {
		v.AddVar(fmt.Sprintf("%s:SOURCE_DIR", p.Name), denv.Path(p.SrcPath))
	}

	var platform = p.Platform
	{
		for _, config := range platform.Configs {
			includes := config.IncludeDirs.Prefix(path, items.PathPrefixer)
			includes = includes.Prefix(path, denv.PathFixer)
			includes.Delimiter = ","
			includes.Quote = `"`
			v.AddVar(fmt.Sprintf("%s:INCLUDE_DIRS", p.Name), includes.String())
		}
	}

}

// setupProjectPaths will set correct paths for the main and dependency packages
// Note: This currently assumes that the dependency packages are in the vendor
//       folder relative to the main package.
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

// GenerateTundraBuildFile will generate the tundra.lua file to be used by the Tundra Build System
func GenerateTundraBuildFile(pkg *denv.Package) error {
	mainprj := pkg.GetMainApp()
	mainapp := true
	if mainprj == nil {
		mainapp = false
		mainprj = pkg.GetUnittest()
	}
	if mainprj == nil {
		return fmt.Errorf("this package has no main app or main test")
	}

	writer := &denv.ProjectTextWriter{}
	slnfilepath := filepath.Join(mainprj.ProjectPath, "tundra.lua")
	if writer.Open(slnfilepath) != nil {
		fmt.Printf("Error opening file '%s'", slnfilepath)
		return fmt.Errorf("error opening file '%s'", slnfilepath)
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

	for _, prj := range projects {
		isdep := prj.PackageURL != mainprj.PackageURL
		addProjectVariables(prj, isdep, variables, replacer)
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
		for n, src := range prj.SrcFiles.Files {
			srcfile := filepath.Join(relpath, src)
			if src_files != "" {
				src_files += ","
			}
			src_files += `"` + srcfile + `"`
			if n%3 == 2 { // max 3 entries on one line
				src_files += "\n"
			}
		}

		// Register the list of source files as a variable for the project
		variables.AddVar(prj.Name+":SOURCE_FILES", src_files)
	}

	writer.WriteLn(`local GlobExtension = require("tundra.syntax.glob")`)
	writer.WriteLn(``)
	writer.WriteLn(`Build {`)
	writer.WriteLn(`+ReplaceEnv = {`)
	writer.WriteLn(`++OBJECTROOT = "target",`)
	writer.WriteLn(`+},`)
	writer.WriteLn(`+Env = {`)
	writer.WriteLn(`++CPPDEFS = {`)

	writer.WriteLn(`+++{ "TARGET_PC_DEV_DEBUG", "TARGET_PC", "PLATFORM_64BIT"; Config = "win64-*-debug-dev" },`)
	writer.WriteLn(`+++{ "TARGET_PC_DEV_RELEASE", "TARGET_PC", "PLATFORM_64BIT"; Config = "win64-*-release-dev" },`)
	writer.WriteLn(`+++{ "TARGET_PC_TEST_DEBUG", "TARGET_PC", "PLATFORM_64BIT"; Config = "win64-*-debug-test" },`)
	writer.WriteLn(`+++{ "TARGET_PC_TEST_RELEASE", "TARGET_PC", "PLATFORM_64BIT"; Config = "win64-*-release-test" },`)

	writer.WriteLn(`+++{ "TARGET_MAC_DEV_DEBUG", "TARGET_MAC", "PLATFORM_64BIT"; Config = "macosx-*-debug-dev" },`)
	writer.WriteLn(`+++{ "TARGET_MAC_DEV_RELEASE", "TARGET_MAC", "PLATFORM_64BIT"; Config = "macosx-*-release-dev" },`)
	writer.WriteLn(`+++{ "TARGET_MAC_TEST_DEBUG", "TARGET_MAC", "PLATFORM_64BIT"; Config = "macosx-*-debug-test" },`)
	writer.WriteLn(`+++{ "TARGET_MAC_TEST_RELEASE", "TARGET_MAC", "PLATFORM_64BIT"; Config = "macosx-*-release-test" },`)

	writer.WriteLn(`+++{ "TARGET_LINUX_DEV_DEBUG", "TARGET_LINUX", "PLATFORM_64BIT"; Config = "linux-*-debug-dev" },`)
	writer.WriteLn(`+++{ "TARGET_LINUX_DEV_RELEASE", "TARGET_LINUX", "PLATFORM_64BIT"; Config = "linux-*-release-dev" },`)
	writer.WriteLn(`+++{ "TARGET_LINUX_TEST_DEBUG", "TARGET_LINUX", "PLATFORM_64BIT"; Config = "linux-*-debug-test" },`)
	writer.WriteLn(`+++{ "TARGET_LINUX_TEST_RELEASE", "TARGET_LINUX", "PLATFORM_64BIT"; Config = "linux-*-release-test" },`)

	writer.WriteLn(`++},`)
	writer.WriteLn(`+},`)
	writer.WriteLn(`+Units = function ()`)

	for _, dep := range dependencies {
		dependency := []string{}
		dependency = append(dependency, `++local ${Name}_library = ${${Name}:TYPE} {`)
		dependency = append(dependency, `+++Name = "${Name}",`)
		if strings.HasSuffix(dep.Name, "test") {
			dependency = append(dependency, `+++Config = "*-*-*-test",`)
		} else {
			dependency = append(dependency, `+++Config = "*-*-*-*",`)
		}

		// Scan the source folder for source files and write them all out
		dependency = append(dependency, `+++Sources = { ${SOURCE_FILES} },`)

		// We do not want to have tundra glob the source files, so we glob them here in the same
		// manner as we would do for Visual Studio.

		// Scan the include folder for folders that need to be added to the include path
		dependency = append(dependency, `+++Includes = { ${INCLUDE_DIRS} },`)
		dependency = append(dependency, `++}`)

		replacer.ReplaceInLines("${SOURCE_FILES}", "${"+dep.Name+":SOURCE_FILES}", dependency)
		replacer.ReplaceInLines("${SOURCE_DIR}", "${"+dep.Name+":SOURCE_DIR}", dependency)

		configitems := map[string]items.List{
			"SOURCE_FILES": items.NewList("${"+dep.Name+":SOURCE_FILES}", ",", ""),
			"INCLUDE_DIRS": items.NewList("${"+dep.Name+":INCLUDE_DIRS}", ",", ""),
		}

		for configitem, defaults := range configitems {
			varkeystr := fmt.Sprintf("${%s}", configitem)
			varlist := defaults.Copy()

			if configitem != "SOURCE_FILES" {
				for _, depDep := range dep.Dependencies {
					varkey := fmt.Sprintf("%s:%s", depDep.Name, configitem)
					varitem, err := variables.GetVar(varkey)
					if err == nil {
						varlist = varlist.Add(varitem)
					} else {
						fmt.Println("ERROR: could not find variable " + varkey)
					}
				}
			}
			varset := items.ListToSet(varlist)
			replacer.InsertInLines(varkeystr, varset.String(), "", dependency)
			replacer.ReplaceInLines(varkeystr, "", dependency)
		}

		replacer.ReplaceInLines("${Name}", dep.Name, dependency)
		variables.ReplaceInLines(replacer, dependency)

		writer.WriteLns(dependency)
	}

	program := []string{}
	program = append(program, `++local ${Main} = ${${Name}:TYPE} {`)
	program = append(program, `+++Name = "${Name}",`)
	if strings.HasSuffix(mainprj.Name, "test") {
		program = append(program, `+++Config = "*-*-*-test",`)
	} else {
		program = append(program, `+++Config = "*-*-*-*",`)
	}
	program = append(program, `+++Sources = { ${SOURCE_FILES} },`)
	program = append(program, `+++Includes = { ${INCLUDE_DIRS} },`)
	program = append(program, `+++Depends = { ${DEPENDS} },`)
	program = append(program, `++}`)

	configitems := map[string]items.List{
		"SOURCE_FILES": items.NewList("${"+mainprj.Name+":SOURCE_FILES}", ",", ""),
		"INCLUDE_DIRS": items.NewList("${"+mainprj.Name+":INCLUDE_DIRS}", ",", ""),
	}

	for configitem, defaults := range configitems {
		varkeystr := fmt.Sprintf("${%s}", configitem)
		varlist := defaults.Copy()

		if configitem != "SOURCE_FILES" {
			for _, depDep := range mainprj.Dependencies {
				varkey := fmt.Sprintf("%s:%s", depDep.Name, configitem)
				varitem, err := variables.GetVar(varkey)
				if err == nil {
					varlist = varlist.Add(varitem)
				} else {
					fmt.Println("ERROR: could not find variable " + varkey)
				}
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
	writer.WriteLns(program)

	if mainapp {
		writer.WriteLn(`++Default(app)`)
	} else {
		writer.WriteLn(`++Default(unittest)`)
	}
	writer.WriteLn(`+end,`)
	writer.WriteLn(`+Configs = {`)
	writer.WriteLn(`++Config {`)
	writer.WriteLn(`+++Name = "macosx-clang",`)
	writer.WriteLn(`+++Env = {`)
	writer.WriteLn(`+++PROGOPTS = { "-lc++" },`)
	writer.WriteLn(`+++CXXOPTS = {`)
	writer.WriteLn(`++++"-std=c++11",`)
	writer.WriteLn(`++++"-arch x86_64",`)
	writer.WriteLn(`++++"-g",`)
	writer.WriteLn(`++++"-Wno-new-returns-null",`)
	writer.WriteLn(`++++"-Wno-missing-braces",`)
	writer.WriteLn(`++++"-Wno-unused-function",`)
	writer.WriteLn(`++++"-Wno-unused-variable",`)
	writer.WriteLn(`++++"-Wno-unused-result",`)
	writer.WriteLn(`++++"-Wno-write-strings",`)
	writer.WriteLn(`++++"-Wno-c++11-compat-deprecated-writable-strings",`)
	writer.WriteLn(`++++"-Wno-null-dereference",`)
	writer.WriteLn(`++++"-Wno-format",`)
	writer.WriteLn(`++++"-fno-strict-aliasing",`)
	writer.WriteLn(`++++"-fno-omit-frame-pointer",`)
	writer.WriteLn(`+++},`)
	writer.WriteLn(`++},`)
	writer.WriteLn(`++DefaultOnHost = "macosx",`)
	writer.WriteLn(`++Tools = { "clang" },`)
	writer.WriteLn(`++},`)
	writer.WriteLn(`++Config {`)
	writer.WriteLn(`+++ReplaceEnv = {`)
	writer.WriteLn(`++++OBJECTROOT = "target",`)
	writer.WriteLn(`+++},`)
	writer.WriteLn(`+++Name = "linux-gcc",`)
	writer.WriteLn(`+++DefaultOnHost = "linux",`)
	writer.WriteLn(`+++Tools = { "gcc" },`)
	writer.WriteLn(`++},`)
	writer.WriteLn(`++Config {`)
	writer.WriteLn(`+++ReplaceEnv = {`)
	writer.WriteLn(`++++OBJECTROOT = "target",`)
	writer.WriteLn(`+++},`)
	writer.WriteLn(`+++Name = "win64-msvc",`)
	writer.WriteLn(`+++Env = {`)
	writer.WriteLn(`++++PROGOPTS = { "/SUBSYSTEM:CONSOLE" },`)
	writer.WriteLn(`++++CXXOPTS = { },`)
	writer.WriteLn(`+++},`)
	writer.WriteLn(`+++DefaultOnHost = "windows",`)
	writer.WriteLn(`+++Tools = { "msvc-vs2017" },`)
	writer.WriteLn(`++},`)
	writer.WriteLn(`+},`)
	writer.WriteLn(``)
	writer.WriteLn(`+SubVariants = { "dev", "test" },`)
	writer.WriteLn(`}`)

	writer.Close()

	return nil
}

// IsTundra checks if IDE is requesting a Tundra build file
func IsTundra(DEV string, OS string, ARCH string) bool {
	return strings.ToLower(DEV) == "tundra"
}
