package toolchain

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
	"github.com/jurgen-kluft/ccode/clay/toolchain/msvc"
	"github.com/jurgen-kluft/ccode/foundation"
)

type WinMsdev struct {
	Name string
	Msvc *msvc.MsvcEnvironment
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Compiler

// Compiler options
//   - https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/compiler-options-listed-by-category.md
//

type WinMsDevCompiler struct {
	toolChain *WinMsdev             // The toolchain this compiler belongs to
	config    *Config               // Build configuration
	args      *foundation.Arguments // Arguments for the compiler
	cmdline   *msvc.CompilerCmdLine // Cmdline for the compiler
}

func (m *WinMsdev) NewCompiler(config *Config) Compiler {
	return &WinMsDevCompiler{
		toolChain: m,
		config:    config,
		args:      foundation.NewArguments(512),
		cmdline:   msvc.NewCompilerCmdLine(foundation.NewArguments(512)),
	}
}

func (cl *WinMsDevCompiler) ObjFilepath(srcRelFilepath string) string {
	srcRelFilepath = strings.TrimSuffix(srcRelFilepath, ".c")
	srcRelFilepath = strings.TrimSuffix(srcRelFilepath, ".cpp")
	return srcRelFilepath + ".obj"
}

func (cl *WinMsDevCompiler) DepFilepath(objRelFilepath string) string {
	objRelFilepath = strings.TrimSuffix(objRelFilepath, ".c")
	objRelFilepath = strings.TrimSuffix(objRelFilepath, ".cpp")
	return objRelFilepath + ".d"
}

func (cl *WinMsDevCompiler) SetupArgs(_defines []string, _includes []string) {
	// Common arguments
	cl.cmdline.CompileOnly()
	cl.cmdline.NoLogo()
	cl.cmdline.DiagnosticsColumnMode()
	cl.cmdline.DiagnosticsEmitFullPathOfSourceFiles()
	cl.cmdline.WarningLevel3()
	cl.cmdline.WarningsAreErrors()

	cl.cmdline.EnableStringPooling()

	if cl.config.IsDebug() {
		// Debug-specific arguments
		cl.cmdline.DisableOptimizations()
		cl.cmdline.GenerateDebugInfo()
		cl.cmdline.DisableFramePointer()
		cl.cmdline.UseMultithreadedDebugRuntime()
	} else if cl.config.IsRelease() {
		// Release-specific arguments
		cl.cmdline.OptimizeForSize()
		cl.cmdline.OptimizeForSpeed()
		cl.cmdline.EnableInlineExpansion(1)
		cl.cmdline.EnableIntrinsicFunctions()
		cl.cmdline.OmitFramePointer()
		cl.cmdline.UseMultithreadedRuntime()
	} else if cl.config.IsFinal() {
		// Final-specific arguments
		cl.cmdline.OptimizeForSize()
		cl.cmdline.OptimizeForSpeed()
		cl.cmdline.EnableInlineExpansion(3)
		cl.cmdline.EnableIntrinsicFunctions()
		cl.cmdline.OmitFramePointer()
		cl.cmdline.UseMultithreadedRuntime()
		cl.cmdline.EnableWholeProgramOptimization()
	}

	// Test-specific arguments
	if cl.config.IsTest() {
		cl.cmdline.EnableExceptionHandling()
	}

	cl.cmdline.Save()
}

func (cl *WinMsDevCompiler) Compile(sourceAbsFilepaths []string, objRelFilepaths []string) error {
	// Analyze all the object filepaths and organize them per directory, we do this because
	// the MSVC compiler outputs object files into a single directory, we do not want this.
	// And we do not want to call the compiler for each source file, but rather for each directory.
	sourceFilesPerDir := make(map[string][]string)
	for i, objFilepath := range objRelFilepaths {
		objFilepath = foundation.PathWindowsPath(objFilepath)
		objFile := foundation.PathFilename(objFilepath, true)
		srcDir := objFile[:len(objFilepath)-len(objFile)]
		srcFile := foundation.PathFilename(sourceAbsFilepaths[i], true)
		if _, ok := sourceFilesPerDir[srcDir]; !ok {
			sourceFiles := make([]string, 0, len(sourceAbsFilepaths)-i)
			sourceFiles = append(sourceFiles, srcFile)
			sourceFilesPerDir[srcDir] = sourceFiles
		} else {
			sourceFilesPerDir[srcDir] = append(sourceFilesPerDir[srcDir], srcFile)
		}
	}

	// Iterate over the source files per directory and compile them.
	for srcDir, srcFiles := range sourceFilesPerDir {
		cl.cmdline.Restore() // Restore the command line arguments

		srcDir = foundation.PathWindowsPath(srcDir)
		cl.cmdline.OutDir(srcDir)
		cl.cmdline.GenerateDependencyFiles(srcDir)
		cl.cmdline.SourceFiles(srcFiles)

		configStr := cl.config.Config.AsString()
		for _, srcFile := range srcFiles {
			foundation.LogInfof("Compiling (%s) %s\n", configStr, srcFile)
		}

		// Prepare the command to execute the compiler.
		compilerPath := filepath.Join(cl.toolChain.Msvc.CompilerPath, cl.toolChain.Msvc.CompilerBin)
		compilerArgs := cl.args.Args
		cmd := exec.Command(compilerPath, compilerArgs...)
		cmd.Env = append(cmd.Env, "PATH="+foundation.PathWindowsPath(cl.toolChain.Msvc.CompilerPath))
		out, err := cmd.CombinedOutput()
		if err != nil {
			foundation.LogInfof("Compile failed, output:\n%s\n", string(out))
			return foundation.LogErrorf(err, "Compiling failed")
		}
		if len(out) > 0 {
			foundation.LogInfof("Compile output:\n%s\n", string(out))
		}
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Archiver

func (m *WinMsdev) NewArchiver(a ArchiverType, config *Config) Archiver {
	return &WinMsDevArchiver{
		toolChain: m,
		config:    config,
		args:      foundation.NewArguments(512),
		cmdline:   msvc.NewArchiverCmdline(foundation.NewArguments(512)),
	}
}

type WinMsDevArchiver struct {
	toolChain *WinMsdev             // The toolchain this archiver belongs to
	config    *Config               // Build configuration
	args      *foundation.Arguments // Arguments for the archiver
	cmdline   *msvc.ArchiverCmdline // Cmdline for the archiver
}

func (a *WinMsDevArchiver) LibFilepath(_filepath string) string {
	filename := foundation.PathFilename(_filepath, true)
	dirpath := foundation.PathDirname(_filepath)
	return filepath.Join(dirpath, filename+".lib")
}

func (a *WinMsDevArchiver) SetupArgs() {
	a.cmdline.NoLogo()
	a.cmdline.MachineX64()
	a.cmdline.Save()
}

func (a *WinMsDevArchiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {
	a.cmdline.Restore()
	a.cmdline.Out(outputArchiveFilepath)
	a.cmdline.ObjectFiles(inputObjectFilepaths)

	archiverArgs := a.args.Args
	archiverPath := filepath.Join(a.toolChain.Msvc.ArchiverPath, a.toolChain.Msvc.ArchiverBin)
	cmd := exec.Command(archiverPath, archiverArgs...)

	foundation.LogInfof("Archiving (%s) %s\n", a.config.Config.AsString(), outputArchiveFilepath)
	cmd.Env = append(cmd.Env, "PATH="+foundation.PathWindowsPath(a.toolChain.Msvc.ArchiverPath))
	out, err := cmd.CombinedOutput()
	if err != nil {
		foundation.LogInfof("Archive failed, output:\n%s\n", string(out))
		return foundation.LogErrorf(err, "Archiving failed")
	}
	if len(out) > 0 {
		foundation.LogInfof("Archive output:\n%s\n", string(out))
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Linker

func (ms *WinMsdev) NewLinker(config *Config) Linker {
	return &WinMsDevLinker{
		toolChain: ms,
		config:    config,
		args:      foundation.NewArguments(512),
		cmdline:   msvc.NewLinkerCmdline(foundation.NewArguments(512)),
	}
}

type WinMsDevLinker struct {
	toolChain *WinMsdev             // The toolchain this archiver belongs to
	config    *Config               // Build configuration
	args      *foundation.Arguments // Arguments for the linker
	cmdline   *msvc.LinkerCmdline   // Cmdline for the linker
}

func (l *WinMsDevLinker) LinkedFilepath(filepath string) string {
	return filepath + ".exe"
}

func (l *WinMsDevLinker) SetupArgs(libraryPaths []string, libraryFiles []string) {
	l.cmdline.ErrorReportPrompt()
	l.cmdline.NoLogo()
	if l.config.IsDebug() {
		l.cmdline.GenerateDebugInfo()
		l.cmdline.GenerateMultithreadedDebugExe()
	}
	if l.config.IsRelease() || l.config.IsFinal() {
		l.cmdline.GenerateMultithreadedExe()
		l.cmdline.OptimizeReferences()
		l.cmdline.OptimizeIdenticalFolding()
	}
	if l.config.IsFinal() {
		l.cmdline.LinkTimeCodeGeneration()
		l.cmdline.DisableIncrementalLinking()
		l.cmdline.UseMultithreadedFinal()
	}

	// Console or GUI application?
	l.cmdline.SubsystemConsole()

	l.cmdline.DynamicBase()
	l.cmdline.EnableDataExecutionPrevention()
	l.cmdline.MachineX64()
}

func (l *WinMsDevLinker) Link(inputArchiveAbsFilepaths []string, outputAppRelFilepath string) error {

	outputAppRelFilepath = foundation.PathWindowsPath(outputAppRelFilepath)

	l.cmdline.GenerateMapfile(foundation.FileChangeExtension(outputAppRelFilepath, ".map"))
	l.cmdline.Out(outputAppRelFilepath)

	libraryPathsMap := map[string]bool{}
	for _, archiveFilepath := range inputArchiveAbsFilepaths {
		archiveFilepath = foundation.PathWindowsPath(archiveFilepath)
		libraryPathsMap[foundation.PathDirname(archiveFilepath)] = true
	}
	libraryPaths := make([]string, 0, len(libraryPathsMap))
	for libPath := range libraryPathsMap {
		libraryPaths = append(libraryPaths, foundation.PathWindowsPath(libPath))
	}
	l.cmdline.LibPaths(libraryPaths)

	// TODO Where do we get this list of libraries from?
	// Note: Could we perhaps scan all the header files that are included for any patterns that
	//       give us hints about the libraries that are needed?
	systemLibraries := []string{
		"kernel32.lib",
		"user32.lib",
		"gdi32.lib",
		"winspool.lib",
		"comdlg32.lib",
		"advapi32.lib",
		"shell32.lib",
		"ole32.lib",
		"oleaut32.lib",
		"uuid.lib",
		"odbc32.lib",
		"odbccp32.lib",
	}
	l.cmdline.Libs(systemLibraries)

	libraryPaths = libraryPaths[0:]
	for _, archiveFile := range inputArchiveAbsFilepaths {
		archiveFile = foundation.PathFilename(archiveFile, true)
		libraryPaths = append(libraryPaths, archiveFile)
	}
	l.cmdline.Libs(libraryPaths)

	linkerPath := filepath.Join(l.toolChain.Msvc.LinkerPath, l.toolChain.Msvc.LinkerBin)
	linkerArgs := l.args.Args

	cmd := exec.Command(linkerPath, linkerArgs...)
	cmd.Env = append(cmd.Env, "PATH="+foundation.PathWindowsPath(l.toolChain.Msvc.LinkerPath))

	foundation.LogInfof("Linking (%s) %s\n", l.config.Config.AsString(), outputAppRelFilepath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		foundation.LogInfof("Linking failed, output:\n%s\n", string(out))
	}
	if len(out) > 0 {
		foundation.LogInfof("Linking output:\n%s\n", string(out))
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Burner

func (t *WinMsdev) NewBurner(config *Config) Burner {
	return &EmptyBurner{}
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Dependency Tracker

func (t *WinMsdev) NewDependencyTracker(dirpath string) deptrackr.FileTrackr {
	// Note: This should be the dependency tracker that can read .json dependency files that are
	// generated by the MSVC compiler.
	return deptrackr.LoadJsonFileTrackr(filepath.Join(dirpath, "deptrackr"))
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Toolchain for Visual Studio on Windows

func NewWinMsdev(arch string, product string) (t *WinMsdev, err error) {
	msdevSetup, err := msvc.InitMsvcVisualStudio(msvc.VsVersion2022, "", msvc.WinArchx64, msvc.WinArchx64)
	if err != nil {
		return nil, err
	}
	if msdevSetup == nil {
		return nil, fmt.Errorf("NewWinMsdev is not implemented yet")
	}

	return &WinMsdev{
		Name: "WinMsdev",
		Msvc: msdevSetup,
	}, nil
}
