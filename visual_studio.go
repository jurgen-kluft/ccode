package xcode

import (
	"fmt"
)

type VisualStudio interface {
	GenerateSolution(path string, pkg Package)
	GenerateProject(path string, pkg Package)
	GenerateFilters(path string, pkg Package)
}

type VisualStudio2015 struct {
}

type VisualStudioVersion int

const (
	VS2012 VisualStudioVersion = 2012
	VS2013 VisualStudioVersion = 2013
	VS2015 VisualStudioVersion = 2015
)

func NewVisualStudioGenerator(version VisualStudioVersion) (VisualStudio, error) {
	switch version {
	case VS2015:
		return &VisualStudio2015{}, nil
	}
	return nil, fmt.Errorf("Wrong visual studio version")
}

func (vs *VisualStudio2015) GenerateSolution(path string, pkg Package) {

	// Write out the sln

}

func (vs *VisualStudio2015) GenerateProject(path string, pkg Package) {

	// Write out the vcxproject

}

func (vs *VisualStudio2015) GenerateFilters(path string, pkg Package) {

	// Write out the vcxproject

}
