package denv

import (
	"runtime"
	"strings"
)

type DevEnum uint64

// All development environment
const (
	DevTundra        DevEnum = 0x0000000000010000
	DevMake          DevEnum = 0x0000000000020000
	DevXcode         DevEnum = 0x0000000000040000
	DevVisualStudio  DevEnum = 0x0000000000080000
	DevVs2015        DevEnum = DevVisualStudio | 2015
	DevVs2017        DevEnum = DevVisualStudio | 2017
	DevVs2019        DevEnum = DevVisualStudio | 2019
	DevVs2022        DevEnum = DevVisualStudio | 2022
	DevEsp32         DevEnum = 0x0000000000100000
	DevEsp32s3       DevEnum = 0x0000000000200000
	DevCompilerMsvc  DevEnum = 0x0000000010000000
	DevCompilerGcc   DevEnum = 0x0000000020000000
	DevCompilerClang DevEnum = 0x0000000040000000
	DevInvalid       DevEnum = 0xFFFFFFFFFFFFFFFF
)

func (d DevEnum) CompilerIsMsvc() bool {
	return d&DevCompilerMsvc != 0
}
func (d DevEnum) CompilerIsClang() bool {
	return d&DevCompilerClang != 0
}
func (d DevEnum) CompilerIsGcc() bool {
	return d&DevCompilerGcc != 0
}

// CompilerAsString
func (d DevEnum) CompilerAsString() string {
	if d.CompilerIsMsvc() {
		return "msvc"
	} else if d.CompilerIsClang() {
		return "clang"
	} else if d.CompilerIsGcc() {
		return "gcc"
	}
	return ""
}

func (d DevEnum) IsValid() bool {
	return d.IsVisualStudio() || d.IsTundra() || d.IsMake() || d.IsXCode()
}
func (d DevEnum) IsVisualStudio() bool {
	return d&DevVisualStudio != 0
}
func (d DevEnum) IsTundra() bool {
	return d == DevTundra
}
func (d DevEnum) IsMake() bool {
	return d == DevMake
}
func (d DevEnum) IsXCode() bool {
	return d == DevXcode
}
func (d DevEnum) IsEsp32() bool {
	return d == DevEsp32
}

var DevEnumToStrMap = map[DevEnum]string{
	DevTundra:       "tundra",
	DevMake:         "make",
	DevXcode:        "xcode",
	DevVisualStudio: "vs2022",
	DevVs2015:       "vs2015",
	DevVs2017:       "vs2017",
	DevVs2019:       "vs2019",
	DevVs2022:       "vs2022",
	DevEsp32:        "esp32",
	DevEsp32s3:      "esp32s3",
}

var DevStrToEnumMap = map[string]DevEnum{
	"tundra":       DevTundra,
	"make":         DevMake,
	"xcode":        DevXcode,
	"vs2015":       DevVs2015,
	"vs2017":       DevVs2017,
	"vs2019":       DevVs2019,
	"vs2022":       DevVs2022,
	"vs":           DevVs2022,
	"esp32":        DevEsp32,
	"esp32s3":      DevEsp32s3,
	"visualstudio": DevVisualStudio,
}

func (d DevEnum) ToString() string {
	if str, ok := DevEnumToStrMap[d]; ok {
		return str
	}
	return "__invalid__"
}

func NewDevEnum(dev string) DevEnum {
	dev = strings.ToLower(dev)
	if devEnum, ok := DevStrToEnumMap[dev]; ok {
		return devEnum
	}

	if runtime.GOOS == "windows" {
		return DevVs2022
	} else if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		return DevTundra
	}

	return DevInvalid
}
