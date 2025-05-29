package clay

import (
	"path/filepath"

	"github.com/jurgen-kluft/ccode/dev"
	cutils "github.com/jurgen-kluft/ccode/utils"
)

type Config struct {
	Config dev.BuildConfig
	Target dev.BuildTarget
}

func NewConfig(os, cpu, build, variant string) *Config {
	config := dev.BuildConfigFromString(build, variant)
	target := dev.BuildTargetFromString(os, cpu)
	return &Config{
		Config: config,
		Target: target,
	}
}

// GetSubDir returns a subdirectory name based on the OS, CPU, Build type, and Variant.
// Example: "linux-x86-release-dev" or "arduino-esp32-debug-prod".
func (c *Config) GetSubDir() string {
	return c.Target.OSAsString() + "-" + c.Target.ArchAsString() + "-" + c.Config.Build() + "-" + c.Config.Variant()
}

func (c *Config) Matches(other *Config) bool {
	if c == nil || other == nil {
		return false
	}
	return c.Target.IsEqual(other.Target) && c.Config.Contains(other.Config)
}

type SourceFile struct {
	SrcAbsPath string // Absolute path to the source file
	SrcRelPath string // Relative path of the source file (based on where it was globbed from)
}

// Library represents a C/C++ library/archive that can be linked with other
// libraries/archives to form an executable binary.
type Library struct {
	Name        string       // Name of the library
	Config      *Config      // Config to identify the library (debug, release, final, test)
	Defines     *ValueSet    // Compiler defines (macros) for the library
	IncludeDirs *IncludeMap  // Include paths for the library (system)
	SourceFiles []SourceFile // C/C++ Source files for the library
}

func NewLibrary(name string, config *Config) *Library {
	return &Library{
		Name:        name,
		Config:      config,
		Defines:     NewValueSet(),
		IncludeDirs: NewIncludeMap(),
		SourceFiles: make([]SourceFile, 0),
	}
}

func (lib *Library) GetOutputFilepath(buildPath string, extension string) string {
	return filepath.Join(buildPath, lib.Name, lib.Name) + extension
}

func (lib *Library) GetOutputDirpath(buildPath string) string {
	return filepath.Join(buildPath, lib.Name)
}

func (lib *Library) AddSourceFile(srcPath string, srcRelPath string) {
	lib.SourceFiles = append(lib.SourceFiles, SourceFile{
		SrcAbsPath: srcPath,
		SrcRelPath: srcRelPath,
	})
}

type AddSourceFileOptions int

const (
	OptionAddCppFiles            AddSourceFileOptions = 1
	OptionAddCFiles              AddSourceFileOptions = 2
	OptionAddRecursively         AddSourceFileOptions = 4
	OptionRecursivelyAddCppFiles AddSourceFileOptions = OptionAddCppFiles | OptionAddRecursively
	OptionRecursivelyAddCFiles   AddSourceFileOptions = OptionAddCFiles | OptionAddRecursively
)

func HasOption(options AddSourceFileOptions, option AddSourceFileOptions) bool {
	return (options & option) != 0
}

func (lib *Library) AddSourceFilesFrom(srcPath string, options AddSourceFileOptions) {
	handleDir := func(rootPath, relPath string) bool {
		return len(relPath) == 0 || HasOption(options, OptionAddRecursively)
	}
	handleFile := func(rootPath, relPath string) {
		isCpp := HasOption(options, OptionAddCppFiles) && (filepath.Ext(relPath) == ".cpp" || filepath.Ext(relPath) == ".cxx")
		isC := !isCpp && (HasOption(options, OptionAddCFiles) && filepath.Ext(relPath) == ".c")
		if isCpp || isC {
			lib.AddSourceFile(filepath.Join(rootPath, relPath), relPath)
		}
	}

	cutils.AddFilesFrom(srcPath, handleDir, handleFile)
}

// Executable represents a C/C++ executable that can be built using the Clay build system.
type Executable struct {
	Name      string     // Name of the executable
	Libraries []*Library // Libraries that this executable is linking with
}

func NewExecutable(name string) *Executable {
	return &Executable{
		Name:      name,
		Libraries: make([]*Library, 0),
	}
}

func (exe *Executable) GetOutputFilename(extension string) string {
	return exe.Name + extension
}

func (exe *Executable) GetOutputFilepath(builddir string, extension string) string {
	return filepath.Join(builddir, exe.Name) + extension
}

func (exe *Executable) AddLibrary(lib *Library) {
	exe.Libraries = append(exe.Libraries, lib)
}
