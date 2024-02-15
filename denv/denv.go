package denv

import (
	"os"
	"strings"
)

// DEV is an enumeration for all possible IDE's that are supported
type DevEnum int

// All development environment
const (
	TUNDRA       DevEnum = 0x20000
	CMAKE        DevEnum = 0x40000
	VISUALSTUDIO DevEnum = 0x80000
	VS2012       DevEnum = VISUALSTUDIO | 2012
	VS2013       DevEnum = VISUALSTUDIO | 2013
	VS2015       DevEnum = VISUALSTUDIO | 2015
	VS2017       DevEnum = VISUALSTUDIO | 2017
	VS2019       DevEnum = VISUALSTUDIO | 2019
	VS2022       DevEnum = VISUALSTUDIO | 2022
	INVALID      DevEnum = 0xFFFFFFFF
)

func GetDevEnum(dev string) DevEnum {
	if dev == "TUNDRA" {
		return TUNDRA
	} else if dev == "CMAKE" {
		return CMAKE
	} else if dev == "VS2022" {
		return VS2022
	} else if dev == "VS2019" {
		return VS2019
	} else if dev == "VS2017" {
		return VS2017
	} else if dev == "VS2015" {
		return VS2015
	} else if dev == "VS2013" {
		return VS2013
	} else if dev == "VS2012" {
		return VS2012
	}
	return INVALID
}

const (
	OS_WINDOWS = "windows"
	OS_MAC     = "mac"
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
