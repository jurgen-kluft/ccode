package msvc

import (
	corepkg "github.com/jurgen-kluft/ccode/core"
)

type ArchiverCmdline struct {
	args   *corepkg.Arguments
	length int
}

func NewArchiverCmdline(args *corepkg.Arguments) *ArchiverCmdline {
	return &ArchiverCmdline{
		args:   args,
		length: 0,
	}
}

func (c *ArchiverCmdline) Add(arg string) {
	c.args.Add(arg)
}
func (c *ArchiverCmdline) AddWithPrefix(prefix string, args ...string) {
	c.args.AddWithPrefix(prefix, args...)
}
func (c *ArchiverCmdline) AddWithFunc(modFunc func(string) string, args ...string) {
	c.args.AddWithFunc(modFunc, args...)
}

func (c *ArchiverCmdline) NoLogo()     { c.Add("/NOLOGO") }      // Suppress display of sign-on banner.
func (c *ArchiverCmdline) MachineX64() { c.Add("/MACHINE:X64") } // Specify the target machine architecture (x64).
func (c *ArchiverCmdline) Out(outputArchiveFilepath string) {
	c.AddWithPrefix("/OUT:", outputArchiveFilepath)
}
func (c *ArchiverCmdline) ObjectFiles(objs []string) {
	c.AddWithFunc(func(arg string) string { return corepkg.PathWindowsPath(arg) }, objs...)
}

func (c *ArchiverCmdline) Save() {
	c.length = c.args.Len()
}
func (c *ArchiverCmdline) Restore() {
	if c.length < c.args.Len() {
		c.args.Args = c.args.Args[:c.length]
	}
}

func GenerateArchiverCmdline(inputObjectFilepaths []string, outputArchiveFilepath string) *corepkg.Arguments {
	args := corepkg.NewArguments(len(inputObjectFilepaths) + 8)

	ac := NewArchiverCmdline(args)

	ac.NoLogo()
	ac.MachineX64()
	ac.Out(outputArchiveFilepath)
	ac.ObjectFiles(inputObjectFilepaths)

	return args
}
