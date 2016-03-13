package denv

import (
	"github.com/jurgen-kluft/xcode/items"
	"github.com/jurgen-kluft/xcode/vars"
)

// Config represents a project build configuration, like 'Debug' or 'Release'
type Config struct {
	Name         string
	Defines      items.List
	IncludeDirs  items.List
	LibraryDirs  items.List
	LibraryFiles items.List
	LibraryFile  string
}

func defaultPlatformConfig(name string) *Config {
	defines := getDefines(name)
	return &Config{Name: name,
		Defines:      defines,
		IncludeDirs:  items.NewList(Path("source\\main\\include"), ";"),
		LibraryDirs:  items.NewList(Path("target\\${Name}\\bin\\$(PackageSignature)"), ";"),
		LibraryFiles: items.NewList("", ";"),
		LibraryFile:  "${Name}_$(PackageSignature).lib",
	}
}

const (
	DevDebugStatic   = "DevDebugStatic"
	DevReleaseStatic = "DevReleaseStatic"
)

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
	newconfig := &Config{Name: config.Name, Defines: config.Defines, IncludeDirs: items.NewList("", ";"), LibraryDirs: items.NewList("", ";"), LibraryFiles: items.NewList("", ";"), LibraryFile: ""}
	newconfig.Defines = items.CopyList(config.Defines)
	newconfig.IncludeDirs = items.CopyList(config.IncludeDirs)
	newconfig.LibraryDirs = items.CopyList(config.LibraryDirs)
	newconfig.LibraryFiles = items.CopyList(config.LibraryFiles)
	newconfig.LibraryFile = config.LibraryFile
	return newconfig
}

// ReplaceVars replaces variables that are present in members of the Config
func (c *Config) ReplaceVars(v vars.Variables, r vars.Replacer) {
	v.ReplaceInLines(r, c.Defines.Items)
	v.ReplaceInLines(r, c.IncludeDirs.Items)
	v.ReplaceInLines(r, c.LibraryDirs.Items)
	v.ReplaceInLines(r, c.LibraryFiles.Items)
	c.LibraryFile = v.ReplaceInLine(r, c.LibraryFile)
}
