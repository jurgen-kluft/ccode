package tchost

import utils "github.com/jurgen-kluft/ccode/utils"

func ApplyMacOsx(env *utils.Vars) {
	env.SetMany(map[string][]string{
		"DOTNETRUN":       {"mono "},
		"HOSTPROGSUFFIX":  {""},
		"HOSTSHLIBSUFFIX": {".dylib"},
		"_COPY_FILE":      {"cp", "-f", "$(<)", "$(@)"},
		"_HARDLINK_FILE":  {"ln", "-f", "$(<)", "$(@)"},
	})
}
