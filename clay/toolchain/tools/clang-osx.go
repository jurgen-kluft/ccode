package tctools

import "github.com/jurgen-kluft/ccode/foundation"

func ApplyClangOsx(env *foundation.Vars, options *foundation.Vars) {
	env.Set("CC", "clang")
	env.Set("CXX", "clang++")
	env.Set("LD", "clang")
}
