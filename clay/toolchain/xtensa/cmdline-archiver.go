package xtensa

import "github.com/jurgen-kluft/ccode/foundation"

type ArchiverContext struct {
	args *foundation.Arguments
}

func (c ArchiverContext) CreateReplace() { c.args.Add("cr") }
func (c ArchiverContext) Out(outputArchiveFilepath string) {
	c.args.Add(outputArchiveFilepath)
}
func (c ArchiverContext) ObjectFiles(objs []string) {
	c.args.Add(objs...)
}

func GenerateArchiverCmdline(objectFiles []string, outputArchiveFilepath string) *foundation.Arguments {
	args := foundation.NewArguments(len(objectFiles) + 8)

	ac := ArchiverContext{args: args}

	ac.CreateReplace()
	ac.Out(outputArchiveFilepath)
	ac.ObjectFiles(objectFiles)

	return args
}
