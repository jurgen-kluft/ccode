package denv

import dev "github.com/jurgen-kluft/ccode/dev"

type DevConfig struct {
	BuildType           dev.BuildType // Static, Dynamic, Executable
	BuildConfig         dev.BuildConfig
	LocalIncludeDirs    []string // Relative paths
	ExternalIncludeDirs []string // Absolute paths
	SourceDirs          []string
	Defines             *DevValueSet
	LinkFlags           *DevValueSet
	Libs                []*DevLib
}

func NewDevConfig(buildType dev.BuildType, buildConfig dev.BuildConfig) *DevConfig {
	var config = &DevConfig{
		// Type:    "Static", // Static, Dynamic, Executable
		BuildType: buildType,
		// Config:  "Debug",  // Debug, Release, Final
		// Build:   "Dev",    // Development(dev), Unittest(test), Profile(prof), Production(prod)
		BuildConfig:         buildConfig,
		LocalIncludeDirs:    []string{},
		ExternalIncludeDirs: []string{},
		SourceDirs:          []string{},
		Defines:             NewDevValueSet(),
		LinkFlags:           NewDevValueSet(),
		Libs:                []*DevLib{},
	}

	return config
}
