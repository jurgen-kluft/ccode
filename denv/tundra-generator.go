package denv

import (
	"path/filepath"
	"runtime"
	"strings"

	cutils "github.com/jurgen-kluft/ccode/cutils"
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
	units := cutils.NewLineWriter(cutils.IndentModeTabs)

	units.WriteLine(`require "tundra.syntax.glob"`)
	units.WriteLine(`require "tundra.path"`)
	units.WriteLine(`require "tundra.util"`)

	default_unit := ""

	// Get all the projects and write them out
	for _, p := range ws.ProjectList.Values {
		switch p.Type.GetProjectType() {
		case DevConfigTypeExecutable:
			continue
		case DevConfigTypeStaticLibrary:
			units.NewLine()
			units.WriteILine("", "local ", p.Name, "_staticlib = ", "StaticLibrary", "{")
		case DevConfigTypeDynamicLibrary:
			units.NewLine()
			units.WriteILine("", "local ", p.Name, "_sharedlib = ", "SharedLibrary", "{")
		}
		g.writeUnit(units, p, false)
		units.WriteLine("}")

		// Set the default unit to the last library project, in case there is no program this
		// will be the default unit. For example if we only have a static library project.
		default_unit = p.Name + "_staticlib"
	}

	for _, p := range ws.ProjectList.Values {
		switch p.Type.GetProjectType() {
		case DevConfigTypeStaticLibrary:
			continue
		case DevConfigTypeDynamicLibrary:
			continue
		case DevConfigTypeExecutable:
			units.NewLine()
			units.WriteILine("", "local ", p.Name, "_program = ", "Program", "{")
		}
		g.writeUnit(units, p, true)
		units.WriteLine("}")

		default_unit = p.Name + "_program"
	}

	units.WriteILine("", "Default(", default_unit, ")")

	units.NewLine()
	units.WriteToFile(filepath.Join(ws.GenerateAbsPath, "units.lua"))
}

func (g *TundraGenerator) writeUnit(units *cutils.LineWriter, p *Project, isProgram bool) {
	units.WriteILine("+", "Name = ", `"`, p.Name, `",`)
	units.WriteILine("+", "Env = {")

	// Library Paths
	if isProgram {
		//    Libs = {
		//        "user32.lib"; Config = "win32-*",
		//        "ws2_32.lib"; Config = "win32-*",
		//        "gdi32.lib"; Config = "win32-*",
		//    },
		units.WriteILine("++", "LIBPATH = {")
		for _, cfg := range p.Resolved.Configs.Values {
			linkDirs, _, _ := p.BuildLibraryInformation(DevTundra, cfg, p.Workspace.GenerateAbsPath)
			for _, linkDir := range linkDirs.Values {
				units.WriteILine("+++", `{"`, linkDir, `", `, `Config = "`, cfg.Type.Tundra(), `"},`)
			}
		}
		units.WriteILine("++", "},")
	}

	// Compiler Defines

	units.WriteILine("++", `CPPDEFS = {`)
	for _, cfg := range p.Resolved.Configs.Values {
		units.WriteILine("+++", `{`)
		for _, def := range cfg.CppDefines.Values {
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
	for _, pcfg := range p.Resolved.Configs.Values {
		for _, inc := range pcfg.IncludeDirs.Values {
			//path := PathGetRelativeTo(filepath.Join(p.ProjectAbsPath, inc), p.Workspace.GenerateAbsPath)
			path := inc.RelativeTo(p.Workspace.GenerateAbsPath)
			path = strings.Replace(path, "\\", "/", -1)
			signature := path + " | " + pcfg.String()
			if _, ok := history[signature]; !ok {
				units.WriteILine("++", `{"`, path, `", Config = "`, pcfg.Type.Tundra(), `"},`)
				history[signature] = 1
			}
		}
	}
	for _, dp := range p.Dependencies.Values {
		for _, dpcfg := range dp.Resolved.Configs.Values {
			for _, inc := range dpcfg.IncludeDirs.Values {
				//path := PathGetRelativeTo(filepath.Join(dp.ProjectAbsPath, inc), p.Workspace.GenerateAbsPath)
				path := inc.RelativeTo(p.Workspace.GenerateAbsPath)
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
	for _, src := range p.SrcFiles.Values {
		if src.Is_SourceFile() {
			path := p.SrcFiles.GetRelativePath(src, p.Workspace.GenerateAbsPath)
			path = strings.Replace(path, "\\", "/", -1)
			units.WriteILine("++", `"`, path, `",`)
		}
	}
	units.WriteILine("+", "},")

	// Library Dependencies

	units.WriteILine("+", "Depends = {")
	for _, dp := range p.Dependencies.Values {
		switch dp.Type.GetProjectType() {
		case DevConfigTypeStaticLibrary:
			units.WriteILine("++", dp.Name, "_staticlib,")
		case DevConfigTypeDynamicLibrary:
			units.WriteILine("++", dp.Name, "_sharedlib,")
		}
	}
	units.WriteILine("+", "},")

	if isProgram {
		//    Libs = {
		//        "user32.lib"; Config = "win32-*",
		//        "ws2_32.lib"; Config = "win32-*",
		//        "gdi32.lib"; Config = "win32-*",
		//    },

		units.WriteILine("+", "Libs = {")
		for _, cfg := range p.Resolved.Configs.Values {
			_, _, linkLibs := p.BuildLibraryInformation(DevTundra, cfg, p.Workspace.GenerateAbsPath)
			for _, lib := range linkLibs.Values {
				lib = strings.Replace(lib, "\\", "/", -1)
				units.WriteILine("++", `{"`, lib, `"`, `, Config = "`, cfg.Type.Tundra(), `"},`)
			}
		}
		units.WriteILine("+", "},")

		// if the platform is Mac also write out the Frameworks we are using
		if p.Workspace.BuildTarget.OSIsMac() {
			units.WriteILine("+", `Frameworks = {`)
			for _, cfg := range p.Resolved.Configs.Values {
				frameworks := p.BuildFrameworkInformation(cfg)
				for _, framework := range frameworks.Values {
					units.WriteILine("++", `{"`, framework, `"`, `, Config = "`, cfg.Type.Tundra(), `"},`)
				}
			}
			units.WriteILine("+", `},`)
		}
	}
}

func (g *TundraGenerator) generateTundraLua(ws *Workspace) {
	tundra := cutils.NewLineWriter(cutils.IndentModeTabs)
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
	tundra.WriteLine(`            "-std=`, ws.Config.CppStd.String(), `",`)
	if ws.Config.CppAdvanced.IsEnabled() {
		tundra.WriteLine(`    "`, ws.Config.CppAdvanced.Tundra(ws.BuildTarget), `",`)
	}
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
	if ws.Config.CppAdvanced.IsEnabled() {
		tundra.WriteLine(`    "`, ws.Config.CppAdvanced.Tundra(ws.BuildTarget), `",`)
	}
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
	tundra.WriteLine(`    "/std:`, ws.Config.CppStd.String(), `",`)
	if ws.Config.CppAdvanced.IsEnabled() {
		tundra.WriteLine(`    "`, ws.Config.CppAdvanced.Tundra(ws.BuildTarget), `",`)
	}
	tundra.WriteLine(`    "/EHsc", "/FS", "/W3", "/I.", "/DUNICODE", "/D_UNICODE", "/DWIN32", "/D_CRT_SECURE_NO_WARNINGS",`)
	tundra.WriteLine(`    "\"/DOBJECT_DIR=$(OBJECTDIR:#)\"",`)
	tundra.WriteLine(`    { "/Od", "/MDd"; Config = "*-*-debug-*" },`)
	tundra.WriteLine(`    { "/O2", "/MD"; Config = "*-*-release-*" },`)
	tundra.WriteLine(`}`)

	tundra.WriteLine(``)
	tundra.WriteLine(`local win64 = {`)
	tundra.WriteLine(`    Env = {`)
	tundra.WriteLine(`        GENERATE_PDB = "1",`)
	tundra.WriteLine(`        PROGOPTS = {`)
	tundra.WriteLine(`            { "/SUBSYSTEM:CONSOLE"; Config = { "*-*-*-test" } },`)
	tundra.WriteLine(`            { "/DEBUG"; Config = { "*-*-debug-*" } },`)
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
