package tctools

import (
	utils "github.com/jurgen-kluft/ccode/utils"
)

func ApplyGenericCpp(env *utils.Vars, options *utils.Vars) {

	env.SetMany(map[string][]string{
		"IGNORED_AUTOEXTS":   {"h", "hpp", "hh", "hxx", "inl", "natvis"},
		"CCEXTS":             {"c"},
		"CXXEXTS":            {"cpp", "cxx", "cc"},
		"OBJCEXTS":           {"m"},
		"PROGSUFFIX":         {"$(HOSTPROGSUFFIX)"},
		"SHLIBSUFFIX":        {"$(HOSTSHLIBSUFFIX)"},
		"CPPPATH":            {""},
		"CPPDEFS":            {""},
		"LIBS":               {""},
		"LIBPATH":            {"$(OBJECTDIR)"},
		"CCOPTS":             {""},
		"CXXOPTS":            {""},
		"CPPDEFS_DEBUG":      {""},
		"CPPDEFS_PRODUCTION": {""},
		"CPPDEFS_RELEASE":    {""},
		"CCOPTS_DEBUG":       {""},
		"CCOPTS_PRODUCTION":  {""},
		"CCOPTS_RELEASE":     {""},
		"CXXOPTS_DEBUG":      {""},
		"CXXOPTS_PRODUCTION": {""},
		"CXXOPTS_RELEASE":    {""},
		"SHLIBLINKSUFFIX":    {""},
	})
}
