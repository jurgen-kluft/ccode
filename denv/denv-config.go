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
