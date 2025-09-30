package denv

import (
	"strings"
)

type BuildType int

const (
	BuildTypeUnknown        BuildType = 0
	BuildTypeStaticLibrary  BuildType = 1
	BuildTypeDynamicLibrary BuildType = 2
	BuildTypeUnittest       BuildType = 3
	BuildTypeCli            BuildType = 4
	BuildTypeApplication    BuildType = 5
)

func (t BuildType) IsStaticLibrary() bool {
	return t&BuildTypeStaticLibrary != 0
}

func (t BuildType) IsDynamicLibrary() bool {
	return t&BuildTypeDynamicLibrary != 0
}

func (t BuildType) IsLibrary() bool {
	return t&(BuildTypeStaticLibrary|BuildTypeDynamicLibrary) != 0
}

func (t BuildType) IsApplication() bool {
	return t&BuildTypeApplication != 0
}

func (t BuildType) IsCli() bool {
	return t&BuildTypeCli != 0
}

func (t BuildType) IsUnittest() bool {
	return t&BuildTypeUnittest != 0
}

func (t BuildType) IsExecutable() bool {
	return t == BuildTypeApplication || t == BuildTypeCli || t == BuildTypeUnittest
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
	}
	return "error"
}

type BuildBuild uint8

const (
	BuildDebug   BuildBuild = 1
	BuildRelease BuildBuild = 2
)

type BuildVariant uint8

const (
	BuildVariantDev   BuildVariant = 1
	BuildVariantFinal BuildVariant = 2
)

type BuildMode uint8

const (
	BuildModeNone    BuildMode = 0
	BuildModeTest    BuildMode = 1
	BuildModeProfile BuildMode = 2
)

type BuildConfig struct {
	Build   BuildBuild
	Variant BuildVariant
	Mode    BuildMode
}

type BuildConfigList struct {
	List []BuildConfig
}

func NewBuildConfigList() *BuildConfigList {
	return &BuildConfigList{
		List: []BuildConfig{},
	}
}

func NewBuildAllConfigList() *BuildConfigList {
	return &BuildConfigList{
		List: []BuildConfig{
			NewDebugDevConfig(),
			NewReleaseDevConfig(),
			NewDebugDevTestConfig(),
			NewReleaseDevTestConfig(),
		},
	}
}

func (b *BuildConfigList) Add(config BuildConfig) {
	if !b.Contains(config) {
		b.List = append(b.List, config)
	}
}

func (b *BuildConfigList) Contains(config BuildConfig) bool {
	for _, existing := range b.List {
		if existing.IsEqual(config) {
			return true // Found a match
		}
	}
	return false // No match found
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

func (t BuildConfig) BuildAsString() string {
	if t.IsDebug() {
		return "debug"
	} else if t.IsRelease() {
		return "release"
	}
	return "unknown"
}

func (t BuildConfig) VariantAsString() string {
	if t.IsDevelopment() {
		return "dev"
	} else if t.IsFinal() {
		return "final"
	}
	return "unknown"
}

func (t BuildConfig) AsString() string {
	str := t.BuildAsString() + "-" + t.VariantAsString()
	if t.Mode == BuildModeTest {
		return str + "-test"
	} else if t.Mode == BuildModeProfile {
		return str + "-profile"
	}
	return str
}

// BuildConfig from config and variant
func BuildConfigFromString(configStr string) BuildConfig {

	// Default configuration
	cfg := BuildConfig{Build: BuildDebug, Variant: BuildVariantDev, Mode: BuildModeNone}

	i := 0
	begin := 0
	for begin < len(configStr) {
		end := strings.Index(configStr[begin:], "-")
		if end == -1 {
			end = len(configStr)
		} else {
			end = begin + end
		}
		config := configStr[begin:end]
		begin = end + 1
		i++

		if i == 1 {
			switch strings.ToLower(config) {
			case "debug":
				cfg.Build = BuildDebug
				continue
			case "release":
				cfg.Build = BuildRelease
				continue
			}
		} else if i == 2 {
			switch strings.ToLower(config) {
			case "dev":
				cfg.Variant = BuildVariantDev
				continue
			case "final":
				cfg.Variant = BuildVariantFinal
				continue
			}
		} else if i == 3 {
			switch strings.ToLower(config) {
			case "test":
				cfg.Mode = BuildModeTest
				continue
			case "profile":
				cfg.Mode = BuildModeProfile
				continue
			}
		}
	}
	return cfg
}

// -----------------------------------------------------------------------------------------------------
// BuildConfig to Tundra string
// -----------------------------------------------------------------------------------------------------

func (t BuildConfig) Tundra() string {
	config := "*-*-"

	if t.IsDebug() {
		config += "debug"
	} else if t.IsRelease() {
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
