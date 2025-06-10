package tchost

import "github.com/jurgen-kluft/ccode/foundation"

func ApplyWindows(env *foundation.Vars) {
	env.SetMany(map[string][]string{
		"DOTNETRUN":       {""},
		"HOSTPROGSUFFIX":  {".exe"},
		"HOSTSHLIBSUFFIX": {".dll"},
		"_COPY_FILE":      {"copy $(<) $(@)"},
		"_HARDLINK_FILE":  {"copy /f $(<) $(@)"},
	})
}
