package denv

import (
	"os"
	"strings"
)

// IDE is an enumeration for all possible IDE's that are supported
type IDE int

const (
	VISUALSTUDIO IDE = 0x80000000
	VS2012       IDE = VISUALSTUDIO | 2012
	VS2013       IDE = VISUALSTUDIO | 2013
	VS2015       IDE = VISUALSTUDIO | 2015
	VS2017       IDE = VISUALSTUDIO | 2017
	CODELITE     IDE = 0x70000000
)

// Fixpath will fix forward/backward slashes to match the current OS
func Path(path string) string {
	path = strings.Replace(path, "\\", string(os.PathSeparator), -1)
	return path
}
