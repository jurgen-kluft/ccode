package toolchain

// Compiler is an interface that defines the methods required for compiling source files
type Compiler interface {
	// Returns the object filepath for the compiler to output
	// e.g. path/to/source/file.c and it will return -> path/to/source/file.c.o or path/to/source/file.c.obj
	// depending on the platform.
	ObjFilepath(srcRelFilepath string) string

	// Returns the filepath for the dependency output
	// e.g. path/to/object/file.obj and it will return -> path/to/object/file.obj.d or path/to/object/file.obj.json
	// depending on the platform.
	DepFilepath(objRelFilepath string) string

	// SetupArgs prepares the arguments for the compiler
	// It should be called before using the Compile method.
	SetupArgs(projectName string, buildPath string, defines []string, includes []string)

	// Compile takes a list of input source file paths and output object file paths
	// The source file paths may be absolute or relative to the build directory, however
	// the object file paths should be relative to the build directory.
	// Returns an array of booleans indicating success or failure for each file compiled,
	// and a boolean indicating if any compilation failed.
	Compile(sourceAbsFilepath []string, objRelFilepath []string) ([]bool, bool)
}
