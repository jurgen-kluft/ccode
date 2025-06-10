package tctools

import "github.com/jurgen-kluft/ccode/foundation"

func ApplyGccOsx(env *foundation.Vars, options *foundation.Vars) {

	ApplyGcc(env, options)

	env.SetMany(map[string][]string{
		"NATIVE_SUFFIXES": {".c", ".cpp", ".cc", ".cxx", ".m", ".mm", ".a", ".o"},
		"CXXEXTS":         {"cpp", "cxx", "cc", "mm"},
		"FRAMEWORKS":      {""},
		"FRAMEWORKPATH":   {},
		"SHLIBPREFIX":     {"lib"},
		"SHLIBOPTS":       {"-shared"},
		"_OS_CCOPTS":      {"-F$(FRAMEWORKPATH)"},
		"_OS_CXXOPTS":     {"-F$(FRAMEWORKPATH)"},
		"SHLIBCOM":        {"$(LD)", "$(SHLIBOPTS)", "-L$(LIBPATH)", "-l$(LIBS)", "-F$(FRAMEWORKPATH)", "-framework$(FRAMEWORKS)", "-o", "$(@)", "$(<)"},
		"PROGCOM":         {"$(LD)", "$(PROGOPTS)", "-L$(LIBPATH)", "-l$(LIBS)", "-F$(FRAMEWORKPATH)", "-framework$(FRAMEWORKS)", "-o", "$(@)", "$(<)"},
		"OBJCCOM":         {"$(CCCOM)"},
	})
}
