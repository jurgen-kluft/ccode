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

	units := &denv.ProjectTextWriter{}
	unitsfilepath := filepath.Join(mainprj.ProjectPath, "units.lua")
	if units.Open(unitsfilepath) != nil {
		fmt.Printf("Error opening file '%s'", unitsfilepath)
		return fmt.Errorf("error opening file '%s'", unitsfilepath)
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

	units.WriteLn(`require "tundra.syntax.glob"`)
	units.WriteLn(`require "tundra.path"`)
	units.WriteLn(`require "tundra.util"`)

	/*
	   -----------------------------------------------------------------------------------------------------------------------

	   local XUNITTEST_SRCDIR = "../xunittest/source/main/cpp/"
	   local XUNITTEST_INCDIR = "../xunittest/source/main/include/"

	   local xunittest_library = StaticLibrary {
	       Name = "xunittest",

	       Env = {
	           CPPPATH = {
	               XUNITTEST_SRCDIR,
	               XUNITTEST_INCDIR,
	           },
	           CPPDEFS = {
	               { "TARGET_DEBUG", Config = "*-*-debug" },
	               { "TARGET_RELEASE", Config = "*-*-release" },
	               { "TARGET_PC", Config = "win64-*-*" },
	               { "TARGET_LINUX", Config = "linux-*-*" },
	               { "TARGET_MAC", Config = "macos-*-*" },
	           },
	       },

	   	   Includes = { XUNITTEST_INCDIR, XUNITTEST_SRCDIR },

	       Sources = {
	           XUNITTEST_SRCDIR .. "entry/ut_Entry_Mac.cpp",XUNITTEST_SRCDIR .. "ut_AssertException.cpp"
	           ,XUNITTEST_SRCDIR .. "ut_Checks.cpp",XUNITTEST_SRCDIR .. "ut_ReportAssert.cpp"
	           ,XUNITTEST_SRCDIR .. "ut_Stdout_Mac.cpp",XUNITTEST_SRCDIR .. "ut_Stdout_Win32.cpp"
	           ,XUNITTEST_SRCDIR .. "ut_StringBuilder.cpp",XUNITTEST_SRCDIR .. "ut_Test.cpp"
	           ,XUNITTEST_SRCDIR .. "ut_TestList.cpp",XUNITTEST_SRCDIR .. "ut_TestReporter.cpp"
	           ,XUNITTEST_SRCDIR .. "ut_TestReporterStdout.cpp",XUNITTEST_SRCDIR .. "ut_TestReporterTeamCity.cpp"
	           ,XUNITTEST_SRCDIR .. "ut_TestResults.cpp",XUNITTEST_SRCDIR .. "ut_TestRunner.cpp"
	           ,XUNITTEST_SRCDIR .. "ut_TestState.cpp",XUNITTEST_SRCDIR .. "ut_Test_Mac.cpp"
	           ,XUNITTEST_SRCDIR .. "ut_Test_Win32.cpp",XUNITTEST_SRCDIR .. "ut_TimeConstraint.cpp"
	           ,XUNITTEST_SRCDIR .. "ut_TimeHelpers_Mac.cpp",XUNITTEST_SRCDIR .. "ut_TimeHelpers_Win32.cpp"
	           ,XUNITTEST_SRCDIR .. "ut_Utils.cpp",XUNITTEST_SRCDIR .. "ut_Utils_Mac.cpp"
	           ,XUNITTEST_SRCDIR .. "ut_Utils_Win32.cpp"
	       },
	   }

	*/

	for _, dep := range dependencies {
		dependency := make([]string, 0)
		dependency = append(dependency, "")
		dependency = append(dependency, `local ${Name}_library = ${${Name}:TYPE} {`)
		dependency = append(dependency, `+Name = "${Name}",`)
		dependency = append(dependency, `+Env = {`)
		dependency = append(dependency, `++CPPPATH = {`)
		dependency = append(dependency, `+++"${${Name}:SOURCE_DIR}",`)
		dependency = append(dependency, `+++${${Name}:INCLUDE_DIRS},`)
		for _, depDep := range dep.Dependencies {
			dependency = append(dependency, `+++${`+depDep.Name+`:INCLUDE_DIRS},`)
			dependency = append(dependency, `+++"${`+depDep.Name+`:SOURCE_DIR}",`)
		}
		dependency = append(dependency, `++},`)
		dependency = append(dependency, `++CPPDEFS = {`)
		dependency = append(dependency, `+++{ "TARGET_DEBUG", Config = "*-*-debug" },`)
		dependency = append(dependency, `+++{ "TARGET_RELEASE", Config = "*-*-release" },`)
		dependency = append(dependency, `+++{ "TARGET_PC", Config = "win64-*-*" },`)
		dependency = append(dependency, `+++{ "TARGET_LINUX", Config = "linux-*-*" },`)
		dependency = append(dependency, `+++{ "TARGET_MAC", Config = "macos-*-*" },`)
		dependency = append(dependency, `+++{ "TARGET_TEST", Config = "*-*-*-test" },`)
		dependency = append(dependency, `++},`)
		dependency = append(dependency, `+},`)
		dependency = append(dependency, `+Includes = {`)
		dependency = append(dependency, `++${${Name}:INCLUDE_DIRS},`)
		for _, depDep := range dep.Dependencies {
			dependency = append(dependency, `++${`+depDep.Name+`:INCLUDE_DIRS},`)
		}
		dependency = append(dependency, `+},`)
		dependency = append(dependency, `+Sources = {`)
		dependency = append(dependency, `++${${Name}:SOURCE_FILES}`)
		dependency = append(dependency, `+},`)
		dependency = append(dependency, `}`)
		dependency = append(dependency, "")

		replacer.ReplaceInLines("${SOURCE_FILES}", "${"+dep.Name+":SOURCE_FILES}", dependency)
		replacer.ReplaceInLines("${SOURCE_DIR}", "${"+dep.Name+":SOURCE_DIR}", dependency)

		configitems := map[string]items.List{
			"INCLUDE_DIRS": items.NewList("${"+dep.Name+":INCLUDE_DIRS}", ",", ""),
		}

		for configitem, defaults := range configitems {
			varkeystr := fmt.Sprintf("${%s}", configitem)
			varlist := defaults.Copy()

			for _, depDep := range dep.Dependencies {
				varkey := fmt.Sprintf("%s:%s", depDep.Name, configitem)
				varitem := variables.GetVar(varkey)
				if len(varitem) > 0 {
					varlist = varlist.Add(varitem)
				}
			}
			varset := items.ListToSet(varlist)
			replacer.InsertInLines(varkeystr, varset.String(), "", dependency)
			replacer.ReplaceInLines(varkeystr, "", dependency)
		}

		replacer.ReplaceInLines("${Name}", dep.Name, dependency)
		variables.ReplaceInLines(replacer, dependency)

		units.WriteLns(dependency)
	}

	program := []string{}
	program = append(program, `local ${Main} = ${${Name}:TYPE} {`)
	program = append(program, `+Name = "${Name}",`)

	program = append(program, `+Env = {`)
	program = append(program, `++CPPPATH = {`)
	//program = append(program, `+++"${${Name}:SOURCE_DIR}",`)
	//program = append(program, `+++${${Name}:INCLUDE_DIRS},`)
	for _, depDep := range mainprj.Dependencies {
		program = append(program, `+++${`+depDep.Name+`:INCLUDE_DIRS},`)
		program = append(program, `+++"${`+depDep.Name+`:SOURCE_DIR}",`)
	}
	program = append(program, `++},`)
	program = append(program, `++CPPDEFS = {`)
	program = append(program, `+++{ "TARGET_DEBUG", Config = "*-*-debug" },`)
	program = append(program, `+++{ "TARGET_RELEASE", Config = "*-*-release" },`)
	program = append(program, `+++{ "TARGET_PC", Config = "win64-*-*" },`)
	program = append(program, `+++{ "TARGET_LINUX", Config = "linux-*-*" },`)
	program = append(program, `+++{ "TARGET_MAC", Config = "macos-*-*" },`)
	program = append(program, `+++{ "TARGET_TEST", Config = "*-*-*-test" },`)
	program = append(program, `++},`)
	program = append(program, `+},`)

	program = append(program, `+Sources = { ${SOURCE_FILES} },`)
	program = append(program, `+Includes = { ${INCLUDE_DIRS} },`)
	program = append(program, `+Depends = { ${DEPENDS} },`)
	program = append(program, `}`)

	replacer.ReplaceInLines("${SOURCE_FILES}", "${"+mainprj.Name+":SOURCE_FILES}", program)

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
	units.WriteLns(program)

	if mainapp {
		units.WriteLn(`Default(app)`)
	} else {
		units.WriteLn(`Default(unittest)`)
	}

	tundra := &denv.ProjectTextWriter{}
	tundrafilepath := filepath.Join(mainprj.ProjectPath, "tundra.lua")
	if tundra.Open(tundrafilepath) != nil {
		fmt.Printf("Error opening file '%s'", tundrafilepath)
		return fmt.Errorf("error opening file '%s'", tundrafilepath)
	}

	tundra.WriteLn(`local native = require('tundra.native')`)
	tundra.WriteLn(``)
	tundra.WriteLn(`-----------------------------------------------------------------------------------------------------------------------`)
	tundra.WriteLn(``)
	tundra.WriteLn(`local mac_opts = {`)
	tundra.WriteLn(`    "-I.",`)
	tundra.WriteLn(`    "-Wno-new-returns-null",`)
	tundra.WriteLn(`    "-Wno-missing-braces",`)
	tundra.WriteLn(`    "-Wno-c++11-compat-deprecated-writable-strings",`)
	tundra.WriteLn(`    "-Wno-null-dereference",`)
	tundra.WriteLn(`    "-Wno-format",`)
	tundra.WriteLn(`    "-fno-strict-aliasing",`)
	tundra.WriteLn(`    "-fno-omit-frame-pointer",`)
	tundra.WriteLn(`	"-Wno-write-strings",`)
	tundra.WriteLn(`    "-Wno-array-bounds",`)
	tundra.WriteLn(`    "-Wno-attributes",`)
	tundra.WriteLn(`    "-Wno-unused-value",`)
	tundra.WriteLn(`    "-Wno-unused-function",`)
	tundra.WriteLn(`    "-Wno-unused-variable",`)
	tundra.WriteLn(`    "-Wno-unused-result",`)
	tundra.WriteLn(`    { "-O2", "-g"; Config = "*-*-test" },`)
	tundra.WriteLn(`    { "-O0", "-g"; Config = "*-*-debug" },`)
	tundra.WriteLn(`    { "-O3", "-g"; Config = "*-*-release" },`)
	tundra.WriteLn(`}`)
	tundra.WriteLn(``)
	tundra.WriteLn(`-----------------------------------------------------------------------------------------------------------------------`)
	tundra.WriteLn(``)
	tundra.WriteLn(`local macosx = {`)
	tundra.WriteLn(`    Env = {`)
	tundra.WriteLn(`        CCOPTS =  {`)
	tundra.WriteLn(`            mac_opts,`)
	tundra.WriteLn(`        },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`        CXXOPTS = {`)
	tundra.WriteLn(`            mac_opts,`)
	tundra.WriteLn(`            "-std=c++14",`)
	tundra.WriteLn(`			"-arch x86_64",`)
	tundra.WriteLn(`        },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`        SHLIBOPTS = {`)
	tundra.WriteLn(`			"-lstdc++",`)
	tundra.WriteLn(`			{ "-fsanitize=address"; Config = "*-*-debug-asan" },`)
	tundra.WriteLn(`		},`)
	tundra.WriteLn(``)
	tundra.WriteLn(`        PROGCOM = {`)
	tundra.WriteLn(`			"-lstdc++",`)
	tundra.WriteLn(`			{ "-fsanitize=address"; Config = "*-*-debug-asan" },`)
	tundra.WriteLn(`		},`)
	tundra.WriteLn(`    },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`	ReplaceEnv = {`)
	tundra.WriteLn(`		OBJECTROOT = "target",`)
	tundra.WriteLn(`	},`)
	tundra.WriteLn(``)
	tundra.WriteLn(`    Frameworks = {`)
	tundra.WriteLn(`        { "Cocoa" },`)
	tundra.WriteLn(`        { "Metal" },`)
	tundra.WriteLn(`        { "QuartzCore" },`)
	tundra.WriteLn(`    },`)
	tundra.WriteLn(`}`)
	tundra.WriteLn(``)
	tundra.WriteLn(`-----------------------------------------------------------------------------------------------------------------------`)
	tundra.WriteLn(``)
	tundra.WriteLn(`local linux_opts = {`)
	tundra.WriteLn(`    "-I.",`)
	tundra.WriteLn(`    "-Wno-new-returns-null",`)
	tundra.WriteLn(`    "-Wno-missing-braces",`)
	tundra.WriteLn(`    "-Wno-c++11-compat-deprecated-writable-strings",`)
	tundra.WriteLn(`    "-Wno-null-dereference",`)
	tundra.WriteLn(`    "-Wno-format",`)
	tundra.WriteLn(`    "-fno-strict-aliasing",`)
	tundra.WriteLn(`    "-fno-omit-frame-pointer",`)
	tundra.WriteLn(`	"-Wno-write-strings",`)
	tundra.WriteLn(`    "-Wno-array-bounds",`)
	tundra.WriteLn(`    "-Wno-attributes",`)
	tundra.WriteLn(`    "-Wno-unused-value",`)
	tundra.WriteLn(`    "-Wno-unused-function",`)
	tundra.WriteLn(`    "-Wno-unused-variable",`)
	tundra.WriteLn(`    "-Wno-unused-result",`)
	tundra.WriteLn(`    "-DOBJECT_DIR=\\\"$(OBJECTDIR)\\\"",`)
	tundra.WriteLn(`    "-I$(OBJECTDIR)",`)
	tundra.WriteLn(`    "-Wall",`)
	tundra.WriteLn(`    "-fPIC",`)
	tundra.WriteLn(`    "-msse2",   -- TODO: Separate gcc options for x64/arm somehow?`)
	tundra.WriteLn(`    { "-O2", "-g"; Config = "*-*-test" },`)
	tundra.WriteLn(`    { "-O0", "-g"; Config = "*-*-debug" },`)
	tundra.WriteLn(`    { "-O3", Config = "*-*-release" },`)
	tundra.WriteLn(`}`)
	tundra.WriteLn(``)
	tundra.WriteLn(`local linux = {`)
	tundra.WriteLn(`    Env = {`)
	tundra.WriteLn(``)
	tundra.WriteLn(`        CCOPTS = {`)
	tundra.WriteLn(`			"-Werror=incompatible-pointer-types",`)
	tundra.WriteLn(`            linux_opts,`)
	tundra.WriteLn(`        },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`        CXXOPTS = {`)
	tundra.WriteLn(`            linux_opts,`)
	tundra.WriteLn(`        },`)
	tundra.WriteLn(`    },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`    ReplaceEnv = {`)
	tundra.WriteLn(`        LD = "c++",`)
	tundra.WriteLn(`		OBJECTROOT = "target",`)
	tundra.WriteLn(`	},`)
	tundra.WriteLn(`}`)
	tundra.WriteLn(``)
	tundra.WriteLn(`-----------------------------------------------------------------------------------------------------------------------`)
	tundra.WriteLn(``)
	tundra.WriteLn(`local win64_opts = {`)
	tundra.WriteLn(`    "/EHsc", "/FS", "/MD", "/W3", "/I.", "/DUNICODE", "/D_UNICODE", "/DWIN32", "/D_CRT_SECURE_NO_WARNINGS",`)
	tundra.WriteLn(`    "\"/DOBJECT_DIR=$(OBJECTDIR:#)\"",`)
	tundra.WriteLn(`    { "/Od"; Config = "*-*-debug" },`)
	tundra.WriteLn(`    { "/O2"; Config = "*-*-release" },`)
	tundra.WriteLn(`}`)
	tundra.WriteLn(``)
	tundra.WriteLn(`local win64 = {`)
	tundra.WriteLn(`    Env = {`)
	tundra.WriteLn(``)
	tundra.WriteLn(`        GENERATE_PDB = "1",`)
	tundra.WriteLn(`        CCOPTS = {`)
	tundra.WriteLn(`            win64_opts,`)
	tundra.WriteLn(`        },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`        CXXOPTS = {`)
	tundra.WriteLn(`            win64_opts,`)
	tundra.WriteLn(`        },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`        OBJCCOM = "meh",`)
	tundra.WriteLn(`    },`)
	tundra.WriteLn(`    ReplaceEnv = {`)
	tundra.WriteLn(`        OBJECTROOT = "target",`)
	tundra.WriteLn(`    },`)
	tundra.WriteLn(`}`)
	tundra.WriteLn(``)
	tundra.WriteLn(`-----------------------------------------------------------------------------------------------------------------------`)
	tundra.WriteLn(``)
	tundra.WriteLn(`Build {`)
	tundra.WriteLn(`    Passes = {`)
	tundra.WriteLn(`        BuildTools = { Name = "BuildTools", BuildOrder = 1 },`)
	tundra.WriteLn(`        GenerateSources = { Name = "GenerateSources", BuildOrder = 2 },`)
	tundra.WriteLn(`    },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`    Units = {`)
	tundra.WriteLn(`        "units.lua",`)
	tundra.WriteLn(`    },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`    Configs = {`)
	tundra.WriteLn(`        Config { Name = "macos-clang", DefaultOnHost = "macosx", Inherit = macosx, Tools = { "clang-osx" } },`)
	tundra.WriteLn(`        Config { Name = "win64-msvc", DefaultOnHost = { "windows" }, Inherit = win64, Tools = { "msvc-vs2019" } },`)
	tundra.WriteLn(`        Config { Name = "linux-gcc", DefaultOnHost = { "linux" }, Inherit = linux, Tools = { "gcc" } },`)
	tundra.WriteLn(`        Config { Name = "linux-clang", DefaultOnHost = { "linux" }, Inherit = linux, Tools = { "clang" } },`)
	tundra.WriteLn(`    },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`    IdeGenerationHints = {`)
	tundra.WriteLn(`        Msvc = {`)
	tundra.WriteLn(`            -- Remap config names to MSVC platform names (affects things like header scanning & debugging)`)
	tundra.WriteLn(`            PlatformMappings = {`)
	tundra.WriteLn(`                ['win64-msvc'] = 'x64',`)
	tundra.WriteLn(`            },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`            -- Remap variant names to MSVC friendly names`)
	tundra.WriteLn(`            VariantMappings = {`)
	tundra.WriteLn(`                ['release']    = 'Release',`)
	tundra.WriteLn(`                ['debug']      = 'Debug',`)
	tundra.WriteLn(`            },`)
	tundra.WriteLn(`        },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`        MsvcSolutions = {`)
	tundra.WriteLn(`            ['libglfw.sln'] = { }`)
	tundra.WriteLn(`        },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`    },`)
	tundra.WriteLn(``)
	tundra.WriteLn(`    -- Variants = { "debug", "test", "release" },`)
	tundra.WriteLn(`    Variants = { "debug", "release" },`)
	tundra.WriteLn(`    SubVariants = { "default", "test" },`)
	tundra.WriteLn(`}`)

	tundra.Close()
	units.Close()

	return nil
}

// IsTundra checks if IDE is requesting a Tundra build file
func IsTundra(DEV string, OS string, ARCH string) bool {
	return strings.ToLower(DEV) == "tundra"
}
