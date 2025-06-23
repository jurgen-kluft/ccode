package toolchain

import "github.com/jurgen-kluft/ccode/foundation"

// Playground:
// Let's see if we can come up with a declarative way of configuring a compiler commandline.

// Current status:
// What we have at the moment is pretty nice and clear to the user

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
	args        *foundation.Arguments
	flags       LinkerFlags // Build configuration
	outputPath  string
	libpaths    []string
	libs        []string
	objectFiles []string
}

func NewLinkerContext(flags LinkerFlags, args *foundation.Arguments) *LinkerContext {
	return &LinkerContext{
		args:        args,
		flags:       flags,
		outputPath:  "",
		libs:        []string{},
		libpaths:    []string{},
		objectFiles: []string{},
	}
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
func (c *LinkerContext) LibPaths() {
	c.AddWithFunc(func(arg string) string { return "/LIBPATH:\"" + foundation.PathWindowsPath(arg) + "\"" }, c.libpaths...)
}
func (c *LinkerContext) Libs() {
	c.AddWithFunc(func(arg string) string { return "\"" + arg + "\"" }, c.libs...)
}

func (c *LinkerContext) GenerateMsdevCmdline() {
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

	c.LibPaths()
	c.Libs()
}
