package vs

import (
	"errors"

	"github.com/jurgen-kluft/xcode/denv"
	"github.com/jurgen-kluft/xcode/vs2015"
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
	switch ide {
	case VS2015:
		vs2015.GenerateVisualStudio2015Solution(project)
		return nil
	}
	return errors.New("IDE is not supported")
}
