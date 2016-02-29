package xcode

import (
	"fmt"
)

type Config struct {
	name      string // Debug, Release
	defines   string //
	includes  []string
	libraries []string
	linking   []string
}

var DefinesPerOS = map[string]string{
	"windows": "TARGET_PC;WIN32",
	"darwin":  "TARGET_OSX",
}

var DefinesPerARCH = map[string]string{
	"x86": "TARGET_32BIT",
	"x64": "TARGET_64BIT",
}

var DefaultConfigs = []Config{
	{name: "DevDebug", defines: "TARGET_DEV_DEBUG;_DEBUG;", libraries: []string{""}, linking: []string{""}},
	{name: "DevRelease", defines: "TARGET_DEV_RELEASE;NDEBUG;"},
	{name: "DevFinal", defines: "TARGET_DEV_FINAL;NDEBUG;"},
	{name: "TestDebug", defines: "TARGET_TEST_DEBUG;_DEBUG;"},
	{name: "TestRelease", defines: "TARGET_TEST_RELEASE;NDEBUG;"},
}

// Version (based on semver)
type Version struct {
	Major uint32
	Minor uint32
	Patch uint32
}

type Package struct {
	name     string
	guid     string
	author   string
	version  Version
	os       string // Windows, Darwin
	arch     string // x86, x64, ARM
	language string
	configs  []Config
}

type Dependency struct {
	packageName string
	version     Version
}

type IDE int

const (
	VISUALSTUDIO IDE = 0x80000000
	VS2012       IDE = VISUALSTUDIO | 2012
	VS2013       IDE = VISUALSTUDIO | 2013
	VS2015       IDE = VISUALSTUDIO | 2015
)

type ProjectGenerator interface {
	Generate(path string, pkg Package)
}

func NewGenerator(version IDE) (ProjectGenerator, error) {
	if (version & VISUALSTUDIO) == VISUALSTUDIO {
		return NewGeneratorForVisualStudio(version)
	}
	return nil, fmt.Errorf("Wrong visual studio version")
}

func StaticLibrary() {

}

func Unittest() {

}
