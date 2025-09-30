package toolchain

import (
	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
	"github.com/jurgen-kluft/ccode/denv"
)

// GetBuildDirname returns "os-arch-build-variant-mode"
func GetBuildDirname(config denv.BuildConfig, target denv.BuildTarget) string {
	return target.Os().String() + "-" + target.Arch().String() + "-" + config.String()
}

type Environment interface {
	NewCompiler(config denv.BuildConfig, target denv.BuildTarget) Compiler
	NewArchiver(a ArchiverType, config denv.BuildConfig, target denv.BuildTarget) Archiver
	NewLinker(config denv.BuildConfig, target denv.BuildTarget) Linker
	//NewInformer(config *Config) Informer  // List information about the executable
	NewBurner(config denv.BuildConfig, target denv.BuildTarget) Burner
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
