package toolchain

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
	"github.com/jurgen-kluft/ccode/foundation"
)

type DarwinClang struct {
	Name string
	Vars *foundation.Vars
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// C/C++ Compiler

type ToolchainDarwinClangCompiler struct {
	toolChain       *DarwinClang
	config          *Config
	cCompilerPath   string
	cppCompilerPath string
	cArgs           *Arguments
	cppArgs         *Arguments
}

func (t *DarwinClang) NewCompiler(config *Config) Compiler {
	return &ToolchainDarwinClangCompiler{
		toolChain:       t,
		config:          config,
		cCompilerPath:   t.Vars.GetFirstOrEmpty("c.compiler"),
		cppCompilerPath: t.Vars.GetFirstOrEmpty("cpp.compiler"),
		cArgs:           NewArguments(512),
		cppArgs:         NewArguments(512),
	}
}

func (cl *ToolchainDarwinClangCompiler) SetupArgs(_defines []string, _includes []string) {
	// Implement the logic to setup arguments for the compiler here
	cl.cArgs.Clear()

	cl.cArgs.Add(`-c`)
	if archFlags, ok := cl.toolChain.Vars.Get(`c.compiler.flags.arch`); ok {
		cl.cArgs.Add(archFlags...)
	}
	if picFlags, ok := cl.toolChain.Vars.Get(`c.compiler.flags.pic`); ok {
		cl.cArgs.Add(picFlags...)
	}
	if stdFlags, ok := cl.toolChain.Vars.Get(`c.compiler.flags.std`); ok {
		cl.cArgs.Add(stdFlags...)
	}

	flagsStr := `c.compiler.flags.release`
	definesStr := `c.compiler.defines.release`
	if cl.config.Config.IsDebug() {
		flagsStr = `c.compiler.flags.debug`
		definesStr = `c.compiler.defines.debug`
	} else if cl.config.Config.IsFinal() {
		flagsStr = `c.compiler.flags.final`
		definesStr = `c.compiler.defines.final`
	}

	if flags, ok := cl.toolChain.Vars.Get(flagsStr); ok {
		cl.cArgs.Add(flags...)
	}
	if defines, ok := cl.toolChain.Vars.Get(definesStr); ok {
		for _, define := range defines {
			cl.cArgs.Add(`-D`, define)
		}
	}
	cl.cArgs.AddWithPrefix(`-D`, _defines...)
	cl.cArgs.AddWithPrefix(`-I`, _includes...)

	cl.cArgs.Add(`-MMD`) // Generate dependency file

	// C++ compiler arguments
	cl.cppArgs.Clear()

	cl.cppArgs.Add(`-c`)
	if archFlags, ok := cl.toolChain.Vars.Get(`cpp.compiler.flags.arch`); ok {
		cl.cppArgs.Add(archFlags...)
	}
	if picFlags, ok := cl.toolChain.Vars.Get(`cpp.compiler.flags.pic`); ok {
		cl.cppArgs.Add(picFlags...)
	}
	if stdFlags, ok := cl.toolChain.Vars.Get(`cpp.compiler.flags.std`); ok {
		cl.cppArgs.Add(stdFlags...)
	}

	flagsStr = `cpp.compiler.flags.release`
	definesStr = `cpp.compiler.defines.release`
	if cl.config.Config.IsDebug() {
		flagsStr = `cpp.compiler.flags.debug`
		definesStr = `cpp.compiler.defines.debug`
	} else if cl.config.Config.IsFinal() {
		flagsStr = `cpp.compiler.flags.final`
		definesStr = `cpp.compiler.defines.final`
	}

	if flags, ok := cl.toolChain.Vars.Get(flagsStr); ok {
		cl.cppArgs.Add(flags...)
	}

	if defines, ok := cl.toolChain.Vars.Get(definesStr); ok {
		cl.cppArgs.AddWithPrefix(`-D`, defines...)
	}
	cl.cppArgs.AddWithPrefix(`-D`, _defines...)
	cl.cppArgs.AddWithPrefix(`-I`, _includes...)

	cl.cppArgs.Add(`-MMD`) // Generate dependency file

}

func (cl *ToolchainDarwinClangCompiler) Compile(sourceAbsFilepath string, objRelFilepath string) error {
	var compilerPath string
	var compilerArgs []string
	if strings.HasSuffix(sourceAbsFilepath, ".c") {
		compilerPath = cl.cCompilerPath
		compilerArgs = cl.cArgs.Args
	} else {
		compilerPath = cl.cppCompilerPath
		compilerArgs = cl.cppArgs.Args
	}

	// The source file and the output object file
	compilerArgs = append(compilerArgs, "-o")
	compilerArgs = append(compilerArgs, objRelFilepath)
	compilerArgs = append(compilerArgs, sourceAbsFilepath)

	foundation.LogInfof("Compiling (%s) %s\n", cl.config.Config.AsString(), filepath.Base(sourceAbsFilepath))

	var cmd *exec.Cmd
	cmd = exec.Command(compilerPath, compilerArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		foundation.LogInfof("Compile failed, output:\n%s\n", string(out))
		return foundation.LogErrorf(err, "Compiling failed")
	}
	if len(out) > 0 {
		foundation.LogInfof("Compile output:\n%s\n", string(out))
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
			archiverPath: t.Vars.GetFirstOrEmpty("archiver.static"),
			args:         []string{},
		}
	} else if at == ArchiverTypeDynamic {
		return &ToolchainDarwinClangDynamicArchiver{
			toolChain:    t,
			archiverPath: t.Vars.GetFirstOrEmpty("archiver.dynamic"),
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
	if archFlags, ok := t.toolChain.Vars.Get(`static.archiver.flags`); ok {
		t.args = append(t.args, archFlags...)
	}
}

func (t *ToolchainDarwinClangStaticArchiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {
	args := t.args
	args = append(args, outputArchiveFilepath)
	for _, inputObjectFilepath := range inputObjectFilepaths {
		args = append(args, inputObjectFilepath)
	}

	cmd := exec.Command(t.archiverPath, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return foundation.LogErrorf(err, "Archiving failed: ", string(out))
	}

	return nil
}

func (t *ToolchainDarwinClangDynamicArchiver) Filename(name string) string {
	return "lib" + name + ".dylib" // The filename for the dynamic library on Darwin is typically "libname.dylib"
}
func (t *ToolchainDarwinClangDynamicArchiver) SetupArgs() {
	t.args = []string{}

	if flags, ok := t.toolChain.Vars.Get(`dynamic.archiver.flags`); ok {
		t.args = append(t.args, flags...)
	}
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

	cmd := exec.Command(t.archiverPath, args...)
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

type ToolchainDarwinClangLinker struct {
	toolChain  *DarwinClang
	linkerPath string
	args       []string
}

func (l *DarwinClang) NewLinker(config *Config) Linker {
	return &ToolchainDarwinClangLinker{
		toolChain:  l,
		linkerPath: l.Vars.GetFirstOrEmpty("linker"),
	}
}

func (l *ToolchainDarwinClangLinker) Filename(name string) string {
	return name // The filename for the output binary on Darwin is typically just the name without extension
}

func (l *ToolchainDarwinClangLinker) SetupArgs(generateMapFile bool, libraryPaths []string, libraryFiles []string) {
	l.args = []string{}

	// Library paths
	if libPaths, ok := l.toolChain.Vars.Get("linker.lib.paths"); ok {
		for _, libPath := range libPaths {
			l.args = append(l.args, `-L`)
			l.args = append(l.args, libPath)
		}
	}

	// Frameworks
	if frameworks, ok := l.toolChain.Vars.Get("linker.frameworks"); ok {
		for _, framework := range frameworks {
			l.args = append(l.args, "-framework")
			l.args = append(l.args, framework)
		}
	}
}

func (l *ToolchainDarwinClangLinker) Link(inputArchiveAbsFilepaths []string, outputAppRelFilepathNoExt string) error {
	args := l.args

	if flags, ok := l.toolChain.Vars.Get(`linker.flags`); ok {
		args = append(args, flags...)
	}

	args = append(args, "-o")
	args = append(args, outputAppRelFilepathNoExt)

	for _, inputArchiveAbsFilepath := range inputArchiveAbsFilepaths {
		args = append(args, inputArchiveAbsFilepath)
	}

	if libFiles, ok := l.toolChain.Vars.Get("linker.lib.files"); ok {
		for _, libFile := range libFiles {
			args = append(args, "-l")
			args = append(args, libFile)
		}
	}

	cmd := exec.Command(l.linkerPath, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		foundation.LogInff("Link failed, output:\n%s\n", string(out))
		return foundation.LogError(err, "Linking failed")
	}
	if len(out) > 0 {
		foundation.LogInfof("Link output:\n%s\n", string(out))
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
func (t *DarwinClang) NewDependencyTracker(dirpath string) deptrackr.FileTrackr {
	return deptrackr.LoadFileTrackr(filepath.Join(dirpath, "deptrackr"))
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
		Vars: foundation.NewVarsCustom(foundation.VarsFormatCurlyBraces),
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
		`dynamic.archiver.flags`: {},

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

	// TODO target x86_64 or aarch64 on macOS

	t.Vars.Cull()
	t.Vars.Resolve()
	return t, nil
}
