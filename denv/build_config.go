package denv

import (
	"strings"
)

type BuildType int

const (
	BuildTypeUnknown        BuildType = 0
	BuildTypeStaticLibrary  BuildType = 1
	BuildTypeDynamicLibrary BuildType = 2
	BuildTypeHeaderOnly     BuildType = 4
	BuildTypeUnittest       BuildType = 16
	BuildTypeCli            BuildType = 32
	BuildTypeApplication    BuildType = 64
)

func (t BuildType) has(o BuildType) bool {
	return (t & o) == o
}

func (t BuildType) IsStaticLibrary() bool {
	return t.has(BuildTypeStaticLibrary)
}

func (t BuildType) IsDynamicLibrary() bool {
	return t.has(BuildTypeDynamicLibrary)
}

func (t BuildType) IsLibrary() bool {
	return t.IsStaticLibrary() || t.IsDynamicLibrary()
}

func (t BuildType) IsApplication() bool {
	return t.has(BuildTypeApplication)
}

func (t BuildType) IsCli() bool {
	return t.has(BuildTypeCli)
}

func (t BuildType) IsUnittest() bool {
	return t.has(BuildTypeUnittest)
}

func (t BuildType) IsExecutable() bool {
	return t.IsApplication() || t.IsCli() || t.IsUnittest()
}

func BuildTypeFromString(str string) BuildType {
	switch strings.ToLower(str) {
	case "application":
		return BuildTypeApplication
	case "cli":
		return BuildTypeCli
	case "unittest":
		return BuildTypeUnittest
	case "dynamic library":
		return BuildTypeDynamicLibrary
	case "static library":
		return BuildTypeStaticLibrary
	case "header only":
		return BuildTypeHeaderOnly
	default:
		return BuildTypeUnknown
	}
}

func (t BuildType) String() string {
	switch t {
	case BuildTypeApplication:
		return "application"
	case BuildTypeCli:
		return "cli"
	case BuildTypeUnittest:
		return "unittest"
	case BuildTypeDynamicLibrary:
		return "dynamic library"
	case BuildTypeStaticLibrary:
		return "static library"
	case BuildTypeHeaderOnly:
		return "header only"
	default:
		return "unknown"
	}
}

func (t BuildType) ProjectString() string {
	switch t {
	case BuildTypeApplication, BuildTypeUnittest, BuildTypeCli:
		return "c_exe"
	case BuildTypeDynamicLibrary:
		return "c_dll"
	case BuildTypeStaticLibrary:
		return "c_lib"
	case BuildTypeHeaderOnly:
		return "c_hol"
	}
	return "error"
}

type BuildBuild uint8

const (
	BuildNone    BuildBuild = 0
	BuildDebug   BuildBuild = 1
	BuildRelease BuildBuild = 2
	BuildAny     BuildBuild = (BuildDebug | BuildRelease)
)

var BuildBuilds = []BuildBuild{BuildDebug, BuildRelease}

type BuildVariant uint8

const (
	BuildVariantNone  BuildVariant = 0
	BuildVariantDev   BuildVariant = 1
	BuildVariantFinal BuildVariant = 2
	BuildVariantAny   BuildVariant = (BuildVariantDev | BuildVariantFinal)
)

var BuildVariants = []BuildVariant{BuildVariantDev, BuildVariantFinal}

type BuildMode uint8

const (
	BuildModeInvalid BuildMode = 0
	BuildModeNone    BuildMode = 1
	BuildModeTest    BuildMode = 2
	BuildModeProfile BuildMode = 4
	BuildModeAny     BuildMode = (BuildModeNone | BuildModeTest | BuildModeProfile)
)

var BuildModes = []BuildMode{BuildModeNone, BuildModeTest, BuildModeProfile}

type BuildConfig struct {
	Build   BuildBuild
	Variant BuildVariant
	Mode    BuildMode
}

func NewDebugDevConfig() BuildConfig {
	return BuildConfig{
		Build:   BuildDebug,
		Variant: BuildVariantDev,
		Mode:    BuildModeNone,
	}
}

func NewReleaseDevConfig() BuildConfig {
	return BuildConfig{
		Build:   BuildRelease,
		Variant: BuildVariantDev,
		Mode:    BuildModeNone,
	}
}

func NewDebugDevTestConfig() BuildConfig {
	return BuildConfig{
		Build:   BuildDebug,
		Variant: BuildVariantDev,
		Mode:    BuildModeTest,
	}
}

func NewReleaseDevTestConfig() BuildConfig {
	return BuildConfig{
		Build:   BuildRelease,
		Variant: BuildVariantDev,
		Mode:    BuildModeTest,
	}
}

func (t BuildConfig) IsEqual(o BuildConfig) bool {
	return t == o
}

func (t BuildConfig) IsDebug() bool {
	return t.Build == BuildDebug
}

func (t BuildConfig) IsRelease() bool {
	return t.Build == BuildRelease
}

func (t BuildConfig) IsDevelopment() bool {
	return t.Variant == BuildVariantDev
}

func (t BuildConfig) IsFinal() bool {
	return t.Variant == BuildVariantFinal
}

func (t BuildConfig) IsTest() bool {
	return t.Mode == BuildModeTest
}

func (t BuildConfig) IsProfile() bool {
	return t.Mode == BuildModeProfile
}

func (t BuildConfig) hasBuild(b BuildBuild) bool {
	return (t.Build & b) == b
}

func (t BuildConfig) hasVariant(v BuildVariant) bool {
	return (t.Variant & v) == v
}

func (t BuildConfig) hasMode(m BuildMode) bool {
	return (t.Mode & m) == m
}

func (t BuildConfig) Contains(o BuildConfig) bool {
	return t.hasBuild(o.Build) && t.hasVariant(o.Variant) && t.hasMode(o.Mode)
}

func (t BuildConfig) BuildAsString() string {
	switch t.Build {
	case BuildNone:
		return "none"
	case BuildDebug:
		return "debug"
	case BuildRelease:
		return "release"
	}
	return "*"
}

func (t BuildConfig) VariantAsString() string {
	switch t.Variant {
	case BuildVariantNone:
		return "none"
	case BuildVariantDev:
		return "dev"
	case BuildVariantFinal:
		return "final"
	}
	return "*"
}

func (t BuildConfig) ModeAsString() string {
	switch t.Mode {
	case BuildModeNone:
		return "none"
	case BuildModeTest:
		return "test"
	case BuildModeProfile:
		return "profile"
	}
	return "*"
}

func (t BuildConfig) String() string {
	return t.BuildAsString() + "-" + t.VariantAsString() + "-" + t.ModeAsString()
}

func BuildConfigIterate(configStr string, cb func(config BuildConfig)) {
	config := BuildConfigFromString(configStr)
	for _, b := range BuildBuilds {
		if config.hasBuild(b) {
			for _, v := range BuildVariants {
				if config.hasVariant(v) {
					for _, m := range BuildModes {
						if config.hasMode(m) {
							cb(BuildConfig{Build: b, Variant: v, Mode: m})
						}
					}
				}
			}
		}
	}
}

// BuildConfig from config and variant
func BuildConfigFromString(configStr string) BuildConfig {
	cfg := BuildConfig{Build: BuildNone, Variant: BuildVariantDev, Mode: BuildModeNone}

	part := 0
	cursor := 0
	for cursor < len(configStr) {
		index := strings.Index(configStr[cursor:], "-")
		if index == -1 {
			index = len(configStr)
		} else {
			index = cursor + index
		}
		config := strings.ToLower(configStr[cursor:index])
		cursor = index + 1

		switch part {
		case 0:
			switch config {
			case "none":
				cfg.Build = BuildNone
			case "debug":
				cfg.Build = BuildDebug
			case "release":
				cfg.Build = BuildRelease
			case "*":
				cfg.Build = BuildAny
			}
		case 1:
			switch config {
			case "none":
				cfg.Variant = BuildVariantNone
			case "dev":
				cfg.Variant = BuildVariantDev
			case "final":
				cfg.Variant = BuildVariantFinal
			case "*":
				cfg.Variant = BuildVariantAny
			}
		case 2:
			switch config {
			case "none":
				cfg.Mode = BuildModeNone
			case "test":
				cfg.Mode = BuildModeTest
			case "profile":
				cfg.Mode = BuildModeProfile
			case "*":
				cfg.Mode = BuildModeAny
			}
		}

		part++
	}
	return cfg
}

// -----------------------------------------------------------------------------------------------------
// BuildConfig to Tundra string
// -----------------------------------------------------------------------------------------------------

func (t BuildConfig) Tundra() string {
	config := "*-*-"

	if t.IsRelease() {
		config += "release"
	} else {
		config += "debug"
	}

	// if t.IsDevelopment() {
	// 	config += "dev"
	// } else if t.IsFinal() {
	// 	config += "final"
	// } else {
	// 	config += "dev"
	// }

	if t.IsTest() {
		config += "-test"
	} else if t.IsProfile() {
		config += "-profile"
	} else {
		config += "-*"
	}

	return config
}
