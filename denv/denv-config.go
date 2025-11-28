package denv

import (
	corepkg "github.com/jurgen-kluft/ccode/core"
)

type DevConfig struct {
	BuildType   BuildType // Static, Dynamic, Executable
	BuildConfig BuildConfig
	IncludeDirs []PinnedPath
	Defines     *corepkg.ValueSet
	LibPaths    []PinnedPath
	Libs        []string // Libraries to link against
}

func NewDevConfig(buildType BuildType, buildConfig BuildConfig) *DevConfig {
	var config = &DevConfig{
		BuildType:   buildType,
		BuildConfig: buildConfig,
		IncludeDirs: []PinnedPath{},
		Defines:     corepkg.NewValueSet(),
		LibPaths:    []PinnedPath{},
		Libs:        []string{},
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
