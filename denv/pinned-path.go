package denv

import (
	"path/filepath"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

type PinnedPath struct {
	Root string // Root directory
	Base string // Base directory
	Sub  string // Sub directory
}

func (fp PinnedPath) String() string {
	return filepath.Join(fp.Root, fp.Base, fp.Sub)
}

func (fp PinnedPath) RelativeTo(root string) string {
	return corepkg.PathGetRelativeTo(fp.String(), root)
}

func (fp PinnedPath) EncodeJson(encoder *corepkg.JsonEncoder, key string) {
	encoder.BeginObject(key)
	{
		encoder.WriteField("root", fp.Root)
		encoder.WriteField("base", fp.Base)
		encoder.WriteField("sub", fp.Sub)
	}
	encoder.EndObject()
}
