package clay

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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

type BuildEnvironmentEsp32 BuildEnvironment

func NewBuildEnvironmentEsp32(buildPath string) *BuildEnvironment {

	EspSdk := "/Users/obnosis5/sdk/arduino/esp32"
	if sdk := os.Getenv("ESP_SDK"); sdk != "" {
		EspSdk = sdk
	}

	be := (*BuildEnvironmentEsp32)(NewBuildEnvironment("esp32", "v1.0.0", EspSdk))

	{ // C Compiler specific
		cc := NewCompiler(EspSdk + "/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-gcc")

		cc.Defines.Add("F_CPU=240000000L")
		cc.Defines.Add("ARDUINO=10605")
		cc.Defines.Add("ARDUINO_ESP32_DEV")
		cc.Defines.Add("ARDUINO_ARCH_ESP32")
		cc.Defines.Add(`ARDUINO_BOARD="ESP32_DEV"`)
		cc.Defines.Add(`ARDUINO_VARIANT="esp32"`)
		cc.Defines.Add("ARDUINO_PARTITION_default")
		cc.Defines.Add(`ARDUINO_HOST_OS="Darwin"`)
		cc.Defines.Add(`ARDUINO_FQBN="generic"`)
		cc.Defines.Add("ESP32=ESP32")
		cc.Defines.Add("CORE_DEBUG_LEVEL=0")
		cc.Defines.Add("ARDUINO_USB_CDC_ON_BOOT=0")

		cc.ResponseFileFlags = filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/flags/c_flags")
		cc.ResponseFileDefines = filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/flags/defines")
		cc.ResponseFileIncludes = filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/flags/includes")

		cc.IncludePaths = NewIncludeMap()
		cc.IncludePaths.Add(filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/include/"), true)
		cc.IncludePaths.Add(filepath.Join(EspSdk, "cores/esp32"), false)
		cc.IncludePaths.Add(filepath.Join(EspSdk, "variants/esp32"), false)
		cc.IncludePaths.Add(filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/dio_qspi/include"), false)

		cc.Switches.Add("-w")  // Suppress all warnings
		cc.Switches.Add("-Os") // Optimize for size

		cc.WarningSwitches.Add("-Werror=return-type")

		be.CCompiler = cc
	}

	{ // C++ Compiler specific

		cxc := NewCompiler(filepath.Join(EspSdk, "tools/xtensa-esp-elf/bin/xtensa-esp32-elf-g++"))

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

		cxc.ResponseFileFlags = filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/flags/cpp_flags")
		cxc.ResponseFileDefines = filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/flags/defines")
		cxc.ResponseFileIncludes = filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/flags/includes")

		cxc.IncludePaths = NewIncludeMap()
		cxc.IncludePaths.Add(filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/include/"), true)
		cxc.IncludePaths.Add(filepath.Join(EspSdk, "cores/esp32"), false)
		cxc.IncludePaths.Add(filepath.Join(EspSdk, "variants/esp32"), false)
		cxc.IncludePaths.Add(filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/dio_qspi/include"), false)

		cxc.Switches.Add("-w")  // Suppress all warnings
		cxc.Switches.Add("-Os") // Optimize for size

		cxc.WarningSwitches.Add("-Werror=return-type")

		be.CppCompiler = cxc
	}

	// Compiler specific
	be.CompileFunc = be.Compile
	be.CCompiler.BuildArgs = BuildCompilerArgs
	be.CppCompiler.BuildArgs = BuildCompilerArgs

	// Archiver specific

	be.Archiver = NewArchiver(filepath.Join(EspSdk, "tools/xtensa-esp-elf/bin/xtensa-esp32-elf-gcc-ar"))
	be.Archiver.BuildArgs = be.BuildArchiverArgs
	be.ArchiveFunc = be.Archive

	// Linker specific

	be.Linker = NewLinker(EspSdk + "/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-g++")
	be.Linker.OutputMapFile = true

	be.Linker.LibraryPaths.Add(filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/lib"))
	be.Linker.LibraryPaths.Add(filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/ld"))
	be.Linker.LibraryPaths.Add(filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/dio_qspi"))

	be.Linker.ResponseFileLdFlags = filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/flags/ld_flags")
	be.Linker.ResponseFileLdScripts = filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/flags/ld_scripts")
	be.Linker.ResponseFileLdLibs = filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/flags/ld_libs")

	be.Linker.BuildArgs = be.BuildLinkerArgs
	be.LinkFunc = be.Link

	// Flashing specific
	be.EspTool = NewEspTool(filepath.Join(EspSdk, "tools/esptool/esptool"))
	be.EspTool.Chip = "esp32"          // --chip esp32
	be.EspTool.Port = ""               // --port /dev/ttyUSB0
	be.EspTool.Baud = "921600"         // --baud 921600
	be.EspTool.FlashMode = "dio"       // --flash_mode dio
	be.EspTool.FlashFrequency = "40m"  // --flash_freq 40m
	be.EspTool.FlashSize = "4MB"       // --flash_size 4MB
	be.EspTool.ElfShareOffset = "0xb0" // --flash_offset 0xb0
	be.EspTool.PartitionCsvFile = filepath.Join(EspSdk, "tools/partitions/default.csv")

	// Image Generation
	be.ImageGenerator = NewImageGenerator("python3", filepath.Join(EspSdk, "tools/gen_esp32part.py"), be.EspTool)

	// Partitions generation specific
	be.ImageGenerator.PartitionsBinToolOutputFile = ""
	be.ImageGenerator.PartitionsBinToolArgs = be.GeneratePartitionsBinArgs
	be.ImageGenerator.ImageBinToolArgs = be.GenerateImageBinArgs
	be.ImageGenerator.ImageBinTool = be.EspTool

	be.ImageStatsTool = NewImageStatsTool(filepath.Join(EspSdk, "/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-size"))
	be.ImageStatsTool.ToolArgs = be.ImageStatsToolArgs
	be.ImageStatsTool.ParseStats = be.ImageStatsParser

	be.PrebuildFunc = be.PreBuild

	be.BuildFunc = be.Build
	be.BuildLibFunc = be.BuildLib

	be.GenerateImageFunc = be.GenerateImage
	be.GenerateStatsFunc = be.GenerateElfSizeStats

	be.BootLoaderCompiler = NewBootLoaderCompiler(be.EspTool)
	be.BootLoaderCompiler.Variables.Add("BootLoaderElfPath", filepath.Join(EspSdk, "tools/esp32-arduino-libs/esp32/bin/bootloader_dio_40m.elf"))
	be.BootLoaderCompiler.Args = be.GenerateBootLoaderArgs
	be.BootLoaderCompiler.Execute = be.CreateBootLoader

	be.FlashTool = NewFlashTool(be.EspTool)
	be.FlashTool.Args = be.FlashToolArgs
	be.FlashTool.Flash = be.FlashToolFlash
	be.FlashTool.Variables.Add("BootApp0BinFile", filepath.Join(EspSdk, "tools/partitions/boot_app0.bin"))
	be.FlashFunc = be.Flash

	return (*BuildEnvironment)(be)
}

// BuildCompilerArgs returns the C and C++ compiler arguments for the requested source file under the provided library.
func BuildCompilerArgs(cl *Compiler, lib *Library, srcFile *SourceFile, outputPath string) []string {
	// -MMD, this is to generate a dependency file
	// -c, this is to compile the source file to an object file
	args := make([]string, 0)
	args = append(args, "-c")
	args = append(args, "-MMD")

	if len(cl.ResponseFileFlags) > 0 {
		args = append(args, "@"+cl.ResponseFileFlags+"")
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

	// Compiler (user/library) defines
	for _, define := range lib.Defines.Values {
		args = append(args, "-D")
		args = append(args, define)
	}

	if len(cl.ResponseFileDefines) > 0 {
		args = append(args, "@"+cl.ResponseFileDefines+"")
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

	// Compiler (user/library) include paths
	for _, include := range lib.IncludeDirs.Values {
		args = append(args, "-I")
		args = append(args, include.IncludePath)
	}

	if len(cl.ResponseFileIncludes) > 0 {
		args = append(args, "@"+cl.ResponseFileIncludes+"")
	}

	// The source file and the output object file
	args = append(args, srcFile.SrcAbsPath)
	args = append(args, "-o")
	args = append(args, srcFile.ObjRelPath)

	return args
}

func (*BuildEnvironmentEsp32) PreBuild(be *BuildEnvironment, buildPath string) error {

	// Make sure our buildPath exists
	MakeDir(buildPath)

	// Copy sdkconfig in the build path from $(SDKROOT)/tools/esp32-arduino-libs/esp32/sdkconfig
	sdkconfigFilepath := filepath.Join(buildPath, "sdkconfig")
	if !FileExists(sdkconfigFilepath) {
		CopyFiles(filepath.Join(be.SdkRoot, "tools/esp32-arduino-libs/esp32/sdkconfig"), sdkconfigFilepath)
	}

	return nil
}

func (*BuildEnvironmentEsp32) Build(be *BuildEnvironment, exe *Executable, buildPath string) error {
	// Generic build flow:
	// - pre-build step
	// - build all libraries
	// - build the executable
	// - generate the image
	// - generate the ELF size stats

	if err := be.PrebuildFunc(be, buildPath); err != nil {
		return err
	}

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
		fmt.Println("Memory summary:")
		fmt.Printf("    RAM:   %d bytes\n", stats.RAMSize)
		fmt.Printf("    Flash: %d bytes\n", stats.FlashSize)
	}

	return nil
}

func (*BuildEnvironmentEsp32) BuildLib(be *BuildEnvironment, lib *Library, buildPath string) error {

	// Compile all source files to object files
	lib.PrepareOutput(buildPath)

	for _, src := range lib.SourceFiles {

		libBuildPath := filepath.Join(buildPath, lib.BuildSubDir, filepath.Dir(src.SrcRelPath))
		MakeDir(libBuildPath)

		if err := be.CompileFunc(be, lib, src, libBuildPath); err != nil {
			return fmt.Errorf("failed to compile source file %s: %w", src.SrcAbsPath, err)
		}
	}

	// Archive all object files into a .lib or .a
	return be.ArchiveFunc(be, lib, buildPath)
}

func (*BuildEnvironmentEsp32) GenerateBootLoaderArgs(bl *BootLoaderCompiler, exe *Executable, buildPath string) []string {
	// Generate a bootloader image if 'NAME.bootloader.bin' not found in the buildPath:

	args := make([]string, 0)
	args = append(args, "--chip")
	args = append(args, bl.EspTool.Chip)
	args = append(args, "elf2image")
	args = append(args, "--flash_mode")
	args = append(args, bl.EspTool.FlashMode)
	args = append(args, "--flash_freq")
	args = append(args, bl.EspTool.FlashFrequency)
	args = append(args, "--flash_size")
	args = append(args, bl.EspTool.FlashSize)

	bootloaderBinFile := filepath.Join(buildPath, FileChangeExtension(exe.OutputFilePath, ".bootloader.bin"))

	args = append(args, "-o")
	args = append(args, bootloaderBinFile)
	args = append(args, bl.Variables.Get("BootLoaderElfPath"))

	return args
}

func (*BuildEnvironmentEsp32) CreateBootLoader(bl *BootLoaderCompiler, exe *Executable, buildPath string) error {
	// Generate the bootloader image
	{
		imgPath := bl.EspTool.ToolPath
		args := bl.Args(bl, exe, buildPath)

		cmd := exec.Command(imgPath, args...)
		log.Printf("Generating bootloader '%s'\n", exe.Name+".bootloader.bin")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Bootloader generation failed with %s\n", err)
		}
		if len(out) > 0 {
			log.Printf("Bootloader generation output:\n%s\n", string(out))
		}
	}

	return nil
}

func (be *BuildEnvironmentEsp32) FlashToolArgs(ft *FlashTool, exe *Executable, outputPath string) []string {
	args := make([]string, 0)

	args = append(args, "--chip")
	args = append(args, ft.Tool.Chip)
	if len(ft.Tool.Port) > 0 {
		args = append(args, "--port")
		args = append(args, ft.Tool.Port)
	}
	if len(ft.Tool.Baud) > 0 {
		args = append(args, "--baud")
		args = append(args, ft.Tool.Baud)
	}
	args = append(args, "--before")
	args = append(args, "default_reset")
	args = append(args, "--after")
	args = append(args, "hard_reset")
	args = append(args, "write_flash")
	args = append(args, "-z")
	args = append(args, "--flash_mode")
	args = append(args, "keep")
	args = append(args, "--flash_freq")
	args = append(args, "keep")
	args = append(args, "--flash_size")
	args = append(args, "keep")
	args = append(args, "0x1000")
	args = append(args, filepath.Join(outputPath, FileChangeExtension(exe.OutputFilePath, ".bootloader.bin")))
	args = append(args, "0x8000")
	args = append(args, filepath.Join(outputPath, FileChangeExtension(exe.OutputFilePath, ".partitions.bin")))
	args = append(args, "0xe000")
	args = append(args, ft.Variables.Get("BootApp0BinFile"))
	args = append(args, "0x10000")
	args = append(args, filepath.Join(outputPath, FileChangeExtension(exe.OutputFilePath, ".bin")))

	return args
}

func (*BuildEnvironmentEsp32) FlashToolFlash(ft *FlashTool, exe *Executable, outputPath string) error {

	// Flash

	flashToolPath := ft.Tool.ToolPath
	flashToolArgs := ft.Args(ft, exe, outputPath)

	flashToolCmd := exec.Command(flashToolPath, flashToolArgs...)
	log.Printf("Flashing '%s'...\n", exe.Name+".bin")
	out, err := flashToolCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Flashing failed with %s\n", err)
	}
	if len(out) > 0 {
		log.Printf("Flashing output:\n%s\n", string(out))
	}

	return nil
}

func (*BuildEnvironmentEsp32) Flash(be *BuildEnvironment, exe *Executable, buildPath string) error {

	if err := be.BootLoaderCompiler.Execute(be.BootLoaderCompiler, exe, buildPath); err != nil {
		return fmt.Errorf("failed to create bootloader: %w", err)
	}

	if err := be.FlashTool.Flash(be.FlashTool, exe, buildPath); err != nil {
		return fmt.Errorf("failed to flash: %w", err)
	}

	return nil
}

func (*BuildEnvironmentEsp32) BuildLinkerArgs(l *Linker, exe *Executable, outputPath string) []string {
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

	if len(l.ResponseFileLdFlags) > 0 {
		args = append(args, "@"+l.ResponseFileLdFlags)
	}
	if len(l.ResponseFileLdScripts) > 0 {
		args = append(args, "@"+l.ResponseFileLdScripts)
	}

	args = append(args, "-Wl,--start-group")
	for _, lib := range exe.Libraries {
		args = append(args, filepath.Join(outputPath, lib.BuildSubDir, lib.OutputFilename))
	}
	if len(l.ResponseFileLdLibs) > 0 {
		args = append(args, "@"+l.ResponseFileLdLibs)
	}
	args = append(args, "-Wl,--end-group")
	args = append(args, "-Wl,-EL")
	args = append(args, "-o")
	args = append(args, filepath.Join(outputPath, exe.OutputFilePath))

	return args
}

func (*BuildEnvironmentEsp32) ImageStatsToolArgs(statsTool *ImageStatsTool, exe *Executable, buildPath string) []string {
	args := make([]string, 0, 2)
	args = append(args, "-A")
	args = append(args, filepath.Join(buildPath, exe.OutputFilePath))
	return args
}

var gPatternsRAM = []string{".iram?.text", ".iram?.vectors", ".dram?.data", ".dram?.bss"}
var gPatternsFLASH = []string{".flash.text", ".flash.rodata", ".flash.appdesc", ".flash.init_array", ".eh_frame"}

// Match does a direct string match, and for the '?' character it will match any character.
func Match(pattern string, str string) bool {
	if len(pattern) != len(str) {
		return false
	}
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == str[i] || pattern[i] == '?' {
			continue
		}
		return false
	}
	return true
}

func (*BuildEnvironmentEsp32) ImageStatsParser(s string, exe *Executable) (*ImageStats, error) {
	stats := &ImageStats{
		FlashSize: 0,
		RAMSize:   0,
	}

	scanner := bufio.NewScanner(bytes.NewBufferString(s))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		// Split the line into fields
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		section := fields[0]
		for _, pattern := range gPatternsFLASH {
			if Match(pattern, section) {
				if size, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					stats.FlashSize += size
				}
				goto found_match
			}
		}
		for _, pattern := range gPatternsRAM {
			if Match(pattern, section) {
				if size, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					stats.RAMSize += size
				}
				goto found_match
			}
		}
	found_match:
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read output: %w", err)
	}

	return stats, nil
}

func (*BuildEnvironmentEsp32) GeneratePartitionsBinArgs(img *ImageGenerator, exe *Executable, buildPath string) []string {
	img.PartitionsBinToolOutputFile = filepath.Join(buildPath, FileChangeExtension(exe.OutputFilePath, ".partitions.bin"))
	args := make([]string, 0)
	args = append(args, img.PartitionsBinToolScript)
	args = append(args, "-q")
	args = append(args, img.ImageBinTool.PartitionCsvFile)
	args = append(args, img.PartitionsBinToolOutputFile)
	return args
}

func (*BuildEnvironmentEsp32) GenerateImageBinArgs(img *ImageGenerator, exe *Executable, buildPath string) []string {
	args := make([]string, 0)
	args = append(args, "--chip")
	args = append(args, img.ImageBinTool.Chip)
	args = append(args, "elf2image")
	args = append(args, "--flash_mode")
	args = append(args, img.ImageBinTool.FlashMode)
	args = append(args, "--flash_freq")
	args = append(args, img.ImageBinTool.FlashFrequency)
	args = append(args, "--flash_size")
	args = append(args, img.ImageBinTool.FlashSize)
	args = append(args, "--elf-sha256-offset")
	args = append(args, img.ImageBinTool.ElfShareOffset)

	args = append(args, "-o")
	args = append(args, filepath.Join(buildPath, FileChangeExtension(exe.OutputFilePath, ".bin")))
	args = append(args, filepath.Join(buildPath, exe.OutputFilePath))

	return args
}

func (*BuildEnvironmentEsp32) GenerateImage(be *BuildEnvironment, exe *Executable, buildPath string) error {

	// Generate the image partitions bin file
	{
		img, _ := exec.LookPath(be.ImageGenerator.PartitionsBinToolPath)
		args := be.ImageGenerator.PartitionsBinToolArgs(be.ImageGenerator, exe, buildPath)

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

	// Generate the image bin file
	{
		imgPath := be.ImageGenerator.ImageBinTool.ToolPath
		args := be.ImageGenerator.ImageBinToolArgs(be.ImageGenerator, exe, buildPath)

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

func (*BuildEnvironmentEsp32) GenerateElfSizeStats(be *BuildEnvironment, exe *Executable, buildPath string) (*ImageStats, error) {

	statsToolPath := be.ImageStatsTool.ElfSizeToolPath
	statsToolArgs := be.ImageStatsTool.ToolArgs(be.ImageStatsTool, exe, buildPath)

	cmd := exec.Command(statsToolPath, statsToolArgs...)
	log.Printf("Generating ELF size stats for '%s'\n", exe.Name+".elf")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ELF size stats query failed with %s\n", err)
	}
	if len(out) > 0 {
		stats, err := be.ImageStatsTool.ParseStats(string(out), exe)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ELF size stats: %w", err)
		}

		return stats, nil
	}

	return nil, fmt.Errorf("failed to generate ELF size stats: %s", string(out))
}

func (*BuildEnvironmentEsp32) Compile(be *BuildEnvironment, lib *Library, srcFile *SourceFile, outputPath string) error {
	var args []string
	var cl string
	if srcFile.IsCpp {
		args = be.CppCompiler.BuildArgs(be.CppCompiler, lib, srcFile, outputPath)
		cl = be.CppCompiler.CompilerPath
		fmt.Printf("Compiling (%s) %s\n", lib.Config, srcFile.SrcRelPath)
	} else {
		args = be.CCompiler.BuildArgs(be.CCompiler, lib, srcFile, outputPath)
		cl = be.CCompiler.CompilerPath
		fmt.Printf("Compiling (%s) %s\n", lib.Config, srcFile.SrcRelPath)
	}
	cmd := exec.Command(cl, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Compile failed, output:\n%s\n", string(out))
		return fmt.Errorf("Compile failed with %s\n", err)
	}
	if len(out) > 0 {
		log.Printf("Compile output:\n%s\n", string(out))
	}
	return nil
}

func (*BuildEnvironmentEsp32) BuildArchiverArgs(ar *Archiver, lib *Library, outputPath string) []string {
	args := make([]string, 0)

	args = append(args, "cr")
	args = append(args, filepath.Join(outputPath, lib.BuildSubDir, lib.OutputFilename))
	for _, src := range lib.SourceFiles {
		args = append(args, src.ObjRelPath)
	}

	return args
}

func (*BuildEnvironmentEsp32) Archive(be *BuildEnvironment, lib *Library, outputPath string) error {
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

func (*BuildEnvironmentEsp32) Link(be *BuildEnvironment, exe *Executable, outputPath string) error {
	var args []string
	var lnk string

	args = be.Linker.BuildArgs(be.Linker, exe, outputPath)
	lnk = be.Linker.LinkerPath

	cmd := exec.Command(lnk, args...)
	log.Printf("Linking '%s'...\n", exe.Name+".elf")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Link failed, output:\n%s\n", string(out))
		return fmt.Errorf("Link failed with %s\n", err)
	}
	if len(out) > 0 {
		log.Printf("Link output:\n%s\n", string(out))
	}

	return nil
}
