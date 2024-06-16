package axe

import (
	"fmt"
	"path/filepath"
)

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type WorkspaceConfig struct {
	ConfigList         []string            // The list of configurations to generate (e.g. ["Debug", "Release", "DebugTest", "ReleaseTest"])
	GenerateAbsPath    string              // The directory where the workspace and project files will be generated
	StartupProject     string              // The name of the project that will be marked as the startup project
	MultiThreadedBuild bool                // Whether to mark 'multi-threaded build' in the project files
	MsDev              *VisualStudioConfig // The project configuration to use for msdev
}

func NewWorkspaceConfig(workspacePath string, projectName string) *WorkspaceConfig {
	wsc := &WorkspaceConfig{
		ConfigList: []string{"Debug", "Release", "DebugTest", "ReleaseTest"},
	}
	wsc.GenerateAbsPath = filepath.Join(workspacePath, projectName, "target")
	wsc.StartupProject = projectName
	wsc.MultiThreadedBuild = true
	wsc.MsDev = NewVisualStudioConfig(VisualStudio2022)
	return wsc
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type Workspace struct {
	Config           *WorkspaceConfig           // The configuration for the workspace
	WorkspaceName    string                     // The name of the workspace (e.g. For VisualStudio -> "cbase.sln", for Xcode -> "cbase.xcworkspace")
	WorkspaceAbsPath string                     // The workspace directory is the path where all the projects and workspace are to be generated
	GenerateAbsPath  string                     // Where to generate the workspace and project files
	Generator        string                     // Name of the generator, ccore compiler define
	Configs          map[string]*Config         // The configuration instances for the workspace
	MakeTarget       MakeTarget                 // The make target for the workspace (e.g. contains details like OS, Compiler, Arch, etc.)
	StartupProject   *Project                   // The project instance that will be marked as the startup project
	ProjectList      *ProjectList               // The project list
	ProjectGroups    *ProjectGroups             // The project groups that are part of the workspace
	MasterWorkspace  *ExtraWorkspace            // The master workspace that contains all projects
	ExtraWorkspaces  map[string]*ExtraWorkspace // The extra workspaces that contain a subset of the projects
}

func NewWorkspace(ccoreAbsPath string, projectRelPath string) *Workspace {
	ws := &Workspace{
		Config:          NewWorkspaceConfig(ccoreAbsPath, projectRelPath),
		Configs:         make(map[string]*Config),
		ProjectList:     NewProjectList(),
		ProjectGroups:   NewProjectGroups(),
		ExtraWorkspaces: make(map[string]*ExtraWorkspace),
	}
	ws.MakeTarget = NewDefaultMakeTarget()
	ws.GenerateAbsPath = ws.Config.GenerateAbsPath
	return ws
}

func (ws *Workspace) NewProject(name string, subPath string, projectType ProjectType, settings *ProjectConfig) *Project {
	p := newProject(ws, name, subPath, projectType, settings)
	ws.ProjectList.Add(p)
	return p
}

func (ws *Workspace) DefaultConfigName() string {
	return ws.Config.ConfigList[0]
}

func (ws *Workspace) AddConfig(config *Config) {
	ws.Configs[config.Name] = config
	config.init(nil)
}

func (ws *Workspace) Finalize() error {
	if ws.StartupProject == nil {
		if startupProject, ok := ws.ProjectList.Get(ws.Config.StartupProject); ok {
			ws.StartupProject = startupProject
		} else {
			return fmt.Errorf("startup project \"%s\" not found as part of workspace \"%s\"", ws.Config.StartupProject, ws.WorkspaceName)
		}
	}

	for _, c := range ws.Configs {
		c.finalize()
	}

	for _, p := range ws.ProjectList.Values {
		ws.ProjectGroups.Add(p)
		if err := p.resolve(); err != nil {
			return err
		}
	}

	ws.MasterWorkspace = NewExtraWorkspace(ws, ws.WorkspaceName)

	ws.MasterWorkspace.ProjectList = NewProjectList()
	for _, p := range ws.ProjectList.Values {
		ws.MasterWorkspace.ProjectList.Add(p)
	}

	for _, ew := range ws.ExtraWorkspaces {
		ew.resolve()
	}

	return nil
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type ExtraWorkspaceConfig struct {
	Projects        []string
	Groups          []string
	ExcludeProjects []string
	ExcludeGroups   []string
}

type ExtraWorkspace struct {
	Workspace   *Workspace
	Name        string
	Config      *ExtraWorkspaceConfig
	ProjectList *ProjectList
	MsDev       *VisualStudioConfig
}

func (ew *ExtraWorkspace) HasProject(project *Project) bool {
	for _, p := range ew.ProjectList.Values {
		if p == project {
			return true
		}
	}
	return false
}

func (ew *ExtraWorkspace) resolve() {
	projectToAdd := NewProjectList()
	projectToRemove := NewProjectList()

	for _, name := range ew.Config.Projects {
		ew.Workspace.ProjectList.CollectByWildcard(name, projectToAdd)
	}

	for _, name := range ew.Config.Groups {
		for _, g := range ew.Workspace.ProjectGroups.Values {
			if PathMatchWildcard(g.Path, name, true) {
				for _, gp := range g.Projects {
					projectToAdd.Add(gp)
				}
			}
		}
	}

	for _, name := range ew.Config.ExcludeProjects {
		for _, p := range ew.Workspace.ProjectList.Values {
			if PathMatchWildcard(p.Name, name, true) {
				projectToRemove.Add(p)
			}
		}
	}

	for _, name := range ew.Config.ExcludeGroups {
		for _, g := range ew.Workspace.ProjectGroups.Values {
			if PathMatchWildcard(g.Path, name, true) {
				for _, gp := range g.Projects {
					projectToRemove.Add(gp)
				}
			}
		}
	}

	for _, p := range projectToAdd.Values {
		ew.ProjectList.Add(p)
	}
}

func NewExtraWorkspace(ws *Workspace, name string) *ExtraWorkspace {
	ew := &ExtraWorkspace{
		Workspace: ws,
		Name:      name,
		MsDev:     ws.Config.MsDev,
		Config:    &ExtraWorkspaceConfig{},
	}
	ew.MsDev = ws.Config.MsDev
	return ew
}
