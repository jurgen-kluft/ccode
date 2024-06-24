package denv

// Config represents a project build configuration, like 'Debug' or 'Release'
type Config struct {
	Name         string
	Type         string // Static, Dynamic, Executable
	Config       string // Debug, Release, Final
	Build        string // Development, Unittest, Profile, Shipping
	Tundra       string // Tundra specific config string
	Defines      []string
	IncludeDirs  []string
	LibraryFiles []string
}

var DebugConfig = &Config{
	Name:         "Debug",
	Type:         "Static",
	Config:       "Debug",
	Build:        "Dev",
	Tundra:       "*-*-debug",
	Defines:      []string{},
	IncludeDirs:  []string{},
	LibraryFiles: []string{},
}

var ReleaseConfig = &Config{
	Name:         "Release",
	Type:         "Static",
	Config:       "Release",
	Build:        "Dev",
	Tundra:       "*-*-release",
	Defines:      []string{},
	IncludeDirs:  []string{},
	LibraryFiles: []string{},
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
	newconfig.Defines = append(newconfig.Defines, config.Defines...)
	newconfig.IncludeDirs = append(newconfig.IncludeDirs, config.IncludeDirs...)
	newconfig.LibraryFiles = append(newconfig.LibraryFiles, config.LibraryFiles...)
	return newconfig
}
