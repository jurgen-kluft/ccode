package denv

type PinnedGlobPath struct {
	Path PinnedPath
	Glob string
}

func NewPinnedGlobPath(root string, base string, sub string, glob string) PinnedGlobPath {
	return PinnedGlobPath{PinnedPath{Root: root, Base: base, Sub: sub}, glob}
}

// func (fp PinnedGlobPath) EncodeJson(encoder *corepkg.JsonEncoder, key string) {
// 	encoder.BeginObject(key)
// 	{
// 		encoder.WriteField("root", fp.Path.Root)
// 		encoder.WriteField("base", fp.Path.Base)
// 		encoder.WriteField("sub", fp.Path.Sub)
// 		encoder.WriteField("glob", fp.Glob)
// 	}
// 	encoder.EndObject()
// }

// func DecodeJsonPinnedGlobPath(decoder *corepkg.JsonDecoder) PinnedGlobPath {
// 	var fp PinnedGlobPath
// 	fields := map[string]corepkg.JsonDecode{
// 		"root": func(decoder *corepkg.JsonDecoder) { fp.Path.Root = decoder.DecodeString() },
// 		"base": func(decoder *corepkg.JsonDecoder) { fp.Path.Base = decoder.DecodeString() },
// 		"sub":  func(decoder *corepkg.JsonDecoder) { fp.Path.Sub = decoder.DecodeString() },
// 		"glob": func(decoder *corepkg.JsonDecoder) { fp.Glob = decoder.DecodeString() },
// 	}
// 	decoder.Decode(fields)
// 	return fp
// }
