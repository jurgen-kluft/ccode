package toolchain

import (
	"fmt"

	"github.com/jurgen-kluft/ccode/clay/toolchain/dpenc"
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

func (t *WinMsdev) NewDependencyTracker(dirpath string) dpenc.FileTrackr {
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
