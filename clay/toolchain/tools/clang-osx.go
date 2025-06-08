package tctools

import (
	utils "github.com/jurgen-kluft/ccode/utils"
)

func ApplyClangOsx(env *utils.Vars, options *utils.Vars) {
	env.Set("CC", "clang")
	env.Set("CXX", "clang++")
	env.Set("LD", "clang")
}
