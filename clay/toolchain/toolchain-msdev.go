package toolchain

import (
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
	corepkg "github.com/jurgen-kluft/ccode/core"
	"github.com/jurgen-kluft/ccode/denv"
)

type WinMsdev struct {
	Name string
	Vars *corepkg.Vars
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Microsoft Visual Studio C/C++ Compiler

type WinMsdevCompiler struct {
	toolChain       *WinMsdev
	buildConfig     denv.BuildConfig
	buildTarget     denv.BuildTarget
	objFilePrefix   string
	objFileSuffix   string
	depFilePrefix   string
	depFileSuffix   string
	cCompilerPath   string
	cCompilerArgs   *corepkg.Arguments
	cppCompilerPath string
	cppCompilerArgs *corepkg.Arguments
	vars            *corepkg.Vars
}

func (t *WinMsdev) NewCompiler(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Compiler {
	objFilePrefix, _ := t.Vars.GetFirst("build.obj.prefix")
	objFileSuffix, _ := t.Vars.GetFirst("build.obj.suffix")
	depFilePrefix, _ := t.Vars.GetFirst("build.dep.prefix")
	depFileSuffix, _ := t.Vars.GetFirst("build.dep.suffix")

	cl := &WinMsdevCompiler{
		toolChain:       t,
		buildConfig:     buildConfig,
		buildTarget:     buildTarget,
		objFilePrefix:   objFilePrefix,
		objFileSuffix:   objFileSuffix,
		depFilePrefix:   depFilePrefix,
		depFileSuffix:   depFileSuffix,
		cCompilerPath:   "",
		cCompilerArgs:   nil,
		cppCompilerPath: "",
		cppCompilerArgs: nil,
		vars:            corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
	}

	return cl
}

func (cl *WinMsdevCompiler) ObjFilepath(srcRelFilepath string) string {
	return cl.objFilePrefix + srcRelFilepath + cl.objFileSuffix
}

func (cl *WinMsdevCompiler) DepFilepath(objRelFilepath string) string {
	return cl.depFilePrefix + objRelFilepath + cl.depFileSuffix
}

func (cl *WinMsdevCompiler) SetupArgs(_defines []string, _includes []string) {
	for i, inc := range _includes {
		if !strings.HasPrefix(inc, "/I") {
			_includes[i] = "/I" + inc
		}
	}
	defines := make([]string, len(_defines)*2)
	for _, def := range _defines {
		if !strings.HasPrefix(def, "/D") {
			defines = append(defines, "/D")
			defines = append(defines, def)
		} else {
			defines = append(defines, "/D")
			defines = append(defines, strings.TrimPrefix(def, "/D"))
		}
	}
	cl.vars.Set("build.includes", _includes...)
	cl.vars.Set("build.defines", defines...)

	cl.cCompilerPath = ""
	cl.cCompilerArgs = corepkg.NewArguments(0)
	if c_compiler_args, ok := cl.toolChain.Vars.Get(`recipe.c.pattern`); ok {
		cl.cCompilerPath = c_compiler_args[0]
		cl.cCompilerArgs.Args = c_compiler_args[1:]

		cl.cCompilerPath = cl.toolChain.Vars.FinalResolveString(cl.cCompilerPath, " ", cl.vars)
		cl.cCompilerArgs.Args = cl.toolChain.Vars.FinalResolveArray(cl.cCompilerArgs.Args, cl.vars)

		cl.cCompilerArgs.Args = slices.DeleteFunc(cl.cCompilerArgs.Args, func(s string) bool { return strings.TrimSpace(s) == "" })
	}

	cl.cppCompilerPath = ""
	cl.cppCompilerArgs = corepkg.NewArguments(0)
	if cpp_compiler_args, ok := cl.toolChain.Vars.Get(`recipe.cpp.pattern`); ok {
		cl.cppCompilerPath = cpp_compiler_args[0]
		cl.cppCompilerArgs.Args = cpp_compiler_args[1:]

		cl.cppCompilerPath = cl.toolChain.Vars.FinalResolveString(cl.cppCompilerPath, " ", cl.vars)
		cl.cppCompilerArgs.Args = cl.toolChain.Vars.FinalResolveArray(cl.cppCompilerArgs.Args, cl.vars)

		cl.cppCompilerArgs.Args = slices.DeleteFunc(cl.cppCompilerArgs.Args, func(s string) bool { return strings.TrimSpace(s) == "" })
	}
}

func (cl *WinMsdevCompiler) Compile(sourceAbsFilepaths []string, objRelFilepaths []string) error {
	for _, sourceAbsFilepath := range sourceAbsFilepaths {
		var compilerPath string
		var compilerArgs []string
		if strings.HasSuffix(sourceAbsFilepath, ".c") {
			compilerPath = cl.cCompilerPath
			compilerArgs = cl.cCompilerArgs.Args
		} else {
			compilerPath = cl.cppCompilerPath
			compilerArgs = cl.cppCompilerArgs.Args
		}

		compilerArgs = append(compilerArgs, sourceAbsFilepath)

		compilerArgs = slices.DeleteFunc(compilerArgs, func(s string) bool { return strings.TrimSpace(s) == "" })
		cmd := exec.Command(compilerPath, compilerArgs...)

		// TODO we could make one single environment and cache it somewhere so that we can reuse it, perhaps
		//      as a member in toolchain? The Archiver and Linker will need it as well.

		path := make([]string, 0)
		path = append(path, `C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Tools\MSVC\14.44.35207\bin\HostX64\x64`)
		path = append(path, `C:\Program Files (x86)\Windows Kits\10\bin\10.0.26100.0\\x64`)
		path = append(path, `C:\Program Files (x86)\Windows Kits\10\bin\\x64`)

		cmd.Env = cmd.Environ()
		for i, env := range cmd.Env {
			if env, found := strings.CutPrefix(env, "Path="); found {
				cmd.Env[i] = "Path=" + strings.Join(path, ";") + ";" + env
				break
			}
		}

		corepkg.LogInfof("Compiling (%s) %s", cl.buildConfig.String(), filepath.Base(sourceAbsFilepath))

		out, err := cmd.CombinedOutput()
		if err != nil {
			corepkg.LogInfof("Compile failed, output:\n%s", string(out))
			return corepkg.LogErrorf(err, "Compiling failed")
		}
	}
	return nil
}

func (ms *WinMsdev) NewArchiver(a ArchiverType, config denv.BuildConfig, target denv.BuildTarget) Archiver {
	return nil
}

func (ms *WinMsdev) NewLinker(config denv.BuildConfig, target denv.BuildTarget) Linker {
	return nil
}

func (ms *WinMsdev) NewBurner(config denv.BuildConfig, target denv.BuildTarget) Burner {
	return nil
}

func (ms *WinMsdev) NewDependencyTracker(dirpath string) deptrackr.FileTrackr {
	return deptrackr.LoadJsonFileTrackr(filepath.Join(dirpath, "deptrackr"))
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Toolchain for Visual Studio on Windows

func NewWinMsdev(vars *corepkg.Vars, projectName string, buildPath string, arch string) (t *WinMsdev, err error) {

	vars.Set("project.name", projectName)
	vars.Set("build.path", buildPath)
	vars.Set("build.arch", arch)

	return &WinMsdev{
		Name: "msdev",
		Vars: vars,
	}, nil
}
