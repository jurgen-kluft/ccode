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

type DarwinClangv2 struct {
	Name string
	Vars *corepkg.Vars
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// C/C++ Compiler

type ToolchainDarwinClangCompilerv2 struct {
	toolChain       *DarwinClangv2
	buildConfig     denv.BuildConfig
	buildTarget     denv.BuildTarget
	cCompilerPath   string
	cCompilerArgs   *corepkg.Arguments
	cppCompilerPath string
	cppCompilerArgs *corepkg.Arguments
	vars            *corepkg.Vars
}

func (t *DarwinClangv2) NewCompiler(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Compiler {
	return &ToolchainDarwinClangCompilerv2{
		toolChain:       t,
		buildConfig:     buildConfig,
		buildTarget:     buildTarget,
		cCompilerPath:   "",
		cCompilerArgs:   nil,
		cppCompilerPath: "",
		cppCompilerArgs: nil,
		vars:            corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
	}
}

func (cl *ToolchainDarwinClangCompilerv2) ObjFilepath(srcRelFilepath string) string {
	return srcRelFilepath + ".o"
}

func (cl *ToolchainDarwinClangCompilerv2) DepFilepath(objRelFilepath string) string {
	return objRelFilepath + ".d"
}

func (cl *ToolchainDarwinClangCompilerv2) SetupArgs(_defines []string, _includes []string) {
	for i, inc := range _includes {
		if !strings.HasPrefix(inc, "-I") {
			_includes[i] = "-I" + inc
		}
	}
	for i, def := range _defines {
		if !strings.HasPrefix(def, "-D") {
			_defines[i] = "-D" + def
		}
	}
	cl.vars.Set("build.includes", _includes...)
	cl.vars.Set("build.defines", _defines...)

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

func (cl *ToolchainDarwinClangCompilerv2) Compile(sourceAbsFilepaths []string, objRelFilepaths []string) error {
	for i, sourceAbsFilepath := range sourceAbsFilepaths {
		var compilerPath string
		var compilerArgs []string
		if strings.HasSuffix(sourceAbsFilepath, ".c") {
			compilerPath = cl.cCompilerPath
			compilerArgs = cl.cCompilerArgs.Args
		} else {
			compilerPath = cl.cppCompilerPath
			compilerArgs = cl.cppCompilerArgs.Args
		}

		// TODO would like this to be part of the resolve step
		compilerArgs = append(compilerArgs, "-o", objRelFilepaths[i])
		compilerArgs = append(compilerArgs, sourceAbsFilepath)

		cmd := exec.Command(compilerPath, compilerArgs...)

		//corepkg.LogInfof("Compiling (%s) %s", cl.buildConfig.String(), filepath.Base(sourceAbsFilepath))
		corepkg.LogInfof("Compiling (%s) %s", cl.buildConfig.String(), sourceAbsFilepath)
		//corepkg.LogInfof("Command: %s %s", compilerPath, strings.Join(compilerArgs, " "))
		out, err := cmd.CombinedOutput()
		if err != nil {
			corepkg.LogInfof("Compile failed, output:\n%s", string(out))
			return corepkg.LogErrorf(err, "Compiling failed")
		}
		if len(out) > 0 {
			corepkg.LogInfof("Compile output:\n%s", string(out))
		}
	}
	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Archiver

type ToolchainDarwinClangStaticArchiverv2 struct {
	toolChain   *DarwinClangv2
	buildConfig denv.BuildConfig
	buildTarget denv.BuildTarget
	arPath      string
	arArgs      *corepkg.Arguments
}

type ToolchainDarwinClangDynamicArchiverv2 struct {
	toolChain   *DarwinClangv2
	buildConfig denv.BuildConfig
	buildTarget denv.BuildTarget
	arPath      string
	arArgs      *corepkg.Arguments
}

func (t *DarwinClangv2) NewArchiver(at ArchiverType, buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Archiver {
	args := corepkg.NewArguments(512)
	switch at {
	case ArchiverTypeStatic:
		return &ToolchainDarwinClangStaticArchiverv2{toolChain: t, buildConfig: buildConfig, buildTarget: buildTarget, arArgs: args}
	case ArchiverTypeDynamic:
		return &ToolchainDarwinClangDynamicArchiverv2{toolChain: t, buildConfig: buildConfig, buildTarget: buildTarget, arArgs: args}
	}
	return nil
}

func (t *ToolchainDarwinClangStaticArchiverv2) LibFilepath(_filepath string) string {
	filename := corepkg.PathFilename(_filepath, true)
	dirpath := corepkg.PathDirname(_filepath)
	return filepath.Join(dirpath, "lib"+filename+".a") // The file extension for the archive on Darwin is typically ".a"
}

func (t *ToolchainDarwinClangStaticArchiverv2) SetupArgs() {
	if archiver_args, ok := t.toolChain.Vars.Get(`recipe.ar.pattern`); ok {
		t.arPath = archiver_args[0]
		t.arArgs = corepkg.NewArguments(0)
		t.arArgs.Args = archiver_args[1:]
		t.arPath = t.toolChain.Vars.FinalResolveString(t.arPath, " ")
		t.arArgs.Args = t.toolChain.Vars.FinalResolveArray(t.arArgs.Args)

		t.arArgs.Args = slices.DeleteFunc(t.arArgs.Args, func(s string) bool { return strings.TrimSpace(s) == "" })
	}
}

func (t *ToolchainDarwinClangStaticArchiverv2) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {
	archiverPath := t.arPath
	archiverArgs := t.arArgs.Args

	// TODO would like this to be part of the resolve step
	archiverArgs = append(archiverArgs, outputArchiveFilepath)
	archiverArgs = append(archiverArgs, inputObjectFilepaths...)

	cmd := exec.Command(archiverPath, archiverArgs...)

	out, err := cmd.CombinedOutput()

	if err != nil {
		return corepkg.LogErrorf(err, "Archiving failed: ", string(out))
	}

	return nil
}

func (t *ToolchainDarwinClangDynamicArchiverv2) LibFilepath(_filepath string) string {
	filename := corepkg.PathFilename(_filepath, true)
	dirpath := corepkg.PathDirname(_filepath)
	return filepath.Join(dirpath, "lib"+filename+".dylib")
}

func (t *ToolchainDarwinClangDynamicArchiverv2) SetupArgs() {
	if archiver_args, ok := t.toolChain.Vars.Get(`recipe.ar.pattern`); ok {
		t.arPath = archiver_args[0]
		t.arArgs = corepkg.NewArguments(0)
		t.arArgs.Args = archiver_args[1:]

		t.arPath = t.toolChain.Vars.FinalResolveString(t.arPath, " ")
		t.arArgs.Args = t.toolChain.Vars.FinalResolveArray(t.arArgs.Args)
	}
}

func (t *ToolchainDarwinClangDynamicArchiverv2) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {
	archiverPath := t.arPath
	archiverArgs := t.arArgs.Args

	// TODO would like this to be part of the resolve step
	archiverArgs = append(archiverArgs, "-dynamiclib", "-o", outputArchiveFilepath)
	archiverArgs = append(archiverArgs, inputObjectFilepaths...)

	cmd := exec.Command(archiverPath, archiverArgs...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return corepkg.LogErrorf(err, "Archiving failed")
	}
	if len(out) > 0 {
		corepkg.LogInfof("Archive output:\n%s", string(out))
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Linker

type ToolchainDarwinClangLinkerv2 struct {
	toolChain   *DarwinClangv2
	buildConfig denv.BuildConfig
	buildTarget denv.BuildTarget
	linkerPath  string
	linkerArgs  *corepkg.Arguments
	vars        *corepkg.Vars
	frameworks  []string
}

func (l *DarwinClangv2) NewLinker(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Linker {
	args := corepkg.NewArguments(512)
	return &ToolchainDarwinClangLinkerv2{
		toolChain:   l,
		buildConfig: buildConfig,
		buildTarget: buildTarget,
		vars:        corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
		linkerArgs:  args,
	}
}

func (l *ToolchainDarwinClangLinkerv2) LinkedFilepath(filepath string) string {
	return filepath
}

func (l *ToolchainDarwinClangLinkerv2) SetupArgs(libraryPaths []string, libraryFiles []string) {
	for i, libPath := range libraryPaths {
		libraryPaths[i] = "-L" + libPath
	}
	for i, libFile := range libraryFiles {
		libraryFiles[i] = "-l" + libFile
	}

	l.vars.Set("library.paths", libraryPaths...)
	l.vars.Set("library.files", libraryFiles...)

	if linker_args, ok := l.toolChain.Vars.Get(`recipe.link.pattern`); ok {
		l.linkerPath = linker_args[0]
		l.linkerArgs = corepkg.NewArguments(0)
		l.linkerArgs.Args = linker_args[1:]

		l.linkerPath = l.toolChain.Vars.FinalResolveString(l.linkerPath, " ", l.vars)
		l.linkerArgs.Args = l.toolChain.Vars.FinalResolveArray(l.linkerArgs.Args, l.vars)

		l.linkerArgs.Args = slices.DeleteFunc(l.linkerArgs.Args, func(s string) bool { return strings.TrimSpace(s) == "" })
	}
}

func (l *ToolchainDarwinClangLinkerv2) Link(inputObjectsAbsFilepaths, inputArchivesAbsFilepaths []string, outputAppRelFilepathNoExt string) error {

	linkerPath := l.linkerPath
	linkerArgs := l.linkerArgs.Args

	// TODO would like this to be part of the resolve step
	linkerArgs = append(linkerArgs, "-Wl,-map,"+outputAppRelFilepathNoExt+".map")
	linkerArgs = append(linkerArgs, "-o", l.LinkedFilepath(outputAppRelFilepathNoExt))
	linkerArgs = append(linkerArgs, inputObjectsAbsFilepaths...)
	linkerArgs = append(linkerArgs, inputArchivesAbsFilepaths...)

	corepkg.LogInff("Linking (%s) %s", l.buildConfig.String(), filepath.Base(outputAppRelFilepathNoExt))

	cmd := exec.Command(linkerPath, linkerArgs...)
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
// Burner

func (t *DarwinClangv2) NewBurner(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Burner {
	return &EmptyBurner{}
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Dependency Tracker
func (t *DarwinClangv2) NewDependencyTracker(dirpath string) deptrackr.FileTrackr {
	return deptrackr.LoadDepFileTrackr(filepath.Join(dirpath, "deptrackr"))
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Toolchain for Clang on MacOS
func NewDarwinClangv2(vars *corepkg.Vars, projectName string, buildPath string, arch string) *DarwinClangv2 {

	vars.Set("project.name", projectName)
	vars.Set("build.path", buildPath)
	vars.Set("build.arch", arch)

	return &DarwinClangv2{Name: "clang", Vars: vars}
}
