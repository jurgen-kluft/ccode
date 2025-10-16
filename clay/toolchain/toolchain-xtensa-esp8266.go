package toolchain

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
	"github.com/jurgen-kluft/ccode/denv"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

type ArduinoEsp8266Toolchain struct {
	Vars        *corepkg.Vars
	ProjectName string // The name of the project, used for output files
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Compiler

type ToolchainArduinoEsp8266Compiler struct {
	toolChain   *ArduinoEsp8266Toolchain
	buildConfig denv.BuildConfig
	buildTarget denv.BuildTarget
	// Configuration for the compiler, e.g., debug or release
	cCompilerPath   string
	cppCompilerPath string
	cCompilerArgs   *corepkg.Arguments
	cppCompilerArgs *corepkg.Arguments
	vars            *corepkg.Vars
}

func (t *ArduinoEsp8266Toolchain) NewCompiler(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Compiler {
	return &ToolchainArduinoEsp8266Compiler{
		toolChain:       t,
		buildConfig:     buildConfig,
		buildTarget:     buildTarget,
		cCompilerPath:   "",
		cppCompilerPath: "",
		cCompilerArgs:   corepkg.NewArguments(64),
		cppCompilerArgs: corepkg.NewArguments(64),
		vars:            corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
	}
}

func (cl *ToolchainArduinoEsp8266Compiler) ObjFilepath(srcRelFilepath string) string {
	return srcRelFilepath + ".o"
}

func (cl *ToolchainArduinoEsp8266Compiler) DepFilepath(objRelFilepath string) string {
	return objRelFilepath + ".d"
}

func setupCompilerArgs(args []string) (string, []string) {
	cpArgs := make([]string, 0, len(args)+32)
	if len(args) > 0 {
		for _, part := range args {
			part = strings.Trim(part, " ")
			if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {
				part = strings.Trim(part, "\"")
			}
			if len(part) > 0 {
				cpArgs = append(cpArgs, part)
			}
		}
		return cpArgs[0], cpArgs[1:]
	}
	return "", []string{}
}

func (cl *ToolchainArduinoEsp8266Compiler) SetupArgs(defines []string, includes []string) {
	for i, inc := range includes {
		includes[i] = "-I" + inc
	}
	for i, def := range defines {
		defines[i] = "-D" + def
	}

	cl.vars.Set("includes", includes...)
	cl.vars.Set("defines", defines...)

	c_compiler := cl.toolChain.Vars.GetFirstOrEmpty(`recipe.c.o.pattern`)
	c_compiler_parts := corepkg.NewArguments(32)
	c_compiler_parts.SmartSplit(c_compiler)
	cl.cCompilerPath, cl.cCompilerArgs.Args = setupCompilerArgs(c_compiler_parts.Args)

	cl.cCompilerPath = cl.toolChain.Vars.FinalResolveString(cl.cCompilerPath, " ")
	cl.cCompilerArgs.Args = cl.toolChain.Vars.FinalResolveArrayWith(cl.cCompilerArgs.Args, cl.vars)

	cpp_compiler := cl.toolChain.Vars.GetFirstOrEmpty(`recipe.cpp.o.pattern`)
	cpp_compiler_parts := corepkg.NewArguments(32)
	cpp_compiler_parts.SmartSplit(cpp_compiler)
	cl.cppCompilerPath, cl.cppCompilerArgs.Args = setupCompilerArgs(cpp_compiler_parts.Args)

	cl.cppCompilerPath = cl.toolChain.Vars.FinalResolveString(cl.cppCompilerPath, " ")
	cl.cppCompilerArgs.Args = cl.toolChain.Vars.FinalResolveArrayWith(cl.cppCompilerArgs.Args, cl.vars)
}

func (cl *ToolchainArduinoEsp8266Compiler) Compile(sourceAbsFilepaths []string, objRelFilepaths []string) error {

	for i, sourceAbsFilepath := range sourceAbsFilepaths {
		objRelFilepath := objRelFilepaths[i]

		var compilerPath string
		var args []string
		if strings.HasSuffix(sourceAbsFilepath, ".c") {
			compilerPath = cl.cCompilerPath
			args = cl.cCompilerArgs.Args
		} else {
			compilerPath = cl.cppCompilerPath
			args = cl.cppCompilerArgs.Args
		}

		// The source file and the output object file
		// sourceAbsFilepath
		// -o
		// sourceRelFilepath + ".o"
		args = append(args, sourceAbsFilepath)
		args = append(args, "-o")
		args = append(args, objRelFilepath)

		corepkg.LogInfof("Compiling (%s) %s", cl.buildConfig.String(), filepath.Base(sourceAbsFilepath))

		cmd := exec.Command(compilerPath, args...)
		out, err := cmd.CombinedOutput()

		if err != nil {
			corepkg.LogInfof("Compile failed, output:\n%s", string(out))
			return corepkg.LogErrorf(err, "Compiling failed")
		}
		if len(out) > 0 {
			corepkg.LogInfof("Compile output:\n%s", string(out))
		}
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Archiver

type ToolchainArduinoEsp8266Archiver struct {
	toolChain    *ArduinoEsp8266Toolchain
	buildConfig  denv.BuildConfig
	buildTarget  denv.BuildTarget
	archiverPath string
	archiverArgs *corepkg.Arguments
}

func (t *ArduinoEsp8266Toolchain) NewArchiver(a ArchiverType, buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Archiver {
	return &ToolchainArduinoEsp8266Archiver{
		toolChain:    t,
		buildConfig:  buildConfig,
		buildTarget:  buildTarget,
		archiverPath: "",
		archiverArgs: corepkg.NewArguments(16),
	}
}

func (t *ToolchainArduinoEsp8266Archiver) LibFilepath(filepath string) string {
	// The file extension for the archive on ESP32 is typically ".a"
	return filepath + ".a"
}

func setupArchiverArgs(args *corepkg.Arguments) (arPath string, arArgs *corepkg.Arguments) {
	arArgs = corepkg.NewArguments(8)
	if args.Len() > 0 {
		arPath = corepkg.StrTrimDelimiters(args.Args[0], '"')
		for _, part := range args.Args[1:] {
			part = corepkg.StrTrim(part)
			part = corepkg.StrTrimDelimiters(part, '"')
			if len(part) > 0 {
				arArgs.Add(part)
			}
		}
	}
	return arPath, arArgs
}

func (a *ToolchainArduinoEsp8266Archiver) SetupArgs() {
	ar_cmdline := a.toolChain.Vars.GetFirstOrEmpty(`recipe.ar.pattern`)
	ar_arg_parts := corepkg.NewArguments(32)
	ar_arg_parts.SmartSplit(ar_cmdline)
	a.archiverPath, a.archiverArgs = setupArchiverArgs(ar_arg_parts)
}

func (a *ToolchainArduinoEsp8266Archiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {

	argLen := a.archiverArgs.Len()

	// {output-archive-filepath}
	a.archiverArgs.Add(outputArchiveFilepath)

	// {input-object-filepaths}
	a.archiverArgs.Add(inputObjectFilepaths...)

	corepkg.LogInfof("Archiving %s", outputArchiveFilepath)

	cmd := exec.Command(a.archiverPath, a.archiverArgs.Args...)
	out, err := cmd.CombinedOutput()

	// Restore the archiver original arguments
	a.archiverArgs.Truncate(argLen)

	if err != nil {
		return corepkg.LogErrorf(err, "Archiving failed")
	}
	if len(out) > 0 {
		corepkg.LogInfof("Archive output:\n%s", string(out))
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Linker

type ToolchainArduinoEsp8266Linker struct {
	toolChain   *ArduinoEsp8266Toolchain
	buildConfig denv.BuildConfig
	buildTarget denv.BuildTarget
	linkerPath  string
	linkerArgs  *corepkg.Arguments
	vars        *corepkg.Vars
}

func (t *ArduinoEsp8266Toolchain) NewLinker(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Linker {
	return &ToolchainArduinoEsp8266Linker{
		toolChain:   t,
		buildConfig: buildConfig,
		buildTarget: buildTarget,
		linkerPath:  "",
		linkerArgs:  corepkg.NewArguments(512),
		vars:        corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
	}
}

// FileExt returns the file extension for the linker output, which is ".elf" for ESP32.
func (l *ToolchainArduinoEsp8266Linker) LinkedFilepath(filepath string) string {
	return filepath + ".elf"
}

func (l *ToolchainArduinoEsp8266Linker) SetupArgs(libraryPaths []string, libraryFiles []string) {
	for i, libPath := range libraryPaths {
		libraryPaths[i] = "-L" + libPath
	}
	for i, libFile := range libraryFiles {
		if strings.HasPrefix(libFile, "lib") && strings.HasSuffix(libFile, ".a") {
			libFile = libFile[3 : len(libFile)-2]
		}
		libraryFiles[i] = "-l" + libFile
	}
	l.vars.Set("build.extra_libs", libraryPaths...)
	l.vars.Set("build.extra_libs", libraryFiles...)

	ld_cmdline := l.toolChain.Vars.GetFirstOrEmpty(`recipe.c.combine.pattern`)
	ld_arg_parts := corepkg.NewArguments(32)
	ld_arg_parts.SmartSplit(ld_cmdline)
	l.linkerPath, l.linkerArgs = setupArchiverArgs(ld_arg_parts)

	l.linkerPath = l.toolChain.Vars.FinalResolveString(l.linkerPath, " ")
	l.linkerArgs.Args = l.toolChain.Vars.FinalResolveArrayWith(l.linkerArgs.Args, l.vars)
}

func (l *ToolchainArduinoEsp8266Linker) Link(inputArchiveAbsFilepaths []string, outputAppRelFilepathNoExt string) error {
	linker := l.linkerPath

	linkerArgs := *l.linkerArgs

	linkerArgs.Add(inputArchiveAbsFilepaths...)

	linkerArgs.Add("-o")
	linkerArgs.Add(outputAppRelFilepathNoExt)

	// Do we need to fill in the arg to generate map file?
	generateMapfile := linkerArgs.Args[0] == "genmap"
	if generateMapfile {
		outputMapFilepath := outputAppRelFilepathNoExt + ".map"
		linkerArgs.Args[0] = "-Wl,--Map=" + outputMapFilepath
	}

	corepkg.LogInfof("Linking '%s'...", outputAppRelFilepathNoExt)
	cmd := exec.Command(linker, linkerArgs.Args...)
	out, err := cmd.CombinedOutput()

	// Reset the map generation command in the arguments so that it
	// will be updated correctly on the next Link() call.
	if generateMapfile {
		l.linkerArgs.Args[0] = "genmap"
	}

	if err != nil {
		corepkg.LogInfof("Link failed, output:\n%s", string(out))
		return corepkg.LogErrorf(err, "Linking failed")
	}
	if len(out) > 0 {
		corepkg.LogInfof("Link output:\n%s", string(out))
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Burner

type ToolchainArduinoEsp8266Burner struct {
	toolChain                            *ArduinoEsp8266Toolchain
	buildConfig                          denv.BuildConfig
	buildTarget                          denv.BuildTarget
	dependencyTracker                    deptrackr.FileTrackr // Dependency tracker for the burner
	hasher                               hash.Hash            // Hasher for generating digests of arguments
	genImageBinToolArgs                  *corepkg.Arguments
	genImageBinToolArgsHash              []byte   // Hash of the arguments for the image bin tool
	genImageBinToolOutputFilepath        string   // The output file for the image bin
	genImageBinToolInputFilepaths        []string // The input files for the image bin
	genImageBinToolPath                  string
	genImagePartitionsToolArgs           *corepkg.Arguments
	genImagePartitionsToolArgsHash       []byte   // Hash of the arguments for the image partitions tool
	genImagePartitionsToolOutputFilepath string   // The output file for the image partitions
	genImagePartitionsToolInputFilepaths []string // The input files for the image partitions
	genImagePartitionsToolPath           string
	genBootloaderToolArgs                *corepkg.Arguments
	genBootloaderToolArgsHash            []byte   // Hash of the arguments for the bootloader tool
	genBootloaderToolOutputFilepath      string   // The output file for the bootloader
	genBootloaderToolInputFilepaths      []string // The input files for the bootloader
	genBootloaderToolPath                string
	flashToolArgs                        *corepkg.Arguments
	flashToolPath                        string
}

func (t *ArduinoEsp8266Toolchain) NewBurner(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Burner {
	return &ToolchainArduinoEsp8266Burner{
		toolChain:                            t,
		buildConfig:                          buildConfig,
		buildTarget:                          buildTarget,
		dependencyTracker:                    nil,
		hasher:                               sha1.New(),
		genImageBinToolArgs:                  corepkg.NewArguments(64),
		genImageBinToolArgsHash:              nil, // Will be set later
		genImageBinToolOutputFilepath:        "",
		genImageBinToolInputFilepaths:        []string{},
		genImageBinToolPath:                  t.Vars.GetFirstOrEmpty("burner.generate-image-bin"),
		genImagePartitionsToolArgs:           corepkg.NewArguments(64),
		genImagePartitionsToolArgsHash:       nil, // Will be set later
		genImagePartitionsToolOutputFilepath: "",
		genImagePartitionsToolInputFilepaths: []string{},
		genImagePartitionsToolPath:           t.Vars.GetFirstOrEmpty("burner.generate-partitions-bin"),
		genBootloaderToolArgs:                corepkg.NewArguments(64),
		genBootloaderToolArgsHash:            nil, // Will be set later
		genBootloaderToolOutputFilepath:      "",
		genBootloaderToolInputFilepaths:      []string{},
		genBootloaderToolPath:                t.Vars.GetFirstOrEmpty("burner.generate-bootloader"),
		flashToolArgs:                        corepkg.NewArguments(64),
		flashToolPath:                        t.Vars.GetFirstOrEmpty("burner.flash"),
	}
}

func (b *ToolchainArduinoEsp8266Burner) hashArguments(args []string) []byte {
	b.hasher.Reset()
	for _, arg := range args {
		b.hasher.Write([]byte(arg))
	}
	return b.hasher.Sum(nil)
}

func (b *ToolchainArduinoEsp8266Burner) SetupBuild(buildPath string) {

	projectElfFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".elf")
	projectBinFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".bin")
	projectPartitionsBinFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".partitions.bin")
	projectBootloaderBinFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".bootloader.bin")

	flashMode := b.toolChain.Vars.GetFirstOrEmpty("build.flash_mode")
	flashFreq := b.toolChain.Vars.GetFirstOrEmpty("build.flash_freq")
	flashSize := b.toolChain.Vars.GetFirstOrEmpty("build.flash_size")

	// The XIAO_ESP32C3 board is configured as "qio" flash mode, 80MHz flash frequency and 4MB flash size.
	// However, the flash that is on the board can only be successfully flashed when using "dio" flash mode.
	if b.toolChain.Vars.GetFirstOrEmpty("build.mcu") == "esp32c3" {
		corepkg.LogWarnf("Overriding flash mode to 'dio' for XIAO_ESP32C3 board")
		flashMode = "dio"
	}

	b.genImageBinToolArgs.Clear()
	b.genImageBinToolArgs.Add("--chip", b.toolChain.Vars.GetFirstOrEmpty("build.mcu"))
	b.genImageBinToolArgs.Add("elf2image")
	b.genImageBinToolArgs.Add("--flash_mode", flashMode)
	b.genImageBinToolArgs.Add("--flash_freq", flashFreq)
	b.genImageBinToolArgs.Add("--flash_size", flashSize)
	b.genImageBinToolArgs.Add("--elf-sha256-offset", b.toolChain.Vars.GetFirstOrEmpty("elf-sha256-offset"))
	b.genImageBinToolArgs.Add("-o", projectBinFilepath)
	b.genImageBinToolArgs.Add(projectElfFilepath)

	b.genImagePartitionsToolArgs.Clear()
	b.genImagePartitionsToolArgs.Add(b.toolChain.Vars.GetFirstOrEmpty("burner.generate-partitions-bin.script"))
	b.genImagePartitionsToolArgs.Add("-q")
	b.genImagePartitionsToolArgs.Add(b.toolChain.Vars.GetFirstOrEmpty("burner.flash.partitions.csv.filepath"))
	b.genImagePartitionsToolArgs.Add(projectPartitionsBinFilepath)

	b.genBootloaderToolArgs.Clear()
	b.genBootloaderToolArgs.Add("--chip", b.toolChain.Vars.GetFirstOrEmpty("build.mcu"))
	b.genBootloaderToolArgs.Add("elf2image")
	b.genBootloaderToolArgs.Add("--flash_mode", flashMode)
	b.genBootloaderToolArgs.Add("--flash_freq", flashFreq)
	b.genBootloaderToolArgs.Add("--flash_size", flashSize)
	b.genBootloaderToolArgs.Add("-o", projectBootloaderBinFilepath)
	sdkBootLoaderElfPath, _ := b.toolChain.Vars.GetFirst("burner.sdk.bootloader.elf.path")
	b.genBootloaderToolArgs.Add(sdkBootLoaderElfPath)

	// File Dependency Tracker and Information
	b.dependencyTracker = deptrackr.LoadDepFileTrackr(filepath.Join(buildPath, "deptrackr.burn"))

	b.genImageBinToolOutputFilepath = projectBinFilepath
	b.genImageBinToolInputFilepaths = []string{projectElfFilepath}
	b.genImageBinToolArgsHash = b.hashArguments(b.genImageBinToolArgs.Args)

	b.genImagePartitionsToolOutputFilepath = projectPartitionsBinFilepath
	b.genImagePartitionsToolInputFilepaths = []string{}
	b.genImagePartitionsToolArgsHash = b.hashArguments(b.genImagePartitionsToolArgs.Args)

	b.genBootloaderToolOutputFilepath = projectBootloaderBinFilepath
	b.genBootloaderToolInputFilepaths = []string{}
	b.genBootloaderToolArgsHash = b.hashArguments(b.genBootloaderToolArgs.Args)
}

func (b *ToolchainArduinoEsp8266Burner) Build() error {

	// - Generate image partitions bin file ('PROJECT.NAME.partitions.bin'):
	// - Generate image bin file ('PROJECT.NAME.bin'):
	// - Generate bootloader image ('PROJECT.NAME.bootloader.bin'):

	// Generate the image partitions bin file
	if !b.dependencyTracker.QueryItemWithExtraData(b.genImagePartitionsToolOutputFilepath, b.genImagePartitionsToolArgsHash) {
		img, _ := exec.LookPath(b.genImagePartitionsToolPath)
		args := b.genImagePartitionsToolArgs.Args

		cmd := exec.Command(img, args...)
		corepkg.LogInfof("Creating image partitions '%s' ...", b.toolChain.ProjectName+".partitions.bin")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return corepkg.LogErrorf(err, "Creating image partitions failed")
		}
		if len(out) > 0 {
			corepkg.LogInfof("Image partitions output:\n%s", string(out))
		}
		fmt.Println()

		b.dependencyTracker.AddItemWithExtraData(b.genImagePartitionsToolOutputFilepath, b.genImagePartitionsToolArgsHash, b.genImagePartitionsToolInputFilepaths)
	} else {
		b.dependencyTracker.CopyItem(b.genImagePartitionsToolOutputFilepath)
	}

	// Generate the image bin file
	if !b.dependencyTracker.QueryItemWithExtraData(b.genImageBinToolOutputFilepath, b.genImageBinToolArgsHash) {
		imgPath := b.genImageBinToolPath
		args := b.genImageBinToolArgs.Args

		cmd := exec.Command(imgPath, args...)
		corepkg.LogInfof("Generating image '%s'", b.toolChain.ProjectName+".bin")
		out, err := cmd.CombinedOutput()
		if err != nil {
			if len(out) > 0 {
				corepkg.LogInfof("Image generation output:\n%s", string(out))
			}
			return corepkg.LogErrorf(err, "Image generation failed")
		}
		if len(out) > 0 {
			corepkg.LogInfof("Image generation output:\n%s", string(out))
		}
		fmt.Println()

		b.dependencyTracker.AddItemWithExtraData(b.genImageBinToolOutputFilepath, b.genImageBinToolArgsHash, b.genImageBinToolInputFilepaths)
	} else {
		b.dependencyTracker.CopyItem(b.genImageBinToolOutputFilepath)
	}

	// Generate the bootloader image
	if !b.dependencyTracker.QueryItemWithExtraData(b.genBootloaderToolOutputFilepath, b.genBootloaderToolArgsHash) {
		imgPath := b.genBootloaderToolPath
		args := b.genBootloaderToolArgs.Args

		cmd := exec.Command(imgPath, args...)
		corepkg.LogInfof("Generating bootloader '%s'", b.toolChain.ProjectName+".bootloader.bin")
		out, err := cmd.CombinedOutput()
		if err != nil {
			corepkg.LogInfof("Bootloader generation failed, output:\n%s", string(out))
			return corepkg.LogErrorf(err, "Bootloader generation failed")
		}
		if len(out) > 0 {
			corepkg.LogInfof("Bootloader generation output:\n%s", string(out))
		}
		fmt.Println()

		b.dependencyTracker.AddItemWithExtraData(b.genBootloaderToolOutputFilepath, b.genBootloaderToolArgsHash, b.genBootloaderToolInputFilepaths)
	} else {
		b.dependencyTracker.CopyItem(b.genBootloaderToolOutputFilepath)
	}

	b.dependencyTracker.Save()

	return nil
}

func (b *ToolchainArduinoEsp8266Burner) SetupBurn(buildPath string) error {

	b.flashToolArgs.Clear()
	b.flashToolArgs.Add("--chip", b.toolChain.Vars.GetFirstOrEmpty("build.mcu"))

	// If we leave this empty then the tool will search for the USB port
	// b.flashToolArgs.Add("--port", "tty port")

	b.flashToolArgs.Add("--baud", b.toolChain.Vars.GetFirstOrEmpty("upload.speed"))
	b.flashToolArgs.Add("--before", "default_reset")
	b.flashToolArgs.Add("--after", "hard_reset")
	b.flashToolArgs.Add("write_flash", "-z")
	b.flashToolArgs.Add("--flash_mode", "keep")
	b.flashToolArgs.Add("--flash_freq", "keep")
	b.flashToolArgs.Add("--flash_size", "keep")

	bootApp0BinFilePath, _ := b.toolChain.Vars.GetFirst("burner.bootapp0.bin.filepath")
	if !corepkg.FileExists(bootApp0BinFilePath) {
		return corepkg.LogErrorf(os.ErrNotExist, "Boot app0 bin file '%s' does not exist", bootApp0BinFilePath)
	}

	b.flashToolArgs.Add(b.toolChain.Vars.GetFirstOrEmpty("build.bootloader_addr"))
	b.flashToolArgs.Add(b.genBootloaderToolOutputFilepath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetFirstOrEmpty("burner.flash.partitions.bin.offset"))
	b.flashToolArgs.Add(b.genImagePartitionsToolOutputFilepath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetFirstOrEmpty("burner.flash.bootapp0.bin.offset"))
	b.flashToolArgs.Add(bootApp0BinFilePath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetFirstOrEmpty("burner.flash.application.bin.offset"))
	b.flashToolArgs.Add(b.genImageBinToolOutputFilepath)

	return nil
}

func (b *ToolchainArduinoEsp8266Burner) Burn() error {
	if !corepkg.FileExists(b.genBootloaderToolOutputFilepath) {
		return corepkg.LogErrorf(os.ErrNotExist, "Cannot burn, bootloader bin file '%s' doesn't exist", b.genBootloaderToolOutputFilepath)
	}
	if !corepkg.FileExists(b.genImagePartitionsToolOutputFilepath) {
		return corepkg.LogErrorf(os.ErrNotExist, "Cannot burn, partitions bin file '%s' doesn't exist", b.genImagePartitionsToolOutputFilepath)
	}
	if !corepkg.FileExists(b.genImageBinToolOutputFilepath) {
		return corepkg.LogErrorf(os.ErrNotExist, "Cannot burn, application bin file '%s' doesn't exist", b.genImageBinToolOutputFilepath)
	}

	flashToolPath := b.flashToolPath
	flashToolArgs := b.flashToolArgs.Args

	corepkg.LogInfof("Flashing '%s'...", b.toolChain.ProjectName+".bin")

	flashToolCmd := exec.Command(flashToolPath, flashToolArgs...)

	// out, err := flashToolCmd.CombinedOutput()
	// if err != nil {
	// 	if len(out) > 0 {
	// 		corepkg.LogInfof("Flashing output:\n%s", string(out))
	// 	}
	// 	return corepkg.LogErrorf(err, "Flashing failed with %s")
	// }
	// if len(out) > 0 {
	// 	corepkg.LogInfof("Flashing output:\n%s", string(out))
	// }

	pipe, _ := flashToolCmd.StdoutPipe()

	if err := flashToolCmd.Start(); err != nil {
		return corepkg.LogErrorf(err, "Flashing failed")
	}

	reader := bufio.NewReader(pipe)
	line, err := reader.ReadString('\n')
	for err == nil {
		line = strings.TrimRight(line, "\n")
		corepkg.LogInfo(line)
		line, err = reader.ReadString('\n')
		if err == io.EOF {
			err = nil
			break
		}
	}

	if err != nil {
		return corepkg.LogErrorf(err, "Flashing failed")
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Dependency Tracker
func (t *ArduinoEsp8266Toolchain) NewDependencyTracker(dirpath string) deptrackr.FileTrackr {
	return deptrackr.LoadDepFileTrackr(filepath.Join(dirpath, "deptrackr"))
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Toolchain for ESP8266 on Arduino
func NewArduinoEsp8266Toolchain(boardVars *corepkg.Vars, projectName string, buildPath string) *ArduinoEsp8266Toolchain {
	tc := &ArduinoEsp8266Toolchain{ProjectName: projectName, Vars: boardVars}

	boardVars.Set("project.name", projectName)
	boardVars.Set("build.path", buildPath)
	boardVars.Set("build.arch", "esp8266")
	boardVars.Set("build.includes", "{runtime.platform.path}/variants/{board.name}")
	boardVars.Set("build.defines", "")

	// TODO we are not doing a final resolve here, instead, each tool should resolve its own
	// command line and arguments based on its needs.
	//boardVars.FinalResolve()

	// Create '{buildPath}/core/build.opt'
	// TODO not sure what this file is for and who should create it with what content
	optFilePath := filepath.Join(buildPath, "core", "build.opt")
	os.MkdirAll(filepath.Dir(optFilePath), os.ModePerm)
	f, err := os.Create(optFilePath)
	if err == nil {
		defer f.Close()
	}

	return tc
}
