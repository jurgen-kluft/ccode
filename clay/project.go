package clay

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain"
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
type Project struct {
	Name            string
	GlobalBuildPath string              // Path where all projects are built (e.g. build/)
	Config          *Config             // Build configuration
	Toolchain       toolchain.Toolchain // Build environment for this project
	Executable      *Executable         // Executable that this project builds (if any)
}

func NewProject(name string, config *Config, buildPath string) *Project {
	exe := NewExecutable(name)
	return &Project{
		Name:            name,
		GlobalBuildPath: buildPath,
		Config:          config,
		Executable:      exe,
	}
}

func (p *Project) GetBuildPath() string {
	if len(p.GlobalBuildPath) == 0 {
		return filepath.Join("build", p.Name)
	}
	return filepath.Join(p.GlobalBuildPath, p.Name)
}

func (p *Project) GetExecutable() *Executable {
	if p.Executable == nil {
		p.Executable = NewExecutable(p.Name)
	}
	return p.Executable
}

func (p *Project) SetToolchain(config *Config) (err error) {

	targetOS := config.Target.OSAsString()

	if targetOS == "arduino" {
		var tc *toolchain.ToolchainArduinoEsp32
		targetMcu := config.Target.ArchAsString()
		tc, err = toolchain.NewToolchainArduinoEsp32(targetMcu, p.GlobalBuildPath, p.Name)
		p.Toolchain = tc

		// TODO This is a HACK!!, this should actually be moved into a 'rdno_sdk' ccode package

		// System Library is at ESP_ROOT+'cores/esp32/', collect
		// all the C and Cpp source files in this directory and create a Library.
		sdkRoot := tc.Vars.GetOne("esp.sdk.path")
		coreLibPath := filepath.Join(sdkRoot, "cores/esp32/")

		coreCppLib := NewLibrary("core-cpp-"+targetMcu, p.Config)

		// Get all the .cpp files from the core library path
		coreCppLib.AddSourceFilesFrom(coreLibPath, OptionAddCppFiles|OptionAddCFiles|OptionAddRecursively)

		p.Executable.Libraries = append(p.Executable.Libraries, coreCppLib)

	} else if targetOS == "windows" {
		p.Toolchain = toolchain.NewToolchainMsdev()
	} else if targetOS == "darwin" {
		p.Toolchain, err = toolchain.NewToolchainClangDarwin()
	} else {
		return fmt.Errorf("error, %s as a build target on %s is not supported", targetOS, runtime.GOOS)
	}

	return nil
}

func (p *Project) AddUserLibrary(lib *Library) {
	p.Executable.Libraries = append(p.Executable.Libraries, lib)
}

func CopyConfig(config *Config) *toolchain.Config {
	return toolchain.NewConfig(config.Config, config.Target)
}

func (p *Project) Build() error {

	cCompiler := p.Toolchain.NewCCompiler(CopyConfig(p.Config))
	cppCompiler := p.Toolchain.NewCppCompiler(CopyConfig(p.Config))
	archiver := p.Toolchain.NewArchiver(CopyConfig(p.Config))

	for _, lib := range p.Executable.Libraries {
		libBuildPath := lib.GetOutputDirpath(p.GetBuildPath())
		MakeDir(libBuildPath)

		cCompiler.SetupArgs(lib.Defines.Values, lib.IncludeDirs.Values)
		cppCompiler.SetupArgs(lib.Defines.Values, lib.IncludeDirs.Values)

		objFilepaths := []string{}
		for _, src := range lib.SourceFiles {

			srcObjRelPath := filepath.Join(libBuildPath, src.SrcRelPath+".o")
			//srcDepRelPath := filepath.Join(libBuildPath, src.SrcRelPath+".d")

			MakeDir(filepath.Dir(srcObjRelPath))

			var err error
			if strings.HasSuffix(src.SrcRelPath, ".c") {
				err = cCompiler.Compile(src.SrcAbsPath, srcObjRelPath)
			} else {
				err = cppCompiler.Compile(src.SrcAbsPath, srcObjRelPath)
			}

			if err != nil {
				return err
			}

			objFilepaths = append(objFilepaths, srcObjRelPath)
		}

		if err := archiver.Archive(objFilepaths, lib.GetOutputFilepath(p.GetBuildPath(), archiver.FileExtension())); err != nil {
			return err
		}
	}

	linker := p.Toolchain.NewLinker(CopyConfig(p.Config))

	generateMapFile := true
	linker.SetupArgs(generateMapFile, []string{}, []string{})

	archivesToLink := []string{}
	for _, lib := range p.Executable.Libraries {
		libAbsFilepath := lib.GetOutputFilepath(p.GetBuildPath(), archiver.FileExtension())
		archivesToLink = append(archivesToLink, libAbsFilepath)
	}
	if err := linker.Link(archivesToLink, p.Executable.GetOutputFilepath(p.GetBuildPath(), linker.FileExt())); err != nil {
		return err
	}
	return nil
}

func (p *Project) Flash() error {

	burner := p.Toolchain.NewBurner(CopyConfig(p.Config))

	burner.SetupBuildArgs()
	if err := burner.Build(); err != nil {
		return err
	}

	burner.SetupBurnArgs()
	if err := burner.Burn(); err != nil {
		return err
	}

	return nil
}
