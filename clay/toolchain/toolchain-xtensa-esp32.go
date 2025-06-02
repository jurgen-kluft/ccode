package toolchain

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain/dpenc"
	utils "github.com/jurgen-kluft/ccode/utils"
)

type ArduinoEsp32 struct {
	Vars        *Vars
	ProjectName string // The name of the project, used for output files
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Compiler

type ToolchainArduinoEsp32Compiler struct {
	toolChain       *ArduinoEsp32
	config          *Config // Configuration for the compiler, e.g., debug or release
	cCompilerPath   string
	cppCompilerPath string
	cCompilerArgs   []string
	cppCompilerArgs []string
}

func (t *ArduinoEsp32) NewCompiler(config *Config) Compiler {
	return &ToolchainArduinoEsp32Compiler{
		toolChain:       t,
		config:          config,
		cCompilerPath:   t.Vars.GetOne("c.compiler"),
		cppCompilerPath: t.Vars.GetOne("cpp.compiler"),
		cCompilerArgs:   make([]string, 0, 64),
		cppCompilerArgs: make([]string, 0, 64),
	}
}

func (cl *ToolchainArduinoEsp32Compiler) SetupArgs(defines []string, includes []string) {
	// C Compiler Arguments
	{
		cl.cCompilerArgs = cl.cCompilerArgs[:0] // Reset the arguments

		cl.cCompilerArgs = append(cl.cCompilerArgs, "-c")
		cl.cCompilerArgs = append(cl.cCompilerArgs, "-MMD")

		responseFileFlags := cl.toolChain.Vars.GetOne("c.compiler.response.flags")
		if len(responseFileFlags) > 0 {
			cl.cCompilerArgs = append(cl.cCompilerArgs, "@"+responseFileFlags)
		}

		switches := cl.toolChain.Vars.GetAll("c.compiler.switches")
		cl.cCompilerArgs = append(cl.cCompilerArgs, switches...)

		warningSwitches := cl.toolChain.Vars.GetAll("c.compiler.warning.switches")
		cl.cCompilerArgs = append(cl.cCompilerArgs, warningSwitches...)

		// Compiler system defines (debug / release ?)
		for _, d := range cl.toolChain.Vars.GetAll("c.compiler.defines") {
			cl.cCompilerArgs = append(cl.cCompilerArgs, "-D")
			cl.cCompilerArgs = append(cl.cCompilerArgs, d)
		}

		// Compiler user defines
		for _, define := range defines {
			cl.cCompilerArgs = append(cl.cCompilerArgs, "-D")
			cl.cCompilerArgs = append(cl.cCompilerArgs, define)
		}

		responseFileDefines := cl.toolChain.Vars.GetOne("c.compiler.response.defines")
		if len(responseFileDefines) > 0 {
			cl.cCompilerArgs = append(cl.cCompilerArgs, "@"+responseFileDefines)
		}

		// Compiler prefix include paths
		compilerPrefixInclude := cl.toolChain.Vars.GetOne("c.compiler.system.prefix.include")
		if len(compilerPrefixInclude) > 0 {
			cl.cCompilerArgs = append(cl.cCompilerArgs, "-iprefix")
			// Make sure the path ends with a /
			if !strings.HasSuffix(compilerPrefixInclude, "/") {
				cl.cCompilerArgs = append(cl.cCompilerArgs, compilerPrefixInclude+"/")
			} else {
				cl.cCompilerArgs = append(cl.cCompilerArgs, compilerPrefixInclude)
			}

			responseFileIncludes := cl.toolChain.Vars.GetOne("c.compiler.response.includes")
			if len(responseFileIncludes) > 0 {
				cl.cCompilerArgs = append(cl.cCompilerArgs, "@"+responseFileIncludes)
			}
		}

		// Compiler system include paths
		systemIncludes := cl.toolChain.Vars.GetAll("c.compiler.system.includes")
		for _, include := range systemIncludes {
			cl.cCompilerArgs = append(cl.cCompilerArgs, "-I")
			cl.cCompilerArgs = append(cl.cCompilerArgs, include)
		}

		// User include paths
		for _, include := range includes {
			cl.cCompilerArgs = append(cl.cCompilerArgs, "-I")
			cl.cCompilerArgs = append(cl.cCompilerArgs, include)
		}
	}

	// C++ Compiler Arguments
	{
		cl.cppCompilerArgs = cl.cppCompilerArgs[:0] // Reset the arguments

		cl.cppCompilerArgs = append(cl.cppCompilerArgs, "-c")
		cl.cppCompilerArgs = append(cl.cppCompilerArgs, "-MMD")

		cppResponseFileFlags := cl.toolChain.Vars.GetOne("cpp.compiler.response.flags")
		if len(cppResponseFileFlags) > 0 {
			cl.cppCompilerArgs = append(cl.cppCompilerArgs, "@"+cppResponseFileFlags)
		}

		cppSwitches := cl.toolChain.Vars.GetAll("cpp.compiler.switches")
		cl.cppCompilerArgs = append(cl.cppCompilerArgs, cppSwitches...)

		cppWarningSwitches := cl.toolChain.Vars.GetAll("cpp.compiler.warning.switches")
		cl.cppCompilerArgs = append(cl.cppCompilerArgs, cppWarningSwitches...)

		// Compiler system defines (debug / release ?)
		for _, d := range cl.toolChain.Vars.GetAll("cpp.compiler.defines") {
			cl.cppCompilerArgs = append(cl.cppCompilerArgs, "-D")
			cl.cppCompilerArgs = append(cl.cppCompilerArgs, d)
		}

		// Compiler user defines
		for _, define := range defines {
			cl.cppCompilerArgs = append(cl.cppCompilerArgs, "-D")
			cl.cppCompilerArgs = append(cl.cppCompilerArgs, define)
		}

		responseFileDefines := cl.toolChain.Vars.GetOne("cpp.compiler.response.defines")
		if len(responseFileDefines) > 0 {
			cl.cppCompilerArgs = append(cl.cppCompilerArgs, "@"+responseFileDefines)
		}

		// Compiler prefix include paths
		compilerPrefixInclude := cl.toolChain.Vars.GetOne("cpp.compiler.system.prefix.include")
		if len(compilerPrefixInclude) > 0 {
			cl.cppCompilerArgs = append(cl.cppCompilerArgs, "-iprefix")
			// Make sure the path ends with a /
			if !strings.HasSuffix(compilerPrefixInclude, "/") {
				cl.cppCompilerArgs = append(cl.cppCompilerArgs, compilerPrefixInclude+"/")
			} else {
				cl.cppCompilerArgs = append(cl.cppCompilerArgs, compilerPrefixInclude)
			}

			responseFileIncludes := cl.toolChain.Vars.GetOne("cpp.compiler.response.includes")
			if len(responseFileIncludes) > 0 {
				cl.cppCompilerArgs = append(cl.cppCompilerArgs, "@"+responseFileIncludes)
			}
		}

		// Compiler system include paths
		systemIncludes := cl.toolChain.Vars.GetAll("cpp.compiler.system.includes")
		for _, include := range systemIncludes {
			cl.cppCompilerArgs = append(cl.cppCompilerArgs, "-I")
			cl.cppCompilerArgs = append(cl.cppCompilerArgs, include)
		}

		// User include paths
		for _, include := range includes {
			cl.cppCompilerArgs = append(cl.cppCompilerArgs, "-I")
			cl.cppCompilerArgs = append(cl.cppCompilerArgs, include)
		}
	}
}

func (cl *ToolchainArduinoEsp32Compiler) Compile(sourceAbsFilepath string, objRelFilepath string) error {
	var compilerPath string
	var args []string
	if strings.HasSuffix(sourceAbsFilepath, ".c") {
		compilerPath = cl.cCompilerPath
		args = cl.cCompilerArgs
	} else {
		compilerPath = cl.cppCompilerPath
		args = cl.cppCompilerArgs
	}

	// The source file and the output object file
	// sourceAbsFilepath
	// -o
	// sourceRelFilepath + ".o"
	args = append(args, sourceAbsFilepath)
	args = append(args, "-o")
	args = append(args, objRelFilepath)

	fmt.Printf("Compiling (%s) %s\n", cl.config.Config.AsString(), filepath.Base(sourceAbsFilepath))

	cmd := exec.Command(compilerPath, args...)
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

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Archiver

type ToolchainArduinoEsp32Archiver struct {
	toolChain    *ArduinoEsp32
	config       *Config
	archiverPath string
	archiverArgs []string
}

func (t *ArduinoEsp32) NewArchiver(a ArchiverType, config *Config) Archiver {
	return &ToolchainArduinoEsp32Archiver{
		toolChain:    t,
		config:       config,
		archiverPath: t.Vars.GetOne("archiver"),
		archiverArgs: []string{},
	}
}

func (t *ToolchainArduinoEsp32Archiver) Filename(name string) string {
	// The file extension for the archive on ESP32 is typically ".a"
	return name + ".a"
}

func (a *ToolchainArduinoEsp32Archiver) SetupArgs(userVars Vars) {
	a.archiverArgs = make([]string, 0, 64)
}

func (a *ToolchainArduinoEsp32Archiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {

	a.archiverArgs = a.archiverArgs[:0] // Reset the arguments
	a.archiverArgs = append(a.archiverArgs, "cr")

	// {output-archive-filepath}
	a.archiverArgs = append(a.archiverArgs, outputArchiveFilepath)

	// {input-object-filepaths}
	a.archiverArgs = a.archiverArgs[0:2]
	for _, objFile := range inputObjectFilepaths {
		a.archiverArgs = append(a.archiverArgs, objFile)
	}

	log.Printf("Archiving %s\n", outputArchiveFilepath)

	cmd := exec.Command(a.archiverPath, a.archiverArgs...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("Archive failed with %s\n", err)
	}
	if len(out) > 0 {
		log.Printf("Archive output:\n%s\n", string(out))
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Linker

type ToolchainArduinoEsp32Linker struct {
	toolChain  *ArduinoEsp32
	config     *Config
	linkerPath string
	linkerArgs []string
}

func (t *ArduinoEsp32) NewLinker(config *Config) Linker {
	return &ToolchainArduinoEsp32Linker{
		toolChain:  t,
		config:     config,
		linkerPath: t.Vars.GetOne("linker"),
		linkerArgs: make([]string, 0, 64),
	}
}

// FileExt returns the file extension for the linker output, which is ".elf" for ESP32.
func (l *ToolchainArduinoEsp32Linker) Filename(name string) string {
	return name + ".elf"
}

func (l *ToolchainArduinoEsp32Linker) SetupArgs(generateMapFile bool, libraryPaths []string, libraryFiles []string) {

	l.linkerArgs = l.linkerArgs[:0] // Reset the arguments

	if generateMapFile {
		l.linkerArgs = append(l.linkerArgs, "genmap")
	}

	linkerSystemLibraryPaths := l.toolChain.Vars.GetAll("linker.system.library.paths")
	for _, libPath := range linkerSystemLibraryPaths {
		l.linkerArgs = append(l.linkerArgs, "-L")
		l.linkerArgs = append(l.linkerArgs, libPath)
	}

	// User library paths
	for _, libPath := range libraryPaths {
		l.linkerArgs = append(l.linkerArgs, "-L")
		l.linkerArgs = append(l.linkerArgs, libPath)
	}

	l.linkerArgs = append(l.linkerArgs, "-Wl,--wrap=esp_panic_handler")

	linkerResponseFile := l.toolChain.Vars.GetAll("linker.response.ldflags")
	if len(linkerResponseFile) == 1 {
		l.linkerArgs = append(l.linkerArgs, "@"+linkerResponseFile[0])
	}

	linkerResponseFile = l.toolChain.Vars.GetAll("linker.response.ldscripts")
	if len(linkerResponseFile) == 1 {
		l.linkerArgs = append(l.linkerArgs, "@"+linkerResponseFile[0])
	}

	l.linkerArgs = append(l.linkerArgs, "-Wl,--start-group")
	{
		// User library files
		for _, libFile := range libraryFiles {
			l.linkerArgs = append(l.linkerArgs, libFile)
		}

		linkerResponseFile = l.toolChain.Vars.GetAll("linker.response.ldlibs")
		if len(linkerResponseFile) == 1 {
			l.linkerArgs = append(l.linkerArgs, "@"+linkerResponseFile[0])
		}
	}
}

func (l *ToolchainArduinoEsp32Linker) Link(inputArchiveAbsFilepaths []string, outputAppRelFilepathNoExt string) error {
	linker := l.linkerPath
	linkerArgs := l.linkerArgs

	for _, libFile := range inputArchiveAbsFilepaths {
		linkerArgs = append(linkerArgs, libFile)
	}
	linkerArgs = append(linkerArgs, "-Wl,--end-group")
	linkerArgs = append(linkerArgs, "-Wl,-EL")

	linkerArgs = append(linkerArgs, "-o")
	linkerArgs = append(linkerArgs, outputAppRelFilepathNoExt)

	// Do we need to fill in the arg to generate map file?
	if linkerArgs[0] == "genmap" {
		outputMapFilepath := outputAppRelFilepathNoExt + ".map"
		linkerArgs[0] = "-Wl,--Map=" + outputMapFilepath
	}

	log.Printf("Linking '%s'...\n", outputAppRelFilepathNoExt)
	cmd := exec.Command(linker, linkerArgs...)
	out, err := cmd.CombinedOutput()

	// Reset the map generation command in the arguments so that it
	// will be updated correctly on the next Link() call.
	l.linkerArgs[0] = "genmap"

	if err != nil {
		log.Printf("Link failed, output:\n%s\n", string(out))
		return fmt.Errorf("Link failed with %s\n", err)
	}
	if len(out) > 0 {
		log.Printf("Link output:\n%s\n", string(out))
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Burner

type ToolchainArduinoEsp32Burner struct {
	toolChain                       *ArduinoEsp32
	config                          *Config // Configuration for the burner, e.g., debug or release
	generateImageBinToolArgs        *Arguments
	generateImageBinToolPath        string
	generateImagePartitionsToolArgs *Arguments
	generateImagePartitionsToolPath string
	generateBootloaderToolArgs      *Arguments
	generateBootloaderToolPath      string
	flashToolArgs                   *Arguments
	flashToolPath                   string
}

func (t *ArduinoEsp32) NewBurner(config *Config) Burner {
	return &ToolchainArduinoEsp32Burner{
		toolChain:                       t,
		config:                          config,
		generateImageBinToolArgs:        NewArguments(64),
		generateImageBinToolPath:        t.Vars.GetOne("burner.generate-image-bin"),
		generateImagePartitionsToolArgs: NewArguments(64),
		generateImagePartitionsToolPath: t.Vars.GetOne("burner.generate-partitions-bin"),
		generateBootloaderToolArgs:      NewArguments(64),
		generateBootloaderToolPath:      t.Vars.GetOne("burner.generate-bootloader"),
		flashToolArgs:                   NewArguments(64),
		flashToolPath:                   t.Vars.GetOne("burner.flash"),
	}
}

func (b *ToolchainArduinoEsp32Burner) SetupBuild(buildPath string) {
	projectElfFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".elf")
	projectBinFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".bin")
	projectPartitionsBinFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".partitions.bin")
	projectBootloaderBinFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".bootloader.bin")

	b.generateImageBinToolArgs.Clear()
	b.generateImageBinToolArgs.Add("--chip", b.toolChain.Vars.GetOne("esp.mcu"))
	b.generateImageBinToolArgs.Add("elf2image")
	b.generateImageBinToolArgs.Add("--flash_mode", b.toolChain.Vars.GetOne("burner.flash.mode"))
	b.generateImageBinToolArgs.Add("--flash_freq", b.toolChain.Vars.GetOne("burner.flash.frequency"))
	b.generateImageBinToolArgs.Add("--flash_size", b.toolChain.Vars.GetOne("burner.flash.size"))
	b.generateImageBinToolArgs.Add("--elf-sha256-offset", b.toolChain.Vars.GetOne("burner.flash.elf.share.offset"))
	b.generateImageBinToolArgs.Add("-o", projectBinFilepath)
	b.generateImageBinToolArgs.Add(projectElfFilepath)

	b.generateImagePartitionsToolArgs.Clear()
	b.generateImagePartitionsToolArgs.Add(b.toolChain.Vars.GetOne("burner.generate-partitions-bin.script"))
	b.generateImagePartitionsToolArgs.Add("-q")
	b.generateImagePartitionsToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.partitions.csv.filepath"))
	b.generateImagePartitionsToolArgs.Add(projectPartitionsBinFilepath)

	b.generateBootloaderToolArgs.Clear()
	b.generateBootloaderToolArgs.Add("--chip", b.toolChain.Vars.GetOne("esp.mcu"))
	b.generateBootloaderToolArgs.Add("elf2image")
	b.generateBootloaderToolArgs.Add("--flash_mode", b.toolChain.Vars.GetOne("burner.flash.mode"))
	b.generateBootloaderToolArgs.Add("--flash_freq", b.toolChain.Vars.GetOne("burner.flash.frequency"))
	b.generateBootloaderToolArgs.Add("--flash_size", b.toolChain.Vars.GetOne("burner.flash.size"))
	b.generateBootloaderToolArgs.Add("-o", projectBootloaderBinFilepath)

	sdkBootLoaderElfPath := b.toolChain.Vars.GetOne("burner.sdk.bootloader.elf.path")
	b.generateBootloaderToolArgs.Add(sdkBootLoaderElfPath)
}

func (b *ToolchainArduinoEsp32Burner) Build() error {

	// - Generate image partitions bin file ('PROJECT.NAME.partitions.bin'):
	// - Generate image bin file ('PROJECT.NAME.bin'):
	// - Generate bootloader image ('PROJECT.NAME.bootloader.bin'):

	// Generate the image partitions bin file
	{
		img, _ := exec.LookPath(b.generateImagePartitionsToolPath)
		args := b.generateImagePartitionsToolArgs

		cmd := exec.Command(img, args.List...)
		log.Printf("Creating image partitions '%s' ...\n", b.toolChain.ProjectName+".partitions.bin")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Creating image partitions failed with %s\n", err)
		}
		if len(out) > 0 {
			log.Printf("Image partitions output:\n%s\n", string(out))
		}
		fmt.Println()
	}

	// Generate the image bin file
	{
		imgPath := b.generateImageBinToolPath
		args := b.generateImageBinToolArgs

		cmd := exec.Command(imgPath, args.List...)
		log.Printf("Generating image '%s'\n", b.toolChain.ProjectName+".bin")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Image generation failed with %s\n", err)
		}
		if len(out) > 0 {
			log.Printf("Image generation output:\n%s\n", string(out))
		}
	}

	// Generate the bootloader image
	{
		imgPath := b.generateBootloaderToolPath
		args := b.generateBootloaderToolArgs

		cmd := exec.Command(imgPath, args.List...)
		log.Printf("Generating bootloader '%s'\n", b.toolChain.ProjectName+".bootloader.bin")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Bootloader generation failed, output:\n%s\n", string(out))
			return fmt.Errorf("Bootloader generation failed with %s\n", err)
		}
		if len(out) > 0 {
			log.Printf("Bootloader generation output:\n%s\n", string(out))
		}
	}

	return nil
}

func (b *ToolchainArduinoEsp32Burner) SetupBurn(buildPath string) error {

	b.flashToolArgs.Clear()

	b.flashToolArgs.Add("--chip", b.toolChain.Vars.GetOne("esp.mcu"))
	b.flashToolArgs.Add("--port", b.toolChain.Vars.GetOne("burner.flash.port"))
	b.flashToolArgs.Add("--baud", b.toolChain.Vars.GetOne("burner.flash.baud"))
	b.flashToolArgs.Add("--before", "default_reset")
	b.flashToolArgs.Add("--after", "hard_reset")
	b.flashToolArgs.Add("write_flash", "-z")
	b.flashToolArgs.Add("--flash_mode", "keep")
	b.flashToolArgs.Add("--flash_freq", "keep")
	b.flashToolArgs.Add("--flash_size", "keep")

	projectNameFilepath := filepath.Join(buildPath, b.toolChain.ProjectName)

	bootloaderBinFilepath := projectNameFilepath + ".bootloader.bin"
	partitionsBinFilepath := projectNameFilepath + ".partitions.bin"
	bootApp0BinFilePath := b.toolChain.Vars.GetOne("burner.bootapp0.bin.filepath")
	applicationBinFilepath := projectNameFilepath + ".bin"

	if !utils.FileExists(bootloaderBinFilepath) {
		return fmt.Errorf("Bootloader bin file '%s' does not exist", bootloaderBinFilepath)
	}
	if !utils.FileExists(partitionsBinFilepath) {
		return fmt.Errorf("Partitions bin file '%s' does not exist", partitionsBinFilepath)
	}
	if !utils.FileExists(bootApp0BinFilePath) {
		return fmt.Errorf("Boot app0 bin file '%s' does not exist", bootApp0BinFilePath)
	}
	if !utils.FileExists(applicationBinFilepath) {
		return fmt.Errorf("Application bin file '%s' does not exist", applicationBinFilepath)
	}

	b.flashToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.bootloader.bin.offset"))
	b.flashToolArgs.Add(bootloaderBinFilepath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.partitions.bin.offset"))
	b.flashToolArgs.Add(partitionsBinFilepath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.bootapp0.bin.offset"))
	b.flashToolArgs.Add(bootApp0BinFilePath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.application.bin.offset"))
	b.flashToolArgs.Add(applicationBinFilepath)

	return nil
}

func (b *ToolchainArduinoEsp32Burner) Burn() error {
	flashToolPath := b.flashToolPath
	flashToolArgs := b.flashToolArgs

	// TODO Verify that the 4 files exist before proceeding?

	flashToolCmd := exec.Command(flashToolPath, flashToolArgs.List...)
	log.Printf("Flashing '%s'...\n", b.toolChain.ProjectName+".bin")
	out, err := flashToolCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Flashing failed with %s\n", err)
	}
	if len(out) > 0 {
		log.Printf("Flashing output:\n%s\n", string(out))
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Dependency Tracker
func (t *ArduinoEsp32) NewDependencyTracker(dirpath string) dpenc.FileTrackr {
	return dpenc.LoadDotdFileTrackr(dirpath)
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Toolchain for ESP32 on Arduino
func NewArduinoEsp32(espMcu string, projectName string) (*ArduinoEsp32, error) {

	var espSdkPath string
	if espSdkPath = os.Getenv("ESP_SDK"); espSdkPath == "" {
		return nil, fmt.Errorf("ESP_SDK environment variable is not set, please set it to the path of the ESP32 SDK")
	}

	type item struct {
		key   string
		value []string
	}

	vars := []item{
		{key: "project.name", value: []string{projectName}},
		{key: "esp.mcu", value: []string{espMcu}},
		{key: "esp.sdk.path", value: []string{espSdkPath}},
		{key: "esp.sdk.version", value: []string{`3.2.0`}},
		{key: "esp.arduino.sdk.path", value: []string{`{esp.sdk.path}/tools/esp32-arduino-libs/{esp.mcu}`}},
		{key: "c.compiler.generate.mapfile", value: []string{"-MMD"}},
		{key: "c.compiler.response.flags", value: []string{`{esp.arduino.sdk.path}/flags/c_flags`}},
		{key: "c.compiler.response.defines", value: []string{`{esp.arduino.sdk.path}/flags/defines`}},
		{key: "c.compiler.response.includes", value: []string{`{esp.arduino.sdk.path}/flags/includes`}},
		{key: "c.compiler.switches", value: []string{`-w`, `-Os`}},
		{key: "c.compiler.warning.switches", value: []string{`-Werror=return-type`}},
		{key: "c.compiler.defines", value: []string{
			`F_CPU=240000000L`,
			`ARDUINO=10605`,
			`ARDUINO_ESP32_DEV`,
			`ARDUINO_ARCH_ESP32`,
			`ARDUINO_BOARD="ESP32_DEV"`,
			`ARDUINO_VARIANT="{esp.mcu}"`,
			`ARDUINO_PARTITION_default`,
			`ARDUINO_HOST_OS="` + runtime.GOOS + `"`,
			`ARDUINO_FQBN="generic"`,
			`ESP32=ESP32`,
			`CORE_DEBUG_LEVEL=0`,
			`ARDUINO_USB_CDC_ON_BOOT=0`,
		}},
		{key: "c.compiler.system.prefix.include", value: []string{`{esp.arduino.sdk.path}/include`}},
		{key: "c.compiler.system.includes", value: []string{
			`{esp.sdk.path}/cores/esp32`,
			`{esp.sdk.path}/variants/{esp.mcu}`,
		}},

		{key: "cpp.compiler.generate.mapfile", value: []string{"-MMD"}},
		{key: "cpp.compiler.response.flags", value: []string{`{esp.arduino.sdk.path}/flags/cpp_flags`}},
		{key: "cpp.compiler.response.defines", value: []string{`{esp.arduino.sdk.path}/flags/defines`}},
		{key: "cpp.compiler.response.includes", value: []string{`{esp.arduino.sdk.path}/flags/includes`}},
		{key: "cpp.compiler.switches", value: []string{`-w`, `-Os`}},
		{key: "cpp.compiler.warning.switches", value: []string{`-Werror=return-type`}},
		{key: "cpp.compiler.defines", value: []string{
			`F_CPU=240000000L`,
			`ARDUINO=10605`,
			`ARDUINO_ESP32_DEV`,
			`ARDUINO_ARCH_ESP32`,
			`ARDUINO_BOARD="ESP32_DEV"`,
			`ARDUINO_VARIANT="{esp.mcu}"`,
			`ARDUINO_PARTITION_default`,
			`ARDUINO_HOST_OS="` + runtime.GOOS + `"`,
			`ARDUINO_FQBN="generic"`,
			`ESP32=ESP32`,
			`CORE_DEBUG_LEVEL=0`,
			`ARDUINO_USB_CDC_ON_BOOT=0`,
		}},
		{key: "cpp.compiler.system.prefix.include", value: []string{`{esp.arduino.sdk.path}/include`}},
		{key: "cpp.compiler.system.includes", value: []string{
			`{esp.sdk.path}/cores/esp32`,
			`{esp.sdk.path}/variants/{esp.mcu}`,
		}},

		{key: "linker.response.ldflags", value: []string{`{esp.arduino.sdk.path}/flags/ld_flags`}},
		{key: "linker.response.ldscripts", value: []string{`{esp.arduino.sdk.path}/flags/ld_scripts`}},
		{key: "linker.response.ldlibs", value: []string{`{esp.arduino.sdk.path}/flags/ld_libs`}},
		{key: "linker.system.library.paths", value: []string{
			`{esp.arduino.sdk.path}/lib`,
			`{esp.arduino.sdk.path}/ld`,
		}},

		{key: "burner.generate-image-bin.script", value: []string{`{esp.sdk}/tools/gen_esp32part.py`}},
		{key: "burner.generate-partitions-bin.script", value: []string{`{esp.sdk.path}/tools/gen_esp32part.py`}},

		{key: "burner.flash.baud", value: []string{`921600`}},
		{key: "burner.flash.mode", value: []string{`dio`}},
		{key: "burner.flash.frequency", value: []string{`40m`}},
		{key: "burner.flash.size", value: []string{`4MB`}},
		{key: "burner.flash.port", value: []string{`/dev/tty.usbmodem4101`}},
		{key: "burner.flash.elf.share.offset", value: []string{`0xb0`}},
		{key: "burner.bootapp0.bin.filepath", value: []string{`{esp.sdk.path}/tools/partitions/boot_app0.bin`}},
		{key: "burner.flash.partitions.csv.filepath", value: []string{`{esp.sdk.path}/tools/partitions/default.csv`}},
		{key: "burner.flash.bootloader.bin.offset", value: []string{`0x1000`}},
		{key: "burner.flash.partitions.bin.offset", value: []string{`0x8000`}},
		{key: "burner.flash.bootapp0.bin.offset", value: []string{`0xe000`}},
		{key: "burner.flash.application.bin.offset", value: []string{`0x10000`}},

		{key: "c.compiler", value: []string{`{esp.sdk.path}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-gcc`}},
		{key: "cpp.compiler", value: []string{`{esp.sdk.path}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-g++`}},
		{key: "archiver", value: []string{`{esp.sdk.path}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-gcc-ar`}},
		{key: "linker", value: []string{`{esp.sdk.path}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-g++`}},
		{key: "burner.generate-bootloader", value: []string{`{esp.sdk.path}/tools/esptool/esptool`}},
		{key: "burner.generate-image-bin", value: []string{`{esp.sdk.path}/tools/esptool/esptool`}},
		{key: "burner.generate-partitions-bin", value: []string{`python3`}},
		{key: "burner.generate-elf-size", value: []string{`{esp.sdk}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-size"`}},
		{key: "burner.flash", value: []string{`{esp.sdk.path}/tools/esptool/esptool`}},
	}

	t := &ArduinoEsp32{
		ProjectName: projectName,
		Vars:        NewVars(),
	}

	for _, kv := range vars {
		t.Vars.Append(kv.key, kv.value...)
	}

	if espMcu == "esp32" {
		// #----------------------------------------------------------------------------------
		//     xtensa-esp32
		t.Vars.Set("burner.sdk.bootloader.elf.path", `{esp.arduino.sdk.path}/bin/bootloader_dio_40m.elf`)
		t.Vars.Set("burner.flash.bootloader.bin.offset", `0x1000`)
		t.Vars.Append("c.compiler.system.includes", `{esp.arduino.sdk.path}/dio_qspi/include`)
		t.Vars.Append("c.compiler.system.includes", "{esp.arduino.sdk.path}/include")
		t.Vars.Append("cpp.compiler.system.includes", `{esp.arduino.sdk.path}/dio_qspi/include`)
		t.Vars.Append("cpp.compiler.system.includes", "{esp.arduino.sdk.path}/include")
		t.Vars.Append("linker.system.library.paths", "{esp.arduino.sdk.path}/dio_qspi")
	} else if espMcu == "esp32s3" {
		// #----------------------------------------------------------------------------------
		//     xtensa-esp32s3
		t.Vars.Set("burner.sdk.bootloader.elf.path", `{esp.arduino.sdk.path}/bin/bootloader_qio_80m.elf`)
		t.Vars.Set("burner.flash.bootloader.bin.offset", `0x0`)
		t.Vars.Append("c.compiler.system.includes", `{esp.arduino.sdk.path}/qio_qspi/include`)
		t.Vars.Append("c.compiler.system.includes", "{esp.arduino.sdk.path}/include")
		t.Vars.Append("cpp.compiler.system.includes", `{esp.arduino.sdk.path}/qio_qspi/include`)
		t.Vars.Append("cpp.compiler.system.includes", "{esp.arduino.sdk.path}/include")
		t.Vars.Append("linker.system.library.paths", "{esp.arduino.sdk.path}/qio_qspi")
	} else {
		return nil, fmt.Errorf("Unsupported ESP32 MCU: %s", espMcu)
	}

	ResolveVars(t.Vars)
	return t, nil
}
