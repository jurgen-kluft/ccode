package toolchain

import (
	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
	"github.com/jurgen-kluft/ccode/denv"
)

type Config struct {
	Config denv.BuildConfig
	Target denv.BuildTarget
}

func NewConfig(config denv.BuildConfig, target denv.BuildTarget) *Config {
	return &Config{
		Config: config,
		Target: target,
	}
}

func (t *Config) IsDebug() bool       { return t.Config.IsDebug() }
func (t *Config) IsRelease() bool     { return t.Config.IsRelease() }
func (t *Config) IsDevelopment() bool { return t.Config.IsDevelopment() }
func (t *Config) IsFinal() bool       { return t.Config.IsFinal() }
func (t *Config) IsTest() bool        { return t.Config.IsTest() }
func (t *Config) IsProfile() bool     { return t.Config.IsProfile() }

// GetDirname returns "os-arch-build-variant-mode"
func (c *Config) GetDirname() string {
	return c.Target.Os().String() + "-" + c.Target.Arch().String() + "-" + c.Config.String()
}

type Environment interface {
	NewCompiler(config *Config) Compiler
	NewArchiver(a ArchiverType, config *Config) Archiver
	NewLinker(config *Config) Linker
	//NewInformer(config *Config) Informer  // List information about the executable
	NewBurner(config *Config) Burner
	NewDependencyTracker(dirpath string) deptrackr.FileTrackr
}

// --------------------------------------------------
// Example:
//   - BuildTargetConfig = "clang-arm64-debug-test"
//   - Config = "*-*-debug-*"
//   - Result = true
//
// Example:
//   - BuildTargetConfig = "clang-arm64-debug-test"
//   - Config = "*-*-*-test"
//   - Result = true
//
// Example:
//   - BuildTargetConfig = "clang-arm64-debug-test"
//   - Config = "*-*-debug-test"
//   - Result = true
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
