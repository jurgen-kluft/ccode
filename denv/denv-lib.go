package denv

import "github.com/jurgen-kluft/ccode/dev"

type DevLib struct {
	LibType      dev.LibraryType
	BuildType    dev.BuildType
	BuildConfigs *dev.BuildConfigList
	Files        []string
	Libs         []string
	Dir          string
}

func NewDevLib() *DevLib {
	return &DevLib{
		LibType:      dev.LibraryTypeUnknown,
		BuildType:    dev.BuildTypeUnknown,
		BuildConfigs: dev.NewBuildConfigList(),
		Files:        make([]string, 0),
		Libs:         make([]string, 0),
		Dir:          "",
	}
}
