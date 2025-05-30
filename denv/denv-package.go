package denv

import (
	"strings"

	"github.com/jurgen-kluft/ccode/dev"
)

// Package hold a defined set of 'Projects'
type Package struct {
	Name     string
	Packages map[string]*Package
	Projects map[string]*DevProject
}

func (p *Package) collectTypes(buildType dev.BuildType) []*DevProject {
	stack := make([]*DevProject, 0)
	for _, prj := range p.Projects {
		stack = append(stack, prj)
	}
	projects := make([]*DevProject, 0)
	projectMap := make(map[string]int)
	for len(stack) > 0 {
		prj := stack[0]
		stack = stack[1:]
		if _, ok := projectMap[prj.Name]; !ok {
			projectMap[prj.Name] = len(projects)
			if prj.BuildType&buildType != 0 {
				projects = append(projects, prj)
			}
			stack = append(stack, prj)
		}
		for _, dprj := range prj.Dependencies {
			if _, ok := projectMap[dprj.Name]; !ok {
				projectMap[dprj.Name] = len(projects)
				if dprj.BuildType&buildType != 0 {
					projects = append(projects, dprj)
				}
				stack = append(stack, dprj)
			}
		}
	}
	return projects
}

// Libraries returns all the libraries in the package
func (p *Package) Libraries() []*DevProject {
	return p.collectTypes(dev.BuildTypeStaticLibrary | dev.BuildTypeDynamicLibrary)
}

func (p *Package) MainProjects() []*DevProject {
	projects := make([]*DevProject, 0)
	if p.GetMainLib() != nil {
		projects = append(projects, p.GetMainLib()...)
	}
	if p.GetMainApp() != nil {
		projects = append(projects, p.GetMainApp()...)
	}
	if p.GetUnittest() != nil {
		projects = append(projects, p.GetUnittest()...)
	}
	return projects
}

// NewPackage creates a new empty package
func NewPackage(name string) *Package {
	return &Package{Name: name, Packages: make(map[string]*Package), Projects: make(map[string]*DevProject)}
}

func (p *Package) HasDependencyOn(name string) bool {
	for _, prj := range p.Projects {
		if prj.Name == name {
			return true
		}

		deps := prj.CollectProjectDependencies()
		for _, dep := range deps {
			if dep.Name == name {
				return true
			}
		}
	}
	return false
}

// AddPackage adds a package to this package
func (p *Package) AddPackage(pkg *Package) {
	p.Packages[pkg.Name] = pkg
}

// AddMainApp adds a project to this package as 'mainapp' the main application
func (p *Package) AddMainApp(prj *DevProject) {
	p.Projects[strings.ToLower(prj.Name)+".mainapp"] = prj
}

// AddMainLib adds a project to this package as 'mainlib' the main library
func (p *Package) AddMainLib(prj *DevProject) {
	p.Projects[strings.ToLower(prj.Name)+".mainlib"] = prj
}

// AddUnittest adds a project to this package as 'unittest' the unittest app
func (p *Package) AddUnittest(prj *DevProject) {
	p.Projects[prj.Name+".unittest"] = prj
}

// AddProject adds a project to this package
func (p *Package) AddProject(name string, prj *DevProject) {
	p.Projects[strings.ToLower(name)] = prj
}

// GetProject returns the project with the specific name, if it exists, if not nil is returned
func (p *Package) GetProject(name string) *DevProject {
	prj := p.Projects[name]
	return prj
}

// GetMainLib returns the projects that are registered as a main library
func (p *Package) GetMainLib() []*DevProject {
	mainlibs := make([]*DevProject, 0)
	for id, prj := range p.Projects {
		if strings.HasSuffix(id, ".mainlib") {
			mainlibs = append(mainlibs, prj)
		}
	}
	return mainlibs
}

// GetUnittest returns the projects that are registered as a unittest
func (p *Package) GetUnittest() []*DevProject {
	unittests := make([]*DevProject, 0)
	for id, prj := range p.Projects {
		if strings.HasSuffix(id, ".unittest") {
			unittests = append(unittests, prj)
		}
	}
	return unittests
}

// GetMainApp returns the projects that are registered as a main application
func (p *Package) GetMainApp() []*DevProject {
	mainapps := make([]*DevProject, 0)
	for id, prj := range p.Projects {
		if strings.HasSuffix(id, ".mainapp") {
			mainapps = append(mainapps, prj)
		}
	}
	return mainapps
}
