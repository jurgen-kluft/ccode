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
	"runtime"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
	"github.com/jurgen-kluft/ccode/foundation"
)

type ArduinoEsp32 struct {
	Vars        *foundation.Vars
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
	cCompilerArgs   *foundation.Arguments
	cppCompilerArgs *foundation.Arguments
}

func (t *ArduinoEsp32) NewCompiler(config *Config) Compiler {
	return &ToolchainArduinoEsp32Compiler{
		toolChain:       t,
		config:          config,
		cCompilerPath:   t.Vars.GetFirstOrEmpty("c.compiler"),
		cppCompilerPath: t.Vars.GetFirstOrEmpty("cpp.compiler"),
		cCompilerArgs:   foundation.NewArguments(64),
		cppCompilerArgs: foundation.NewArguments(64),
	}
}

func (cl *ToolchainArduinoEsp32Compiler) ObjFilepath(srcRelFilepath string) string {
	return srcRelFilepath + ".o"
}

func (cl *ToolchainArduinoEsp32Compiler) DepFilepath(objRelFilepath string) string {
	return objRelFilepath + ".d"
}

func (cl *ToolchainArduinoEsp32Compiler) SetupArgs(defines []string, includes []string) {
	// C Compiler Arguments
	{
		cl.cCompilerArgs.Clear()

		cl.cCompilerArgs.Add("-c")
		cl.cCompilerArgs.Add("-MMD")

		if responseFileFlags, ok := cl.toolChain.Vars.GetFirst("c.compiler.response.flags"); ok {
			cl.cCompilerArgs.Add("@" + responseFileFlags)
		}

		if switches, ok := cl.toolChain.Vars.Get("c.compiler.switches"); ok {
			cl.cCompilerArgs.Add(switches...)
		}

		if warningSwitches, ok := cl.toolChain.Vars.Get("c.compiler.warning.switches"); ok {
			cl.cCompilerArgs.Add(warningSwitches...)
		}

		// Compiler system defines (debug / release ?)
		if compilerDefines, ok := cl.toolChain.Vars.Get("c.compiler.defines"); ok {
			cl.cCompilerArgs.AddWithPrefix("-D", compilerDefines...)
		}

		// Compiler user defines
		cl.cCompilerArgs.AddWithPrefix("-D", defines...)

		if responseFileDefines, ok := cl.toolChain.Vars.GetFirst("c.compiler.response.defines"); ok {
			cl.cCompilerArgs.Add("@" + responseFileDefines)
		}

		// Compiler prefix include paths
		if compilerPrefixInclude, ok := cl.toolChain.Vars.GetFirst("c.compiler.system.prefix.include"); ok {
			cl.cCompilerArgs.Add("-iprefix")
			// Make sure the path ends with a /
			if !strings.HasSuffix(compilerPrefixInclude, "/") {
				cl.cCompilerArgs.Add(compilerPrefixInclude + "/")
			} else {
				cl.cCompilerArgs.Add(compilerPrefixInclude)
			}

			if responseFileIncludes, ok := cl.toolChain.Vars.GetFirst("c.compiler.response.includes"); ok {
				cl.cCompilerArgs.Add("@" + responseFileIncludes)
			}
		}

		// Compiler system include paths
		if systemIncludes, ok := cl.toolChain.Vars.Get("c.compiler.system.includes"); ok {
			cl.cCompilerArgs.AddWithPrefix("-I", systemIncludes...)
		}

		// User include paths
		cl.cCompilerArgs.AddWithPrefix("-I", includes...)
	}

	// C++ Compiler Arguments
	{
		cl.cppCompilerArgs.Clear()

		cl.cppCompilerArgs.Add("-c")
		cl.cppCompilerArgs.Add("-MMD")

		if cppResponseFileFlags, ok := cl.toolChain.Vars.GetFirst("cpp.compiler.response.flags"); ok {
			cl.cppCompilerArgs.Add("@" + cppResponseFileFlags)
		}

		if cppSwitches, ok := cl.toolChain.Vars.Get("cpp.compiler.switches"); ok {
			cl.cppCompilerArgs.Add(cppSwitches...)
		}

		if cppWarningSwitches, ok := cl.toolChain.Vars.Get("cpp.compiler.warning.switches"); ok {
			cl.cppCompilerArgs.Add(cppWarningSwitches...)
		}

		// Compiler system defines (debug / release ?)
		if compilerDefines, ok := cl.toolChain.Vars.Get("cpp.compiler.defines"); ok {
			cl.cppCompilerArgs.AddWithPrefix("-D", compilerDefines...)
		}

		// Compiler user defines
		cl.cppCompilerArgs.AddWithPrefix("-D", defines...)

		if responseFileDefines, ok := cl.toolChain.Vars.GetFirst("cpp.compiler.response.defines"); ok {
			cl.cppCompilerArgs.Add("@" + responseFileDefines)
		}

		// Compiler prefix include paths
		if compilerPrefixInclude, ok := cl.toolChain.Vars.GetFirst("cpp.compiler.system.prefix.include"); ok {
			cl.cppCompilerArgs.Add("-iprefix")
			// Make sure the path ends with a /
			if !strings.HasSuffix(compilerPrefixInclude, "/") {
				cl.cppCompilerArgs.Add(compilerPrefixInclude + "/")
			} else {
				cl.cppCompilerArgs.Add(compilerPrefixInclude)
			}

			if responseFileIncludes, ok := cl.toolChain.Vars.GetFirst("cpp.compiler.response.includes"); ok {
				cl.cppCompilerArgs.Add("@" + responseFileIncludes)
			}
		}

		// Compiler system include paths
		if systemIncludes, ok := cl.toolChain.Vars.Get("cpp.compiler.system.includes"); ok {
			cl.cppCompilerArgs.AddWithPrefix("-I", systemIncludes...)
		}

		// User include paths
		cl.cppCompilerArgs.AddWithPrefix("-I", includes...)
	}
}

func (cl *ToolchainArduinoEsp32Compiler) Compile(sourceAbsFilepaths []string, objRelFilepaths []string) error {

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

		foundation.LogInfof("Compiling (%s) %s\n", cl.config.Config.AsString(), filepath.Base(sourceAbsFilepath))

		cmd := exec.Command(compilerPath, args...)
		out, err := cmd.CombinedOutput()

		if err != nil {
			foundation.LogInfof("Compile failed, output:\n%s\n", string(out))
			return foundation.LogErrorf(err, "Compiling failed")
		}
		if len(out) > 0 {
			foundation.LogInfof("Compile output:\n%s\n", string(out))
		}
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
	archiverArgs *foundation.Arguments
}

func (t *ArduinoEsp32) NewArchiver(a ArchiverType, config *Config) Archiver {
	return &ToolchainArduinoEsp32Archiver{
		toolChain:    t,
		config:       config,
		archiverPath: t.Vars.GetFirstOrEmpty("archiver"),
		archiverArgs: foundation.NewArguments(16),
	}
}

func (t *ToolchainArduinoEsp32Archiver) LibFilepath(filepath string) string {
	// The file extension for the archive on ESP32 is typically ".a"
	return filepath + ".a"
}

func (a *ToolchainArduinoEsp32Archiver) SetupArgs() {

}

func (a *ToolchainArduinoEsp32Archiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {

	a.archiverArgs.Clear() // Reset the arguments
	a.archiverArgs.Add("cr")

	// {output-archive-filepath}
	a.archiverArgs.Add(outputArchiveFilepath)

	// {input-object-filepaths}
	a.archiverArgs.Add(inputObjectFilepaths...)

	foundation.LogInfof("Archiving %s\n", outputArchiveFilepath)

	cmd := exec.Command(a.archiverPath, a.archiverArgs.Args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return foundation.LogErrorf(err, "Archiving failed")
	}
	if len(out) > 0 {
		foundation.LogInfof("Archive output:\n%s\n", string(out))
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
	linkerArgs *foundation.Arguments
}

func (t *ArduinoEsp32) NewLinker(config *Config) Linker {
	return &ToolchainArduinoEsp32Linker{
		toolChain:  t,
		config:     config,
		linkerPath: t.Vars.GetFirstOrEmpty("linker"),
		linkerArgs: foundation.NewArguments(512),
	}
}

// FileExt returns the file extension for the linker output, which is ".elf" for ESP32.
func (l *ToolchainArduinoEsp32Linker) LinkedFilepath(filepath string) string {
	return filepath + ".elf"
}

func (l *ToolchainArduinoEsp32Linker) SetupArgs(libraryPaths []string, libraryFiles []string) {

	l.linkerArgs.Clear()

	if genMapfile, ok := l.toolChain.Vars.GetFirst("linker.generate.mapfile"); ok && genMapfile == "true" {
		l.linkerArgs.Add("genmap")
	}

	if linkerSystemLibraryPaths, ok := l.toolChain.Vars.Get("linker.system.library.paths"); ok {
		l.linkerArgs.AddWithPrefix("-L", linkerSystemLibraryPaths...)
	}

	// User library paths
	l.linkerArgs.AddWithPrefix("-L", libraryPaths...)

	l.linkerArgs.Add("-Wl,--wrap=esp_panic_handler")

	linkerResponseFile, _ := l.toolChain.Vars.GetFirst("linker.response.ldflags")
	if len(linkerResponseFile) > 0 {
		l.linkerArgs.Add("@" + linkerResponseFile)
	}

	linkerResponseFile, _ = l.toolChain.Vars.GetFirst("linker.response.ldscripts")
	if len(linkerResponseFile) > 0 {
		l.linkerArgs.Add("@" + linkerResponseFile)
	}

	l.linkerArgs.Add("-Wl,--start-group")
	{
		// User library files
		l.linkerArgs.Add(libraryFiles...)

		linkerResponseFile, _ = l.toolChain.Vars.GetFirst("linker.response.ldlibs")
		if len(linkerResponseFile) > 0 {
			l.linkerArgs.Add("@" + linkerResponseFile)
		}
	}
}

func (l *ToolchainArduinoEsp32Linker) Link(inputArchiveAbsFilepaths []string, outputAppRelFilepathNoExt string) error {
	linker := l.linkerPath

	linkerArgs := *l.linkerArgs

	linkerArgs.Add(inputArchiveAbsFilepaths...)
	linkerArgs.Add("-Wl,--end-group")
	linkerArgs.Add("-Wl,-EL") //

	linkerArgs.Add("-o")
	linkerArgs.Add(outputAppRelFilepathNoExt)

	// Do we need to fill in the arg to generate map file?
	generateMapfile := linkerArgs.Args[0] == "genmap"
	if generateMapfile {
		outputMapFilepath := outputAppRelFilepathNoExt + ".map"
		linkerArgs.Args[0] = "-Wl,--Map=" + outputMapFilepath
	}

	foundation.LogInfof("Linking '%s'...\n", outputAppRelFilepathNoExt)
	cmd := exec.Command(linker, linkerArgs.Args...)
	out, err := cmd.CombinedOutput()

	// Reset the map generation command in the arguments so that it
	// will be updated correctly on the next Link() call.
	if generateMapfile {
		l.linkerArgs.Args[0] = "genmap"
	}

	if err != nil {
		foundation.LogInfof("Link failed, output:\n%s\n", string(out))
		return foundation.LogErrorf(err, "Linking failed")
	}
	if len(out) > 0 {
		foundation.LogInfof("Link output:\n%s\n", string(out))
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Burner

type ToolchainArduinoEsp32Burner struct {
	toolChain                            *ArduinoEsp32
	config                               *Config              // Configuration for the burner, e.g., debug or release
	dependencyTracker                    deptrackr.FileTrackr // Dependency tracker for the burner
	hasher                               hash.Hash            // Hasher for generating digests of arguments
	genImageBinToolArgs                  *foundation.Arguments
	genImageBinToolArgsHash              []byte   // Hash of the arguments for the image bin tool
	genImageBinToolOutputFilepath        string   // The output file for the image bin
	genImageBinToolInputFilepaths        []string // The input files for the image bin
	genImageBinToolPath                  string
	genImagePartitionsToolArgs           *foundation.Arguments
	genImagePartitionsToolArgsHash       []byte   // Hash of the arguments for the image partitions tool
	genImagePartitionsToolOutputFilepath string   // The output file for the image partitions
	genImagePartitionsToolInputFilepaths []string // The input files for the image partitions
	genImagePartitionsToolPath           string
	genBootloaderToolArgs                *foundation.Arguments
	genBootloaderToolArgsHash            []byte   // Hash of the arguments for the bootloader tool
	genBootloaderToolOutputFilepath      string   // The output file for the bootloader
	genBootloaderToolInputFilepaths      []string // The input files for the bootloader
	genBootloaderToolPath                string
	flashToolArgs                        *foundation.Arguments
	flashToolPath                        string
}

func (t *ArduinoEsp32) NewBurner(config *Config) Burner {
	return &ToolchainArduinoEsp32Burner{
		toolChain:                            t,
		config:                               config,
		dependencyTracker:                    nil,
		hasher:                               sha1.New(),
		genImageBinToolArgs:                  foundation.NewArguments(64),
		genImageBinToolArgsHash:              nil, // Will be set later
		genImageBinToolOutputFilepath:        "",
		genImageBinToolInputFilepaths:        []string{},
		genImageBinToolPath:                  t.Vars.GetFirstOrEmpty("burner.generate-image-bin"),
		genImagePartitionsToolArgs:           foundation.NewArguments(64),
		genImagePartitionsToolArgsHash:       nil, // Will be set later
		genImagePartitionsToolOutputFilepath: "",
		genImagePartitionsToolInputFilepaths: []string{},
		genImagePartitionsToolPath:           t.Vars.GetFirstOrEmpty("burner.generate-partitions-bin"),
		genBootloaderToolArgs:                foundation.NewArguments(64),
		genBootloaderToolArgsHash:            nil, // Will be set later
		genBootloaderToolOutputFilepath:      "",
		genBootloaderToolInputFilepaths:      []string{},
		genBootloaderToolPath:                t.Vars.GetFirstOrEmpty("burner.generate-bootloader"),
		flashToolArgs:                        foundation.NewArguments(64),
		flashToolPath:                        t.Vars.GetFirstOrEmpty("burner.flash"),
	}
}

func (b *ToolchainArduinoEsp32Burner) hashArguments(args []string) []byte {
	b.hasher.Reset()
	for _, arg := range args {
		b.hasher.Write([]byte(arg))
	}
	return b.hasher.Sum(nil)
}

func (b *ToolchainArduinoEsp32Burner) SetupBuild(buildPath string) {

	projectElfFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".elf")
	projectBinFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".bin")
	projectPartitionsBinFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".partitions.bin")
	projectBootloaderBinFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".bootloader.bin")

	b.genImageBinToolArgs.Clear()
	b.genImageBinToolArgs.Add("--chip", b.toolChain.Vars.GetFirstOrEmpty("esp.mcu"))
	b.genImageBinToolArgs.Add("elf2image")
	b.genImageBinToolArgs.Add("--flash_mode", b.toolChain.Vars.GetFirstOrEmpty("burner.flash.mode"))
	b.genImageBinToolArgs.Add("--flash_freq", b.toolChain.Vars.GetFirstOrEmpty("burner.flash.frequency"))
	b.genImageBinToolArgs.Add("--flash_size", b.toolChain.Vars.GetFirstOrEmpty("burner.flash.size"))
	b.genImageBinToolArgs.Add("--elf-sha256-offset", b.toolChain.Vars.GetFirstOrEmpty("burner.flash.elf.share.offset"))
	b.genImageBinToolArgs.Add("-o", projectBinFilepath)
	b.genImageBinToolArgs.Add(projectElfFilepath)

	b.genImagePartitionsToolArgs.Clear()
	b.genImagePartitionsToolArgs.Add(b.toolChain.Vars.GetFirstOrEmpty("burner.generate-partitions-bin.script"))
	b.genImagePartitionsToolArgs.Add("-q")
	b.genImagePartitionsToolArgs.Add(b.toolChain.Vars.GetFirstOrEmpty("burner.flash.partitions.csv.filepath"))
	b.genImagePartitionsToolArgs.Add(projectPartitionsBinFilepath)

	b.genBootloaderToolArgs.Clear()
	b.genBootloaderToolArgs.Add("--chip", b.toolChain.Vars.GetFirstOrEmpty("esp.mcu"))
	b.genBootloaderToolArgs.Add("elf2image")
	b.genBootloaderToolArgs.Add("--flash_mode", b.toolChain.Vars.GetFirstOrEmpty("burner.flash.mode"))
	b.genBootloaderToolArgs.Add("--flash_freq", b.toolChain.Vars.GetFirstOrEmpty("burner.flash.frequency"))
	b.genBootloaderToolArgs.Add("--flash_size", b.toolChain.Vars.GetFirstOrEmpty("burner.flash.size"))
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

func (b *ToolchainArduinoEsp32Burner) Build() error {

	// - Generate image partitions bin file ('PROJECT.NAME.partitions.bin'):
	// - Generate image bin file ('PROJECT.NAME.bin'):
	// - Generate bootloader image ('PROJECT.NAME.bootloader.bin'):

	// Generate the image partitions bin file
	if !b.dependencyTracker.QueryItemWithExtraData(b.genImagePartitionsToolOutputFilepath, b.genImagePartitionsToolArgsHash) {
		img, _ := exec.LookPath(b.genImagePartitionsToolPath)
		args := b.genImagePartitionsToolArgs.Args

		cmd := exec.Command(img, args...)
		foundation.LogInfof("Creating image partitions '%s' ...\n", b.toolChain.ProjectName+".partitions.bin")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return foundation.LogErrorf(err, "Creating image partitions failed")
		}
		if len(out) > 0 {
			foundation.LogInfof("Image partitions output:\n%s\n", string(out))
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
		foundation.LogInfof("Generating image '%s'\n", b.toolChain.ProjectName+".bin")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return foundation.LogErrorf(err, "Image generation failed")
		}
		if len(out) > 0 {
			foundation.LogInfof("Image generation output:\n%s\n", string(out))
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
		foundation.LogInfof("Generating bootloader '%s'\n", b.toolChain.ProjectName+".bootloader.bin")
		out, err := cmd.CombinedOutput()
		if err != nil {
			foundation.LogInfof("Bootloader generation failed, output:\n%s\n", string(out))
			return foundation.LogErrorf(err, "Bootloader generation failed")
		}
		if len(out) > 0 {
			foundation.LogInfof("Bootloader generation output:\n%s\n", string(out))
		}
		fmt.Println()

		b.dependencyTracker.AddItemWithExtraData(b.genBootloaderToolOutputFilepath, b.genBootloaderToolArgsHash, b.genBootloaderToolInputFilepaths)
	} else {
		b.dependencyTracker.CopyItem(b.genBootloaderToolOutputFilepath)
	}

	b.dependencyTracker.Save()

	return nil
}

func (b *ToolchainArduinoEsp32Burner) SetupBurn(buildPath string) error {

	b.flashToolArgs.Clear()
	b.flashToolArgs.Add("--chip", b.toolChain.Vars.GetFirstOrEmpty("esp.mcu"))
	b.flashToolArgs.Add("--port", b.toolChain.Vars.GetFirstOrEmpty("burner.flash.port"))
	b.flashToolArgs.Add("--baud", b.toolChain.Vars.GetFirstOrEmpty("burner.flash.baud"))
	b.flashToolArgs.Add("--before", "default_reset")
	b.flashToolArgs.Add("--after", "hard_reset")
	b.flashToolArgs.Add("write_flash", "-z")
	b.flashToolArgs.Add("--flash_mode", "keep")
	b.flashToolArgs.Add("--flash_freq", "keep")
	b.flashToolArgs.Add("--flash_size", "keep")

	bootApp0BinFilePath, _ := b.toolChain.Vars.GetFirst("burner.bootapp0.bin.filepath")
	if !foundation.FileExists(bootApp0BinFilePath) {
		return foundation.LogErrorf(os.ErrNotExist, "Boot app0 bin file '%s' does not exist", bootApp0BinFilePath)
	}

	b.flashToolArgs.Add(b.toolChain.Vars.GetFirstOrEmpty("burner.flash.bootloader.bin.offset"))
	b.flashToolArgs.Add(b.genBootloaderToolOutputFilepath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetFirstOrEmpty("burner.flash.partitions.bin.offset"))
	b.flashToolArgs.Add(b.genImagePartitionsToolOutputFilepath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetFirstOrEmpty("burner.flash.bootapp0.bin.offset"))
	b.flashToolArgs.Add(bootApp0BinFilePath)
	b.flashToolArgs.Add(b.toolChain.Vars.GetFirstOrEmpty("burner.flash.application.bin.offset"))
	b.flashToolArgs.Add(b.genImageBinToolOutputFilepath)

	return nil
}

func (b *ToolchainArduinoEsp32Burner) Burn() error {
	if !foundation.FileExists(b.genBootloaderToolOutputFilepath) {
		return foundation.LogErrorf(os.ErrNotExist, "Cannot burn, bootloader bin file '%s' doesn't exist", b.genBootloaderToolOutputFilepath)
	}
	if !foundation.FileExists(b.genImagePartitionsToolOutputFilepath) {
		return foundation.LogErrorf(os.ErrNotExist, "Cannot burn, partitions bin file '%s' doesn't exist", b.genImagePartitionsToolOutputFilepath)
	}
	if !foundation.FileExists(b.genImageBinToolOutputFilepath) {
		return foundation.LogErrorf(os.ErrNotExist, "Cannot burn, application bin file '%s' doesn't exist", b.genImageBinToolOutputFilepath)
	}

	flashToolPath := b.flashToolPath
	flashToolArgs := b.flashToolArgs.Args

	foundation.LogInfof("Flashing '%s'...\n", b.toolChain.ProjectName+".bin")

	flashToolCmd := exec.Command(flashToolPath, flashToolArgs...)

	// out, err := flashToolCmd.CombinedOutput()
	// if err != nil {
	// 	return foundation.LogErrorf("Flashing failed with %s\n", err)
	// }
	// if len(out) > 0 {
	// 	foundation.LogInfof("Flashing output:\n%s\n", string(out))
	// }

	pipe, _ := flashToolCmd.StdoutPipe()

	if err := flashToolCmd.Start(); err != nil {
		return foundation.LogErrorf(err, "Flashing failed")
	}

	reader := bufio.NewReader(pipe)
	line, err := reader.ReadString('\n')
	for err == nil {
		foundation.LogPrint(line)
		line, err = reader.ReadString('\n')
		if err == io.EOF {
			err = nil
			break
		}
	}
	foundation.LogPrintln()

	if err != nil {
		return foundation.LogErrorf(err, "Flashing failed")
	}

	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Dependency Tracker
func (t *ArduinoEsp32) NewDependencyTracker(dirpath string) deptrackr.FileTrackr {
	return deptrackr.LoadDepFileTrackr(filepath.Join(dirpath, "deptrackr"))
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Toolchain for ESP32 on Arduino
func NewArduinoEsp32(espMcu string, projectName string) (*ArduinoEsp32, error) {

	var espSdkPath string
	if espSdkPath = os.Getenv("ESP_SDK"); espSdkPath == "" {
		return nil, foundation.LogErrorf(os.ErrNotExist, "ESP_SDK environment variable is not set, please set it to the path of the ESP32 SDK")
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

		{key: "linker.generate.mapfile", value: []string{`true`}},
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
		Vars:        foundation.NewVarsCustom(foundation.VarsFormatCurlyBraces),
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
		return nil, foundation.LogErrorf(os.ErrInvalid, "unsupported ESP32 MCU: %s", espMcu)
	}

	t.Vars.Resolve()
	return t, nil
}
