package toolchain

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/csv"
	"fmt"
	"hash"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
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
// Partition Scheme Generator

type Partition struct {
	Name    string
	Type    string
	SubType string
	Offset  string
	Size    string
	Flags   string
}

// filepath = Output CSV file path
// ota = ( >= 0) Enable OTA partitions, ( < 0) Enable factory partition with size = -ota
// nvsSize = Size of NVS partition (e.g., 24K, 0x6000)
// fsType = Filesystem type (spiffs|littlefs|fatfs)
// fsSize = Size of filesystem partition (e.g., 1M)")
// coreDump = Enable core dump partition
func alignInt64(offset int64, alignment int64) int64 {
	return (offset + (alignment - 1)) & ^(alignment - 1)
}

func generatePartitions(filepath string, ota bool, otaSize int64, nvsSize int64, fsType string, fsSize int64, coreDump bool) error {
	partitions := []Partition{}
	offset := int64(0x9000)     // Start offset
	align4KB := int64(0x1000)   // 4KB align4KB
	align64KB := int64(0x10000) // 64KB align4KB

	if nvsSize >= 0 { // NVS
		nvsSize = alignInt64(nvsSize, align4KB)
		partitions = append(partitions, Partition{"nvs", "data", "nvs", fmt.Sprintf("0x%X", offset), fmt.Sprintf("0x%X", nvsSize), ""})
		offset = alignInt64(offset+nvsSize, align4KB)
	}

	// Note: There seems to be another unwritten rule and that is that the filesystem partitions should always come after
	// the application partitions. So we first write the app partitions and then the FS partition.

	if ota { // OTA
		otaSize := alignInt64(otaSize, align64KB)

		otaDataSize := int64(0x2000)
		partitions = append(partitions, Partition{"otadata", "data", "ota", fmt.Sprintf("0x%X", offset), fmt.Sprintf("0x%X", otaDataSize), ""})

		offset = alignInt64(offset+otaDataSize, align64KB)
		partitions = append(partitions, Partition{"ota_0", "app", "ota_0", fmt.Sprintf("0x%X", offset), fmt.Sprintf("0x%X", otaSize), ""})

		offset = alignInt64(offset+otaSize, align64KB)
		partitions = append(partitions, Partition{"ota_1", "app", "ota_1", fmt.Sprintf("0x%X", offset), fmt.Sprintf("0x%X", otaSize), ""})

		offset = alignInt64(offset+otaSize, align64KB)
	} else { // Factory app, no OTA
		offset = alignInt64(offset, align64KB)
		factorySize := alignInt64(otaSize, align4KB)
		partitions = append(partitions, Partition{"factory", "app", "factory", fmt.Sprintf("0x%X", offset), fmt.Sprintf("0x%X", factorySize), ""})
		offset = alignInt64(offset+factorySize, align4KB)
	}

	if fsSize >= 0 { // FS partition
		fsSize = alignInt64(fsSize, align4KB)
		partitions = append(partitions, Partition{fsType, "data", fsType, fmt.Sprintf("0x%X", offset), fmt.Sprintf("0x%X", fsSize), ""})
		offset = alignInt64(offset+fsSize, align4KB)
	}

	if coreDump { // Crash dump partition
		coreDumpSize := int64(0x10000)
		partitions = append(partitions, Partition{"coredump", "data", "coredump", fmt.Sprintf("0x%X", offset), fmt.Sprintf("0x%X", coreDumpSize), ""})
		offset = alignInt64(offset+coreDumpSize, align4KB)
	}

	// Write CSV
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create partitions CSV: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"# Name", "Type", "SubType", "Offset", "Size", "Flags"})
	for _, p := range partitions {
		writer.Write([]string{p.Name, p.Type, p.SubType, p.Offset, p.Size, p.Flags})
	}

	fmt.Println("Partition CSV generated: " + filepath)
	return nil
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
		compilerArgs = slices.DeleteFunc(compilerArgs, func(s string) bool { return strings.TrimSpace(s) == "" })

		//fmt.Printf("Compiler path %s and args %v\n", compilerPath, compilerArgs)

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
	partitionsFilepath                   string
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
	elfSizeToolArgs                      []string
	elfSizeToolPath                      string
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
		partitionsFilepath:                   "",
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
		elfSizeToolArgs:                      make([]string, 0, 32),
		elfSizeToolPath:                      "",
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

	if b.toolChain.Vars.GetFirstOrEmpty("build.mcu") == "esp32c3" {
		corepkg.LogWarnf("Overriding flash mode to 'dio' for XIAO_ESP32C3 board")
		//flashMode = "dio"
		b.toolChain.Vars.Set("build.flash.mode", "dio")
	}

	// ------------------------------------------------------------------------------------------------
	// Partitions file setup
	b.partitionsFilepath = filepath.Join(buildPath, "partitions.csv")

	// ------------------------------------------------------------------------------------------------
	// Elf size tool setup
	b.elfSizeToolArgs, _ = b.toolChain.Vars.Get("recipe.size.pattern")
	b.elfSizeToolArgs = b.toolChain.Vars.FinalResolveArray(b.elfSizeToolArgs, b.vars)
	b.elfSizeToolPath = b.elfSizeToolArgs[0]
	b.elfSizeToolArgs = b.elfSizeToolArgs[1:]

	// ------------------------------------------------------------------------------------------------
	// Image partitions tool setup
	genImagePartitionsArgs, _ := b.toolChain.Vars.Get("recipe.objcopy.partitions.bin.pattern")
	genImagePartitionsArgs = b.toolChain.Vars.FinalResolveArray(genImagePartitionsArgs, b.vars)

	b.genImagePartitionsToolPath = genImagePartitionsArgs[0]
	b.genImagePartitionsToolArgs = genImagePartitionsArgs[1:]
	b.genImagePartitionsToolOutputFilepath = filepath.Join(buildPath, b.toolChain.ProjectName+".partitions.bin")
	b.genImagePartitionsToolInputFilepaths = []string{b.partitionsFilepath}

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

// Note: SRAM = IRAM + DRAM

type ImageStats struct {
	ImageSize int64 // Total image size = Flash size + RAM size
	FlashSize int64 // Flash size
	FlashCode int64 // Flash .text size
	FlashData int64 // Flash .rodata size
	RAMSize   int64 // RAM size = IRAM + DRAM + RTC
	IRAM0Size int64 // Instruction RAM size
	DRAM0Size int64 // Data RAM size
	RTCSize   int64 // RTC RAM size
}

func (s *ImageStats) Print() {
	corepkg.LogInfof("Image Size: %d bytes", s.ImageSize)
	corepkg.LogInfof("  Flash: %d bytes", s.FlashSize)
	corepkg.LogInfof("    Code Size: %d bytes", s.FlashCode)
	corepkg.LogInfof("    Data Size: %d bytes", s.FlashData)
	corepkg.LogInfof("  RAM: %d bytes", s.RAMSize)
	corepkg.LogInfof("    IRAM0 Size: %d bytes", s.IRAM0Size)
	corepkg.LogInfof("    DRAM0 Size: %d bytes", s.DRAM0Size)
	corepkg.LogInfof("    RTC Size: %d bytes", s.RTCSize)
}

func (b *ToolchainArduinoEsp32Burnerv2) AnalyzeElfSize(s string) (*ImageStats, error) {
	var gPatternsRTC = []string{".rtc_reserved"}
	var gPatternsIRAM = []string{".iram?.text", ".iram?.vectors", ".iram?.data"}
	var gPatternsDRAM = []string{".dram?.data", ".dram?.bss"}
	var gPatternsFLASHText = []string{".flash.text"}
	var gPatternsFLASHData = []string{".flash.rodata", ".flash.appdesc", ".flash.init_array", ".eh_frame"}

	// Match does a direct string match, and for the '?' character it will match any character.
	match := func(pattern string, str string) bool {
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

	collect := func(fields []string, patterns []string) (collected int64, found bool) {
		for _, pattern := range patterns {
			if match(pattern, fields[0]) {
				if size, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					collected += size
				}
				found = true
				return collected, found
			}
		}
		return 0, false
	}

	stats := &ImageStats{}
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

		if size, found := collect(fields, gPatternsIRAM); found {
			stats.IRAM0Size += size
		} else if size, found := collect(fields, gPatternsDRAM); found {
			stats.DRAM0Size += size
		} else if size, found := collect(fields, gPatternsFLASHText); found {
			stats.FlashCode += size
		} else if size, found := collect(fields, gPatternsFLASHData); found {
			stats.FlashData += size
		} else if size, found := collect(fields, gPatternsRTC); found {
			stats.RTCSize += size
		}
	}

	stats.FlashSize = stats.FlashCode + stats.FlashData
	stats.RAMSize = stats.IRAM0Size + stats.DRAM0Size + stats.RTCSize
	stats.ImageSize = stats.FlashSize + stats.RAMSize

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read output: %w", err)
	}

	return stats, nil
}

func (b *ToolchainArduinoEsp32Burnerv2) Build() error {

	// - Analyze the ELF file to get section sizes
	// - Generate the partitions.csv file
	// - Generate image partitions bin file ('PROJECT.NAME.partitions.bin')
	// - Generate image bin file ('PROJECT.NAME.bin')
	// - Generate bootloader image ('PROJECT.NAME.bootloader.bin')

	// ------------------------------------------------------------------------------------------------
	// ELF size analysis
	elfSizeToolPath := b.elfSizeToolPath
	elfSizeToolArgs := b.elfSizeToolArgs

	fmt.Printf("Analyzing ELF size, cmd args: %s %s\n", elfSizeToolPath, strings.Join(elfSizeToolArgs, "|"))

	cmd := exec.Command(elfSizeToolPath, elfSizeToolArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return corepkg.LogErrorf(err, "ELF size analysis failed")
	}

	//AnalyzeElfSize(string(out))
	elfStats, err := b.AnalyzeElfSize(string(out))
	if err != nil {
		return corepkg.LogErrorf(err, "ELF size analysis parsing failed")
	}
	elfStats.Print()

	// ------------------------------------------------------------------------------------------------
	// Generate the partitions.csv file, check if the file exists, hash the parameters to see if we need to
	// regenerate it.

	// TEST
	ota := true                                               // Enable OTA partitions
	otaSize := alignInt64(elfStats.ImageSize, int64(0x20000)) // OTA size aligned to 128KB
	nvsSize := int64(0x5000)                                  // 16KB NVS size
	fsType := "spiffs"                                        // Filesystem type
	fsSize := int64(0x40000)                                  // 64KB FS size
	coreDump := true                                          // Enable core dump partition

	// Hash the parameters
	partitionParams := fmt.Sprintf("%d|%d|%s|%d|%v", otaSize, nvsSize, fsType, fsSize, coreDump)
	partitionParamsHash := b.hashArguments([]string{partitionParams})

	if !b.dependencyTracker.QueryItemWithExtraData(b.partitionsFilepath, partitionParamsHash) {
		err := generatePartitions(b.partitionsFilepath, ota, otaSize, nvsSize, fsType, fsSize, coreDump)
		if err != nil {
			corepkg.LogErrorf(err, "Failed to generate partitions file")
		}
		b.dependencyTracker.AddItemWithExtraData(b.partitionsFilepath, partitionParamsHash, []string{})
	} else {
		b.dependencyTracker.CopyItem(b.partitionsFilepath)
	}

	// Generate the image partitions bin file
	if !b.dependencyTracker.QueryItemWithExtraData(b.genImagePartitionsToolOutputFilepath, b.genImagePartitionsToolArgsHash) {

		toolPath, _ := exec.LookPath(b.genImagePartitionsToolPath)
		toolArgs := b.genImagePartitionsToolArgs

		fmt.Printf("Creating image partitions, cmd args: %s %s\n", toolPath, strings.Join(toolArgs, "|"))

		cmd := exec.Command(toolPath, toolArgs...)
		corepkg.LogInfof("Creating image partitions '%s' ...", b.toolChain.ProjectName+".partitions.bin")
		out, err := cmd.CombinedOutput()
		if err != nil {
			corepkg.LogInfof("Image partitions output:\n%s", string(out))
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
		toolPath := b.genImageBinToolPath
		toolArgs := b.genImageBinToolArgs

		fmt.Printf("Generating image, cmd args: %s %s\n", toolPath, strings.Join(toolArgs, "|"))

		cmd := exec.Command(toolPath, toolArgs...)
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
		toolPath := b.genBootloaderToolPath
		toolArgs := b.genBootloaderToolArgs

		fmt.Printf("Generating bootloader, cmd args: %s %s\n", toolPath, strings.Join(toolArgs, "|"))

		cmd := exec.Command(toolPath, toolArgs...)
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

	flashToolArgs = slices.DeleteFunc(flashToolArgs, func(s string) bool { return strings.TrimSpace(s) == "--port" })

	corepkg.LogInfof("Flashing '%s'...", b.toolChain.ProjectName+".bin")

	fmt.Printf("Flashing with command: %s %s\n", flashToolPath, strings.Join(flashToolArgs, "|"))

	flashToolCmd := exec.Command(flashToolPath, flashToolArgs...)

	out, err := flashToolCmd.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			corepkg.LogInfof("Flashing output:\n%s", string(out))
		}
		return corepkg.LogErrorf(err, "Flashing failed with %s")
	}
	if len(out) > 0 {
		corepkg.LogInfof("Flashing output:\n%s", string(out))
	}

	//pipe, _ := flashToolCmd.StdoutPipe()
	//
	//if err := flashToolCmd.Start(); err != nil {
	//	return corepkg.LogErrorf(err, "Flashing failed")
	//}
	//
	//reader := bufio.NewReader(pipe)
	//line, err := reader.ReadString('\n')
	//for err == nil {
	//	line = strings.TrimRight(line, "\n")
	//	corepkg.LogInfo(line)
	//	line, err = reader.ReadString('\n')
	//	if err == io.EOF {
	//		err = nil
	//		break
	//	}
	//}
	//
	//if err != nil {
	//	return corepkg.LogErrorf(err, "Flashing failed")
	//}

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
