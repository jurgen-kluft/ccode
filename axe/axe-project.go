package axe

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/denv"
	ccode_utils "github.com/jurgen-kluft/ccode/utils"
)

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type ProjectType int

const (
	ProjectTypeNone        ProjectType = iota
	ProjectTypeHeaders                 = 1
	ProjectTypeLib                     = 2
	ProjectTypeExe                     = 4
	ProjectTypeDll                     = 8
	ProjectTypeCLanguage               = 64
	ProjectTypeCppLanguage             = 128
	ProjectTypeCHeaders                = ProjectTypeHeaders | ProjectTypeCLanguage
	ProjectTypeCppHeaders              = ProjectTypeHeaders | ProjectTypeCppLanguage
	ProjectTypeCLib                    = ProjectTypeLib | ProjectTypeCLanguage
	ProjectTypeCppLib                  = ProjectTypeLib | ProjectTypeCppLanguage
	ProjectTypeCExe                    = ProjectTypeExe | ProjectTypeCLanguage
	ProjectTypeCppExe                  = ProjectTypeExe | ProjectTypeCppLanguage
	ProjectTypeCDll                    = ProjectTypeDll | ProjectTypeCLanguage
	ProjectTypeCppDll                  = ProjectTypeDll | ProjectTypeCppLanguage
)

func MakeProjectType(t denv.ProjectType) ProjectType {

	projectType := ProjectTypeNone
	if t.IsStaticLibrary() {
		projectType |= ProjectTypeLib
	} else if t.IsSharedLibrary() {
		projectType |= ProjectTypeDll
	} else if t.IsExecutable() {
		projectType |= ProjectTypeExe
	}

	if t.IsCLanguage() {
		projectType |= ProjectTypeCLanguage
	} else if t.IsCppLanguage() {
		projectType |= ProjectTypeCppLanguage
	}

	if t.IsHeaders() {
		projectType |= ProjectTypeHeaders
	}

	return projectType
}

func (t ProjectType) IsExecutable() bool {
	return t&ProjectTypeExe == ProjectTypeExe
}

func (t ProjectType) IsStaticLibrary() bool {
	return t&ProjectTypeLib == ProjectTypeLib
}

func (t ProjectType) IsSharedLibrary() bool {
	return t&ProjectTypeDll == ProjectTypeDll
}

func (t ProjectType) IsHeaders() bool {
	return t&ProjectTypeHeaders == ProjectTypeHeaders
}

func (t ProjectType) String() string {
	switch t {
	case ProjectTypeCHeaders:
		return "c_headers"
	case ProjectTypeCppHeaders:
		return "cpp_headers"
	case ProjectTypeCLib:
		return "c_lib"
	case ProjectTypeCExe:
		return "c_exe"
	case ProjectTypeCDll:
		return "c_dll"
	case ProjectTypeCppLib:
		return "cpp_lib"
	case ProjectTypeCppDll:
		return "cpp_dll"
	case ProjectTypeCppExe:
		return "cpp_exe"
	}

	return "error"
}

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
	Uuid                      ccode_utils.UUID
	TargetUuid                ccode_utils.UUID
	TargetProductUuid         ccode_utils.UUID
	ConfigListUuid            ccode_utils.UUID
	TargetConfigListUuid      ccode_utils.UUID
	DependencyProxyUuid       ccode_utils.UUID
	DependencyTargetUuid      ccode_utils.UUID
	DependencyTargetProxyUuid ccode_utils.UUID
}

type MsDevProjectConfig struct {
	UUID ccode_utils.UUID
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
		if ccode_utils.PathMatchWildcard(p.Name, name, true) {
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
	Workspace       *Workspace  // The workspace this project is part of
	Name            string      // The name of the project
	Type            ProjectType // The type of the project
	ProjectAbsPath  string      // The path where the project is located on disk, under the workspace directory
	GenerateAbsPath string      // Where the project will be saved on disk
	Settings        *ProjectConfig
	Group           *ProjectGroup // Set when project is added into ProjectGroups
	FileEntries     *FileEntryDict
	ResourceEntries *FileEntryDict
	VirtualFolders  *VirtualDirectories
	PchCpp          *FileEntry
	ProjectFilename string
	ConfigsLocal    *ConfigList
	Dependencies    *ProjectList

	Resolved *ProjectResolved
}

func newProject(ws *Workspace, name string, projectAbsPath string, projectType denv.ProjectType, settings *ProjectConfig) *Project {
	p := &Project{
		Workspace:       ws,
		Name:            name,
		Type:            MakeProjectType(projectType),
		ProjectAbsPath:  projectAbsPath,
		GenerateAbsPath: ws.GenerateAbsPath,
		Settings:        settings,
		Group:           nil,
		FileEntries:     NewFileEntryDict(ws, projectAbsPath),
		ResourceEntries: NewFileEntryDict(ws, projectAbsPath),
		ConfigsLocal:    NewConfigList(),
		Dependencies:    NewProjectList(),
	}
	p.VirtualFolders = NewVirtualFolders(p.ProjectAbsPath) // The path that is the root path of the virtual folder/file structure

	p.Settings.MultiThreadedBuild = Boolean(ws.Config.MultiThreadedBuild)
	p.Settings.Xcode.BundleIdentifier = "$(PROJECT_NAME)"

	return p
}

func (p *Project) TypeIsCpp() bool {
	return p.Type == ProjectTypeCppLib || p.Type == ProjectTypeCppDll || p.Type == ProjectTypeCppExe
}
func (p *Project) TypeIsC() bool {
	return p.Type == ProjectTypeCLib || p.Type == ProjectTypeCDll || p.Type == ProjectTypeCExe
}
func (p *Project) TypeIsExe() bool {
	return p.Type == ProjectTypeCExe || p.Type == ProjectTypeCppExe
}
func (p *Project) TypeIsDll() bool {
	return p.Type == ProjectTypeCDll || p.Type == ProjectTypeCppDll
}
func (p *Project) TypeIsLib() bool {
	return p.Type == ProjectTypeCLib || p.Type == ProjectTypeCppLib
}
func (p *Project) TypeIsHeaders() bool {
	return p.Type == ProjectTypeCHeaders || p.Type == ProjectTypeCppHeaders
}
func (p *Project) TypeIsExeOrDll() bool {
	return p.TypeIsExe() || p.TypeIsDll()
}

func (p *Project) GetOrCreateConfig(t ConfigType) *Config {
	c, ok := p.ConfigsLocal.Get(t)
	if !ok {
		c = NewConfig(t, p.Workspace, p)
	}
	return c
}

func (p *Project) FindConfig(t ConfigType) *Config {
	c, ok := p.ConfigsLocal.Get(t)
	if !ok {
		return nil
	}
	return c
}

func (p *Project) FileEntriesGenerateUUIDs() {

	for _, i := range p.FileEntries.Dict {
		f := p.FileEntries.Values[i]
		f.UUID = ccode_utils.GenerateUUID()
		f.BuildUUID = ccode_utils.GenerateUUID()
	}

	for _, i := range p.ResourceEntries.Dict {
		f := p.FileEntries.Values[i]
		f.UUID = ccode_utils.GenerateUUID()
		f.BuildUUID = ccode_utils.GenerateUUID()
	}

	for _, f := range p.VirtualFolders.Folders {
		f.UUID = ccode_utils.GenerateUUID()
	}
}

func (p *Project) CreateConfiguration(cfg *denv.Config, configType ConfigType) *Config {
	config := p.GetOrCreateConfig(configType)

	// C++ defines
	for _, define := range cfg.Defines.Values {
		config.CppDefines.ValuesToAdd(define)
	}

	// Library
	for _, lib := range cfg.Libs {
		config.Library.Add(p.ProjectAbsPath, lib)
	}

	// Include directories
	for _, include := range cfg.IncludeDirs {
		config.AddIncludeDir(include)
	}

	if configType.IsTest() {
		config.VisualStudioClCompile.AddOrSet("ExceptionHandling", "Sync")
	}

	return config
}

func (p *Project) AddConfigurations(configs []*denv.Config) {
	for _, cfg := range configs {
		configType := MakeFromDenvConfigType(cfg.ConfigType)
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

func (p *ProjectResolved) InitXCodeConfig(prj *Project) {
	gd := &XcodeProjectConfig{}
	gd.XcodeProj = NewFileEntry()
	gd.XcodeProj.Init(filepath.Join(prj.Workspace.GenerateAbsPath, prj.Name, ".xcodeproj"), true)
	gd.PbxProj = filepath.Join(gd.XcodeProj.Path, "project.pbxproj")
	p.GenDataXcode = gd
}

func (p *ProjectResolved) GenerateUUIDs(dev DevEnum) {
	if dev.IsXCode() {
		p.GenDataXcode.Uuid = ccode_utils.GenerateUUID()
		p.GenDataXcode.TargetUuid = ccode_utils.GenerateUUID()
		p.GenDataXcode.TargetProductUuid = ccode_utils.GenerateUUID()
		p.GenDataXcode.ConfigListUuid = ccode_utils.GenerateUUID()
		p.GenDataXcode.TargetConfigListUuid = ccode_utils.GenerateUUID()
		p.GenDataXcode.DependencyProxyUuid = ccode_utils.GenerateUUID()
		p.GenDataXcode.DependencyTargetUuid = ccode_utils.GenerateUUID()
		p.GenDataXcode.DependencyTargetProxyUuid = ccode_utils.GenerateUUID()

		for _, config := range p.Configs.Values {
			config.GenDataXcode.ProjectConfigUuid = ccode_utils.GenerateUUID()
			config.GenDataXcode.TargetUuid = ccode_utils.GenerateUUID()
			config.GenDataXcode.TargetConfigUuid = ccode_utils.GenerateUUID()
		}
	}

	p.GenDataMsDev.UUID = ccode_utils.GenerateUUID()
}

func (p *Project) Resolve(dev DevEnum) error {
	resolved := NewProjectResolved()

	if p.Type.IsExecutable() {
		resolved.HasOutputTarget = true
	} else if p.Type.IsSharedLibrary() {
		resolved.HasOutputTarget = true
	} else if p.Type.IsStaticLibrary() {
		resolved.HasOutputTarget = true
	} else if p.Type.IsHeaders() {
		// ...
	} else {
		return fmt.Errorf("project %q has unknown project type %q", p.Name, p.Type.String())
	}

	resolved.GeneratedFilesDir = filepath.Join(p.Workspace.GenerateAbsPath, "_generated_", p.Name)

	if p.Settings.PchHeader != "" {
		resolved.PchHeader = NewFileEntry()
		resolved.PchHeader.Init(p.Settings.PchHeader, false)
	}

	for _, f := range p.FileEntries.Values {
		if f.Generated {
			p.VirtualFolders.AddFile(f)
		} else {
			p.VirtualFolders.AddFile(f)
		}
	}

	configsPerConfigTypeDb := map[ConfigType][]*Config{}

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
	p.FileEntries.SortByKey()
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
	dir = ccode_utils.PathNormalize(dir)
	pattern = ccode_utils.PathNormalize(pattern)
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
		p.FileEntries.Add(filepath.Join(pp[0], file), false)
	}
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

func (p *Project) BuildLibraryInformation(dev DevEnum, config *Config, workspaceGenerateAbsPath string) (linkDirs, linkFiles, linkLibs *ValueSet) {
	linkDirs = NewValueSet()
	linkFiles = NewValueSet()
	linkLibs = NewValueSet()

	// Library directories, these will be relative to the workspace generate path
	for _, dir := range config.Library.Dirs.Values {
		relpath := ccode_utils.PathGetRelativeTo(dir.String(), workspaceGenerateAbsPath)
		linkDirs.Add(relpath)
	}

	// Library libs
	for _, file := range config.Library.Libs.Values {
		linkLibs.Add(file)
	}

	// For all project dependencies, get their matching config and take the OutputLib and add it to the linkLibs
	if dev.IsVisualStudio() {
		for _, dep := range p.Dependencies.Values {
			if cfg, has := dep.Resolved.Configs.Get(config.Type); has {
				relpath := ccode_utils.PathGetRelativeTo(cfg.Resolved.OutputLib.Path, workspaceGenerateAbsPath)
				linkLibs.Add(relpath)
			}
		}
	}

	// Library files
	for _, file := range config.Library.Files.Values {
		linkFiles.Add(file)
	}

	return
}

func (p *Project) BuildFrameworkInformation(config *Config) (frameworks *ValueSet) {
	frameworks = NewValueSet()

	// Library directories and files
	for _, fw := range config.Library.Frameworks.Values {
		frameworks.Add(fw)
	}

	return
}
