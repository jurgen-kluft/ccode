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

func (c *DevConfig) EncodeJson(encoder *corepkg.JsonEncoder, key string) {
	encoder.BeginObject(key)
	{
		encoder.WriteField("build_type", c.BuildType.String())
		encoder.WriteField("build_config", c.BuildConfig.String())
		encoder.WriteField("include_dirs", c.IncludeDirs)
		{
			if len(c.IncludeDirs) > 0 {
				encoder.BeginArray("")
				for _, dir := range c.IncludeDirs {
					dir.EncodeJson(encoder, "")
				}
				encoder.EndArray()
			}
		}
		encoder.WriteField("defines", c.Defines)
		{
			c.Defines.EncodeJson(encoder, "")
		}
		encoder.WriteField("libs", c.Libs)
		{
			if len(c.Libs) > 0 {
				encoder.BeginArray("")
				for _, lib := range c.Libs {
					lib.EncodeJson(encoder, "")
				}
				encoder.EndArray()
			}
		}
	}
	encoder.EndObject()
}

func DecodeJsonDevConfig(decoder *corepkg.JsonDecoder) *DevConfig {
	var buildType BuildType
	var buildConfig BuildConfig
	var includeDirs []PinnedPath
	var defines = corepkg.NewValueSet()
	var libs []PinnedFilepath

	fields := map[string]corepkg.JsonDecode{
		"build_type":   func(decoder *corepkg.JsonDecoder) { buildType = BuildTypeFromString(decoder.DecodeString()) },
		"build_config": func(decoder *corepkg.JsonDecoder) { buildConfig = BuildConfigFromString(decoder.DecodeString()) },
		"include_dirs": func(decoder *corepkg.JsonDecoder) {
			includeDirs = make([]PinnedPath, 0)
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				includeDirs = append(includeDirs, DecodeJsonPinnedPath(decoder, ""))
			})
		},
		"defines": func(decoder *corepkg.JsonDecoder) { defines = corepkg.DecodeJsonValueSet(decoder) },
		"libs": func(decoder *corepkg.JsonDecoder) {
			libs = make([]PinnedFilepath, 0)
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				libs = append(libs, DecodeJsonPinnedFilepath(decoder))
			})
		},
	}
	decoder.Decode(fields)

	return &DevConfig{
		BuildType:   buildType,
		BuildConfig: buildConfig,
		IncludeDirs: includeDirs,
		Defines:     defines,
		Libs:        libs,
	}
}
