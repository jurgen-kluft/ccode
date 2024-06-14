package xcode

import "fmt"

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

type ProjectConfig struct {
	Group                               string
	Type                                string
	GuiApp                              bool
	PchHeader                           string
	Dependencies                        []string
	MultithreadBuild                    Boolean
	CppAsObjcpp                         Boolean
	XcodeBundleIdentifier               string
	VisualcProjectTools                 string
	VisualcPlatformToolset              string
	VisualcWindowsTargetPlatformVersion string
}

type XcodeProjectConfig struct {
	Xcodeproj                 *FileEntry
	Pbxproj                   string
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
	Vcxproj string
	UUID    UUID
}

func NewXcodeProjectConfig() *XcodeProjectConfig {
	return &XcodeProjectConfig{
		Xcodeproj: NewFileEntry(),
	}
}

type Vs2015ProjectConfig struct {
	VcxProj string
	UUID    string
}

type MakefileProjectConfig struct {
	Makefile string
}

type Project struct {
	Workspace           *Workspace
	Input               ProjectConfig
	Group               string
	FileEntries         map[string]*FileEntry
	ResourceDirs        map[string]*FileEntry
	HasOutputTarget     bool
	VirtualFolders      *VirtualFolders
	PchCpp              *FileEntry
	AxprojFilename      string
	AxprojDir           string
	Name                string
	Type                ProjectType
	Configs             map[string]*Config
	GeneratedFileDir    string
	Dependencies        []*Project
	DependenciesInherit []*Project
	PchHeader           *FileEntry
	Resolved            bool
	Resolving           bool

	GenDataXcode *XcodeProjectConfig
	GenDataMsDev *MsDevProjectConfig
}

func NewProject(ws *Workspace, name string, input ProjectConfig) *Project {
	p := &Project{
		Workspace:           ws,
		Input:               input,
		FileEntries:         map[string]*FileEntry{},
		ResourceDirs:        map[string]*FileEntry{},
		VirtualFolders:      &VirtualFolders{},
		Name:                name,
		Configs:             map[string]*Config{},
		Dependencies:        []*Project{},
		DependenciesInherit: []*Project{},
		GenDataXcode:        &XcodeProjectConfig{},
	}

	p.Input.MultithreadBuild = Boolean(ws.Config.MultithreadBuild)
	p.Input.VisualcPlatformToolset = ws.Config.VisualcPlatformToolset
	p.Input.VisualcWindowsTargetPlatformVersion = ws.Config.VisualcWindowsTargetPlatformVersion
	p.Input.XcodeBundleIdentifier = "$(PROJECT_NAME)"

	for _, src := range ws.Configs {
		dst := &Config{}
		dst.Init(p, src, src.Name)
		p.Configs[src.Name] = dst
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

func (p *Project) DefaultConfig() *Config {
	name := p.Workspace.DefaultConfigName()
	c := p.Configs[name]
	if c == nil {
		c = NewDefaultConfig(p.Workspace)
		p.Configs[name] = c
	}
	return c
}

func (p *Project) GenProjectGenUuid() {
	gd := &XcodeProjectConfig{}
	gd.Xcodeproj = NewFileEntry()
	gd.Xcodeproj.Init(p.Workspace.BuildDir+p.Name+".xcodeproj", false, true, p.Workspace)
	gd.Pbxproj = gd.Xcodeproj.AbsPath + "/project.pbxproj"
	gd.Uuid = GenerateUUID()
	gd.TargetUuid = GenerateUUID()
	gd.TargetProductUuid = GenerateUUID()
	gd.ConfigListUuid = GenerateUUID()
	gd.TargetConfigListUuid = GenerateUUID()
	gd.DependencyProxyUuid = GenerateUUID()
	gd.DependencyTargetUuid = GenerateUUID()
	gd.DependencyTargetProxyUuid = GenerateUUID()

	for _, f := range p.FileEntries {
		f.GenDataXcode.UUID = GenerateUUID()
		f.GenDataXcode.BuildUUID = GenerateUUID()
	}

	for _, f := range p.ResourceDirs {
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

func (p *Project) Resolve() {
	if p.Resolved {
		return
	}

	if p.Resolving {
		panic("Cycle dependencies in project " + p.Name)
	}

	p.Resolving = true
	p.ResolveInternal()
	p.Resolving = false

	// fileEntries.sort();
	// virtualFolders.sort();
}

func (p *Project) ResolveInternal() error {
	p.ResolveFiles()

	if p.Input.Type == "cpp_exe" {
		p.HasOutputTarget = true
		p.Type = ProjectTypeCppExe
	} else if p.Input.Type == "c_exe" {
		p.HasOutputTarget = true
		p.Type = ProjectTypeCExe
	} else if p.Input.Type == "cpp_dll" {
		p.HasOutputTarget = true
		p.Type = ProjectTypeCppDll
	} else if p.Input.Type == "c_dll" {
		p.HasOutputTarget = true
		p.Type = ProjectTypeCDll
	} else if p.Input.Type == "cpp_lib" {
		p.HasOutputTarget = true
		p.Type = ProjectTypeCppLib
	} else if p.Input.Type == "c_lib" {
		p.HasOutputTarget = true
		p.Type = ProjectTypeCLib
	} else if p.Input.Type == "c_headers" {
		p.Type = ProjectTypeCHeaders
	} else if p.Input.Type == "cpp_headers" {
		p.Type = ProjectTypeCppHeaders
	} else {
		// panic("Unknown project type " + p.Input.Type + " from project " + p.Name)
		return fmt.Errorf("Unknown project type %s from project %s", p.Input.Type, p.Name)
	}

	i := 0
	for _, c := range p.Workspace.Configs {
		c.Inherit(c)
		i++
	}

	for _, d := range p.Input.Dependencies {
		dp := p.Workspace.Projects[d]
		if dp == nil {
			//panic("Cannot find dependency project '" + d + "' for project '" + p.Name + "'")
			return fmt.Errorf("Cannot find dependency project '%s' for project '%s'", d, p.Name)
		}

		if dp == p {
			// panic("project depends on itself, project='" + p.Name + "'")
			return fmt.Errorf("project depends on itself, project='%s'", p.Name)
		}

		p.Dependencies = append(p.Dependencies, dp)
		dp.Resolve()

		i = 0
		for _, c := range dp.Configs {
			c.Inherit(c)
			i++
		}

		for _, dpdp := range dp.DependenciesInherit {
			p.DependenciesInherit = append(p.DependenciesInherit, dpdp)
		}

		p.DependenciesInherit = append(p.DependenciesInherit, dp)
	}

	i = 0
	for _, c := range p.Configs {
		c.ComputeFinal()
		i++
	}

	return nil
}

func (p *Project) ResolveFiles() {
	p.GeneratedFileDir = p.Workspace.BuildDir + "_generated_/" + p.Name + "/"

	if p.Input.PchHeader != "" {
		p.PchHeader = NewFileEntry()
		p.PchHeader.Init(p.Input.PchHeader, false, false, p.Workspace)
	}

	for _, f := range p.FileEntries {
		if f.Generated {
			p.VirtualFolders.AddFile(p.Workspace.BuildDir, f)
		} else {
			p.VirtualFolders.AddFile(p.AxprojDir, f)
		}
	}
}
