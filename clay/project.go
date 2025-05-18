package clay

import (
	"path/filepath"
)

// TODO:
// - MINOR: Clay build CLI-APP for the user
//          flash
// - MINOR: Add Build Info as a Library to the Executable when BuildDir/buildinfo.h/.cpp exist
// - MINOR: ESP32 S3 Target
// - MAJOR: To reduce compile/link time we need Dependency Tracking (Database)
//   - source file -> object file
//     + args (excluding '<src> + -o <obj>')
//     + response files, they should be a dependency
//     + header files
//     + tool(s) used

// DONE:
// - MINOR: Clay build CLI-APP for the user
//          build, clean, build-info, list libraries, list boards,  list flash sizes
// - MINOR: Build Info (build_info.h and build_info.cpp as a Library)
// - MINOR: Archiver
// - MINOR: Linker
// - MINOR: Image Generator
// - MINOR: Elf Size Stats

// Project represents a C/C++ project that can be built using the Clay build system.
type Project struct {
	Name       string
	Version    string
	BuildPath  string      // Path to the build directory
	Executable *Executable // Executable that this project builds (if any)
}

func NewProject(name string, version string, buildPath string) *Project {
	exe := NewExecutable(name, version, buildPath)
	return &Project{
		Name:       name,
		Version:    version,
		BuildPath:  buildPath,
		Executable: exe,
	}
}

func (p *Project) GetExecutable() *Executable {
	if p.Executable == nil {
		p.Executable = NewExecutable(p.Name, p.Version, p.BuildPath)
	}
	return p.Executable
}

func (p *Project) SetBuildEnvironment(be *BuildEnvironment) error {

	sdkRoot := be.SdkRoot

	//// System Library is at ESP_ROOT+'cores/esp32/', collect
	//// all the C and Cpp source files in this directory and create a Library.
	coreLibPath := filepath.Join(sdkRoot, "cores/esp32/")

	coreCppLib := NewCppLibrary("esp32-core-cpp", "1.0.0", "esp32-core", "libesp32-core-cpp.a")
	coreCppLib.IsSystemLibrary = true

	coreCppLib.IncludeDirs.Add(coreLibPath, false)
	coreCppLib.IncludeDirs.Add(filepath.Join(sdkRoot, "tools/esp32-arduino-libs/esp32/include/"), true)
	coreCppLib.IncludeDirs.Add(filepath.Join(sdkRoot, "cores/esp32"), false)
	coreCppLib.IncludeDirs.Add(filepath.Join(sdkRoot, "variants/esp32"), false)

	// Flash Type
	coreCppLib.IncludeDirs.Add(filepath.Join(sdkRoot, "tools/esp32-arduino-libs/esp32/dio_qspi/include"), false)

	// Get all the .cpp files from the core library path
	coreCppLib.AddSourceFilesFrom(coreLibPath, OptionAddCppFiles|OptionAddCFiles|OptionAddRecursively)
	coreCppLib.PrepareOutput(p.BuildPath)

	p.Executable.Libraries = append(p.Executable.Libraries, coreCppLib)

	return nil
}

func (p *Project) AddUserLibrary(lib *Library) {
	p.Executable.Libraries = append(p.Executable.Libraries, lib)
}

func (p *Project) Build(be *BuildEnvironment) error {
	return be.BuildFunc(be, p.Executable, p.BuildPath)
}
