package msvc

import "github.com/jurgen-kluft/ccode/foundation"

type ArchiverContext struct {
	args *foundation.Arguments
}

func (c ArchiverContext) Add(arg string) {
	c.args.Add(arg)
}
func (c ArchiverContext) AddWithPrefix(prefix string, args ...string) {
	c.args.AddWithPrefix(prefix, args...)
}
func (c ArchiverContext) AddWithFunc(modFunc func(string) string, args ...string) {
	c.args.AddWithFunc(modFunc, args...)
}

func (c ArchiverContext) NoLogo()     { c.Add("/NOLOGO") }      // Suppress display of sign-on banner.
func (c ArchiverContext) MachineX64() { c.Add("/MACHINE:X64") } // Specify the target machine architecture (x64).
func (c ArchiverContext) Out(outputArchiveFilepath string) {
	c.AddWithPrefix("/OUT:", outputArchiveFilepath)
}
func (c ArchiverContext) ObjectFiles(objs []string) {
	c.AddWithFunc(func(arg string) string { return "\"" + arg + "\"" }, objs...)
}

func GenerateArchiverCmdline(inputObjectFilepaths []string, outputArchiveFilepath string) *foundation.Arguments {
	args := foundation.NewArguments(len(inputObjectFilepaths) + 8)

	ac := ArchiverContext{args: args}

	ac.NoLogo()
	ac.MachineX64()
	ac.Out(outputArchiveFilepath)
	ac.ObjectFiles(inputObjectFilepaths)

	return args
}
