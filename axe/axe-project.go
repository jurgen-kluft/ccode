package axe

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/glob"
)

type ProjectType int

const (
	ProjectTypeNone ProjectType = iota
	ProjectTypeCHeaders
	ProjectTypeCppHeaders
	ProjectTypeCLib
	ProjectTypeCExe
	ProjectTypeCDll
	ProjectTypeCppLib
	ProjectTypeCppDll
	ProjectTypeCppExe
)

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

type ExclusionFilter struct {
	Exclusions []string
}

func (f *ExclusionFilter) IsExcluded(filepath string) bool {
	parts := PathSplitRelativeFilePath(filepath, true)
	for i := 0; i < len(parts)-1; i++ {
		for _, exclusion := range f.Exclusions {
			if strings.HasSuffix(parts[i], exclusion) {
				return true
			}
		}
	}
	return false
}

func NewExclusionFilter(target MakeTarget) *ExclusionFilter {
	if target.OSIsMac() {
		return &ExclusionFilter{Exclusions: []string{"_win32", "_win64", "_pc", "_linux", "_nob"}}
	} else if target.OSIsWindows() {
		return &ExclusionFilter{Exclusions: []string{"_mac", "_macos", "_darwin", "_linux", "_unix", "_nob"}}
	} else if target.OSIsLinux() {
		return &ExclusionFilter{Exclusions: []string{"_win32", "_win64", "_pc", "_mac", "_macos", "_darwin", "_nob"}}
	}
	return &ExclusionFilter{Exclusions: []string{"_nob"}}
}

type ProjectDependencies struct {
	Dict   map[string]int
	Values []*Project
}

func NewProjectDependencies() *ProjectDependencies {
	return &ProjectDependencies{
		Dict:   map[string]int{},
		Values: []*Project{},
	}
}

func (pd *ProjectDependencies) AddUnique(p *Project) {
	if _, ok := pd.Dict[p.Name]; !ok {
		pd.Dict[p.Name] = len(pd.Values)
		pd.Values = append(pd.Values, p)
	}
}

type ProjectConfig struct {
	Group              string
	Type               ProjectType // ccore compiler define (CCORE_GEN_TYPE_{Type})
	IsGuiApp           bool
	PchHeader          string
	Dependencies       []string
	MultiThreadedBuild Boolean
	CppAsObjCpp        Boolean
	Xcode              struct {
		BundleIdentifier string
	}
	MsDev VisualStudioConfig
}

func NewProjectConfig() *ProjectConfig {
	return &ProjectConfig{}
}

func NewVisualStudioProjectConfig(version EnumVisualStudio) *ProjectConfig {
	config := &ProjectConfig{}
	config.Dependencies = make([]string, 0)
	config.MsDev = NewVisualStudioConfig(version)
	return config
}

type XcodeProjectConfig struct {
	XcodeProj                 *FileEntry
	PbxProj                   string
	InfoPlistFile             string
	Uuid                      UUID
	TargetUuid                UUID
	TargetProductUuid         UUID
	ConfigListUuid            UUID
	TargetConfigListUuid      UUID
	DependencyProxyUuid       UUID
	DependencyTargetUuid      UUID
	DependencyTargetProxyUuid UUID
}

type MsDevProjectConfig struct {
	VcxProj string
	UUID    UUID
}

func NewXcodeProjectConfig() *XcodeProjectConfig {
	return &XcodeProjectConfig{
		XcodeProj: NewFileEntry(),
	}
}

type Project struct {
	Workspace           *Workspace  // The workspace this project is part of
	Name                string      // The name of the project
	Type                ProjectType // The type of the project
	ProjectAbsPath      string      // The path where the project is located on disk, under the workspace directory
	GenerateAbsPath     string      // Where the project will be saved on disk
	Settings            *ProjectConfig
	Group               *ProjectGroup
	FileEntries         *FileEntryDict
	ResourceDirs        *FileEntryDict
	HasOutputTarget     bool
	VirtualFolders      *VirtualFolders
	PchCpp              *FileEntry
	ProjectFilename     string
	Configs             map[string]*Config
	GeneratedFilesDir   string
	Dependencies        *ProjectDependencies
	DependenciesInherit []*Project
	PchHeader           *FileEntry
	Resolved            bool
	Resolving           bool

	GenDataXcode *XcodeProjectConfig
	GenDataMsDev *MsDevProjectConfig
}

func NewProject(ws *Workspace, name string, subPath string, projectType ProjectType, settings *ProjectConfig) *Project {
	p := &Project{
		Workspace:           ws,
		Name:                name,
		Type:                projectType,
		ProjectAbsPath:      filepath.Join(ws.WorkspaceAbsPath, subPath),
		GenerateAbsPath:     filepath.Join(ws.WorkspaceAbsPath, subPath, ws.GenerateAbsPath, ws.Generator),
		Settings:            settings,
		FileEntries:         NewFileEntryDict(ws),
		ResourceDirs:        NewFileEntryDict(ws),
		HasOutputTarget:     false,
		Configs:             map[string]*Config{},
		Dependencies:        NewProjectDependencies(),
		DependenciesInherit: []*Project{},
		GenDataXcode:        &XcodeProjectConfig{},
	}
	p.VirtualFolders = NewVirtualFolders(p.ProjectAbsPath) // The path that is the root path of the virtual folder/file structure

	p.Settings.MultiThreadedBuild = Boolean(ws.Config.MultiThreadedBuild)
	p.Settings.MsDev.PlatformToolset = ws.Config.MsDev.PlatformToolset
	p.Settings.MsDev.WindowsTargetPlatformVersion = ws.Config.MsDev.WindowsTargetPlatformVersion
	p.Settings.Xcode.BundleIdentifier = "$(PROJECT_NAME)"

	for _, srcConfig := range ws.Configs {
		dstConfig := NewConfig(srcConfig.Name, ws, p)
		p.Configs[srcConfig.Name] = dstConfig
		dstConfig.init(srcConfig)
	}

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

func (p *Project) GetDefaultConfig() *Config {
	name := p.Workspace.Config.ConfigList[0]
	return p.GetOrCreateConfig(name)
}

func (p *Project) GetOrCreateConfig(name string) *Config {
	c := p.Configs[name]
	if c == nil {
		c = NewConfig(name, p.Workspace, p)
		p.Configs[name] = c
	}
	return c
}

func (p *Project) GenProjectGenUuid() {
	gd := &XcodeProjectConfig{}
	gd.XcodeProj = NewFileEntry()
	gd.XcodeProj.Init(filepath.Join(p.Workspace.GenerateAbsPath, p.Name, ".xcodeproj"), true)
	gd.PbxProj = filepath.Join(gd.XcodeProj.Path, "project.pbxproj")
	gd.Uuid = GenerateUUID()
	gd.TargetUuid = GenerateUUID()
	gd.TargetProductUuid = GenerateUUID()
	gd.ConfigListUuid = GenerateUUID()
	gd.TargetConfigListUuid = GenerateUUID()
	gd.DependencyProxyUuid = GenerateUUID()
	gd.DependencyTargetUuid = GenerateUUID()
	gd.DependencyTargetProxyUuid = GenerateUUID()

	for _, i := range p.FileEntries.Dict {
		f := p.FileEntries.Values[i]
		f.GenDataXcode.UUID = GenerateUUID()
		f.GenDataXcode.BuildUUID = GenerateUUID()
	}

	for _, i := range p.ResourceDirs.Dict {
		f := p.FileEntries.Values[i]
		f.GenDataXcode.UUID = GenerateUUID()
		f.GenDataXcode.BuildUUID = GenerateUUID()
	}

	for _, f := range p.VirtualFolders.Folders {
		f.GenData_xcode.UUID = GenerateUUID()
	}

	for _, config := range p.Configs {
		config.GenDataXcode.ProjectConfigUuid = GenerateUUID()
		config.GenDataXcode.TargetUuid = GenerateUUID()
		config.GenDataXcode.TargetConfigUuid = GenerateUUID()
	}

	p.GenDataXcode = gd
}

func (p *Project) resolve() error {
	if p.Resolved {
		return nil
	}

	if p.Resolving {
		return fmt.Errorf("cyclic dependencies in project %s", p.Name)
	}

	p.Resolving = true
	if err := p.resolveInternal(); err != nil {
		return err
	}
	p.Resolving = false

	//p.FileEntries.SortByKey()
	//p.VirtualFolders.SortByKey()
	return nil
}

func (p *Project) resolveInternal() error {
	p.resolveFiles()

	p.Type = p.Settings.Type

	if p.Settings.Type == ProjectTypeCppExe {
		p.HasOutputTarget = true
	} else if p.Settings.Type == ProjectTypeCExe {
		p.HasOutputTarget = true
	} else if p.Settings.Type == ProjectTypeCppDll {
		p.HasOutputTarget = true
	} else if p.Settings.Type == ProjectTypeCDll {
		p.HasOutputTarget = true
	} else if p.Settings.Type == ProjectTypeCppLib {
		p.HasOutputTarget = true
	} else if p.Settings.Type == ProjectTypeCLib {
		p.HasOutputTarget = true
	} else if p.Settings.Type == ProjectTypeCHeaders {
		// ...
	} else if p.Settings.Type == ProjectTypeCppHeaders {
		// ...
	} else {
		return fmt.Errorf("unknown project type %q from project %q", p.Settings.Type, p.Name)
	}

	for _, c := range p.Workspace.Configs {
		c.inherit(c)
	}

	for _, d := range p.Settings.Dependencies {
		dp := p.Workspace.Projects[d]
		if dp == nil {
			return fmt.Errorf("cannot find dependency project '%s' for project '%s'", d, p.Name)
		}
		if dp == p {
			return fmt.Errorf("project depends on itself, project='%s'", p.Name)
		}

		p.Dependencies.AddUnique(dp)
		if err := dp.resolve(); err != nil {
			return err
		}

		for pcKey, pc := range p.Configs {
			pc.inherit(dp.Configs[pcKey])
		}

		for _, dpdp := range dp.DependenciesInherit {
			p.DependenciesInherit = append(p.DependenciesInherit, dpdp)
		}

		p.DependenciesInherit = append(p.DependenciesInherit, dp)
	}

	for _, c := range p.Configs {
		c.finalize()
	}

	return nil
}

func (p *Project) resolveFiles() {
	p.GeneratedFilesDir = filepath.Join(p.Workspace.GenerateAbsPath, "_generated_", p.Name)

	if p.Settings.PchHeader != "" {
		p.PchHeader = NewFileEntry()
		p.PchHeader.Init(p.Settings.PchHeader, false)
	}

	for _, f := range p.FileEntries.Values {
		if f.Generated {
			p.VirtualFolders.AddFile(f)
		} else {
			p.VirtualFolders.AddFile(f)
		}
	}
}

func (p *Project) GlobFiles(dir string, pattern string) {
	pp := strings.Split(pattern, "^")
	path := filepath.Join(dir, pp[0])
	files, err := glob.GlobFiles(path, pp[1])
	if err != nil {
		panic(err)
	}

	exclusionFilter := NewExclusionFilter(p.Workspace.MakeTarget)
	for _, file := range files {
		if exclusionFilter.IsExcluded(file) {
			continue
		}
		p.FileEntries.Add(filepath.Join(pp[0], file), false)
	}
}