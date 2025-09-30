package denv

import corepkg "github.com/jurgen-kluft/ccode/core"

type PinnedGlobPath struct {
	Path PinnedPath
	Glob string
}

func (fp PinnedGlobPath) EncodeJson(encoder *corepkg.JsonEncoder, key string) {
	encoder.BeginObject(key)
	{
		encoder.WriteField("root", fp.Path.Root)
		encoder.WriteField("base", fp.Path.Base)
		encoder.WriteField("sub", fp.Path.Sub)
		encoder.WriteField("glob", fp.Glob)
	}
	encoder.EndObject()
}
