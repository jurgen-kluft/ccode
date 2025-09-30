package denv

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

// Package holds sets of 'Projects'
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

func collectTypes(buildType BuildType, projects []*DevProject, list *DevProjectList) {
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

// Libraries returns all the libraries, including dependencies, in the package (used by fixer)
func (p *Package) Libraries() []*DevProject {
	list := NewDevProjectList()
	collectTypes(BuildTypeStaticLibrary|BuildTypeDynamicLibrary, p.MainLibs, list)
	collectTypes(BuildTypeStaticLibrary|BuildTypeDynamicLibrary, p.TestLibs, list)
	collectTypes(BuildTypeStaticLibrary|BuildTypeDynamicLibrary, p.MainApps, list)
	collectTypes(BuildTypeStaticLibrary|BuildTypeDynamicLibrary, p.Unittests, list)
	return list.Values

}

// Executables returns all the executable projects in the package (used by fixer)
func (p *Package) Executables() []*DevProject {
	projects := NewDevProjectList()
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

func (p *Package) EncodeJson(encoder *corepkg.JsonEncoder, key string) {
	encoder.BeginObject(key)
	{
		encoder.WriteField("root_path", p.RootPath)
		encoder.WriteField("repo_path", p.RepoPath)
		encoder.WriteField("repo_name", p.RepoName)

		if len(p.Packages) > 0 {
			encoder.BeginArray("packages")
			for _, pkg := range p.Packages {
				pkg.EncodeJson(encoder, pkg.RepoName)
			}
			encoder.EndObject()
		}

		if len(p.MainApps) > 0 {
			encoder.BeginArray("main_apps")
			for _, prj := range p.MainApps {
				prj.EncodeJson(encoder, "")
			}
			encoder.EndArray()
		}
		if len(p.MainLibs) > 0 {
			encoder.BeginArray("main_libs")
			for _, prj := range p.MainLibs {
				prj.EncodeJson(encoder, "")
			}
			encoder.EndArray()
		}
		if len(p.Unittests) > 0 {
			encoder.BeginArray("unittests")
			for _, prj := range p.Unittests {
				prj.EncodeJson(encoder, "")
			}
			encoder.EndArray()
		}
		if len(p.TestLibs) > 0 {
			encoder.BeginArray("test_libs")
			for _, prj := range p.TestLibs {
				prj.EncodeJson(encoder, "")
			}
			encoder.EndArray()
		}
	}
	encoder.EndObject()
}
