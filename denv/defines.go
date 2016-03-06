package denv

import (
	"github.com/jurgen-kluft/xcode/items"
)

var DevDebugDefines = items.NewList("TARGET_DEV_DEBUG;_DEBUG", ";")
var DevReleaseDefines = items.NewList("TARGET_DEV_RELEASE;NDEBUG", ";")
var TestDebugDefines = items.NewList("TARGET_TEST_DEBUG;_DEBUG", ";")
var TestReleaseDefines = items.NewList("TARGET_TEST_RELEASE;NDEBUG", ";")

// DefaultDefineMap defines for every default config a ItemsList
var DefaultDefineMap = map[string]items.List{
	"DevDebugStatic":    DevDebugDefines,
	"DevReleaseStatic":  DevReleaseDefines,
	"TestDebugStatic":   TestDebugDefines,
	"TestReleaseStatic": TestReleaseDefines,
}
