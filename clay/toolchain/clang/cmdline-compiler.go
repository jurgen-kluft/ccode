package clang

import "github.com/jurgen-kluft/ccode/foundation"

type CompilerFlags uint64

const (
	CompilerFlagDebug CompilerFlags = 1 << iota
	CompilerFlagRelease
	CompilerFlagFinal
	CompilerFlagTest
	CompilerFlagCpp
)

func (c CompilerFlags) IsDebug() bool {
	return c&CompilerFlagDebug != 0
}
func (c CompilerFlags) IsRelease() bool {
	return c&CompilerFlagRelease != 0
}
func (c CompilerFlags) IsFinal() bool {
	return c&CompilerFlagFinal != 0
}
func (c CompilerFlags) IsTest() bool {
	return c&CompilerFlagTest != 0
}
func (c CompilerFlags) IsCpp() bool {
	return c&CompilerFlagCpp != 0
}
func (c CompilerFlags) IsC() bool {
	return c&CompilerFlagCpp == 0
}

type CompilerCmdLine struct {
	args   *foundation.Arguments
	length int
}

func NewCompilerCmdline(args *foundation.Arguments) *CompilerCmdLine {
	return &CompilerCmdLine{
		args:   args,
		length: 0,
	}
}

func (c *CompilerCmdLine) Add(arg string) { c.args.Add(arg) }
func (c *CompilerCmdLine) AddWithPrefix(prefix string, args ...string) {
	c.args.AddWithPrefix(prefix, args...)
}

func (c *CompilerCmdLine) CompileOnly()                 { c.Add("-c") }                        // Compile only; do not link.
func (c *CompilerCmdLine) NoLogo()                      { c.Add("-nologo") }                   // Suppress the display of the compiler's startup banner and copyright message.
func (c *CompilerCmdLine) WarningsAreErrors()           { c.Add("-Werror") }                   // Treat warnings as errors.
func (c *CompilerCmdLine) DisableOptimizations()        { c.Add("-O0") }                       // Disable optimizations for debugging.
func (c *CompilerCmdLine) GenerateDebugInfo()           { c.Add("-g") }                        // Generate complete debugging information.
func (c *CompilerCmdLine) DisableFramePointer()         { c.Add("-fno-omit-frame-pointer") }   // Do not omit frame pointer.
func (c *CompilerCmdLine) OptimizeForSize()             { c.Add("-Os") }                       // Optimize for size.
func (c *CompilerCmdLine) OptimizeForSpeed()            { c.Add("-O2") }                       // Optimize for speed.
func (c *CompilerCmdLine) OptimizeHard()                { c.Add("-O3") }                       // Optimize hard, enabling aggressive optimizations.
func (c *CompilerCmdLine) EnableInlineExpansion()       { c.Add("-finline-functions") }        // Enable inline expansion for functions that are small and frequently called.
func (c *CompilerCmdLine) EnableIntrinsicFunctions()    { c.Add("-ffunction-sections") }       // Enable intrinsic functions.
func (c *CompilerCmdLine) OmitFramePointer()            { c.Add("-fomit-frame-pointer") }      // Omit frame pointer for functions that do not require one.
func (c *CompilerCmdLine) EnableExceptionHandling()     { c.Add("-fexceptions") }              // Enable C++ exception handling.
func (c *CompilerCmdLine) UseFloatingPointPrecise()     { c.Add("-ffp-model=precise") }        // Use floating-point model: precise
func (c *CompilerCmdLine) UseCpp14()                    { c.Add("-std=c++14") }                // Use C++14 standard.
func (c *CompilerCmdLine) UseCpp17()                    { c.Add("-std=c++17") }                // Use C++17 standard.
func (c *CompilerCmdLine) UseCpp20()                    { c.Add("-std=c++20") }                // Use C++20 standard.
func (c *CompilerCmdLine) UseCppLatest()                { c.Add("-std=c++latest") }            // Use the latest C++ standard.
func (c *CompilerCmdLine) UseC11()                      { c.Add("-std=c11") }                  // Use C11 standard.
func (c *CompilerCmdLine) UseC17()                      { c.Add("-std=c17") }                  // Use C17 standard.
func (c *CompilerCmdLine) UseCLatest()                  { c.Add("-std=clatest") }              // Use the latest C standard.
func (c *CompilerCmdLine) GenerateDependencyFiles()     { c.Add("-MMD") }                      // Generate a dependency file for every source files being compiled.
func (c *CompilerCmdLine) Includes(includes []string)   { c.AddWithPrefix("-I", includes...) } // Add include paths.
func (c *CompilerCmdLine) Defines(defines []string)     { c.AddWithPrefix("-D", defines...) }  // Add preprocessor definitions.
func (c *CompilerCmdLine) ObjectFile(outputFile string) { c.AddWithPrefix("-o", outputFile) }
func (c *CompilerCmdLine) SourceFile(sourceFile string) { c.Add(sourceFile) }
func (c *CompilerCmdLine) Save()                        { c.length = c.args.Len() }
func (c *CompilerCmdLine) Restore()                     { c.args.Args = c.args.Args[:c.length] }

func GenerateCompilerCmdline(flags CompilerFlags, includes []string, defines []string, sourceFiles []string, objectFiles []string) *foundation.Arguments {
	args := foundation.NewArguments(len(sourceFiles) + len(objectFiles) + 20)

	c := NewCompilerCmdline(args)

	c.CompileOnly()
	c.NoLogo()
	c.WarningsAreErrors()

	// Debug-specific arguments
	if flags.IsDebug() {
		c.DisableOptimizations()
		c.GenerateDebugInfo()
		c.DisableFramePointer()
	}

	c.Includes(includes)
	c.Defines(defines)

	// Release and Final specific arguments
	if flags.IsRelease() || flags.IsFinal() {
		c.OptimizeForSize()
		c.OptimizeForSpeed()
		c.OptimizeHard()
		c.EnableInlineExpansion()
		c.EnableIntrinsicFunctions()
		c.OmitFramePointer()
	}
	// Test-specific arguments
	if flags.IsTest() {
		c.EnableExceptionHandling()
	}
	c.UseFloatingPointPrecise()
	// C++ specific arguments
	if flags.IsCpp() {
		c.UseCpp17()
	}
	// C specific arguments
	if flags.IsC() {
		c.UseC11()
	}
	c.GenerateDependencyFiles()

	return args
}
