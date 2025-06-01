package clay

import (
	"github.com/jurgen-kluft/ccode/dev"
)

type Config struct {
	Config dev.BuildConfig
	Target dev.BuildTarget
}

func NewConfig(os, cpu string, config string) *Config {
	buildConfig := dev.BuildConfigFromString(config)
	buildTarget := dev.BuildTargetFromString(os, cpu)
	return &Config{
		Config: buildConfig,
		Target: buildTarget,
	}
}

// GetSubDir returns a subdirectory name based on the OS, CPU, Build type, and Variant.
// Example: "linux-x86-release-dev" or "arduino-esp32-debug-prod".
func (c *Config) GetSubDir() string {
	return c.Target.OSAsString() + "-" + c.Target.ArchAsString() + "-" + c.Config.AsString()
}

func (c *Config) Matches(other *Config) bool {
	if c == nil || other == nil {
		return false
	}
	return c.Target.IsEqual(other.Target) && c.Config == other.Config
}

func (c *Config) ConfigString() string {
	return c.Config.AsString()
}

func (c *Config) String() string {
	return c.Target.OSAsString() + "-" + c.Target.ArchAsString() + "-" + c.Config.AsString()
}
