package toolchain

var MsDevOptions = map[string][]string{
	"dynamic.debugging":              {"/dynamicdeopt"},       // Enable dynamic debugging; allows for dynamic analysis of the program.
	"optimize.for.size":              {"/O1"},                 // Optimize for size; enables optimizations that reduce code size.
	"optimize.for.speed":             {"/O2"},                 // Optimize for speed; enables most optimizations.
	"optimize.none":                  {"/Od"},                 // No optimizations; useful for debugging.
	"optimize.global":                {"/Og"},                 // Global optimizations, including inlining and loop unrolling.
	"optimize.full":                  {"/Ox"},                 // Full optimization, including speed and size.
	"optimize.favor.small":           {"/Os"},                 // Optimize for size, not speed.
	"optimize.favor.fast":            {"/Ot"},                 // Optimize for speed, not size.
	"generate.debug.info":            {"/Z7", "/Zi"},          // Generate debug information; `/Z7` for embedded, `/Zi` for separate PDB.
	"generate.intrinsic.functions":   {"/Oi"},                 // Enable intrinsic functions (e.g., `_memcpy`, `_memset`).
	"omit.frame.pointer":             {"/Oy"},                 // Omit frame pointer for functions that do not require one.
	"inline.expansion.level.0":       {"/Ob0"},                // Disable inline expansion; functions are not inlined.
	"inline.expansion.level.1":       {"/Ob1"},                // Enable inline expansion for functions that are small and frequently called.
	"inline.expansion.level.2":       {"/Ob2"},                // Enable inline expansion for functions that are small and frequently called, with more aggressive inlining.
	"inline.expansion.level.3":       {"/Ob3"},                // Enable inline expansion for functions that are small and frequently called, with the most aggressive inlining.
	"exception.handling.std":         {"/EHsc"},               // Default: enable C++ exception handling with standard semantics.
	"exception.handling.async":       {"/EHa"},                // Enable C++ exception handling (with SEH exceptions).
	"exception.handling.c":           {"/EHc"},                // `extern "C"` defaults to `nothrow`.
	"exception.handling.r":           {"/EHr"},                // Always generate `noexcept` runtime termination checks.
	"exception.handling.s":           {"/EHs"},                // Enable C++ exception handling (no SEH exceptions).
	"fp.behavior.contract":           {"/fp:contract"},        // Consider floating-point contractions when generating code.
	"fp.behavior.except":             {"/fp:except"},          // Consider floating-point exceptions when generating code.
	"fp.behavior.fast":               {"/fp:fast"},            // "fast" floating-point model; results are less predictable.
	"fp.behavior.precise":            {"/fp:precise"},         // "precise" floating-point model; results are predictable.
	"fp.behavior.strict":             {"/fp:strict"},          // "strict" floating-point model (implies `/fp:except`).
	"string.pooling":                 {"/GF"},                 // Enable string pooling to reduce memory usage.
	"whole.program.optimization":     {"/Gw"},                 // Enable whole program optimization.
	"rtti.enable":                    {"/GR"},                 // Enable run-time type information (RTTI).
	"rtti.disable":                   {"/GR-"},                // Disable run-time type information (RTTI).
	"function.level.linking.enable":  {"/Gy"},                 // Enable function-level linking.
	"function.level.linking.disable": {"/Gy-"},                // Disable function-level linking.
	"map.filepath":                   {"/Fm"},                 // Create a map file.
	"exe.filepath":                   {"/Fe"},                 // Specify the output executable file path.
	"pdb.filepath":                   {"/Fd"},                 // Specify the output program database (PDB) file path.
	"compiler.defines":               {"/D"},                  // Define preprocessor macros.
	"compiler.includes":              {"/I"},                  // Specify additional include directories.
	"generate.dependency.files":      {"/sourceDependencies"}, // Generate source-level dependency files.
	"build.concurrently":             {"/MP"},                 // Build multiple source files concurrently.
	"compile.all.as.c":               {"/TC"},                 // Treat all source files as C.
	"compile.all.as.cpp":             {"/TP"},                 // Treat all source files
	"warnings.disable.all":           {"/w"},                  // Disable all warnings.
	"warnings.enable.all":            {"/Wall"},               // Enable all warnings, including those disabled by default.
	"warnings.are.errors":            {"/WX"},                 // Treat all warnings as errors.
	"warnings.output.level.0":        {"/W0"},                 // Set output warning level to 0 (no warnings).
	"warnings.output.level.1":        {"/W1"},                 // Set output warning level to 1 (basic warnings).
	"warnings.output.level.2":        {"/W2"},                 // Set output warning level to 2 (moderate warnings).
	"warnings.output.level.3":        {"/W3"},                 // Set output warning level to 3 (high warnings).
	"warnings.output.level.4":        {"/W4"},                 // Set output warning level to 4 (very high warnings).
	"c++14":                          {"/std:c++14"},          // Specify the C++ standard version (c++14, c++17, c++20, c++latest).
	"c++17":                          {"/std:c++17"},          // Specify the C++ standard version (c++14, c++17, c++20, c++latest).
	"c++20":                          {"/std:c++20"},          // Specify the C++ standard version (c++14, c++17, c++20, c++latest).
	"c++latest":                      {"/std:c++latest"},      // Specify the C++ standard version (c++14, c++17, c++20, c++latest).
	"c11":                            {"/std:c11"},            // Specify the C standard version (c11, c17, clatest).
	"c17":                            {"/std:c17"},            // Specify the C standard version (c11, c17, clatest).
	"clatest":                        {"/std:clatest"},        // Specify the C standard version (c11, c17, clatest).
	"link.dll":                       {"/LD"},                 // Generate a dynamic-link library (DLL).
	"link.debug.dll":                 {"/LDd"},                // Generate a debug dynamic
	"link.multithreaded.dll":         {"/MD"},                 // Generate a multithreaded DLL, by using *MSVCRT.lib*
	"link.multithreaded.debug.dll":   {"/MDd"},                // Generate a debug multithreaded DLL, by using *MSVCRTD.lib*
	"link.multithreaded.exe":         {"/MT"},                 // Generate a multithreaded executable file, by using *LIBCMT.lib*
	"link.multithreaded.debug.exe":   {"/MTd"},                // Generate a debug multithreaded executable file, by using *LIBCMTD.lib*
}
