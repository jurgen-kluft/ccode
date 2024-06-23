package denv

import (
	"github.com/jurgen-kluft/ccode/items"
	"github.com/jurgen-kluft/ccode/vars"
)

// Config represents a project build configuration, like 'Debug' or 'Release'
type Config struct {
	Name         string
	Type         string // Static, Dynamic, Executable
	Config       string // Debug, Release, Final
	Build        string // Dev, Test, Retail
	Tundra       string // Tundra specific config string
	Defines      items.List
	IncludeDirs  items.List
	LibraryFiles items.List
	LibraryFile  string
	Vars         vars.Variables
}

var DevDebugStatic = &Config{
	Name:         "Debug",
	Type:         "Static",
	Config:       "Debug",
	Build:        "Dev",
	Tundra:       "*-*-debug",
	Defines:      items.NewList("TARGET_DEBUG;TARGET_DEV;_DEBUG", ";", ""),
	IncludeDirs:  items.NewList(Path("source/main/include"), ";", ""),
	LibraryFiles: items.NewList("", ";", ""),
	Vars:         vars.NewVars(),
}

var DevReleaseStatic = &Config{
	Name:         "Release",
	Type:         "Static",
	Config:       "Release",
	Build:        "Dev",
	Tundra:       "*-*-release",
	Defines:      items.NewList("TARGET_RELEASE;TARGET_DEV;NDEBUG", ";", ""),
	IncludeDirs:  items.NewList(Path("source/main/include"), ";", ""),
	LibraryFiles: items.NewList("", ";", ""),
	Vars:         vars.NewVars(),
}

// ConfigSet type for mapping a config-name to a config-object
type ConfigSet map[string]*Config

// NewConfigSet returns a new ConfigSet
func NewConfigSet() ConfigSet {
	return ConfigSet{}
}

// CopyConfigSet returns a copy of @set
func CopyConfigSet(set ConfigSet) ConfigSet {
	newset := ConfigSet{}
	for name, config := range set {
		newset[name] = CopyConfig(config)
	}
	return newset
}

// Copy returns a copy of @set
func (set ConfigSet) Copy() ConfigSet {
	return CopyConfigSet(set)
}

// HasConfig returns true if the project has that configuration
func (set ConfigSet) HasConfig(configname string) bool {
	for _, config := range set {
		if configname == config.Name {
			return true
		}
	}
	return false
}

// Copy returns a copy of @c
func (c *Config) Copy() *Config {
	return CopyConfig(c)
}

// CopyConfig makes a deep copy of a Config
func CopyConfig(config *Config) *Config {
	newconfig := &Config{Name: config.Name, Config: config.Config}
	newconfig.Config = config.Config
	newconfig.Defines = items.CopyList(config.Defines)
	newconfig.IncludeDirs = items.CopyList(config.IncludeDirs)
	newconfig.LibraryFiles = items.CopyList(config.LibraryFiles)
	newconfig.LibraryFile = config.LibraryFile
	newconfig.Vars = config.Vars.Copy()
	return newconfig
}

// ReplaceVars replaces variables that are present in members of the Config
func (c *Config) ReplaceVars(v vars.Variables, r vars.Replacer) {
	v.ReplaceInLines(r, c.Defines.Items)
	v.ReplaceInLines(r, c.IncludeDirs.Items)
	v.ReplaceInLines(r, c.LibraryFiles.Items)
	c.LibraryFile = v.ReplaceInLine(r, c.LibraryFile)
}
