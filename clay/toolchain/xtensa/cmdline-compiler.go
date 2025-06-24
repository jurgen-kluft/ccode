package xtensa

import (
	"runtime"

	"github.com/jurgen-kluft/ccode/foundation"
)

type CompilerFlags uint64

const (
	CompilerFlagDebug CompilerFlags = 1 << iota
	CompilerFlagRelease
	CompilerFlagFinal
	CompilerFlagTest
	CompilerFlagCpp
)

func (c CompilerFlags) IsCpp() bool {
	return c&CompilerFlagCpp != 0
}
func (c CompilerFlags) IsC() bool {
	return c&CompilerFlagCpp == 0
}

type CompilerCmdLine struct {
	args           *foundation.Arguments
	flags          CompilerFlags // Build configuration
	arduinoSdkPath string
	espMcu         string
}

func NewCompilerContext(flags CompilerFlags, args *foundation.Arguments) *CompilerCmdLine {
	return &CompilerCmdLine{
		args:  args,
		flags: flags,
	}
}

func (c *CompilerCmdLine) Add(arg string)        { c.args.Add(arg) }
func (c *CompilerCmdLine) AddMany(arg ...string) { c.args.Add(arg...) }
func (c *CompilerCmdLine) AddWithPrefix(prefix string, args ...string) {
	c.args.AddWithPrefix(prefix, args...)
}

func (c *CompilerCmdLine) CompileOnly()     { c.Add("-c") }
func (c *CompilerCmdLine) GenerateMapfile() { c.Add("-MMD") }
func (c *CompilerCmdLine) ResponseFlags() {
	if c.flags.IsC() {
		c.Add(`@` + c.arduinoSdkPath + `/flags/c_flags`)
	} else {
		c.Add(`@` + c.arduinoSdkPath + `/flags/cpp_flags`)
	}
}
func (c *CompilerCmdLine) ResponseDefines()  { c.Add(`@` + c.arduinoSdkPath + `/flags/defines`) }
func (c *CompilerCmdLine) ResponseIncludes() { c.Add(`@` + c.arduinoSdkPath + `/flags/includes`) }
func (c *CompilerCmdLine) CompilerSwitches() { c.AddMany("-w", "-Os") }
func (c *CompilerCmdLine) WarningSwitches()  { c.Add("-Werror=return-type") }

func (c *CompilerCmdLine) SystemDefines() {
	c.Add(`-DF_CPU=240000000L`)
	c.Add(`-DARDUINO=10605`)
	c.Add(`-DARDUINO_ESP32_DEV`)
	c.Add(`-DARDUINO_ARCH_ESP32`)
	c.Add(`-DARDUINO_BOARD="ESP32_DEV"`)
	c.Add(`-DARDUINO_VARIANT="` + c.espMcu + `"`)
	c.Add(`-DARDUINO_PARTITION_default`)
	c.Add(`-DARDUINO_HOST_OS="` + runtime.GOOS + `"`)
	c.Add(`-DARDUINO_FQBN="generic"`)
	c.Add(`-DESP32=ESP32`)
	c.Add(`-DCORE_DEBUG_LEVEL=0`)
	c.Add(`-DARDUINO_USB_CDC_ON_BOOT=0`)

}
func (c *CompilerCmdLine) Defines(defines []string) {
	c.args.AddWithPrefix("-D", defines...)
}
func (c *CompilerCmdLine) SystemPrefixInclude() { c.AddMany(`-iprefix`, c.arduinoSdkPath+`/include/`) }
func (c *CompilerCmdLine) SystemIncludes() {
	c.AddMany(
		c.arduinoSdkPath+`/cores/esp32`,
		c.arduinoSdkPath+`/variants/`+c.espMcu,
	)
}
func (c *CompilerCmdLine) Includes(includes []string) {
	c.args.AddWithPrefix("-I", includes...)
}

func GenerateCompilerCmdline(flags CompilerFlags, includes []string, defines []string, sourceFiles []string, objectFiles []string) *foundation.Arguments {
	args := foundation.NewArguments(64)

	c := NewCompilerContext(flags, args)
	c.CompileOnly()
	c.GenerateMapfile()
	c.ResponseFlags()
	c.CompilerSwitches()
	c.WarningSwitches()
	c.SystemDefines()
	c.Defines(defines)
	c.ResponseDefines()
	c.SystemPrefixInclude()
	c.ResponseIncludes()
	c.SystemIncludes()
	c.Includes(includes)

	return args
}
