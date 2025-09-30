package ide_generators

import (
	"fmt"
	"path/filepath"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
	"github.com/jurgen-kluft/ccode/denv"
)

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type CppStdType int

const (
	CppStdUnknown CppStdType = iota
	CppStd11
	CppStd14
	CppStd17
	CppStd20
	CppStd23
	CppStdLatest
)

func (cst CppStdType) String() string {
	switch cst {
	case CppStd11:
		return "c++11"
	case CppStd14:
		return "c++14"
	case CppStd17:
		return "c++17"
	case CppStd20:
		return "c++20"
	case CppStd23:
		return "c++23"
	case CppStdLatest:
		return "c++latest"
	}
	return ""
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type CppAdvancedType int

const (
	CppAdvancedNone CppAdvancedType = iota
	CppAdvancedSSE
	CppAdvancedSSE2
	CppAdvancedAVX
	CppAdvancedAVX2
	CppAdvancedAVX512
)

func (cat CppAdvancedType) IsEnabled() bool {
	return cat != CppAdvancedNone
}

func (cat CppAdvancedType) ToString() string {
	switch cat {
	case CppAdvancedSSE:
		return "SSE"
	case CppAdvancedSSE2:
		return "SSE2"
	case CppAdvancedAVX:
		return "AVX"
	case CppAdvancedAVX2:
		return "AVX2"
	case CppAdvancedAVX512:
		return "AVX512"
	}
	return ""
}
func (cat CppAdvancedType) Tundra(d denv.DevEnum, t denv.BuildTarget) string {
	if d.IsVisualStudio() && cat.IsEnabled() {
		return "/arch:" + cat.ToString()
	} else if d.CompilerIsClang() || d.CompilerIsGcc() {
		return "-m" + strings.ToLower(cat.ToString())
	}
	return ""
}

func (cat CppAdvancedType) VisualStudio() string {
	// Streaming SIMD Extensions (X86) (/arch:SSE)
	//    <EnableEnhancedInstructionSet>StreamingSIMDExtensions</EnableEnhancedInstructionSet>
	// Streaming SIMD Extensions 2 (X86) (/arch:SSE2)
	//    <EnableEnhancedInstructionSet>StreamingSIMDExtensions2</EnableEnhancedInstructionSet>
	// Advanced Vector Extensions (X86/X64) (/arch:AVX)
	//    <EnableEnhancedInstructionSet>AdvancedVectorExtensions</EnableEnhancedInstructionSet>
	// Advanced Vector Extensions 2 (X86/X64) (/arch:AVX2)
	//    <EnableEnhancedInstructionSet>AdvancedVectorExtensions2</EnableEnhancedInstructionSet>
	// Advanced Vector Extensions 512 (X86/X64) (/arch:AVX512)
	//    <EnableEnhancedInstructionSet>AdvancedVectorExtensions512</EnableEnhancedInstructionSet>

	cppAdvanced := ""
	switch cat {
	case CppAdvancedSSE:
		cppAdvanced = "StreamingSIMDExtensions"
	case CppAdvancedSSE2:
		cppAdvanced = "StreamingSIMDExtensions2"
	case CppAdvancedAVX:
		cppAdvanced = "AdvancedVectorExtensions"
	case CppAdvancedAVX2:
		cppAdvanced = "AdvancedVectorExtensions2"
	case CppAdvancedAVX512:
		cppAdvanced = "AdvancedVectorExtensions512"
	}
	return cppAdvanced
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type WorkspaceConfig struct {
	Dev                denv.DevEnum        // The development environment (tundra, make, xcode, vs2022, espmake)
	BuildTarget        denv.BuildTarget    // The build target (windows, linux, macos, etc.)
	BuildTargetOs      denv.BuildTargetOs  // The build target (windows, linux, macos, etc.)
	GenerateAbsPath    string              // The directory where the workspace and project files will be generated
	StartupProject     string              // The name of the project that will be marked as the startup project
	CppStd             CppStdType          // The C++ standard to use for this workspace and all projects
	CppAdvanced        CppAdvancedType     // The C++ advanced features to use for this workspace and all projects
	MultiThreadedBuild bool                // Whether to mark 'multi-threaded build' in the project files
	MsDev              *VisualStudioConfig // The project configuration to use for msdev

	ExeTargetPrefix string
	ExeTargetSuffix string
	DllTargetPrefix string
	DllTargetSuffix string
	LibTargetPrefix string
	LibTargetSuffix string
}

func NewWorkspaceConfig(_dev denv.DevEnum, _buildTargetOs denv.BuildTargetOs, workspacePath string, projectName string) *WorkspaceConfig {
	wsc := &WorkspaceConfig{}
	wsc.Dev = _dev
	wsc.BuildTargetOs = _buildTargetOs
	wsc.GenerateAbsPath = filepath.Join(workspacePath, projectName, "target")
	wsc.StartupProject = projectName
	wsc.CppStd = CppStd17
	wsc.CppAdvanced = CppAdvancedNone
	wsc.MultiThreadedBuild = true
	wsc.MsDev = NewVisualStudioConfig(VisualStudio2022)

	return wsc
}

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type Workspace struct {
	Config           *WorkspaceConfig   // The configuration for the workspace
	WorkspaceName    string             // The name of the workspace (e.g. For VisualStudio -> "cbase.sln", for Xcode -> "cbase.xcworkspace")
	WorkspaceAbsPath string             // The workspace directory is the path where all the projects and workspace are to be generated
	GenerateAbsPath  string             // Where to generate the workspace and project files
	BuildTargetHost  denv.BuildTarget   // The make target for the workspace (e.g. contains details like OS, Compiler, Arch, etc.)
	BuildTarget      denv.BuildTarget   // The make target for the workspace (e.g. contains details like OS, Compiler, Arch, etc.)
	BuildTargetOs    denv.BuildTargetOs // The make target for the workspace (e.g. contains details like OS, Compiler, Arch, etc.)
	StartupProject   *Project           // The project instance that will be marked as the startup project
	ProjectList      *ProjectList       // The project list
	ProjectGroups    *ProjectGroups     // The project groups that are part of the workspace
	//cMasterWorkspace   *ExtraWorkspace            // The master workspace that contains all projects
	//ExtraWorkspaces   map[string]*ExtraWorkspace // The extra workspaces that contain a subset of the projects
}

func NewWorkspace(wsc *WorkspaceConfig) *Workspace {
	ws := &Workspace{
		Config:        wsc,
		ProjectList:   NewProjectList(),
		ProjectGroups: NewProjectGroups(),
		//ExtraWorkspaces: make(map[string]*ExtraWorkspace),
	}
	ws.BuildTarget = ws.Config.BuildTarget
	ws.BuildTargetOs = ws.Config.BuildTargetOs
	ws.BuildTargetHost = denv.GetBuildTargetTargettingHost()
	ws.GenerateAbsPath = ws.Config.GenerateAbsPath

	if ws.BuildTargetOs.Windows() {
		wsc.ExeTargetSuffix = ".exe"
		wsc.DllTargetSuffix = ".dll"
	} else {
		wsc.ExeTargetSuffix = ""
		wsc.DllTargetSuffix = ".so"
	}

	if ws.Config.Dev.IsVisualStudio() {
		wsc.LibTargetPrefix = ""
		wsc.LibTargetSuffix = ".lib"
	} else {
		wsc.LibTargetPrefix = "lib"
		wsc.LibTargetSuffix = ".a"
	}

	return ws
}

func (ws *Workspace) NewProject2(prj *denv.DevProject, settings *ProjectSettings) *Project {
	p := newProject2(ws.BuildTarget, prj, ws.GenerateAbsPath, settings)
	ws.ProjectList.Add(p)
	return p
}

func (ws *Workspace) Resolve(dev denv.DevEnum) error {
	if ws.StartupProject == nil {
		if startupProject, ok := ws.ProjectList.Get(ws.Config.StartupProject); ok {
			ws.StartupProject = startupProject
		} else {
			return fmt.Errorf("startup project \"%s\" not found as part of workspace \"%s\"", ws.Config.StartupProject, ws.WorkspaceName)
		}
	}

	// Sort ProjectList in topological order
	err := ws.ProjectList.TopoSort()
	if err != nil {
		return err
	}

	// Now Resolve all projects
	for _, p := range ws.ProjectList.Values {
		ws.ProjectGroups.Add(p)
		if err := p.Resolve(dev); err != nil {
			return err
		}
	}

	//ws.MasterWorkspace = NewExtraWorkspace(ws, ws.WorkspaceName)
	// for _, p := range ws.ProjectList.Values {
	// 	ws.MasterWorkspace.ProjectList.Add(p)
	// }

	// for _, ew := range ws.ExtraWorkspaces {
	// 	ew.resolve()
	// }

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
			if corepkg.PathMatchWildcard(g.Path, name, true) {
				for _, gp := range g.Projects {
					projectToAdd.Add(gp)
				}
			}
		}
	}

	for _, name := range ew.Config.ExcludeProjects {
		for _, p := range ew.Workspace.ProjectList.Values {
			if corepkg.PathMatchWildcard(p.Name, name, true) {
				projectToRemove.Add(p)
			}
		}
	}

	for _, name := range ew.Config.ExcludeGroups {
		for _, g := range ew.Workspace.ProjectGroups.Values {
			if corepkg.PathMatchWildcard(g.Path, name, true) {
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
		Workspace:   ws,
		Name:        name,
		Config:      &ExtraWorkspaceConfig{},
		ProjectList: NewProjectList(),
		MsDev:       ws.Config.MsDev,
	}
	ew.MsDev = ws.Config.MsDev
	return ew
}
