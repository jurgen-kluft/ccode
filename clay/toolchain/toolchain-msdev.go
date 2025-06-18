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
	Setup *MsDevSetup
}

// Compiler options
//      https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/compiler-options-listed-by-category.md
// Linker options
//      https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/linker-options.md

func (ms *WinMsdev) NewCompiler(config *Config) Compiler {
	// PathToTools\CL.exe
	// /c
	// /I..\..\..\ccore\source\main\include
	// /Zi                                                            Generates complete debugging information.
	// /nologo                                                        Suppresses display of sign-on banner.
	// /W3                                                            Set output warning level.
	// /WX                                                            Treat warnings as errors.
	// /diagnostics:column                                            Diagnostics format: prints column information.
	// /sdl                                                           Enable more security features and warnings.
	// /MP                                                            Builds multiple source files concurrently.
	// /Od                                                            Disables optimization.
	// /Oy-                                                           Omits frame pointer
	// /D TARGET_DEBUG
	// /D _DEBUG
	// /D TARGET_TEST
	// /D TARGET_PC
	// /D _UNICODE
	// /D UNICODE
	// /D CCORE_GEN_CPU_X86_64
	// /D CCORE_GEN_OS_WINDOWS
	// /D CCORE_GEN_COMPILER_VC
	// /D CCORE_GEN_GENERATOR_VS2022
	// /D CCORE_GEN_CONFIG_DEBUGTEST
	// /D "CCORE_GEN_PLATFORM_NAME=\"WINDOWS\""
	// /D CCORE_GEN_PROJECT_CCORE
	// /D CCORE_GEN_TYPE_CPP_LIB
	// /D _UNICODE
	// /D UNICODE
	// /Gm-                                                           Deprecated; disables minimal rebuild.
	// /EHsc                                                          Enable standard C++ exception handling.
	// /MTd                                                           Use the multithreaded debug version of the C runtime library.
	// /GS                                                            Enable security checks to detect buffer overflows.
	// /fp:precise                                                    Floating-point model: precise.
	// /Zc:wchar_t                                                    Treats wchar_t as a built-in type.
	// /Zc:forScope                                                   Enables C++ scoping rules for for-loop variables.
	// /Zc:inline                                                     Enables C++ inline function semantics.
	// /std:c++17                                                     Specifies the C++ language standard to use (C++17).
	// /Fo"obj\ccore\DebugTest_x86_64_v143\\"                         Output directory for object files.
	// /Fd"lib\ccore\DebugTest_x86_64_v143\ccore.pdb"                 Output filepath for program database (PDB)
	// /external:W3                                                   External compiler warnings: set to level 3.
	// /Gd                                                            Use the default calling convention (Cdecl).
	// /TP                                                            Treats all source files as C++ files.
	// /FC                                                            Full path of source files in diagnostics.
	// /errorReport:prompt                                            Deprecated; prompts for error report.
	// ..\..\..\ccore\source\main\cpp\c_allocator.cpp
	// ..\..\..\ccore\source\main\cpp\c_binary_search.cpp
	// ..\..\..\ccore\source\main\cpp\c_binmap1.cpp
	// ..\..\..\ccore\source\main\cpp\c_debug.cpp
	// ..\..\..\ccore\source\main\cpp\c_error.cpp
	// ..\..\..\ccore\source\main\cpp\c_memcpy_neon.cpp
	// ..\..\..\ccore\source\main\cpp\c_memcpy_sse.cpp
	// ..\..\..\ccore\source\main\cpp\c_memory.cpp
	// ..\..\..\ccore\source\main\cpp\c_qsort.cpp

	return nil
}

func (ms *WinMsdev) NewArchiver(a ArchiverType, config *Config) Archiver {
	return nil
}

func (ms *WinMsdev) NewLinker(config *Config) Linker {
	return nil
}

func (t *WinMsdev) NewBurner(config *Config) Burner {
	return &EmptyBurner{}
}

func (t *WinMsdev) NewDependencyTracker(dirpath string) deptrackr.FileTrackr {
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

	return &WinMsdev{
		Name:  "WinMsdev",
		Vars:  foundation.NewVars(),
		Tools: make(map[string]string),
		Setup: msdevSetup,
	}, nil
}

func determineMsDevSetup(arch string, product string) *MsDevSetup {
	// This function should determine the installation of Microsoft Visual Studio
	// and return a MsDevSetup struct with the relevant paths and options.
	// For now, we return nil to indicate that this is not implemented yet.
	return nil
}
