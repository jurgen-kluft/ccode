package denv

import dev "github.com/jurgen-kluft/ccode/dev"

type DevConfig struct {
	BuildType   dev.BuildType // Static, Dynamic, Executable
	BuildConfig dev.BuildConfig
	IncludeDirs []dev.PinPath
	Defines     *DevValueSet
	LinkFlags   *DevValueSet
	Libs        []*DevLib
}

func NewDevConfig(buildType dev.BuildType, buildConfig dev.BuildConfig) *DevConfig {
	var config = &DevConfig{
		// Type:    "Static", // Static, Dynamic, Executable
		BuildType: buildType,
		// Config:  "Debug",  // Debug, Release, Final
		// Build:   "Dev",    // Development(dev), Unittest(test), Profile(prof), Production(prod)
		BuildConfig: buildConfig,
		IncludeDirs: []dev.PinPath{},
		Defines:     NewDevValueSet(),
		LinkFlags:   NewDevValueSet(),
		Libs:        []*DevLib{},
	}

	return config
}
