package clay

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Project represents a C/C++ project that can be built using the Clay build system.

type TargetEsp32 struct {
	cp *CompilerPackage
}

func NewTargetEsp32(buildPath string) *TargetEsp32 {
	t := &TargetEsp32{}

	cp := &CompilerPackage{}
	t.cp = cp
	cp.Name = "esp32-compiler"
	cp.Version = "1.0.0"
	cp.ArchiverPath = "/Users/obnosis5/sdk/arduino/esp32/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-gcc-ar"
	cp.LinkerPath = "/Users/obnosis5/sdk/arduino/esp32/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-g++"

	// C specific
	cc := &Compiler{}

	cc.CompilerPath = "/Users/obnosis5/sdk/arduino/esp32/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-gcc"

	cc.Defines = NewValueMap()
	cc.Defines.Add("F_CPU=240000000L")
	cc.Defines.Add("ARDUINO=10605")
	cc.Defines.Add("ARDUINO_ESP32_DEV")
	cc.Defines.Add("ARDUINO_ARCH_ESP32")
	cc.Defines.Add("ARDUINO_BOARD=\"ESP32_DEV\"")
	cc.Defines.Add("ARDUINO_VARIANT=\"esp32\"")
	cc.Defines.Add("ARDUINO_PARTITION_default")
	cc.Defines.Add("ARDUINO_HOST_OS=\"Darwin\"")
	cc.Defines.Add("ARDUINO_FQBN=\"generic\"")
	cc.Defines.Add("ESP32=ESP32")
	cc.Defines.Add("CORE_DEBUG_LEVEL=0")
	cc.Defines.Add("ARDUINO_USB_CDC_ON_BOOT=0")

	cc.AtFlagsFile = "/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/c_flags"
	cc.AtDefinesFile = "/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/defines"
	cc.AtIncludesFile = "/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/includes"

	cc.IncludePaths = NewIncludeMap()
	cc.IncludePaths.Add("/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/include/", true)
	cc.IncludePaths.Add("/Users/obnosis5/sdk/arduino/esp32/cores/esp32", false)
	cc.IncludePaths.Add("/Users/obnosis5/sdk/arduino/esp32/variants/esp32", false)
	cc.IncludePaths.Add("/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/dio_qspi/include", false)

	cc.Switches = NewValueMap()
	cc.Switches.Add("-w")  // Suppress all warnings
	cc.Switches.Add("-Os") // Optimize for size

	cc.WarningSwitches = NewValueMap()
	cc.WarningSwitches.Add("-Werror=return-type")

	cc.BuildArgs = BuildCompilerArgs

	// C++ specific

	cxc := &Compiler{}
	cxc.CompilerPath = "/Users/obnosis5/sdk/arduino/esp32/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-g++"

	cxc.Defines = NewValueMap()
	cxc.Defines.Add("F_CPU=240000000L")
	cxc.Defines.Add("ARDUINO=10605")
	cxc.Defines.Add("ARDUINO_ESP32_DEV")
	cxc.Defines.Add("ARDUINO_ARCH_ESP32")
	cxc.Defines.Add("ARDUINO_BOARD=\"ESP32_DEV\"")
	cxc.Defines.Add("ARDUINO_VARIANT=\"esp32\"")
	cxc.Defines.Add("ARDUINO_PARTITION_default")
	cxc.Defines.Add("ARDUINO_HOST_OS=\"Darwin\"")
	cxc.Defines.Add("ARDUINO_FQBN=\"generic\"")
	cxc.Defines.Add("ESP32=ESP32")
	cxc.Defines.Add("CORE_DEBUG_LEVEL=0")
	cxc.Defines.Add("ARDUINO_USB_CDC_ON_BOOT=0")

	cxc.AtFlagsFile = "/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/cpp_flags"
	cxc.AtDefinesFile = "/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/defines"
	cxc.AtIncludesFile = "/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/includes"

	cxc.IncludePaths = NewIncludeMap()
	cxc.IncludePaths.Add("/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/include/", true)
	cxc.IncludePaths.Add("/Users/obnosis5/sdk/arduino/esp32/cores/esp32", false)
	cxc.IncludePaths.Add("/Users/obnosis5/sdk/arduino/esp32/variants/esp32", false)
	cxc.IncludePaths.Add("/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/dio_qspi/include", false)

	cxc.Switches = NewValueMap()
	cxc.Switches.Add("-w")  // Suppress all warnings
	cxc.Switches.Add("-Os") // Optimize for size

	cxc.WarningSwitches = NewValueMap()
	cxc.WarningSwitches.Add("-Werror=return-type")

	cxc.BuildArgs = BuildCompilerArgs

	// Compilers

	cp.CCompiler = cc
	cp.CppCompiler = cxc

	// Linker specific

	// ...

	// Functions to be implemented

	cp.PreBuild = func() error { return nil }
	cp.Setup = func() error { return nil }
	cp.Compile = func(srcFile *SourceFile, outputPath string) error {
		var args []string
		var cl string
		if srcFile.IsCpp {
			args = cp.CppCompiler.BuildArgs(cp.CppCompiler, srcFile, outputPath)
			cl = cp.CppCompiler.CompilerPath
		} else {
			args = cp.CCompiler.BuildArgs(cp.CCompiler, srcFile, outputPath)
			cl = cp.CCompiler.CompilerPath
		}
		cmd := exec.Command(cl, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Compile failed with %s\n", err)
		}
		if len(out) > 0 {
			log.Printf("Compile output:\n%s\n", string(out))
		}
		return nil
	}
	cp.PreArchive = func() error { return nil }
	cp.Archive = func(lib *Library, outputPath string) error { return nil }
	cp.PreLink = func() error { return nil }
	cp.Link = func(exe Executable) error { return nil }

	return t
}

func (t *TargetEsp32) Init() error {
	// Initialize the compiler package
	if t.cp == nil {
		return fmt.Errorf("compiler package is not initialized")
	}

	// Set up the compiler package
	if err := t.cp.Setup(); err != nil {
		return fmt.Errorf("failed to set up compiler package: %w", err)
	}

	return nil
}

func BuildCompilerArgs(cl *Compiler, srcFile *SourceFile, outputPath string) []string {
	//exeBin := cl.CompilerPath

	// -MMD, this is to generate a dependency file
	// -c, this is to compile the source file to an object file
	args := make([]string, 0)
	args = append(args, "-c")
	args = append(args, "-MMD")

	if len(cl.AtFlagsFile) > 0 {
		args = append(args, "@"+cl.AtFlagsFile+"")
	}

	for _, s := range cl.Switches.Values {
		args = append(args, s)
	}
	for _, s := range cl.WarningSwitches.Values {
		args = append(args, s)
	}
	for _, d := range cl.Defines.Values {
		args = append(args, "-D")
		args = append(args, d)
	}

	if len(cl.AtDefinesFile) > 0 {
		args = append(args, "@"+cl.AtDefinesFile+"")
	}

	// Compiler system include paths
	for _, include := range cl.IncludePaths.Values {
		if include.Prefix {
			args = append(args, "-iprefix")
			args = append(args, include.IncludePath)
		} else {
			args = append(args, "-I")
			args = append(args, include.IncludePath)
		}
	}

	if len(cl.AtIncludesFile) > 0 {
		args = append(args, "@"+cl.AtIncludesFile+"")
	}

	// The source file and the output object file
	args = append(args, srcFile.SrcAbsPath)
	args = append(args, "-o")
	args = append(args, filepath.Join(outputPath, filepath.Base(srcFile.SrcAbsPath))+".o")

	srcFile.ObjRelPath = filepath.Join(outputPath, srcFile.ObjRelPath) + ".o"
	srcFile.DepRelPath = filepath.Join(outputPath, srcFile.DepRelPath) + ".d"

	return args
}

func (t *TargetEsp32) Prebuild() error {

	// Call the pre-build function if it exists
	if err := t.cp.PreBuild(); err != nil {
		return fmt.Errorf("failed to run prebuild: %w", err)
	}

	return nil
}

func (t *TargetEsp32) Build() error {

	// System Library is at '/Users/obnosis5/sdk/arduino/esp32/cores/esp32/', collect
	// all the C and Cpp source files in this directory and create a Library.
	coreLibPath := "/Users/obnosis5/sdk/arduino/esp32/cores/esp32/"
	buildPath := "build"

	coreCLib := &Library{
		Name:              "esp32-core-c",
		Version:           "1.0.0",
		IsSystemLibrary:   true,
		IsCppLibrary:      false,
		OutputRelFilePath: "core/libesp32-core-c",
		IncludeDirs:       NewIncludeMap(),
		SourceFiles:       []*SourceFile{},
		Libraries:         []*Library{},
	}
	coreCLib.IncludeDirs.Add(coreLibPath, false)
	coreCLib.IncludeDirs.Add("/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/include/", true)
	coreCLib.IncludeDirs.Add("/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/dio_qspi/include", false)
	coreCLib.IncludeDirs.Add("/Users/obnosis5/sdk/arduino/esp32/cores/esp32", false)
	coreCLib.IncludeDirs.Add("/Users/obnosis5/sdk/arduino/esp32/variants/esp32", false)

	// Get all the .c files from the core library path
	coreCLib.AddSourceFilesFrom(coreLibPath, buildPath, false, false)

	// Compile C source files
	buildPathC := filepath.Join(buildPath, coreCLib.OutputRelFilePath)
	MakeDir(buildPathC)
	for _, src := range coreCLib.SourceFiles {
		if err := t.cp.Compile(src, buildPathC); err != nil {
			return fmt.Errorf("failed to compile C source file %s: %w", src.SrcAbsPath, err)
		}
	}

	coreCppLib := &Library{
		Name:              "esp32-core-cpp",
		Version:           "1.0.0",
		IsSystemLibrary:   true,
		IsCppLibrary:      true,
		OutputRelFilePath: "core/libesp32-core-cpp",
		Defines:           NewValueMap(),
		IncludeDirs:       NewIncludeMap(),
		SourceFiles:       []*SourceFile{},
		Libraries:         []*Library{},
	}
	coreCppLib.IncludeDirs.Add(coreLibPath, false)
	coreCppLib.IncludeDirs.Add("/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/include/", true)
	coreCppLib.IncludeDirs.Add("/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/dio_qspi/include", false)
	coreCppLib.IncludeDirs.Add("/Users/obnosis5/sdk/arduino/esp32/cores/esp32", false)
	coreCppLib.IncludeDirs.Add("/Users/obnosis5/sdk/arduino/esp32/variants/esp32", false)

	// Get all the .cpp files from the core library path
	coreCppLib.AddSourceFilesFrom(coreLibPath, buildPath, false, true)

	// Compile C++ source files
	buildPathCpp := filepath.Join(buildPath, coreCppLib.OutputRelFilePath)
	MakeDir(buildPathCpp)
	for _, src := range coreCppLib.SourceFiles {
		if err := t.cp.Compile(src, buildPathCpp); err != nil {
			return fmt.Errorf("failed to compile C++ source file %s: %w", src.SrcAbsPath, err)
		}
	}

	return nil
}

func MakeDir(path string) error {
	// Create the directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}
