package toolchain

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain/dpenc"
	utils "github.com/jurgen-kluft/ccode/utils"
)

type DarwinClang struct {
	Name string
	Vars *utils.Vars
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// C/C++ Compiler

type ToolchainDarwinClangCompiler struct {
	toolChain       *DarwinClang
	config          *Config
	cCompilerPath   string
	cppCompilerPath string
	cArgs           []string
	cppArgs         []string
}

func (t *DarwinClang) NewCompiler(config *Config) Compiler {
	return &ToolchainDarwinClangCompiler{
		toolChain:       t,
		config:          config,
		cCompilerPath:   t.Vars.GetOne("c.compiler"),
		cppCompilerPath: t.Vars.GetOne("cpp.compiler"),
		cArgs:           []string{},
		cppArgs:         []string{},
	}
}

func (cl *ToolchainDarwinClangCompiler) SetupArgs(_defines []string, _includes []string) {
	// Implement the logic to setup arguments for the compiler here
	cl.cArgs = make([]string, 0, 64)

	cl.cArgs = append(cl.cArgs, `-c`)
	archFlags := cl.toolChain.Vars.GetAll(`c.compiler.flags.arch`)
	cl.cArgs = append(cl.cArgs, archFlags...)
	picFlags := cl.toolChain.Vars.GetAll(`c.compiler.flags.pic`)
	cl.cArgs = append(cl.cArgs, picFlags...)
	stdFlags := cl.toolChain.Vars.GetAll(`c.compiler.flags.std`)
	cl.cArgs = append(cl.cArgs, stdFlags...)

	flagsStr := `c.compiler.flags.release`
	definesStr := `c.compiler.defines.release`
	if cl.config.Config.IsDebug() {
		flagsStr = `c.compiler.flags.debug`
		definesStr = `c.compiler.defines.debug`
	} else if cl.config.Config.IsFinal() {
		flagsStr = `c.compiler.flags.final`
		definesStr = `c.compiler.defines.final`
	}

	flags := cl.toolChain.Vars.GetAll(flagsStr)
	cl.cArgs = append(cl.cArgs, flags...)
	defines := cl.toolChain.Vars.GetAll(definesStr)
	for _, define := range defines {
		cl.cArgs = append(cl.cArgs, `-D`, define)
	}
	for _, define := range _defines {
		cl.cArgs = append(cl.cArgs, `-D`, define)
	}
	for _, include := range _includes {
		cl.cArgs = append(cl.cArgs, `-I`, include)
	}

	cl.cArgs = append(cl.cArgs, `-MMD`) // Generate dependency file

	// C++ compiler arguments
	cl.cppArgs = make([]string, 0, 64)

	cl.cppArgs = append(cl.cppArgs, `-c`)
	archFlags = cl.toolChain.Vars.GetAll(`cpp.compiler.flags.arch`)
	cl.cppArgs = append(cl.cppArgs, archFlags...)
	picFlags = cl.toolChain.Vars.GetAll(`cpp.compiler.flags.pic`)
	cl.cppArgs = append(cl.cppArgs, picFlags...)
	stdFlags = cl.toolChain.Vars.GetAll(`cpp.compiler.flags.std`)
	cl.cppArgs = append(cl.cppArgs, stdFlags...)

	flagsStr = `cpp.compiler.flags.release`
	definesStr = `cpp.compiler.defines.release`
	if cl.config.Config.IsDebug() {
		flagsStr = `cpp.compiler.flags.debug`
		definesStr = `cpp.compiler.defines.debug`
	} else if cl.config.Config.IsFinal() {
		flagsStr = `cpp.compiler.flags.final`
		definesStr = `cpp.compiler.defines.final`
	}

	flags = cl.toolChain.Vars.GetAll(flagsStr)
	cl.cppArgs = append(cl.cppArgs, flags...)
	defines = cl.toolChain.Vars.GetAll(definesStr)
	for _, define := range defines {
		cl.cppArgs = append(cl.cppArgs, `-D`, define)
	}
	for _, define := range _defines {
		cl.cppArgs = append(cl.cppArgs, `-D`, define)
	}
	for _, include := range _includes {
		cl.cppArgs = append(cl.cppArgs, `-I`, include)
	}

	cl.cppArgs = append(cl.cppArgs, `-MMD`) // Generate dependency file

}

func (cl *ToolchainDarwinClangCompiler) Compile(sourceAbsFilepath string, objRelFilepath string) error {

	var args []string
	if strings.HasSuffix(sourceAbsFilepath, ".c") {
		args = cl.cArgs
	} else {
		args = cl.cppArgs
	}

	// The source file and the output object file
	args = append(args, "-o")
	args = append(args, objRelFilepath)
	args = append(args, sourceAbsFilepath)

	fmt.Printf("Compiling (%s) %s\n", cl.config.Config.AsString(), filepath.Base(sourceAbsFilepath))

	cmd := exec.Command(cl.cCompilerPath, args...)
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

type ToolchainDarwinClangStaticArchiver struct {
	toolChain    *DarwinClang
	archiverPath string
	args         []string
}

type ToolchainDarwinClangDynamicArchiver struct {
	toolChain    *DarwinClang
	archiverPath string
	args         []string
}

func (t *DarwinClang) NewArchiver(at ArchiverType, config *Config) Archiver {
	if at == ArchiverTypeStatic {
		return &ToolchainDarwinClangStaticArchiver{
			toolChain:    t,
			archiverPath: t.Vars.GetOne("archiver.static"),
			args:         []string{},
		}
	} else if at == ArchiverTypeDynamic {
		return &ToolchainDarwinClangDynamicArchiver{
			toolChain:    t,
			archiverPath: t.Vars.GetOne("archiver.dynamic"),
			args:         []string{},
		}
	}
	return nil
}

func (t *ToolchainDarwinClangStaticArchiver) Filename(name string) string {
	return "lib" + name + ".a" // The file extension for the archive on Darwin is typically ".a"
}

func (t *ToolchainDarwinClangStaticArchiver) SetupArgs() {
	t.args = []string{}
	archFlags := t.toolChain.Vars.GetAll(`static.archiver.flags`)
	t.args = append(t.args, archFlags...)
}

func (t *ToolchainDarwinClangStaticArchiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {
	args := t.args
	args = append(args, outputArchiveFilepath)
	for _, inputObjectFilepath := range inputObjectFilepaths {
		args = append(args, inputObjectFilepath)
	}

	// log.Printf("Archiving static library %s\n", outputArchiveFilepath)

	cmd := exec.Command(t.archiverPath, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("Archive failed with output:\n%s\n", string(out))
		return fmt.Errorf("Archive failed with %s\n", err)
	}

	return nil
}

func (t *ToolchainDarwinClangDynamicArchiver) Filename(name string) string {
	return "lib" + name + ".dylib" // The filename for the dynamic library on Darwin is typically "libname.dylib"
}
func (t *ToolchainDarwinClangDynamicArchiver) SetupArgs() {
	t.args = []string{}

	flags := t.toolChain.Vars.GetAll(`dynamic.archiver.flags`)
	t.args = append(t.args, flags...)
}

func (t *ToolchainDarwinClangDynamicArchiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {
	args := t.args
	args = append(args, "-dynamiclib")
	args = append(args, "-install_name")
	args = append(args, filepath.Base(outputArchiveFilepath))

	args = append(args, "-o")
	args = append(args, outputArchiveFilepath)

	for _, inputObjectFilepath := range inputObjectFilepaths {
		args = append(args, inputObjectFilepath)
	}

	// log.Printf("Archiving dynamic library %s\n", outputArchiveFilepath)

	cmd := exec.Command(t.archiverPath, args...)
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

type ToolchainDarwinClangLinker struct {
	toolChain  *DarwinClang
	linkerPath string
	args       []string
}

func (l *DarwinClang) NewLinker(config *Config) Linker {
	return &ToolchainDarwinClangLinker{
		toolChain:  l,
		linkerPath: l.Vars.GetOne("linker"),
	}
}

func (l *ToolchainDarwinClangLinker) Filename(name string) string {
	return name // The filename for the output binary on Darwin is typically just the name without extension
}

func (l *ToolchainDarwinClangLinker) SetupArgs(generateMapFile bool, libraryPaths []string, libraryFiles []string) {
	l.args = []string{}

	// Library paths
	libPaths := l.toolChain.Vars.GetAll("linker.lib.paths")
	for _, libPath := range libPaths {
		l.args = append(l.args, `-L`)
		l.args = append(l.args, libPath)
	}

	// Frameworks
	frameworks := l.toolChain.Vars.GetAll("linker.frameworks")
	for _, framework := range frameworks {
		l.args = append(l.args, "-framework")
		l.args = append(l.args, framework)
	}
}

func (l *ToolchainDarwinClangLinker) Link(inputArchiveAbsFilepaths []string, outputAppRelFilepathNoExt string) error {
	args := l.args

	flags := l.toolChain.Vars.GetAll(`linker.flags`)
	args = append(args, flags...)

	args = append(args, "-o")
	args = append(args, outputAppRelFilepathNoExt)

	for _, inputArchiveAbsFilepath := range inputArchiveAbsFilepaths {
		args = append(args, inputArchiveAbsFilepath)
	}

	libFiles := l.toolChain.Vars.GetAll("linker.lib.files")
	for _, libFile := range libFiles {
		args = append(args, "-l")
		args = append(args, libFile)
	}

	// log.Printf("Linking '%s'...\n", outputAppRelFilepathNoExt)
	cmd := exec.Command(l.linkerPath, args...)
	out, err := cmd.CombinedOutput()

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

func (t *DarwinClang) NewBurner(config *Config) Burner {
	return &EmptyBurner{}
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Dependency Tracker
func (t *DarwinClang) NewDependencyTracker(dirpath string) dpenc.FileTrackr {
	return dpenc.LoadFileTrackr(filepath.Join(dirpath, "deptrackr"))
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Toolchain for Clang on MacOS
const (
	archtype_arm    string = "arm"    // ARM: arm, armv.*, xscale
	archtype_arm64  string = "arm64"  // ARM: arm, armv.*, xscale
	archtype_x86    string = "x86"    // X86: i[3-9]86
	archtype_x86_64 string = "x86_64" // X86-64: amd64, x86_64
)

func NewDarwinClang(arch string, frameworks []string) (t *DarwinClang, err error) {
	var clangPath string
	if clangPath, err = exec.LookPath("clang"); err != nil {
		return nil, err
	}
	clangPath = filepath.Dir(clangPath)

	var arPath string
	if arPath, err = exec.LookPath("ar"); err != nil {
		return nil, err
	}
	arPath = filepath.Dir(arPath)

	t = &DarwinClang{
		Name: "clang",
		Vars: utils.NewVars(),
	}

	vars := map[string][]string{
		"ar.path":        {arPath},
		"clang.path":     {clangPath},
		"clang.lib.path": {`{clang.path}/lib`},

		"c.compiler.flags.arch":      {`-arch`, archtype_arm64},
		"c.compiler.flags.std":       {`-std=c11`},
		"c.compiler.flags.debug":     {`-g`, `-O0`},
		"c.compiler.flags.release":   {`-O2`},
		"c.compiler.flags.final":     {`-O3`},
		"c.compiler.defines.debug":   {},
		"c.compiler.defines.release": {},
		"c.compiler.defines.final":   {},

		"cpp.compiler.flags":           {`-arch`, archtype_arm64},
		"cpp.compiler.flags.std":       {`-std=c++17`},
		"cpp.compiler.flags.debug":     {`-g`, `-O0`},
		"cpp.compiler.flags.release":   {`-O2`},
		"cpp.compiler.flags.final":     {`-O3`},
		"cpp.compiler.defines.debug":   {},
		"cpp.compiler.defines.release": {},
		"cpp.compiler.defines.final":   {},

		"m.compiler.includes": {},
		"m.compiler.flags":    {`-arch`, archtype_arm64, `-fobjc-arc`},

		// specific flags for archiver
		`static.archiver.flags`:  {`-rs`},
		`dynamic.archiver.flags`: {``},

		`linker.flags`:      {},
		"linker.lib.paths":  {},
		"linker.lib.files":  {`stdc++`},
		"linker.frameworks": {},

		"c.compiler":       {`{clang.path}/clang`},
		"cpp.compiler":     {`{clang.path}/clang++`},
		"archiver.static":  {`{ar.path}/ar`},
		"archiver.dynamic": {`{clang.path}/clang`},
		"linker":           {`{clang.path}/clang`},
	}

	for key, value := range vars {
		t.Vars.Set(key, value...)
	}

	if len(frameworks) > 0 {
		t.Vars.Set("linker.frameworks", frameworks...)
	}

	// We can target x86_64 and aarch64 on macOS

	ResolveVars(t.Vars)
	return t, nil
}
