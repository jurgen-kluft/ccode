package toolchain

import (
	"fmt"

	"github.com/jurgen-kluft/ccode/clay/toolchain/dpenc"
	utils "github.com/jurgen-kluft/ccode/utils"
)

type Msdev struct {
	Name  string
	Vars  *utils.Vars
	Tools map[string]string
}

// Compiler options
//      https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/compiler-options-listed-by-category.md
// Linker options
//      https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/linker-options.md

func NewWindowsMsdev() (*Msdev, error) {
	return nil, fmt.Errorf("Msdev is not implemented yet")
}

func (ms *Msdev) NewCompiler(config *Config) Compiler {
	return nil
}
func (ms *Msdev) NewArchiver(a ArchiverType, config *Config) Archiver {
	return nil
}
func (ms *Msdev) NewLinker(config *Config) Linker {
	return nil
}
func (t *Msdev) NewBurner(config *Config) Burner {
	return &EmptyBurner{}
}
func (t *Msdev) NewDependencyTracker(dirpath string) dpenc.FileTrackr {
	return nil
}

// MsDevInstallation represents the installation of Microsoft Visual Studio that was found.
type MsDevInstallation struct {
	RootPath     string   // The root path of the installation, e.g., "C:\\Program Files (x86)\\Microsoft Visual Studio\\2019\\Community"
	Version      string   // The version of the installation, e.g., "16.0"
	Arch         string   // The architecture of the installation, e.g., "x86", "x64", "arm64"
	BinPath      string   // The path to the bin directory, e.g., "C:\\Program Files (x86)\\Microsoft Visual Studio\\2019\\Community\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64"
	CCPath       string   // The path to the cl.exe compiler, e.g., "C:\\Program Files (x86)\\Microsoft Visual Studio\\2019\\Community\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64\\cl.exe"
	CXXPath      string   // The path to the cl.exe compiler, e.g., "C:\\Program Files (x86)\\Microsoft Visual Studio\\2019\\Community\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64\\cl.exe"
	LIBPath      string   // The path to the lib directory, e.g., "C:\\Program Files (x86)\\Microsoft Visual Studio\\2019\\Community\\VC\\Tools\\MSVC\\14.29.30133\\lib\\x64"
	LDPath       string   // The path to the link.exe linker, e.g., "C:\\Program Files (x86)\\Microsoft Visual Studio\\2019\\Community\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64\\link.exe"
	RCPath       string   // The path to the rc.exe resource compiler, e.g., "C:\\Program Files (x86)\\Microsoft Visual Studio\\2019\\Community\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64\\rc.exe"
	IncludePaths []string //
	LibraryPaths []string // The paths to the library directories, e.g., "C:\\Program Files (x86)\\Microsoft Visual Studio\\2019\\Community\\VC\\Tools\\MSVC\\14.29.30133\\lib\\x64"
	CCOpts       []string // Compiler options, e.g., "/nologo /W3 /O2 /DWIN32 /D_WINDOWS /D_USRDLL /D_MBCS"
	CXXOpts      []string // C++ compiler options, e.g., "/nologo /W3 /O2 /DWIN32 /D_WINDOWS /D_USRDLL /D_MBCS"
}
