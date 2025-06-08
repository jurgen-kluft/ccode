package tchost

import utils "github.com/jurgen-kluft/ccode/utils"

func ApplyWindows(env *utils.Vars) {
	env.SetMany(map[string][]string{
		"DOTNETRUN":       {""},
		"HOSTPROGSUFFIX":  {".exe"},
		"HOSTSHLIBSUFFIX": {".dll"},
		"_COPY_FILE":      {"copy $(<) $(@)"},
		"_HARDLINK_FILE":  {"copy /f $(<) $(@)"},
	})
}
