package denv

import (
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

// ----------------------------------------------------------------------------------------------
// Exclusion filter
// ----------------------------------------------------------------------------------------------
var gValidSuffixDefault = []string{"_nob", "_null", "_nill"}
var gValidSuffixWindows = []string{"_win", "_pc", "_win32", "_win64", "_windows", "_d3d11", "_d3d12"}
var gValidSuffixMac = []string{"_mac", "_macos", "_darwin", "_cocoa", "_metal", "_osx"}
var gValidSuffixIOs = []string{"_ios", "_iphone", "_ipad", "_ipod"}
var gValidSuffixLinux = []string{"_linux", "_unix"}
var gValidSuffixArduinoEsp32 = []string{"_arduino", "_esp32"}
var gValidSuffixArduinoEsp8266 = []string{"_arduino", "_esp32"}

func IsExcludedOn(str string, suffixes []string) bool {
	for _, e := range suffixes {
		if strings.HasSuffix(str, e) {
			return true
		}
	}
	return false
}

func IsExcludedOnMac(str string) bool {
	return IsExcludedOn(str, gValidSuffixMac)
}
func IsExcludedOnWindows(str string) bool {
	return IsExcludedOn(str, gValidSuffixWindows)
}
func IsExcludedOnLinux(str string) bool {
	return IsExcludedOn(str, gValidSuffixLinux)
}
func IsExcludedOnArduinoEsp32(str string) bool {
	return IsExcludedOn(str, gValidSuffixArduinoEsp32)
}
func IsExcludedOnArduinoEsp8266(str string) bool {
	return IsExcludedOn(str, gValidSuffixArduinoEsp8266)
}
func IsExcludedDefault(str string) bool {
	return IsExcludedOn(str, gValidSuffixDefault)
}

func NewExclusionFilter(target BuildTarget) *ExclusionFilter {
	if target.Mac() {
		return &ExclusionFilter{Filter: IsExcludedOnMac}
	} else if target.Windows() {
		return &ExclusionFilter{Filter: IsExcludedOnWindows}
	} else if target.Linux() {
		return &ExclusionFilter{Filter: IsExcludedOnLinux}
	} else if target.Arduino() && target.Esp32() {
		return &ExclusionFilter{Filter: IsExcludedOnArduinoEsp32}
	} else if target.Arduino() && target.Esp8266() {
		return &ExclusionFilter{Filter: IsExcludedOnArduinoEsp8266}
	}
	return &ExclusionFilter{Filter: IsExcludedDefault}
}

type ExclusionFilter struct {
	Filter func(filepath string) bool
}

func (f *ExclusionFilter) IsExcluded(filepath string) bool {
	parts := corepkg.PathSplitRelativeFilePath(filepath, true)
	for i := 0; i < len(parts)-1; i++ {
		p := strings.ToLower(parts[i])
		if f.Filter(p) {
			return true
		}
	}
	return false
}
