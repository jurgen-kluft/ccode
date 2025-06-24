package msvc

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
	args  *foundation.Arguments
	flags CompilerFlags // Build configuration
}

func (c *CompilerCmdLine) Add(arg string) { c.args.Add(arg) }
func (c *CompilerCmdLine) AddWithPrefix(prefix string, args ...string) {
	c.args.AddWithPrefix(prefix, args...)
}
func (c *CompilerCmdLine) CompileOnly()                          { c.Add("/c") }                        // Compile only; do not link. This is useful for generating object files without creating an executable.
func (c *CompilerCmdLine) NoLogo()                               { c.Add("/nologo") }                   // Suppress the display of the compiler's startup banner and copyright message.
func (c *CompilerCmdLine) DiagnosticsColumnMode()                { c.Add("/diagnostics:column") }       // Enable column mode for diagnostics, which provides more detailed error messages.
func (c *CompilerCmdLine) WarningLevel3()                        { c.Add("/W3") }                       // Set output warning level to 3 (high warnings).
func (c *CompilerCmdLine) WarningsAreErrors()                    { c.Add("/WX") }                       // Treat warnings as errors.
func (c *CompilerCmdLine) BuildMultipleSourceFilesConcurrently() { c.Add("/MP") }                       // Build multiple source files concurrently.
func (c *CompilerCmdLine) DisableOptimizations()                 { c.Add("/Od") }                       // Disable optimizations for debugging.
func (c *CompilerCmdLine) GenerateDebugInfo()                    { c.Add("/Zi") }                       // Generate complete debugging information.
func (c *CompilerCmdLine) DisableFramePointer()                  { c.Add("/Oy-") }                      // Do not omit frame pointer.
func (c *CompilerCmdLine) UseMultithreadedDebugRuntime()         { c.Add("/MTd") }                      // Use the multithreaded debug version of the C runtime library.
func (c *CompilerCmdLine) Includes(includes []string)            { c.AddWithPrefix("/I", includes...) } //
func (c *CompilerCmdLine) Defines(defines []string)              { c.AddWithPrefix("/D", defines...) }  //
func (c *CompilerCmdLine) OptimizeForSize()                      { c.Add("/O1") }                       // Optimize for size.
func (c *CompilerCmdLine) OptimizeForSpeed()                     { c.Add("/O2") }                       // Optimize for speed.
func (c *CompilerCmdLine) EnableInlineExpansion()                { c.Add("/Ob2") }                      // Enable inline expansion for functions that are small and frequently called.
func (c *CompilerCmdLine) EnableIntrinsicFunctions()             { c.Add("/Oi") }                       // Enable intrinsic functions.
func (c *CompilerCmdLine) OmitFramePointer()                     { c.Add("/Oy") }                       // Omit frame pointer for functions that do not require one.
func (c *CompilerCmdLine) UseMultithreadedRuntime()              { c.Add("/MT") }                       // Use the multithreaded version of the C runtime library.
func (c *CompilerCmdLine) EnableExceptionHandling()              { c.Add("/EHsc") }                     // Enable C++ exception handling.
func (c *CompilerCmdLine) DiagnosticsEmitFullPathOfSourceFiles() { c.Add("/FC") }                       // Full path of source files in diagnostics.
func (c *CompilerCmdLine) UseFloatingPointPrecise()              { c.Add("/fp:precise") }               // Use floating-point model: precise
func (c *CompilerCmdLine) UseCpp14()                             { c.Add("/std:c++14") }                // Use C++14 standard.
func (c *CompilerCmdLine) UseCpp17()                             { c.Add("/std:c++17") }                // Use C++17 standard.
func (c *CompilerCmdLine) UseCpp20()                             { c.Add("/std:c++20") }                // Use C++20 standard.
func (c *CompilerCmdLine) UseCppLatest()                         { c.Add("/std:c++latest") }            // Use the latest C++ standard.
func (c *CompilerCmdLine) UseC11()                               { c.Add("/std:c11") }                  // Use C11 standard.
func (c *CompilerCmdLine) UseC17()                               { c.Add("/std:c17") }                  // Use C17 standard.
func (c *CompilerCmdLine) UseCLatest()                           { c.Add("/std:clatest") }              // Use the latest C standard.
func (c *CompilerCmdLine) GenerateDependencyFiles(outputPath string) {
	c.args.Add("/sourceDependencies", outputPath)
} // Generate a dependency file for every source files being compiled.

func GenerateCompilerCmdline(flags CompilerFlags, ouputPath string, includes []string, defines []string, sourceFiles []string, objectFiles []string) *foundation.Arguments {
	args := foundation.NewArguments(len(sourceFiles) + len(objectFiles) + 20)

	c := &CompilerCmdLine{args: args}

	// Common arguments
	c.CompileOnly()
	c.NoLogo()
	c.DiagnosticsColumnMode()
	c.DiagnosticsEmitFullPathOfSourceFiles()
	c.WarningLevel3()
	c.WarningsAreErrors()

	// Debug-specific arguments
	if c.flags.IsDebug() {
		c.DisableOptimizations()
		c.GenerateDebugInfo()
		c.DisableFramePointer()
		c.UseMultithreadedDebugRuntime()
	}

	// Release-specific arguments
	if c.flags.IsRelease() || c.flags.IsFinal() {
		c.OptimizeForSize()
		c.OptimizeForSpeed()
		c.EnableInlineExpansion()
		c.EnableIntrinsicFunctions()
		c.OmitFramePointer()
		c.UseMultithreadedRuntime()
	}

	// Test-specific arguments
	if c.flags.IsTest() {
		c.EnableExceptionHandling()
	}

	// Include paths and defines
	c.Includes(includes)
	c.Defines(defines)

	// More common arguments
	c.UseFloatingPointPrecise()
	c.GenerateDependencyFiles(ouputPath)
	c.BuildMultipleSourceFilesConcurrently()

	// C++ specific arguments
	if c.flags.IsCpp() {
		c.UseCpp17()
	}
	// C specific arguments
	if c.flags.IsC() {
		c.UseC11()
	}

	return args
}
