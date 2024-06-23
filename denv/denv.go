package denv

import (
	"os"
	"strings"
)

// DEV is an enumeration for all possible IDE's that are supported
type DevEnum uint

// All development environment
const (
	TUNDRA       DevEnum = 0x010000
	MAKE         DevEnum = 0x020000
	CMAKE        DevEnum = 0x040000
	XCODE        DevEnum = 0x080000
	VISUALSTUDIO DevEnum = 0x100000
	VS2015       DevEnum = VISUALSTUDIO | 2015
	VS2017       DevEnum = VISUALSTUDIO | 2017
	VS2019       DevEnum = VISUALSTUDIO | 2019
	VS2022       DevEnum = VISUALSTUDIO | 2022
	INVALID      DevEnum = 0xFFFFFFFF
)

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

const (
	OS_WINDOWS = "windows"
	OS_MAC     = "darwin"
	OS_LINUX   = "linux"
)

// XCodeDEV constant: Visual Studio, Tundra
var DEV string

// XCodeOS constant: Windows, Darwin, Linux
var OS string

// XCodeARCH constant: x64 ?
var ARCH string

// Path will fix forward/backward slashes to match the current OS
func Path(path string) string {
	to := string(os.PathSeparator)
	if strings.EqualFold(DEV, "tundra") {
		to = "/"
	}

	path = strings.Replace(path, "\\\\", "\\", -1)

	from := "\\"
	if to == "\\" {
		from = "/"
	}
	path = strings.Replace(path, from, to, -1)

	return path
}

// PathFixer is a delegate used by items.List
func PathFixer(item string, prefix string) string {
	return Path(item)
}
