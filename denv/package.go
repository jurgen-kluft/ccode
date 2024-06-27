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

func (p *Package) collect(projectTypesToCollect ProjectType) []*Project {
	stack := make([]*Project, 0)
	for _, prj := range p.Projects {
		stack = append(stack, prj)
	}
	projects := make([]*Project, 0)
	projectMap := make(map[string]int)
	for len(stack) > 0 {
		prj := stack[0]
		stack = stack[1:]
		if _, ok := projectMap[prj.Name]; !ok {
			projectMap[prj.Name] = len(projects)
			if prj.Type&projectTypesToCollect != 0 {
				projects = append(projects, prj)
			}
			stack = append(stack, prj)
		}
		for _, dprj := range prj.Dependencies {
			if _, ok := projectMap[dprj.Name]; !ok {
				projectMap[dprj.Name] = len(projects)
				if dprj.Type&projectTypesToCollect != 0 {
					projects = append(projects, dprj)
				}
				stack = append(stack, dprj)
			}
		}
	}
	return projects
}

// Libraries returns all the libraries in the package
func (p *Package) Libraries() []*Project {
	return p.collect(StaticLibrary | SharedLibrary)
}

// Applications returns all the applications in the package
func (p *Package) Applications() []*Project {
	return p.collect(Executable)
}

func (p *Package) MainProjects() []*Project {
	projects := make([]*Project, 0)
	if p.GetMainLib() != nil {
		projects = append(projects, p.GetMainLib())
	}
	if p.GetMainApp() != nil {
		projects = append(projects, p.GetMainApp())
	}
	if p.GetUnittest() != nil {
		projects = append(projects, p.GetUnittest())
	}
	return projects
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

// GetProject returns the project with the specific name, if it exists, if not nil is returned
func (p *Package) GetProject(name string) *Project {
	prj := p.Projects[name]
	return prj
}

// GetMainLib returns the project that is registered as the main library
func (p *Package) GetMainLib() *Project {
	prj := p.Projects["mainlib"]
	return prj
}

// GetUnittest returns the project that is registered as the unittest
func (p *Package) GetUnittest() *Project {
	prj := p.Projects["unittest"]
	return prj
}

// GetMainApp returns the project that is registered as the main application
func (p *Package) GetMainApp() *Project {
	prj := p.Projects["mainapp"]
	return prj
}
