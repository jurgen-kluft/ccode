package denv

import (
	"github.com/jurgen-kluft/xcode/items"
	"github.com/jurgen-kluft/xcode/vars"
)

const (
	// PlatformWin32 OS is Windows, 32-bit
	PlatformWin32 = "Win32"
	// PlatformWin64 OS is Windows, 64-bit
	PlatformWin64 = "x64"
	// PlatformDarwin64 is OSX, 64-bit
	PlatformDarwin64 = "Darwin64"
)

// Platform represents a platform and holds configurations for that platform
type Platform struct {
	Name    string
	Defines items.List
	Configs ConfigSet
}

// HasConfig will return true if the platform contains a config with @name
func (p *Platform) HasConfig(configname string) bool {
	for _, config := range p.Configs {
		if config.Name == configname {
			return true
		}
	}
	return false
}

// GetConfig will return the Config with name @configname
func (p *Platform) GetConfig(configname string) (*Config, bool) {
	for _, config := range p.Configs {
		if config.Name == configname {
			return config, true
		}
	}
	return nil, false
}

// PlatformSet type for mapping a config-name to a config-object
type PlatformSet map[string]*Platform

// HasPlatform returns true if the set has that platform
func (pset PlatformSet) HasPlatform(platformname string) bool {
	for _, entry := range pset {
		if platformname == entry.Name {
			return true
		}
	}
	return false
}

// ReplaceVars replaces any variable that exists in members of Project
func (pset PlatformSet) ReplaceVars(v vars.Variables, r vars.Replacer) {
	for _, platform := range pset {
		for _, config := range platform.Configs {
			config.ReplaceVars(v, r)
		}
	}
}

// DefaultPlatforms defines a set of supported platforms
var defaultPlatforms = PlatformSet{
	PlatformWin32: &Platform{
		Name:    PlatformWin32,
		Defines: items.NewList("PLATFORM_PC;TARGET_PC;PLATFORM_32BIT", ";"),
		Configs: ConfigSet{
			DevDebugStatic:   defaultPlatformConfig(DevDebugStatic),
			DevReleaseStatic: defaultPlatformConfig(DevReleaseStatic),
		},
	},
	PlatformWin64: &Platform{
		Name:    PlatformWin64,
		Defines: items.NewList("PLATFORM_PC;TARGET_PC;PLATFORM_64BIT", ";"),
		Configs: ConfigSet{
			DevDebugStatic:   defaultPlatformConfig(DevDebugStatic),
			DevReleaseStatic: defaultPlatformConfig(DevReleaseStatic),
		},
	},
}

// Copy returns a copy of @pset (PlatformSet)
func (pset PlatformSet) Copy() PlatformSet {
	set := PlatformSet{}
	for pn, p := range defaultPlatforms {
		platform := &Platform{Name: pn}
		platform.Defines = p.Defines.Copy()
		platform.Configs = p.Configs.Copy()
		set[platform.Name] = platform
	}
	return set
}

// GetDefaultPlatforms returns the default set of platforms
func GetDefaultPlatforms() PlatformSet {
	pset := defaultPlatforms.Copy()

	// Merge the platform defines into the configurations
	for _, platform := range pset {
		for _, config := range platform.Configs {
			config.Defines = config.Defines.Merge(platform.Defines)
		}
	}
	return pset
}
