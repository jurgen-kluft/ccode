package clang

import "github.com/jurgen-kluft/ccode/foundation"

type ArchiverCmdline struct {
	args   *foundation.Arguments
	length int
}

func NewArchiverCmdline(args *foundation.Arguments) *ArchiverCmdline {
	return &ArchiverCmdline{args: args}
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

func (c *ArchiverCmdline) ReplaceCreateSort() { c.Add("-rs") }
func (c *ArchiverCmdline) DynamicLib()        { c.Add("-dynamiclib") }
func (c *ArchiverCmdline) InstallName(outputArchiveFilepath string) {
	c.args.Add("-install_name", outputArchiveFilepath)
}

func (c *ArchiverCmdline) Out() {
}
func (c *ArchiverCmdline) OutputArchiveAndObjectFiles(outputArchiveFilepath string, objs []string) {
    c.args.Add(outputArchiveFilepath)
	c.args.Add(objs...)
}
func (c *ArchiverCmdline) Save()    { c.length = c.args.Len() }
func (c *ArchiverCmdline) Restore() { c.args.Args = c.args.Args[:c.length] }

func GenerateArchiverCmdline(objectFiles []string, outputArchiveFilepath string) *foundation.Arguments {
	args := foundation.NewArguments(len(objectFiles) + 8)

	ac := NewArchiverCmdline(args)

	ac.ReplaceCreateSort()
	ac.OutputArchiveAndObjectFiles(outputArchiveFilepath, objectFiles)

	return args
}
