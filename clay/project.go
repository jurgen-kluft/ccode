package clay

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jurgen-kluft/ccode/clay/toolchain"
	"github.com/jurgen-kluft/ccode/clay/toolchain/dpenc"
	utils "github.com/jurgen-kluft/ccode/utils"
)

// TODO:
// - MINOR: Clay build CLI-APP for the user
//          90%
//          Next: emit the project and library info in the cmd.go, build
//                the app and copy/release it in the build directory
// - MEDIOR: Parse the boards.txt file and be able to extract compiler, linker and other info
// - MINOR: ESP32 S3 Target (could be automatic if we are able to fully parse the boards.txt file)
// - MAJOR: To reduce compile/link time we need Dependency Tracking (Database)
//   - source file <-> [command-line args], [object-file + header-files], [tools?]

// DONE:
// - MINOR: Clay build CLI-APP for the user
//          build, clean, build-info, list libraries, list boards, list flash sizes (done)
//          flash (done)
// - MINOR: Build Info (build_info.h and build_info.cpp as a Library)
// - MINOR: Archiver
// - MINOR: Linker
// - MINOR: Image Generator
// - MINOR: Elf Size Stats

// Project represents a C/C++ project that can be built using the Clay build system.
// It can be a library or an executable.
type Project struct {
	Toolchain    toolchain.Environment // Build environment for this project
	IsExecutable bool                  // Is this project an executable (true) or a library (false)

	Name         string
	Config       *Config      // Build configuration
	Defines      *ValueSet    // Compiler defines (macros) for the library
	IncludeDirs  *IncludeMap  // Include paths for the library (system)
	SourceFiles  []SourceFile // C/C++ Source files for the library
	Dependencies []*Project   // Libraries that this project depends on
	Frameworks   []string     // Frameworks to link against (for macOS)
}

func NewExecutableProject(name string, config *Config) *Project {
	return &Project{
		Name:         name,
		Config:       config,
		Toolchain:    nil, // Will be set later
		IsExecutable: true,
		Defines:      NewValueSet(),
		IncludeDirs:  NewIncludeMap(),
		SourceFiles:  []SourceFile{},
		Dependencies: []*Project{},
	}
}

func NewLibraryProject(name string, config *Config) *Project {
	return &Project{
		Name:         name,
		Config:       config,
		Toolchain:    nil, // Will be set later
		IsExecutable: false,
		Defines:      NewValueSet(),
		IncludeDirs:  NewIncludeMap(),
		SourceFiles:  []SourceFile{},
		Dependencies: []*Project{},
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
		p.Toolchain, err = toolchain.NewWindowsMsdev()
	} else if targetOS == "mac" || targetOS == "macos" || targetOS == "darwin" {
		p.Toolchain, err = toolchain.NewDarwinClang(runtime.GOARCH, p.Frameworks)
	} else {
		err = fmt.Errorf("error, %s as a build target on %s is not supported", targetOS, runtime.GOOS)
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
	staticArchiver := p.Toolchain.NewArchiver(toolchain.ArchiverTypeStatic, CopyConfig(buildConfig))
	//dynamicArchiver := p.Toolchain.NewArchiver(toolchain.ArchiverTypeDynamic, CopyConfig(p.Config))

	compiler.SetupArgs(p.Defines.Values, p.IncludeDirs.Values)

	projectBuildPath := p.GetBuildPath(buildPath)
	projectDepFileTrackr := p.Toolchain.NewDependencyTracker(projectBuildPath)

	prjObjFilepaths := []string{}
	for _, src := range p.SourceFiles {
		srcObjRelPath := filepath.Join(projectBuildPath, src.SrcRelPath+".o")
		prjObjFilepaths = append(prjObjFilepaths, srcObjRelPath)
		if !projectDepFileTrackr.QueryItem(srcObjRelPath) {
			outOfDate += 1
		}
	}

	buildStartTime := time.Now()
	if outOfDate > 0 {
		log.Printf("Building project: %s, config: %s\n", p.Name, p.Config.String())
		buildStartTime = time.Now()

		for _, src := range p.SourceFiles {
			srcObjRelPath := filepath.Join(projectBuildPath, src.SrcRelPath+".o")
			if !projectDepFileTrackr.QueryItem(srcObjRelPath) {
				MakeDir(filepath.Dir(srcObjRelPath))
				if err := compiler.Compile(src.SrcAbsPath, srcObjRelPath); err != nil {
					return outOfDate, err
				}
				srcDepRelPath := filepath.Join(projectBuildPath, src.SrcRelPath+".d")
				if mainItem, depItems, err := dpenc.ParseDotDependencyFile(srcDepRelPath); err == nil {
					projectDepFileTrackr.AddItem(mainItem, depItems)
				} else {
					return outOfDate, err
				}
			} else {
				projectDepFileTrackr.CopyItem(srcObjRelPath)
			}
		}
	}

	if p.IsExecutable {
		generateMapFile := true

		linker := p.Toolchain.NewLinker(CopyConfig(buildConfig))
		linker.SetupArgs(generateMapFile, []string{}, []string{})

		archivesToLink := []string{}
		executableOutputFilepath := p.GetOutputFilepath(buildPath, linker.Filename(p.Name))

		if outOfDate > 0 || !projectDepFileTrackr.QueryItem(executableOutputFilepath) {
			if outOfDate == 0 {
				log.Printf("Linking project: %s, config: %s\n", p.Name, p.Config.String())
				buildStartTime = time.Now()
				outOfDate += 1
			}

			// Project object files
			for _, obj := range prjObjFilepaths {
				archivesToLink = append(archivesToLink, obj)
			}
			// Project dependency archives
			for _, dep := range p.Dependencies {
				libAbsFilepath := dep.GetOutputFilepath(buildPath, staticArchiver.Filename(dep.Name))
				archivesToLink = append(archivesToLink, libAbsFilepath)
			}
			// Link them all together into a single executable
			if err := linker.Link(archivesToLink, executableOutputFilepath); err != nil {
				return outOfDate, err
			}
			projectDepFileTrackr.AddItem(executableOutputFilepath, archivesToLink)

		} else {
			projectDepFileTrackr.CopyItem(executableOutputFilepath)
		}
		_, err = projectDepFileTrackr.Save()
	} else {
		archiveOutputFilepath := p.GetOutputFilepath(buildPath, staticArchiver.Filename(p.Name))
		if outOfDate > 0 || !projectDepFileTrackr.QueryItem(archiveOutputFilepath) {
			if outOfDate == 0 {
				log.Printf("Archiving project: %s, config: %s\n", p.Name, p.Config.String())
				buildStartTime = time.Now()
				outOfDate += 1
			}

			// If this is a library, we don't link it, but we can create a static archive
			staticArchiver.SetupArgs(toolchain.Vars{})
			if err := staticArchiver.Archive(prjObjFilepaths, archiveOutputFilepath); err != nil {
				return outOfDate, err
			}
			projectDepFileTrackr.AddItem(archiveOutputFilepath, prjObjFilepaths)
			_, err = projectDepFileTrackr.Save()
			if err != nil {
				return outOfDate, err
			}
		} else {
			projectDepFileTrackr.CopyItem(archiveOutputFilepath)
			_, err = projectDepFileTrackr.Save()
		}
	}

	if outOfDate > 0 {
		log.Printf("Building done ... (duration %s)\n", time.Since(buildStartTime).Round(time.Second))
	}

	return outOfDate, err
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

	utils.AddFilesFrom(srcPath, handleDir, handleFile)
}
