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

type ArduinoEsp32Toolchainv2 struct {
	ProjectName string // The name of the project, used for output files
	Vars        *corepkg.Vars
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Compiler

type ToolchainArduinoEsp32Compilerv2 struct {
	toolChain   *ArduinoEsp32Toolchainv2
	buildConfig denv.BuildConfig // Configuration for the compiler, e.g., debug or release
	buildTarget denv.BuildTarget
	vars        *corepkg.Vars // Local variables for the compiler
}

func (t *ArduinoEsp32Toolchainv2) NewCompiler(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Compiler {
	return &ToolchainArduinoEsp32Compilerv2{
		toolChain:   t,
		buildConfig: buildConfig,
		buildTarget: buildTarget,
		vars:        corepkg.NewVars(16),
	}
}

func (cl *ToolchainArduinoEsp32Compilerv2) ObjFilepath(srcRelFilepath string) string {
	return srcRelFilepath + ".o"
}

func (cl *ToolchainArduinoEsp32Compilerv2) DepFilepath(objRelFilepath string) string {
	return objRelFilepath + ".d"
}

func (cl *ToolchainArduinoEsp32Compilerv2) SetupArgs(defines []string, includes []string) {
	for i, inc := range includes {
		includes[i] = "-I" + inc
	}
	for i, def := range defines {
		if !strings.HasPrefix(def, "-D") {
			defines[i] = "-D" + def
		}
	}
    cl.toolChain.Vars.Append("build.defines", defines...)

	cl.vars.Clear()
	cl.vars.Append("includes", "-I{runtime.platform.path}/variants/{board.name}")
	cl.vars.Append("includes", includes...)
}

func (cl *ToolchainArduinoEsp32Compilerv2) Compile(sourceAbsFilepaths []string, objRelFilepaths []string) error {
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

type ToolchainArduinoEsp32Archiverv2 struct {
	toolChain   *ArduinoEsp32Toolchainv2
	buildConfig denv.BuildConfig
	buildTarget denv.BuildTarget
	vars        *corepkg.Vars
}

func (t *ArduinoEsp32Toolchainv2) NewArchiver(a ArchiverType, buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Archiver {
	return &ToolchainArduinoEsp32Archiverv2{
		toolChain:   t,
		buildConfig: buildConfig,
		buildTarget: buildTarget,
		vars:        corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
	}
}

func (t *ToolchainArduinoEsp32Archiverv2) LibFilepath(filepath string) string {
	// The file extension for an archive/library on ESP32 is typically ".a"
	return filepath + ".a"
}

func (a *ToolchainArduinoEsp32Archiverv2) SetupArgs() {
}

func (a *ToolchainArduinoEsp32Archiverv2) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {
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

type ToolchainArduinoEsp32Linkerv2 struct {
	toolChain   *ArduinoEsp32Toolchainv2
	buildConfig denv.BuildConfig
	buildTarget denv.BuildTarget
	vars        *corepkg.Vars // Local variables for the linker
}

func (t *ArduinoEsp32Toolchainv2) NewLinker(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Linker {
	return &ToolchainArduinoEsp32Linkerv2{
		toolChain:   t,
		buildConfig: buildConfig,
		buildTarget: buildTarget,
		vars:        corepkg.NewVars(16),
	}
}

// FileExt returns the file extension for the linker output, which is ".elf" for ESP32.
func (l *ToolchainArduinoEsp32Linkerv2) LinkedFilepath(filepath string) string {
	return filepath + ".elf"
}

func (l *ToolchainArduinoEsp32Linkerv2) SetupArgs(libraryPaths []string, libraryFiles []string) {
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

func (l *ToolchainArduinoEsp32Linkerv2) Link(inputObjectsAbsFilepaths, inputArchivesAbsFilepaths []string, outputAppRelFilepathNoExt string) error {
	corepkg.LogInfof("Linking '%s'...", outputAppRelFilepathNoExt)

	linkerArgs, _ := l.toolChain.Vars.Get(`recipe.c.combine.pattern`)

	linkerPath := linkerArgs[0]
	linkerArgs = linkerArgs[1:]

	l.vars.Set("object_files", inputObjectsAbsFilepaths...)
	l.vars.Append("object_files", inputArchivesAbsFilepaths...)

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

type ToolchainArduinoEsp32Burnerv2 struct {
	toolChain                            *ArduinoEsp32Toolchainv2
	buildConfig                          denv.BuildConfig
	buildTarget                          denv.BuildTarget
	dependencyTracker                    deptrackr.FileTrackr // Dependency tracker for the burner
	hasher                               hash.Hash            // Hasher for generating digests of arguments
	genImageBinToolArgs                  []string
	genImageBinToolArgsHash              []byte   // Hash of the arguments for the image bin tool
	genImageBinToolOutputFilepath        string   // The output file for the image bin
	genImageBinToolInputFilepaths        []string // The input files for the image bin
	genImageBinToolPath                  string
	genImagePartitionsToolArgs           []string
	genImagePartitionsToolArgsHash       []byte   // Hash of the arguments for the image partitions tool
	genImagePartitionsToolOutputFilepath string   // The output file for the image partitions
	genImagePartitionsToolInputFilepaths []string // The input files for the image partitions
	genImagePartitionsToolPath           string
	genBootloaderToolArgs                []string
	genBootloaderToolArgsHash            []byte   // Hash of the arguments for the bootloader tool
	genBootloaderToolOutputFilepath      string   // The output file for the bootloader
	genBootloaderToolInputFilepaths      []string // The input files for the bootloader
	genBootloaderToolPath                string
	flashToolArgs                        []string
	flashToolPath                        string
	vars                                 *corepkg.Vars // Local variables for the burner
}

func (t *ArduinoEsp32Toolchainv2) NewBurner(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) Burner {
	return &ToolchainArduinoEsp32Burnerv2{
		toolChain:                            t,
		buildConfig:                          buildConfig,
		buildTarget:                          buildTarget,
		dependencyTracker:                    nil,
		hasher:                               sha1.New(),
		genImageBinToolArgs:                  make([]string, 0, 32),
		genImageBinToolArgsHash:              nil, // Will be set later
		genImageBinToolOutputFilepath:        "",
		genImageBinToolInputFilepaths:        []string{},
		genImageBinToolPath:                  "",
		genImagePartitionsToolArgs:           make([]string, 0, 32),
		genImagePartitionsToolArgsHash:       nil, // Will be set later
		genImagePartitionsToolOutputFilepath: "",
		genImagePartitionsToolInputFilepaths: []string{},
		genImagePartitionsToolPath:           "",
		genBootloaderToolArgs:                make([]string, 0, 32),
		genBootloaderToolArgsHash:            nil, // Will be set later
		genBootloaderToolOutputFilepath:      "",
		genBootloaderToolInputFilepaths:      []string{},
		genBootloaderToolPath:                "",
		flashToolArgs:                        make([]string, 0, 32),
		flashToolPath:                        "",
		vars:                                 corepkg.NewVars(16),
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
	// File Dependency Tracker and Information
	b.dependencyTracker = deptrackr.LoadDepFileTrackr(filepath.Join(buildPath, "deptrackr.burn"))

	b.vars.Set("build.project_name", b.toolChain.ProjectName)
	b.vars.Set("build.source.path", filepath.Join(buildPath, "void"))
	b.vars.Set("build.variant.path", filepath.Join(buildPath, "void"))

	// ------------------------------------------------------------------------------------------------
	// Image partitions tool setup
	dstPartitionsFilepath := filepath.Join(buildPath, "partitions.csv")
	if !b.dependencyTracker.QueryItem(dstPartitionsFilepath) {
		// Create a 'partitions.csv' file in the build path if it doesn't exist, take the one from
		//    {runtime.platform.path}/tools/partitions/{build.partitions}.csv
		srcPartitionsFilepath := b.toolChain.Vars.FinalResolveString("{runtime.platform.path}/tools/partitions/{build.partitions}.csv", " ", b.vars)
		input, err := os.ReadFile(srcPartitionsFilepath)
		if err == nil {
			err = os.WriteFile(filepath.Join(buildPath, "partitions.csv"), input, 0644)
			if err != nil {
				corepkg.LogErrorf(err, "Failed to create partitions file in build path")
			}
		} else {
			corepkg.LogErrorf(err, "Failed to read partitions file from platform path")
		}

		b.dependencyTracker.AddItem(dstPartitionsFilepath, []string{srcPartitionsFilepath})
	} else {
		b.dependencyTracker.CopyItem(dstPartitionsFilepath)
	}

	genImagePartitionsArgs, _ := b.toolChain.Vars.Get("recipe.objcopy.partitions.bin.pattern")
	genImagePartitionsArgs = b.toolChain.Vars.FinalResolveArray(genImagePartitionsArgs, b.vars)

	b.genImagePartitionsToolPath = genImagePartitionsArgs[0]
	b.genImagePartitionsToolArgs = genImagePartitionsArgs[1:]
	b.genImagePartitionsToolOutputFilepath = filepath.Join(buildPath, b.toolChain.ProjectName+".partitions.bin")
	b.genImagePartitionsToolInputFilepaths = []string{}

	// Remove any empty entries from genImagePartitionsArgs
	b.genImagePartitionsToolArgs = slices.DeleteFunc(b.genImagePartitionsToolArgs, func(s string) bool { return strings.TrimSpace(s) == "" })
	b.genImagePartitionsToolArgsHash = b.hashArguments(b.genImagePartitionsToolArgs)

	// ------------------------------------------------------------------------------------------------
	// Image bin tool setup

	genImageBinArgs, _ := b.toolChain.Vars.Get("recipe.objcopy.bin.pattern")
	genImageBinArgs = b.toolChain.Vars.FinalResolveArray(genImageBinArgs, b.vars)

	b.genImageBinToolPath = genImageBinArgs[0]
	b.genImageBinToolArgs = genImageBinArgs[1:]
	b.genImageBinToolOutputFilepath = filepath.Join(buildPath, b.toolChain.ProjectName+".bin")
	b.genImageBinToolInputFilepaths = []string{filepath.Join(buildPath, b.toolChain.ProjectName+".elf")}

	// Remove any empty entries from genImageBinArgs
	b.genImageBinToolArgs = slices.DeleteFunc(b.genImageBinToolArgs, func(s string) bool { return strings.TrimSpace(s) == "" })
	b.genImageBinToolArgsHash = b.hashArguments(b.genImageBinToolArgs)

	// ------------------------------------------------------------------------------------------------
	// Bootloader tool setup
	genBootloaderToolArgs, _ := b.toolChain.Vars.Get("recipe.hooks.prebuild.4.pattern")
	genBootloaderToolArgs = b.toolChain.Vars.FinalResolveArray(genBootloaderToolArgs, b.vars)

	b.genBootloaderToolPath = genBootloaderToolArgs[0]
	b.genBootloaderToolArgs = genBootloaderToolArgs[1:]
	b.genBootloaderToolOutputFilepath = filepath.Join(buildPath, b.toolChain.ProjectName+".bootloader.bin")
	b.genBootloaderToolInputFilepaths = []string{}

	// Remove any empty entries from genBootloaderToolArgs
	b.genBootloaderToolArgs = slices.DeleteFunc(b.genBootloaderToolArgs, func(s string) bool { return strings.TrimSpace(s) == "" })
	b.genBootloaderToolArgsHash = b.hashArguments(b.genBootloaderToolArgs)
}

func (b *ToolchainArduinoEsp32Burnerv2) Build() error {

	// - Generate image partitions bin file ('PROJECT.NAME.partitions.bin')
	// - Generate image bin file ('PROJECT.NAME.bin')
	// - Generate bootloader image ('PROJECT.NAME.bootloader.bin')

	// Generate the image partitions bin file
	if !b.dependencyTracker.QueryItemWithExtraData(b.genImagePartitionsToolOutputFilepath, b.genImagePartitionsToolArgsHash) {

		img, _ := exec.LookPath(b.genImagePartitionsToolPath)
		args := b.genImagePartitionsToolArgs

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

	// Generate the bootloader image
	if !b.dependencyTracker.QueryItemWithExtraData(b.genBootloaderToolOutputFilepath, b.genBootloaderToolArgsHash) {
		imgPath := b.genBootloaderToolPath
		args := b.genBootloaderToolArgs

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
	if !corepkg.FileExists(b.genBootloaderToolOutputFilepath) {
		return corepkg.LogErrorf(os.ErrNotExist, "Cannot burn, bootloader bin file '%s' doesn't exist", b.genBootloaderToolOutputFilepath)
	}
	if !corepkg.FileExists(b.genImagePartitionsToolOutputFilepath) {
		return corepkg.LogErrorf(os.ErrNotExist, "Cannot burn, partitions bin file '%s' doesn't exist", b.genImagePartitionsToolOutputFilepath)
	}
	if !corepkg.FileExists(b.genImageBinToolOutputFilepath) {
		return corepkg.LogErrorf(os.ErrNotExist, "Cannot burn, application bin file '%s' doesn't exist", b.genImageBinToolOutputFilepath)
	}

	// ------------------------------------------------------------------------------------------------
	// Flash tool setup
	flashToolArgs, _ := b.toolChain.Vars.Get("tools.esptool_py.upload.pattern")

	b.vars.Set("cmd", "{tools.esptool_py.cmd}")
	b.vars.Set("path", "{tools.esptool_py.path}")

	b.vars.Set("upload.network_pattern", "{tools.esptool_py.upload.network_pattern}")
	b.vars.Set("upload.params.quiet", "{tools.esptool_py.upload.params.quiet}")
	b.vars.Set("upload.params", "{tools.esptool_py.upload.params.verbose}")
	b.vars.Set("upload.pattern_args", "{tools.esptool_py.upload.pattern_args}")
	b.vars.Set("upload.protocol", "{tools.esptool_py.upload.protocol}")

	flashToolArgs = b.toolChain.Vars.FinalResolveArray(flashToolArgs, b.vars)

	b.flashToolPath = flashToolArgs[0]
	b.flashToolArgs = flashToolArgs[1:]

	// Remove any empty entries from flashToolArgs
	b.flashToolArgs = slices.DeleteFunc(b.flashToolArgs, func(s string) bool { return strings.TrimSpace(s) == "" })

	return nil
}

func (b *ToolchainArduinoEsp32Burnerv2) Burn() error {
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
func (t *ArduinoEsp32Toolchainv2) NewDependencyTracker(dirpath string) deptrackr.FileTrackr {
	return deptrackr.LoadDepFileTrackr(filepath.Join(dirpath, "deptrackr"))
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------

func NewArduinoEsp32Toolchainv2(boardVars *corepkg.Vars, projectName string, buildPath string) *ArduinoEsp32Toolchainv2 {
	tc := &ArduinoEsp32Toolchainv2{ProjectName: projectName, Vars: boardVars}

	boardVars.Set("project.name", projectName)
	boardVars.Set("build.path", buildPath)
	boardVars.Set("build.arch", "ESP32")
	boardVars.Set("build.includes", "{runtime.platform.path}/variants/{board.name}")

	boardVars.SortByKey()

	// Create '{buildPath}/build.opt'
	// Create '{buildPath}/file.opt'

	// TODO not sure what these files are for and who should create it with what content

	buildOptFilePath := filepath.Join(buildPath, "build_opt.h")
	os.MkdirAll(filepath.Dir(buildOptFilePath), os.ModePerm)
	f, err := os.Create(buildOptFilePath)
	if err == nil {
		defer f.Close()
	}

	fileOptFilePath := filepath.Join(buildPath, "file_opts")
	os.MkdirAll(filepath.Dir(fileOptFilePath), os.ModePerm)
	f, err = os.Create(fileOptFilePath)
	if err == nil {
		defer f.Close()
	}

	return tc
}
