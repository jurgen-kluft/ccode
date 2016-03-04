package denv

var DevDebugDefines ItemsList = "TARGET_DEV_DEBUG;_DEBUG"
var DevReleaseDefines ItemsList = "TARGET_DEV_RELEASE;NDEBUG"
var TestDebugDefines ItemsList = "TARGET_TEST_DEBUG;_DEBUG"
var TestReleaseDefines ItemsList = "TARGET_TEST_RELEASE;NDEBUG"

// DefaultDefineMap defines for every default config a ItemsList
var DefaultDefineMap = map[string]ItemsList{
	"DevDebugStatic":    DevDebugDefines,
	"DevReleaseStatic":  DevReleaseDefines,
	"TestDebugStatic":   TestDebugDefines,
	"TestReleaseStatic": TestReleaseDefines,
}
