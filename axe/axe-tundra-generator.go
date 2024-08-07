package axe

import (
	"path/filepath"
	"runtime"
	"strings"
)

type TundraGenerator struct {
	Workspace  *Workspace
	VcxProjCpu string
}

func NewTundraGenerator(ws *Workspace) *TundraGenerator {
	g := &TundraGenerator{
		Workspace: ws,
	}

	return g
}

func (g *TundraGenerator) Generate() {
	g.generateUnitsLua(g.Workspace)
	g.generateTundraLua(g.Workspace)
}

func (g *TundraGenerator) generateUnitsLua(ws *Workspace) {
	units := NewLineWriter(IndentModeTabs)

	units.WriteLine(`require "tundra.syntax.glob"`)
	units.WriteLine(`require "tundra.path"`)
	units.WriteLine(`require "tundra.util"`)

	// Sort the projects by their dependencies using topological sort
	ws.ProjectList.TopoSort()

	// Get all the projects and write them out
	for _, p := range ws.ProjectList.Values {
		switch p.Type {
		case ProjectTypeCppLib, ProjectTypeCLib:
			units.NewLine()
			units.WriteILine("", "local ", p.Name, "_staticlib = ", "StaticLibrary", "{")
		case ProjectTypeCppDll, ProjectTypeCDll:
			units.NewLine()
			units.WriteILine("", "local ", p.Name, "_sharedlib = ", "SharedLibrary", "{")
		case ProjectTypeCppExe, ProjectTypeCExe:
			continue
		}
		g.writeUnit(units, p)
		units.WriteLine("}")
	}

	for _, p := range ws.ProjectList.Values {
		switch p.Type {
		case ProjectTypeCppLib, ProjectTypeCLib:
			continue
		case ProjectTypeCppDll, ProjectTypeCDll:
			continue
		case ProjectTypeCppExe, ProjectTypeCExe:
			units.NewLine()
			units.WriteILine("", "local ", p.Name, "_program = ", "Program", "{")
		}
		g.writeUnit(units, p)
		units.WriteLine("}")

		units.WriteILine("", "Default(", p.Name, "_program)")
	}

	units.NewLine()
	units.WriteToFile(filepath.Join(ws.GenerateAbsPath, "units.lua"))
}

func (g *TundraGenerator) writeUnit(units *LineWriter, p *Project) {
	units.WriteILine("+", "Name = ", `"`, p.Name, `",`)
	units.WriteILine("+", "Env = {")

	// Compiler Defines

	units.WriteILine("++", `CPPDEFS = {`)
	for _, cfg := range p.Configs.Values {
		units.WriteILine("+++", `{`)
		for _, def := range cfg.CppDefines.FinalDict.Values {
			escapedDef := g.escapeString(def)
			units.WriteILine("++++", `"`, escapedDef, `",`)
		}
		units.WriteILine("++++", `Config = "`, cfg.Type.Tundra(), `"`)
		units.WriteILine("+++", `},`)
	}
	units.WriteILine("+++", `{ "TARGET_PC", Config = "win64-*-*-*" },`)
	units.WriteILine("+++", `{ "TARGET_MAC", Config = "macos-*-*-*" },`)
	units.WriteILine("+++", `{ "TARGET_TEST", Config = "*-*-*-test" },`)
	units.WriteILine("++", `},`)
	units.WriteILine("+", `},`)

	// Include Directories

	units.WriteILine("+", `Includes = {`)
	history := make(map[string]int)
	for _, pcfg := range p.Configs.Values {
		for _, inc := range pcfg.IncludeDirs.FinalDict.Values {
			path := PathGetRel(filepath.Join(p.ProjectAbsPath, inc), p.Workspace.GenerateAbsPath)
			path = strings.Replace(path, "\\", "/", -1)
			signature := path + " | " + pcfg.String()
			if _, ok := history[signature]; !ok {
				units.WriteILine("++", `{"`, path, `", Config = "`, pcfg.Type.Tundra(), `"},`)
				history[signature] = 1
			}
		}
	}
	for _, dp := range p.Dependencies.Values {
		for _, dpcfg := range dp.Configs.Values {
			for _, inc := range dpcfg.IncludeDirs.FinalDict.Values {
				path := PathGetRel(filepath.Join(dp.ProjectAbsPath, inc), p.Workspace.GenerateAbsPath)
				path = strings.Replace(path, "\\", "/", -1)
				signature := path + " | " + dpcfg.String()
				if _, ok := history[signature]; !ok {
					units.WriteILine("++", `{"`, path, `", Config = "`, dpcfg.Type.Tundra(), `"},`)
					history[signature] = 1
				}
			}
		}
	}
	units.WriteILine("+", `},`)

	// Source Files

	units.WriteILine("+", `Sources = {`)
	for _, src := range p.FileEntries.Values {
		path := p.FileEntries.GetRelativePath(src, p.Workspace.GenerateAbsPath)
		path = strings.Replace(path, "\\", "/", -1)
		units.WriteILine("++", `"`, path, `",`)
	}
	units.WriteILine("+", "},")

	// Library Dependencies

	units.WriteILine("+", "Depends = {")
	for _, dp := range p.Dependencies.Values {
		switch dp.Type {
		case ProjectTypeCppLib, ProjectTypeCLib:
			units.WriteILine("++", dp.Name, "_staticlib,")
		case ProjectTypeCppDll, ProjectTypeCDll:
			units.WriteILine("++", dp.Name, "_sharedlib,")
		}
	}
	units.WriteILine("+", "},")

	// Framework Dependencies

	// if the platform is Mac also write out the Frameworks we are using
	if p.Workspace.MakeTarget.OSIsMac() {
		units.WriteILine("+", `Frameworks = {`)
		units.WriteILine("++", `{ "Cocoa" },`)
		units.WriteILine("++", `{ "Metal" },`)
		units.WriteILine("++", `{ "OpenGL" },`)
		units.WriteILine("++", `{ "IOKit" },`)
		units.WriteILine("++", `{ "Carbon" },`)
		units.WriteILine("++", `{ "CoreVideo" },`)
		units.WriteILine("++", `{ "QuartzCore" },`)
		units.WriteILine("+", `},`)
	}
}

func (g *TundraGenerator) generateTundraLua(ws *Workspace) {
	tundra := NewLineWriter(IndentModeTabs)
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
	tundra.WriteLine(`    { "-O2", "-g"; Config = "*-*-*-test" },`)
	tundra.WriteLine(`    { "-O0", "-g"; Config = "*-*-debug-*" },`)
	tundra.WriteLine(`    { "-O3", "-g"; Config = "*-*-release-*" },`)
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
	tundra.WriteLine(`		OBJECTROOT = "../tundra",`)
	tundra.WriteLine(`	},`)
	tundra.WriteLine(``)
	tundra.WriteLine(`    Frameworks = {`)
	tundra.WriteLine(`        { "Foundation" },`)
	tundra.WriteLine(`        { "Cocoa" },`)
	tundra.WriteLine(`        { "Carbon" },`)
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
	tundra.WriteLine(`    { "-O2", "-g"; Config = "*-*-*-test" },`)
	tundra.WriteLine(`    { "-O0", "-g"; Config = "*-*-debug-*" },`)
	tundra.WriteLine(`    { "-O3", Config = "*-*-release-*" },`)
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
	tundra.WriteLine(`        OBJECTROOT = "../tundra",`)
	tundra.WriteLine(`	},`)
	tundra.WriteLine(`}`)
	tundra.WriteLine(``)
	tundra.WriteLine(`-----------------------------------------------------------------------------------------------------------------------`)
	tundra.WriteLine(``)
	tundra.WriteLine(`local win64_opts = {`)
	tundra.WriteLine(`    "/EHsc", "/FS", "/MD", "/W3", "/I.", "/DUNICODE", "/D_UNICODE", "/DWIN32", "/D_CRT_SECURE_NO_WARNINGS",`)
	tundra.WriteLine(`    "\"/DOBJECT_DIR=$(OBJECTDIR:#)\"",`)
	tundra.WriteLine(`    { "/Od"; Config = "*-*-debug-*" },`)
	tundra.WriteLine(`    { "/O2"; Config = "*-*-release-*" },`)
	tundra.WriteLine(`}`)
	tundra.WriteLine(``)
	tundra.WriteLine(`local win64 = {`)
	tundra.WriteLine(`    Env = {`)
	tundra.WriteLine(`        GENERATE_PDB = "1",`)
	tundra.WriteLine(`        PROGOPTS = {`)
	tundra.WriteLine(`            { "/SUBSYSTEM:CONSOLE", "/DEBUG"; Config = { "*-*-*-test" } },`)
	tundra.WriteLine(`        },`)
	tundra.WriteLine(`        CCOPTS = {`)
	tundra.WriteLine(`            win64_opts,`)
	tundra.WriteLine(`        },`)
	tundra.WriteLine(`        CXXOPTS = {`)
	tundra.WriteLine(`            win64_opts,`)
	tundra.WriteLine(`        },`)
	tundra.WriteLine(`        OBJCCOM = "meh",`)
	tundra.WriteLine(`    },`)
	tundra.WriteLine(`	  ReplaceEnv = {`)
	tundra.WriteLine(`	      OBJECTROOT = "../tundra",`)
	tundra.WriteLine(`	  },`)
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
	tundra.WriteLine(`        Config { Name = "macos-clang", DefaultOnHost = "macosx", Inherit = macosx, Tools = { "clang-osx" } },`)
	tundra.WriteLine(`        Config { Name = "win64-msvc", DefaultOnHost = "windows", Inherit = win64, Tools = { "msvc-vs2022" } },`)
	tundra.WriteLine(`        Config { Name = "linux-gcc", DefaultOnHost = "linux", Inherit = linux, Tools = { "gcc" } },`)
	tundra.WriteLine(`        Config { Name = "linux-clang", DefaultOnHost = "linux", Inherit = linux, Tools = { "clang" } },`)
	tundra.WriteLine(`    },`)
	tundra.WriteLine(``)
	tundra.WriteLine(`    Variants = { "debug", "release" },`)
	tundra.WriteLine(`    SubVariants = { "test" },`)
	tundra.WriteLine(`}`)

	tundrafilepath := filepath.Join(ws.GenerateAbsPath, "tundra.lua")
	tundra.WriteToFile(tundrafilepath)
}

func (g *TundraGenerator) escapeString(s string) string {
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "\"", "\\\"", -1)
	s = strings.Replace(s, "-", "_", -1)
	return s
}
