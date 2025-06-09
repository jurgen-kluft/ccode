package tctools

import utils "github.com/jurgen-kluft/ccode/utils"

func ApplyGcc(env *utils.Vars, options *utils.Vars) {

	ApplyGenericCpp(env, options)

	env.SetMany(map[string][]string{
		"NATIVE_SUFFIXES": {".c", ".cpp", ".cc", ".cxx", ".a", ".o"},
		"OBJECTSUFFIX":    {".o"},
		"LIBPREFIX":       {"lib"},
		"LIBSUFFIX":       {".a"},
		"_GCC_BINPREFIX":  {""},
		"CC":              {"$(_GCC_BINPREFIX)gcc"},
		"CXX":             {"$(_GCC_BINPREFIX)g++"},
		"LIB":             {"$(_GCC_BINPREFIX)ar"},
		"LD":              {"$(_GCC_BINPREFIX)gcc"},
		"_OS_CCOPTS":      {""},
		"_OS_CXXOPTS":     {""},
		"_PCH_SUPPORTED":  {"1"},
		"_PCH_SUFFIX":     {".gch"},
		"_PCH_WRITES_OBJ": {"0"},
		"_USE_PCH_OPT":    {"-include", "$(_PCH_INCLUDE_PATH)"},
		"_USE_PCH":        {""},
		"CCCOM":           {"$(CC)", "$(_OS_CCOPTS)", "-c", "-D$(CPPDEFS)", "-I$(CPPPATH:f)", "$(CCOPTS)", "$(CCOPTS_$(CURRENT_VARIANT:u))", "$(_USE_PCH)", "-o", "$(@)", "$(<)"},
		"CXXCOM":          {"$(CXX)", "$(_OS_CXXOPTS)", "-c", "-D$(CPPDEFS)", "-I$(CPPPATH:f)", "$(CXXOPTS)", "$(CXXOPTS_$(CURRENT_VARIANT:u))", "$(_USE_PCH)", "-o", "$(@)", "$(<)"},
		"PCHCOMPILE_CC":   {"$(CC)", "$(_OS_CCOPTS)", "-x", "c-header", "-c", "-D$(CPPDEFS)", "-I$(CPPPATH:f)", "$(CCOPTS)", "$(CCOPTS_$(CURRENT_VARIANT:u))", "-o", "$(@)", "$(<)"},
		"PCHCOMPILE_CXX":  {"$(CXX)", "$(_OS_CXXOPTS)", "-x", "c++-header", "-c", "-D$(CPPDEFS)", "-I$(CPPPATH:f)", "$(CXXOPTS)", "$(CXXOPTS_$(CURRENT_VARIANT:u))", "-o", "$(@)", "$(<)"},
		"PROGOPTS":        {""},
		"PROGCOM":         {"$(LD)", "$(PROGOPTS)", "-L$(LIBPATH)", "-o", "$(@)", "$(<)", "-l$(LIBS)"},
		"PROGPREFIX":      {""},
		"LIBOPTS":         {""},
		"LIBCOM":          {"$(LIB) -rs", "$(LIBOPTS)", "$(@)", "$(<)"},
		"SHLIBPREFIX":     {"lib"},
		"SHLIBOPTS":       {"-shared"},
		"SHLIBCOM":        {"$(LD)", "$(SHLIBOPTS)", "-L$(LIBPATH)", "-o", "$(@)", "$(<)", "-l$(LIBS)"},
	})

}
