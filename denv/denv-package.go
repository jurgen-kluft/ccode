package denv

import (
	"os"
	"path/filepath"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

// GetVarsFunc is a function that returns variables for a specific build target and build configuration, the
// hardwareId is used for Arduino to specify the board name.
type GetVarsFunc func(buildTarget BuildTarget, buildConfig BuildConfig, hardwareId string, vars *corepkg.Vars)

// Package holds sets of 'Projects'
type Package struct {
	RootPath  string
	RepoPath  string
	RepoName  string
	Packages  map[string]*Package
	MainApps  map[string]*DevProject
	MainLibs  map[string]*DevProject
	Library   map[string]*DevProject
	Unittests map[string]*DevProject
	TestLibs  map[string]*DevProject
	getVars   GetVarsFunc
}

type ArduinoConfig struct {
	PSRam           string // e.g. "opi or enabled"
	FlashMode       string // e.g. "dio"
	FlashSize       string // e.g. "4MB"
	PartitionScheme string // e.g. "huge_app"
	CPUFreq         string // e.g. "240"
	UploadSpeed     string // e.g. "921600"
	DebugLevel      string // e.g. "none", "error", "warn", "info", "debug", "verbose"
}

// Root path is the path where you have cloned many repositories, e.g. $GOPATH/src
// Repo path is the path to all the repositories relative to the root path, e.g. github.com/your_name
// Repo name is the name of the repository, e.g. your_repo

func (p *Package) Path() string {
	return filepath.Join(p.RootPath, p.RepoPath)
}

func iterateAllPackagesRecursive(pkg *Package, iterator func(pkg *Package), history map[string]bool) {
	if _, ok := history[pkg.RepoName]; ok {
		return
	}
	history[pkg.RepoName] = true
	for _, depPkg := range pkg.Packages {
		iterateAllPackagesRecursive(depPkg, iterator, history)
	}
	iterator(pkg)
}

func iterateAllPackages(pkg *Package, iterator func(pkg *Package)) {
	history := make(map[string]bool)
	iterateAllPackagesRecursive(pkg, iterator, history)
}

func iterateProjects(projects map[string]*DevProject, iterateProject func(prj *DevProject)) {
	projectStack := make([]*DevProject, 0)
	projectMap := make(map[string]bool)

	for _, prj := range projects {
		if _, ok := projectMap[prj.Name]; !ok {
			projectMap[prj.Name] = true
			projectStack = append(projectStack, prj)
		}
	}

	for len(projectStack) > 0 {
		prj := projectStack[0]
		projectStack = projectStack[1:]

		iterateProject(prj)

		// Add dependencies to the stack
		for _, dprj := range prj.Dependencies.Values {
			if _, ok := projectMap[dprj.Name]; !ok {
				projectMap[dprj.Name] = true
				projectStack = append(projectStack, dprj)
			}
		}
	}
}

// AllProjects returns all the projects, including dependencies, in the package
func (p *Package) AllProjects() []*DevProject {
	list := NewDevProjectList()
	addProjectFunc := func(prj *DevProject) { list.Add(prj) }
	iterateProjects(p.MainLibs, addProjectFunc)
	iterateProjects(p.TestLibs, addProjectFunc)
	iterateProjects(p.MainApps, addProjectFunc)
	iterateProjects(p.Unittests, addProjectFunc)
	return list.Values
}

// Libraries returns all the libraries, including dependencies, in the package
func (p *Package) Libraries() []*DevProject {
	list := NewDevProjectList()
	addProjectFunc := func(prj *DevProject) {
		if prj.BuildType&(BuildTypeStaticLibrary|BuildTypeDynamicLibrary) != 0 {
			list.Add(prj)
		}
	}
	iterateProjects(p.MainLibs, addProjectFunc)
	iterateProjects(p.TestLibs, addProjectFunc)
	return list.Values
}

// Executables returns all the executable projects in the package
func (p *Package) Executables() []*DevProject {
	projects := NewDevProjectList()
	addProjectFunc := func(prj *DevProject) {
		if prj.BuildType&(BuildTypeUnittest|BuildTypeCli|BuildTypeApplication) != 0 {
			projects.Add(prj)
		}
	}
	iterateProjects(p.MainApps, addProjectFunc)
	iterateProjects(p.Unittests, addProjectFunc)
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
		MainApps:  make(map[string]*DevProject, 0),
		MainLibs:  make(map[string]*DevProject, 0),
		Library:   make(map[string]*DevProject, 0),
		Unittests: make(map[string]*DevProject, 0),
		TestLibs:  make(map[string]*DevProject, 0),
		getVars:   nil,
	}
}

func hasDependencyOn(name string, projects map[string]*DevProject) bool {
	for _, prj := range projects {
		if prj.RepoName == name {
			return true
		}

		deps := NewDevProjectList()
		prj.CollectProjectDependencies(deps)
		for _, dep := range deps.Values {
			if dep.RepoName == name {
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

type Var struct {
	Config string
	Value  []string
	Append bool
}

// AddVars adds variables to this package for a specific build target
func (p *Package) SetGetVarsFunc(getVars GetVarsFunc) {
	p.getVars = getVars
}

func (p *Package) GetVars(buildTarget BuildTarget, buildConfig BuildConfig, hardwareId string, vars *corepkg.Vars) {
	iterateAllPackages(p, func(pkg *Package) {
		if pkg.getVars != nil {
			pkg.getVars(buildTarget, buildConfig, hardwareId, vars)
		}
	})
}

// AddMainApp adds a project to this package under 'name.mainapp'
func (p *Package) AddMainApp(prj *DevProject) {
	p.MainApps[prj.Name] = prj
}

// AddMainLib adds a project to this package under 'name.mainlib'
func (p *Package) AddMainLib(prj *DevProject) {
	//p.MainLibs = append(p.MainLibs, prj)
	p.MainLibs[prj.Name] = prj
}

// AddLibrary adds a library to this package under its own name, these
// libraries are not automatically added by consumers of this package.
// Instead a consumer must explicitly add these libraries as dependency
// by name.
func (p *Package) AddLibrary(prj *DevProject) {
	p.Library[prj.Name] = prj
}

// AddTestLib adds a project to this package under 'name.testlib'
func (p *Package) AddTestLib(prj *DevProject) {
	//p.TestLibs = append(p.TestLibs, prj)
	p.TestLibs[prj.Name] = prj
}

// AddUnittest adds a project to this package under 'name.unittest'
func (p *Package) AddUnittest(prj *DevProject) {
	//p.Unittests = append(p.Unittests, prj)
	p.Unittests[prj.Name] = prj
}

// GetMainLib returns the projects that are registered as a main library
func (p *Package) GetMainLib() map[string]*DevProject {
	return p.MainLibs
}

// GetLibrary returns a project that is registered as a library, it checks
// all registered packages as well.
func (p *Package) GetLibrary(name string) *DevProject {
	if prj, ok := p.Library[name]; ok {
		return prj
	}
	for _, pkg := range p.Packages {
		if prj, ok := pkg.Library[name]; ok {
			return prj
		}
	}
	return nil
}

// GetTestLib returns the projects that are registered as a test library
func (p *Package) GetTestLib() map[string]*DevProject {
	return p.TestLibs
}

// GetUnittest returns the projects that are registered as a unittest
func (p *Package) GetUnittest() map[string]*DevProject {
	return p.Unittests
}

// GetMainApp returns the projects that are registered as a main application
func (p *Package) GetMainApp() map[string]*DevProject {
	return p.MainApps
}

func (p *Package) GetLibraries() map[string]*DevProject {
	return p.Library
}

func (p *Package) CollectAllPackages(collectPackages func(pkg *Package)) {
	collectPackages(p)
	for _, pkg := range p.Packages {
		collectPackages(pkg)
	}
}

// func (p *Package) SaveJson(filepath string) error {
// 	encoder := corepkg.NewJsonEncoder("    ")
// 	encoder.Begin()
// 	{
// 		encoder.BeginObject("")
// 		{
// 			encoder.WriteFieldString("main", p.RepoName)

// 			packages := make([]*Package, 0, 32)
// 			iterateAllPackages(p, func(pkg *Package) { packages = append(packages, pkg) })

// 			projects := make(map[string]*DevProject)
// 			encoder.BeginArray("packages")
// 			{
// 				for _, pkg := range packages {
// 					pkg.encodeJson(encoder, "", &projects)
// 				}
// 			}
// 			encoder.EndArray()

// 			// Sort projects by name
// 			projectNames := make([]string, 0, len(projects))
// 			for name := range projects {
// 				projectNames = append(projectNames, name)
// 			}
// 			slices.Sort(projectNames)

// 			encoder.BeginArray("projects")
// 			{
// 				for _, name := range projectNames {
// 					prj := projects[name]
// 					prj.EncodeJson(encoder, "")
// 				}
// 			}
// 			encoder.EndArray()
// 		}
// 		encoder.EndObject()
// 	}
// 	json := encoder.End()

// 	return corepkg.FileOpenWriteClose(filepath, []byte(json))
// }

// func LoadPackageFromJson(filepath string) (*Package, error) {
// 	data, err := corepkg.FileOpenReadClose(filepath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	decoder := corepkg.NewJsonDecoder()

// 	mainPkg := ""
// 	packages := make(map[string]*Package)
// 	projects := make(map[string]*DevProject)
// 	decoder.Begin(string(data))
// 	{
// 		fields := map[string]corepkg.JsonDecode{
// 			"main": func(decoder *corepkg.JsonDecoder) { mainPkg = decoder.DecodeString() },
// 			"packages": func(decoder *corepkg.JsonDecoder) {
// 				decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
// 					pkg := decodeJsonPackage(decoder)
// 					packages[pkg.RepoName] = pkg
// 				})
// 			},
// 			"projects": func(decoder *corepkg.JsonDecoder) {
// 				decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
// 					prj := DecodeJsonDevProject(decoder)
// 					projects[prj.Name] = prj
// 				})
// 			},
// 		}
// 		if err := decoder.Decode(fields); err != nil {
// 			corepkg.LogErrorf(err, "error decoding Package: %s", err.Error())
// 		}
// 	}
// 	decoder.End()

// 	// For every project process the array of dependency names and populate the Dependencies list
// 	for _, prj := range projects {
// 		for i, name := range prj.Dependencies.Keys {
// 			if dep, ok := projects[name]; ok {
// 				prj.Dependencies.Values[i] = dep
// 			} else {
// 				corepkg.LogErrorf(nil, "error: project '%s' depends on unknown project '%s'", prj.Name, name)
// 			}
// 		}
// 	}

// 	// For every package process the array of packages and get them
// 	for _, pkg := range packages {
// 		for k := range pkg.Packages {
// 			pkg.Packages[k] = packages[k]
// 		}

// 		for i := range pkg.MainApps {
// 			if prj, ok := projects[i]; ok {
// 				pkg.MainApps[i] = prj
// 			} else {
// 				corepkg.LogErrorf(nil, "error: package '%s' has unknown main app project '%s'", pkg.RepoName, i)
// 			}
// 		}
// 		for i := range pkg.MainLibs {
// 			if prj, ok := projects[i]; ok {
// 				pkg.MainLibs[i] = prj
// 			} else {
// 				corepkg.LogErrorf(nil, "error: package '%s' has unknown main lib project '%s'", pkg.RepoName, i)
// 			}
// 		}
// 		for i := range pkg.Unittests {
// 			if prj, ok := projects[i]; ok {
// 				pkg.Unittests[i] = prj
// 			} else {
// 				corepkg.LogErrorf(nil, "error: package '%s' has unknown unittest project '%s'", pkg.RepoName, i)
// 			}
// 		}
// 		for i := range pkg.TestLibs {
// 			if prj, ok := projects[i]; ok {
// 				pkg.TestLibs[i] = prj
// 			} else {
// 				corepkg.LogErrorf(nil, "error: package '%s' has unknown test lib project '%s'", pkg.RepoName, i)
// 			}
// 		}
// 	}

// 	return packages[mainPkg], nil
// }

// func (p *Package) encodeJson(encoder *corepkg.JsonEncoder, key string, projects *map[string]*DevProject) {
// 	encoder.BeginObject(key)
// 	{
// 		encoder.WriteField("root_path", p.RootPath)
// 		encoder.WriteField("repo_path", p.RepoPath)
// 		encoder.WriteField("repo_name", p.RepoName)

// 		if len(p.Packages) > 0 {
// 			encoder.BeginArray("packages")
// 			for _, pkg := range p.Packages {
// 				encoder.WriteArrayElement(pkg.RepoName)
// 			}
// 			encoder.EndArray()
// 		}

// 		if len(p.MainApps) > 0 {
// 			encoder.BeginArray("main_apps")
// 			for _, prj := range p.MainApps {
// 				if _, ok := (*projects)[prj.Name]; !ok {
// 					(*projects)[prj.Name] = prj
// 				}
// 				encoder.WriteArrayElement(prj.Name)
// 			}
// 			encoder.EndArray()
// 		}
// 		if len(p.MainLibs) > 0 {
// 			encoder.BeginArray("main_libs")
// 			for _, prj := range p.MainLibs {
// 				if _, ok := (*projects)[prj.Name]; !ok {
// 					(*projects)[prj.Name] = prj
// 				}
// 				encoder.WriteArrayElement(prj.Name)
// 			}
// 			encoder.EndArray()
// 		}
// 		if len(p.Unittests) > 0 {
// 			encoder.BeginArray("unittests")
// 			for _, prj := range p.Unittests {
// 				if _, ok := (*projects)[prj.Name]; !ok {
// 					(*projects)[prj.Name] = prj
// 				}
// 				encoder.WriteArrayElement(prj.Name)
// 			}
// 			encoder.EndArray()
// 		}
// 		if len(p.TestLibs) > 0 {
// 			encoder.BeginArray("test_libs")
// 			for _, prj := range p.TestLibs {
// 				if _, ok := (*projects)[prj.Name]; !ok {
// 					(*projects)[prj.Name] = prj
// 				}
// 				encoder.WriteArrayElement(prj.Name)
// 			}
// 			encoder.EndArray()
// 		}
// 	}
// 	encoder.EndObject()
// }

// func decodeJsonPackage(decoder *corepkg.JsonDecoder) *Package {
// 	pkg := &Package{
// 		RootPath:  "",
// 		RepoPath:  "",
// 		RepoName:  "",
// 		Packages:  make(map[string]*Package),
// 		MainApps:  make(map[string]*DevProject, 0),
// 		MainLibs:  make(map[string]*DevProject, 0),
// 		Unittests: make(map[string]*DevProject, 0),
// 		TestLibs:  make(map[string]*DevProject, 0),
// 	}

// 	fields := map[string]corepkg.JsonDecode{
// 		"root_path": func(decoder *corepkg.JsonDecoder) { pkg.RootPath = decoder.DecodeString() },
// 		"repo_path": func(decoder *corepkg.JsonDecoder) { pkg.RepoPath = decoder.DecodeString() },
// 		"repo_name": func(decoder *corepkg.JsonDecoder) { pkg.RepoName = decoder.DecodeString() },
// 		"packages": func(decoder *corepkg.JsonDecoder) {
// 			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
// 				pkgName := decoder.DecodeString()
// 				pkg.Packages[pkgName] = nil
// 			})
// 		},
// 		"main_apps": func(decoder *corepkg.JsonDecoder) {
// 			pkg.MainApps = make(map[string]*DevProject, 0)
// 			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
// 				prjName := decoder.DecodeString()
// 				pkg.MainApps[prjName] = nil
// 			})
// 		},
// 		"main_libs": func(decoder *corepkg.JsonDecoder) {
// 			pkg.MainLibs = make(map[string]*DevProject, 0)
// 			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
// 				prjName := decoder.DecodeString()
// 				pkg.MainLibs[prjName] = nil
// 			})
// 		},
// 		"unittests": func(decoder *corepkg.JsonDecoder) {
// 			pkg.Unittests = make(map[string]*DevProject, 0)
// 			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
// 				prjName := decoder.DecodeString()
// 				pkg.Unittests[prjName] = nil
// 			})
// 		},
// 		"test_libs": func(decoder *corepkg.JsonDecoder) {
// 			pkg.TestLibs = make(map[string]*DevProject, 0)
// 			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
// 				prjName := decoder.DecodeString()
// 				pkg.TestLibs[prjName] = nil
// 			})
// 		},
// 	}
// 	if err := decoder.Decode(fields); err != nil {
// 		corepkg.LogErrorf(err, "error decoding Package: %s", err.Error())
// 	}

// 	return pkg
// }
