package denv

import (
	corepkg "github.com/jurgen-kluft/ccode/core"
)

type DevConfig struct {
	BuildType   BuildType // Static, Dynamic, Executable
	BuildConfig BuildConfig
	IncludeDirs []PinnedPath
	Defines     *corepkg.ValueSet
	Libs        []PinnedFilepath // Libraries to link against
}

func NewDevConfig(buildType BuildType, buildConfig BuildConfig) *DevConfig {
	var config = &DevConfig{
		BuildType:   buildType,
		BuildConfig: buildConfig,
		IncludeDirs: []PinnedPath{},
		Defines:     corepkg.NewValueSet(),
		Libs:        []PinnedFilepath{},
	}

	return config
}

func (c *DevConfig) GetIncludeDirs() []string {
	dirs := make([]string, 0, len(c.IncludeDirs))
	for _, dir := range c.IncludeDirs {
		dirs = append(dirs, dir.String())
	}
	return dirs
}

// func (c *DevConfig) EncodeJson(encoder *corepkg.JsonEncoder, key string) {
// 	encoder.BeginObject(key)
// 	{
// 		encoder.WriteField("build_type", c.BuildType.String())
// 		encoder.WriteField("build_config", c.BuildConfig.String())

// 		if len(c.IncludeDirs) > 0 {
// 			encoder.BeginArray("include_dirs")
// 			for _, dir := range c.IncludeDirs {
// 				dir.EncodeJson(encoder, "")
// 			}
// 			encoder.EndArray()
// 		}

// 		if len(c.Defines.Values) > 0 {
// 			encoder.BeginArray("defines")
// 			for _, define := range c.Defines.Values {
// 				encoder.WriteArrayElement(define)
// 			}
// 			encoder.EndArray()
// 		}

// 		if len(c.Libs) > 0 {
// 			encoder.BeginArray("libs")
// 			for _, lib := range c.Libs {
// 				lib.EncodeJson(encoder, "")
// 			}
// 			encoder.EndArray()
// 		}

// 	}
// 	encoder.EndObject()
// }

// func DecodeJsonDevConfig(decoder *corepkg.JsonDecoder) *DevConfig {
// 	var buildType BuildType
// 	var buildConfig BuildConfig
// 	var includeDirs []PinnedPath
// 	var defines = corepkg.NewValueSet()
// 	var libs []PinnedFilepath

// 	fields := map[string]corepkg.JsonDecode{
// 		"build_type":   func(decoder *corepkg.JsonDecoder) { buildType = BuildTypeFromString(decoder.DecodeString()) },
// 		"build_config": func(decoder *corepkg.JsonDecoder) { buildConfig = BuildConfigFromString(decoder.DecodeString()) },
// 		"include_dirs": func(decoder *corepkg.JsonDecoder) {
// 			includeDirs = make([]PinnedPath, 0)
// 			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
// 				includeDirs = append(includeDirs, DecodeJsonPinnedPath(decoder, ""))
// 			})
// 		},
// 		"defines": func(decoder *corepkg.JsonDecoder) {
// 			definesStringArray := decoder.DecodeStringArray()
// 			defines.AddMany(definesStringArray...)
// 		},
// 		"libs": func(decoder *corepkg.JsonDecoder) {
// 			libs = make([]PinnedFilepath, 0)
// 			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
// 				libs = append(libs, DecodeJsonPinnedFilepath(decoder))
// 			})
// 		},
// 	}
// 	decoder.Decode(fields)

// 	return &DevConfig{
// 		BuildType:   buildType,
// 		BuildConfig: buildConfig,
// 		IncludeDirs: includeDirs,
// 		Defines:     defines,
// 		Libs:        libs,
// 	}
// }
