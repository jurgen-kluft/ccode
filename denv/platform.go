package denv

import (
	"github.com/jurgen-kluft/ccode/items"
	"github.com/jurgen-kluft/ccode/vars"
)

const (
	// PlatformWin32 OS is Windows, 32-bit
	PlatformWin32 = "Win32"
	// PlatformWin64 OS is Windows, 64-bit
	PlatformWin64 = "x64"
	// PlatformDarwin64 is OSX, 64-bit
	PlatformDarwin64 = "Darwin64"
	// PlatformLinux64 is Linux, 64-bit
	PlatformLinux64 = "Linux64"
)

// Platform represents a platform and holds configurations for that platform
type Platform struct {
	Name                 string
	OS                   string
	FilePatternsToIgnore []string
	Defines              items.List
	Configs              ConfigSet
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
func (p *Platform) GetConfig(configname string) *Config {
	for _, config := range p.Configs {
		if config.Name == configname {
			return config
		}
	}
	return nil
}

func (p *Platform) ReplaceVars(v vars.Variables, r vars.Replacer) {
	for _, config := range p.Configs {
		config.ReplaceVars(v, r)
	}
}

func (p *Platform) AddIncludeDir(includeDir string) {
	var platform = p
	for _, config := range platform.Configs {
		config.IncludeDirs = config.IncludeDirs.Add(includeDir)
	}
}

func (p *Platform) AddDefine(define string) {
	var platform = p
	for _, config := range platform.Configs {
		config.Defines = config.Defines.Add(define)
	}
}

func (p *Platform) AddVar(varname, varvalue string) {
	var platform = p
	for _, config := range platform.Configs {
		config.Vars.AddVar(varname, varvalue)
	}
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
var defaultWinPlatform = Platform{
	Name:                 PlatformWin64,
	OS:                   "windows",
	FilePatternsToIgnore: []string{"_darwin", "_linux", "_nob"},
	Defines:              items.NewList("TARGET_PC", ";", ""),
	Configs: ConfigSet{
		"DevDebugStatic":   DevDebugStatic.Copy(),
		"DevReleaseStatic": DevReleaseStatic.Copy(),
	},
}

var defaultDarwinPlatform = Platform{
	Name:                 PlatformDarwin64,
	OS:                   "darwin",
	FilePatternsToIgnore: []string{"_win32", "_linux", "_nob"},
	Defines:              items.NewList("TARGET_MAC", ";", ""),
	Configs: ConfigSet{
		"DevDebugStatic":   DevDebugStatic.Copy(),
		"DevReleaseStatic": DevReleaseStatic.Copy(),
	},
}

var defaultLinuxPlatform = Platform{
	Name:                 PlatformLinux64,
	OS:                   "linux",
	FilePatternsToIgnore: []string{"_win32", "_darwin", "_nob"},
	Defines:              items.NewList("TARGET_LINUX", ";", ""),
	Configs: ConfigSet{
		"DevDebugStatic":   DevDebugStatic.Copy(),
		"DevReleaseStatic": DevReleaseStatic.Copy(),
	},
}

// Copy returns a copy of @pset (PlatformSet)
func (pset PlatformSet) Copy() PlatformSet {
	set := PlatformSet{}
	for pn, p := range pset {
		platform := &Platform{Name: pn}
		platform.Defines = p.Defines.Copy()
		platform.Configs = p.Configs.Copy()
		set[platform.Name] = platform
	}
	return set
}

// GetDefaultPlatforms returns the default platform according to the OS we are running on at the moment
func GetDefaultPlatform() *Platform {

	//dev := GetDevEnum(DEV)

	platform := &Platform{}
	if OS == "windows" {
		var p = defaultWinPlatform
		platform.Name = p.Name
		platform.OS = p.OS
		platform.Defines = p.Defines.Copy()
		platform.Configs = p.Configs.Copy()
		platform.FilePatternsToIgnore = p.FilePatternsToIgnore
	} else if OS == "linux" {
		var p = defaultLinuxPlatform
		platform.Name = p.Name
		platform.OS = p.OS
		platform.Defines = p.Defines.Copy()
		platform.Configs = p.Configs.Copy()
		platform.FilePatternsToIgnore = p.FilePatternsToIgnore
	} else {
		var p = defaultDarwinPlatform
		platform.Name = p.Name
		platform.OS = p.OS
		platform.Defines = p.Defines.Copy()
		platform.Configs = p.Configs.Copy()
		platform.FilePatternsToIgnore = p.FilePatternsToIgnore
	}

	// Merge the platform defines into the configurations
	for _, config := range platform.Configs {
		config.Defines = config.Defines.Merge(platform.Defines)
	}

	return platform
}

func (pset PlatformSet) AddIncludeDir(includeDir string) {
	for _, platform := range pset {
		for _, config := range platform.Configs {
			config.IncludeDirs = config.IncludeDirs.Add(includeDir)
		}
	}
}

func (pset PlatformSet) AddDefine(define string) {
	for _, platform := range pset {
		for _, config := range platform.Configs {
			config.Defines = config.Defines.Add(define)
		}
	}
}

func (pset PlatformSet) AddVar(varname, varvalue string) {
	for _, platform := range pset {
		for _, config := range platform.Configs {
			config.Vars.AddVar(varname, varvalue)
		}
	}
}
