package xcode

type WorkspaceConfig struct {
	ConfigList                          []string
	BuildDir                            string
	StartupProject                      string
	MultithreadBuild                    bool
	VisualcPlatformToolset              string
	VisualcWindowsTargetPlatformVersion string
}

type Workspace struct {
	Config              *WorkspaceConfig
	WorkspaceName       string
	Configs             map[string]*Config
	Generator           string
	HostOs              string
	HostCpu             string
	Os                  string
	Compiler            string
	Cpu                 string
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
