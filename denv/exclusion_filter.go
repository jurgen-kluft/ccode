package denv

import (
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

// ----------------------------------------------------------------------------------------------
// Exclusion filter
// ----------------------------------------------------------------------------------------------
var gExcludeDefault = map[string]bool{
	"_nob": false, "_null": false, "_nill": false,
	"_win": true, "_pc": true, "_win32": true, "_win64": true, "_windows": true, "_d3d11": true, "_d3d12": true,
	"_mac": true, "_macos": true, "_darwin": true, "_cocoa": true, "_metal": true, "_osx": true,
	"_ios": true, "_iphone": true, "_ipad": true, "_ipod": true,
	"_linux": true, "_unix": true,
	"_arduino": true, "_esp32": true, "_esp8266": true,
}

var gExcludeWindows = map[string]bool{
	"_nob": true, "_null": true, "_nill": true,
	"_win": false, "_pc": false, "_win32": false, "_win64": false, "_windows": false, "_d3d11": false, "_d3d12": false,
	"_mac": true, "_macos": true, "_darwin": true, "_cocoa": true, "_metal": true, "_osx": true,
	"_ios": true, "_iphone": true, "_ipad": true, "_ipod": true,
	"_linux": true, "_unix": true,
	"_arduino": true, "_esp32": true, "_esp8266": true,
}
var gExcludeMac = map[string]bool{
	"_nob": true, "_null": true, "_nill": true,
	"_win": true, "_pc": true, "_win32": true, "_win64": true, "_windows": true, "_d3d11": true, "_d3d12": true,
	"_mac": false, "_macos": false, "_darwin": false, "_cocoa": false, "_metal": false, "_osx": false,
	"_ios": true, "_iphone": true, "_ipad": true, "_ipod": true,
	"_linux": true, "_unix": true,
	"_arduino": true, "_esp32": true, "_esp8266": true,
}
var gExcludeIOs = map[string]bool{
	"_nob": true, "_null": true, "_nill": true,
	"_win": true, "_pc": true, "_win32": true, "_win64": true, "_windows": true, "_d3d11": true, "_d3d12": true,
	"_mac": true, "_macos": true, "_darwin": true, "_cocoa": true, "_metal": true, "_osx": true,
	"_ios": false, "_iphone": false, "_ipad": false, "_ipod": false,
	"_linux": true, "_unix": true,
	"_arduino": true, "_esp32": true, "_esp8266": true,
}
var gExcludeLinux = map[string]bool{
	"_nob": true, "_null": true, "_nill": true,
	"_win": true, "_pc": true, "_win32": true, "_win64": true, "_windows": true, "_d3d11": true, "_d3d12": true,
	"_mac": true, "_macos": true, "_darwin": true, "_cocoa": true, "_metal": true, "_osx": true,
	"_ios": true, "_iphone": true, "_ipad": true, "_ipod": true,
	"_linux": false, "_unix": false,
	"_arduino": true, "_esp32": true, "_esp8266": true,
}
var gExcludeArduinoEsp32 = map[string]bool{
	"_nob": true, "_null": true, "_nill": true,
	"_win": true, "_pc": true, "_win32": true, "_win64": true, "_windows": true, "_d3d11": true, "_d3d12": true,
	"_mac": true, "_macos": true, "_darwin": true, "_cocoa": true, "_metal": true, "_osx": true,
	"_ios": true, "_iphone": true, "_ipad": true, "_ipod": true,
	"_linux": true, "_unix": true,
	"_arduino": false, "_esp32": false, "_esp8266": true,
}
var gExcludeArduinoEsp8266 = map[string]bool{
	"_nob": true, "_null": true, "_nill": true,
	"_win": true, "_pc": true, "_win32": true, "_win64": true, "_windows": true, "_d3d11": true, "_d3d12": true,
	"_mac": true, "_macos": true, "_darwin": true, "_cocoa": true, "_metal": true, "_osx": true,
	"_ios": true, "_iphone": true, "_ipad": true, "_ipod": true,
	"_linux": true, "_unix": true,
	"_arduino": false, "_esp32": true, "_esp8266": false,
}

func IsExcludedOn(str string, suffixes map[string]bool) bool {
	// find '_' in str from the end
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] == '_' {
			suffix := str[i:]
			if val, ok := suffixes[suffix]; ok {
				return val
			}
		}
	}
	return false
}

func IsExcludedOnMac(str string) bool {
	return IsExcludedOn(str, gExcludeMac)
}
func IsExcludedOnWindows(str string) bool {
	return IsExcludedOn(str, gExcludeWindows)
}
func IsExcludedOnLinux(str string) bool {
	return IsExcludedOn(str, gExcludeLinux)
}
func IsExcludedOnArduinoEsp32(str string) bool {
	return IsExcludedOn(str, gExcludeArduinoEsp32)
}
func IsExcludedOnArduinoEsp8266(str string) bool {
	return IsExcludedOn(str, gExcludeArduinoEsp8266)
}
func IsExcludedDefault(str string) bool {
	return IsExcludedOn(str, gExcludeDefault)
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
