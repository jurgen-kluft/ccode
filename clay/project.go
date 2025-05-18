package clay

import (
	"path/filepath"
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
	Name       string
	Version    string
	BuildPath  string            // Path to the build directory
	BuildEnv   *BuildEnvironment // Build environment for this project
	Executable *Executable       // Executable that this project builds (if any)
}

func NewProject(name string, version string, buildPath string) *Project {
	buildPath = filepath.Join(buildPath, name)
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

	p.BuildEnv = be

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

func (p *Project) Build() error {
	return p.BuildEnv.BuildFunc(p.BuildEnv, p.Executable, p.BuildPath)
}

func (p *Project) Flash() error {
	return p.BuildEnv.FlashFunc(p.BuildEnv, p.Executable, p.BuildPath)
}
