package xcode

type WorkspaceConfig struct {
	ConfigList                          []string
	BuildDir                            string
	StartupProject                      string
	MultithreadBuild                    bool
	VisualcProjectTools                 string
	VisualcPlatformToolset              string
	VisualcWindowsTargetPlatformVersion string
}

func NewWorkspaceConfig() *WorkspaceConfig {
	wsc := &WorkspaceConfig{
		ConfigList: []string{"Debug", "Release"},
	}
	wsc.BuildDir = "target"
	wsc.StartupProject = "main"
	wsc.MultithreadBuild = true
	wsc.VisualcProjectTools = "14.0"
	wsc.VisualcPlatformToolset = "v142"
	wsc.VisualcWindowsTargetPlatformVersion = "10.0"
	return wsc
}

type Workspace struct {
	Config              *WorkspaceConfig
	WorkspaceName       string
	Configs             map[string]*Config
	Generator           string
	MakeTarget          MakeTarget
	PlatformName        string
	StartupProject      *Project
	Projects            map[string]*Project
	ProjectGroups       ProjectGroupDict
	AxworkspaceFilename string
	AxworkspaceDir      string
	BuildDir            string
	MasterWorkspace     *ExtraWorkspace
	ExtraWorkspaces     map[string]*ExtraWorkspace
	Xcworkspace         string
}

func (ws *Workspace) DefaultConfigName() string {
	return ws.Config.ConfigList[0]
}

type ExtraWorkspaceConfig struct {
	Projects        []string
	Groups          []string
	ExcludeProjects []string
	ExcludeGroups   []string
}

type ExtraWorkspace struct {
	Config        *ExtraWorkspaceConfig
	Name          string
	Projects      []*Project
	GenDataVs2015 struct {
		Sln string
	}
}

func (ew *ExtraWorkspace) IndexOfProject(project *Project) int {
	for i, p := range ew.Projects {
		if p == project {
			return i
		}
	}
	return -1
}
