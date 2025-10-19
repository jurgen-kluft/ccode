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
	"slices"
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

func (cl *ToolchainArduinoEsp8266Compiler) SetupArgs(defines []string, includes []string) {
	for i, inc := range includes {
		includes[i] = "-I" + inc
	}
	for i, def := range defines {
		defines[i] = "-D" + def
	}

	cl.vars.Clear()
	cl.vars.Append("includes", "-I{runtime.platform.path}/variants/{board.name}")
	cl.vars.Append("includes", includes...)
	cl.vars.Append("build.defines", defines...)
}

func (cl *ToolchainArduinoEsp8266Compiler) Compile(sourceAbsFilepaths []string, objRelFilepaths []string) error {
	for i, sourceAbsFilepath := range sourceAbsFilepaths {
		corepkg.LogInfof("Compiling %s", filepath.Base(sourceAbsFilepath))

		objRelFilepath := objRelFilepaths[i]
		cl.vars.Set("build.source.path", corepkg.PathDirname(sourceAbsFilepath))
		cl.vars.Set("source_file", sourceAbsFilepath)
		cl.vars.Set("object_file", objRelFilepath)

		var compilerPath string
		var compilerArgs []string
		if strings.HasSuffix(sourceAbsFilepath, ".c") {
			c_compiler, _ := cl.toolChain.Vars.Get(`recipe.c.o.pattern`)
			compilerPath = c_compiler[0]
			compilerArgs = c_compiler[1:]
			compilerPath = cl.toolChain.Vars.FinalResolveString(compilerPath, " ", cl.vars)
			compilerArgs = cl.toolChain.Vars.FinalResolveArray(compilerArgs, cl.vars)
		} else {
			cpp_compiler, _ := cl.toolChain.Vars.Get(`recipe.cpp.o.pattern`)
			compilerPath = cpp_compiler[0]
			compilerArgs = cpp_compiler[1:]
			compilerPath = cl.toolChain.Vars.FinalResolveString(compilerPath, " ", cl.vars)
			compilerArgs = cl.toolChain.Vars.FinalResolveArray(compilerArgs, cl.vars)
		}
		compilerPath = corepkg.StrTrimDelimiters(compilerPath, '"')

		// Remove empty entries from compilerArgs
		compilerArgs = slices.DeleteFunc(compilerArgs, func(s string) bool {
			return strings.TrimSpace(s) == ""
		})

		cmd := exec.Command(compilerPath, compilerArgs...)
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
	toolChain   *ArduinoEsp8266Toolchain
	buildConfig denv.BuildConfig
	buildTarget denv.BuildTarget
	vars        *corepkg.Vars
}

func (t *ArduinoEsp8266Toolchain) NewArchiver(a ArchiverType, buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Archiver {
	return &ToolchainArduinoEsp8266Archiver{
		toolChain:   t,
		buildConfig: buildConfig,
		buildTarget: buildTarget,
		vars:        corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
	}
}

func (t *ToolchainArduinoEsp8266Archiver) LibFilepath(filepath string) string {
	// The file extension for the archive on ESP32 is typically ".a"
	return filepath + ".a"
}

func (a *ToolchainArduinoEsp8266Archiver) SetupArgs() {
}

func (a *ToolchainArduinoEsp8266Archiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {

	corepkg.LogInfof("Archiving %s", outputArchiveFilepath)

	a.vars.Set("archive_file_path", outputArchiveFilepath)
	a.vars.Set("object_file", inputObjectFilepaths...)

	archiverArgs, _ := a.toolChain.Vars.Get(`recipe.ar.pattern`)
	archiverPath := archiverArgs[0]
	archiverArgs = archiverArgs[1:]

	// Resolve archiverPath and archiverArgs
	archiverPath = a.toolChain.Vars.FinalResolveString(archiverPath, " ", a.vars)
	archiverArgs = a.toolChain.Vars.FinalResolveArray(archiverArgs, a.vars)

	// Remove any empty entries from archiverArgs
	archiverArgs = slices.DeleteFunc(archiverArgs, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})

	cmd := exec.Command(archiverPath, archiverArgs...)
	out, err := cmd.CombinedOutput()
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
	vars        *corepkg.Vars
}

func (t *ArduinoEsp8266Toolchain) NewLinker(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Linker {
	return &ToolchainArduinoEsp8266Linker{
		toolChain:   t,
		buildConfig: buildConfig,
		buildTarget: buildTarget,
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
		libFile = strings.TrimPrefix(libFile, "lib")
		libFile = strings.TrimSuffix(libFile, ".a")
		libraryFiles[i] = "-l" + libFile
	}

	l.vars.Clear()
	l.vars.Set("build.extra_libs", libraryPaths...)
	l.vars.Set("build.extra_libs", libraryFiles...)
}

func (l *ToolchainArduinoEsp8266Linker) Link(inputArchiveAbsFilepaths []string, outputAppRelFilepathNoExt string) error {
	corepkg.LogInfof("Linking '%s'...", outputAppRelFilepathNoExt)

	linkerArgs, _ := l.toolChain.Vars.Get(`recipe.c.combine.pattern`)

	linkerPath := linkerArgs[0]
	linkerArgs = linkerArgs[1:]

	l.vars.Set("object_files", inputArchiveAbsFilepaths...)

	// Split 'outputAppRelFilepathNoExt' into directory and filename parts
	// Remove the extension of the output file if present
	outputDir := filepath.Dir(outputAppRelFilepathNoExt)
	outputFile := corepkg.PathFilename(outputAppRelFilepathNoExt, false)

	l.toolChain.Vars.Set("build.path", outputDir)
	l.toolChain.Vars.Set("build.project_name", outputFile)

	// Resolve linkerPath and linkerArgs
	linkerPath = l.toolChain.Vars.FinalResolveString(linkerPath, " ", l.vars)
	linkerArgs = l.toolChain.Vars.FinalResolveArray(linkerArgs, l.vars)

	// Do we need to fill in the arg to generate map file?
	generateMapfile := true
	if generateMapfile {
		outputMapFilepath := outputAppRelFilepathNoExt + ".map"
		linkerArgs = append(linkerArgs, "-Wl,--Map="+outputMapFilepath)
	}

	// Remove any empty entries from linkerArgs
	linkerArgs = slices.DeleteFunc(linkerArgs, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})

	cmd := exec.Command(linkerPath, linkerArgs...)
	out, err := cmd.CombinedOutput()

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
