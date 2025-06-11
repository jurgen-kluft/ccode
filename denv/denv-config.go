package denv

import (
	"github.com/jurgen-kluft/ccode/dev"
	"github.com/jurgen-kluft/ccode/foundation"
)

type DevConfig struct {
	BuildType   dev.BuildType // Static, Dynamic, Executable
	BuildConfig dev.BuildConfig
	IncludeDirs []dev.PinPath
	Defines     *foundation.ValueSet
	LinkFlags   *foundation.ValueSet
	Libs        []dev.PinFilepath // Libraries to link against
}

func NewDevConfig(buildType dev.BuildType, buildConfig dev.BuildConfig) *DevConfig {
	var config = &DevConfig{
		// Type:    "Static", // Static, Dynamic, Executable
		BuildType: buildType,
		// Config:  "Debug",  // Debug, Release, Final
		// Build:   "Dev",    // Development(dev), Unittest(test), Profile(prof), Production(prod)
		BuildConfig: buildConfig,
		IncludeDirs: []dev.PinPath{},
		Defines:     foundation.NewValueSet(),
		LinkFlags:   foundation.NewValueSet(),
		Libs:        []dev.PinFilepath{},
	}

	return config
}
