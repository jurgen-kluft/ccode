package tchost

import "github.com/jurgen-kluft/ccode/foundation"

func ApplyMacOsx(env *foundation.Vars) {
	env.SetMany(map[string][]string{
		"DOTNETRUN":       {"mono "},
		"HOSTPROGSUFFIX":  {""},
		"HOSTSHLIBSUFFIX": {".dylib"},
		"_COPY_FILE":      {"cp", "-f", "$(<)", "$(@)"},
		"_HARDLINK_FILE":  {"ln", "-f", "$(<)", "$(@)"},
	})
}
