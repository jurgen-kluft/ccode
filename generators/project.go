package ide_generators

import (
	"fmt"
	"path/filepath"

	corepkg "github.com/jurgen-kluft/ccode/core"
	"github.com/jurgen-kluft/ccode/denv"
)

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type Project struct {
	Workspace       *Workspace       // The workspace this project is part of
	Name            string           // The name of the project
	BuildType       denv.BuildType   // The build type of the project
	BuildTargets    denv.BuildTarget // The targets that this project supports
	ProjectAbsPath  string           // The path where the project is located on disk, under the workspace directory
	GenerateAbsPath string           // Where the project will be saved on disk
	Group           *ProjectGroup    // Set when project is added into ProjectGroups
	SrcFileGroups   []*FileEntryDict
	ResFileGroups   []*FileEntryDict
	VirtualFolders  *VirtualDirectories // For IDE generation, this is the path that is the root path of the virtual folder/file structure
	PchCpp          *FileEntry
	Settings        *ProjectSettings
	ProjectFilename string
	ConfigsLocal    *ConfigList
	Dependencies    *ProjectList

	Resolved *ProjectResolved
}

func newProject2(ws *Workspace, buildTarget denv.BuildTarget, prj *denv.DevProject, generateAbsPath string, settings *ProjectSettings) *Project {
	projectAbsPath := prj.Path()
	p := &Project{
		Workspace:       ws,
		Name:            prj.Name,
		BuildType:       prj.BuildType,
		BuildTargets:    prj.BuildTargets,
		ProjectAbsPath:  projectAbsPath,
		GenerateAbsPath: generateAbsPath,
		Group:           nil,
		SrcFileGroups:   []*FileEntryDict{NewFileEntryDict(projectAbsPath)},
		ResFileGroups:   []*FileEntryDict{NewFileEntryDict(projectAbsPath)},
		Settings:        settings,
		ConfigsLocal:    NewConfigList(),
		Dependencies:    NewProjectList(),
	}
	p.VirtualFolders = NewVirtualFolders(p.ProjectAbsPath) // The path that is the root path of the virtual folder/file structure

	// TODO Should we copy the configurations here ?
	for _, devCfg := range prj.Configs {
		cfg := p.CreateConfiguration(buildTarget, devCfg)
		p.ConfigsLocal.Add(cfg)
	}

	//p.Settings.MultiThreadedBuild = ws.Config.MultiThreadedBuild
	//p.Settings.Xcode.BundleIdentifier = "$(PROJECT_NAME)"

	return p
}

func (p *Project) TypeIsExe() bool {
	return p.BuildType.IsExecutable()
}
func (p *Project) TypeIsDll() bool {
	return p.BuildType == denv.BuildTypeDynamicLibrary
}
func (p *Project) TypeIsLib() bool {
	return p.BuildType == denv.BuildTypeDynamicLibrary || p.BuildType == denv.BuildTypeStaticLibrary
}
func (p *Project) TypeIsExeOrDll() bool {
	return p.TypeIsExe() || p.TypeIsDll()
}

func (p *Project) GetOrCreateConfig(b denv.BuildTarget, t denv.BuildConfig) *Config {
	c, ok := p.ConfigsLocal.Get(t)
	if !ok {
		c = NewConfig(b, t, p)
	}
	return c
}

func (p *Project) FindConfig(t denv.BuildConfig) *Config {
	c, ok := p.ConfigsLocal.Get(t)
	if !ok {
		return nil
	}
	return c
}

func (p *Project) FileEntriesGenerateUUIDs() {
	for _, g := range p.SrcFileGroups {
		for _, i := range g.Dict {
			f := g.Values[i]
			f.UUID = corepkg.GenerateUUID()
			f.BuildUUID = corepkg.GenerateUUID()
		}
	}
	for _, g := range p.ResFileGroups {
		for _, i := range g.Dict {
			f := g.Values[i]
			f.UUID = corepkg.GenerateUUID()
			f.BuildUUID = corepkg.GenerateUUID()
		}
	}
	for _, f := range p.VirtualFolders.Folders {
		f.UUID = corepkg.GenerateUUID()
	}
}

func (p *Project) CreateConfiguration(buildTarget denv.BuildTarget, cfg *denv.DevConfig) *Config {
	config := p.GetOrCreateConfig(buildTarget, cfg.BuildConfig)

	// C++ defines
	for _, define := range cfg.Defines.Values {
		config.CppDefines.ValuesToAdd(define)
	}

	// Library
	for _, lib := range cfg.Libs {
		config.AddLibrary(p.ProjectAbsPath, lib)
	}

	// Include directories
	for _, include := range cfg.IncludeDirs {
		config.AddIncludeDir(include)
	}

	return config
}

func (p *Project) AddConfigurations(configs []*denv.DevConfig) {
	for _, cfg := range configs {
		if !p.ConfigsLocal.Has(cfg.BuildConfig) {
			config := p.CreateConfiguration(p.Workspace.BuildTarget, cfg)
			p.ConfigsLocal.Add(config)
		}
	}
}

// ProjectResolved contains resolved information, and can be used by a generator
type ProjectResolved struct {
	HasOutputTarget   bool
	GeneratedFilesDir string
	Configs           *ConfigList
	PchHeader         *FileEntry
	PchSuffix         string

	GenDataMake struct {
		Makefile string
	}
	GenDataXcode *XcodeProjectConfig
	GenDataMsDev *MsDevProjectConfig
}

func NewProjectResolved() *ProjectResolved {
	return &ProjectResolved{
		Configs:      NewConfigList(),
		GenDataXcode: NewXcodeProjectConfig(),
		GenDataMsDev: NewMsDevProjectConfig(),
	}
}

func (p *ProjectResolved) FindConfig(t denv.BuildConfig) *Config {
	c, ok := p.Configs.Get(t)
	if !ok {
		return nil
	}
	return c
}

func (p *ProjectResolved) InitXCodeConfig(prj *Project) {
	gd := &XcodeProjectConfig{}
	gd.XcodeProj = NewFileEntry()
	gd.XcodeProj.Init(filepath.Join(prj.Workspace.GenerateAbsPath, prj.Name, ".xcodeproj"), true)
	gd.PbxProj = filepath.Join(gd.XcodeProj.Path, "project.pbxproj")
	p.GenDataXcode = gd
}

func (p *ProjectResolved) GenerateUUIDs(dev DevEnum) {
	if dev.IsXCode() {
		p.GenDataXcode.Uuid = corepkg.GenerateUUID()
		p.GenDataXcode.TargetUuid = corepkg.GenerateUUID()
		p.GenDataXcode.TargetProductUuid = corepkg.GenerateUUID()
		p.GenDataXcode.ConfigListUuid = corepkg.GenerateUUID()
		p.GenDataXcode.TargetConfigListUuid = corepkg.GenerateUUID()
		p.GenDataXcode.DependencyProxyUuid = corepkg.GenerateUUID()
		p.GenDataXcode.DependencyTargetUuid = corepkg.GenerateUUID()
		p.GenDataXcode.DependencyTargetProxyUuid = corepkg.GenerateUUID()

		for _, config := range p.Configs.Values {
			config.GenDataXcode.ProjectConfigUuid = corepkg.GenerateUUID()
			config.GenDataXcode.TargetUuid = corepkg.GenerateUUID()
			config.GenDataXcode.TargetConfigUuid = corepkg.GenerateUUID()
		}
	}

	p.GenDataMsDev.UUID = corepkg.GenerateUUID()
}

func (p *Project) Resolve(devEnum DevEnum) error {
	resolved := NewProjectResolved()

	if p.BuildType.IsExecutable() {
		resolved.HasOutputTarget = true
	} else if p.BuildType.IsDynamicLibrary() {
		resolved.HasOutputTarget = true
	} else if p.BuildType.IsStaticLibrary() {
		resolved.HasOutputTarget = true
	} else {
		return fmt.Errorf("project %q has unknown project type %q", p.Name, p.BuildType.ProjectString())
	}

	resolved.GeneratedFilesDir = filepath.Join(p.Workspace.GenerateAbsPath, "_generated_", p.Name)

	// if p.Settings.PchHeader != "" {
	// 	resolved.PchHeader = NewFileEntry()
	// 	resolved.PchHeader.Init(p.Settings.PchHeader, false)
	// }

	for _, g := range p.SrcFileGroups {
		for _, f := range g.Values {
			p.VirtualFolders.AddFile(f)
		}
	}
	configsPerConfigTypeDb := map[denv.BuildConfig][]*Config{}

	err := p.Dependencies.TopoSort()
	if err != nil {
		return err
	}

	for _, depProject := range p.Dependencies.Values {
		if depProject == p {
			return fmt.Errorf("project depends on itself, project='%s'", p.Name)
		}

		for _, config := range p.ConfigsLocal.Values {
			if dpConfig, ok := depProject.ConfigsLocal.Get(config.BuildConfig); ok {
				configsPerConfigTypeDb[config.BuildConfig] = append(configsPerConfigTypeDb[config.BuildConfig], dpConfig)
			}
		}
	}

	// For each config of this project, merge it will all the configs of the dependencies using the configsPerConfigTypeDb
	for _, config := range p.ConfigsLocal.Values {
		if configsOfSpecificConfigType, ok := configsPerConfigTypeDb[config.BuildConfig]; ok {
			mergedConfig := config.BuildResolved(configsOfSpecificConfigType)
			resolved.Configs.Add(mergedConfig)
		} else {
			mergedConfig := config.BuildResolved([]*Config{})
			resolved.Configs.Add(mergedConfig)
		}
	}

	// Should we copy these and then sort ?
	for _, g := range p.SrcFileGroups {
		g.SortByKey()
	}
	p.VirtualFolders.SortByKey()
	p.FileEntriesGenerateUUIDs()

	// XCode ?
	if devEnum.IsXCode() {
		resolved.InitXCodeConfig(p)
	}
	resolved.GenerateUUIDs(devEnum)

	p.Resolved = resolved

	return nil
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

func (p *Project) GlobFiles(path string, sub string, pattern string, isExcluded func(string) bool) {
	path = corepkg.PathNormalize(path)
	sub = corepkg.PathNormalize(sub)
	pattern = corepkg.PathNormalize(pattern)

	dirFunc := func(rootPath, relPath string) bool {
		return true // We want to include all directories
	}

	dirpath := filepath.Join(path, sub)
	filepaths := []string{}
	fileFunc := func(rootPath, relPath string) {
		if !isExcluded(relPath) {
			if match := corepkg.GlobMatching(relPath, pattern); match {
				filepaths = append(filepaths, relPath)
			}
		}
	}

	err := corepkg.FileEnumerate(dirpath, dirFunc, fileFunc)
	if err != nil {
		corepkg.LogErrorf(err, "failed to enumerate files in %q: %v", dirpath)
		return
	}

	if len(filepaths) > 0 {
		fileGroup := NewFileEntryDict(path)
		for _, file := range filepaths {
			fileGroup.Add(filepath.Join(sub, file))
		}
		p.SrcFileGroups = append(p.SrcFileGroups, fileGroup)
	}
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

func (p *Project) BuildLibraryInformation(devEnum DevEnum, config *Config, workspaceGenerateAbsPath string) (linkDirs, linkFiles, linkLibs *corepkg.ValueSet) {
	linkDirs = corepkg.NewValueSet()
	linkFiles = corepkg.NewValueSet()
	linkLibs = corepkg.NewValueSet()

	// Library directories
	for _, dir := range config.LibraryPaths.Values {
		relpath := corepkg.PathGetRelativeTo(dir.String(), workspaceGenerateAbsPath)
		linkDirs.Add(relpath)
	}

	// Library files
	for _, file := range config.LibraryFiles.Values {
		linkLibs.Add(file)
	}

	// For all project dependencies, get their matching config and take the OutputLib and add it to the linkLibs
	if devEnum.IsVisualStudio() {
		for _, dep := range p.Dependencies.Values {
			if cfg, has := dep.Resolved.Configs.Get(config.BuildConfig); has {
				relpath := corepkg.PathGetRelativeTo(cfg.Resolved.OutputLib.Path, workspaceGenerateAbsPath)
				linkLibs.Add(relpath)
			}
		}
	}

	// Library files
	for _, file := range config.LibraryFiles.Values {
		linkFiles.Add(file)
	}

	return
}

func (p *Project) BuildFrameworkInformation(config *Config) (frameworks *corepkg.ValueSet) {
	frameworks = corepkg.NewValueSet()

	// Library directories and files
	for _, fw := range config.LibraryFrameworks.Values {
		frameworks.Add(fw)
	}

	return
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type ProjectSettings struct {
	Group              string
	IsGuiApp           bool
	PchHeader          string
	MultiThreadedBuild bool
	CppAsObjCpp        bool
	Xcode              struct {
		BundleIdentifier string
	}
}

func NewProjectSettings() *ProjectSettings {
	config := &ProjectSettings{}
	return config
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type XcodeProjectConfig struct {
	XcodeProj                 *FileEntry
	PbxProj                   string
	InfoPlistFile             string
	Uuid                      corepkg.UUID
	TargetUuid                corepkg.UUID
	TargetProductUuid         corepkg.UUID
	ConfigListUuid            corepkg.UUID
	TargetConfigListUuid      corepkg.UUID
	DependencyProxyUuid       corepkg.UUID
	DependencyTargetUuid      corepkg.UUID
	DependencyTargetProxyUuid corepkg.UUID
}

type MsDevProjectConfig struct {
	UUID corepkg.UUID
}

func NewXcodeProjectConfig() *XcodeProjectConfig {
	return &XcodeProjectConfig{
		XcodeProj: NewFileEntry(),
	}
}

func NewMsDevProjectConfig() *MsDevProjectConfig {
	return &MsDevProjectConfig{}
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type ProjectList struct {
	Dict   map[string]int
	Values []*Project
	Keys   []string
}

func NewProjectList() *ProjectList {
	return &ProjectList{
		Dict:   map[string]int{},
		Values: []*Project{},
		Keys:   []string{},
	}
}

func (p *ProjectList) Len() int {
	return len(p.Values)
}

func (p *ProjectList) IsEmpty() bool {
	return len(p.Values) == 0
}

func (p *ProjectList) Add(project *Project) bool {
	if _, ok := p.Dict[project.Name]; !ok {
		p.Dict[project.Name] = len(p.Values)
		p.Values = append(p.Values, project)
		p.Keys = append(p.Keys, project.Name)
		return true
	}
	return false
}

func (p *ProjectList) Get(name string) (*Project, bool) {
	if i, ok := p.Dict[name]; ok {
		return p.Values[i], true
	}
	return nil, false
}

func (p *ProjectList) CollectByWildcard(name string, list *ProjectList) {
	for _, p := range p.Values {
		if corepkg.PathMatchWildcard(p.Name, name, true) {
			list.Add(p)
		}
	}
}

func (p *ProjectList) TopoSort() error {
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

	var sortedProjects []*Project
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
