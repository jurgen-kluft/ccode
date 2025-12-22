package toolchain

import (
	"os"
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

	// Environment variables for the toolchain processes, these are environment variables necessary
	// to configure Microsoft Visual Studio command line tools.
	Env []string
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// File Commander
func (t *WinMsdev) NewFileCommander(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) FileCommander {
	return &BasicFileCommander{}
}

func (t *WinMsdev) ChangeFileExtension(_filepath string, newExt string) string {
	ext := filepath.Ext(_filepath)
	return strings.TrimSuffix(_filepath, ext) + newExt
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
	return srcRelFilepath + cl.objFileSuffix
}

func (cl *WinMsdevCompiler) DepFilepath(objRelFilepath string) string {
	return objRelFilepath + cl.depFileSuffix
}

func (cl *WinMsdevCompiler) SetupArgs(projectName string, buildPath string, _defines []string, _includes []string) {
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

		// Append program database file argument
		cl.cCompilerArgs.Args = append(cl.cCompilerArgs.Args, "/Fd:"+buildPath+string(filepath.Separator)+projectName+".pdb")
		// Append source dependencies argument
		cl.cCompilerArgs.Args = append(cl.cCompilerArgs.Args, "/sourceDependencies")
		cl.cCompilerArgs.Args = append(cl.cCompilerArgs.Args, buildPath)
	}

	cl.cppCompilerPath = ""
	cl.cppCompilerArgs = corepkg.NewArguments(0)
	if cpp_compiler_args, ok := cl.toolChain.Vars.Get(`recipe.cpp.pattern`); ok {
		cl.cppCompilerPath = cpp_compiler_args[0]
		cl.cppCompilerArgs.Args = cpp_compiler_args[1:]

		cl.cppCompilerPath = cl.toolChain.Vars.FinalResolveString(cl.cppCompilerPath, " ", cl.vars)
		cl.cppCompilerArgs.Args = cl.toolChain.Vars.FinalResolveArray(cl.cppCompilerArgs.Args, cl.vars)

		cl.cppCompilerArgs.Args = slices.DeleteFunc(cl.cppCompilerArgs.Args, func(s string) bool { return strings.TrimSpace(s) == "" })

		// Append program database file argument
		cl.cppCompilerArgs.Args = append(cl.cppCompilerArgs.Args, "/Fd:"+buildPath+string(filepath.Separator)+projectName+".pdb")
	}
}

func (cl *WinMsdevCompiler) Compile(sourceAbsFilepaths []string, objRelFilepaths []string) ([]bool, bool) {
	errors := 0
	compiled := make([]bool, len(sourceAbsFilepaths))
	for s, sourceAbsFilepath := range sourceAbsFilepaths {
		var compilerPath string
		var compilerArgs []string
		if strings.HasSuffix(sourceAbsFilepath, ".c") {
			compilerPath = cl.cCompilerPath
			compilerArgs = cl.cCompilerArgs.Args
		} else {
			compilerPath = cl.cppCompilerPath
			compilerArgs = cl.cppCompilerArgs.Args
		}

		compilerArgs = append(compilerArgs, "/sourceDependencies")
		compilerArgs = append(compilerArgs, cl.toolChain.ChangeFileExtension(objRelFilepaths[s], ".json"))

		compilerArgs = append(compilerArgs, "/Fo"+objRelFilepaths[s])

		compilerArgs = append(compilerArgs, sourceAbsFilepath)
		compilerArgs = slices.DeleteFunc(compilerArgs, func(s string) bool { return strings.TrimSpace(s) == "" })

		//fmt.Printf("Compiler path %s and args %v\n", compilerPath, compilerArgs)

		cmd := exec.Command(compilerPath, compilerArgs...)
		cmd.Env = cl.toolChain.Env

		// path := make([]string, 0)
		// path = append(path, `C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Tools\MSVC\14.44.35207\bin\HostX64\x64`)
		// path = append(path, `C:\Program Files (x86)\Windows Kits\10\bin\10.0.26100.0\\x64`)
		// path = append(path, `C:\Program Files (x86)\Windows Kits\10\bin\\x64`)

		// cmd.Env = cmd.Environ()
		// for i, env := range cmd.Env {
		// 	if env, found := strings.CutPrefix(env, "Path="); found {
		// 		cmd.Env[i] = "Path=" + strings.Join(path, ";") + ";" + env
		// 		break
		// 	}
		// }

		corepkg.LogInfof("Compiling (%s) %s", cl.buildConfig.String(), filepath.Base(sourceAbsFilepath))

		out, err := cmd.CombinedOutput()
		if err != nil {
			corepkg.LogInfof("Compile failed, output:\n%s", string(out))
			compiled[s] = false
			errors = errors + 1
		} else {
			//if len(out) > 0 {
			//	corepkg.LogInfof("Compile output:\n%s", string(out))
			//}
			compiled[s] = true
		}
	}
	return compiled, errors == 0
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Microsoft Visual Studio Archiver

type WinMsdevArchiver struct {
	toolChain   *WinMsdev
	buildConfig denv.BuildConfig
	buildTarget denv.BuildTarget
	arPath      string
	arArgs      *corepkg.Arguments
}

func (t *WinMsdev) NewArchiver(at ArchiverType, buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Archiver {
	args := corepkg.NewArguments(512)
	switch at {
	case ArchiverTypeStatic:
		return &WinMsdevArchiver{toolChain: t, buildConfig: buildConfig, buildTarget: buildTarget, arArgs: args}
	case ArchiverTypeDynamic:
		return &WinMsdevArchiver{toolChain: t, buildConfig: buildConfig, buildTarget: buildTarget, arArgs: args}
	}
	return nil
}

func (t *WinMsdevArchiver) LibFilepath(_filepath string) string {
	filename := corepkg.PathFilename(_filepath, true)
	dirpath := corepkg.PathDirname(_filepath)
	return filepath.Join(dirpath, filename+".lib") // The file extension for the archive on Windows is typically ".lib"
}

func (t *WinMsdevArchiver) SetupArgs() {
	if archiver_args, ok := t.toolChain.Vars.Get(`recipe.lib.pattern`); ok {
		t.arPath = archiver_args[0]
		t.arArgs = corepkg.NewArguments(0)
		t.arArgs.Args = archiver_args[1:]
		t.arPath = t.toolChain.Vars.FinalResolveString(t.arPath, " ")
		t.arArgs.Args = t.toolChain.Vars.FinalResolveArray(t.arArgs.Args)

		t.arArgs.Args = slices.DeleteFunc(t.arArgs.Args, func(s string) bool { return strings.TrimSpace(s) == "" })
	}
}

func (t *WinMsdevArchiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {
	archiverPath := t.arPath
	archiverArgs := t.arArgs.Args

	// TODO would like this to be part of the resolve step
	archiverArgs = append(archiverArgs, "/OUT:"+outputArchiveFilepath)
	archiverArgs = append(archiverArgs, inputObjectFilepaths...)

	cmd := exec.Command(archiverPath, archiverArgs...)
	cmd.Env = t.toolChain.Env

	out, err := cmd.CombinedOutput()

	if err != nil {
		return corepkg.LogErrorf(err, "Archiving failed: ", string(out))
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------

type WinMsdevLinker struct {
	toolChain    *WinMsdev
	buildConfig  denv.BuildConfig
	buildTarget  denv.BuildTarget
	linkerPath   string
	linkerArgs   *corepkg.Arguments
	vars         *corepkg.Vars
	frameworks   []string
	libraryFiles []string
}

func (l *WinMsdev) NewLinker(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Linker {
	args := corepkg.NewArguments(512)
	return &WinMsdevLinker{
		toolChain:    l,
		buildConfig:  buildConfig,
		buildTarget:  buildTarget,
		linkerPath:   "",
		linkerArgs:   args,
		vars:         corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
		frameworks:   []string{},
		libraryFiles: []string{},
	}
}

func (l *WinMsdevLinker) LinkedFilepath(filepath string) string {
	return filepath + ".exe"
}

func (l *WinMsdevLinker) SetupArgs(libraryPaths []string, libraryFiles []string) {
	// for i, libPath := range libraryPaths {
	// 	libraryPaths[i] = "/LIBPATH:" + libPath
	// }
	// for i, libFile := range libraryFiles {
	// 	libraryFiles[i] = libFile
	// }
	//l.vars.Prepend("library.paths", libraryPaths...)

	//l.libraryFiles = append(l.libraryFiles, libraryFiles...)

	if linker_args, ok := l.toolChain.Vars.Get(`recipe.link.pattern`); ok {
		l.linkerPath = linker_args[0]
		l.linkerArgs = corepkg.NewArguments(0)
		l.linkerArgs.Args = linker_args[1:]

		l.linkerPath = l.toolChain.Vars.FinalResolveString(l.linkerPath, " ", l.vars)
		l.linkerArgs.Args = l.toolChain.Vars.FinalResolveArray(l.linkerArgs.Args, l.vars)

		l.linkerArgs.Args = slices.DeleteFunc(l.linkerArgs.Args, func(s string) bool { return strings.TrimSpace(s) == "" })
	}
}

func (l *WinMsdevLinker) Link(inputObjectsAbsFilepaths, inputArchivesAbsFilepaths []string, outputAppRelFilepath string) error {

	linkerPath := l.linkerPath
	linkerArgs := l.linkerArgs.Args

	// TODO would like this to be part of the resolve step
	linkerArgs = append(linkerArgs, "/MAP:"+l.toolChain.ChangeFileExtension(outputAppRelFilepath, ".map"))
	linkerArgs = append(linkerArgs, "/OUT:"+outputAppRelFilepath)
	linkerArgs = append(linkerArgs, inputObjectsAbsFilepaths...)
	linkerArgs = append(linkerArgs, inputArchivesAbsFilepaths...)
	// for _, libFile := range l.libraryFiles {
	// 	linkerArgs = append(linkerArgs, libFile)
	// }

	corepkg.LogInff("Linking (%s) %s", l.buildConfig.String(), filepath.Base(outputAppRelFilepath))

	// corepkg.LogInfof("Linker command: %s %s", linkerPath, strings.Join(linkerArgs, " "))

	cmd := exec.Command(linkerPath, linkerArgs...)
	cmd.Env = l.toolChain.Env

	out, err := cmd.CombinedOutput()

	if err != nil {
		corepkg.LogInff("Link failed, output:\n%s", string(out))
		return corepkg.LogError(err, "Linking failed")
	}
	if len(out) > 0 {
		corepkg.LogInfof("Link output:\n%s", string(out))
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------

func (ms *WinMsdev) NewBurner(config denv.BuildConfig, target denv.BuildTarget) Burner {
	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------

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

	// Get the current environment
	osEnv := os.Environ()

	// Make a copy and adjust it for the toolchain processes
	env := slices.Clone(osEnv)

	// PATH=
	//      C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Tools\MSVC\14.44.35207\bin\HostX64\x64;
	//      C:\Program Files (x86)\Windows Kits\10\bin\10.0.26100.0\\x64;
	//      C:\Program Files (x86)\Windows Kits\10\bin\\x64;

	path, _ := vars.Get("PATH")

	// LIBPATH=
	//     C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Tools\MSVC\14.41.34120\lib\x64;
	//     C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Tools\MSVC\14.41.34120\atlmfc\lib\x64;
	//     C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Auxiliary\VS\lib\x64;
	//     C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Auxiliary\VS\UnitTest\lib;
	//     C:\Program Files (x86)\Windows Kits\10\lib\10.0.22621.0\ucrt\x64;
	//     C:\Program Files (x86)\Windows Kits\10\lib\10.0.22621.0\um\x64;
	//     C:\Program Files (x86)\Windows Kits\NETFXSDK\4.8\lib\um\x64;

	libPath, _ := vars.Get("LIBPATH")
	lib := libPath

	// INCLUDE=
	//         C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Tools\MSVC\14.41.34120\include;
	//         C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Tools\MSVC\14.41.34120\atlmfc\include;
	//         C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Auxiliary\VS\include;
	//         C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Auxiliary\VS\UnitTest\include;
	//         C:\Program Files (x86)\Windows Kits\10\Include\10.0.22621.0\ucrt;
	//         C:\Program Files (x86)\Windows Kits\10\Include\10.0.22621.0\um;
	//         C:\Program Files (x86)\Windows Kits\10\Include\10.0.22621.0\shared;
	//         C:\Program Files (x86)\Windows Kits\10\Include\10.0.22621.0\winrt;
	//         C:\Program Files (x86)\Windows Kits\10\Include\10.0.22621.0\cppwinrt;
	//         C:\Program Files (x86)\Windows Kits\NETFXSDK\4.8\Include\um;

	include, _ := vars.Get("INCLUDE")

	// Update the environment variables
	for i, e := range env {
		if e, found := strings.CutPrefix(e, "Path="); found {
			if path != nil {
				env[i] = "Path=" + strings.Join(path, ";") + ";" + e
				path = nil
			}
		} else if e, found := strings.CutPrefix(e, "LIB="); found {
			if lib != nil {
				env[i] = "LIB=" + strings.Join(lib, ";") + ";" + e
				lib = nil
			}
		} else if e, found := strings.CutPrefix(e, "LIBPATH="); found {
			if libPath != nil {
				env[i] = "LIBPATH=" + strings.Join(libPath, ";") + ";" + e
				libPath = nil
			}
		} else if e, found := strings.CutPrefix(e, "INCLUDE="); found {
			if include != nil {
				env[i] = "INCLUDE=" + strings.Join(include, ";") + ";" + e
				include = nil
			}
		}
	}

	// Add any missing environment variables
	if path != nil {
		env = append(env, "Path="+strings.Join(path, ";"))
	}
	if lib != nil {
		env = append(env, "LIB="+strings.Join(lib, ";"))
	}
	if libPath != nil {
		env = append(env, "LIBPATH="+strings.Join(libPath, ";"))
	}
	if include != nil {
		env = append(env, "INCLUDE="+strings.Join(include, ";"))
	}

	// Debug, print all environment variables
	// for _, e := range env {
	// 	corepkg.LogInfof("Env: %s", e)
	// }

	return &WinMsdev{
		Name: "msdev",
		Vars: vars,
		Env:  env,
	}, nil
}
