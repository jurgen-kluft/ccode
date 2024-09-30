package axe

import "strings"

type DevEnum uint

// All development environment
const (
	DevTundra  DevEnum = 0x020000
	DevMake    DevEnum = 0x080000
	DevXcode   DevEnum = 0x100000
	DevVs2015  DevEnum = 0x200000 | 2015
	DevVs2017  DevEnum = 0x200000 | 2017
	DevVs2019  DevEnum = 0x200000 | 2019
	DevVs2022  DevEnum = 0x200000 | 2022
	DevInvalid DevEnum = 0xFFFFFFFF
)

func (d DevEnum) IsValid() bool {
	return d.IsVisualStudio() || d.IsTundra() || d.IsMake() || d.IsXCode()
}
func (d DevEnum) IsVisualStudio() bool {
	return d == DevVs2015 || d == DevVs2017 || d == DevVs2019 || d == DevVs2022
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

func (d DevEnum) String() string {
	switch d {
	case DevTundra:
		return "tundra"
	case DevMake:
		return "make"
	case DevXcode:
		return "xcode"
	case DevVs2015:
		return "vs2015"
	case DevVs2017:
		return "vs2017"
	case DevVs2019:
		return "vs2019"
	case DevVs2022:
		return "vs2022"
	default:
		return "__invalid__"
	}
}
func DevEnumFromString(dev string) DevEnum {
	dev = strings.ToLower(dev)
	if dev == "tundra" {
		return DevTundra
	} else if dev == "make" {
		return DevMake
	} else if dev == "xcode" {
		return DevXcode
	} else if dev == "vs2022" {
		return DevVs2022
	} else if dev == "vs2019" {
		return DevVs2019
	} else if dev == "vs2017" {
		return DevVs2017
	} else if dev == "vs2015" {
		return DevVs2015
	} else {
		if strings.HasPrefix(dev, "vs") {
			return DevVs2022
		}
	}

	return DevInvalid
}
