package denv

import (
	"os"
	"strings"
)

// IDE is an enumeration for all possible IDE's that are supported
type DEV int

const (
	VISUALSTUDIO DEV = 0x8000
	VS2012       DEV = VISUALSTUDIO | 2012
	VS2013       DEV = VISUALSTUDIO | 2013
	VS2015       DEV = VISUALSTUDIO | 2015
	VS2017       DEV = VISUALSTUDIO | 2017
	CODELITE     DEV = 0x10000
	TUNDRA       DEV = 0x20000
)

var XCodeDEV string
var XCodeOS string
var XCodeARCH string

func Init(DEV string, OS string, ARCH string) {
	XCodeDEV = DEV
	XCodeOS = OS
	XCodeARCH = ARCH
}

// Path will fix forward/backward slashes to match the current OS
func Path(path string) string {
	to := string(os.PathSeparator)
	if strings.EqualFold(XCodeDEV, "tundra") {
		to = "/"
	}

	from := "\\"
	if to == "\\" {
		from = "/"
	} else {
		from = "\\"
	}
	path = strings.Replace(path, from, to, -1)
	return path
}

// PathFixer is a delegate used by items.List
func PathFixer(item string, prefix string) string {
	return Path(item)
}
