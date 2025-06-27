package clay

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jurgen-kluft/ccode/clay/toolchain"
	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
	"github.com/jurgen-kluft/ccode/foundation"
)

type IncludeMap = foundation.ValueSet

func NewIncludeMap(size int) *IncludeMap {
	return foundation.NewValueSet2(size)
}

type DefineMap = foundation.ValueSet

func NewDefineMap(size int) *DefineMap {
	return foundation.NewValueSet2(size)
}

// Project represents a C/C++ project that can be built using the Clay build system.
// A project can be a library or an executable.
type Project struct {
	Toolchain    toolchain.Environment // Build environment for this project
	IsExecutable bool                  // Is this project an executable (true) or a library (false)
	Name         string                // Name of the Library or Executable
	Config       *Config               // Build configuration
	Defines      *DefineMap            // Compiler defines (macros) for the library
	IncludeDirs  *IncludeMap           // Include paths for the library (system)
	SourceFiles  []SourceFile          // C/C++ Source files for the library
	Dependencies []*Project            // Libraries that this project depends on
	Frameworks   []string              // Frameworks to link against (for macOS)
}

func NewExecutableProject(name string, config *Config) *Project {
	return &Project{
		Name:         name,
		Config:       config,
		Toolchain:    nil, // Will be set later
		IsExecutable: true,
		Defines:      nil,
		IncludeDirs:  nil,
		SourceFiles:  nil,
		Dependencies: nil,
	}
}

func NewLibraryProject(name string, config *Config) *Project {
	return &Project{
		Name:         name,
		Config:       config,
		Toolchain:    nil, // Will be set later
		IsExecutable: false,
		Defines:      nil,
		IncludeDirs:  nil,
		SourceFiles:  nil,
		Dependencies: nil,
	}
}

type SourceFile struct {
	SrcAbsPath string // Absolute path to the source file
	SrcRelPath string // Relative path of the source file (based on where it was globbed from)
}

func (p *Project) GetOutputFilepath(buildPath, filename string) string {
	return filepath.Join(buildPath, p.Name, filename)
}

func (p *Project) GetBuildPath(buildPath string) string {
	return filepath.Join(buildPath, p.Name)
}

func (p *Project) SetToolchain(config *Config) (err error) {
	targetOS := config.Target.OSAsString()
	if targetOS == "arduino" {
		p.Toolchain, err = toolchain.NewArduinoEsp32(config.Target.ArchAsString(), p.Name)
	} else if targetOS == "windows" {
		p.Toolchain, err = toolchain.NewWinMsdev(config.Target.ArchAsString(), "Desktop")
	} else if targetOS == "mac" || targetOS == "macos" || targetOS == "darwin" {
		p.Toolchain, err = toolchain.NewDarwinClang(runtime.GOARCH, p.Frameworks)
	} else {
		err = foundation.LogErrorf(os.ErrNotExist, "error, %s as a build target on %s is not supported", targetOS, runtime.GOOS)
	}
	return err
}

func (p *Project) AddSourceFile(srcPath string, srcRelPath string) {
	p.SourceFiles = append(p.SourceFiles, SourceFile{
		SrcAbsPath: srcPath,
		SrcRelPath: srcRelPath,
	})
}

func (p *Project) AddLibrary(lib *Project) {
	p.Dependencies = append(p.Dependencies, lib)
}

func CopyConfig(config *Config) *toolchain.Config {
	return toolchain.NewConfig(config.Config, config.Target)
}

func (p *Project) Build(buildConfig *Config, buildPath string) (outOfDate int, err error) {

	compiler := p.Toolchain.NewCompiler(CopyConfig(buildConfig))
	compiler.SetupArgs(p.Defines.Values, p.IncludeDirs.Values)

	projectBuildPath := p.GetBuildPath(buildPath)
	projectDepFileTrackr := p.Toolchain.NewDependencyTracker(projectBuildPath)

	srcFilesOutOfDate := make([]SourceFile, 0, len(p.SourceFiles))
	srcFilesUpToDate := make([]SourceFile, 0, len(p.SourceFiles))
	for _, src := range p.SourceFiles {
		srcObjRelPath := filepath.Join(projectBuildPath, compiler.ObjFilepath(src.SrcRelPath))
		if !projectDepFileTrackr.QueryItem(srcObjRelPath) {
			foundation.DirMake(filepath.Dir(srcObjRelPath))
			srcFilesOutOfDate = append(srcFilesOutOfDate, src)
		} else {
			srcFilesUpToDate = append(srcFilesUpToDate, src)
		}
	}

	absSrcFilepaths := make([]string, len(srcFilesOutOfDate))
	objRelFilepaths := make([]string, len(srcFilesOutOfDate))
	for i, src := range srcFilesOutOfDate {
		absSrcFilepaths[i] = src.SrcAbsPath
		objRelFilepaths[i] = filepath.Join(projectBuildPath, compiler.ObjFilepath(src.SrcRelPath))
	}

	buildStartTime := time.Now()
	outOfDate = len(srcFilesOutOfDate)
	if outOfDate > 0 {
		foundation.LogInfof("Building project: %s, config: %s\n", p.Name, p.Config.String())
		buildStartTime = time.Now()

		// Give the compiler the array of out-of-date source files (input) and their object files (output)
		if err := compiler.Compile(absSrcFilepaths, objRelFilepaths); err != nil {
			return outOfDate, err
		}

		// Update the dependency tracker
		for _, src := range srcFilesUpToDate {
			objRelFilepath := filepath.Join(projectBuildPath, compiler.ObjFilepath(src.SrcRelPath))
			projectDepFileTrackr.CopyItem(objRelFilepath)
		}
		for _, src := range srcFilesOutOfDate {
			depFilepath := filepath.Join(projectBuildPath, compiler.DepFilepath(src.SrcRelPath))
			if mainItem, depItems, err := deptrackr.ParseDotDependencyFile(depFilepath); err == nil {
				projectDepFileTrackr.AddItem(mainItem, depItems)
			}
		}
	}

	staticArchiver := p.Toolchain.NewArchiver(toolchain.ArchiverTypeStatic, CopyConfig(buildConfig))

	if p.IsExecutable {
		linker := p.Toolchain.NewLinker(CopyConfig(buildConfig))
		linker.SetupArgs([]string{}, []string{})

		executableOutputFilepath := p.GetOutputFilepath(buildPath, linker.LinkedFilepath(p.Name))

		if outOfDate > 0 || !projectDepFileTrackr.QueryItem(executableOutputFilepath) {
			if outOfDate == 0 {
				foundation.LogInfof("Linking project: %s, config: %s\n", p.Name, p.Config.String())
				buildStartTime = time.Now()
				outOfDate += 1
			}

			// Project object files
			archivesToLink := make([]string, 0, len(p.SourceFiles)+len(p.Dependencies))
			for _, src := range p.SourceFiles {
				srcObjRelPath := filepath.Join(projectBuildPath, compiler.ObjFilepath(src.SrcRelPath))
				archivesToLink = append(archivesToLink, srcObjRelPath)
			}
			// Project archive dependencies (only those matching the build config)
			for _, dep := range p.Dependencies {
				if dep.Config.Matches(buildConfig) {
					libAbsFilepath := dep.GetOutputFilepath(buildPath, staticArchiver.LibFilepath(dep.Name))
					archivesToLink = append(archivesToLink, libAbsFilepath)
				}
			}
			// Link them all together into a single executable
			if err := linker.Link(archivesToLink, executableOutputFilepath); err != nil {
				return outOfDate, err
			}
			projectDepFileTrackr.AddItem(executableOutputFilepath, archivesToLink)

		} else {
			projectDepFileTrackr.CopyItem(executableOutputFilepath)
		}

	} else {
		archiveOutputFilepath := p.GetOutputFilepath(buildPath, staticArchiver.LibFilepath(p.Name))
		if outOfDate > 0 || !projectDepFileTrackr.QueryItem(archiveOutputFilepath) {
			if outOfDate == 0 {
				foundation.LogInfof("Archiving project: %s, config: %s\n", p.Name, p.Config.String())
				buildStartTime = time.Now()
				outOfDate += 1
			}

			// If this is a library, we don't link it, but we can create a static archive
			staticArchiver.SetupArgs()

			objFilesToArchive := make([]string, 0, len(p.SourceFiles))
			for _, src := range p.SourceFiles {
				srcObjRelPath := filepath.Join(projectBuildPath, compiler.ObjFilepath(src.SrcRelPath))
				objFilesToArchive = append(objFilesToArchive, srcObjRelPath)
			}
			if err := staticArchiver.Archive(objFilesToArchive, archiveOutputFilepath); err != nil {
				return outOfDate, err
			}

			projectDepFileTrackr.AddItem(archiveOutputFilepath, objFilesToArchive)
		} else {
			projectDepFileTrackr.CopyItem(archiveOutputFilepath)
		}
	}

	_, err = projectDepFileTrackr.Save()
	if err != nil {
		return outOfDate, err
	}

	if outOfDate > 0 {
		foundation.LogInfof("Building done ... (duration %s)\n", time.Since(buildStartTime).Round(time.Second))
	}

	return outOfDate, nil
}

func (p *Project) Flash(buildConfig *Config, buildPath string) error {
	burner := p.Toolchain.NewBurner(CopyConfig(buildConfig))

	buildPath = p.GetBuildPath(buildPath)

	burner.SetupBuild(buildPath)
	if err := burner.Build(); err != nil {
		return err
	}

	if err := burner.SetupBurn(buildPath); err != nil {
		return err
	}
	if err := burner.Burn(); err != nil {
		return err
	}

	return nil
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

func (p *Project) AddSourceFilesFrom(srcPath string, options AddSourceFileOptions) {
	handleDir := func(rootPath, relPath string) bool {
		return len(relPath) == 0 || HasOption(options, OptionAddRecursively)
	}
	handleFile := func(rootPath, relPath string) {
		isCpp := HasOption(options, OptionAddCppFiles) && (filepath.Ext(relPath) == ".cpp" || filepath.Ext(relPath) == ".cxx")
		isC := !isCpp && (HasOption(options, OptionAddCFiles) && filepath.Ext(relPath) == ".c")
		if isCpp || isC {
			p.AddSourceFile(filepath.Join(rootPath, relPath), relPath)
		}
	}

	foundation.FileEnumerate(srcPath, handleDir, handleFile)
}
