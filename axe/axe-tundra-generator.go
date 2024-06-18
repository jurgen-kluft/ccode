package axe

import (
	"path/filepath"
	"runtime"
)

type TundraGenerator struct {
	LastGenId  UUID
	Workspace  *Workspace
	VcxProjCpu string
}

func NewTundraGenerator(ws *Workspace) *TundraGenerator {
	g := &TundraGenerator{
		LastGenId: GenerateUUID(),
		Workspace: ws,
	}

	return g
}

func (g *TundraGenerator) Generate() {

}

func (g *TundraGenerator) generateUnitsLua(ws *Workspace) {
	units := NewLineWriter()

	units.WriteLine(`require "tundra.syntax.glob"`)
	units.WriteLine(`require "tundra.path"`)
	units.WriteLine(`require "tundra.util"`)

	// Get all the library projects and write them out

}

func (g *TundraGenerator) writeCppLibrary(units *LineWriter, p *Project) {

	projectType := ""
	switch p.Type {
	case ProjectTypeCppLib, ProjectTypeCLib:
		projectType = "StaticLibrary"
	case ProjectTypeCppDll, ProjectTypeCDll:
		projectType = "SharedLibrary"
	case ProjectTypeCppExe, ProjectTypeCExe:
		projectType = "Program"
	}

	units.NewLine()
	units.WriteILine("local ", p.Name, "_unit = ", projectType)
	units.WriteILine("+", "Name = ", p.Name, ",")
	units.WriteILine("+", "Env = {")
	units.WriteILine(`++`, `CPPPATH = {`)
	units.WriteILine(`+++`, `${${Name}:SOURCE_DIR}",`)
	units.WriteILine(`+++`, `${${Name}:INCLUDE_DIRS},`)

	// for _, depDep := range dep.Dependencies {
	// 	units.WriteLine(`+++${` + depDep.Name + `:INCLUDE_DIRS},`)
	// 	units.WriteLine(`+++"${` + depDep.Name + `:SOURCE_DIR}",`)
	// }

	units.WriteLine(`++},`)
	units.WriteLine(`++CPPDEFS = {`)
	// units.WriteLine( `+++{ "TARGET_DEBUG", Config = "*-*-debug" },`)
	// units.WriteLine( `+++{ "TARGET_RELEASE", Config = "*-*-release" },`)

	// for _, cfg := range dep.Platform.Configs {
	// 	units.WriteLine(`+++{`)
	// 	for _, def := range cfg.Defines.Items {
	// 		units.WriteLine(`++++"` + def + `",`)
	// 	}
	// 	units.WriteLine(`++++Config = "*-*-` + strings.ToLower(strings.ToLower(cfg.Config)) + `" `)
	// 	units.WriteLine(`+++},`)
	// }

	units.WriteLine(`+++{ "TARGET_MAC", Config = "macos-*-*" },`)
	units.WriteLine(`+++{ "TARGET_TEST", Config = "*-*-test" },`)
	units.WriteLine(`++},`)
	units.WriteLine(`+},`)
	units.WriteLine(`+Includes = {`)
	units.WriteLine(`++${${Name}:INCLUDE_DIRS},`)

	// for _, depDep := range dep.Dependencies {
	// 	units.WriteLine(`++${` + depDep.Name + `:INCLUDE_DIRS},`)
	// }

	units.WriteLine(`+},`)
	units.WriteLine(`+Sources = {`)
	units.WriteLine(`++${${Name}:SOURCE_FILES}`)
	units.WriteLine(`+},`)
	units.WriteLine(`}`)
	units.WriteLine("")

	// replacer.ReplaceInLines("${SOURCE_FILES}", "${"+dep.Name+":SOURCE_FILES}", dependency)
	// replacer.ReplaceInLines("${SOURCE_DIR}", "${"+dep.Name+":SOURCE_DIR}", dependency)

	// configitems := map[string]items.List{
	// 	"INCLUDE_DIRS": items.NewList("${"+dep.Name+":INCLUDE_DIRS}", ",", ""),
	// }

	// for configitem, defaults := range configitems {
	// 	varkeystr := fmt.Sprintf("${%s}", configitem)
	// 	varlist := defaults.Copy()

	// 	for _, depDep := range dep.Dependencies {
	// 		varkey := fmt.Sprintf("%s:%s", depDep.Name, configitem)
	// 		varitem := variables.GetVar(varkey)
	// 		if len(varitem) > 0 {
	// 			varlist = varlist.Add(varitem)
	// 		}
	// 	}
	// 	varset := items.ListToSet(varlist)
	// 	replacer.InsertInLines(varkeystr, varset.String(), "", dependency)
	// 	replacer.ReplaceInLines(varkeystr, "", dependency)
	// }

	// replacer.ReplaceInLines("${Name}", dep.Name, dependency)
	// variables.ReplaceInLines(replacer, dependency)

	// units.WriteLns(dependency)

}

func (g *TundraGenerator) generateTundraLua(ws *Workspace) {
	tundra := NewLineWriter()
	tundra.WriteLine(`local native = require('tundra.native')`)
	tundra.WriteLine(``)
	tundra.WriteLine(`-----------------------------------------------------------------------------------------------------------------------`)
	tundra.WriteLine(``)
	tundra.WriteLine(`local mac_opts = {`)
	tundra.WriteLine(`    "-I.",`)
	tundra.WriteLine(`    "-Wno-new-returns-null",`)
	tundra.WriteLine(`    "-Wno-missing-braces",`)
	tundra.WriteLine(`    "-Wno-c++11-compat-deprecated-writable-strings",`)
	tundra.WriteLine(`    "-Wno-null-dereference",`)
	tundra.WriteLine(`    "-Wno-format",`)
	tundra.WriteLine(`    "-fno-strict-aliasing",`)
	tundra.WriteLine(`    "-fno-omit-frame-pointer",`)
	tundra.WriteLine(`	"-Wno-write-strings",`)
	tundra.WriteLine(`    "-Wno-array-bounds",`)
	tundra.WriteLine(`    "-Wno-attributes",`)
	tundra.WriteLine(`    "-Wno-unused-value",`)
	tundra.WriteLine(`    "-Wno-unused-function",`)
	tundra.WriteLine(`    "-Wno-unused-variable",`)
	tundra.WriteLine(`    "-Wno-unused-result",`)
	tundra.WriteLine(`    { "-O2", "-g"; Config = "*-*-test" },`)
	tundra.WriteLine(`    { "-O0", "-g"; Config = "*-*-debug" },`)
	tundra.WriteLine(`    { "-O3", "-g"; Config = "*-*-release" },`)
	tundra.WriteLine(`}`)
	tundra.WriteLine(``)
	tundra.WriteLine(`-----------------------------------------------------------------------------------------------------------------------`)
	tundra.WriteLine(``)
	tundra.WriteLine(`local macosx = {`)
	tundra.WriteLine(`    Env = {`)
	tundra.WriteLine(`        CCOPTS =  {`)
	tundra.WriteLine(`            mac_opts,`)
	tundra.WriteLine(`        },`)
	tundra.WriteLine(``)
	tundra.WriteLine(`        CXXOPTS = {`)
	tundra.WriteLine(`            mac_opts,`)
	tundra.WriteLine(`            "-std=c++14",`)
	if runtime.GOARCH == "amd64" {
		tundra.WriteLine(`			"-arch x86_64",`)
	} else if runtime.GOARCH == "arm64" {
		tundra.WriteLine(`			"-arch arm64",`)
	}
	tundra.WriteLine(`        },`)
	tundra.WriteLine(``)
	tundra.WriteLine(`        SHLIBOPTS = {`)
	tundra.WriteLine(`			"-lstdc++",`)
	tundra.WriteLine(`			{ "-fsanitize=address"; Config = "*-*-debug-asan" },`)
	tundra.WriteLine(`		},`)
	tundra.WriteLine(``)
	tundra.WriteLine(`        PROGCOM = {`)
	tundra.WriteLine(`			"-lstdc++",`)
	tundra.WriteLine(`			{ "-fsanitize=address"; Config = "*-*-debug-asan" },`)
	tundra.WriteLine(`		},`)
	tundra.WriteLine(`    },`)
	tundra.WriteLine(``)
	tundra.WriteLine(`	ReplaceEnv = {`)
	tundra.WriteLine(`		OBJECTROOT = "../../target",`)
	tundra.WriteLine(`	},`)
	tundra.WriteLine(``)
	tundra.WriteLine(`    Frameworks = {`)
	tundra.WriteLine(`        { "Cocoa" },`)
	tundra.WriteLine(`        { "Metal" },`)
	tundra.WriteLine(`        { "OpenGL" },`)
	tundra.WriteLine(`        { "IOKit" },`)
	tundra.WriteLine(`        { "AppKit" },`)
	tundra.WriteLine(`        { "CoreVideo" },`)
	tundra.WriteLine(`        { "QuartzCore" },`)
	tundra.WriteLine(`    },`)
	tundra.WriteLine(`}`)
	tundra.WriteLine(``)
	tundra.WriteLine(`-----------------------------------------------------------------------------------------------------------------------`)
	tundra.WriteLine(``)
	tundra.WriteLine(`local linux_opts = {`)
	tundra.WriteLine(`    "-I.",`)
	tundra.WriteLine(`    "-Wno-new-returns-null",`)
	tundra.WriteLine(`    "-Wno-missing-braces",`)
	tundra.WriteLine(`    "-Wno-c++11-compat-deprecated-writable-strings",`)
	tundra.WriteLine(`    "-Wno-null-dereference",`)
	tundra.WriteLine(`    "-Wno-format",`)
	tundra.WriteLine(`    "-fno-strict-aliasing",`)
	tundra.WriteLine(`    "-fno-omit-frame-pointer",`)
	tundra.WriteLine(`	"-Wno-write-strings",`)
	tundra.WriteLine(`    "-Wno-array-bounds",`)
	tundra.WriteLine(`    "-Wno-attributes",`)
	tundra.WriteLine(`    "-Wno-unused-value",`)
	tundra.WriteLine(`    "-Wno-unused-function",`)
	tundra.WriteLine(`    "-Wno-unused-variable",`)
	tundra.WriteLine(`    "-Wno-unused-result",`)
	tundra.WriteLine(`    "-DOBJECT_DIR=\\\"$(OBJECTDIR)\\\"",`)
	tundra.WriteLine(`    "-I$(OBJECTDIR)",`)
	tundra.WriteLine(`    "-Wall",`)
	tundra.WriteLine(`    "-fPIC",`)
	tundra.WriteLine(`    "-msse2",   -- TODO: Separate gcc options for x64/arm somehow?`)
	tundra.WriteLine(`    { "-O2", "-g"; Config = "*-*-test" },`)
	tundra.WriteLine(`    { "-O0", "-g"; Config = "*-*-debug" },`)
	tundra.WriteLine(`    { "-O3", Config = "*-*-release" },`)
	tundra.WriteLine(`}`)
	tundra.WriteLine(``)
	tundra.WriteLine(`local linux = {`)
	tundra.WriteLine(`    Env = {`)
	tundra.WriteLine(``)
	tundra.WriteLine(`        CCOPTS = {`)
	tundra.WriteLine(`			"-Werror=incompatible-pointer-types",`)
	tundra.WriteLine(`            linux_opts,`)
	tundra.WriteLine(`        },`)
	tundra.WriteLine(``)
	tundra.WriteLine(`        CXXOPTS = {`)
	tundra.WriteLine(`            linux_opts,`)
	tundra.WriteLine(`        },`)
	tundra.WriteLine(`    },`)
	tundra.WriteLine(``)
	tundra.WriteLine(`    ReplaceEnv = {`)
	tundra.WriteLine(`        LD = "c++",`)
	tundra.WriteLine(`		OBJECTROOT = "target",`)
	tundra.WriteLine(`	},`)
	tundra.WriteLine(`}`)
	tundra.WriteLine(``)
	tundra.WriteLine(`-----------------------------------------------------------------------------------------------------------------------`)
	tundra.WriteLine(``)
	tundra.WriteLine(`local win64_opts = {`)
	tundra.WriteLine(`    "/EHsc", "/FS", "/MD", "/W3", "/I.", "/DUNICODE", "/D_UNICODE", "/DWIN32", "/D_CRT_SECURE_NO_WARNINGS",`)
	tundra.WriteLine(`    "\"/DOBJECT_DIR=$(OBJECTDIR:#)\"",`)
	tundra.WriteLine(`    { "/Od"; Config = "*-*-debug" },`)
	tundra.WriteLine(`    { "/O2"; Config = "*-*-release" },`)
	tundra.WriteLine(`}`)
	tundra.WriteLine(``)
	tundra.WriteLine(`local win64 = {`)
	tundra.WriteLine(`    Env = {`)
	tundra.WriteLine(``)
	tundra.WriteLine(`        GENERATE_PDB = "1",`)
	tundra.WriteLine(`        CCOPTS = {`)
	tundra.WriteLine(`            win64_opts,`)
	tundra.WriteLine(`        },`)
	tundra.WriteLine(``)
	tundra.WriteLine(`        CXXOPTS = {`)
	tundra.WriteLine(`            win64_opts,`)
	tundra.WriteLine(`        },`)
	tundra.WriteLine(``)
	tundra.WriteLine(`        OBJCCOM = "meh",`)
	tundra.WriteLine(`    },`)
	tundra.WriteLine(`    ReplaceEnv = {`)
	tundra.WriteLine(`        OBJECTROOT = "target",`)
	tundra.WriteLine(`    },`)
	tundra.WriteLine(`}`)
	tundra.WriteLine(``)
	tundra.WriteLine(`-----------------------------------------------------------------------------------------------------------------------`)
	tundra.WriteLine(``)
	tundra.WriteLine(`Build {`)
	tundra.WriteLine(`    Passes = {`)
	tundra.WriteLine(`        BuildTools = { Name = "BuildTools", BuildOrder = 1 },`)
	tundra.WriteLine(`        GenerateSources = { Name = "GenerateSources", BuildOrder = 2 },`)
	tundra.WriteLine(`    },`)
	tundra.WriteLine(``)
	tundra.WriteLine(`    Units = {`)
	tundra.WriteLine(`        "units.lua",`)
	tundra.WriteLine(`    },`)
	tundra.WriteLine(``)
	tundra.WriteLine(`    Configs = {`)
	tundra.WriteLine(`        Config { Name = "macos-clang", DefaultOnHost = { "macosx" }, Inherit = macosx, Tools = { "clang-osx" } },`)
	tundra.WriteLine(`        Config { Name = "win64-msvc", DefaultOnHost = { "windows" }, Inherit = win64, Tools = { "msvc-vs2019" } },`)
	tundra.WriteLine(`        Config { Name = "linux-gcc", DefaultOnHost = { "linux" }, Inherit = linux, Tools = { "gcc" } },`)
	tundra.WriteLine(`        Config { Name = "linux-clang", DefaultOnHost = { "linux" }, Inherit = linux, Tools = { "clang" } },`)
	tundra.WriteLine(`    },`)
	tundra.WriteLine(``)
	tundra.WriteLine(`    -- Variants = { "debug", "test", "release" },`)
	tundra.WriteLine(`    Variants = { "debug", "release" },`)
	tundra.WriteLine(`    SubVariants = { "default", "test" },`)
	tundra.WriteLine(`}`)

	tundrafilepath := filepath.Join(ws.GenerateAbsPath, "tundra", "tundra.lua")
	tundra.WriteToFile(tundrafilepath)
}
