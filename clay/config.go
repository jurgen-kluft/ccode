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

func (c *Config) GetCppDefines() []string {
	defines := []string{}

	if c.Target.Arduino() {
		defines = append(defines, "TARGET_ESP32")
	} else if c.Target.Linux() {
		defines = append(defines, "TARGET_LINUX")
	} else if c.Target.Windows() {
		defines = append(defines, "TARGET_PC")
	} else if c.Target.Mac() {
		defines = append(defines, "TARGET_MAC")
	}

	if c.Target.Arm32() {
		defines = append(defines, "TARGET_ARCH_ARM32")
	} else if c.Target.Arm64() {
		defines = append(defines, "TARGET_ARCH_ARM64")
	} else if c.Target.X86() {
		defines = append(defines, "TARGET_ARCH_X86")
	} else if c.Target.X64() {
		defines = append(defines, "TARGET_ARCH_X64")
	} else if c.Target.Esp32() {
		defines = append(defines, "TARGET_ARCH_ESP32")
	} else if c.Target.Esp32s3() {
		defines = append(defines, "TARGET_ARCH_ESP32S3")
	}

	if c.Config.IsDebug() {
		defines = append(defines, "TARGET_DEBUG")
	} else if c.Config.IsRelease() {
		defines = append(defines, "TARGET_RELEASE")
	}

	if c.Config.IsDevelopment() {
		defines = append(defines, "TARGET_DEV")
	} else if c.Config.IsFinal() {
		defines = append(defines, "TARGET_FINAL")
	}

	if c.Config.IsTest() {
		defines = append(defines, "TARGET_TEST")
	} else if c.Config.IsProfile() {
		defines = append(defines, "TARGET_PROFILE")
	}

	return defines
}
