package toolchain

import "github.com/jurgen-kluft/ccode/dev"

type Config struct {
	Config dev.BuildConfig
	Target dev.BuildTarget
}

func NewConfig(config dev.BuildConfig, target dev.BuildTarget) *Config {
	return &Config{
		Config: config,
		Target: target,
	}
}

// GetDirname returns "os-arch-build-variant"
func (c *Config) GetDirname() string {
	return c.Target.OSAsString() + "-" + c.Target.ArchAsString() + "-" + c.Config.Build() + "-" + c.Config.Variant()
}

type Toolchain interface {
	NewCCompiler(config *Config) Compiler
	NewCppCompiler(config *Config) Compiler
	NewArchiver(config *Config) Archiver
	NewLinker(config *Config) Linker
	//NewInformer(config *Config) Informer  // List information about the executable
	NewBurner(config *Config) Burner
}

type ToolchainInstance struct {
	Name  string
	Vars  Vars
	Tools map[string]string
}

func (t *ToolchainInstance) ResolveVars() {

	for key, values := range t.Vars {
		for i, value := range values {
			t.Vars[key][i] = ResolveString(value, t.Vars)
		}
	}

	for toolKey, toolPath := range t.Tools {
		t.Tools[toolKey] = ResolveString(toolPath, t.Vars)
	}
}

// Example:
//
//	   BuildTargetConfig = "clang-arm64-debug-test"
//		  Config = "*-*-debug-*"
//	   Result = true
//
// Example:
//
//	BuildTargetConfig = "clang-arm64-debug-test"
//	Config = "*-*-*-test"
//	Result = true
//
// Example:
//
//	BuildTargetConfig = "clang-arm64-debug-test"
//	Config = "*-*-debug-test"
//	Result = true
func ConfigMatches(lhsConfig string, rhsConfig string) bool {
	if lhsConfig == "*-*-*-*" || rhsConfig == "*-*-*-*" {
		return true
	}

	// Do manual pattern matching
	li := 0
	ri := 0
	for li < len(lhsConfig) && ri < len(rhsConfig) {
		lhs := ""
		for i := li; i < len(lhsConfig); i++ {
			if lhsConfig[i] != '-' {
				continue

			}
			lhs = lhsConfig[li:i]
			li = i + 1
			break
		}
		if lhs == "" {
			lhs = lhsConfig[li:]
			li = len(lhsConfig)
		}

		rhs := ""
		for i := ri; i < len(rhsConfig); i++ {
			if rhsConfig[i] != '-' {
				continue
			}
			rhs = rhsConfig[ri:i]
			ri = i + 1
			break
		}
		if rhs == "" {
			rhs = rhsConfig[ri:]
			ri = len(rhsConfig)
		}

		if lhs == "*" || rhs == "*" {
			continue
		}

		if lhs != rhs {
			return false
		}
	}

	return true
}
