package axe

import "fmt"

type WorkspaceConfig struct {
	ConfigList         []string           // The list of configurations to generate (e.g. ["Debug", "Release", "Debug Test", "Release Test"])
	GenerateAbsPath    string             // The directory where the workspace and project files will be generated
	StartupProject     string             // The name of the project that will be marked as the startup project
	MultiThreadedBuild bool               // Whether to mark 'multi-threaded build' in the project files
	MsDev              VisualStudioConfig // The project configuration to use for msdev
}

func NewWorkspaceConfig() *WorkspaceConfig {
	wsc := &WorkspaceConfig{
		ConfigList: []string{"Debug", "Release", "Debug Test", "Release Test"},
	}
	wsc.GenerateAbsPath = "target"
	wsc.StartupProject = "main"
	wsc.MultiThreadedBuild = true
	wsc.MsDev = NewVisualStudioConfig(VisualStudio2022)
	return wsc
}

type Workspace struct {
	Config           *WorkspaceConfig           // The configuration for the workspace
	WorkspaceName    string                     // The name of the workspace (e.g. For VisualStudio -> "cbase.sln", for Xcode -> "cbase.xcworkspace")
	WorkspaceAbsPath string                     // The workspace directory is the path where all the projects and workspace are to be generated
	Configs          map[string]*Config         // The configuration instances for the workspace
	Generator        string                     // ccore compiler define
	MakeTarget       MakeTarget                 // The make target for the workspace (e.g. contains details like OS, Compiler, Arch, etc.)
	StartupProject   *Project                   // The project instance that will be marked as the startup project
	Projects         map[string]*Project        // The project instances that are part of the workspace
	ProjectGroups    *ProjectGroups             // The project groups that are part of the workspace
	GenerateAbsPath  string                     // Where to generate the workspace and project files
	MasterWorkspace  *ExtraWorkspace            // The master workspace that contains all projects
	ExtraWorkspaces  map[string]*ExtraWorkspace // The extra workspaces that contain a subset of the projects
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
	Workspace *Workspace
	Name      string
	Config    *ExtraWorkspaceConfig
	Projects  []*Project
	MsDev     VisualStudioConfig
}

func (ew *ExtraWorkspace) IndexOfProject(project *Project) int {
	for i, p := range ew.Projects {
		if p == project {
			return i
		}
	}
	return -1
}

func (ws *Workspace) AddConfig(config *Config) {
	ws.Configs[config.Name] = config
	config.init(nil)
}

func (ws *Workspace) Finalize() error {
	if ws.StartupProject == nil {
		if startupProject, ok := ws.Projects[ws.Config.StartupProject]; ok {
			ws.StartupProject = startupProject
		} else {
			return fmt.Errorf("startup project \"%s\" not found as part of workspace \"%s\"", ws.Config.StartupProject, ws.WorkspaceName)
		}
	}

	for _, c := range ws.Configs {
		c.finalize()
	}

	for _, p := range ws.Projects {
		ws.ProjectGroups.Add(p)
		if err := p.resolve(); err != nil {
			return err
		}
	}

	ws.MasterWorkspace.Name = ws.WorkspaceName
	ws.MasterWorkspace.Projects = make([]*Project, 0, len(ws.Projects))
	for _, p := range ws.Projects {
		ws.MasterWorkspace.Projects = append(ws.MasterWorkspace.Projects, p)
	}

	for _, ew := range ws.ExtraWorkspaces {
		ew.resolve()
	}

	return nil
}

func (ew *ExtraWorkspace) resolve() {
	projectToAdd := make([]*Project, 0)
	projectToRemove := make([]*Project, 0)

	for _, name := range ew.Config.Projects {
		for _, p := range ew.Workspace.Projects {
			if PathMatchWildcard(p.Name, name, true) {
				projectToAdd = append(projectToAdd, p)
			}
		}
	}

	for _, name := range ew.Config.Groups {
		for _, g := range ew.Workspace.ProjectGroups.Values {
			if PathMatchWildcard(g.Path, name, true) {
				for _, gp := range g.Projects {
					projectToAdd = append(projectToAdd, gp)
				}
			}
		}
	}

	for _, name := range ew.Config.ExcludeProjects {
		for _, p := range ew.Workspace.Projects {
			if PathMatchWildcard(p.Name, name, true) {
				projectToRemove = append(projectToRemove, p)
			}
		}
	}

	for _, name := range ew.Config.ExcludeGroups {
		for _, g := range ew.Workspace.ProjectGroups.Values {
			if PathMatchWildcard(g.Path, name, true) {
				for _, gp := range g.Projects {
					projectToRemove = append(projectToRemove, gp)
				}
			}
		}
	}

	for _, p := range projectToAdd {
		if ew.IndexOfProject(p) >= 0 {
			continue
		}
		if ew.IndexOfProject(p) >= 0 {
			continue
		}
		ew.Projects = append(ew.Projects, p)
	}
}

func NewWorkspace() *Workspace {
	ws := &Workspace{
		Config:          NewWorkspaceConfig(),
		Configs:         make(map[string]*Config),
		Projects:        make(map[string]*Project),
		ProjectGroups:   NewProjectGroups(nil),
		ExtraWorkspaces: make(map[string]*ExtraWorkspace),
	}
	ws.MakeTarget = NewDefaultMakeTarget()
	ws.GenerateAbsPath = ws.Config.GenerateAbsPath
	ws.MasterWorkspace = &ExtraWorkspace{
		Workspace: ws,
		Config: &ExtraWorkspaceConfig{
			Projects: []string{"*"},
		},
	}
	return ws
}
