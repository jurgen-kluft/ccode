package denv

import (
	"fmt"
	"path/filepath"

	"github.com/jurgen-kluft/ccode/dev"
	"github.com/jurgen-kluft/ccode/foundation"
)

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type Project struct {
	Workspace        *Workspace       // The workspace this project is part of
	Name             string           // The name of the project
	Version          string           // The version of the project
	BuildType        dev.BuildType    // The build type of the project
	SupportedTargets dev.BuildTarget  // The targets that this project supports
	ProjectAbsPath   string           // The path where the project is located on disk, under the workspace directory
	GenerateAbsPath  string           // Where the project will be saved on disk
	Settings         *ProjectSettings //
	Group            *ProjectGroup    // Set when project is added into ProjectGroups
	SrcFileGroups    []*FileEntryDict
	ResFileGroups    []*FileEntryDict
	VirtualFolders   *VirtualDirectories // For IDE generation, this is the path that is the root path of the virtual folder/file structure
	PchCpp           *FileEntry
	ProjectFilename  string
	ConfigsLocal     *ConfigList
	Dependencies     *ProjectList

	Resolved *ProjectResolved
}

func newProject2(prj *DevProject, ws *Workspace, settings *ProjectSettings) *Project {
	projectAbsPath := prj.Package.PackagePath()
	p := &Project{
		Workspace:        ws,
		Name:             prj.Name,
		BuildType:        prj.BuildType,
		SupportedTargets: prj.Supported,
		ProjectAbsPath:   projectAbsPath,
		GenerateAbsPath:  ws.GenerateAbsPath,
		Settings:         settings,
		Group:            nil,
		SrcFileGroups:    []*FileEntryDict{NewFileEntryDict(projectAbsPath)},
		ResFileGroups:    []*FileEntryDict{NewFileEntryDict(projectAbsPath)},
		ConfigsLocal:     NewConfigList(),
		Dependencies:     NewProjectList(),
	}
	p.VirtualFolders = NewVirtualFolders(p.ProjectAbsPath) // The path that is the root path of the virtual folder/file structure

	// TODO Should we copy the configurations here ?
	for _, devCfg := range prj.Configs {
		cfg := p.CreateConfiguration(devCfg)
		p.ConfigsLocal.Add(cfg)
	}

	p.Settings.MultiThreadedBuild = ws.Config.MultiThreadedBuild
	p.Settings.Xcode.BundleIdentifier = "$(PROJECT_NAME)"

	return p
}

func (p *Project) TypeIsExe() bool {
	return p.BuildType.IsExecutable()
}
func (p *Project) TypeIsDll() bool {
	return p.BuildType == dev.BuildTypeDynamicLibrary
}
func (p *Project) TypeIsLib() bool {
	return p.BuildType == dev.BuildTypeDynamicLibrary || p.BuildType == dev.BuildTypeStaticLibrary
}
func (p *Project) TypeIsExeOrDll() bool {
	return p.TypeIsExe() || p.TypeIsDll()
}

func (p *Project) GetOrCreateConfig(t dev.BuildConfig) *Config {
	c, ok := p.ConfigsLocal.Get(t)
	if !ok {
		c = NewConfig(t, p.Workspace, p)
	}
	return c
}

func (p *Project) FindConfig(t dev.BuildConfig) *Config {
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
			f.UUID = foundation.GenerateUUID()
			f.BuildUUID = foundation.GenerateUUID()
		}
	}
	for _, g := range p.ResFileGroups {
		for _, i := range g.Dict {
			f := g.Values[i]
			f.UUID = foundation.GenerateUUID()
			f.BuildUUID = foundation.GenerateUUID()
		}
	}
	for _, f := range p.VirtualFolders.Folders {
		f.UUID = foundation.GenerateUUID()
	}
}

func (p *Project) CreateConfiguration(cfg *DevConfig) *Config {
	config := p.GetOrCreateConfig(cfg.BuildConfig)

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

	if cfg.BuildConfig.IsTest() {
		config.VisualStudioClCompile.AddOrSet("ExceptionHandling", "Sync")
	}

	return config
}

func (p *Project) AddConfigurations(configs []*DevConfig) {
	for _, cfg := range configs {
		if !p.ConfigsLocal.Has(cfg.BuildConfig) {
			config := p.CreateConfiguration(cfg)
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

func (p *ProjectResolved) FindConfig(t dev.BuildConfig) *Config {
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
		p.GenDataXcode.Uuid = foundation.GenerateUUID()
		p.GenDataXcode.TargetUuid = foundation.GenerateUUID()
		p.GenDataXcode.TargetProductUuid = foundation.GenerateUUID()
		p.GenDataXcode.ConfigListUuid = foundation.GenerateUUID()
		p.GenDataXcode.TargetConfigListUuid = foundation.GenerateUUID()
		p.GenDataXcode.DependencyProxyUuid = foundation.GenerateUUID()
		p.GenDataXcode.DependencyTargetUuid = foundation.GenerateUUID()
		p.GenDataXcode.DependencyTargetProxyUuid = foundation.GenerateUUID()

		for _, config := range p.Configs.Values {
			config.GenDataXcode.ProjectConfigUuid = foundation.GenerateUUID()
			config.GenDataXcode.TargetUuid = foundation.GenerateUUID()
			config.GenDataXcode.TargetConfigUuid = foundation.GenerateUUID()
		}
	}

	p.GenDataMsDev.UUID = foundation.GenerateUUID()
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

	if p.Settings.PchHeader != "" {
		resolved.PchHeader = NewFileEntry()
		resolved.PchHeader.Init(p.Settings.PchHeader, false)
	}

	for _, g := range p.SrcFileGroups {
		for _, f := range g.Values {
			p.VirtualFolders.AddFile(f)
		}
	}
	configsPerConfigTypeDb := map[dev.BuildConfig][]*Config{}

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
	path = foundation.PathNormalize(path)
	sub = foundation.PathNormalize(sub)
	pattern = foundation.PathNormalize(pattern)
	files, err := GlobFiles(filepath.Join(path, sub), pattern, isExcluded)
	if err != nil {
		return
	}

	if len(files) > 0 {
		fileGroup := NewFileEntryDict(path)
		for _, file := range files {
			fileGroup.Add(filepath.Join(sub, file))
		}
		p.SrcFileGroups = append(p.SrcFileGroups, fileGroup)
	}
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

func (p *Project) BuildLibraryInformation(devEnum DevEnum, config *Config, workspaceGenerateAbsPath string) (linkDirs, linkFiles, linkLibs *foundation.ValueSet) {
	linkDirs = foundation.NewValueSet()
	linkFiles = foundation.NewValueSet()
	linkLibs = foundation.NewValueSet()

	// Library directories
	for _, dir := range config.LibraryPaths.Values {
		relpath := foundation.PathGetRelativeTo(dir.String(), workspaceGenerateAbsPath)
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
				relpath := foundation.PathGetRelativeTo(cfg.Resolved.OutputLib.Path, workspaceGenerateAbsPath)
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

func (p *Project) BuildFrameworkInformation(config *Config) (frameworks *foundation.ValueSet) {
	frameworks = foundation.NewValueSet()

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
	Uuid                      foundation.UUID
	TargetUuid                foundation.UUID
	TargetProductUuid         foundation.UUID
	ConfigListUuid            foundation.UUID
	TargetConfigListUuid      foundation.UUID
	DependencyProxyUuid       foundation.UUID
	DependencyTargetUuid      foundation.UUID
	DependencyTargetProxyUuid foundation.UUID
}

type MsDevProjectConfig struct {
	UUID foundation.UUID
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

func (p *ProjectList) Add(project *Project) {
	if _, ok := p.Dict[project.Name]; !ok {
		p.Dict[project.Name] = len(p.Values)
		p.Values = append(p.Values, project)
		p.Keys = append(p.Keys, project.Name)
	}
}

func (p *ProjectList) Get(name string) (*Project, bool) {
	if i, ok := p.Dict[name]; ok {
		return p.Values[i], true
	}
	return nil, false
}

func (p *ProjectList) CollectByWildcard(name string, list *ProjectList) {
	for _, p := range p.Values {
		if foundation.PathMatchWildcard(p.Name, name, true) {
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
