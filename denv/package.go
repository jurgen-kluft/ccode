package denv

import (
	"strings"
)

// Package hold a defined set of 'Projects'
type Package struct {
	Name     string
	Packages map[string]*Package
	Projects map[string]*Project
}

// NewPackage creates a new empty package
func NewPackage(name string) *Package {
	return &Package{Name: name, Packages: make(map[string]*Package), Projects: make(map[string]*Project)}
}

// AddPackage adds a package to this package
func (p *Package) AddPackage(pkg *Package) {
	p.Packages[pkg.Name] = pkg
}

// AddMainApp adds a project to this package as 'mainapp' the main application
func (p *Package) AddMainApp(prj *Project) {
	p.Projects["mainapp"] = prj
}

// AddMainLib adds a project to this package as 'mainlib' the main library
func (p *Package) AddMainLib(prj *Project) {
	p.Projects["mainlib"] = prj
}

// AddUnittest adds a project to this package as 'unittest' the unittest app
func (p *Package) AddUnittest(prj *Project) {
	p.Projects["unittest"] = prj
}

// AddProject adds a project to this package
func (p *Package) AddProject(name string, prj *Project) {
	p.Projects[strings.ToLower(name)] = prj
}

// GetMainLib returns the project that is registered as the main library
func (p *Package) GetMainLib() *Project {
	mainlib := p.Projects["mainlib"]
	return mainlib
}

// GetMainApp returns the project that is registered as the main application
func (p *Package) GetMainApp() *Project {
	mainlib := p.Projects["mainapp"]
	return mainlib
}
