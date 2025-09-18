package xtensa

import (
	corepkg "github.com/jurgen-kluft/ccode/core"
)

type ArchiverContext struct {
	args *corepkg.Arguments
}

func (c ArchiverContext) CreateReplace() { c.args.Add("cr") }
func (c ArchiverContext) Out(outputArchiveFilepath string) {
	c.args.Add(outputArchiveFilepath)
}
func (c ArchiverContext) ObjectFiles(objs []string) {
	c.args.Add(objs...)
}

func GenerateArchiverCmdline(objectFiles []string, outputArchiveFilepath string) *corepkg.Arguments {
	args := corepkg.NewArguments(len(objectFiles) + 8)

	ac := ArchiverContext{args: args}

	ac.CreateReplace()
	ac.Out(outputArchiveFilepath)
	ac.ObjectFiles(objectFiles)

	return args
}
