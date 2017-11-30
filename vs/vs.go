package vs

import (
	"errors"

	"github.com/jurgen-kluft/xcode/denv"
)

// IsVisualStudio returns true if the incoming string @ide is equal to any of
// the Visual Studio formats that xcode supports.
func IsVisualStudio(ide string) bool {
	vs := GetVisualStudio(ide)
	return vs != -1
}

// GetVisualStudio returns a value for type IDE deduced from the incoming string @ide
func GetVisualStudio(ide string) denv.IDE {
	if ide == "VS2017" {
		return denv.VS2017
	} else if ide == "VS2015" {
		return denv.VS2015
	} else if ide == "VS2013" {
		return denv.VS2013
	} else if ide == "VS2012" {
		return denv.VS2012
	}
	return -1
}

// Generate will generate the Solution and Project files for the incoming project
func Generate(ide denv.IDE, path string, targets []string, project *denv.Project) error {
	switch ide {
	case denv.VS2015:
		GenerateVisualStudio2015Solution(project)
		return nil
	}
	return errors.New("IDE is not supported")
}
