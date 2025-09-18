package xtensa

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

type LinkerCmdLine struct {
	args           *corepkg.Arguments
	flags          LinkerFlags // Build configuration
	arduinoSdkPath string
}

func (c *LinkerCmdLine) WhenDebug() bool {
	return c.flags.WhenDebug()
}
func (c *LinkerCmdLine) WhenRelease() bool {
	return c.flags.WhenRelease()
}
func (c *LinkerCmdLine) WhenFinal() bool {
	return c.flags.WhenFinal()
}
func (c *LinkerCmdLine) WhenConsole() bool {
	return c.flags.WhenConsole()
}
func (c *LinkerCmdLine) Add(arg string) {
	c.args.Add(arg)
}
func (c *LinkerCmdLine) AddWithPrefix(prefix string, args ...string) {
	c.args.AddWithPrefix(prefix, args...)
}
func (c *LinkerCmdLine) AddWithFunc(modFunc func(string) string, args ...string) {
	c.args.AddWithFunc(modFunc, args...)
}

func (c *LinkerCmdLine) ResponseFlags()                   { c.Add(`@` + c.arduinoSdkPath + `/flags/ld_flags`) }
func (c *LinkerCmdLine) ResponseScripts()                 { c.Add(`@` + c.arduinoSdkPath + `/flags/ld_scripts`) }
func (c *LinkerCmdLine) ResponseLibs()                    { c.Add(`@` + c.arduinoSdkPath + `/flags/ld_libs`) }
func (c *LinkerCmdLine) GenerateMapfile(_filepath string) { c.Add("-Wl,--Map=" + _filepath) }
func (c *LinkerCmdLine) SystemLibraryPaths() {
	c.AddWithPrefix("-L", c.arduinoSdkPath+"/lib", c.arduinoSdkPath+"/ld")
}
func (c *LinkerCmdLine) UserLibraryPaths(libraryPaths ...string) {
	c.AddWithPrefix("-L", libraryPaths...)
}
func (c *LinkerCmdLine) AddPanicHandler() { c.Add("-Wl,--wrap=esp_panic_handler") }
func (c *LinkerCmdLine) StartGroup()      { c.Add("-Wl,--start-group") }
func (c *LinkerCmdLine) EndGroup()        { c.Add("-Wl,--end-group") }
func (c *LinkerCmdLine) UseElfFormat()    { c.Add("-Wl,-EL") } // Use ELF format for the output file

func (c *LinkerCmdLine) UserLibraryFiles(libraryFiles []string) {
	c.args.Add(libraryFiles...)
}
func (c *LinkerCmdLine) UserObjectFiles(objectFiles []string) {
	c.args.Add(objectFiles...)
}
func (c *LinkerCmdLine) Out(outputAppRelFilepathNoExt string) {
	c.args.Add("-o", outputAppRelFilepathNoExt)
}

func GenerateLinkerCmdline(flags LinkerFlags, arduinoSdkPath string, libpaths []string, libs []string, objectFiles []string, outputAppRelFilepathNoExt string) *corepkg.Arguments {
	args := corepkg.NewArguments(len(libpaths) + len(libs) + len(objectFiles) + 20)

	c := &LinkerCmdLine{args: args, flags: flags, arduinoSdkPath: arduinoSdkPath}

	c.GenerateMapfile(outputAppRelFilepathNoExt + ".map")
	c.SystemLibraryPaths()
	c.UserLibraryPaths(libpaths...)
	c.AddPanicHandler()
	c.ResponseFlags()
	c.ResponseScripts()

	c.StartGroup()
	{
		c.UserLibraryFiles(libs)
		c.ResponseLibs()
		c.UserObjectFiles(objectFiles)
	}
	c.EndGroup()

	c.UseElfFormat()
	c.Out(outputAppRelFilepathNoExt)

	return args
}
