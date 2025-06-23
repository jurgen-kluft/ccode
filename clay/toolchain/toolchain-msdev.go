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
	Name  string
	Vars  *foundation.Vars
	Tools map[string]string
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Compiler

// Compiler options
//   - https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/compiler-options-listed-by-category.md
//

type WinMsDevCompiler struct {
	toolChain       *WinMsdev             // The toolchain this compiler belongs to
	config          *Config               // Build configuration
	cCompilerPath   string                // Path to the C compiler executable (e.g., cl.exe)
	cppCompilerPath string                // Path to the C++ compiler executable (e.g., cl.exe)
	cArgs           *foundation.Arguments // Arguments for the C compiler
	cppArgs         *foundation.Arguments // Arguments for the C++ compiler
}

func (m *WinMsdev) NewCompiler(config *Config) Compiler {
	return &WinMsDevCompiler{
		toolChain:       m,
		config:          config,
		cCompilerPath:   m.Vars.GetFirstOrEmpty("c.compiler"),
		cppCompilerPath: m.Vars.GetFirstOrEmpty("cpp.compiler"),
		cArgs:           foundation.NewArguments(512),
		cppArgs:         foundation.NewArguments(512),
	}
}

func (cl *WinMsDevCompiler) ObjFilepath(srcRelFilepath string) string {
	srcRelFilepath = strings.TrimSuffix(srcRelFilepath, ".c")
	srcRelFilepath = strings.TrimSuffix(srcRelFilepath, ".cpp")
	return srcRelFilepath + ".obj"
}

func WindowsCompilerIncludePath(incl string) string {
	return "/I" + foundation.PathWindowsPath(incl)
}

func (cl *WinMsDevCompiler) DepFilepath(objRelFilepath string) string {
	objRelFilepath = strings.TrimSuffix(objRelFilepath, ".c")
	objRelFilepath = strings.TrimSuffix(objRelFilepath, ".cpp")
	return objRelFilepath + ".d"
}

func (cl *WinMsDevCompiler) SetupArgs(_defines []string, _includes []string) {
	argsArray := []*foundation.Arguments{cl.cArgs, cl.cppArgs}
	compilerOptions := []string{"c.compiler.options", "cpp.compiler.options"}
	compilerIncludes := []string{"c.compiler.includes", "cpp.compiler.includes"}
	for i, args := range argsArray {
		isCpp := i == 1 // C++ compiler is the second in the array

		args.Add("/c")                  // Compile only, do not link.
		args.Add("/nologo")             // Suppress display of sign-on banner.
		args.Add("/diagnostics:column") // Diagnostics format: prints column information.

		if options, ok := cl.toolChain.Vars.Get(compilerOptions[i]); ok {
			args.Add(options...)
		}

		// Add include directories
		if cppIncludes, ok := cl.toolChain.Vars.Get(compilerIncludes[i]); ok {
			args.AddWithFunc(WindowsCompilerIncludePath, cppIncludes...)
		}
		args.AddWithFunc(WindowsCompilerIncludePath, _includes...)

		args.Add("/W3") // Set output warning level to 3 (high warnings).
		args.Add("/WX") // Treat warnings as errors.
		args.Add("/MP") // Build multiple source files concurrently.

		if cl.config.Config.IsDebug() {
			args.Add("/Od")  // Disable optimizations for debugging.
			args.Add("/Zi")  // Generate complete debugging information.
			args.Add("/Oy-") // Do not omit frame pointer.
			args.Add("/MTd") // Use the multithreaded debug version of the C runtime library.
		} else {
			args.Add("/O2")  // Optimize for speed.
			args.Add("/Ob2") // Enable inline expansion for functions that are small and frequently called.
			args.Add("/Oi")  // Enable intrinsic functions.
			args.Add("/Oy")  // Omit frame pointer for functions that do not require one.
			args.Add("/MT")  // Use the multithreaded version of the C runtime library.
		}

		args.AddWithPrefix("/D", _defines...)

		args.Add("/sdl") // Enable more security features and warnings.
		args.Add("/GS")  // Enable security checks to detect buffer overflows.

		if isCpp && cl.config.Config.IsTest() {
			args.Add("/EHsc") // Enable standard C++ exception handling.
		}

		args.Add("/fp:precise") // Floating-point model: precise.
		args.Add("/Zc:wchar_t") // Treats wchar_t as a built-in type.

		if isCpp {
			args.Add("/std:c++17") // Specifies the C++ language standard to use (C++17).
			args.Add("/TP")        // Treats all source files as C++ files.
		} else {
			args.Add("/std:c17") // Specifies the C language standard to use (C17).
		}
		args.Add("/FC") // Full path of source files in diagnostics.
	}
}

func (cl *WinMsDevCompiler) Compile(sourceAbsFilepaths []string, objRelFilepaths []string) error {
	// Analyze all the object filepaths and organize them per directory, we do this because
	// the MSVC compiler outputs object files into a single directory, we do not want this.
	// And we do not want to call the compiler for each source file, but rather for each directory.
	sourceFilesPerDir := make(map[string][]string)
	for i, objFilepath := range objRelFilepaths {
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
		compilerPath := cl.cppCompilerPath
		compilerArgs := cl.cppArgs.Args

		srcDir = foundation.PathWindowsPath(srcDir)
		if len(srcDir) > 0 {
			compilerArgs = append(compilerArgs, "/Fo\""+srcDir+"\"")
		}

		compilerArgs = append(compilerArgs, "/sourceDependencies")
		compilerArgs = append(compilerArgs, srcDir)

		configStr := cl.config.Config.AsString()
		for _, srcFile := range srcFiles {
			srcFile = foundation.PathWindowsPath(srcFile)
			compilerArgs = append(compilerArgs, srcFile)
			foundation.LogInfof("Compiling (%s) %s\n", configStr, srcFile)
		}

		// Prepare the command to execute the compiler.
		var cmd *exec.Cmd
		cmd = exec.Command(compilerPath, compilerArgs...)
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

func (ms *WinMsdev) NewArchiver(a ArchiverType, config *Config) Archiver {
	return &WinMsDevArchiver{
		toolChain:    ms,
		config:       config,
		archiverPath: ms.Vars.GetFirstOrEmpty("archiver.path"),
		args:         foundation.NewArguments(512),
	}
}

type WinMsDevArchiver struct {
	toolChain    *WinMsdev             // The toolchain this archiver belongs to
	config       *Config               // Build configuration
	archiverPath string                // Path to the archiver executable (e.g., lib.exe)
	args         *foundation.Arguments // Arguments for the archiver
}

func (a *WinMsDevArchiver) LibFilepath(_filepath string) string {
	filename := foundation.PathFilename(_filepath, true)
	dirpath := foundation.PathDirname(_filepath)
	return filepath.Join(dirpath, filename+".lib")
}

func (a *WinMsDevArchiver) SetupArgs() {
	a.args.Add("/NOLOGO")      // Suppress display of sign-on banner.
	a.args.Add("/MACHINE:X64") // Specify the target machine architecture (x64).
}

func (a *WinMsDevArchiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {
	archiverPath := a.archiverPath
	archiverArgs := a.args.Args
	archiverArgs = append(archiverArgs, "/OUT:\""+foundation.PathWindowsPath(outputArchiveFilepath)+"\"")

	for _, objFile := range inputObjectFilepaths {
		archiverArgs = append(archiverArgs, "\""+foundation.PathWindowsPath(objFile)+"\"")
	}

	foundation.LogInfof("Archiving (%s) %s\n", a.config.Config.AsString(), outputArchiveFilepath)

	cmd := exec.Command(archiverPath, archiverArgs...)
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

// Linker options
//	- https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/linker-options.md

func (ms *WinMsdev) NewLinker(config *Config) Linker {
	return &WinMsDevLinker{
		toolChain:  ms,
		config:     config,
		linkerPath: ms.Vars.GetFirstOrEmpty("linker.path"),
		args:       foundation.NewArguments(512),
	}
}

type WinMsDevLinker struct {
	toolChain  *WinMsdev             // The toolchain this archiver belongs to
	config     *Config               // Build configuration
	linkerPath string                // Path to the linker executable (e.g., ld.exe)
	args       *foundation.Arguments // Arguments for the linker
}

func (l *WinMsDevLinker) LinkedFilepath(filepath string) string {
	return filepath + ".exe"
}

func (l *WinMsDevLinker) SetupArgs(libraryPaths []string, libraryFiles []string) {

	// Note: According to the MSVC documentation, the linker options can be specified in any order.

	l.args.Add("/ERRORREPORT:PROMPT")
	l.args.Add("/NOLOGO")

	// TODO: Need to figure out if we want to use the manifest options.
	// /MANIFEST
	// /MANIFESTUAC:"level='asInvoker' uiAccess='false'"
	// /manifest:embed

	if l.config.Config.IsDebug() {
		l.args.Add("/DEBUG") // Generate debug information.
		l.args.Add("/MTd")   // Use the multithreaded debug version of the C runtime library.
	} else if l.config.Config.IsRelease() {
		l.args.Add("/MT") // Use the multithreaded version of the C runtime library.
	} else if l.config.Config.IsFinal() {
		l.args.Add("/OPT:REF")        // Enable optimization for references.
		l.args.Add("/OPT:ICF")        // Enable identical COMDAT folding.
		l.args.Add("/LTCG")           // Enable link-time code generation.
		l.args.Add("/INCREMENTAL:NO") // Disable incremental linking.
		l.args.Add("/MT")             // Use the multithreaded version of the C runtime library.
	}

	// vars.Set("linker.libraries", msdevSetup.Libs...)
	// vars.Set("linker.library.paths", msdevSetup.LibPaths...)
	if libraryPaths, ok := l.toolChain.Vars.Get("linker.library.paths"); ok {
		for _, libPath := range libraryPaths {
			l.args.Add("/LIBPATH:\"" + foundation.PathWindowsPath(libPath) + "\"")
		}
	}

	// What is this used for?
	// /TLBID:1

	l.args.Add("/SUBSYSTEM:CONSOLE") // Specify the subsystem type (console application).
	l.args.Add("/DYNAMICBASE")       // Enable ASLR (Address Space Layout Randomization).
	l.args.Add("/NXCOMPAT")          // Enable DEP (Data Execution Prevention).
	l.args.Add("/MACHINE:X64")       // Specify the target machine architecture (x64).
}

func (l *WinMsDevLinker) Link(inputArchiveAbsFilepaths []string, outputAppRelFilepath string) error {

	linkerPath := l.linkerPath
	linkerArgs := l.args.Args

	outputAppRelFilepath = foundation.PathWindowsPath(outputAppRelFilepath)

	linkerArgs = append(linkerArgs, "/OUT:\""+outputAppRelFilepath+"\"")
	linkerArgs = append(linkerArgs, "/MAP:"+foundation.FileChangeExtension(outputAppRelFilepath, ".map"))

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
	for _, systemLib := range systemLibraries {
		linkerArgs = append(linkerArgs, "\""+systemLib+"\"")
	}

	for _, archiveFile := range inputArchiveAbsFilepaths {
		archiveFile = foundation.PathWindowsPath(archiveFile)
		linkerArgs = append(linkerArgs, "\""+archiveFile+"\"")
	}

	foundation.LogInfof("Linking (%s) %s\n", l.config.Config.AsString(), outputAppRelFilepath)

	cmd := exec.Command(linkerPath, linkerArgs...)
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
	return nil
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

	msdevOptions := map[string][]string{
		"dynamic.debugging":              {"/dynamicdeopt"},       // Enable dynamic debugging; allows for dynamic analysis of the program.
		"optimize.for.size":              {"/O1"},                 // Optimize for size; enables optimizations that reduce code size.
		"optimize.for.speed":             {"/O2"},                 // Optimize for speed; enables most optimizations.
		"optimize.none":                  {"/Od"},                 // No optimizations; useful for debugging.
		"optimize.global":                {"/Og"},                 // Global optimizations, including inlining and loop unrolling.
		"optimize.full":                  {"/Ox"},                 // Full optimization, including speed and size.
		"optimize.favor.small":           {"/Os"},                 // Optimize for size, not speed.
		"optimize.favor.fast":            {"/Ot"},                 // Optimize for speed, not size.
		"generate.debug.info":            {"/Z7", "/Zi"},          // Generate debug information; `/Z7` for embedded, `/Zi` for separate PDB.
		"generate.intrinsic.functions":   {"/Oi"},                 // Enable intrinsic functions (e.g., `_memcpy`, `_memset`).
		"omit.frame.pointer":             {"/Oy"},                 // Omit frame pointer for functions that do not require one.
		"inline.expansion.level.0":       {"/Ob0"},                // Disable inline expansion; functions are not inlined.
		"inline.expansion.level.1":       {"/Ob1"},                // Enable inline expansion for functions that are small and frequently called.
		"inline.expansion.level.2":       {"/Ob2"},                // Enable inline expansion for functions that are small and frequently called, with more aggressive inlining.
		"inline.expansion.level.3":       {"/Ob3"},                // Enable inline expansion for functions that are small and frequently called, with the most aggressive inlining.
		"exception.handling.std":         {"/EHsc"},               // Default: enable C++ exception handling with standard semantics.
		"exception.handling.async":       {"/EHa"},                // Enable C++ exception handling (with SEH exceptions).
		"exception.handling.c":           {"/EHc"},                // `extern "C"` defaults to `nothrow`.
		"exception.handling.r":           {"/EHr"},                // Always generate `noexcept` runtime termination checks.
		"exception.handling.s":           {"/EHs"},                // Enable C++ exception handling (no SEH exceptions).
		"fp.behavior.contract":           {"/fp:contract"},        // Consider floating-point contractions when generating code.
		"fp.behavior.except":             {"/fp:except"},          // Consider floating-point exceptions when generating code.
		"fp.behavior.fast":               {"/fp:fast"},            // "fast" floating-point model; results are less predictable.
		"fp.behavior.precise":            {"/fp:precise"},         // "precise" floating-point model; results are predictable.
		"fp.behavior.strict":             {"/fp:strict"},          // "strict" floating-point model (implies `/fp:except`).
		"string.pooling":                 {"/GF"},                 // Enable string pooling to reduce memory usage.
		"whole.program.optimization":     {"/Gw"},                 // Enable whole program optimization.
		"rtti.enable":                    {"/GR"},                 // Enable run-time type information (RTTI).
		"rtti.disable":                   {"/GR-"},                // Disable run-time type information (RTTI).
		"function.level.linking.enable":  {"/Gy"},                 // Enable function-level linking.
		"function.level.linking.disable": {"/Gy-"},                // Disable function-level linking.
		"map.filepath":                   {"/Fm"},                 // Create a map file.
		"exe.filepath":                   {"/Fe"},                 // Specify the output executable file path.
		"pdb.filepath":                   {"/Fd"},                 // Specify the output program database (PDB) file path.
		"compiler.defines":               {"/D"},                  // Define preprocessor macros.
		"compiler.includes":              {"/I"},                  // Specify additional include directories.
		"generate.dependency.files":      {"/sourceDependencies"}, // Generate source-level dependency files.
		"build.concurrently":             {"/MP"},                 // Build multiple source files concurrently.
		"compile.all.as.c":               {"/TC"},                 // Treat all source files as C.
		"compile.all.as.cpp":             {"/TP"},                 // Treat all source files
		"warnings.disable.all":           {"/w"},                  // Disable all warnings.
		"warnings.enable.all":            {"/Wall"},               // Enable all warnings, including those disabled by default.
		"warnings.are.errors":            {"/WX"},                 // Treat all warnings as errors.
		"warnings.output.level.0":        {"/W0"},                 // Set output warning level to 0 (no warnings).
		"warnings.output.level.1":        {"/W1"},                 // Set output warning level to 1 (basic warnings).
		"warnings.output.level.2":        {"/W2"},                 // Set output warning level to 2 (moderate warnings).
		"warnings.output.level.3":        {"/W3"},                 // Set output warning level to 3 (high warnings).
		"warnings.output.level.4":        {"/W4"},                 // Set output warning level to 4 (very high warnings).
		"c++14":                          {"/std:c++14"},          // Specify the C++ standard version (c++14, c++17, c++20, c++latest).
		"c++17":                          {"/std:c++17"},          // Specify the C++ standard version (c++14, c++17, c++20, c++latest).
		"c++20":                          {"/std:c++20"},          // Specify the C++ standard version (c++14, c++17, c++20, c++latest).
		"c++latest":                      {"/std:c++latest"},      // Specify the C++ standard version (c++14, c++17, c++20, c++latest).
		"c11":                            {"/std:c11"},            // Specify the C standard version (c11, c17, clatest).
		"c17":                            {"/std:c17"},            // Specify the C standard version (c11, c17, clatest).
		"clatest":                        {"/std:clatest"},        // Specify the C standard version (c11, c17, clatest).
		"link.dll":                       {"/LD"},                 // Generate a dynamic-link library (DLL).
		"link.debug.dll":                 {"/LDd"},                // Generate a debug dynamic
		"link.multithreaded.dll":         {"/MD"},                 // Generate a multithreaded DLL, by using *MSVCRT.lib*
		"link.multithreaded.debug.dll":   {"/MDd"},                // Generate a debug multithreaded DLL, by using *MSVCRTD.lib*
		"link.multithreaded.exe":         {"/MT"},                 // Generate a multithreaded executable file, by using *LIBCMT.lib*
		"link.multithreaded.debug.exe":   {"/MTd"},                // Generate a debug multithreaded executable file, by using *LIBCMTD.lib*
	}

	vars := foundation.NewVars()
	vars.Set("c.compiler", msdevSetup.CompilerBin)
	vars.Set("c.compiler.includes", msdevSetup.IncludePaths...)
	vars.Set("c.compiler.options", msdevSetup.CcOptions...)

	vars.Set("cpp.compiler", msdevSetup.CompilerBin)
	vars.Set("cpp.compiler.includes", msdevSetup.IncludePaths...)
	vars.Set("cpp.compiler.options", msdevSetup.CxxOptions...)

	vars.Set("archiver.path", msdevSetup.ArchiverBin)
	vars.Set("rc.compiler", msdevSetup.RcBin)

	vars.Set("linker.path", msdevSetup.LinkerBin)
	vars.Set("linker.libraries", msdevSetup.Libs...)
	vars.Set("linker.library.paths", msdevSetup.LibPaths...)

	vars.SetMany(msdevOptions)

	return &WinMsdev{
		Name:  "WinMsdev",
		Vars:  vars,
		Tools: make(map[string]string),
	}, nil
}
