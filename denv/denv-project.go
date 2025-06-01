package denv

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/dev"
	utils "github.com/jurgen-kluft/ccode/utils"
)

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------
type DevProjectList struct {
	Dict   map[string]int
	Values []*DevProject
	Keys   []string
}

func NewDevProjectList() *DevProjectList {
	return &DevProjectList{
		Dict:   map[string]int{},
		Values: []*DevProject{},
		Keys:   []string{},
	}
}

func (p *DevProjectList) Len() int {
	return len(p.Values)
}

func (p *DevProjectList) IsEmpty() bool {
	return len(p.Values) == 0
}

func (p *DevProjectList) Add(project *DevProject) {
	if _, ok := p.Dict[project.Name]; !ok {
		p.Dict[project.Name] = len(p.Values)
		p.Values = append(p.Values, project)
		p.Keys = append(p.Keys, project.Name)
	}
}

func (p *DevProjectList) Has(name string) bool {
	_, ok := p.Dict[name]
	return ok
}

func (p *DevProjectList) Get(name string) (*DevProject, bool) {
	if i, ok := p.Dict[name]; ok {
		return p.Values[i], true
	}
	return nil, false
}

func (p *DevProjectList) CollectByWildcard(name string, list *DevProjectList) {
	for _, p := range p.Values {
		if utils.PathMatchWildcard(p.Name, name, true) {
			list.Add(p)
		}
	}
}

func (p *DevProjectList) TopoSort() error {
	var edges []Edge

	// Sort the projects by dependencies
	for i, project := range p.Values {
		if project.Dependencies.IsEmpty() {
			edges = append(edges, Edge{Vertex(i), InvalidVertex})
		} else {
			for _, dep := range project.Dependencies.Values {
				edges = append(edges, Edge{S: Vertex(i), D: Vertex(p.Dict[dep.Name])})
			}
		}
	}

	sorted, err := Toposort(edges)
	if err != nil {
		return err
	}

	var sortedProjects []*DevProject
	for i := len(sorted) - 1; i >= 0; i-- {
		sortedProjects = append(sortedProjects, p.Values[sorted[i]])
	}

	p.Dict = map[string]int{}
	p.Values = sortedProjects
	p.Keys = []string{}

	for i, project := range sortedProjects {
		p.Dict[project.Name] = i
		p.Keys = append(p.Keys, project.Name)
	}

	return nil
}

// DevProject is a structure that holds all the information that defines a project in an IDE
type DevProject struct {
	Package      *Package
	Name         string
	BuildType    dev.BuildType
	Supported    dev.BuildTarget
	Vars         map[string]string
	SourceDirs   []dev.PinPathGlob
	Configs      []*DevConfig
	Dependencies *DevProjectList
}

func NewProject(pkg *Package, name string) *DevProject {
	return &DevProject{
		Package:      pkg,
		Name:         name,
		BuildType:    dev.BuildTypeUnknown,
		Supported:    dev.EmptyBuildTarget,
		Vars:         make(map[string]string),
		Configs:      []*DevConfig{},
		Dependencies: NewDevProjectList(),
	}
}

func (prj *DevProject) AddEnvironmentVariable(ev string) {
	// Environment variable should exist
	if value, ok := os.LookupEnv(ev); ok {
		prj.Vars[strings.ToLower(ev)] = value
	}
}

func (prj *DevProject) ResolveEnvironmentVariables(str string) string {
	// Replace all environment variables in the string
	// Variables can be nested, so we need to know when a replace was
	// done, and repeat the replace until no more replacements are done.
	for {
		// Check if there are any environment variables in the string
		start := strings.Index(str, "{")
		if start < 0 {
			break
		}
		end := strings.Index(str[start:], "}")
		if end < 0 {
			break
		}

		if value, ok := prj.Vars[strings.ToLower(str[start+1:end])]; ok {
			str = strings.ReplaceAll(str, str[start:end+1], value)
			end = start
			continue
		}
		break
	}
	return str
}

func (prj *DevProject) AddDependency(dep *DevProject) {
	if dep != nil {
		prj.Dependencies.Add(dep)
	}
}

func (prj *DevProject) AddDependencies(deps ...*DevProject) {
	for _, dep := range deps {
		if dep != nil {
			prj.Dependencies.Add(dep)
		}
	}
}

func (prj *DevProject) AddInclude(root string, base string, sub string) {
	root = prj.ResolveEnvironmentVariables(root)
	base = prj.ResolveEnvironmentVariables(base)
	sub = prj.ResolveEnvironmentVariables(sub)
	for _, cfg := range prj.Configs {
		cfg.IncludeDirs = append(cfg.IncludeDirs, dev.PinPath{Root: root, Base: base, Sub: sub})
	}
}

func (prj *DevProject) SourceFilesFrom(root, base, sub string) {
	root = prj.ResolveEnvironmentVariables(root)
	base = prj.ResolveEnvironmentVariables(base)
	sub = prj.ResolveEnvironmentVariables(sub)
	prj.SourceDirs = append(prj.SourceDirs, dev.PinPathGlob{Path: dev.PinPath{Root: root, Base: base, Sub: sub}, Glob: "**/*.cpp"})
}

func (prj *DevProject) AddDefine(define string) {
	define = prj.ResolveEnvironmentVariables(define)
	for _, cfg := range prj.Configs {
		cfg.Defines.Add(define)
	}
}

func (prj *DevProject) AddLib(path string, file string) {
	for _, cfg := range prj.Configs {
		lib := dev.PinFilepath{Path: dev.PinPath{Root: prj.ResolveEnvironmentVariables(path), Base: "", Sub: ""},
			Filename: file,
		}
		cfg.Libs = append(cfg.Libs, lib)
	}
}

// Used by IncludeFixer
func (proj *DevProject) CollectIncludeDirs() []dev.PinPath {
	includes := make([]dev.PinPath, 0)
	history := make(map[string]bool)
	for _, cfg := range proj.Configs {
		for _, inc := range cfg.IncludeDirs {
			if inc.Base != "" && inc.Sub != "" {
				// Resolve the base and sub path
				base := proj.ResolveEnvironmentVariables(inc.Base)
				sub := proj.ResolveEnvironmentVariables(inc.Sub)
				fullPath := filepath.Join(base, sub)

				// Check if we have already added this path
				if !history[fullPath] {
					history[fullPath] = true
					inc.Base = base
					inc.Sub = sub
					includes = append(includes, inc)
				}
			}
		}
	}
	return includes
}

func (proj *DevProject) CollectSourceDirs() []dev.PinPathGlob {
	sourceDirs := make([]dev.PinPathGlob, 0)
	history := make(map[string]bool)
	for _, dir := range proj.SourceDirs {
		// Resolve the base and sub path
		root := proj.ResolveEnvironmentVariables(dir.Path.Root)
		base := proj.ResolveEnvironmentVariables(dir.Path.Base)
		sub := proj.ResolveEnvironmentVariables(dir.Path.Sub)
		fullPath := filepath.Join(base, sub)
		// Check if we have already added this path
		if !history[fullPath] {
			history[fullPath] = true
			ppg := dev.PinPathGlob{Path: dev.PinPath{Root: root, Base: base, Sub: sub}, Glob: dir.Glob}
			sourceDirs = append(sourceDirs, ppg)
		}
	}
	return sourceDirs
}

func (proj *DevProject) CollectProjectDependencies() *DevProjectList {

	// Traverse and collect all dependencies
	list := NewDevProjectList()
	for _, dp := range proj.Dependencies.Values {
		list.Add(dp)
	}

	i := 0
	for i < list.Len() {
		cp := list.Values[i]
		for _, dp := range cp.Dependencies.Values {
			list.Add(dp)
		}
		i++
	}
	return list
}

// SetupDefaultCppLibProject returns a default C++ library project, since such a project can be used by
// an application as well as an unittest we need to add the appropriate configurations.
// Example:
//
//	SetupDefaultCppLibProject("cbase", "github.com/jurgen-kluft")
func SetupDefaultCppLibProject(pkg *Package, name string, dir string, buildTarget dev.BuildTarget) *DevProject {
	project := NewProject(pkg, name)
	project.BuildType = dev.BuildTypeStaticLibrary
	project.Dependencies = NewDevProjectList()

	project.SourceDirs = append(project.SourceDirs, dev.PinPathGlob{Path: dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/" + dir + "/cpp"}, Glob: "**/*.c"})
	project.SourceDirs = append(project.SourceDirs, dev.PinPathGlob{Path: dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/" + dir + "/cpp"}, Glob: "**/*.cpp"})

	return project
}

func SetupCppLibProject(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppLibProject(pkg, "library_"+name, "main", dev.GetBuildTarget())
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeStaticLibrary, dev.NewDebugDevConfig()))
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeStaticLibrary, dev.NewReleaseDevConfig()))
	project.Supported = dev.BuildTargetsAll
	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/main/include"})
	}
	return project
}

func SetupCppTestLibProject(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppLibProject(pkg, "unittest_library_"+name, "main", dev.GetBuildTarget())
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeStaticLibrary, dev.NewDebugDevTestConfig()))
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeStaticLibrary, dev.NewReleaseDevTestConfig()))
	project.Supported = dev.BuildTargetsAll
	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/main/include"})
	}
	return project
}

func SetupCppLibProjectForDesktop(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppLibProject(pkg, "program_"+name, "main", dev.GetBuildTarget())
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeStaticLibrary, dev.NewDebugDevConfig()))
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeStaticLibrary, dev.NewReleaseDevConfig()))
	project.Supported = dev.BuildTargetsDesktop
	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/main/include"})
	}
	return project
}

func SetupCppLibProjectForArduino(pkg *Package, name string) *DevProject {
	// Arduino Esp32
	project := SetupDefaultCppLibProject(pkg, "library_"+name, "main", dev.BuildTargetArduinoEsp32)
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeStaticLibrary, dev.NewDebugDevConfig()))
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeStaticLibrary, dev.NewReleaseDevConfig()))
	project.Supported = dev.BuildTargetsArduino
	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/main/include"})
	}
	return project
}

func SetupDefaultCppTestProject(pkg *Package, name string, buildTarget dev.BuildTarget) *DevProject {
	project := NewProject(pkg, name)
	project.BuildType = dev.BuildTypeUnittest
	project.Supported = dev.BuildTargetsDesktop
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeUnittest, dev.NewDebugDevTestConfig()))
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeUnittest, dev.NewReleaseDevTestConfig()))
	project.Dependencies = NewDevProjectList()

	project.SourceDirs = append(project.SourceDirs, dev.PinPathGlob{Path: dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/test/cpp"}, Glob: "**/*.c"})
	project.SourceDirs = append(project.SourceDirs, dev.PinPathGlob{Path: dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/test/cpp"}, Glob: "**/*.cpp"})

	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/test/include"})
		cfg.IncludeDirs = append(cfg.IncludeDirs, dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/main/include"})
	}

	return project
}

func SetupCppTestProject(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppTestProject(pkg, "unittest_"+name, dev.GetBuildTarget())
	project.Supported = dev.BuildTargetsDesktop
	return project
}

func SetupCppTestProjectForDesktop(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppTestProject(pkg, "unittest_"+name, dev.GetBuildTarget())
	project.Supported = dev.BuildTargetsDesktop
	return project
}

// SetupDefaultCppCliProject returns a default C++ command-line program project
// Example:
//
//	SetupDefaultCppCliProject("cmycli", "github.com\\jurgen-kluft")
func SetupDefaultCppCliProject(pkg *Package, name string, buildTarget dev.BuildTarget) *DevProject {
	project := NewProject(pkg, name)
	project.BuildType = dev.BuildTypeCli
	project.Supported = dev.BuildTargetsDesktop
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeCli, dev.NewDebugDevConfig()))
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeCli, dev.NewReleaseDevConfig()))
	project.Dependencies = NewDevProjectList()

	project.SourceDirs = append(project.SourceDirs, dev.PinPathGlob{Path: dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/cli/cpp"}, Glob: "**/*.c"})
	project.SourceDirs = append(project.SourceDirs, dev.PinPathGlob{Path: dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/cli/cpp"}, Glob: "**/*.cpp"})

	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/cli/include"})
	}

	return project
}

// SetupDefaultCppAppProject returns a default C++ application project
// Example:
//
//	SetupDefaultCppAppProject("cmyapp", "github.com\\jurgen-kluft")
func SetupDefaultCppAppProject(pkg *Package, name string, buildTarget dev.BuildTarget) *DevProject {
	project := NewProject(pkg, name)
	project.BuildType = dev.BuildTypeApplication
	project.Supported = dev.BuildTargetsDesktop
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeApplication, dev.NewDebugDevConfig()))
	project.Configs = append(project.Configs, NewDevConfig(dev.BuildTypeApplication, dev.NewReleaseDevConfig()))
	project.Dependencies = NewDevProjectList()

	project.SourceDirs = append(project.SourceDirs, dev.PinPathGlob{Path: dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/app/cpp"}, Glob: "**/*.c"})
	project.SourceDirs = append(project.SourceDirs, dev.PinPathGlob{Path: dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/app/cpp"}, Glob: "**/*.cpp"})

	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, dev.PinPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/app/include"})
	}

	return project
}

func SetupCppAppProject(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppAppProject(pkg, "app_"+name, dev.GetBuildTarget())
	project.Supported = dev.BuildTargetsAll
	return project
}

func SetupCppAppProjectForDesktop(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppAppProject(pkg, "app_"+name, dev.GetBuildTarget())
	project.Supported = dev.BuildTargetsDesktop
	return project
}

func SetupCppAppProjectForArduino(pkg *Package, name string) *DevProject {
	// Arduino project
	project := SetupDefaultCppAppProject(pkg, "app_"+name, dev.BuildTargetArduinoEsp32)
	project.Supported = dev.BuildTargetsArduino
	return project
}

func configureProjectCompilerDefines(config *DevConfig) {
	configType := config.BuildConfig
	if configType.IsDebug() {
		config.Defines.AddMany("TARGET_DEBUG", "_DEBUG")
	} else if configType.IsRelease() {
		config.Defines.AddMany("TARGET_RELEASE", "NDEBUG")
	} else if configType.IsFinal() {
		config.Defines.AddMany("TARGET_FINAL", "NDEBUG")
	}

	if configType.IsProfile() {
		config.Defines.AddMany("TARGET_RELEASE", "TARGET_PROFILE", "NDEBUG")
	}

	if configType.IsTest() {
		config.Defines.AddMany("TARGET_TEST")
	}
}
