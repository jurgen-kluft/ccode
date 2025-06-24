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

type LinkerContext struct {
	args  *foundation.Arguments
	flags LinkerFlags // Build configuration
}

func (c *LinkerContext) WhenDebug() bool {
	return c.flags.WhenDebug()
}
func (c *LinkerContext) WhenRelease() bool {
	return c.flags.WhenRelease()
}
func (c *LinkerContext) WhenFinal() bool {
	return c.flags.WhenFinal()
}
func (c *LinkerContext) WhenConsole() bool {
	return c.flags.WhenConsole()
}
func (c *LinkerContext) Add(arg string) {
	c.args.Add(arg)
}
func (c *LinkerContext) AddWithPrefix(prefix string, args ...string) {
	c.args.AddWithPrefix(prefix, args...)
}
func (c *LinkerContext) AddWithFunc(modFunc func(string) string, args ...string) {
	c.args.AddWithFunc(modFunc, args...)
}

func (c *LinkerContext) ErrorReportPrompt()             { c.Add("/ERRORREPORT:PROMPT") }
func (c *LinkerContext) NoLogo()                        { c.Add("/NOLOGO") }
func (c *LinkerContext) GenerateDebugInfo()             { c.Add("/DEBUG") }
func (c *LinkerContext) UseMultithreadedDebug()         { c.Add("/MTd") }
func (c *LinkerContext) UseMultithreaded()              { c.Add("/MT") }
func (c *LinkerContext) OptimizeReferences()            { c.Add("/OPT:REF") }
func (c *LinkerContext) OptimizeIdenticalFolding()      { c.Add("/OPT:ICF") }
func (c *LinkerContext) LinkTimeCodeGeneration()        { c.Add("/LTCG") }
func (c *LinkerContext) DisableIncrementalLinking()     { c.Add("/INCREMENTAL:NO") }
func (c *LinkerContext) UseMultithreadedFinal()         { c.Add("/MT") }
func (c *LinkerContext) SubsystemConsole()              { c.Add("/SUBSYSTEM:CONSOLE") }
func (c *LinkerContext) SubsystemWindows()              { c.Add("/SUBSYSTEM:WINDOWS") }
func (c *LinkerContext) DynamicBase()                   { c.Add("/DYNAMICBASE") }
func (c *LinkerContext) EnableDataExecutionPrevention() { c.Add("/NXCOMPAT") }
func (c *LinkerContext) MachineX64()                    { c.Add("/MACHINE:X64") }
func (c *LinkerContext) LibPaths(libpaths []string) {
	c.AddWithFunc(func(arg string) string { return "/LIBPATH:\"" + foundation.PathWindowsPath(arg) + "\"" }, libpaths...)
}
func (c *LinkerContext) Libs(libs []string) {
	c.AddWithFunc(func(arg string) string { return "\"" + arg + "\"" }, libs...)
}
func (c *LinkerContext) ObjectFiles(objs []string) {
	c.AddWithFunc(func(arg string) string { return "\"" + arg + "\"" }, objs...)
}

func GenerateLinkerCmdline(flags LinkerFlags, libpaths []string, libs []string, objectFiles []string) *foundation.Arguments {
	args := foundation.NewArguments(len(libpaths) + len(libs) + len(objectFiles) + 20)

	c := &LinkerContext{args: args}
	c.flags = flags

	c.ErrorReportPrompt()
	c.NoLogo()
	if c.WhenDebug() {
		c.GenerateDebugInfo()
		c.UseMultithreadedDebug()
	}
	if c.WhenRelease() || c.WhenFinal() {
		c.UseMultithreaded()
		c.OptimizeReferences()
		c.OptimizeIdenticalFolding()
	}
	if c.WhenFinal() {
		c.LinkTimeCodeGeneration()
		c.DisableIncrementalLinking()
		c.UseMultithreadedFinal()
	}
	if c.WhenConsole() {
		c.SubsystemConsole()
	}

	c.DynamicBase()
	c.EnableDataExecutionPrevention()
	c.MachineX64()

	c.LibPaths(libpaths)
	c.Libs(libs)
	c.ObjectFiles(objectFiles)

	return args
}
