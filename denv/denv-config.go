package denv

import (
	corepkg "github.com/jurgen-kluft/ccode/core"
	"github.com/jurgen-kluft/ccode/dev"
)

type DevConfig struct {
	BuildType   dev.BuildType // Static, Dynamic, Executable
	BuildConfig dev.BuildConfig
	IncludeDirs []dev.PinnedPath
	Defines     *corepkg.ValueSet
	LinkFlags   *corepkg.ValueSet
	Libs        []dev.PinnedFilepath // Libraries to link against
}

func NewDevConfig(buildType dev.BuildType, buildConfig dev.BuildConfig) *DevConfig {
	var config = &DevConfig{
		// Type:    "Static", // Static, Dynamic, Executable
		BuildType: buildType,
		// Config:  "Debug",  // Debug, Release, Final
		// Build:   "Dev",    // Development(dev), Unittest(test), Profile(prof), Production(prod)
		BuildConfig: buildConfig,
		IncludeDirs: []dev.PinnedPath{},
		Defines:     corepkg.NewValueSet(),
		LinkFlags:   corepkg.NewValueSet(),
		Libs:        []dev.PinnedFilepath{},
	}

	return config
}
