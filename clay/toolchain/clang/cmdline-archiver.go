package clang

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

func (c ArchiverContext) ReplaceCreateSort() { c.Add("-rcs") }
func (c ArchiverContext) Out(outputArchiveFilepath string) {
	c.Add(outputArchiveFilepath)
}
func (c ArchiverContext) ObjectFiles(objs []string) {
	c.args.Add(objs...)
}

func GenerateArchiverCmdline(objectFiles []string, outputArchiveFilepath string) *foundation.Arguments {
	args := foundation.NewArguments(len(objectFiles) + 8)

	ac := ArchiverContext{args: args}

	ac.ReplaceCreateSort()
	ac.Out(outputArchiveFilepath)
	ac.ObjectFiles(objectFiles)

	return args
}
