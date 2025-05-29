package denv

import "github.com/jurgen-kluft/ccode/dev"

type DevLib struct {
	LibType      dev.LibType
	BuildType    dev.BuildType
	BuildConfigs dev.BuildConfig
	Files        []string
	Libs         []string
	Dir          string
}
