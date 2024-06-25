package axe

import "strings"

type DevEnum uint

// All development environment
const (
	TUNDRA       DevEnum = 0x020000
	CMAKE        DevEnum = 0x040000
	MAKE         DevEnum = 0x080000
	XCODE        DevEnum = 0x100000
	VISUALSTUDIO DevEnum = 0x200000
	VS2015       DevEnum = VISUALSTUDIO | 2015
	VS2017       DevEnum = VISUALSTUDIO | 2017
	VS2019       DevEnum = VISUALSTUDIO | 2019
	VS2022       DevEnum = VISUALSTUDIO | 2022
	INVALID      DevEnum = 0xFFFFFFFF
)

func (d DevEnum) String() string {
	switch d {
	case TUNDRA:
		return "tundra"
	case CMAKE:
		return "cmake"
	case MAKE:
		return "make"
	case XCODE:
		return "xcode"
	case VS2015:
		return "vs2015"
	case VS2017:
		return "vs2017"
	case VS2019:
		return "vs2019"
	case VS2022:
		return "vs2022"
	default:
		return "__invalid__"
	}
}
func GetDevEnum(dev string) DevEnum {
	dev = strings.ToLower(dev)
	if dev == "tundra" {
		return TUNDRA
	} else if dev == "make" {
		return MAKE
	} else if dev == "cmake" {
		return CMAKE
	} else if dev == "xcode" {
		return XCODE
	} else if dev == "vs2022" {
		return VS2022
	} else if dev == "vs2019" {
		return VS2019
	} else if dev == "vs2017" {
		return VS2017
	} else if dev == "vs2015" {
		return VS2015
	}
	return INVALID
}

func (d DevEnum) IsXCode() bool {
	return d == XCODE
}
