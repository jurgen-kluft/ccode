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

type ArduinoEsp32Toolchainv2 struct {
	ProjectName string // The name of the project, used for output files
	Vars        *corepkg.Vars
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Compiler

type ToolchainArduinoEsp32Compilerv2 struct {
	toolChain       *ArduinoEsp32Toolchainv2
	buildConfig     denv.BuildConfig // Configuration for the compiler, e.g., debug or release
	buildTarget     denv.BuildTarget
	cCompilerPath   string
	cppCompilerPath string
	cCompilerArgs   *corepkg.Arguments
	cppCompilerArgs *corepkg.Arguments
}

func (t *ArduinoEsp32Toolchainv2) NewCompiler(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Compiler {
	return &ToolchainArduinoEsp32Compilerv2{
		toolChain:       t,
		buildConfig:     buildConfig,
		buildTarget:     buildTarget,
		cCompilerPath:   t.Vars.GetFirstOrEmpty("c.compiler"),
		cppCompilerPath: t.Vars.GetFirstOrEmpty("cpp.compiler"),
		cCompilerArgs:   corepkg.NewArguments(64),
		cppCompilerArgs: corepkg.NewArguments(64),
	}
}

func (cl *ToolchainArduinoEsp32Compilerv2) ObjFilepath(srcRelFilepath string) string {
	return srcRelFilepath + ".o"
}

func (cl *ToolchainArduinoEsp32Compilerv2) DepFilepath(objRelFilepath string) string {
	return objRelFilepath + ".d"
}

func (cl *ToolchainArduinoEsp32Compilerv2) SetupArgs(defines []string, includes []string) {
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

		// Optimization
		if cl.buildConfig.IsDebug() {
			cl.cCompilerArgs.Add("-Os")
		} else {
			cl.cCompilerArgs.Add("-Og", "-g3")
		}

		// Compiler system defines (debug / release ?)
		if compilerDefines, ok := cl.toolChain.Vars.Get("c.compiler.defines"); ok {
			cl.cCompilerArgs.AddWithPrefix("-D", compilerDefines...)
		}

		// Compiler user defines
		cl.cCompilerArgs.AddWithPrefix("-D", defines...)
		// Depending on Debug/Release/Final, we add CORE_DEBUG_LEVEL
		if cl.buildConfig.IsDebug() {
			cl.cCompilerArgs.AddWithPrefix("-D", "CORE_DEBUG_LEVEL=1")
		} else {
			cl.cCompilerArgs.AddWithPrefix("-D", "CORE_DEBUG_LEVEL=0")
		}

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

		// Optimization
		if cl.buildConfig.IsDebug() {
			cl.cCompilerArgs.Add("-Os")
		} else {
			cl.cCompilerArgs.Add("-Og", "-g3")
		}

		// Compiler system defines (debug / release ?)
		if compilerDefines, ok := cl.toolChain.Vars.Get("cpp.compiler.defines"); ok {
			cl.cppCompilerArgs.AddWithPrefix("-D", compilerDefines...)
		}

		// Compiler user defines
		cl.cppCompilerArgs.AddWithPrefix("-D", defines...)

		// Depending on Debug/Release/Final, we add CORE_DEBUG_LEVEL
		if cl.buildConfig.IsDebug() {
			cl.cCompilerArgs.AddWithPrefix("-D", "CORE_DEBUG_LEVEL=1")
		} else {
			cl.cCompilerArgs.AddWithPrefix("-D", "CORE_DEBUG_LEVEL=0")
		}

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

func (cl *ToolchainArduinoEsp32Compilerv2) Compile(sourceAbsFilepaths []string, objRelFilepaths []string) error {

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

type ToolchainArduinoEsp32Archiverv2 struct {
	toolChain    *ArduinoEsp32Toolchainv2
	buildConfig  denv.BuildConfig
	buildTarget  denv.BuildTarget
	archiverPath string
	archiverArgs *corepkg.Arguments
}

func (t *ArduinoEsp32Toolchainv2) NewArchiver(a ArchiverType, buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Archiver {
	return &ToolchainArduinoEsp32Archiverv2{
		toolChain:    t,
		buildConfig:  buildConfig,
		buildTarget:  buildTarget,
		archiverPath: t.Vars.GetFirstOrEmpty("archiver"),
		archiverArgs: corepkg.NewArguments(16),
	}
}

func (t *ToolchainArduinoEsp32Archiverv2) LibFilepath(filepath string) string {
	// The file extension for the archive on ESP32 is typically ".a"
	return filepath + ".a"
}

func (a *ToolchainArduinoEsp32Archiverv2) SetupArgs() {

}

func (a *ToolchainArduinoEsp32Archiverv2) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {

	a.archiverArgs.Clear() // Reset the arguments
	a.archiverArgs.Add("cr")

	// {output-archive-filepath}
	a.archiverArgs.Add(outputArchiveFilepath)

	// {input-object-filepaths}
	a.archiverArgs.Add(inputObjectFilepaths...)

	corepkg.LogInfof("Archiving %s", outputArchiveFilepath)

	cmd := exec.Command(a.archiverPath, a.archiverArgs.Args...)
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

type ToolchainArduinoEsp32Linkerv2 struct {
	toolChain   *ArduinoEsp32Toolchainv2
	buildConfig denv.BuildConfig
	buildTarget denv.BuildTarget
	linkerPath  string
	linkerArgs  *corepkg.Arguments
}

func (t *ArduinoEsp32Toolchainv2) NewLinker(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Linker {
	return &ToolchainArduinoEsp32Linkerv2{
		toolChain:   t,
		buildConfig: buildConfig,
		buildTarget: buildTarget,
		linkerPath:  t.Vars.GetFirstOrEmpty("linker"),
		linkerArgs:  corepkg.NewArguments(512),
	}
}

// FileExt returns the file extension for the linker output, which is ".elf" for ESP32.
func (l *ToolchainArduinoEsp32Linkerv2) LinkedFilepath(filepath string) string {
	return filepath + ".elf"
}

func (l *ToolchainArduinoEsp32Linkerv2) SetupArgs(libraryPaths []string, libraryFiles []string) {

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

func (l *ToolchainArduinoEsp32Linkerv2) Link(inputArchiveAbsFilepaths []string, outputAppRelFilepathNoExt string) error {
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

type ToolchainArduinoEsp32Burnerv2 struct {
	toolChain                            *ArduinoEsp32Toolchainv2
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

func (t *ArduinoEsp32Toolchainv2) NewBurner(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Burner {
	return &ToolchainArduinoEsp32Burnerv2{
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

func (b *ToolchainArduinoEsp32Burnerv2) hashArguments(args []string) []byte {
	b.hasher.Reset()
	for _, arg := range args {
		b.hasher.Write([]byte(arg))
	}
	return b.hasher.Sum(nil)
}

func (b *ToolchainArduinoEsp32Burnerv2) SetupBuild(buildPath string) {

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

func (b *ToolchainArduinoEsp32Burnerv2) Build() error {

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

func (b *ToolchainArduinoEsp32Burnerv2) SetupBurn(buildPath string) error {

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

func (b *ToolchainArduinoEsp32Burnerv2) Burn() error {
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
func (t *ArduinoEsp32Toolchainv2) NewDependencyTracker(dirpath string) deptrackr.FileTrackr {
	return deptrackr.LoadDepFileTrackr(filepath.Join(dirpath, "deptrackr"))
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------

func NewArduinoEsp32Toolchainv2(boardVars *corepkg.Vars, projectName string, buildPath string) *ArduinoEsp32Toolchainv2 {
	tc := &ArduinoEsp32Toolchainv2{ProjectName: projectName, Vars: boardVars}

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
