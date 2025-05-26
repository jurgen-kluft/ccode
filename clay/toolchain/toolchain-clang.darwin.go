package toolchain

import (
	"os/exec"
)

type ToolchainDarwinClang struct {
	ToolchainInstance
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// C Compiler

type ToolchainDarwinClangCCompiler struct {
	toolChain *ToolchainDarwinClang
	toolPath  string
	args      []string
	config    string
}

func (t *ToolchainDarwinClang) NewCCompiler(config string) Compiler {
	return &ToolchainDarwinClangCCompiler{
		toolChain: t,
		toolPath:  t.Tools["c.compiler"],
		config:    config,
	}
}

func (cl *ToolchainDarwinClangCCompiler) AddDefine(define string) {
	// Implement the logic to add a define here
}
func (cl *ToolchainDarwinClangCCompiler) AddIncludePath(path string) {
	// Implement the logic to add an include path here
}
func (cl *ToolchainDarwinClangCCompiler) SetupArgs(userVars Vars) {
	// Implement the logic to setup arguments for the compiler here
	cl.args = []string{}
	if ConfigMatches(cl.config, "*-*-debug-*") {
		cl.args = append(cl.args, `-g`, `-O0`)
	} else if ConfigMatches(cl.config, "*-*-release-*") {
		cl.args = append(cl.args, `-O3`)
	}
	cl.args = append(cl.args, `-fPIC`)
}
func (cl *ToolchainDarwinClangCCompiler) Compile(sourceAbsFilepath string, sourceRelFilepath string) (string, error) {
	// Implement the compile logic here
	return "objectfilepath", nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// C++ Compiler

type ToolchainDarwinClangCppCompiler struct {
	toolChain    *ToolchainDarwinClang
	compilerPath string
	compilerArgs []string
	config       string
}

func (t *ToolchainDarwinClang) NewCppCompiler(config string) Compiler {
	return &ToolchainDarwinClangCppCompiler{
		toolChain:    t,
		compilerPath: t.Tools["cpp.compiler"],
		compilerArgs: []string{},
		config:       config,
	}
}

func (cl *ToolchainDarwinClangCppCompiler) AddDefine(define string) {
	// Implement the logic to add a define here
}
func (cl *ToolchainDarwinClangCppCompiler) AddIncludePath(path string) {
	// Implement the logic to add an include path here
}
func (cl *ToolchainDarwinClangCppCompiler) SetupArgs(userVars Vars) {
	// Implement the logic to setup arguments for the compiler here
	cl.compilerArgs = []string{}
	if ConfigMatches(cl.config, "*-*-debug-*") {
		cl.compilerArgs = append(cl.compilerArgs, `-g`, `-O0`)
	} else if ConfigMatches(cl.config, "*-*-release-*") {
		cl.compilerArgs = append(cl.compilerArgs, `-O3`)
	}
	cl.compilerArgs = append(cl.compilerArgs, `-fPIC`)
}
func (cl *ToolchainDarwinClangCppCompiler) Compile(sourceFilepath string, objectFilepath string) (string, error) {
	// Implement the compile logic here
	return "objectfilepath", nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Archiver

type ToolchainDarwinClangArchiver struct {
	toolChain *ToolchainDarwinClang
	toolPath  string
}

func (t *ToolchainDarwinClang) NewArchiver() Archiver {
	return &ToolchainDarwinClangArchiver{
		toolChain: t,
		toolPath:  t.Tools["archiver"],
	}
}

func (t *ToolchainDarwinClangArchiver) SetupArgs(userVars Vars) {
	// Implement the logic here
}

func (cl *ToolchainDarwinClangArchiver) Archive(inputObjectFilepaths []string, outputArchiveFilepath string) error {
	// Implement the logic here
	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Linker

type ToolchainDarwinClangLinker struct {
	toolChain *ToolchainDarwinClang
	toolPath  string
	args      []string
}

func (t *ToolchainDarwinClang) NewLinker() Linker {
	return &ToolchainDarwinClangLinker{
		toolChain: t,
		toolPath:  t.Tools["linker"],
	}
}

func (cl *ToolchainDarwinClangLinker) GenerateMapFile() {
	// Implement the logic to generate a map file here
}
func (cl *ToolchainDarwinClangLinker) AddLibraryPath(path string) {
	// Implement the logic to add a library path here
}
func (cl *ToolchainDarwinClangLinker) AddLibraryFile(lib string) {
	// Implement the logic to add a library file here
}
func (t *ToolchainDarwinClangLinker) SetupArgs(userVars Vars) {
	libPaths := t.toolChain.Vars["linker.lib.paths"]
	t.args = []string{}
	for _, libPath := range libPaths {
		t.args = append(t.args, `-L`)
		t.args = append(t.args, libPath)
	}
	libFiles := t.toolChain.Vars["linker.lib.files"]
	for _, libFile := range libFiles {
		t.args = append(t.args, `-l`)
		t.args = append(t.args, libFile)
	}
}
func (cl *ToolchainDarwinClangLinker) Link(inputArchiveAbsFilepaths []string, outputAppRelFilepathNoExt string) error {
	// Implement the compile logic here
	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Burner

func (t *ToolchainDarwinClang) NewBurner() Burner {
	return &ToolchainEmptyBurner{}
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Toolchain for Clang on MacOS
func NewToolchainClangDarwin() (t *ToolchainDarwinClang, err error) {
	var clangPath string
	if clangPath, err = exec.LookPath("clang"); err != nil {
		return nil, err
	}

	t = &ToolchainDarwinClang{
		ToolchainInstance{
			Name: "clang",
			Vars: map[string][]string{
				"clang.path": {
					clangPath,
				},
				"clang.bin.path": {
					`{clang.path}/bin`,
				},
				"clang.lib.path": {
					`{clang.path}/lib`,
				},
				"c.compiler.includes": {
					`{clang.path}/include`,
					`{clang.path}/usr/include`,
					`{clang.path}/usr/local/include`,
				},
				"c++.compiler.includes": {
					`{clang.path}/include`,
					`{clang.path}/usr/include`,
					`{clang.path}/usr/local/include`,
				},
				"linker.lib.paths": {
					`{clang.lib.path}`,
				},
				"linker.lib.files": {
					`libc++.dylib`,
					`libSystem.dylib`,
					`libc++abi.dylib`,
					`libobjc.A.dylib`,
				},
			},

			// #--------------------------------------------------
			Tools: map[string]string{
				"c.compiler":   `{clang.bin.path}/clang`,
				"cpp.compiler": `{clang.bin.path}/clang`,
				"archiver":     `{clang.bin.path}/clang`,
				"linker":       `{clang.bin.path}/clang`,
			},
		},
	}

	t.ResolveVars()
	return t, nil
}
