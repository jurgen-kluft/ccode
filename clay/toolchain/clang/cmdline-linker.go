package clang

import (
	corepkg "github.com/jurgen-kluft/ccode/core"
)

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
	args   *corepkg.Arguments
	length int
}

func NewLinkerCmdline(args *corepkg.Arguments) *LinkerCmdline {
	return &LinkerCmdline{args: args, length: 0}
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

func (c *LinkerCmdline) ErrorReportPrompt()             {}
func (c *LinkerCmdline) NoLogo()                        {}
func (c *LinkerCmdline) GenerateDebugInfo()             {}
func (c *LinkerCmdline) UseMultithreadedDebug()         {}
func (c *LinkerCmdline) UseMultithreaded()              {}
func (c *LinkerCmdline) OptimizeReferences()            {}
func (c *LinkerCmdline) OptimizeIdenticalFolding()      {}
func (c *LinkerCmdline) LinkTimeCodeGeneration()        {}
func (c *LinkerCmdline) DisableIncrementalLinking()     {}
func (c *LinkerCmdline) UseMultithreadedFinal()         {}
func (c *LinkerCmdline) SubsystemConsole()              {}
func (c *LinkerCmdline) SubsystemWindows()              {}
func (c *LinkerCmdline) DynamicBase()                   {}
func (c *LinkerCmdline) EnableDataExecutionPrevention() {}
func (c *LinkerCmdline) MachineX64()                    {}
func (c *LinkerCmdline) LibPaths(libpaths []string)     { c.AddWithPrefix("-L", libpaths...) }
func (c *LinkerCmdline) Libs(libs []string)             { c.AddWithPrefix("-l", libs...) }
func (c *LinkerCmdline) Frameworks(frameworks []string) { c.AddWithPrefix("-framework", frameworks...) }
func (c *LinkerCmdline) ObjectFiles(objs []string)      { c.args.Add(objs...) }
func (c *LinkerCmdline) Out(outputFilepath string)      { c.args.Add("-o", outputFilepath) }
func (c *LinkerCmdline) Save()                          { c.length = c.args.Len() }
func (c *LinkerCmdline) Restore()                       { c.args.Args = c.args.Args[:c.length] }

func GenerateLinkerCmdline(flags LinkerFlags, libpaths []string, libs []string, objectFiles []string) *corepkg.Arguments {
	args := corepkg.NewArguments(len(libpaths) + len(libs) + len(objectFiles) + 20)

	c := &LinkerCmdline{args: args}

	c.ErrorReportPrompt()
	c.NoLogo()

	if flags.WhenDebug() {
		c.GenerateDebugInfo()
		c.UseMultithreadedDebug()
	}
	if flags.WhenRelease() || flags.WhenFinal() {
		c.UseMultithreaded()
		c.OptimizeReferences()
		c.OptimizeIdenticalFolding()
	}
	if flags.WhenFinal() {
		c.LinkTimeCodeGeneration()
		c.DisableIncrementalLinking()
		c.UseMultithreadedFinal()
	}
	if flags.WhenConsole() {
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
