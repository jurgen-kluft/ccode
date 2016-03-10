package denv

import (
	"github.com/jurgen-kluft/xcode/items"
)

var DevDebugDefines = items.NewList("TARGET_DEV_DEBUG;TARGET_PC;_DEBUG", ";")
var DevReleaseDefines = items.NewList("TARGET_DEV_RELEASE;TARGET_PC;NDEBUG", ";")
var TestDebugDefines = items.NewList("TARGET_TEST_DEBUG;TARGET_PC;_DEBUG", ";")
var TestReleaseDefines = items.NewList("TARGET_TEST_RELEASE;TARGET_PC;NDEBUG", ";")

// DefaultDefineMap defines for every default config a ItemsList
var DefaultDefineMap = map[string]items.List{
	"DevDebugStatic":    DevDebugDefines,
	"DevReleaseStatic":  DevReleaseDefines,
	"TestDebugStatic":   TestDebugDefines,
	"TestReleaseStatic": TestReleaseDefines,
}
