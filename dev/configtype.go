package dev

import (
    "strings"
)

type BuildType int

const (
	BuildTypeStaticLibrary  BuildType = 1
	BuildTypeDynamicLibrary BuildType = 2
	BuildTypeExecutable     BuildType = 4
	BuildTypeOutputMask               = BuildTypeStaticLibrary | BuildTypeDynamicLibrary | BuildTypeExecutable
)

func (t BuildType) GetProjectType() BuildType {
	return t & (BuildTypeStaticLibrary | BuildTypeDynamicLibrary | BuildTypeExecutable)
}

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
	return t&BuildTypeExecutable != 0
}
func (t BuildType) IsExecutable() bool {
	return t&BuildTypeExecutable != 0
}

func (t BuildType) ProjectString() string {
	switch t.GetProjectType() {
	case BuildTypeExecutable:
		return "c_exe"
	case BuildTypeDynamicLibrary:
		return "c_dll"
	case BuildTypeStaticLibrary:
		return "c_lib"
	}
	return "error"
}

type BuildConfig int

const (
	BuildConfigDebug       BuildConfig = 8
	BuildConfigRelease     BuildConfig = 16
	BuildConfigFinal       BuildConfig = 64
	BuildConfigConfigAll               = BuildConfigDebug | BuildConfigRelease | BuildConfigFinal
	BuildConfigConfigMask              = BuildConfigDebug | BuildConfigRelease | BuildConfigFinal
	BuildConfigDevelopment BuildConfig = 128
	BuildConfigTest        BuildConfig = 256
	BuildConfigProfile     BuildConfig = 512
	BuildConfigProduction  BuildConfig = 1024
	BuildConfigVariantMask             = BuildConfigDevelopment | BuildConfigTest | BuildConfigProfile | BuildConfigProduction
	BuildConfigAll         BuildConfig = BuildConfigConfigMask | BuildConfigVariantMask
)

func (t BuildConfig) IsEqual(o BuildConfig) bool {
	return t == o
}

func (t BuildConfig) Contains(o BuildConfig) bool {
	return t&o == o
}

func (t BuildConfig) GetBuildConfig() BuildConfig {
	return t & (BuildConfigDebug | BuildConfigRelease | BuildConfigFinal)
}

func (t BuildConfig) GetBuildConfigVariant() BuildConfig {
	return t & (BuildConfigDevelopment | BuildConfigTest | BuildConfigProfile | BuildConfigProduction)
}

func (t BuildConfig) IsDebug() bool {
	return t&BuildConfigDebug != 0
}

func (t BuildConfig) IsRelease() bool {
	return t&BuildConfigRelease != 0
}

func (t BuildConfig) IsFinal() bool {
	return t&BuildConfigFinal != 0
}

func (t BuildConfig) IsDevelopment() bool {
	return t&BuildConfigDevelopment != 0
}

func (t BuildConfig) IsTest() bool {
	return t&BuildConfigTest != 0
}

func (t BuildConfig) IsProfile() bool {
	return t&BuildConfigProfile != 0
}

func (t BuildConfig) IsProduction() bool {
	return t&BuildConfigProduction != 0
}

func (t BuildConfig) Build() string {
	str := "Debug"

	if t.IsDebug() {
		str = "Debug"
	} else if t.IsRelease() {
		str = "Release"
	} else if t.IsFinal() {
		str = "Final"
	}
	return str
}

func (t BuildConfig) Variant() string {
	str := "Dev"

	if t.IsTest() {
		str += "Test"
	} else if t.IsProfile() {
		str += "Profile"
	} else if t.IsProduction() {
		str += "Prod"
	}
	return str
}

func (t BuildConfig) ConfigString() string {
	str := "Debug"
	if t.IsRelease() {
		str = "Release"
	} else if t.IsFinal() {
		str = "Final"
	}

	if t.IsTest() {
		str += "Test"
	} else if t.IsProfile() {
		str += "Profile"
	} else if t.IsProduction() {
		str += "Prod"
	} else if t.IsDevelopment() {
		str += "Dev"
	}

	return str
}

// BuildConfig from config and variant
func BuildConfigFromString(config string, variant string) BuildConfig {
	var cfg BuildConfig

	switch strings.ToLower(config) {
	case "debug":
		cfg |= BuildConfigDebug
	case "release":
		cfg |= BuildConfigRelease
	case "final":
		cfg |= BuildConfigFinal
	default:
		cfg |= BuildConfigDebug // Default to debug if not specified
	}

	switch strings.ToLower(variant) {
	case "dev":
		cfg |= BuildConfigDevelopment
	case "test":
		cfg |= BuildConfigTest
	case "profile":
		cfg |= BuildConfigProfile
	case "prod":
		cfg |= BuildConfigProduction
	default:
		cfg |= BuildConfigDevelopment // Default to development if not specified
	}

	return cfg
}

// -----------------------------------------------------------------------------------------------------
// BuildConfig to Tundra string
// -----------------------------------------------------------------------------------------------------

func (t BuildConfig) Tundra() string {
	config := "*-*-"
	switch t & BuildConfigConfigMask {
	case BuildConfigDebug:
		config += "debug"
	case BuildConfigRelease:
		config += "release"
	case BuildConfigFinal:
		config += "final"
	}

	switch t & BuildConfigVariantMask {
	case BuildConfigDevelopment:
		config += "-dev"
	case BuildConfigTest:
		config += "-test"
	case BuildConfigProfile:
		config += "-profile"
	case BuildConfigProduction:
		config += "-prod"
	default:
		config += "-*"
	}

	return config
}
