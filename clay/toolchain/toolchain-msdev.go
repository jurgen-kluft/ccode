package toolchain

import (
	"fmt"

	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
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
	toolChain       *WinMsdev  // The toolchain this compiler belongs to
	config          *Config    // Build configuration
	cCompilerPath   string     // Path to the C compiler executable (e.g., cl.exe)
	cppCompilerPath string     // Path to the C++ compiler executable (e.g., cl.exe)
	cArgs           *Arguments // Arguments for the C compiler
	cppArgs         *Arguments // Arguments for the C++ compiler
}

func (m *WinMsdev) NewCompiler(config *Config) Compiler {
	return &WinMsDevCompiler{
		toolChain:       m,
		config:          config,
		cCompilerPath:   m.Vars.GetFirstOrEmpty("c.compiler"),
		cppCompilerPath: m.Vars.GetFirstOrEmpty("cpp.compiler"),
		cArgs:           NewArguments(512),
		cppArgs:         NewArguments(512),
	}
}

func (cl *WinMsDevCompiler) ObjFilepath(srcRelFilepath string) string {
	return srcRelFilepath + ".obj"
}

func (cl *WinMsDevCompiler) DepFilepath(objRelFilepath string) string {
	return objRelFilepath + ".d"
}

func (cl *WinMsDevCompiler) SetupArgs(_defines []string, _includes []string) {

	// PathToTools\CL.exe
	// /c
	// /I{compiler.includes}

	// /nologo                                                        Suppresses display of sign-on banner.
	// /diagnostics:column                                            Diagnostics format: prints column information.

	// /W3                                                            Set output warning level.
	// /WX                                                            Treat warnings as errors.
	// /MP                                                            Builds multiple source files concurrently.

	// DEBUG
	// /Zi                                                            Generates complete debugging information.
	// /Od                                                            Disables optimization.
	// /Oy-                                                           Omits frame pointer

	// /D{compiler.defines}

	// /sdl                                                           Enable more security features and warnings.
	// /GS                                                            Enable security checks to detect buffer overflows.

	// TEST
	// /EHsc                                                          Enable standard C++ exception handling.

	// /MTd                                                           Use the multithreaded debug version of the C runtime library.
	// /fp:precise                                                    Floating-point model: precise.
	// /Zc:wchar_t                                                    Treats wchar_t as a built-in type.
	// /std:c++17                                                     Specifies the C++ language standard to use (C++17).

	// /Fo"obj\ccore\DebugTest_x86_64_v143\\"                         Output directory for object files.
	// /Fd"lib\ccore\DebugTest_x86_64_v143\ccore.pdb"                 Output filepath for program database (PDB)

	// /TP                                                            Treats all source files as C++ files.
	// /FC                                                            Full path of source files in diagnostics.

	// ..\..\..\ccore\source\main\cpp\c_allocator.cpp
	// ..\..\..\ccore\source\main\cpp\c_binary_search.cpp
	// ..\..\..\ccore\source\main\cpp\c_binmap1.cpp
	// ..\..\..\ccore\source\main\cpp\c_debug.cpp
	// ..\..\..\ccore\source\main\cpp\c_error.cpp
	// ..\..\..\ccore\source\main\cpp\c_memcpy_neon.cpp
	// ..\..\..\ccore\source\main\cpp\c_memcpy_sse.cpp
	// ..\..\..\ccore\source\main\cpp\c_memory.cpp
	// ..\..\..\ccore\source\main\cpp\c_qsort.cpp

}

func (cl *WinMsDevCompiler) Compile(sourceAbsFilepaths []string, objRelFilepaths []string) error {

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Archiver

func (ms *WinMsdev) NewArchiver(a ArchiverType, config *Config) Archiver {

	// Lib.exe /OUT:"lib\cbase\DebugTest_x86_64_v143\cbase.lib"
	// /NOLOGO
	// /MACHINE:X64
	// obj\cbase\DebugTest_x86_64_v143\c_allocator.obj
	// obj\cbase\DebugTest_x86_64_v143\c_allocator_system_win32.obj

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Linker

// Linker options
//	- https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/linker-options.md

func (ms *WinMsdev) NewLinker(config *Config) Linker {
	// link.exe
	// /ERRORREPORT:PROMPT
	// /OUT:"bin\cbase_test\DebugTest_x86_64_v143\cbase_test.exe"
	// /NOLOGO
	// lib\cunittest\DebugTest_x86_64_v143\cunittest.lib
	// lib\ccore\DebugTest_x86_64_v143\ccore.lib
	// lib\cbase\DebugTest_x86_64_v143\cbase.lib
	// kernel32.lib
	// user32.lib
	// gdi32.lib
	// winspool.lib
	// comdlg32.lib
	// advapi32.lib
	// shell32.lib
	// ole32.lib
	// oleaut32.lib
	// uuid.lib
	// odbc32.lib
	// odbccp32.lib
	// /MANIFEST
	// /MANIFESTUAC:"level='asInvoker' uiAccess='false'"
	// /manifest:embed
	// /DEBUG
	// /PDB:"bin\cbase_test\DebugTest_x86_64_v143\cbase_test.pdb"
	// /SUBSYSTEM:CONSOLE
	// /TLBID:1
	// /DYNAMICBASE
	// /NXCOMPAT
	// /IMPLIB:"bin\cbase_test\DebugTest_x86_64_v143\cbase_test.lib"
	// /MACHINE:X64
	// obj\cbase_test\DebugTest_x86_64_v143\test_allocator.obj
	// obj\cbase_test\DebugTest_x86_64_v143\test_binary_search.obj
	// ...
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

// MsDevSetup represents the installation of Microsoft Visual Studio that was found.
type MsDevSetup struct {
	RootPath     string   // The root path of the installation, e.g., "C:\\Program Files (x86)\\Microsoft Visual Studio\\2019\\Community"
	Version      string   // The version of the installation, e.g., "16.0"
	Arch         string   // The architecture of the installation, e.g., "x86", "x64", "arm64"
	BinPath      string   // The path to the bin directory, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64"
	CCPath       string   // The path to the cl.exe compiler, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64\\cl.exe"
	CXXPath      string   // The path to the cl.exe compiler, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64\\cl.exe"
	LIBPath      string   // The path to the lib directory, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\lib\\x64"
	LDPath       string   // The path to the link.exe linker, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64\\link.exe"
	RCPath       string   // The path to the rc.exe resource compiler, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64\\rc.exe"
	IncludePaths []string //
	LibraryPaths []string // The paths to the library directories, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\lib\\x64"
	CCOpts       []string // Compiler options, e.g., "/nologo /W3 /O2 /DWIN32 /D_WINDOWS /D_USRDLL /D_MBCS"
	CXXOpts      []string // C++ compiler options, e.g., "/nologo /W3 /O2 /DWIN32 /D_WINDOWS /D_USRDLL /D_MBCS"
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Toolchain for Visual Studio on Windows

func NewWinMsdev(arch string, product string) (t *WinMsdev, err error) {
	msdevSetup := determineMsDevSetup(arch, product)
	if msdevSetup == nil {
		return nil, fmt.Errorf("NewWinMsdev is not implemented yet")
	}

	vars := foundation.NewVars()

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

	// "NATIVE_SUFFIXES":         {".c", ".cpp", ".cc", ".cxx", ".lib", ".obj", ".res", ".rc"},
	// "OBJECTSUFFIX":            {".obj"},
	// "LIBSUFFIX":               {".lib"},
	// "CC":                      {"cl"},
	// "CXX":                     {"cl"},
	// "LIB":                     {"lib"},
	// "LD":                      {"link"},
	// "RCCOM":                   {"$(RC)", "$(RCOPTS)", "/fo$(@:b)", "$(_CPPDEFS)", "/i$(CPPPATH:b:q)", "$(b)"},
	// "CCCOM":                   {"$(CC)", "/c", "@RESPONSE|@|", "$(_CPPDEFS)", "/I$(CPPPATH:b:q)", "/nologo", "$(CCOPTS)", "$(CCOPTS_$(CURRENT_VARIANT:u))", "$(_USE_PCH)", "$(_USE_PDB_CC)", "/Fo$(@:b)", "$(b)"},
	// "CXXCOM":                  {"$(CC)", "/c", "@RESPONSE|@|", "$(_CPPDEFS)", "/I$(CPPPATH:b:q)", "/nologo", "$(CXXOPTS)", "$(CXXOPTS_$(CURRENT_VARIANT:u))", "$(_USE_PCH)", "$(_USE_PDB_CC)", "/Fo$(@:b)", "$(b)"},
	// "PCHCOMPILE_CC":           {"$(CC)", "/c", "$(_CPPDEFS)", "/I$(CPPPATH:b:q)", "/nologo", "$(CCOPTS)", "$(CCOPTS_$(CURRENT_VARIANT:u))", "$(_USE_PDB_CC)", "/Yc$(_PCH_HEADER)", "/Fp$(@:i1:b)", "/Fo$(@:i2:b)", "$(b:i1:b)"},
	// "PCHCOMPILE_CXX":          {"$(CXX)", "/c", "$(_CPPDEFS)", "/I$(CPPPATH:b:q)", "/nologo", "$(CXXOPTS)", "$(CXXOPTS_$(CURRENT_VARIANT:u))", "$(_USE_PDB_CC)", "/Yc$(_PCH_HEADER)", "/Fp$(@:i1:b)", "/Fo$(@:i2:b)", "$(b:i1:b)"},
	// "PROGCOM":                 {"$(LD)", "/nologo", "@RESPONSE|@|", "$(_USE_PDB_LINK)", "$(PROGOPTS)", "/LIBPATH\\:$(LIBPATH:b:q)", "$(_USE_MODDEF)", "$(LIBS:q)", "/out:$(@:b:q)", "$(b:q:p\n)"},
	// "LIBCOM":                  {"$(LIB)", "/nologo", "@RESPONSE|@|", "$(LIBOPTS)", "/out:$(@:b:q)", "$(_USE_MODDEF)", "$(b:q:p\n)"},
	// "SHLIBLINKSUFFIX":         {".lib"},
	// "SHLIBCOM":                {"$(LD)", "/DLL", "/nologo", "@RESPONSE|@|", "$(_USE_PDB_LINK)", "$(SHLIBOPTS)", "/LIBPATH\\:$(LIBPATH:b:q)", "$(_USE_MODDEF)", "$(LIBS:q)", "/out:$(@:b)", "$(b)"},

	vars.SetMany(msdevOptions)

	return &WinMsdev{
		Name:  "WinMsdev",
		Vars:  vars,
		Tools: make(map[string]string),
	}, nil
}

func determineMsDevSetup(arch string, product string) *MsDevSetup {
	// This function should determine the installation of Microsoft Visual Studio
	// and return a MsDevSetup struct with the relevant paths and options.
	// For now, we return nil to indicate that this is not implemented yet.
	return nil
}
