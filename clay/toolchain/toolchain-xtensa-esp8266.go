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
	cl.vars.Append("build.extra_flags", defines...)
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

		// corepkg.LogInfof("Using compiler: %s", compilerPath)
		// corepkg.LogInfof("Compiler args: %s", strings.Join(compilerArgs, " "))

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
	toolChain                     *ArduinoEsp8266Toolchain
	buildConfig                   denv.BuildConfig
	buildTarget                   denv.BuildTarget
	dependencyTracker             deptrackr.FileTrackr // Dependency tracker for the burner
	hasher                        hash.Hash            // Hasher for generating digests of arguments
	genImageBinToolArgs           []string
	genImageBinToolArgsHash       []byte   // Hash of the arguments for the image bin tool
	genImageBinToolOutputFilepath string   // The output file for the image bin
	genImageBinToolInputFilepaths []string // The input files for the image bin
	genImageBinToolPath           string
	flashToolArgs                 []string
	flashToolPath                 string
	vars                          *corepkg.Vars
}

func (t *ArduinoEsp8266Toolchain) NewBurner(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Burner {
	return &ToolchainArduinoEsp8266Burner{
		toolChain:                     t,
		buildConfig:                   buildConfig,
		buildTarget:                   buildTarget,
		dependencyTracker:             nil,
		hasher:                        sha1.New(),
		genImageBinToolArgs:           make([]string, 0, 16),
		genImageBinToolArgsHash:       nil, // Will be set later
		genImageBinToolOutputFilepath: "",
		genImageBinToolInputFilepaths: []string{},
		genImageBinToolPath:           "",
		flashToolArgs:                 make([]string, 0, 16),
		flashToolPath:                 "",
		vars:                          corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
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
	b.vars.Set("build.project_name", b.toolChain.ProjectName)

	projectElfFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".elf")
	projectBinFilepath := filepath.Join(buildPath, b.toolChain.ProjectName+".bin")

	// The XIAO_ESP32C3 board is configured as "qio" flash mode, 80MHz flash frequency and 4MB flash size.
	// However, the flash that is on the board can only be successfully flashed when using "dio" flash mode.
	if b.toolChain.Vars.GetFirstOrEmpty("build.mcu") == "esp32c3" {
		corepkg.LogWarnf("Overriding flash mode to 'dio' ESP32C3 boards")
		b.toolChain.Vars.Set("build.flash_mode", "dio")
	}

	// File Dependency Tracker and Information
	b.dependencyTracker = deptrackr.LoadDepFileTrackr(filepath.Join(buildPath, "deptrackr.burn"))

	b.genImageBinToolOutputFilepath = projectBinFilepath
	b.genImageBinToolInputFilepaths = []string{projectElfFilepath}
	genImageBinToolArgs, _ := b.toolChain.Vars.Get("recipe.objcopy.hex.1.pattern")
	b.genImageBinToolPath = genImageBinToolArgs[0]
	b.genImageBinToolArgs = genImageBinToolArgs[1:]
	b.genImageBinToolPath = b.toolChain.Vars.FinalResolveString(b.genImageBinToolPath, " ", nil)
	b.genImageBinToolArgs = b.toolChain.Vars.FinalResolveArray(b.genImageBinToolArgs, nil)
	b.genImageBinToolArgsHash = b.hashArguments(b.genImageBinToolArgs)
}

func (b *ToolchainArduinoEsp8266Burner) Build() error {

	// - Generate image bin file ('PROJECT.NAME.bin'):

	// Generate the image bin file
	if !b.dependencyTracker.QueryItemWithExtraData(b.genImageBinToolOutputFilepath, b.genImageBinToolArgsHash) {
		imgPath := b.genImageBinToolPath
		args := b.genImageBinToolArgs

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

	b.dependencyTracker.Save()

	return nil
}

func (b *ToolchainArduinoEsp8266Burner) SetupBurn(buildPath string) error {
	if !corepkg.FileExists(b.genImageBinToolOutputFilepath) {
		return corepkg.LogErrorf(os.ErrNotExist, "Cannot burn, application bin file '%s' doesn't exist", b.genImageBinToolOutputFilepath)
	}

	b.vars.Set("cmd", "{tools.esptool.cmd}")

	flashToolArgs, _ := b.toolChain.Vars.Get("tools.esptool.upload.pattern")
	b.flashToolPath = flashToolArgs[0]
	b.flashToolArgs = flashToolArgs[1:]
	b.flashToolPath = b.toolChain.Vars.FinalResolveString(b.flashToolPath, " ", nil)
	b.flashToolArgs = b.toolChain.Vars.FinalResolveArray(b.flashToolArgs, nil)

	return nil
}

func (b *ToolchainArduinoEsp8266Burner) Burn() error {
	flashToolPath := b.flashToolPath
	flashToolArgs := b.flashToolArgs

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
