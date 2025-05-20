package denv

import (
	"fmt"
	"path/filepath"
	"strings"

	cutils "github.com/jurgen-kluft/ccode/cutils"
)

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type ProjectConfig struct {
	Group              string
	IsGuiApp           bool
	PchHeader          string
	MultiThreadedBuild Boolean
	CppAsObjCpp        Boolean
	Xcode              struct {
		BundleIdentifier string
	}
}

func NewProjectConfig() *ProjectConfig {
	config := &ProjectConfig{}
	return config
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type XcodeProjectConfig struct {
	XcodeProj                 *FileEntry
	PbxProj                   string
	InfoPlistFile             string
	Uuid                      cutils.UUID
	TargetUuid                cutils.UUID
	TargetProductUuid         cutils.UUID
	ConfigListUuid            cutils.UUID
	TargetConfigListUuid      cutils.UUID
	DependencyProxyUuid       cutils.UUID
	DependencyTargetUuid      cutils.UUID
	DependencyTargetProxyUuid cutils.UUID
}

type MsDevProjectConfig struct {
	UUID cutils.UUID
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
		if cutils.PathMatchWildcard(p.Name, name, true) {
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

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type Project struct {
	Workspace        *Workspace     // The workspace this project is part of
	Name             string         // The name of the project
	Version          string         // The version of the project
	Type             DevConfigType  // The type of the project
	SupportedTargets BuildTargets   // The targets that this project supports
	ProjectAbsPath   string         // The path where the project is located on disk, under the workspace directory
	GenerateAbsPath  string         // Where the project will be saved on disk
	Settings         *ProjectConfig //
	Group            *ProjectGroup  // Set when project is added into ProjectGroups
	SrcFiles         *FileEntryDict
	ExternalSrcFiles []*FileEntryDict
	ResourceEntries  *FileEntryDict
	VirtualFolders   *VirtualDirectories // For IDE generation, this is the path that is the root path of the virtual folder/file structure
	PchCpp           *FileEntry
	ProjectFilename  string
	ConfigsLocal     *ConfigList
	Dependencies     *ProjectList

	Resolved *ProjectResolved
}

func newProject2(prj *DevProject, ws *Workspace, name string, projectAbsPath string, settings *ProjectConfig) *Project {
	p := &Project{
		Workspace:        ws,
		Name:             name,
		Type:             prj.Type,
		SupportedTargets: prj.BuildTargets,
		ProjectAbsPath:   projectAbsPath,
		GenerateAbsPath:  ws.GenerateAbsPath,
		Settings:         settings,
		Group:            nil,
		SrcFiles:         NewFileEntryDict(projectAbsPath),
		ExternalSrcFiles: []*FileEntryDict{},
		ResourceEntries:  NewFileEntryDict(projectAbsPath),
		ConfigsLocal:     NewConfigList(),
		Dependencies:     NewProjectList(),
	}
	p.VirtualFolders = NewVirtualFolders(p.ProjectAbsPath) // The path that is the root path of the virtual folder/file structure

	p.Settings.MultiThreadedBuild = Boolean(ws.Config.MultiThreadedBuild)
	p.Settings.Xcode.BundleIdentifier = "$(PROJECT_NAME)"

	return p
}

func (p *Project) TypeIsCpp() bool {
	return p.Type == DevConfigTypeStaticLibrary || p.Type == DevConfigTypeDynamicLibrary || p.Type == DevConfigTypeExecutable
}
func (p *Project) TypeIsExe() bool {
	return p.Type == DevConfigTypeExecutable || p.Type == DevConfigTypeExecutable
}
func (p *Project) TypeIsDll() bool {
	return p.Type == DevConfigTypeDynamicLibrary || p.Type == DevConfigTypeDynamicLibrary
}
func (p *Project) TypeIsLib() bool {
	return p.Type == DevConfigTypeDynamicLibrary || p.Type == DevConfigTypeStaticLibrary
}
func (p *Project) TypeIsExeOrDll() bool {
	return p.TypeIsExe() || p.TypeIsDll()
}

func (p *Project) GetOrCreateConfig(t DevConfigType) *Config {
	c, ok := p.ConfigsLocal.Get(t)
	if !ok {
		c = NewConfig(t, p.Workspace, p)
	}
	return c
}

func (p *Project) FindConfig(t DevConfigType) *Config {
	c, ok := p.ConfigsLocal.Get(t)
	if !ok {
		return nil
	}
	return c
}

func (p *Project) FileEntriesGenerateUUIDs() {

	for _, i := range p.SrcFiles.Dict {
		f := p.SrcFiles.Values[i]
		f.UUID = cutils.GenerateUUID()
		f.BuildUUID = cutils.GenerateUUID()
	}

	for _, i := range p.ResourceEntries.Dict {
		f := p.SrcFiles.Values[i]
		f.UUID = cutils.GenerateUUID()
		f.BuildUUID = cutils.GenerateUUID()
	}

	for _, f := range p.VirtualFolders.Folders {
		f.UUID = cutils.GenerateUUID()
	}
}

func (p *Project) CreateConfiguration(cfg *DevConfig, configType DevConfigType) *Config {
	config := p.GetOrCreateConfig(configType)

	// C++ defines
	for _, define := range cfg.Defines.Values {
		config.CppDefines.ValuesToAdd(define)
	}

	// Library
	for _, lib := range cfg.Libs {
		config.AddLibrary(p.ProjectAbsPath, lib)
	}

	// Local (Package) Include directories
	for _, include := range cfg.LocalIncludeDirs {
		config.AddLocalIncludeDir(include)
	}

	// External (absolute path) Include directories
	for _, include := range cfg.ExternalIncludeDirs {
		config.AddExternalIncludeDir(include)
	}

	if configType.IsTest() {
		config.VisualStudioClCompile.AddOrSet("ExceptionHandling", "Sync")
	}

	return config
}

func (p *Project) AddConfigurations(configs []*DevConfig) {
	for _, cfg := range configs {
		configType := cfg.ConfigType
		config := p.CreateConfiguration(cfg, configType)
		p.ConfigsLocal.Add(config)
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

func (p *ProjectResolved) FindConfig(t DevConfigType) *Config {
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
		p.GenDataXcode.Uuid = cutils.GenerateUUID()
		p.GenDataXcode.TargetUuid = cutils.GenerateUUID()
		p.GenDataXcode.TargetProductUuid = cutils.GenerateUUID()
		p.GenDataXcode.ConfigListUuid = cutils.GenerateUUID()
		p.GenDataXcode.TargetConfigListUuid = cutils.GenerateUUID()
		p.GenDataXcode.DependencyProxyUuid = cutils.GenerateUUID()
		p.GenDataXcode.DependencyTargetUuid = cutils.GenerateUUID()
		p.GenDataXcode.DependencyTargetProxyUuid = cutils.GenerateUUID()

		for _, config := range p.Configs.Values {
			config.GenDataXcode.ProjectConfigUuid = cutils.GenerateUUID()
			config.GenDataXcode.TargetUuid = cutils.GenerateUUID()
			config.GenDataXcode.TargetConfigUuid = cutils.GenerateUUID()
		}
	}

	p.GenDataMsDev.UUID = cutils.GenerateUUID()
}

func (p *Project) Resolve(dev DevEnum) error {
	resolved := NewProjectResolved()

	if p.Type.IsExecutable() {
		resolved.HasOutputTarget = true
	} else if p.Type.IsDynamicLibrary() {
		resolved.HasOutputTarget = true
	} else if p.Type.IsStaticLibrary() {
		resolved.HasOutputTarget = true
	} else {
		return fmt.Errorf("project %q has unknown project type %q", p.Name, p.Type.ProjectString())
	}

	resolved.GeneratedFilesDir = filepath.Join(p.Workspace.GenerateAbsPath, "_generated_", p.Name)

	if p.Settings.PchHeader != "" {
		resolved.PchHeader = NewFileEntry()
		resolved.PchHeader.Init(p.Settings.PchHeader, false)
	}

	for _, f := range p.SrcFiles.Values {
		if f.Generated {
			p.VirtualFolders.AddFile(f)
		} else {
			p.VirtualFolders.AddFile(f)
		}
	}

	configsPerConfigTypeDb := map[DevConfigType][]*Config{}

	err := p.Dependencies.TopoSort()
	if err != nil {
		return err
	}

	for _, depProject := range p.Dependencies.Values {
		if depProject == p {
			return fmt.Errorf("project depends on itself, project='%s'", p.Name)
		}

		for _, config := range p.ConfigsLocal.Values {
			if dpConfig, ok := depProject.ConfigsLocal.Get(config.Type); ok {
				configsPerConfigTypeDb[config.Type] = append(configsPerConfigTypeDb[config.Type], dpConfig)
			}
		}
	}

	// For each config of this project, merge it will all the configs of the dependencies using the configsPerConfigTypeDb
	for _, config := range p.ConfigsLocal.Values {
		if configsOfSpecificConfigType, ok := configsPerConfigTypeDb[config.Type]; ok {
			mergedConfig := config.BuildResolved(configsOfSpecificConfigType)
			resolved.Configs.Add(mergedConfig)
		} else {
			mergedConfig := config.BuildResolved([]*Config{})
			resolved.Configs.Add(mergedConfig)
		}
	}

	// Should we copy these and then sort ?
	p.SrcFiles.SortByKey()
	p.VirtualFolders.SortByKey()
	p.FileEntriesGenerateUUIDs()

	// XCode ?
	if dev.IsXCode() {
		resolved.InitXCodeConfig(p)
	}
	resolved.GenerateUUIDs(dev)

	p.Resolved = resolved

	return nil
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

func (p *Project) GlobFiles(dir string, pattern string, isExcluded func(string) bool) {
	dir = cutils.PathNormalize(dir)
	pattern = cutils.PathNormalize(pattern)
	pp := strings.Split(pattern, "^")
	path := filepath.Join(dir, pp[0])
	files, err := GlobFiles(path, pp[1])
	if err != nil {
		return
	}

	for _, file := range files {
		if isExcluded(file) {
			continue
		}
		p.SrcFiles.Add(filepath.Join(pp[0], file))
	}
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

func (p *Project) BuildLibraryInformation(dev DevEnum, config *Config, workspaceGenerateAbsPath string) (linkDirs, linkFiles, linkLibs *ValueSet) {
	linkDirs = NewValueSet()
	linkFiles = NewValueSet()
	linkLibs = NewValueSet()

	// Library directories, these will be relative to the workspace generate path
	for _, dir := range config.LibraryDirs.Values {
		relpath := cutils.PathGetRelativeTo(dir.String(), workspaceGenerateAbsPath)
		linkDirs.Add(relpath)
	}

	// Library libs
	for _, file := range config.LibraryLibs.Values {
		linkLibs.Add(file)
	}

	// For all project dependencies, get their matching config and take the OutputLib and add it to the linkLibs
	if dev.IsVisualStudio() {
		for _, dep := range p.Dependencies.Values {
			if cfg, has := dep.Resolved.Configs.Get(config.Type); has {
				relpath := cutils.PathGetRelativeTo(cfg.Resolved.OutputLib.Path, workspaceGenerateAbsPath)
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

func (p *Project) BuildFrameworkInformation(config *Config) (frameworks *ValueSet) {
	frameworks = NewValueSet()

	// Library directories and files
	for _, fw := range config.LibraryFrameworks.Values {
		frameworks.Add(fw)
	}

	return
}
