package denv

import (
	"os"
	"path/filepath"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

// DevProject is a structure that holds all the information that defines a project in an IDE
type DevProject struct {
	Package      *Package
	Name         string
	DirName      string
	BuildType    BuildType
	BuildTargets BuildTarget
	EnvVars      map[string]string
	SourceDirs   []PinnedGlobPath
	Configs      []*DevConfig
	Dependencies *DevProjectList
}

func NewProject(pkg *Package, name string, dirname string) *DevProject {
	return &DevProject{
		Package:      pkg,
		Name:         name,
		DirName:      dirname,
		BuildType:    BuildTypeUnknown,
		BuildTargets: EmptyBuildTarget,
		EnvVars:      make(map[string]string),
		Configs:      []*DevConfig{},
		Dependencies: NewDevProjectList(),
	}
}

func (prj *DevProject) AddEnvironmentVariable(ev string) {
	// Environment variable should exist
	if value, ok := os.LookupEnv(ev); ok {
		prj.EnvVars[strings.ToLower(ev)] = value
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

		if value, ok := prj.EnvVars[strings.ToLower(str[start+1:end])]; ok {
			str = strings.ReplaceAll(str, str[start:end+1], value)
			end = start
			continue
		}
		break
	}
	return str
}

func (prj *DevProject) AddDependency(dep *DevProject) {
	if strings.HasPrefix(prj.Name, "unittest_") {
		// The dependency project has to be a unittest library, if not panic
		if dep != nil && strings.HasPrefix(dep.Name, "unittest_library_") {
			prj.Dependencies.Add(dep)
		} else {
			panic("Cannot add dependency " + dep.Name + " to project " + prj.Name + ", because it is not a unittest library")
		}
	} else {
		// The dependency project can be any type of project, except a unittest project
		if dep != nil && !strings.HasPrefix(dep.Name, "unittest_") {
			prj.Dependencies.Add(dep)
		} else {
			panic("Cannot add dependency " + dep.Name + " to project " + prj.Name + ", because it is a unittest project")
		}
	}
}

func (prj *DevProject) AddDependencies(deps ...*DevProject) {
	for _, dep := range deps {
		prj.AddDependency(dep)
	}
}

func (prj *DevProject) ClearIncludes() {
	for _, cfg := range prj.Configs {
		cfg.IncludeDirs = make([]PinnedPath, 0)
	}
}

func (prj *DevProject) AddInclude(root string, base string, sub string) {
	root = prj.ResolveEnvironmentVariables(root)
	base = prj.ResolveEnvironmentVariables(base)
	sub = prj.ResolveEnvironmentVariables(sub)
	for _, cfg := range prj.Configs {
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: root, Base: base, Sub: sub})
	}
}

func (prj *DevProject) ClearSourcePaths() {
	prj.SourceDirs = make([]PinnedGlobPath, 0)
}

func (prj *DevProject) SourceFilesFrom(root, base, sub string) {
	root = prj.ResolveEnvironmentVariables(root)
	base = prj.ResolveEnvironmentVariables(base)
	sub = prj.ResolveEnvironmentVariables(sub)
	prj.SourceDirs = append(prj.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: root, Base: base, Sub: sub}, Glob: "**/*.c"})
	prj.SourceDirs = append(prj.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: root, Base: base, Sub: sub}, Glob: "**/*.cpp"})
}

func (prj *DevProject) AddDefine(define string) {
	define = prj.ResolveEnvironmentVariables(define)
	for _, cfg := range prj.Configs {
		cfg.Defines.Add(define)
	}
}

func (prj *DevProject) AddLib(path string, file string) {
	for _, cfg := range prj.Configs {
		lib := PinnedFilepath{Path: PinnedPath{Root: prj.ResolveEnvironmentVariables(path), Base: "", Sub: ""},
			Filename: file,
		}
		cfg.Libs = append(cfg.Libs, lib)
	}
}

// Used by IncludeFixer
func (proj *DevProject) CollectIncludeDirs() []PinnedPath {
	includes := make([]PinnedPath, 0)
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

func (proj *DevProject) CollectSourceDirs() []PinnedGlobPath {
	sourceDirs := make([]PinnedGlobPath, 0)
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
			ppg := PinnedGlobPath{Path: PinnedPath{Root: root, Base: base, Sub: sub}, Glob: dir.Glob}
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

func (prj *DevProject) AddSharedSource(name string) {
	prj.SourceDirs = append(prj.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: prj.Package.WorkspacePath(), Base: prj.Package.RepoName, Sub: "source/" + name + "/cpp"}, Glob: "**/*.c"})
	prj.SourceDirs = append(prj.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: prj.Package.WorkspacePath(), Base: prj.Package.RepoName, Sub: "source/" + name + "/cpp"}, Glob: "**/*.cpp"})
	for _, cfg := range prj.Configs {
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: prj.Package.WorkspacePath(), Base: prj.Package.RepoName, Sub: "source/" + name + "/include"})
	}
}

func (p *DevProject) EncodeJson(encoder *corepkg.JsonEncoder, key string) {
	encoder.BeginObject(key)
	{
		encoder.WriteField("name", p.Name)
		encoder.WriteField("dir_name", p.DirName)
		encoder.WriteField("build_type", p.BuildType.String())
		encoder.WriteField("build_targets", p.BuildTargets.String())
		encoder.WriteField("", p.EnvVars)
		if len(p.EnvVars) > 0 {
			encoder.BeginMap("env_vars")
			{
				for k, v := range p.EnvVars {
					encoder.WriteMapElement(k, v)
				}
			}
			encoder.EndMap()
		}
		if len(p.SourceDirs) > 0 {
			encoder.BeginArray("source_dirs")
			for _, dir := range p.SourceDirs {
				dir.EncodeJson(encoder, "")
			}
			encoder.EndArray()
		}
		if len(p.Configs) > 0 {
			encoder.BeginArray("configs")
			for _, cfg := range p.Configs {
				cfg.EncodeJson(encoder, "")
			}
			encoder.EndArray()
		}
		if !p.Dependencies.IsEmpty() {
			encoder.BeginArray("dependencies")
			for _, dep := range p.Dependencies.Values {
				encoder.WriteArrayElement(dep.Name)
			}
			encoder.EndArray()
		}
	}
	encoder.EndObject()
}

func DecodeJsonDevProject(decoder *corepkg.JsonDecoder, pkg *Package) *DevProject {
	project := &DevProject{
		Package:      pkg,
		EnvVars:      make(map[string]string),
		Configs:      make([]*DevConfig, 0),
		Dependencies: NewDevProjectList(),
		SourceDirs:   make([]PinnedGlobPath, 0),
	}

	fields := map[string]corepkg.JsonDecode{
		"name":       func(decoder *corepkg.JsonDecoder) { project.Name = decoder.DecodeString() },
		"dir_name":   func(decoder *corepkg.JsonDecoder) { project.DirName = decoder.DecodeString() },
		"build_type": func(decoder *corepkg.JsonDecoder) { project.BuildType = BuildTypeFromString(decoder.DecodeString()) },
		"build_targets": func(decoder *corepkg.JsonDecoder) {
			project.BuildTargets = BuildTargetFromString(decoder.DecodeString())
		},
		"env_vars": func(decoder *corepkg.JsonDecoder) {
			project.EnvVars = decoder.DecodeStringMapString()
		},
		"source_dirs": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				dir := DecodeJsonPinnedGlobPath(decoder)
				project.SourceDirs = append(project.SourceDirs, dir)
			})
		},
		"configs": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				cfg := DecodeJsonDevConfig(decoder)
				project.Configs = append(project.Configs, cfg)
			})
		},
		"dependencies": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeStringArray()
		},
	}
	decoder.Decode(fields)

	return project
}

// SetupDefaultCppLibProject returns a default C++ library project, since such a project can be used by
// an application as well as an unittest we need to add the appropriate configurations.
// Example:
//
//	SetupDefaultCppLibProject("cbase", "github.com/jurgen-kluft")
func SetupDefaultCppLibProject(pkg *Package, name string, dir string, buildTarget BuildTarget) *DevProject {
	project := NewProject(pkg, name, dir)
	project.BuildType = BuildTypeStaticLibrary
	project.Dependencies = NewDevProjectList()

	project.SourceDirs = append(project.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/" + dir + "/cpp"}, Glob: "**/*.c"})
	project.SourceDirs = append(project.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/" + dir + "/cpp"}, Glob: "**/*.cpp"})

	return project
}

func SetupCppLibProject(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppLibProject(pkg, "library_"+name, "main", GetBuildTarget())

	// TODO we should create all possible configuration, not just debug-dev/release-dev
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugDevConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseDevConfig()))

	project.BuildTargets = BuildTargetsAll

	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/main/include"})
	}
	return project
}

func SetupCppTestLibProject(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppLibProject(pkg, "unittest_library_"+name, "main", GetBuildTarget())

	// TODO we should create all possible configuration, not just debug-dev/release-dev
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugDevTestConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseDevTestConfig()))

	project.BuildTargets = BuildTargetsDesktop

	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/main/include"})
	}
	return project
}

func SetupCppLibProjectForDesktop(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppLibProject(pkg, "library_"+name, "main", GetBuildTarget())

	// TODO we should create all possible configuration, not just debug-dev/release-dev
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugDevConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseDevConfig()))

	project.BuildTargets = BuildTargetsDesktop

	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/main/include"})
	}
	return project
}

// Arduino Esp32
func SetupCppLibProjectForArduinoEsp32(pkg *Package, name string) *DevProject {
	project := SetupDefaultCppLibProject(pkg, "library_"+name, "main", BuildTargetArduinoEsp32)

	// TODO we should create all possible configuration, not just debug-dev/release-dev
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugDevConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseDevConfig()))

	project.BuildTargets = BuildTargetArduinoEsp32

	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/main/include"})
	}
	return project
}

// Arduino Esp8266
func SetupCppLibProjectForArduinoEsp8266(pkg *Package, name string) *DevProject {
	project := SetupDefaultCppLibProject(pkg, "library_"+name, "main", BuildTargetArduinoEsp8266)

	// TODO we should create all possible configuration, not just debug-dev/release-dev
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugDevConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseDevConfig()))

	project.BuildTargets = BuildTargetArduinoEsp8266
	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/main/include"})
	}
	return project
}

func SetupDefaultCppTestProject(pkg *Package, name string, buildTarget BuildTarget) *DevProject {
	project := NewProject(pkg, name, "test")
	project.BuildType = BuildTypeUnittest
	project.BuildTargets = BuildTargetsDesktop

	// TODO we should create all possible configuration, not just debug-dev/release-dev
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeUnittest, NewDebugDevTestConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeUnittest, NewReleaseDevTestConfig()))

	project.Dependencies = NewDevProjectList()

	project.SourceDirs = append(project.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/test/cpp"}, Glob: "**/*.c"})
	project.SourceDirs = append(project.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/test/cpp"}, Glob: "**/*.cpp"})

	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/test/include"})
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/main/include"})
	}

	return project
}

func SetupCppTestProject(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppTestProject(pkg, "unittest_"+name, GetBuildTarget())
	project.BuildTargets = BuildTargetsDesktop
	return project
}

// SetupDefaultCppCliProject returns a default C++ command-line program project
// Example:
//
//	SetupDefaultCppCliProject("cmycli", "github.com\\jurgen-kluft")
func SetupDefaultCppCliProject(pkg *Package, name string, buildTarget BuildTarget) *DevProject {
	project := NewProject(pkg, name, "cli")
	project.BuildType = BuildTypeCli
	project.BuildTargets = BuildTargetsDesktop
	// TODO we should create all possible configuration, not just debug-dev/release-dev
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeCli, NewDebugDevConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeCli, NewReleaseDevConfig()))
	project.Dependencies = NewDevProjectList()

	project.SourceDirs = append(project.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/cli/cpp"}, Glob: "**/*.c"})
	project.SourceDirs = append(project.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/cli/cpp"}, Glob: "**/*.cpp"})

	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/cli/include"})
	}

	return project
}

// SetupDefaultCppAppProject returns a default C++ application project
// Example:
//
//	SetupDefaultCppAppProject("cmyapp", "github.com\\jurgen-kluft")
func SetupDefaultCppAppProject(pkg *Package, name string, dirname string, buildTarget BuildTarget) *DevProject {
	project := NewProject(pkg, name, dirname)
	project.BuildType = BuildTypeApplication
	project.BuildTargets = BuildTargetsDesktop
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeApplication, NewDebugDevConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeApplication, NewReleaseDevConfig()))
	project.Dependencies = NewDevProjectList()

	project.SourceDirs = append(project.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/" + dirname + "/cpp"}, Glob: "**/*.c"})
	project.SourceDirs = append(project.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/" + dirname + "/cpp"}, Glob: "**/*.cpp"})

	for _, cfg := range project.Configs {
		configureProjectCompilerDefines(cfg)
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/" + dirname + "/include"})
	}

	return project
}

func SetupCppAppProject(pkg *Package, name string, dirname string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppAppProject(pkg, "app_"+name, dirname, GetBuildTarget())
	project.BuildTargets = BuildTargetsAll
	return project
}

func SetupCppAppProjectForDesktop(pkg *Package, name string, dirname string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppAppProject(pkg, "app_"+name, dirname, GetBuildTarget())
	project.BuildTargets = BuildTargetsDesktop
	return project
}

func SetupCppAppProjectForArduino(pkg *Package, name string, dirname string) *DevProject {
	// Arduino project
	project := SetupDefaultCppAppProject(pkg, "app_"+name, dirname, BuildTargetArduinoEsp32)
	project.SourceDirs = append(project.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: pkg.WorkspacePath(), Base: pkg.RepoName, Sub: "source/" + dirname + "/partitions"}, Glob: "**/*.csv"})
	project.BuildTargets = BuildTargetsArduino
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

func (p *DevProjectList) AddMany(project ...*DevProject) {
	for _, prj := range project {
		if prj != nil {
			p.Add(prj)
		}
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
		if corepkg.PathMatchWildcard(p.Name, name, true) {
			list.Add(p)
		}
	}
}

func (p *DevProjectList) TopoSort() error {
	var edges []corepkg.Edge

	// Sort the projects by dependencies
	for i, project := range p.Values {
		if project.Dependencies.IsEmpty() {
			edges = append(edges, corepkg.Edge{S: corepkg.Vertex(i), D: corepkg.InvalidVertex})
		} else {
			for _, dep := range project.Dependencies.Values {
				edges = append(edges, corepkg.Edge{S: corepkg.Vertex(i), D: corepkg.Vertex(p.Dict[dep.Name])})
			}
		}
	}

	sorted, err := corepkg.Toposort(edges)
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
