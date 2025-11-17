package denv

import (
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

// ----------------------------------------------------------------------------------------------
// Exclusion filter
// ----------------------------------------------------------------------------------------------
type PlatformFilter int

const (
	PlatformFilterNone PlatformFilter = iota
	PlatformFilterWindows
	PlatformFilterMac
	PlatformFilterIOs
	PlatformFilterLinux
	PlatformFilterArduinoEsp32
	PlatformFilterArduinoEsp8266
)

var gIncludeMap = map[PlatformFilter][]string{
	PlatformFilterNone:           {"nob", "null", "nill"},
	PlatformFilterWindows:        {"win", "pc", "win32", "win64", "windows", "d3d11", "d3d12"},
	PlatformFilterMac:            {"mac", "macos", "darwin", "cocoa", "metal", "osx"},
	PlatformFilterIOs:            {"ios", "iphone", "ipad", "ipod"},
	PlatformFilterLinux:          {"linux", "unix"},
	PlatformFilterArduinoEsp32:   {"arduino", "esp32"},
	PlatformFilterArduinoEsp8266: {"arduino", "esp8266"},
}
var gExcludeFilter = map[string]PlatformFilter{}

type ExclusionFilter struct {
	filter   map[string]PlatformFilter
	platform PlatformFilter
}

func (f *ExclusionFilter) Filter(str string) bool {
	// find '_' in str from the end
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] == '_' && i+1 < len(str) {
			suffix := str[i+1:]
			if p, ok := f.filter[suffix]; ok {
				return p != f.platform
			}
		}
	}
	return false
}

func NewExclusionFilter(target BuildTarget) *ExclusionFilter {
	if len(gExcludeFilter) == 0 {
		for pf, includes := range gIncludeMap {
			for _, inc := range includes {
				gExcludeFilter[inc] = pf
			}
		}
	}

	if target.Mac() {
		return &ExclusionFilter{filter: gExcludeFilter, platform: PlatformFilterMac}
	} else if target.Windows() {
		return &ExclusionFilter{filter: gExcludeFilter, platform: PlatformFilterWindows}
	} else if target.Linux() {
		return &ExclusionFilter{filter: gExcludeFilter, platform: PlatformFilterLinux}
	} else if target.Arduino() && target.Esp32() {
		return &ExclusionFilter{filter: gExcludeFilter, platform: PlatformFilterArduinoEsp32}
	} else if target.Arduino() && target.Esp8266() {
		return &ExclusionFilter{filter: gExcludeFilter, platform: PlatformFilterArduinoEsp8266}
	}
	return &ExclusionFilter{filter: gExcludeFilter, platform: PlatformFilterNone}
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
