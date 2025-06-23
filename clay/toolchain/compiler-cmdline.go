package toolchain

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

type CompilerContext struct {
	args        *foundation.Arguments
	flags       CompilerFlags // Build configuration
	includes    []string
	defines     []string
	outputPath  string
	sourceFiles []string
	objectFiles []string
}

func NewCompilerContext(flags CompilerFlags, args *foundation.Arguments) *CompilerContext {
	return &CompilerContext{
		args:        args,
		flags:       flags,
		includes:    []string{},
		defines:     []string{},
		outputPath:  "",
		sourceFiles: []string{},
		objectFiles: []string{},
	}
}

func (c *CompilerContext) Add(arg string) { c.args.Add(arg) }
func (c *CompilerContext) AddWithPrefix(prefix string, args ...string) {
	c.args.AddWithPrefix(prefix, args...)
}
func (c *CompilerContext) CompileOnly()                          { c.Add("/c") }                                     // Compile only; do not link. This is useful for generating object files without creating an executable.
func (c *CompilerContext) NoLogo()                               { c.Add("/nologo") }                                // Suppress the display of the compiler's startup banner and copyright message.
func (c *CompilerContext) DiagnosticsColumnMode()                { c.Add("/diagnostics:column") }                    // Enable column mode for diagnostics, which provides more detailed error messages.
func (c *CompilerContext) WarningLevel3()                        { c.Add("/W3") }                                    // Set output warning level to 3 (high warnings).
func (c *CompilerContext) WarningsAreErrors()                    { c.Add("/WX") }                                    // Treat warnings as errors.
func (c *CompilerContext) BuildMultipleSourceFilesConcurrently() { c.Add("/MP") }                                    // Build multiple source files concurrently.
func (c *CompilerContext) DisableOptimizations()                 { c.Add("/Od") }                                    // Disable optimizations for debugging.
func (c *CompilerContext) GenerateDebugInfo()                    { c.Add("/Zi") }                                    // Generate complete debugging information.
func (c *CompilerContext) DisableFramePointer()                  { c.Add("/Oy-") }                                   // Do not omit frame pointer.
func (c *CompilerContext) UseMultithreadedDebugRuntime()         { c.Add("/MTd") }                                   // Use the multithreaded debug version of the C runtime library.
func (c *CompilerContext) IncludesForMsdev()                     { c.AddWithPrefix("/I", c.includes...) }            //
func (c *CompilerContext) DefinesForMsdev()                      { c.AddWithPrefix("/D", c.defines...) }             //
func (c *CompilerContext) OptimizeForSize()                      { c.Add("/O1") }                                    // Optimize for size.
func (c *CompilerContext) OptimizeForSpeed()                     { c.Add("/O2") }                                    // Optimize for speed.
func (c *CompilerContext) EnableInlineExpansion()                { c.Add("/Ob2") }                                   // Enable inline expansion for functions that are small and frequently called.
func (c *CompilerContext) EnableIntrinsicFunctions()             { c.Add("/Oi") }                                    // Enable intrinsic functions.
func (c *CompilerContext) OmitFramePointer()                     { c.Add("/Oy") }                                    // Omit frame pointer for functions that do not require one.
func (c *CompilerContext) UseMultithreadedRuntime()              { c.Add("/MT") }                                    // Use the multithreaded version of the C runtime library.
func (c *CompilerContext) EnableExceptionHandling()              { c.Add("/EHsc") }                                  // Enable C++ exception handling.
func (c *CompilerContext) DiagnosticsEmitFullPathOfSourceFiles() { c.Add("/FC") }                                    // Full path of source files in diagnostics.
func (c *CompilerContext) UseFloatingPointPrecise()              { c.Add("/fp:precise") }                            // Use floating-point model: precise
func (c *CompilerContext) UseCpp14()                             { c.Add("/std:c++14") }                             // Use C++14 standard.
func (c *CompilerContext) UseCpp17()                             { c.Add("/std:c++17") }                             // Use C++17 standard.
func (c *CompilerContext) UseCpp20()                             { c.Add("/std:c++20") }                             // Use C++20 standard.
func (c *CompilerContext) UseCppLatest()                         { c.Add("/std:c++latest") }                         // Use the latest C++ standard.
func (c *CompilerContext) UseC11()                               { c.Add("/std:c11") }                               // Use C11 standard.
func (c *CompilerContext) UseC17()                               { c.Add("/std:c17") }                               // Use C17 standard.
func (c *CompilerContext) UseCLatest()                           { c.Add("/std:clatest") }                           // Use the latest C standard.
func (c *CompilerContext) GenerateDependencyFiles()              { c.args.Add("/sourceDependencies", c.outputPath) } // Generate a dependency file for every source files being compiled.

func (c *CompilerContext) GenerateMsdevCmdline() {
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
	c.IncludesForMsdev()
	c.DefinesForMsdev()

	// More common arguments
	c.UseFloatingPointPrecise()
	c.GenerateDependencyFiles()
	c.BuildMultipleSourceFilesConcurrently()

	// C++ specific arguments
	if c.flags.IsCpp() {
		c.UseCpp17()
	}
	// C specific arguments
	if c.flags.IsC() {
		c.UseC11()
	}
}

func (c *CompilerContext) ClangCompileOnly()                          { c.Add("-c") }                          // Compile only; do not link.
func (c *CompilerContext) ClangNoLogo()                               { c.Add("-nologo") }                     // Suppress the display of the compiler's startup banner and copyright message.
func (c *CompilerContext) ClangWarningLevel3()                        {}                                       // Set output warning level to 3 (high warnings).
func (c *CompilerContext) ClangWarningsAreErrors()                    { c.Add("-Werror") }                     // Treat warnings as errors.
func (c *CompilerContext) ClangBuildMultipleSourceFilesConcurrently() {}                                       // Build multiple source files concurrently.
func (c *CompilerContext) ClangDisableOptimizations()                 { c.Add("-O0") }                         // Disable optimizations for debugging.
func (c *CompilerContext) ClangGenerateDebugInfo()                    { c.Add("-g") }                          // Generate complete debugging information.
func (c *CompilerContext) ClangDisableFramePointer()                  { c.Add("-fno-omit-frame-pointer") }     // Do not omit frame pointer.
func (c *CompilerContext) ClangIncludesForClang()                     { c.AddWithPrefix("-I", c.includes...) } // Add include paths.
func (c *CompilerContext) ClangDefinesForClang()                      { c.AddWithPrefix("-D", c.defines...) }  // Add preprocessor definitions.
func (c *CompilerContext) ClangOptimizeForSize()                      { c.Add("-Os") }                         // Optimize for size.
func (c *CompilerContext) ClangOptimizeForSpeed()                     { c.Add("-O2") }                         // Optimize for speed.
func (c *CompilerContext) ClangOptimizeHard()                         { c.Add("-O3") }                         // Optimize hard, enabling aggressive optimizations.
func (c *CompilerContext) ClangEnableInlineExpansion()                { c.Add("-finline-functions") }          // Enable inline expansion for functions that are small and frequently called.
func (c *CompilerContext) ClangEnableIntrinsicFunctions()             { c.Add("-ffunction-sections") }         // Enable intrinsic functions.
func (c *CompilerContext) ClangOmitFramePointer()                     { c.Add("-fomit-frame-pointer") }        // Omit frame pointer for functions that do not require one.
func (c *CompilerContext) ClangUseMultithreadedRuntime()              {}                                       // Use the multithreaded version of the C runtime library.
func (c *CompilerContext) ClangEnableExceptionHandling()              { c.Add("-fexceptions") }                // Enable C++ exception handling.
func (c *CompilerContext) ClangDiagnosticsEmitFullPathOfSourceFiles() {}                                       // Full path of source files in diagnostics.
func (c *CompilerContext) ClangUseFloatingPointPrecise()              { c.Add("-ffp-model=precise") }          // Use floating-point model: precise
func (c *CompilerContext) ClangUseCpp14()                             { c.Add("-std=c++14") }                  // Use C++14 standard.
func (c *CompilerContext) ClangUseCpp17()                             { c.Add("-std=c++17") }                  // Use C++17 standard.
func (c *CompilerContext) ClangUseCpp20()                             { c.Add("-std=c++20") }                  // Use C++20 standard.
func (c *CompilerContext) ClangUseCppLatest()                         { c.Add("-std=c++latest") }              // Use the latest C++ standard.
func (c *CompilerContext) ClangUseC11()                               { c.Add("-std=c11") }                    // Use C11 standard.
func (c *CompilerContext) ClangUseC17()                               { c.Add("-std=c17") }                    // Use C17 standard.
func (c *CompilerContext) ClangUseCLatest()                           { c.Add("-std=clatest") }                // Use the latest C standard.
func (c *CompilerContext) ClangGenerateDependencyFiles()              { c.Add("-MMD") }                        // Generate a dependency file for every source files being compiled.

func (c *CompilerContext) GenerateClangCmdline() {
	c.ClangCompileOnly()
	c.ClangNoLogo()
	c.ClangWarningLevel3()
	c.ClangWarningsAreErrors()
	c.ClangBuildMultipleSourceFilesConcurrently()
	// Debug-specific arguments
	if c.flags.IsDebug() {
		c.ClangDisableOptimizations()
		c.ClangGenerateDebugInfo()
		c.ClangDisableFramePointer()
	}
	c.ClangIncludesForClang()
	c.ClangDefinesForClang()
	// Release and Final specific arguments
	if c.flags.IsRelease() || c.flags.IsFinal() {
		c.ClangOptimizeForSize()
		c.ClangOptimizeForSpeed()
		c.ClangOptimizeHard()
		c.ClangEnableInlineExpansion()
		c.ClangEnableIntrinsicFunctions()
		c.ClangOmitFramePointer()
		c.ClangUseMultithreadedRuntime()
	}
	// Test-specific arguments
	if c.flags.IsTest() {
		c.ClangEnableExceptionHandling()
	}
	c.ClangDiagnosticsEmitFullPathOfSourceFiles()
	c.ClangUseFloatingPointPrecise()
	// C++ specific arguments
	if c.flags.IsCpp() {
		c.ClangUseCpp17()
	}
	// C specific arguments
	if c.flags.IsC() {
		c.ClangUseC11()
	}
	c.ClangGenerateDependencyFiles()
}
