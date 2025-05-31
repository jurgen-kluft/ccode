package toolchain

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
)

type ToolchainArduinoEsp32 struct {
	ToolchainInstance
	depTrackr   deptrackr.DepTrackr
	projectName string // Name of the project, used for output files
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Compiler

type ToolchainArduinoEsp32Compiler struct {
	toolChain       *ToolchainArduinoEsp32
	config          *Config // Configuration for the compiler, e.g., debug or release
	cCompilerPath   string
	cppCompilerPath string
	compilerArgs    []string
}

func (t *ToolchainArduinoEsp32) NewCompiler(config *Config) Compiler {
	return &ToolchainArduinoEsp32Compiler{
		toolChain:       t,
		config:          config,
		cCompilerPath:   t.Tools["c.compiler"],
		cppCompilerPath: t.Tools["cpp.compiler"],
		compilerArgs:    make([]string, 0, 64),
	}
}

func (cl *ToolchainArduinoEsp32Compiler) SetupArgs(defines []string, includes []string) {

	cl.compilerArgs = cl.compilerArgs[:0] // Reset the arguments

	cl.compilerArgs = append(cl.compilerArgs, "-c")
	cl.compilerArgs = append(cl.compilerArgs, "-MMD")

	responseFileFlags := cl.toolChain.Vars.GetOne("c.compiler.response.flags")
	if len(responseFileFlags) > 0 {
		cl.compilerArgs = append(cl.compilerArgs, "@"+responseFileFlags)
	}

	switches := cl.toolChain.Vars.GetAll("c.compiler.switches")
	cl.compilerArgs = append(cl.compilerArgs, switches...)

	warningSwitches := cl.toolChain.Vars.GetAll("c.compiler.warning.switches")
	cl.compilerArgs = append(cl.compilerArgs, warningSwitches...)

	// Compiler system defines (debug / release ?)
	for _, d := range cl.toolChain.Vars.GetAll("c.compiler.defines") {
		cl.compilerArgs = append(cl.compilerArgs, "-D")
		cl.compilerArgs = append(cl.compilerArgs, d)
	}

	// Compiler user defines
	for _, define := range defines {
		cl.compilerArgs = append(cl.compilerArgs, "-D")
		cl.compilerArgs = append(cl.compilerArgs, define)
	}

	responseFileDefines := cl.toolChain.Vars.GetOne("c.compiler.response.defines")
	if len(responseFileDefines) > 0 {
		cl.compilerArgs = append(cl.compilerArgs, "@"+responseFileDefines)
	}

	// Compiler prefix include paths
	compilerPrefixInclude := cl.toolChain.Vars.GetOne("c.compiler.system.prefix.include")
	if len(compilerPrefixInclude) > 0 {
		cl.compilerArgs = append(cl.compilerArgs, "-iprefix")
		// Make sure the path ends with a /
		if !strings.HasSuffix(compilerPrefixInclude, "/") {
			cl.compilerArgs = append(cl.compilerArgs, compilerPrefixInclude+"/")
		} else {
			cl.compilerArgs = append(cl.compilerArgs, compilerPrefixInclude)
		}

		responseFileIncludes := cl.toolChain.Vars.GetOne("c.compiler.response.includes")
		if len(responseFileIncludes) > 0 {
			cl.compilerArgs = append(cl.compilerArgs, "@"+responseFileIncludes)
		}
	}

	// Compiler system include paths
	systemIncludes := cl.toolChain.Vars.GetAll("c.compiler.system.includes")
	for _, include := range systemIncludes {
		cl.compilerArgs = append(cl.compilerArgs, "-I")
		cl.compilerArgs = append(cl.compilerArgs, include)
	}

	// User include paths
	for _, include := range includes {
		cl.compilerArgs = append(cl.compilerArgs, "-I")
		cl.compilerArgs = append(cl.compilerArgs, include)
	}
}

func (cl *ToolchainArduinoEsp32Compiler) Compile(sourceAbsFilepath string, objRelFilepath string) error {

	args := cl.compilerArgs

	// The source file and the output object file
	// sourceAbsFilepath
	// -o
	// sourceRelFilepath + ".o"
	args = append(args, sourceAbsFilepath)
	args = append(args, "-o")
	args = append(args, objRelFilepath)

	fmt.Printf("Compiling (%s) %s\n", cl.config.Config.AsString(), filepath.Base(sourceAbsFilepath))

	var err error
	var out []byte
	if strings.HasSuffix(sourceAbsFilepath, ".cpp") {
		cmd := exec.Command(cl.cppCompilerPath, args...)
		out, err = cmd.CombinedOutput()
	} else {
		cmd := exec.Command(cl.cCompilerPath, args...)
		out, err = cmd.CombinedOutput()
	}

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
	toolChain    *ToolchainArduinoEsp32
	config       *Config
	archiverPath string
	archiverArgs []string
}

func (t *ToolchainArduinoEsp32) NewArchiver(a ArchiverType, config *Config) Archiver {
	return &ToolchainArduinoEsp32Archiver{
		toolChain:    t,
		config:       config,
		archiverPath: t.Tools["archiver"],
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
	toolChain  *ToolchainArduinoEsp32
	config     *Config
	linkerPath string
	linkerArgs []string
}

func (t *ToolchainArduinoEsp32) NewLinker(config *Config) Linker {
	return &ToolchainArduinoEsp32Linker{
		toolChain:  t,
		config:     config,
		linkerPath: t.Tools["linker"],
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

	linkerSystemLibraryPaths := l.toolChain.Vars["linker.system.library.paths"]
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

	linkerResponseFile := l.toolChain.Vars["linker.response.ldflags"]
	if len(linkerResponseFile) == 1 {
		l.linkerArgs = append(l.linkerArgs, "@"+linkerResponseFile[0])
	}

	linkerResponseFile = l.toolChain.Vars["linker.response.ldscripts"]
	if len(linkerResponseFile) == 1 {
		l.linkerArgs = append(l.linkerArgs, "@"+linkerResponseFile[0])
	}

	l.linkerArgs = append(l.linkerArgs, "-Wl,--start-group")
	{
		// User library files
		for _, libFile := range libraryFiles {
			l.linkerArgs = append(l.linkerArgs, libFile)
		}

		linkerResponseFile = l.toolChain.Vars["linker.response.ldlibs"]
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
	toolChain                       *ToolchainArduinoEsp32
	projectName                     string
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

func (t *ToolchainArduinoEsp32) NewBurner(config *Config) Burner {
	return &ToolchainArduinoEsp32Burner{
		toolChain:                       t,
		projectName:                     t.projectName,
		config:                          config,
		generateImageBinToolArgs:        NewArguments(64),
		generateImageBinToolPath:        t.Tools["burner.generate-image-bin"],
		generateImagePartitionsToolArgs: NewArguments(64),
		generateImagePartitionsToolPath: t.Tools["burner.generate-partitions-bin"],
		generateBootloaderToolArgs:      NewArguments(64),
		generateBootloaderToolPath:      t.Tools["burner.generate-bootloader"],
		flashToolArgs:                   NewArguments(64),
		flashToolPath:                   t.Tools["burner.flash"],
	}
}

func (b *ToolchainArduinoEsp32Burner) SetupBuildArgs(buildPath string) {
	configDir := b.config.GetDirname()
	projectElfFilepath := filepath.Join(buildPath, configDir, b.projectName, b.projectName+".elf")
	projectBinFilepath := filepath.Join(buildPath, configDir, b.projectName, b.projectName+".bin")
	projectPartitionsBinFilepath := filepath.Join(buildPath, configDir, b.projectName, b.projectName+".partitions.bin")
	projectBootloaderBinFilepath := filepath.Join(buildPath, configDir, b.projectName, b.projectName+".bootloader.bin")

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
	b.generateBootloaderToolArgs.Add(projectElfFilepath)
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
		log.Printf("Creating image partitions '%s' ...\n", b.projectName+".partitions.bin")
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
		log.Printf("Generating image '%s'\n", b.projectName+".bin")
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
		log.Printf("Generating bootloader '%s'\n", b.projectName+".bootloader.bin")
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

func (b *ToolchainArduinoEsp32Burner) SetupBurnArgs(buildPath string) {

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

	configDir := b.config.GetDirname()
	projectNameFilepath := filepath.Join(buildPath, configDir, b.projectName, b.projectName)

	bootloaderBinFilepath := projectNameFilepath + ".bootloader.bin"
	partitionsBinFilepath := projectNameFilepath + ".partitions.bin"
	bootApp0BinFilePath := b.toolChain.Vars.GetOne("burner.bootapp0.bin.filepath")
	applicationBinFilepath := projectNameFilepath + ".bin"

	b.flashToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.bootloader.bin.offset"))
	b.flashToolArgs.Add(bootloaderBinFilepath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.partitions.bin.offset"))
	b.flashToolArgs.Add(partitionsBinFilepath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.bootapp0.bin.offset"))
	b.flashToolArgs.Add(bootApp0BinFilePath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.application.bin.offset"))
	b.flashToolArgs.Add(applicationBinFilepath)
}

func (b *ToolchainArduinoEsp32Burner) Burn() error {
	flashToolPath := b.flashToolPath
	flashToolArgs := b.flashToolArgs

	// TODO Verify that the 4 files exist before proceeding?

	flashToolCmd := exec.Command(flashToolPath, flashToolArgs.List...)
	log.Printf("Flashing '%s'...\n", b.projectName+".bin")
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
// Toolchain for ESP32 on Arduino
func NewToolchainArduinoEsp32(espMcu string, projectName string) (t *ToolchainArduinoEsp32, err error) {

	var espSdkPath string
	if espSdkPath = os.Getenv("ESP_SDK"); espSdkPath == "" {
		return nil, err
	}

	t = &ToolchainArduinoEsp32{
		projectName: projectName,
		ToolchainInstance: ToolchainInstance{
			Name: "xtensa-esp32",
			Vars: Vars{
				"esp.mcu":              {espMcu},
				"esp.sdk.path":         {espSdkPath},
				"esp.sdk.version":      {`3.2.0`},
				"esp.arduino.sdk.path": {`{esp.sdk.path}/tools/esp32-arduino-libs/{esp.mcu}`},

				"c.compiler.generate.mapfile":  {"-MMD"},
				"c.compiler.response.flags":    {`{esp.arduino.sdk.path}/flags/c_flags`},
				"c.compiler.response.defines":  {`{esp.arduino.sdk.path}/flags/defines`},
				"c.compiler.response.includes": {`{esp.arduino.sdk.path}/flags/includes`},
				"c.compiler.switches":          {`-w`, `-Os`},
				"c.compiler.warning.switches":  {`-Werror=return-type`},
				"c.compiler.defines": {
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
				},
				"c.compiler.system.prefix.include": {`{esp.arduino.sdk.path}/include`},
				"c.compiler.system.includes": {
					`{esp.sdk.path}/cores/esp32`,
					`{esp.sdk.path}/variants/{esp.mcu}`,
				},

				"cpp.compiler.generate.mapfile":  {"-MMD"},
				"cpp.compiler.response.flags":    {`{esp.arduino.sdk.path}/flags/cpp_flags`},
				"cpp.compiler.response.defines":  {`{esp.arduino.sdk.path}/flags/defines`},
				"cpp.compiler.response.includes": {`{esp.arduino.sdk.path}/flags/includes`},
				"cpp.compiler.switches":          {`-w`, `-Os`},
				"cpp.compiler.warning.switches":  {`-Werror=return-type`},
				"cpp.compiler.defines": {
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
				},
				"cpp.compiler.system.prefix.include": {`{esp.arduino.sdk.path}/include`},
				"cpp.compiler.system.includes": {
					`{esp.sdk.path}/cores/esp32`,
					`{esp.sdk.path}/variants/{esp.mcu}`,
				},

				"archiver": {},

				"linker":                    {},
				"linker.response.ldflags":   {`{esp.arduino.sdk.path}/flags/ld_flags`},
				"linker.response.ldscripts": {`{esp.arduino.sdk.path}/flags/ld_scripts`},
				"linker.response.ldlibs":    {`{esp.arduino.sdk.path}/flags/ld_libs`},
				"linker.system.library.paths": {
					`{esp.arduino.sdk.path}/lib`,
					`{esp.arduino.sdk.path}/ld`,
				},

				"burner.generate-image-bin.script":      {`{esp.sdk}/tools/gen_esp32part.py`},
				"burner.generate-partitions-bin.script": {`{esp.sdk.path}/tools/gen_esp32part.py`},

				"burner.flash.baud":                    {`921600`},
				"burner.flash.mode":                    {`dio`},
				"burner.flash.frequency":               {`40m`},
				"burner.flash.size":                    {`4MB`},
				"burner.flash.port":                    {`/dev/tty.usbmodem4101`},
				"burner.flash.elf.share.offset":        {`0xb0`},
				"burner.bootapp0.bin.filepath":         {`{esp.sdk.path}/tools/partitions/boot_app0.bin`},
				"burner.flash.partitions.csv.filepath": {`{esp.sdk.path}/tools/partitions/default.csv`},
				"burner.flash.bootloader.bin.offset":   {`0x1000`},
				"burner.flash.partitions.bin.offset":   {`0x8000`},
				"burner.flash.bootapp0.bin.offset":     {`0xe000`},
				"burner.flash.application.bin.offset":  {`0x10000`},
			},
			Tools: map[string]string{
				"c.compiler":                     `{esp.sdk.path}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-gcc`,
				"cpp.compiler":                   `{esp.sdk.path}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-g++`,
				"archiver":                       `{esp.sdk.path}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-gcc-ar`,
				"linker":                         `{esp.sdk.path}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-g++`,
				"burner.generate-bootloader":     `{esp.sdk.path}/tools/esptool/esptool`,
				"burner.generate-image-bin":      `{esp.sdk.path}/tools/esptool/esptool`,
				"burner.generate-partitions-bin": `python3`,
				"burner.generate-elf-size":       `{esp.sdk}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-size"`,
				"burner.flash":                   `{esp.sdk.path}/tools/esptool/esptool`,
			},
		},
	}

	if espMcu == "esp32" {
		// #----------------------------------------------------------------------------------
		//     xtensa-esp32

		t.Vars.Set("bootloader.elf.path", `{esp.arduino.sdk.path}/bin/bootloader_dio_40m.elf`)
		t.Vars.Set("flash.bootloader_bin_offset", `0x1000`)
		t.Vars.Append("c.compiler.system.includes", `{esp.arduino.sdk.path}/dio_qspi/include`)
		t.Vars.Append("c.compiler.system.includes", "{esp.arduino.sdk.path}/include")
		t.Vars.Append("cpp.compiler.system.includes", `{esp.arduino.sdk.path}/dio_qspi/include`)
		t.Vars.Append("cpp.compiler.system.includes", "{esp.arduino.sdk.path}/include")
		t.Vars.Append("linker.system.library.paths", "{esp.arduino.sdk.path}/dio_qspi")

	} else if espMcu == "esp32s3" {
		// #----------------------------------------------------------------------------------
		//     xtensa-esp32s3

		t.Vars.Set("bootloader.elf.path", `{esp.arduino.sdk.path}/bin/bootloader_qio_80m.elf`)
		t.Vars.Set("flash.bootloader_bin_offset", `0x0`)
		t.Vars.Append("c.compiler.system.includes", `{esp.arduino.sdk.path}/qio_qspi/include`)
		t.Vars.Append("c.compiler.system.includes", "{esp.arduino.sdk.path}/include")
		t.Vars.Append("cpp.compiler.system.includes", `{esp.arduino.sdk.path}/qio_qspi/include`)
		t.Vars.Append("cpp.compiler.system.includes", "{esp.arduino.sdk.path}/include")
		t.Vars.Append("linker.system.library.paths", "{esp.arduino.sdk.path}/qio_qspi")
	}

	t.ResolveVars()
	return t, nil
}
