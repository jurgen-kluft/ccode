package clay

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/jurgen-kluft/ccode/clay/toolchain"
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
	Toolchain    toolchain.Toolchain // Build environment for this project
	IsExecutable bool                // Is this project an executable (true) or a library (false)

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
		var tc *toolchain.ToolchainArduinoEsp32
		targetMcu := config.Target.ArchAsString()
		tc, err = toolchain.NewToolchainArduinoEsp32(targetMcu, p.Name)
		p.Toolchain = tc

		// TODO This is a HACK!!, this should actually be moved into a 'rdno_sdk' ccode package

		// System Library is at ESP_ROOT+'cores/esp32/', collect
		// all the C and Cpp source files in this directory and create a Library.
		sdkRoot := tc.Vars.GetOne("esp.sdk.path")
		coreLibPath := filepath.Join(sdkRoot, "cores/esp32/")

		coreCppLib := NewLibraryProject("core-cpp-"+targetMcu, p.Config)

		// Get all the .cpp files from the core library path
		coreCppLib.AddSourceFilesFrom(coreLibPath, OptionAddCppFiles|OptionAddCFiles|OptionAddRecursively)

		p.Dependencies = append(p.Dependencies, coreCppLib)

	} else if targetOS == "windows" {
		p.Toolchain = toolchain.NewToolchainMsdev()
	} else if targetOS == "mac" || targetOS == "macos" || targetOS == "darwin" {
		p.Toolchain, err = toolchain.NewToolchainClangDarwin(runtime.GOARCH, p.Frameworks)
	} else {
		return fmt.Errorf("error, %s as a build target on %s is not supported", targetOS, runtime.GOOS)
	}

	return nil
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

func (p *Project) Build(buildPath string) error {

	compiler := p.Toolchain.NewCompiler(CopyConfig(p.Config))
	staticArchiver := p.Toolchain.NewArchiver(toolchain.ArchiverTypeStatic, CopyConfig(p.Config))
	//dynamicArchiver := p.Toolchain.NewArchiver(toolchain.ArchiverTypeDynamic, CopyConfig(p.Config))

	//for _, dep := range p.Dependencies {
	//	depBuildPath := dep.GetBuildPath(buildPath)
	//	MakeDir(depBuildPath)
	//
	//	compiler.SetupArgs(dep.Defines.Values, dep.IncludeDirs.Values)
	//
	//	objFilepaths := []string{}
	//	for _, src := range dep.SourceFiles {
	//
	//		srcObjRelPath := filepath.Join(depBuildPath, src.SrcRelPath+".o")
	//		//srcDepRelPath := filepath.Join(libBuildPath, src.SrcRelPath+".d")
	//
	//		MakeDir(filepath.Dir(srcObjRelPath))
	//
	//		if err := compiler.Compile(src.SrcAbsPath, srcObjRelPath); err != nil {
	//			return err
	//		}
	//
	//		objFilepaths = append(objFilepaths, srcObjRelPath)
	//	}
	//
	//	// Static library ?
	//	staticArchiver.SetupArgs(toolchain.Vars{})
	//	if err := staticArchiver.Archive(objFilepaths, dep.GetOutputFilepath(buildPath, staticArchiver.Filename(dep.Name))); err != nil {
	//		return err
	//	}
	//}

	compiler.SetupArgs(p.Defines.Values, p.IncludeDirs.Values)

	MakeDir(p.GetBuildPath(buildPath))
	prjObjFilepaths := []string{}
	for _, src := range p.SourceFiles {
		srcObjRelPath := filepath.Join(p.GetBuildPath(buildPath), src.SrcRelPath+".o")
		//srcDepRelPath := filepath.Join(libBuildPath, src.SrcRelPath+".d")

		MakeDir(filepath.Dir(srcObjRelPath))
		if err := compiler.Compile(src.SrcAbsPath, srcObjRelPath); err != nil {
			return err
		}

		prjObjFilepaths = append(prjObjFilepaths, srcObjRelPath)
	}

	if p.IsExecutable {
		linker := p.Toolchain.NewLinker(CopyConfig(p.Config))

		generateMapFile := true
		linker.SetupArgs(generateMapFile, []string{}, []string{})

		archivesToLink := []string{}
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
		if err := linker.Link(archivesToLink, p.GetOutputFilepath(buildPath, linker.Filename(p.Name))); err != nil {
			return err
		}
	} else {
		// If this is a library, we don't link it, but we can create a static archive
		staticArchiver.SetupArgs(toolchain.Vars{})
		if err := staticArchiver.Archive(prjObjFilepaths, p.GetOutputFilepath(buildPath, staticArchiver.Filename(p.Name))); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) Flash(buildPath string) error {
	burner := p.Toolchain.NewBurner(CopyConfig(p.Config))

	buildPath = p.GetBuildPath(buildPath)

	burner.SetupBuildArgs(buildPath)
	if err := burner.Build(); err != nil {
		return err
	}

	burner.SetupBurnArgs(buildPath)
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
