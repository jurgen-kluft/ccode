package denv

import (
	"github.com/jurgen-kluft/xcode/items"
)

var defaultDefines = map[string]string{
	DevDebugStatic:   "TARGET_DEV_DEBUG;_DEBUG",
	DevReleaseStatic: "TARGET_DEV_RELEASE;NDEBUG",
}

func getDefines(config string) items.List {
	defines, ok := defaultDefines[config]
	if !ok {
		items.NewList("", ";", "")
	}
	return items.NewList(defines, ";", "")
}
