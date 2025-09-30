package denv

import (
	"path/filepath"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

type PinnedFilepath struct {
	Path     PinnedPath
	Filename string // Filename (without extension)
}

func (fp PinnedFilepath) String() string {
	return filepath.Join(fp.Path.Root, fp.Path.Base, fp.Path.Sub, fp.Filename)
}

func (fp PinnedFilepath) RelativeTo(root string) string {
	return corepkg.PathGetRelativeTo(fp.String(), root)
}

func (fp PinnedFilepath) EncodeJson(encoder *corepkg.JsonEncoder, key string) {
	encoder.BeginObject(key)
	{
		encoder.WriteField("root", fp.Path.Root)
		encoder.WriteField("base", fp.Path.Base)
		encoder.WriteField("sub", fp.Path.Sub)
		encoder.WriteField("filename", fp.Filename)
	}
	encoder.EndObject()
}
