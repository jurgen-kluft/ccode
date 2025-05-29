package clay

import (
	"os"
	"path/filepath"
	"runtime"
)

//
// This is a build environment for ESP32 boards.
// It uses the ESP32 Arduino core and the ESP32 toolchain.
//
// An example of a generic ESP32 board is the YD ESP32 board.
//
// NOTE: Currently a lot of paths and details are hardcoded,
// necessary information from the boards.txt file and with that
// generate the necessary paths and flags is planned.
//
// With that, it should not be too much effort to add support for
// any ESP32 board.
//

func NewBuildEnvironmentEsp32s3(buildPath string) *BuildEnvironment {

	EspSdk := "/Users/obnosis5/sdk/arduino/esp32"
	if sdk := os.Getenv("ESP_SDK"); sdk != "" {
		EspSdk = sdk
	}

	mcu := "esp32s3"

	ArduinoEspSdk := filepath.Join(EspSdk, "tools/esp32-arduino-libs", mcu)

	cCompilerPath := filepath.Join(EspSdk, "tools/xtensa-esp-elf/bin/xtensa-"+mcu+"-elf-gcc")
	cCompilerDefines := []string{
		"F_CPU=240000000L",
		"ARDUINO=10605",
		"ARDUINO_ESP32_DEV",
		"ARDUINO_ARCH_ESP32",
		`ARDUINO_BOARD="ESP32_DEV"`,
		`ARDUINO_VARIANT="` + mcu + `"`,
		"ARDUINO_PARTITION_default",
		"ARDUINO_HOST_OS=\"" + runtime.GOOS + "\"",
		`ARDUINO_FQBN="generic"`,
		"ESP32=ESP32",
		"CORE_DEBUG_LEVEL=0",
		"ARDUINO_USB_CDC_ON_BOOT=0",
	}

	cppCompilerPath := filepath.Join(EspSdk, "tools/xtensa-esp-elf/bin/xtensa-"+mcu+"-elf-g++")
	cppCompilerDefines := []string{
		"F_CPU=240000000L",
		"ARDUINO=10605",
		"ARDUINO_ESP32_DEV",
		"ARDUINO_ARCH_ESP32",
		"ARDUINO_BOARD=\"ESP32_DEV\"",
		`ARDUINO_VARIANT="` + mcu + `"`,
		"ARDUINO_PARTITION_default",
		"ARDUINO_HOST_OS=\"" + runtime.GOOS + "\"",
		"ARDUINO_FQBN=\"generic\"",
		"ESP32=ESP32",
		"CORE_DEBUG_LEVEL=0",
		"ARDUINO_USB_CDC_ON_BOOT=0",
	}

	archiverPath := filepath.Join(EspSdk, "tools/xtensa-esp-elf/bin/xtensa-"+mcu+"-elf-gcc-ar")
	linkerPath := filepath.Join(EspSdk, "tools/xtensa-esp-elf/bin/xtensa-"+mcu+"-elf-g++")
	espToolPath := filepath.Join(EspSdk, "tools/esptool/esptool")
	genEsp32PartPath := filepath.Join(EspSdk, "tools/gen_esp32part.py")
	elfSizeToolPath := filepath.Join(EspSdk, "tools/xtensa-esp-elf/bin/xtensa-"+mcu+"-elf-size")
	bootLoaderElfPath := filepath.Join(ArduinoEspSdk, "bin/bootloader_qio_80m.elf")
	bootApp0BinFilePath := filepath.Join(EspSdk, "tools/partitions/boot_app0.bin")

	flashBaud := "921600"                                                          // --baud 921600
	flashMode := "dio"                                                             // --flash_mode dio
	flashFrequency := "80m"                                                        //
	flashSize := "4MB"                                                             // --flash_size 4MB
	flashElfShareOffset := "0xb0"                                                  // --flash_offset 0xb0
	flashPartitionCsvFile := filepath.Join(EspSdk, "tools/partitions/default.csv") //
	flashBootLoaderBinOffset := "0x0"                                              // esp32s3=0x0, esp32=0x1000
	flashPartitionsBinOffset := "0x8000"                                           // esp32s3=0x8000, esp32=0x8000
	flashBootApp0BinOffset := "0xe000"                                             // esp32s3=0xe000, esp32=0xe000
	flashApplicationBinOffset := "0x10000"                                         // esp32s3=0x10000, esp32=0x10000

	be := (*BuildEnvironmentEsp32)(NewBuildEnvironment(mcu, "v1.0.0", EspSdk, ArduinoEspSdk))

	{ // C Compiler specific
		cc := NewCompiler(cCompilerPath)
		for _, d := range cCompilerDefines {
			cc.Defines.Add(d)
		}

		cc.ResponseFileFlags = filepath.Join(ArduinoEspSdk, "flags/c_flags")
		cc.ResponseFileDefines = filepath.Join(ArduinoEspSdk, "flags/defines")
		cc.ResponseFileIncludes = filepath.Join(ArduinoEspSdk, "flags/includes")

		cc.IncludePaths = NewIncludeMap()
		cc.IncludePaths.Add(filepath.Join(EspSdk, "cores/esp32"))
		cc.IncludePaths.Add(filepath.Join(EspSdk, "variants", mcu))
		cc.PrefixPaths.Add(filepath.Join(ArduinoEspSdk, "include/"))
		cc.IncludePaths.Add(filepath.Join(ArduinoEspSdk, "qio_qspi/include"))

		cc.Switches.Add("-w")  // Suppress all warnings
		cc.Switches.Add("-Os") // Optimize for size

		cc.WarningSwitches.Add("-Werror=return-type")

		be.CCompiler = cc
		be.CCompiler.BuildArgs = be.BuildCompilerArgs
	}

	{ // C++ Compiler specific
		cxc := NewCompiler(cppCompilerPath)
		for _, d := range cppCompilerDefines {
			cxc.Defines.Add(d)
		}

		cxc.ResponseFileFlags = filepath.Join(ArduinoEspSdk, "flags/cpp_flags")
		cxc.ResponseFileDefines = filepath.Join(ArduinoEspSdk, "flags/defines")
		cxc.ResponseFileIncludes = filepath.Join(ArduinoEspSdk, "flags/includes")

		cxc.IncludePaths = NewIncludeMap()
		cxc.IncludePaths.Add(filepath.Join(EspSdk, "cores/esp32"))
		cxc.IncludePaths.Add(filepath.Join(EspSdk, "variants", mcu))
		cxc.PrefixPaths.Add(filepath.Join(ArduinoEspSdk, "include/"))
		cxc.IncludePaths.Add(filepath.Join(ArduinoEspSdk, "qio_qspi/include"))

		cxc.Switches.Add("-w")  // Suppress all warnings
		cxc.Switches.Add("-Os") // Optimize for size

		cxc.WarningSwitches.Add("-Werror=return-type")

		be.CppCompiler = cxc
		be.CppCompiler.BuildArgs = be.BuildCompilerArgs
	}

	// Compiler specific
	be.CompileFunc = be.Compile

	// Archiver specific

	be.Archiver = NewArchiver(archiverPath)
	be.Archiver.BuildArgs = be.BuildArchiverArgs
	be.ArchiveFunc = be.Archive

	// Linker specific

	be.Linker = NewLinker(linkerPath)
	be.Linker.OutputMapFile = true

	be.Linker.LibraryPaths.Add(filepath.Join(ArduinoEspSdk, "lib"))
	be.Linker.LibraryPaths.Add(filepath.Join(ArduinoEspSdk, "ld"))
	be.Linker.LibraryPaths.Add(filepath.Join(ArduinoEspSdk, "dio_qspi"))

	be.Linker.ResponseFileLdFlags = filepath.Join(ArduinoEspSdk, "flags/ld_flags")
	be.Linker.ResponseFileLdScripts = filepath.Join(ArduinoEspSdk, "flags/ld_scripts")
	be.Linker.ResponseFileLdLibs = filepath.Join(ArduinoEspSdk, "flags/ld_libs")

	be.Linker.BuildArgs = be.BuildLinkerArgs
	be.LinkFunc = be.Link

	// Flashing specific
	be.EspTool = NewEspTool(espToolPath)
	be.EspTool.Chip = mcu                                       // --chip esp32
	be.EspTool.Port = ""                                        // --port /dev/ttyUSB0
	be.EspTool.Baud = flashBaud                                 // --baud 921600
	be.EspTool.FlashMode = flashMode                            // --flash_mode dio
	be.EspTool.FlashFrequency = flashFrequency                  // --flash_freq
	be.EspTool.FlashSize = flashSize                            // --flash_size 4MB
	be.EspTool.ElfShareOffset = flashElfShareOffset             // --flash_offset 0xb0
	be.EspTool.PartitionCsvFile = flashPartitionCsvFile         //
	be.EspTool.BootLoaderBinOffset = flashBootLoaderBinOffset   // esp32s3=0x0, esp32=0x1000
	be.EspTool.PartitionsBinOffset = flashPartitionsBinOffset   // esp32s3=0x8000, esp32=0x8000
	be.EspTool.BootApp0BinOffset = flashBootApp0BinOffset       // esp32s3=0xe000, esp32=0xe000
	be.EspTool.ApplicationBinOffset = flashApplicationBinOffset // esp32s3=0x10000, esp32=0x10000

	// Image Generation
	be.ImageGenerator = NewImageGenerator("python3", genEsp32PartPath, be.EspTool)

	// Partitions generation specific
	be.ImageGenerator.PartitionsBinToolOutputFile = ""
	be.ImageGenerator.PartitionsBinToolArgs = be.GeneratePartitionsBinArgs
	be.ImageGenerator.ImageBinToolArgs = be.GenerateImageBinArgs
	be.ImageGenerator.ImageBinTool = be.EspTool

	be.ImageStatsTool = NewImageStatsTool(elfSizeToolPath)
	be.ImageStatsTool.ToolArgs = be.ImageStatsToolArgs
	be.ImageStatsTool.ParseStats = be.ImageStatsParser

	be.PrebuildFunc = be.PreBuild

	be.BuildFunc = be.Build
	be.BuildLibFunc = be.BuildLib

	be.GenerateImageFunc = be.GenerateImage
	be.GenerateStatsFunc = be.GenerateElfSizeStats

	be.BootLoaderCompiler = NewBootLoaderCompiler(be.EspTool)
	be.BootLoaderCompiler.BootLoaderElfPath = bootLoaderElfPath
	be.BootLoaderCompiler.Args = be.GenerateBootLoaderArgs
	be.BootLoaderCompiler.Execute = be.CreateBootLoader

	be.FlashTool = NewFlashTool(be.EspTool)
	be.FlashTool.Args = be.FlashToolArgs
	be.FlashTool.Flash = be.FlashToolFlash
	be.FlashTool.BootApp0BinFile = bootApp0BinFilePath
	be.FlashFunc = be.Flash

	return (*BuildEnvironment)(be)
}
