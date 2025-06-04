package denv

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jurgen-kluft/ccode/dev"
)

// Package hold a defined set of 'Projects'
type Package struct {
	RootPath  string
	RepoPath  string
	RepoName  string
	Packages  map[string]*Package
	MainApps  []*DevProject
	MainLibs  []*DevProject
	Unittests []*DevProject
	TestLibs  []*DevProject
}

func (p *Package) WorkspacePath() string {
	return filepath.Join(p.RootPath, p.RepoPath)
}

func (p *Package) PackagePath() string {
	return filepath.Join(p.RootPath, p.RepoPath, p.RepoName)
}

func collectTypes(buildType dev.BuildType, projects []*DevProject, list *DevProjectList) {
	stack := slices.Clone(projects)
	for len(stack) > 0 {
		prj := stack[0]
		stack = stack[1:]
		if !list.Has(prj.Name) {
			if prj.BuildType&buildType != 0 {
				list.Add(prj)
			}
			stack = append(stack, prj)
		}
		for _, dprj := range prj.Dependencies.Values {
			if !list.Has(dprj.Name) {
				if dprj.BuildType&buildType != 0 {
					list.Add(dprj)
				}
				stack = append(stack, dprj)
			}
		}
	}
}

// Libraries returns all the libraries in the package
func (p *Package) Libraries() []*DevProject {
	list := NewDevProjectList()
	collectTypes(dev.BuildTypeStaticLibrary|dev.BuildTypeDynamicLibrary, p.MainApps, list)
	collectTypes(dev.BuildTypeStaticLibrary|dev.BuildTypeDynamicLibrary, p.MainLibs, list)
	collectTypes(dev.BuildTypeStaticLibrary|dev.BuildTypeDynamicLibrary, p.Unittests, list)
	collectTypes(dev.BuildTypeStaticLibrary|dev.BuildTypeDynamicLibrary, p.TestLibs, list)
	return list.Values

}

func (p *Package) MainProjects() []*DevProject {
	projects := NewDevProjectList()
	projects.AddMany(p.GetMainLib()...)
	projects.AddMany(p.GetMainApp()...)
	projects.AddMany(p.GetUnittest()...)
	return projects.Values
}

// NewPackage creates a new empty package
func NewPackage(repo_path string, repo_name string) *Package {
	repo_path = strings.ReplaceAll(repo_path, "\\", "/")
	rootPath := filepath.Join(os.Getenv("GOPATH"), "src")
	return &Package{
		RootPath:  rootPath,
		RepoPath:  repo_path,
		RepoName:  repo_name,
		Packages:  make(map[string]*Package),
		MainApps:  make([]*DevProject, 0),
		MainLibs:  make([]*DevProject, 0),
		Unittests: make([]*DevProject, 0),
		TestLibs:  make([]*DevProject, 0),
	}
}

func hasDependencyOn(name string, projects []*DevProject) bool {
	for _, prj := range projects {
		if prj.Package.RepoName == name {
			return true
		}

		deps := prj.CollectProjectDependencies()
		for _, dep := range deps.Values {
			if dep.Package.RepoName == name {
				return true
			}
		}
	}
	return false
}

func (p *Package) TestingHasDependencyOn(name string) bool {
	if hasDependencyOn(name, p.Unittests) {
		return true
	}
	if hasDependencyOn(name, p.TestLibs) {
		return true
	}
	return false
}

// AddPackage adds a package to this package
func (p *Package) AddPackage(pkg *Package) {
	p.Packages[pkg.RepoName] = pkg
}

// AddMainApp adds a project to this package under 'name.mainapp'
func (p *Package) AddMainApp(prj *DevProject) {
	p.MainApps = append(p.MainApps, prj)
}

// AddMainLib adds a project to this package under 'name.mainlib'
func (p *Package) AddMainLib(prj *DevProject) {
	p.MainLibs = append(p.MainLibs, prj)
}

// AddTestLib adds a project to this package under 'name.testlib'
func (p *Package) AddTestLib(prj *DevProject) {
	p.TestLibs = append(p.TestLibs, prj)
}

// AddUnittest adds a project to this package under 'name.unittest'
func (p *Package) AddUnittest(prj *DevProject) {
	p.Unittests = append(p.Unittests, prj)
}

// GetMainLib returns the projects that are registered as a main library
func (p *Package) GetMainLib() []*DevProject {
	return p.MainLibs
}

// GetTestLib returns the projects that are registered as a test library
func (p *Package) GetTestLib() []*DevProject {
	return p.TestLibs
}

// GetUnittest returns the projects that are registered as a unittest
func (p *Package) GetUnittest() []*DevProject {
	return p.Unittests
}

// GetMainApp returns the projects that are registered as a main application
func (p *Package) GetMainApp() []*DevProject {
	return p.MainApps
}
