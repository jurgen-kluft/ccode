package vs

import (
	"github.com/jurgen-kluft/xcode/denv"
)

func IsVisualStudio(ide string) bool {
	vs := GetVisualStudio(ide)
	return vs != -1
}

func GetVisualStudio(ide string) denv.IDE {
	if ide == "VS2015" {
		return denv.VS2015
	} else if ide == "VS2013" {
		return denv.VS2013
	} else if ide == "VS2012" {
		return denv.VS2012
	}
	return -1
}

func Generate(ide denv.IDE, path string, targets []string, project *denv.Project) error {
	return nil
}
