package clay

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Project represents a C/C++ project that can be built using the Clay build system.

type BuildEnvironmentEsp32 BuildEnvironment

func NewBuildEnvironmentEsp32(buildPath string) *BuildEnvironment {

	ESP_ROOT := "/Users/obnosis5/sdk/arduino/esp32"
	be := (*BuildEnvironmentEsp32)(NewBuildEnvironment("esp32", "v1.0.0", ESP_ROOT))

	{ // C specific
		cc := NewCompiler(ESP_ROOT + "/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-gcc")

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

		cc.AtFlagsFile = ESP_ROOT + "/tools/esp32-arduino-libs/esp32/flags/c_flags"
		cc.AtDefinesFile = ESP_ROOT + "/tools/esp32-arduino-libs/esp32/flags/defines"
		cc.AtIncludesFile = ESP_ROOT + "/tools/esp32-arduino-libs/esp32/flags/includes"

		cc.IncludePaths = NewIncludeMap()
		cc.IncludePaths.Add(ESP_ROOT+"/tools/esp32-arduino-libs/esp32/include/", true)
		cc.IncludePaths.Add(ESP_ROOT+"/cores/esp32", false)
		cc.IncludePaths.Add(ESP_ROOT+"/variants/esp32", false)
		cc.IncludePaths.Add(ESP_ROOT+"/tools/esp32-arduino-libs/esp32/dio_qspi/include", false)

		cc.Switches = NewValueMap()
		cc.Switches.Add("-w")  // Suppress all warnings
		cc.Switches.Add("-Os") // Optimize for size

		cc.WarningSwitches = NewValueMap()
		cc.WarningSwitches.Add("-Werror=return-type")

		cc.BuildArgs = BuildCompilerArgs

		be.CCompiler = cc
	}
	{ // C++ specific

		cxc := NewCompiler(ESP_ROOT + "/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-g++")

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

		cxc.AtFlagsFile = ESP_ROOT + "/tools/esp32-arduino-libs/esp32/flags/cpp_flags"
		cxc.AtDefinesFile = ESP_ROOT + "/tools/esp32-arduino-libs/esp32/flags/defines"
		cxc.AtIncludesFile = ESP_ROOT + "/tools/esp32-arduino-libs/esp32/flags/includes"

		cxc.IncludePaths = NewIncludeMap()
		cxc.IncludePaths.Add(ESP_ROOT+"/tools/esp32-arduino-libs/esp32/include/", true)
		cxc.IncludePaths.Add(ESP_ROOT+"/cores/esp32", false)
		cxc.IncludePaths.Add(ESP_ROOT+"/variants/esp32", false)
		cxc.IncludePaths.Add(ESP_ROOT+"/tools/esp32-arduino-libs/esp32/dio_qspi/include", false)

		cxc.Switches = NewValueMap()
		cxc.Switches.Add("-w")  // Suppress all warnings
		cxc.Switches.Add("-Os") // Optimize for size

		cxc.WarningSwitches = NewValueMap()
		cxc.WarningSwitches.Add("-Werror=return-type")

		cxc.BuildArgs = BuildCompilerArgs

		be.CppCompiler = cxc
	}

	// Archiver specific

	be.Archiver = NewArchiver(ESP_ROOT + "/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-gcc-ar")
	be.Archiver.BuildArgs = func(ar *Archiver, lib *Library, outputPath string) []string {
		args := make([]string, 0)

		args = append(args, "cr")
		args = append(args, filepath.Join(outputPath, lib.BuildSubDir, lib.OutputFilename))
		for _, src := range lib.SourceFiles {
			args = append(args, src.ObjRelPath)
		}

		return args
	}

	// Linker specific

	be.Linker = NewLinker(ESP_ROOT + "/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-g++")

	be.Linker.OutputMapFile = true

	be.Linker.LibraryPaths.Add(filepath.Join(ESP_ROOT, "tools/esp32-arduino-libs/esp32/lib"))
	be.Linker.LibraryPaths.Add(filepath.Join(ESP_ROOT, "tools/esp32-arduino-libs/esp32/ld"))
	be.Linker.LibraryPaths.Add(filepath.Join(ESP_ROOT, "tools/esp32-arduino-libs/esp32/dio_qspi"))

	be.Linker.AtLdFlagsFile = filepath.Join(ESP_ROOT, "tools/esp32-arduino-libs/esp32/flags/ld_flags")
	be.Linker.AtLdScriptsFile = filepath.Join(ESP_ROOT, "tools/esp32-arduino-libs/esp32/flags/ld_scripts")
	be.Linker.AtLdLibsFile = filepath.Join(ESP_ROOT, "tools/esp32-arduino-libs/esp32/flags/ld_libs")

	be.Linker.BuildArgs = func(l *Linker, exe *Executable, outputPath string) []string {
		args := make([]string, 0)

		if l.OutputMapFile {
			mapFilePath := filepath.Join(outputPath, FileChangeExtension(exe.OutputFilePath, ".map"))
			args = append(args, "-Wl,--Map="+mapFilePath)
		}

		for _, libPath := range l.LibraryPaths.Values {
			args = append(args, "-L")
			args = append(args, libPath)
		}

		args = append(args, "-Wl,--wrap=esp_panic_handler")

		if len(l.AtLdFlagsFile) > 0 {
			args = append(args, "@"+l.AtLdFlagsFile)
		}
		if len(l.AtLdScriptsFile) > 0 {
			args = append(args, "@"+l.AtLdScriptsFile)
		}

		args = append(args, "-Wl,--start-group")
		for _, lib := range exe.Libraries {
			args = append(args, filepath.Join(outputPath, lib.BuildSubDir, lib.OutputFilename))
		}
		if len(l.AtLdLibsFile) > 0 {
			args = append(args, "@"+l.AtLdLibsFile)
		}
		args = append(args, "-Wl,--end-group")
		args = append(args, "-Wl,-EL")
		args = append(args, "-o")
		args = append(args, filepath.Join(outputPath, exe.OutputFilePath))

		return args
	}

	// Image Generation

	be.ImageGenerator = NewImageGenerator("python3", ESP_ROOT+"/tools/gen_esp32part.py", ESP_ROOT+"/tools/esptool/esptool")

	// Partitions generation specific
	be.ImageGenerator.PartitionCsvFile = ESP_ROOT + "/tools/partitions/default.csv"
	be.ImageGenerator.PartitionsBinOutputFile = ""

	// Image generation specific
	be.ImageGenerator.Chip = "esp32"          // --chip esp32
	be.ImageGenerator.FlashMode = "dio"       // --flash_mode dio
	be.ImageGenerator.FlashFrequency = "40m"  // --flash_freq 40m
	be.ImageGenerator.FlashSize = "4MB"       // --flash_size 4MB
	be.ImageGenerator.ElfShareOffset = "0xb0" // --flash_offset 0xb0

	be.ImageGenerator.GenEspPartArgs = func(img *ImageGenerator, exe *Executable, buildPath string) []string {

		img.PartitionsBinOutputFile = filepath.Join(buildPath, FileChangeExtension(exe.OutputFilePath, ".partitions.bin"))

		args := make([]string, 0)
		args = append(args, img.EspPartitionsToolScript)
		args = append(args, "-q")
		args = append(args, img.PartitionCsvFile)
		args = append(args, img.PartitionsBinOutputFile)

		return args
	}

	be.ImageGenerator.GenEspToolArgs = func(img *ImageGenerator, exe *Executable, buildPath string) []string {

		args := make([]string, 0)
		args = append(args, "--chip")
		args = append(args, img.Chip)
		args = append(args, "elf2image")
		args = append(args, "--flash_mode")
		args = append(args, img.FlashMode)
		args = append(args, "--flash_freq")
		args = append(args, img.FlashFrequency)
		args = append(args, "--flash_size")
		args = append(args, img.FlashSize)
		args = append(args, "--elf-sha256-offset")
		args = append(args, img.ElfShareOffset)

		args = append(args, "-o")
		args = append(args, filepath.Join(buildPath, FileChangeExtension(exe.OutputFilePath, ".bin")))
		args = append(args, filepath.Join(buildPath, exe.OutputFilePath))

		return args
	}

	// Functions

	be.SetupFunc = func(be *BuildEnvironment) error { return nil }
	be.BuildFunc = be.Build
	be.BuildLibFunc = be.BuildLib

	be.CompileFunc = func(be *BuildEnvironment, srcFile *SourceFile, outputPath string) error {
		var args []string
		var cl string
		if srcFile.IsCpp {
			args = be.CppCompiler.BuildArgs(be.CppCompiler, srcFile, outputPath)
			cl = be.CppCompiler.CompilerPath
			fmt.Printf("Compiling C++ file, %s\n", srcFile.SrcRelPath)
		} else {
			args = be.CCompiler.BuildArgs(be.CCompiler, srcFile, outputPath)
			cl = be.CCompiler.CompilerPath
			fmt.Printf("Compiling C file, %s\n", srcFile.SrcRelPath)
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

	// Archiver

	be.ArchiveFunc = func(be *BuildEnvironment, lib *Library, outputPath string) error {
		var args []string
		var ar string

		args = be.Archiver.BuildArgs(be.Archiver, lib, outputPath)
		ar = be.Archiver.ArchiverPath

		cmd := exec.Command(ar, args...)
        log.Printf("Archiving %s\n", lib.OutputFilename)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Archive failed with %s\n", err)
		}
		if len(out) > 0 {
			log.Printf("Archive output:\n%s\n", string(out))
		}

		return nil
	}

	// Linker

	be.LinkFunc = func(be *BuildEnvironment, exe *Executable, outputPath string) error {
		var args []string
		var lnk string

		args = be.Linker.BuildArgs(be.Linker, exe, outputPath)
		lnk = be.Linker.LinkerPath

		cmd := exec.Command(lnk, args...)
        log.Printf("Linking '%s'...\n", exe.Name+".elf")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Link failed with %s\n", err)
		}
		if len(out) > 0 {
			log.Printf("Link output:\n%s\n", string(out))
		}

		return nil
	}

	// Generate Image
	be.GenerateImageFunc = func(be *BuildEnvironment, exe *Executable, buildPath string) error {

		{
			img, _ := exec.LookPath(be.ImageGenerator.EspPartitionsToolPath)
			args := be.ImageGenerator.GenEspPartArgs(be.ImageGenerator, exe, buildPath)

			cmd := exec.Command(img, args...)
			log.Printf("Creating image partitions '%s' ...\n", exe.Name+".partitions.bin")
			out, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("Creating image partitions failed with %s\n", err)
			}
			if len(out) > 0 {
				log.Printf("Image partitions output:\n%s\n", string(out))
			}
		}
		{
			imgPath := be.ImageGenerator.EspImageToolPath
			args := be.ImageGenerator.GenEspToolArgs(be.ImageGenerator, exe, buildPath)

			cmd := exec.Command(imgPath, args...)
            log.Printf("Generating image '%s'\n", exe.Name+".bin")
			out, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("Image generation failed with %s\n", err)
			}
			if len(out) > 0 {
				log.Printf("Image generation output:\n%s\n", string(out))
			}
		}

		return nil
	}

	return (*BuildEnvironment)(be)
}

func BuildCompilerArgs(cl *Compiler, srcFile *SourceFile, outputPath string) []string {
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
	args = append(args, srcFile.ObjRelPath)

	return args
}

func (*BuildEnvironmentEsp32) PreBuild(be *BuildEnvironment) error {

	// Call the pre-build function if it exists
	if err := be.PrebuildFunc(be); err != nil {
		return fmt.Errorf("failure running prebuild: %w", err)
	}

	return nil
}

func (*BuildEnvironmentEsp32) Build(be *BuildEnvironment, exe *Executable, buildPath string) error {
	// Generic build flow:
	// - build all libraries
	// - build the executable
	// - generate the image
	// - generate the ELF size stats

	// Build all libraries
	for _, lib := range exe.Libraries {
		if err := be.BuildLibFunc(be, lib, buildPath); err != nil {
			return err
		}
	}

	// Build the executable
	if err := be.LinkFunc(be, exe, buildPath); err != nil {
		return err
	}

	// Generate the Image
	if err := be.GenerateImageFunc(be, exe, buildPath); err != nil {
		return err
	}

	// Generate the ELF size stats
	if stats, err := be.GenerateStatsFunc(be, exe, buildPath); err != nil {
		return err
	} else {
		// Print the ELF size stats
		fmt.Println("ELF Size Stats:")
		fmt.Printf("RAM: %d bytes\n", stats.RAMSize)
		fmt.Printf("Flash: %d bytes\n", stats.FlashSize)
	}

	return nil
}

func (*BuildEnvironmentEsp32) BuildLib(be *BuildEnvironment, lib *Library, buildPath string) error {
	// Compile all source files to object files
	for _, src := range lib.SourceFiles {

		libBuildPath := filepath.Join(buildPath, lib.BuildSubDir, filepath.Dir(src.SrcRelPath))
		MakeDir(libBuildPath)

		if err := be.CompileFunc(be, src, libBuildPath); err != nil {
			return fmt.Errorf("failed to compile source file %s: %w", src.SrcAbsPath, err)
		}
	}

	// Archive all object files into a .lib or .a
	return be.ArchiveFunc(be, lib, buildPath)
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
