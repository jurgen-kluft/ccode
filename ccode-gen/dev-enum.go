package ccode_gen

import "strings"

type DevEnum uint

// All development environment
const (
	DevTundra       DevEnum = 0x020000
	DevMake         DevEnum = 0x080000
	DevXcode        DevEnum = 0x100000
	DevVisualStudio DevEnum = 0x200000
	DevVs2015       DevEnum = DevVisualStudio | 2015
	DevVs2017       DevEnum = DevVisualStudio | 2017
	DevVs2019       DevEnum = DevVisualStudio | 2019
	DevVs2022       DevEnum = DevVisualStudio | 2022
	DevEspMake      DevEnum = 0x400000
	DevInvalid      DevEnum = 0xFFFFFFFF
)

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
func (d DevEnum) IsEspMake() bool {
	return d == DevEspMake
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
	DevEspMake:      "espmake",
}

var DevStrToEnumMap = map[string]DevEnum{
	"tundra":       DevTundra,
	"make":         DevMake,
	"xcode":        DevXcode,
	"vs2022":       DevVs2022,
	"vs2015":       DevVs2015,
	"vs2017":       DevVs2017,
	"vs2019":       DevVs2019,
	"espmake":      DevEspMake,
	"visualstudio": DevVisualStudio,
}

func (d DevEnum) String() string {
	if str, ok := DevEnumToStrMap[d]; ok {
		return str
	}
	return "__invalid__"
}

func DevEnumFromString(dev string) DevEnum {
	dev = strings.ToLower(dev)
	if devEnum, ok := DevStrToEnumMap[dev]; ok {
		return devEnum
	}

	if strings.HasPrefix(dev, "vs") {
		return DevVs2022
	}

	return DevInvalid
}
