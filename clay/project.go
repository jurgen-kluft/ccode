package clay

import "fmt"

// TODO:
// - MINOR: Archiver
// - MINOR: Linker
// - MINOR: Image Generator
// - MINOR: Elf Size Stats
// - MINOR: ESP32 S3 Target
// - MAJOR: To reduce compile/link time we need Dependency Tracking (Database)
//   - source file -> object file
//     + args (excluding '<src> + -o <obj>')
//     + response files, they should be a dependency
//     + header files
//     + tool(s) used

// Project represents a C/C++ project that can be built using the Clay build system.
type Project struct {
	Name       string
	Version    string
	BuildPath  string      // Path to the build directory
	Libraries  []*Library  // Libraries that this project depends on
	Executable *Executable // Executable that this project builds (if any)
}

func NewProject(name string, version string, buildPath string) *Project {
	exe := NewExecutable(name, version, buildPath)
	return &Project{
		Name:       name,
		Version:    version,
		BuildPath:  buildPath,
		Libraries:  make([]*Library, 0),
		Executable: exe,
	}
}

func (p *Project) InitCore(be *BuildEnvironment) error {

	sdkRoot := be.SdkRoot

	// System Library is at ESP_ROOT+'/cores/esp32/', collect
	// all the C and Cpp source files in this directory and create a Library.
	coreLibPath := sdkRoot + "/cores/esp32/"

	coreCLib := NewCLibrary("esp32-core-c", "1.0.0", "esp32-core", "libesp32-core-c.a")
	coreCLib.IsSystemLibrary = true

	coreCLib.IncludeDirs.Add(coreLibPath, false)
	coreCLib.IncludeDirs.Add(sdkRoot+"/tools/esp32-arduino-libs/esp32/include/", true)
	coreCLib.IncludeDirs.Add(sdkRoot+"/tools/esp32-arduino-libs/esp32/dio_qspi/include", false)
	coreCLib.IncludeDirs.Add(sdkRoot+"/cores/esp32", false)
	coreCLib.IncludeDirs.Add(sdkRoot+"/variants/esp32", false)

	// Get all the .c files from the core library path
	coreCLib.AddSourceFilesFrom(coreLibPath, OptionAddCFiles)

	coreCppLib := NewCppLibrary("esp32-core-cpp", "1.0.0", "esp32-core", "libesp32-core-cpp.a")
	coreCppLib.IsSystemLibrary = true

	coreCppLib.IncludeDirs.Add(coreLibPath, false)
	coreCppLib.IncludeDirs.Add(sdkRoot+"/tools/esp32-arduino-libs/esp32/include/", true)
	coreCppLib.IncludeDirs.Add(sdkRoot+"/tools/esp32-arduino-libs/esp32/dio_qspi/include", false)
	coreCppLib.IncludeDirs.Add(sdkRoot+"/cores/esp32", false)
	coreCppLib.IncludeDirs.Add(sdkRoot+"/variants/esp32", false)

	// Get all the .cpp files from the core library path
	coreCppLib.AddSourceFilesFrom(coreLibPath, OptionAddCppFiles)

	p.Libraries = append(p.Libraries, coreCLib)
	p.Libraries = append(p.Libraries, coreCppLib)

	p.Executable = NewExecutable(p.Name, p.Version, p.BuildPath)

	return nil
}

func (p *Project) AddUserLibrary(lib *Library) {
	p.Libraries = append(p.Libraries, lib)
}

func (p *Project) Build(be *BuildEnvironment) error {

	// Generic build flow:
	// - build all libraries
	// - build the executable
	// - generate the image
	// - generate the ELF size stats

	// Build all libraries
	for _, lib := range p.Libraries {
		if err := be.BuildLibFunc(be, lib, p.BuildPath); err != nil {
			return err
		}
	}

	// Build the executable
	if p.Executable != nil {
		if err := be.LinkFunc(be, *p.Executable, p.BuildPath); err != nil {
			return err
		}
	}

	// Generate the Image
	if err := be.GenerateImageFunc(be, *p.Executable, p.BuildPath); err != nil {
		return err
	}

	// Generate the ELF size stats
	if stats, err := be.GenerateStatsFunc(be, *p.Executable, p.BuildPath); err != nil {
		return err
	} else {
		// Print the ELF size stats
		fmt.Println("ELF Size Stats:")
		fmt.Printf("RAM: %d bytes\n", stats.RAMSize)
		fmt.Printf("Flash: %d bytes\n", stats.FlashSize)
	}

	return nil
}
