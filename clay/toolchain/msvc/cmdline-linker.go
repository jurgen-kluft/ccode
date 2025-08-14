package msvc

import "github.com/jurgen-kluft/ccode/foundation"

type LinkerFlags uint64

const (
	LinkerFlagDebug LinkerFlags = 1 << iota
	LinkerFlagRelease
	LinkerFlagFinal
	LinkerFlagConsole
)

func (f LinkerFlags) WhenDebug() bool {
	return f&LinkerFlagDebug != 0
}
func (f LinkerFlags) WhenRelease() bool {
	return f&LinkerFlagRelease != 0
}
func (f LinkerFlags) WhenFinal() bool {
	return f&LinkerFlagFinal != 0
}
func (f LinkerFlags) WhenConsole() bool {
	return f&LinkerFlagConsole != 0
}

type LinkerCmdline struct {
	args   *foundation.Arguments
	length int
}

func NewLinkerCmdline(args *foundation.Arguments) *LinkerCmdline {
	return &LinkerCmdline{
		args: args,
	}
}

func (c *LinkerCmdline) Add(arg string) {
	c.args.Add(arg)
}
func (c *LinkerCmdline) AddWithPrefix(prefix string, args ...string) {
	c.args.AddWithPrefix(prefix, args...)
}
func (c *LinkerCmdline) AddWithFunc(modFunc func(string) string, args ...string) {
	c.args.AddWithFunc(modFunc, args...)
}

// Linker options
//	- https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/linker-options.md

func (c *LinkerCmdline) ErrorReportPrompt()             { c.Add("/ERRORREPORT:PROMPT") }
func (c *LinkerCmdline) NoLogo()                        { c.Add("/NOLOGO") }
func (c *LinkerCmdline) GenerateMapfile(fp string)      { c.Add("/MAP:" + fp) }
func (c *LinkerCmdline) GenerateDebugInfo()             { c.Add("/DEBUG") }
func (c *LinkerCmdline) OptimizeReferences()            { c.Add("/OPT:REF") }
func (c *LinkerCmdline) OptimizeIdenticalFolding()      { c.Add("/OPT:ICF") }
func (c *LinkerCmdline) LinkTimeCodeGeneration()        { c.Add("/LTCG") }
func (c *LinkerCmdline) DisableIncrementalLinking()     { c.Add("/INCREMENTAL:NO") }
func (c *LinkerCmdline) SubsystemConsole()              { c.Add("/SUBSYSTEM:CONSOLE") }
func (c *LinkerCmdline) SubsystemWindows()              { c.Add("/SUBSYSTEM:WINDOWS") }
func (c *LinkerCmdline) DynamicBase()                   { c.Add("/DYNAMICBASE") }
func (c *LinkerCmdline) EnableDataExecutionPrevention() { c.Add("/NXCOMPAT") }
func (c *LinkerCmdline) MachineX64()                    { c.Add("/MACHINE:X64") }
func (c *LinkerCmdline) LibPaths(libpaths []string) {
	c.AddWithFunc(func(arg string) string { return "/LIBPATH:" + foundation.PathWindowsPath(arg) }, libpaths...)
}
func (c *LinkerCmdline) Libs(libs []string) {
	c.AddWithFunc(func(arg string) string { return arg }, libs...)
}
func (c *LinkerCmdline) ObjectFiles(objs []string) {
	c.AddWithFunc(func(arg string) string { return "\"" + arg + "\"" }, objs...)
}
func (c *LinkerCmdline) Out(outputFilepath string) {
	c.Add("/OUT:" + outputFilepath)
}
func (c *LinkerCmdline) Save() { c.length = c.args.Len() }
func (c *LinkerCmdline) Restore() {
	if c.length < c.args.Len() {
		c.args.Args = c.args.Args[:c.length]
	}
}

func GenerateLinkerCmdline(flags LinkerFlags, libpaths []string, libs []string, objectFiles []string) *foundation.Arguments {
	args := foundation.NewArguments(len(libpaths) + len(libs) + len(objectFiles) + 20)

	c := &LinkerCmdline{args: args}

	c.ErrorReportPrompt()
	c.NoLogo()
	if flags.WhenDebug() {
		c.GenerateDebugInfo()
	}
	if flags.WhenRelease() || flags.WhenFinal() {
		c.OptimizeReferences()
		c.OptimizeIdenticalFolding()
	}
	if flags.WhenFinal() {
		c.LinkTimeCodeGeneration()
		c.DisableIncrementalLinking()
	}
	if flags.WhenConsole() {
		c.SubsystemConsole()
	}

	c.DynamicBase()
	c.EnableDataExecutionPrevention()
	c.MachineX64()

	c.GenerateMapfile("clay.exe.map")
	c.Out("clay.exe")
	c.LibPaths(libpaths)
	c.Libs(libs)
	c.ObjectFiles(objectFiles)

	return args
}
