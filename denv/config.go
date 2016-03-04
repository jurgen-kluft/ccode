package denv

import "github.com/jurgen-kluft/xcode/vars"

// Config represents a project build configuration, like 'Debug' or 'Release'
type Config struct {
	Name         string
	Defines      ItemsList
	IncludeDirs  ItemsList
	LibraryDirs  ItemsList
	LibraryFiles ItemsList
	LibraryFile  string
}

// ConfigSet type for mapping a config-name to a config-object
type ConfigSet map[string]*Config

// HasConfig returns true if the project has that configuration
func (set ConfigSet) HasConfig(configname string) bool {
	for _, config := range set {
		if configname == config.Name {
			return true
		}
	}
	return false
}

// DefaultConfigs $(Configuration)_$(Platform)
var DefaultConfigs = []Config{
	{Name: "DevDebugStatic", Defines: DevDebugDefines, IncludeDirs: "source\\main\\include", LibraryDirs: "target\\$(Configuration)_$(Platform)_$(ToolSet)", LibraryFiles: "", LibraryFile: "${Name}_$(Configuration)_$(Platform)_$(ToolSet).lib"},
	{Name: "DevReleaseStatic", Defines: DevReleaseDefines, IncludeDirs: "source\\main\\include", LibraryDirs: "target\\$(Configuration)_$(Platform)_$(ToolSet)", LibraryFiles: "", LibraryFile: "${Name}_$(Configuration)_$(Platform)_$(ToolSet).lib"},
	{Name: "TestDebugStatic", Defines: TestDebugDefines, IncludeDirs: "source\\main\\include", LibraryDirs: "target\\$(Configuration)_$(Platform)_$(ToolSet)", LibraryFiles: "", LibraryFile: "${Name}_$(Configuration)_$(Platform)_$(ToolSet).lib"},
	{Name: "TestReleaseStatic", Defines: TestReleaseDefines, IncludeDirs: "source\\main\\include", LibraryDirs: "target\\$(Configuration)_$(Platform)_$(ToolSet)", LibraryFiles: "", LibraryFile: "${Name}_$(Configuration)_$(Platform)_$(ToolSet).lib"},
}

// CopyStringArray makes a copy of an array of strings
func CopyStringArray(strarray []string) []string {
	newstrarray := make([]string, len(strarray))
	for i, str := range strarray {
		newstrarray[i] = str
	}
	return newstrarray
}

// CopyConfig makes a deep copy of a Config
func CopyConfig(config Config) *Config {
	newconfig := &Config{Name: config.Name, IncludeDirs: "", LibraryDirs: "", LibraryFiles: "", LibraryFile: ""}
	newconfig.IncludeDirs = config.IncludeDirs
	newconfig.LibraryDirs = config.LibraryDirs
	newconfig.LibraryFiles = config.LibraryFiles
	newconfig.LibraryFile = config.LibraryFile
	return newconfig
}

// ReplaceVars replaces variables that are present in members of the Config
func (c *Config) ReplaceVars(v vars.Variables, r vars.Replacer) {
	c.IncludeDirs = ItemsList(v.ReplaceInLine(r, string(c.IncludeDirs)))
	c.LibraryDirs = ItemsList(v.ReplaceInLine(r, string(c.LibraryDirs)))
	c.LibraryFiles = ItemsList(v.ReplaceInLine(r, string(c.LibraryFiles)))
	c.LibraryFile = v.ReplaceInLine(r, c.LibraryFile)
}

// GetDefaultConfigs returns a map of default configs
func GetDefaultConfigs() map[string]*Config {
	configs := make(map[string]*Config)
	for _, config := range DefaultConfigs {
		configs[config.Name] = CopyConfig(config)
	}
	return configs
}
