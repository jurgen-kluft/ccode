package denv

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

// DevProject is a structure that holds all the information that defines a project in an IDE
type DevProject struct {
	Name         string
	RepoName     string
	PackagePath  string
	BuildType    BuildType
	BuildTargets BuildTarget
	Copy2Output  map[string]string
	SourceDirs   []PinnedGlobPath
	Configs      []*DevConfig
	Dependencies *DevProjectList
}

func NewProject(pkg *Package, name string) *DevProject {
	return &DevProject{
		Name:         name,
		RepoName:     pkg.RepoName,
		PackagePath:  pkg.Path(),
		BuildType:    BuildTypeUnknown,
		BuildTargets: EmptyBuildTarget,
		Copy2Output:  make(map[string]string),
		Configs:      []*DevConfig{},
		Dependencies: NewDevProjectList(),
	}
}

func (p *DevProject) RootPath() string {
	return p.PackagePath
}

func (p *DevProject) ProjectPath() string {
	return filepath.Join(p.PackagePath, p.RepoName)
}

func (p *DevProject) SetBuildTargets(target ...BuildTarget) {
	p.BuildTargets = BuildTarget{}
	for _, t := range target {
		p.BuildTargets = p.BuildTargets.Union(t)
	}
}

// Copy copies the contents of the file at srcpath to a regular file
// at dstpath. If the file named by dstpath already exists, it is
// truncated. The function does not copy the file mode, file
// permission bits, or file attributes.
func FileCopy(srcpath, dstpath string) (err error) {
	r, err := os.Open(srcpath)
	if err != nil {
		return err
	}
	defer r.Close() // ignore error: file was opened read-only.

	w, err := os.Create(dstpath)
	if err != nil {
		return err
	}

	defer func() {
		// Report the error, if any, from Close, but do so
		// only if there isn't already an outgoing error.
		if c := w.Close(); err == nil {
			err = c
		}
	}()

	_, err = io.Copy(w, r)
	return err
}

func (prj *DevProject) DoCopyToOutput(outputPath string) {
	for src, dest := range prj.Copy2Output {
		srcPath := filepath.Join(prj.RootPath(), prj.RepoName, src)
		destPath := filepath.Join(outputPath, dest)

		// Create destination directory if it does not exist
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			os.MkdirAll(destPath, os.ModePerm)
		}

		// Copy files from srcPath to destPath, preserving directory structure
		// Use filepath.Walk to traverse the source directory
		filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Determine the relative path from srcPath
			relPath, err := filepath.Rel(srcPath, path)
			if err != nil {
				return err
			}

			destFilePath := filepath.Join(destPath, relPath)

			if info.IsDir() {
				// Create the directory in the destination
				os.MkdirAll(destFilePath, os.ModePerm)
			} else {
				// Copy the file
				FileCopy(path, destFilePath)
			}
			return nil
		})
	}
}

func (prj *DevProject) HasMatchingConfigForTarget(config BuildConfig, target BuildTarget) bool {
	if prj.BuildTargets.HasOverlap(target) {
		for _, cfg := range prj.Configs {
			if cfg.BuildConfig.Contains(config) {
				return true
			}
		}
	}
	return false
}

func (prj *DevProject) ResolveEnvironmentVariables(str string) string {
	// Replace all environment variables in the string
	// Variables can be nested, so we need to know when a replace was
	// done, and repeat the replace until no more replacements are done.
	vars := make(map[string]string)
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			vars[strings.ToLower(parts[0])] = parts[1]
		}
	}

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

		if value, ok := vars[strings.ToLower(str[start+1:end])]; ok {
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

func (prj *DevProject) CopyToOutput(srcdir string, destdir string) {
	srcdir = prj.ResolveEnvironmentVariables(srcdir)
	destdir = prj.ResolveEnvironmentVariables(destdir)
	prj.Copy2Output[srcdir] = destdir
}

func (prj *DevProject) AddDependencies(deps map[string]*DevProject) {
	for _, dep := range deps {
		prj.AddDependency(dep)
	}
}

func (prj *DevProject) AddDefine(defines ...string) {
	for _, define := range defines {
		define = prj.ResolveEnvironmentVariables(define)
		for _, cfg := range prj.Configs {
			cfg.Defines.Add(define)
		}
	}
}

func (prj *DevProject) AddLib(libs ...string) {
	for _, lib := range libs {
		for _, cfg := range prj.Configs {
			cfg.Libs = append(cfg.Libs, lib)
		}
	}
}

func (prj *DevProject) AddLibPath(root string, base string, sub string) {
	for _, cfg := range prj.Configs {
		cfg.LibPaths = append(cfg.LibPaths, PinnedPath{Root: root, Base: base, Sub: sub})
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

func (proj *DevProject) CollectProjectDependencies(deps *DevProjectList) {
	// Traverse and collect all dependencies
	for _, dp := range proj.Dependencies.Values {
		deps.Add(dp)
	}

	i := 0
	for i < deps.Len() {
		cp := deps.Values[i]
		for _, dp := range cp.Dependencies.Values {
			deps.Add(dp)
		}
		i++
	}
}

func (prj *DevProject) AddSourceFiles(dir string, extensions ...string) {
	for _, ext := range extensions {
		prj.SourceDirs = append(prj.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: prj.RootPath(), Base: prj.RepoName, Sub: dir}, Glob: "**/*" + ext})
	}
}

func (prj *DevProject) AddSourceFilesFrom(root, base, sub string, extensions ...string) {
	root = prj.ResolveEnvironmentVariables(root)
	base = prj.ResolveEnvironmentVariables(base)
	sub = prj.ResolveEnvironmentVariables(sub)
	for _, ext := range extensions {
		prj.SourceDirs = append(prj.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: root, Base: base, Sub: sub}, Glob: "**/*" + ext})
	}
}

func (prj *DevProject) AddDefaultInclude(dir string) {
	for _, cfg := range prj.Configs {
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: prj.RootPath(), Base: prj.RepoName, Sub: dir})
	}
}

func (prj *DevProject) AddInclude(root string, base string, sub string, extensions ...string) {
	// Environment variable resolve on root
	root = prj.ResolveEnvironmentVariables(root)
	for _, cfg := range prj.Configs {
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: root, Base: base, Sub: sub})
	}
}

func AddDefaultIncludePaths(pkg *Package, prj *DevProject, dir string) {
	for _, cfg := range prj.Configs {
		cfg.IncludeDirs = append(cfg.IncludeDirs, PinnedPath{Root: pkg.Path(), Base: pkg.RepoName, Sub: "source/" + dir + "/include"})
	}
}

func AddDefaultSourcePaths(pkg *Package, prj *DevProject, dir string, extensions ...string) {
	for _, ext := range extensions {
		prj.SourceDirs = append(prj.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: pkg.Path(), Base: pkg.RepoName, Sub: "source/" + dir + "/cpp"}, Glob: "**/*" + ext})
	}
}

func SetupDefaultCppLibProject(pkg *Package, name string) *DevProject {
	project := NewProject(pkg, name)
	project.BuildType = BuildTypeStaticLibrary
	return project
}

func SetupCppHeaderProject(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	name = "library_" + name
	dir := "main"

	project := NewProject(pkg, name)
	project.BuildType = BuildTypeHeaderOnly
	project.BuildTargets = BuildTargetsAll

	// TODO we should create all possible configuration, not just debug/release
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseConfig()))

	AddDefaultIncludePaths(pkg, project, dir)

	return project
}

func SetupCppLibProject(pkg *Package, name string) *DevProject {
	dir := "main"

	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppLibProject(pkg, "library_"+name)
	project.BuildTargets = BuildTargetsAll

	// TODO we should create all possible configuration, not just debug/release
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseConfig()))

	AddDefaultIncludePaths(pkg, project, dir)
	AddDefaultSourcePaths(pkg, project, dir, ".c", ".cpp")

	return project
}

func SetupCppLibraryDefault(pkg *Package, dir string, name string) *DevProject {

	project := SetupDefaultCppLibProject(pkg, "library_"+name)
	project.BuildTargets = BuildTargetsAll

	// TODO we should create all possible configuration, not just debug/release
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseConfig()))

	AddDefaultIncludePaths(pkg, project, dir)
	AddDefaultSourcePaths(pkg, project, dir, ".c", ".cpp")

	return project
}

func SetupCppLibraryCustom(pkg *Package, name string) *DevProject {

	project := SetupDefaultCppLibProject(pkg, "library_"+name)
	project.BuildTargets = BuildTargetsAll

	// TODO we should create all possible configuration, not just debug/release
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseConfig()))

	return project
}

func SetupCppTestLibProject(pkg *Package, name string) *DevProject {
	dir := "main"

	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppLibProject(pkg, "unittest_library_"+name)
	project.BuildType = BuildTypeStaticLibrary
	project.BuildTargets = BuildTargetsDesktop

	AddDefaultSourcePaths(pkg, project, dir, ".c", ".cpp")

	// TODO we should create all possible configuration, not just debug/release
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugTestConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseTestConfig()))

	AddDefaultIncludePaths(pkg, project, dir)

	return project
}

func SetupCppLibProjectForDesktop(pkg *Package, name string) *DevProject {
	dir := "main"

	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppLibProject(pkg, "library_"+name)
	project.BuildTargets = BuildTargetsDesktop

	// project.SourceDirs = append(project.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: pkg.Path(), Base: pkg.RepoName, Sub: "source/" + dir + "/cpp"}, Glob: "**/*.c"})
	// project.SourceDirs = append(project.SourceDirs, PinnedGlobPath{Path: PinnedPath{Root: pkg.Path(), Base: pkg.RepoName, Sub: "source/" + dir + "/cpp"}, Glob: "**/*.cpp"})
	AddDefaultSourcePaths(pkg, project, dir, ".c", ".cpp")

	// TODO we should create all possible configuration, not just debug/release
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseConfig()))

	AddDefaultIncludePaths(pkg, project, dir)

	return project
}

// Arduino Esp32
func SetupCppLibProjectForArduinoEsp32(pkg *Package, name string) *DevProject {
	project := SetupDefaultCppLibProject(pkg, "library_"+name)
	project.BuildTargets = BuildTargetArduinoEsp32

	// TODO we should create all possible configuration, not just debug/release
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseConfig()))

	return project
}

// Arduino
func SetupCppLibProjectForArduino(pkg *Package, name string) *DevProject {
	project := SetupDefaultCppLibProject(pkg, "library_"+name)
	project.BuildTargets = BuildTargetsArduino

	// TODO we should create all possible configuration, not just debug/release
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseConfig()))

	return project
}

func SetupDefaultCppLibProjectForArduino(pkg *Package, name string) *DevProject {
	project := SetupCppLibProjectForArduino(pkg, name)

	dir := "main"
	AddDefaultIncludePaths(pkg, project, dir)
	AddDefaultSourcePaths(pkg, project, dir, ".c", ".cpp")

	return project
}

func SetupCppLibraryForArduinoEsp32(pkg *Package, dir string, name string) *DevProject {

	project := SetupDefaultCppLibProject(pkg, "library_"+name)
	project.BuildTargets = BuildTargetArduinoEsp32

	// TODO we should create all possible configuration, not just debug/release
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseConfig()))

	AddDefaultIncludePaths(pkg, project, dir)
	AddDefaultSourcePaths(pkg, project, dir, ".c", ".cpp")

	return project
}

// Arduino Esp8266
func SetupCppLibProjectForArduinoEsp8266(pkg *Package, name string) *DevProject {
	dir := "main"

	project := SetupDefaultCppLibProject(pkg, "library_"+name)
	project.BuildTargets = BuildTargetArduinoEsp8266

	// TODO we should create all possible configuration, not just debug/release
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewDebugConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeStaticLibrary, NewReleaseConfig()))

	AddDefaultIncludePaths(pkg, project, dir)

	return project
}

func SetupDefaultCppTestProject(pkg *Package, name string, buildTarget BuildTarget) *DevProject {
	dir := "test"

	project := NewProject(pkg, name)
	project.BuildType = BuildTypeUnittest
	project.BuildTargets = BuildTargetsDesktop

	// TODO we should create all possible configuration, not just debug/release
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeUnittest, NewDebugTestConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeUnittest, NewReleaseTestConfig()))

	AddDefaultSourcePaths(pkg, project, dir, ".c", ".cpp")
	AddDefaultIncludePaths(pkg, project, dir)

	return project
}

func SetupCppTestProject(pkg *Package, name string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppTestProject(pkg, "unittest_"+name, BuildTargetsDesktop)
	return project
}

// SetupDefaultCppCliProject returns a default C++ command-line program project
// Example:
//
//	SetupDefaultCppCliProject("cmycli", "github.com\\jurgen-kluft")
func SetupDefaultCppCliProject(pkg *Package, name string, buildTarget BuildTarget) *DevProject {
	dir := "cli"

	project := NewProject(pkg, name)
	project.BuildType = BuildTypeCli
	project.BuildTargets = BuildTargetsDesktop
	// TODO we should create all possible configuration, not just debug/release
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeCli, NewDebugConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeCli, NewReleaseConfig()))

	AddDefaultSourcePaths(pkg, project, dir, ".c", ".cpp")
	AddDefaultIncludePaths(pkg, project, dir)

	return project
}

// SetupDefaultCppAppProject returns a default C++ application project
// Example:
//
//	SetupDefaultCppAppProject("cmyapp", "github.com\\jurgen-kluft")
func SetupDefaultCppAppProject(pkg *Package, name string, dirname string, buildTarget BuildTarget) *DevProject {
	project := NewProject(pkg, name)

	project.BuildType = BuildTypeApplication
	project.BuildTargets = buildTarget

	project.Configs = append(project.Configs, NewDevConfig(BuildTypeApplication, NewDebugConfig()))
	project.Configs = append(project.Configs, NewDevConfig(BuildTypeApplication, NewReleaseConfig()))

	AddDefaultSourcePaths(pkg, project, dirname, ".c", ".cpp")
	AddDefaultIncludePaths(pkg, project, dirname)

	return project
}

func SetupCppAppProject(pkg *Package, name string, dirname string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppAppProject(pkg, "app_"+name, dirname, BuildTargetsAll)
	return project
}

func SetupCppAppProjectForDesktop(pkg *Package, name string, dirname string) *DevProject {
	// Windows, Mac and Linux, build for the Host platform
	project := SetupDefaultCppAppProject(pkg, "app_"+name, dirname, BuildTargetsDesktop)
	return project
}

func SetupCppAppProjectForArduino(pkg *Package, name string, dirname string) *DevProject {
	project := SetupDefaultCppAppProject(pkg, "app_"+name, dirname, BuildTargetsArduino)
	return project
}

func SetupCppAppProjectForArduinoEsp32(pkg *Package, name string, dirname string) *DevProject {
	// Arduino project only for ESP32
	project := SetupDefaultCppAppProject(pkg, "app_"+name, dirname, BuildTargetArduinoEsp32)
	project.BuildTargets = BuildTarget{BuildTargetOsArduino: BuildTargetArchEsp32}
	return project
}

func SetupCppAppProjectForArduinoEsp8266(pkg *Package, name string, dirname string) *DevProject {
	// Arduino project only for ESP8266
	project := SetupDefaultCppAppProject(pkg, "app_"+name, dirname, BuildTargetArduinoEsp8266)
	project.BuildTargets = BuildTarget{BuildTargetOsArduino: BuildTargetArchEsp8266}
	return project
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

func (p *DevProjectList) Reset() {
	p.Dict = map[string]int{}
	p.Values = p.Values[:0]
	p.Keys = p.Keys[:0]
}

func (p *DevProjectList) Pop() *DevProject {
	if len(p.Values) == 0 {
		return nil
	}
	prj := p.Values[len(p.Values)-1]
	p.Values = p.Values[:len(p.Values)-1]
	p.Keys = p.Keys[:len(p.Keys)-1]
	delete(p.Dict, prj.Name)
	return prj
}

func (p *DevProjectList) Add(project *DevProject) {
	if _, ok := p.Dict[project.Name]; !ok {
		p.Dict[project.Name] = len(p.Values)
		p.Values = append(p.Values, project)
		p.Keys = append(p.Keys, project.Name)
	}
}

func (p *DevProjectList) AddMany(project map[string]*DevProject) {
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
