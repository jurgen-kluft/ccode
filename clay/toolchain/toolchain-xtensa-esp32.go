package toolchain

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
)

type ToolchainArduinoEsp32 struct {
	ToolchainInstance
	depTrackr   deptrackr.DepTrackr
	buildPath   string // Path to the build directory
	projectName string // Name of the project, used for output files
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// C Compiler

type ToolchainArduinoEsp32CCompiler struct {
	toolChain            *ToolchainArduinoEsp32
	compilerPath         string
	compilerArgs         []string
	compilerUserDefines  []string // User-defined macros for the compiler
	compilerUserIncludes []string // User-defined macros for the compiler
	config               string   // Configuration for the compiler, e.g., debug or release
}

func (t *ToolchainArduinoEsp32) NewCCompiler(config string) Compiler {
	return &ToolchainArduinoEsp32CCompiler{
		toolChain:            t,
		compilerPath:         t.Tools["c.compiler"],
		compilerArgs:         make([]string, 0, 64),
		compilerUserDefines:  make([]string, 0, 16),
		compilerUserIncludes: make([]string, 0, 16),
		config:               config,
	}
}

func (cl *ToolchainArduinoEsp32CCompiler) AddDefine(define string) {
	cl.compilerUserDefines = append(cl.compilerUserDefines, define)
}
func (cl *ToolchainArduinoEsp32CCompiler) AddIncludePath(path string) {
	cl.compilerUserIncludes = append(cl.compilerUserIncludes, path)
}
func (cl *ToolchainArduinoEsp32CCompiler) SetupArgs(userVars Vars) {

	cl.compilerArgs = cl.compilerArgs[:0] // Reset the arguments

	cl.compilerArgs = append(cl.compilerArgs, "-c")
	cl.compilerArgs = append(cl.compilerArgs, "-MMD")

	responseFileFlags := userVars.GetOne("c.compiler.response.flags")
	if len(responseFileFlags) > 0 {
		cl.compilerArgs = append(cl.compilerArgs, "@"+responseFileFlags)
	}

	switches := userVars.GetAll("c.compiler.switches")
	for _, s := range switches {
		cl.compilerArgs = append(cl.compilerArgs, s)
	}

	warningSwitches := userVars.GetAll("c.compiler.warning.switches")
	for _, s := range warningSwitches {
		cl.compilerArgs = append(cl.compilerArgs, s)
	}

	// Compiler system defines (debug / release ?)
	defines := userVars.GetAll("c.compiler.defines")
	for _, d := range defines {
		cl.compilerArgs = append(cl.compilerArgs, "-D")
		cl.compilerArgs = append(cl.compilerArgs, d)
	}

	// Compiler user defines
	for _, define := range cl.compilerUserDefines {
		cl.compilerArgs = append(cl.compilerArgs, "-D")
		cl.compilerArgs = append(cl.compilerArgs, define)
	}

	responseFileDefines := userVars.GetOne("c.compiler.response.defines")
	if len(responseFileDefines) > 0 {
		cl.compilerArgs = append(cl.compilerArgs, "@"+responseFileDefines)
	}

	// Compiler prefix include paths
	compilerPrefixInclude := userVars.GetOne("c.compiler.system.prefix.include")
	if len(compilerPrefixInclude) > 0 {
		cl.compilerArgs = append(cl.compilerArgs, "-iprefix")
		// Make sure the path ends with a /
		if !strings.HasSuffix(compilerPrefixInclude, "/") {
			cl.compilerArgs = append(cl.compilerArgs, compilerPrefixInclude+"/")
		} else {
			cl.compilerArgs = append(cl.compilerArgs, compilerPrefixInclude)
		}
	}

	// Compiler system include paths
	systemIncludes := userVars.GetAll("c.compiler.system.includes")
	for _, include := range systemIncludes {
		cl.compilerArgs = append(cl.compilerArgs, "-I")
		cl.compilerArgs = append(cl.compilerArgs, include)
	}

	responseFileIncludes := userVars.GetOne("c.compiler.response.includes")
	if len(responseFileIncludes) > 0 {
		cl.compilerArgs = append(cl.compilerArgs, "@"+responseFileIncludes)
	}
}

func (cl *ToolchainArduinoEsp32CCompiler) Compile(sourceAbsFilepath string, sourceRelFilepath string) (string, error) {

	numArgs := len(cl.compilerArgs)
	objFilepath := sourceRelFilepath + ".o"

	// The source file and the output object file
	// sourceAbsFilepath
	// -o
	// sourceRelFilepath + ".o"
	cl.compilerArgs = append(cl.compilerArgs, sourceAbsFilepath)
	cl.compilerArgs = append(cl.compilerArgs, "-o")
	cl.compilerArgs = append(cl.compilerArgs, objFilepath)

	cmd := exec.Command(cl.compilerPath, cl.compilerArgs...)
	out, err := cmd.CombinedOutput()

	// Reset the arguments to the initial state
	cl.compilerArgs = cl.compilerArgs[:numArgs]

	if err != nil {
		log.Printf("Compile failed, output:\n%s\n", string(out))
		return "", fmt.Errorf("Compile failed with %s\n", err)
	}
	if len(out) > 0 {
		log.Printf("Compile output:\n%s\n", string(out))
	}

	return objFilepath, nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// C++ Compiler

type ToolchainArduinoEsp32CppCompiler struct {
	toolChain            *ToolchainArduinoEsp32
	compilerPath         string
	compilerArgs         []string
	compilerUserDefines  []string // User-defined macros for the compiler
	compilerUserIncludes []string // User-defined macros for the compiler
	config               string   // Configuration for the compiler, e.g., debug or release
}

func (t *ToolchainArduinoEsp32) NewCppCompiler(config string) Compiler {
	return &ToolchainArduinoEsp32CppCompiler{
		toolChain:            t,
		compilerPath:         t.Tools["cpp.compiler"],
		compilerArgs:         []string{},
		compilerUserDefines:  make([]string, 0, 16),
		compilerUserIncludes: make([]string, 0, 16),
		config:               config,
	}
}

func (cl *ToolchainArduinoEsp32CppCompiler) AddDefine(define string) {
	cl.compilerUserDefines = append(cl.compilerUserDefines, define)
}
func (cl *ToolchainArduinoEsp32CppCompiler) AddIncludePath(path string) {
	cl.compilerUserIncludes = append(cl.compilerUserIncludes, path)
}
func (cl *ToolchainArduinoEsp32CppCompiler) SetupArgs(userVars Vars) {

	cl.compilerArgs = cl.compilerArgs[:0] // Reset the arguments

	cl.compilerArgs = append(cl.compilerArgs, "-c")
	cl.compilerArgs = append(cl.compilerArgs, "-MMD")

	responseFileFlags := userVars.GetOne("cpp.compiler.response.flags")
	if len(responseFileFlags) > 0 {
		cl.compilerArgs = append(cl.compilerArgs, "@"+responseFileFlags)
	}

	switches := userVars.GetAll("cpp.compiler.switches")
	for _, s := range switches {
		cl.compilerArgs = append(cl.compilerArgs, s)
	}

	warningSwitches := userVars.GetAll("cpp.compiler.warning.switches")
	for _, s := range warningSwitches {
		cl.compilerArgs = append(cl.compilerArgs, s)
	}

	// Compiler system defines (debug / release ?)
	defines := userVars.GetAll("cpp.compiler.defines")
	for _, d := range defines {
		cl.compilerArgs = append(cl.compilerArgs, "-D")
		cl.compilerArgs = append(cl.compilerArgs, d)
	}

	// Compiler user defines
	for _, define := range cl.compilerUserDefines {
		cl.compilerArgs = append(cl.compilerArgs, "-D")
		cl.compilerArgs = append(cl.compilerArgs, define)
	}

	responseFileDefines := userVars.GetOne("cpp.compiler.response.defines")
	if len(responseFileDefines) > 0 {
		cl.compilerArgs = append(cl.compilerArgs, "@"+responseFileDefines)
	}

	// Compiler prefix include paths
	compilerPrefixInclude := userVars.GetOne("cpp.compiler.system.prefix.include")
	if len(compilerPrefixInclude) > 0 {
		cl.compilerArgs = append(cl.compilerArgs, "-iprefix")
		// Make sure the path ends with a /
		if !strings.HasSuffix(compilerPrefixInclude, "/") {
			cl.compilerArgs = append(cl.compilerArgs, compilerPrefixInclude+"/")
		} else {
			cl.compilerArgs = append(cl.compilerArgs, compilerPrefixInclude)
		}
	}

	// Compiler system include paths
	systemIncludes := userVars.GetAll("cpp.compiler.system.includes")
	for _, include := range systemIncludes {
		cl.compilerArgs = append(cl.compilerArgs, "-I")
		cl.compilerArgs = append(cl.compilerArgs, include)
	}

	responseFileIncludes := userVars.GetOne("cpp.compiler.response.includes")
	if len(responseFileIncludes) > 0 {
		cl.compilerArgs = append(cl.compilerArgs, "@"+responseFileIncludes)
	}
}
func (cl *ToolchainArduinoEsp32CppCompiler) Compile(sourceAbsFilepath string, sourceRelFilepath string) (string, error) {
	numArgs := len(cl.compilerArgs)
	objFilepath := sourceRelFilepath + ".o"

	// The source file and the output object file
	// sourceAbsFilepath
	// -o
	// sourceRelFilepath + ".o"
	cl.compilerArgs = append(cl.compilerArgs, sourceAbsFilepath)
	cl.compilerArgs = append(cl.compilerArgs, "-o")
	cl.compilerArgs = append(cl.compilerArgs, objFilepath)

	cmd := exec.Command(cl.compilerPath, cl.compilerArgs...)
	out, err := cmd.CombinedOutput()

	// Reset the arguments to the initial state
	cl.compilerArgs = cl.compilerArgs[:numArgs]

	if err != nil {
		log.Printf("Compile failed, output:\n%s\n", string(out))
		return "", fmt.Errorf("Compile failed with %s\n", err)
	}
	if len(out) > 0 {
		log.Printf("Compile output:\n%s\n", string(out))
	}

	return objFilepath, nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Archiver

type ToolchainArduinoEsp32Archiver struct {
	toolChain    *ToolchainArduinoEsp32
	buildPath    string // Path to the build directory
	config       string
	archiverPath string
	archiverArgs []string
}

func (t *ToolchainArduinoEsp32) NewArchiver(config string) Archiver {
	return &ToolchainArduinoEsp32Archiver{
		toolChain:    t,
		buildPath:    t.buildPath,
		config:       config,
		archiverPath: t.Tools["archiver"],
		archiverArgs: []string{},
	}
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
	toolChain     *ToolchainArduinoEsp32
	buildPath     string // Path to the build directory
	config        string
	createMapFile bool
	libraryPaths  []string // Paths to libraries to link against
	libraryFiles  []string // Library files to link against
	linkerPath    string
	linkerArgs    []string
}

func (t *ToolchainArduinoEsp32) NewLinker(config string) Linker {
	return &ToolchainArduinoEsp32Linker{
		toolChain:     t,
		buildPath:     t.buildPath,
		config:        config,
		createMapFile: false, // Set to true if you want to generate a map file
		libraryPaths:  make([]string, 0, 16),
		libraryFiles:  make([]string, 0, 16),
		linkerPath:    t.Tools["linker"],
		linkerArgs:    make([]string, 0, 64),
	}
}

func (l *ToolchainArduinoEsp32Linker) GenerateMapFile() {
	l.createMapFile = true
}
func (l *ToolchainArduinoEsp32Linker) AddLibraryPath(path string) {
	l.libraryPaths = append(l.libraryPaths, path)
}
func (l *ToolchainArduinoEsp32Linker) AddLibraryFile(lib string) {
	l.libraryFiles = append(l.libraryFiles, lib)
}

func (l *ToolchainArduinoEsp32Linker) SetupArgs(userVars Vars) {

	l.linkerArgs = l.linkerArgs[:0] // Reset the arguments

	if l.createMapFile {
		l.linkerArgs = append(l.linkerArgs, "genmap")
	}

	l.linkerArgs = append(l.linkerArgs, "-Wl,--wrap=esp_panic_handler")

	linkerResponseFile := userVars["linker.response.ldflags"]
	if len(linkerResponseFile) == 1 {
		l.linkerArgs = append(l.linkerArgs, "@"+linkerResponseFile[0])
	}

	linkerResponseFile = userVars["linker.response.ldscripts"]
	if len(linkerResponseFile) == 1 {
		l.linkerArgs = append(l.linkerArgs, "@"+linkerResponseFile[0])
	}

	linkerResponseFile = userVars["linker.response.ldlibs"]
	if len(linkerResponseFile) == 1 {
		l.linkerArgs = append(l.linkerArgs, "-Wl,--start-group")
		l.linkerArgs = append(l.linkerArgs, "@"+linkerResponseFile[0])
		l.linkerArgs = append(l.linkerArgs, "-Wl,--end-group")
	}

	l.linkerArgs = append(l.linkerArgs, "-Wl,-EL")
}

func (l *ToolchainArduinoEsp32Linker) Link(inputArchiveAbsFilepaths []string, outputAppRelFilepathNoExt string) error {
	//
	if l.linkerArgs[0] == "genmap" {
		outputMapFilepath := outputAppRelFilepathNoExt + ".map"
		l.linkerArgs = append(l.linkerArgs, "-Wl,--Map="+outputMapFilepath)
	}

	// Finalize the linker arguments, remember how many arguments we initially had
	argCount := len(l.linkerArgs)

	linker := l.linkerPath
	linkerArgs := l.linkerArgs

	for _, libFile := range inputArchiveAbsFilepaths {
		linkerArgs = append(linkerArgs, "-l")
		linkerArgs = append(linkerArgs, libFile)
	}
	linkerArgs = append(linkerArgs, "-o")
	linkerArgs = append(linkerArgs, filepath.Join(l.buildPath, outputAppRelFilepathNoExt+".elf"))

	log.Printf("Linking '%s'...\n", outputAppRelFilepathNoExt+".elf")
	cmd := exec.Command(linker, linkerArgs...)
	out, err := cmd.CombinedOutput()

	// Reset the args to the initial state
	linkerArgs = linkerArgs[:argCount]

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
	buildPath                       string // Path to the build directory
	projectName                     string
	config                          string // Configuration for the burner, e.g., debug or release
	generateImageBinToolArgs        Arguments
	generateImageBinToolPath        string
	generateImagePartitionsToolArgs Arguments
	generateImagePartitionsToolPath string
	generateBootloaderToolArgs      Arguments
	generateBootloaderToolPath      string
	flashToolArgs                   Arguments
	flashToolPath                   string
}

func (t *ToolchainArduinoEsp32) NewBurner(config string) Burner {
	return &ToolchainArduinoEsp32Burner{
		toolChain:                       t,
		buildPath:                       t.buildPath,
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

func (b *ToolchainArduinoEsp32Burner) SetupBuildArgs(userVars Vars) {

	projectElfFilepath := filepath.Join(b.buildPath, b.projectName+".elf")
	projectBinFilepath := filepath.Join(b.buildPath, b.projectName+".bin")
	projectPartitionsBinFilepath := filepath.Join(b.buildPath, b.projectName+".partitions.bin")
	projectBootloaderBinFilepath := filepath.Join(b.buildPath, b.projectName+".bootloader.bin")

	b.generateImageBinToolArgs.Clear()
	b.generateImageBinToolArgs.Add("--chip", b.toolChain.Vars.GetOne("esp.mcu"))
	b.generateImageBinToolArgs.Add("elf2image")
	b.generateImageBinToolArgs.Add("--flash_mode", userVars.GetOne("burner.flash.mode"))
	b.generateImageBinToolArgs.Add("--flash_freq", userVars.GetOne("burner.flash.frequency"))
	b.generateImageBinToolArgs.Add("--flash_size", userVars.GetOne("burner.flash.size"))
	b.generateImageBinToolArgs.Add("--elf-sha256-offset", userVars.GetOne("burner.elf.sha256.offset"))
	b.generateImageBinToolArgs.Add("-o", projectBinFilepath)
	b.generateImageBinToolArgs.Add(projectElfFilepath)

	b.generateImagePartitionsToolArgs.Clear()
	b.generateImagePartitionsToolArgs.Add(b.toolChain.Vars.GetOne("burner.partitions.bin.script"))
	b.generateImagePartitionsToolArgs.Add("-q")
	b.generateImagePartitionsToolArgs.Add(b.toolChain.Vars.GetOne("burner.partitions.csv.filepath"))
	b.generateImagePartitionsToolArgs.Add(projectPartitionsBinFilepath)

	b.generateBootloaderToolArgs.Clear()
	b.generateBootloaderToolArgs.Add("--chip", b.toolChain.Vars.GetOne("esp.mcu"))
	b.generateBootloaderToolArgs.Add("elf2image")
	b.generateBootloaderToolArgs.Add("--flash_mode", userVars.GetOne("burner.flash.mode"))
	b.generateBootloaderToolArgs.Add("--flash_freq", userVars.GetOne("burner.flash.frequency"))
	b.generateBootloaderToolArgs.Add("--flash_size", userVars.GetOne("burner.flash.size"))
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

func (b *ToolchainArduinoEsp32Burner) SetupBurnArgs(userVars Vars) {

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

	projectNameFilepath := filepath.Join(b.buildPath, b.projectName)
	bootloaderBinFilepath := projectNameFilepath + ".bootloader.bin"
	partitionsBinFilepath := projectNameFilepath + ".partitions.bin"
	bootApp0BinFilePath := b.toolChain.Vars.GetOne("burner.bootapp0.bin.filepath")
	applicationBinFilepath := projectNameFilepath + ".bin"

	b.flashToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.bootloader.offset"))
	b.flashToolArgs.Add(bootloaderBinFilepath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.partitions.offset"))
	b.flashToolArgs.Add(partitionsBinFilepath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.bootapp0.offset"))
	b.flashToolArgs.Add(bootApp0BinFilePath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetOne("burner.flash.application.offset"))
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
func NewToolchainArduinoEsp32(espMcu string, buildPath string, projectName string) (t *ToolchainArduinoEsp32, err error) {

	var espSdkPath string
	if espSdkPath, err = exec.LookPath("ESP_SDK"); err != nil {
		return nil, err
	}

	t = &ToolchainArduinoEsp32{
		buildPath:   buildPath,
		projectName: projectName,
		ToolchainInstance: ToolchainInstance{
			Name: "xtensa-esp32",
			Vars: Vars{
				"esp.mcu":              {espMcu},
				"esp.sdk.path":         {espSdkPath},
				"esp.sdk.version":      {`3.2.0`},
				"esp.arduino.sdk.path": {`{esp.sdk.path}/tools/esp32-arduino-libs`},

				"c.compiler.generate.mapfile":  {"-MMD"},
				"c.compiler.response.flags":    {`@{esp.sdk.path}/flags/c_flags`},
				"c.compiler.response.defines":  {`@{esp.sdk.path}/flags/defines`},
				"c.compiler.response.includes": {`@{esp.sdk.path}/flags/includes`},
				"c.compiler.switches":          {`-w`, `-Os`},
				"c.compiler.warning.switches":  {`-Werror=return-type`},
				"c.compiler.defines": {
					`F_CPU=240000000L`,
					`ARDUINO=10605`,
					`ARDUINO_ESP32_DEV`,
					`ARDUINO_ARCH_ESP32`,
					`ARDUINO_BOARD="ESP32_DEV"`,
					`ARDUINO_VARIANT={esp.mcu}`,
					`ARDUINO_PARTITION_default`,
					`ARDUINO_HOST_OS=darwin`,
					`ARDUINO_FQBN="generic"`,
					`ESP32=ESP32`,
					`CORE_DEBUG_LEVEL=0`,
					`ARDUINO_USB_CDC_ON_BOOT=0`,
				},
				"c.compiler.system.prefix.include": {`{esp.sdk.path}/include`},
				"c.compiler.system.includes": {
					`{esp.sdk.path}/cores/esp32`,
					`{esp.sdk.path}/variants/{esp.mcu}`,
				},

				"cpp.compiler.generate.mapfile":  {"-MMD"},
				"cpp.compiler.response.flags":    {`@{esp.sdk.path}/flags/cpp_flags`},
				"cpp.compiler.response.defines":  {`@{esp.sdk.path}/flags/defines`},
				"cpp.compiler.response.includes": {`@{esp.sdk.path}/flags/includes`},
				"cpp.compiler.switches":          {`-w`, `-Os`},
				"cpp.compiler.warning.switches":  {`-Werror=return-type`},
				"cpp.compiler.defines": {
					`F_CPU=240000000L`,
					`ARDUINO=10605`,
					`ARDUINO_ESP32_DEV`,
					`ARDUINO_ARCH_ESP32`,
					`ARDUINO_BOARD="ESP32_DEV"`,
					`ARDUINO_VARIANT={esp.mcu}`,
					`ARDUINO_PARTITION_default`,
					`ARDUINO_HOST_OS=darwin`,
					`ARDUINO_FQBN="generic"`,
					`ESP32=ESP32`,
					`CORE_DEBUG_LEVEL=0`,
					`ARDUINO_USB_CDC_ON_BOOT=0`,
				},
				"cpp.compiler.system.prefix.include": {`{esp.sdk.path}/include`},
				"cpp.compiler.system.includes": {
					`{esp.sdk.path}/cores/esp32`,
					`{esp.sdk.path}/variants/{esp.mcu}`,
				},

				"archiver": {},

				"linker":                    {},
				"linker.response.ldflags":   {`@{esp.arduino.sdk.path}/flags/ld_flags`},
				"linker.response.ldscripts": {`@{esp.arduino.sdk.path}/flags/ld_scripts`},
				"linker.response.ldlibs":    {`@{esp.arduino.sdk.path}/flags/ld_libs`},

				"burner.generate-image-bin.script": {`{esp.sdk}/tools/gen_esp32part.py`},

				"burner.generate-partitions-bin.script": {`{esp.sdk.path}/tools/gen_esp32part.py`},

				"burner.flash.baud":                   {`921600`},
				"burner.flash.mode":                   {`dio`},
				"burner.flash.frequency":              {`40m`},
				"burner.flash.size":                   {`4MB`},
				"burner.flash.port":                   {`/dev/tty.usbmodem4101`},
				"burner.flash.elf_share_offset":       {`0xb0`},
				"burner.flash.partition_csv_file":     {`{esp.sdk.path}/tools/partitions/default.csv`},
				"burner.flash.bootloader_bin_offset":  {`0x1000`},
				"burner.flash.partitions_bin_offset":  {`0x8000`},
				"burner.flash.bootapp0_bin_offset":    {`0xe000`},
				"burner.flash.application_bin_offset": {`0x10000`},
			},
			Tools: map[string]string{
				"c.compiler":                     `{esp.sdk.path}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-gcc`,
				"cpp.compiler":                   `{esp.sdk.path}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-g++`,
				"archiver":                       `{esp.sdk.path}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-gcc-ar`,
				"linker":                         `{esp.sdk.path}tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-g++`,
				"burner.generate-bootloader":     `{esp.sdk.path}/tools/esptool/esptool`,
				"burner.generate-image-bin":      `python3`,
				"burner.generate-partitions-bin": `python3`,
				"burner.generate-elf-size":       `{esp.sdk}/tools/xtensa-esp-elf/bin/xtensa-{esp.mcu}-elf-size"`,
				"burner.flash":                   `{esp.sdk.path}/tools/esptool/esptool`,
			},
		},
	}

	if espMcu == "esp32" {
		// #----------------------------------------------------------------------------------
		//     xtensa-esp32
		t.Vars.Set("esp.mcu", `esp32`)
		t.Vars.Set("bootloader.elf.path", `{esp.arduino.sdk.path}/bin/bootloader_dio_40m.elf`)
		t.Vars.Set("flash.bootloader_bin_offset", `0x1000`)
		t.Vars.Append("c.compiler.system.includes", `{esp.sdk.path}/dio_qspi/include`)
		t.Vars.Append("cpp.compiler.system.includes", `{esp.sdk.path}/dio_qspi/include`)
	} else if espMcu == "esp32s3" {
		// #----------------------------------------------------------------------------------
		//     xtensa-esp32s3
		t.Vars.Set("esp.mcu", `esp32s3`)
		t.Vars.Set("bootloader.elf.path", `{esp.arduino.sdk.path}/bin/bootloader_qio_80m.elf`)
		t.Vars.Set("flash.bootloader_bin_offset", `0x0`)
		t.Vars.Append("c.compiler.system.includes", `{esp.sdk.path}/qio_qspi/include`)
		t.Vars.Append("cpp.compiler.system.includes", `{esp.sdk.path}/qio_qspi/include`)
	}

	t.ResolveVars()
	return t, nil
}
