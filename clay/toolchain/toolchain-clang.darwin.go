package toolchain

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain/clang"
	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
	corepkg "github.com/jurgen-kluft/ccode/core"
)

type DarwinClang struct {
	Name            string
	cCompilerPath   string   // Path to the C compiler
	cxxCompilerPath string   // Path to the C++ compiler
	arPath          string   // Path to the archiver (ar)
	linkerPath      string   // Path to the linker (clang)
	frameworks      []string // List of frameworks to link against
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// C/C++ Compiler

type ToolchainDarwinClangCompiler struct {
	toolChain *DarwinClang
	config    *Config
	args      *corepkg.Arguments
	cmdline   *clang.CompilerCmdLine
}

func (t *DarwinClang) NewCompiler(config *Config) Compiler {
	args := corepkg.NewArguments(512)
	return &ToolchainDarwinClangCompiler{
		toolChain: t,
		config:    config,
		args:      args,
		cmdline:   clang.NewCompilerCmdline(args),
	}
}

func (cl *ToolchainDarwinClangCompiler) ObjFilepath(srcRelFilepath string) string {
	return srcRelFilepath + ".o"
}

func (cl *ToolchainDarwinClangCompiler) DepFilepath(objRelFilepath string) string {
	return objRelFilepath + ".d"
}

func (cl *ToolchainDarwinClangCompiler) SetupArgs(_defines []string, _includes []string) {
	cl.cmdline.CompileOnly()
	cl.cmdline.NoLogo()
	cl.cmdline.WarningsAreErrors()

	if cl.config.IsDebug() {
		// Debug-specific arguments
		cl.cmdline.DisableOptimizations()
		cl.cmdline.GenerateDebugInfo()
		cl.cmdline.DisableFramePointer()
	} else if cl.config.IsRelease() {
		// Release and Final specific arguments
		cl.cmdline.OptimizeForSpeed()
		cl.cmdline.EnableInlineExpansion()
		cl.cmdline.EnableIntrinsicFunctions()
	} else if cl.config.IsFinal() {
		// Final-specific arguments
		cl.cmdline.OptimizeHard()
		cl.cmdline.EnableInlineExpansion()
		cl.cmdline.EnableIntrinsicFunctions()
		cl.cmdline.OmitFramePointer()
	}

	// Test-specific arguments
	// if cl.config.IsTest() {
	// 	cl.cmdline.EnableExceptionHandling()
	// }

	cl.cmdline.UseFloatingPointPrecise()

	cl.cmdline.Includes(_includes)
	cl.cmdline.Defines(_defines)

	cl.cmdline.GenerateDependencyFiles()

	cl.cmdline.Save()
}

func (cl *ToolchainDarwinClangCompiler) Compile(sourceAbsFilepaths []string, objRelFilepaths []string) error {

	for i, sourceAbsFilepath := range sourceAbsFilepaths {
		cl.cmdline.Restore()

		var compilerPath string
		if strings.HasSuffix(sourceAbsFilepath, ".c") {
			cl.cmdline.UseC11()
			compilerPath = cl.toolChain.cCompilerPath
		} else {
			cl.cmdline.UseCpp17()
			compilerPath = cl.toolChain.cxxCompilerPath
		}

		// The source file and the output object file
		cl.cmdline.ObjectFile(objRelFilepaths[i])
		cl.cmdline.SourceFile(sourceAbsFilepath)

		compilerArgs := cl.args.Args
		cmd := exec.Command(compilerPath, compilerArgs...)

		corepkg.LogInfof("Compiling (%s) %s", cl.config.Config.AsString(), filepath.Base(sourceAbsFilepath))
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

type ToolchainDarwinClangStaticArchiver struct {
	toolChain *DarwinClang
	config    *Config
	args      *corepkg.Arguments
	cmdline   *clang.ArchiverCmdline
}

type ToolchainDarwinClangDynamicArchiver struct {
	toolChain *DarwinClang
	config    *Config
	args      *corepkg.Arguments
	cmdline   *clang.ArchiverCmdline
}

func (t *DarwinClang) NewArchiver(at ArchiverType, config *Config) Archiver {
	args := corepkg.NewArguments(512)
	cmdline := clang.NewArchiverCmdline(args)
	switch at {
	case ArchiverTypeStatic:
		return &ToolchainDarwinClangStaticArchiver{toolChain: t, config: config, args: args, cmdline: cmdline}
	case ArchiverTypeDynamic:
		return &ToolchainDarwinClangDynamicArchiver{toolChain: t, config: config, args: args, cmdline: cmdline}
	}
	return nil
}

func (t *ToolchainDarwinClangStaticArchiver) LibFilepath(_filepath string) string {
	filename := corepkg.PathFilename(_filepath, true)
	dirpath := corepkg.PathDirname(_filepath)
	return filepath.Join(dirpath, "lib"+filename+".a") // The file extension for the archive on Darwin is typically ".a"
}

func (t *ToolchainDarwinClangStaticArchiver) SetupArgs() {
	t.cmdline.ReplaceCreateSort()
	t.cmdline.Save()
}

func (t *ToolchainDarwinClangStaticArchiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {
	t.cmdline.Restore()
	t.cmdline.OutputArchiveAndObjectFiles(outputArchiveFilepath, inputObjectFilepaths)

	archiverPath := t.toolChain.arPath
	archiverArgs := t.args.Args

	cmd := exec.Command(archiverPath, archiverArgs...)

	out, err := cmd.CombinedOutput()

	if err != nil {
		return corepkg.LogErrorf(err, "Archiving failed: ", string(out))
	}

	return nil
}

func (t *ToolchainDarwinClangDynamicArchiver) LibFilepath(_filepath string) string {
	filename := corepkg.PathFilename(_filepath, true)
	dirpath := corepkg.PathDirname(_filepath)
	return filepath.Join(dirpath, "lib"+filename+".dylib")
}

func (t *ToolchainDarwinClangDynamicArchiver) SetupArgs() {
	t.cmdline.DynamicLib()
	t.cmdline.Save()
}

func (t *ToolchainDarwinClangDynamicArchiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {
	t.cmdline.Restore()
	t.cmdline.InstallName(outputArchiveFilepath)
	t.cmdline.OutputArchiveAndObjectFiles(outputArchiveFilepath, inputObjectFilepaths)

	archiverPath := t.toolChain.arPath
	archiverArgs := t.args.Args

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

type ToolchainDarwinClangLinker struct {
	toolChain *DarwinClang
	config    *Config
	args      *corepkg.Arguments
	cmdline   *clang.LinkerCmdline
}

func (l *DarwinClang) NewLinker(config *Config) Linker {
	args := corepkg.NewArguments(512)
	return &ToolchainDarwinClangLinker{
		toolChain: l,
		config:    config,
		args:      args,
		cmdline:   clang.NewLinkerCmdline(args),
	}
}

func (l *ToolchainDarwinClangLinker) LinkedFilepath(filepath string) string {
	return filepath
}

func (l *ToolchainDarwinClangLinker) SetupArgs(libraryPaths []string, libraryFiles []string) {
	l.cmdline.ErrorReportPrompt()
	l.cmdline.NoLogo()

	if l.config.IsDebug() {
		l.cmdline.GenerateDebugInfo()
		l.cmdline.UseMultithreadedDebug()
	}
	if l.config.IsRelease() || l.config.IsFinal() {
		l.cmdline.UseMultithreaded()
		l.cmdline.OptimizeReferences()
		l.cmdline.OptimizeIdenticalFolding()
	}
	if l.config.IsFinal() {
		l.cmdline.LinkTimeCodeGeneration()
		l.cmdline.DisableIncrementalLinking()
		l.cmdline.UseMultithreadedFinal()
	}

	l.cmdline.SubsystemConsole()

	l.cmdline.DynamicBase()
	l.cmdline.EnableDataExecutionPrevention()
	l.cmdline.MachineX64()

	l.cmdline.Frameworks(l.toolChain.frameworks)

	l.cmdline.Save()
}

func (l *ToolchainDarwinClangLinker) Link(inputArchiveAbsFilepaths []string, outputAppRelFilepathNoExt string) error {
	l.cmdline.Restore()

	l.cmdline.Out(outputAppRelFilepathNoExt)
	l.cmdline.Libs([]string{"stdc++"})
	l.cmdline.ObjectFiles(inputArchiveAbsFilepaths)

	linkerPath := l.toolChain.linkerPath
	linkerArgs := l.args.Args
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

func (t *DarwinClang) NewBurner(config *Config) Burner {
	return &EmptyBurner{}
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Dependency Tracker
func (t *DarwinClang) NewDependencyTracker(dirpath string) deptrackr.FileTrackr {
	return deptrackr.LoadDepFileTrackr(filepath.Join(dirpath, "deptrackr"))
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Toolchain for Clang on MacOS
const (
	archtype_arm    string = "arm"    // ARM: arm, armv.*, xscale
	archtype_arm64  string = "arm64"  // ARM: arm, armv.*, xscale
	archtype_x86    string = "x86"    // X86: i[3-9]86
	archtype_x86_64 string = "x86_64" // X86-64: amd64, x86_64
)

func NewDarwinClang(arch string, frameworks []string) (t *DarwinClang, err error) {
	var cCompilerPath string
	if cCompilerPath, err = exec.LookPath("clang"); err != nil {
		return nil, err
	}

	var cxxCompilerPath string
	if cxxCompilerPath, err = exec.LookPath("clang++"); err != nil {
		return nil, err
	}

	var arPath string
	if arPath, err = exec.LookPath("ar"); err != nil {
		return nil, err
	}

	linkerPath := cCompilerPath

	t = &DarwinClang{
		Name:            "clang",
		cCompilerPath:   cCompilerPath,
		cxxCompilerPath: cxxCompilerPath,
		arPath:          arPath,
		linkerPath:      linkerPath,
		frameworks:      []string{},
	}
	t.frameworks = append(t.frameworks, frameworks...)

	return t, nil
}
