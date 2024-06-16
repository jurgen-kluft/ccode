package axe

import (
	"path/filepath"
)

type MsDevGenerator struct {
	LastGenId  UUID
	Workspace  *Workspace
	VcxProjCpu string
}

func NewMsDevGenerator(ws *Workspace) *MsDevGenerator {
	g := &MsDevGenerator{
		LastGenId: GenerateUUID(),
		Workspace: ws,
	}
	g.init(ws)
	return g
}

func (g *MsDevGenerator) Generate() {
	for _, p := range g.Workspace.ProjectList.Values {
		g.genProject(p)
		g.genProjectFilters(p)
	}
	g.genWorkspace(g.Workspace.MasterWorkspace)
	for _, ew := range g.Workspace.ExtraWorkspaces {
		g.genWorkspace(ew)
	}
}

func (g *MsDevGenerator) init(ws *Workspace) {
	if ws.MakeTarget == nil {
		ws.MakeTarget = NewDefaultMakeTarget()
	}

	if ws.MakeTarget.ArchIsX64() {
		g.VcxProjCpu = "x64"
	} else {
		g.VcxProjCpu = "Win32"
	}

	for _, p := range ws.ProjectList.Values {
		g.genProject(p)
		g.genProjectFilters(p)
	}

	g.genWorkspace(ws.MasterWorkspace)
	for _, ew := range ws.ExtraWorkspaces {
		g.genWorkspace(ew)
	}
}

func (g *MsDevGenerator) PlatformToolset(proj *Project) string {
	if proj.Settings.MsDev.PlatformToolset != "" {
		return proj.Settings.MsDev.PlatformToolset
	}
	return g.Workspace.Config.MsDev.PlatformToolset
}

func (g *MsDevGenerator) TargetPlatformVersion(proj *Project) string {
	if proj.Settings.MsDev.WindowsTargetPlatformVersion != "" {
		return proj.Settings.MsDev.WindowsTargetPlatformVersion
	}
	return g.Workspace.Config.MsDev.WindowsTargetPlatformVersion
}

func (g *MsDevGenerator) genProject(proj *Project) {
	projectFilepath := filepath.Join(g.Workspace.GenerateAbsPath, proj.ProjectFilename+".vcxproj")

	proj.GenDataMsDev.UUID = GenerateUUID()

	wr := NewXmlWriter()
	{
		wr.WriteHeader()
		tag := wr.TagScope("Project")

		wr.Attr("DefaultTargets", "Build")
		wr.Attr("ToolsVersion", proj.Settings.MsDev.ProjectTools)
		wr.Attr("xmlns", "http://schemas.microsoft.com/developer/msbuild/2003")

		{
			tag := wr.TagScope("ItemGroup")
			wr.Attr("Label", "ProjectConfigurations")

			for _, config := range proj.Configs {
				tag := wr.TagScope("ProjectConfiguration")
				wr.Attr("Include", config.Name+"|"+g.VcxProjCpu)

				wr.TagWithBody("Configuration", config.Name)
				wr.TagWithBody("Platform", g.VcxProjCpu)
				tag.Close()
			}
			tag.Close()
		}
		{
			tag := wr.TagScope("PropertyGroup")
			wr.Attr("Label", "Globals")
			wr.TagWithBody("ProjectGuid", proj.GenDataMsDev.UUID.String())
			wr.TagWithBody("Keyword", "Win32Proj")
			wr.TagWithBody("RootNamespace", proj.Name)
			wr.TagWithBody("WindowsTargetPlatformVersion", proj.Settings.MsDev.WindowsTargetPlatformVersion)
			tag.Close()
		}

		g.genProjectFiles(wr, proj)

		//-----------
		{
			tag := wr.TagScope("Import")
			wr.Attr("Project", "$(VCTargetsPath)\\Microsoft.Cpp.Default.props")
			tag.Close()
		}
		{
			tag := wr.TagScope("PropertyGroup")
			wr.Attr("Label", "Configuration")

			var productType string
			if proj.TypeIsHeaders() || proj.TypeIsLib() {
				productType = "StaticLibrary"
			} else if proj.TypeIsDll() {
				productType = "DynamicLibrary"
			} else if proj.TypeIsExe() {
				productType = "Application"
			} else {
				panic("Unhandled project type")
			}

			wr.TagWithBody("ConfigurationType", productType)
			wr.TagWithBody("CharacterSet", "Unicode")

			if g.Workspace.MakeTarget.OSIsLinux() {
				wr.TagWithBody("PlatformToolset", "Remote_GCC_1_0")
				if g.Workspace.MakeTarget.CompilerIsGcc() {
					wr.TagWithBody("RemoteCCompileToolExe", "gcc")
					wr.TagWithBody("RemoteCppCompileToolExe", "g++")
				} else if g.Workspace.MakeTarget.CompilerIsClang() {
					wr.TagWithBody("RemoteCCompileToolExe", "clang")
					wr.TagWithBody("RemoteCppCompileToolExe", "clang++")
				} else {
					panic("Unsupported compiler")
				}

				relBuildDir := PathGetRel(g.Workspace.GenerateAbsPath, g.Workspace.WorkspaceAbsPath)

				remoteRootDir := "_vsForLinux/" + g.Workspace.MakeTarget.OSAsString() + "/" + relBuildDir
				wr.TagWithBody("RemoteRootDir", remoteRootDir)

				projectDir := "$(RemoteRootDir)/../ProjectDir_" + proj.Name
				wr.TagWithBody("RemoteProjectDir", projectDir)

				fileToCopy := ""
				for _, f := range proj.FileEntries.Values {
					fileToCopy += f.Path + ":=" + projectDir + "/" + f.Path + ";"
				}

				wr.TagWithBody("SourcesToCopyRemotelyOverride", "")
				wr.TagWithBody("AdditionalSourcesToCopyMapping", fileToCopy)
			} else {
				wr.TagWithBody("PlatformToolset", g.PlatformToolset(proj))
			}

			tag.Close()
		}

		//-----------
		{
			tag := wr.TagScope("Import")
			wr.Attr("Project", "$(VCTargetsPath)\\Microsoft.Cpp.props")
			tag.Close()
		}
		{
			tag := wr.TagScope("ImportGroup")
			wr.Attr("Label", "PropertySheets")
			tag.Close()
		}

		if g.Workspace.MakeTarget.OSIsLinux() {
			{
				tag := wr.TagScope("ImportGroup")
				wr.Attr("Label", "ExtensionSettings")
				tag.Close()
			}

			{
				tag := wr.TagScope("ImportGroup")
				wr.Attr("Label", "Shared")
				tag.Close()
			}
		}

		{
			tag := wr.TagScope("PropertyGroup")
			wr.Attr("Label", "UserMacros")
			tag.Close()
		}

		//-----------
		for _, config := range proj.Configs {
			g.genProjectConfig(wr, proj, config)
		}

		//-----------
		{
			tag := wr.TagScope("Import")
			wr.Attr("Project", "$(VCTargetsPath)\\Microsoft.Cpp.targets")
			tag.Close()
		}
		{
			tag := wr.TagScope("ImportGroup")
			wr.Attr("Label", "ExtensionTargets")
			tag.Close()
		}

		tag.Close()
	}

	wr.WriteToFile(projectFilepath)
}

func (g *MsDevGenerator) genProjectFiles(wr *XmlWriter, proj *Project) {
	tag := wr.TagScope("ItemGroup")

	g.genProjectPch(wr, proj)

	for _, f := range proj.FileEntries.Values {
		tagName := ""
		excludedFromBuild := false
		remoteCopyFile := false
		switch f.Type {
		case FileTypeIxx, FileTypeCSource, FileTypeCppSource:
			tagName = "ClCompile"
			excludedFromBuild = f.ExcludedFromBuild
		case FileTypeCuHeader, FileTypeCuSource:
			tagName = "CudaCompile"
			excludedFromBuild = f.ExcludedFromBuild
		default:
			tagName = "ClInclude"
		}

		tag := wr.TagScope(tagName)
		wr.Attr("Include", f.Path)
		if excludedFromBuild {
			wr.TagWithBody("ExcludedFromBuild", "true")
		}
		if remoteCopyFile {
			wr.TagWithBody("RemoteCopyFile", "false")
		}
		tag.Close()
	}

	tag.Close()
}

func (g *MsDevGenerator) genProjectPch(wr *XmlWriter, proj *Project) {
	if proj.PchHeader == nil {
		return
	}

	var pchExt string
	if proj.TypeIsCpp() {
		pchExt = ".cpp"
	} else if proj.TypeIsC() {
		pchExt = ".c"
	} else {
		return
	}

	filename := proj.GeneratedFilesDir + proj.Name + "-precompiledHeader" + pchExt
	tmp := PathGetRel(proj.PchHeader.Path, proj.GeneratedFilesDir)

	proj.PchHeader.Init(filename, true)
	code := "//-- Auto Generated File for Visual C++ precompiled header\n"
	code += "#include \"" + tmp + "\"\n"

	WriteTextFile(filename, code)

	tag := wr.TagScope("ClCompile")
	{
		wr.Attr("Include", proj.PchHeader.Path)
		wr.TagWithBody("PrecompiledHeader", "Create")
	}
	tag.Close()
}

func (g *MsDevGenerator) genProjectConfig(wr *XmlWriter, proj *Project, config *Config) {
	cond := "'$(Configuration)|$(Platform)'=='" + config.Name + "|" + g.VcxProjCpu + "'"
	{
		tag := wr.TagScope("PropertyGroup")
		wr.Attr("Condition", cond)

		var outputTarget string
		if g.Workspace.MakeTarget.OSIsLinux() {
			outputTarget = "../" + g.Workspace.MakeTarget.OSAsString() + "/"
		}
		outputTarget += PathDirname(config.OutputTarget.Path)
		if outputTarget == "" {
			outputTarget = "."
		}
		outputTarget += "/"

		intermediaDir := config.BuildTmpDir.Path
		if intermediaDir == "" {
			intermediaDir = "."
		}
		intermediaDir += "/"

		targetName := PathBasename(config.OutputTarget.Path, false)
		targetExt := PathExtension(config.OutputTarget.Path)

		wr.TagWithBody("OutDir", outputTarget)
		wr.TagWithBody("IntDir", intermediaDir)
		wr.TagWithBody("TargetName", targetName)
		if targetExt != "" {
			wr.TagWithBody("TargetExt", "."+targetExt)
		}

		tag.Close()
	}
	{
		tag := wr.TagScope("ItemDefinitionGroup")
		wr.Attr("Condition", cond)
		{
			tag := wr.TagScope("ClCompile")

			cppStd := "stdcpp14"
			switch config.CppStd {
			case "c++11", "c++14":
				cppStd = "stdcpp14"
			case "c++17":
				cppStd = "stdcpp17"
			case "latest":
				cppStd = "stdcpplatest"
			}

			wr.TagWithBody("LanguageStandard", cppStd)
			wr.TagWithBody("PreprocessorDefinitions", "%(PreprocessorDefinitions)")

			if g.Workspace.MakeTarget.OSIsLinux() {
				wr.TagWithBody("Verbose", "true")

				if config.IsDebug {
					wr.TagWithBody("DebugInformationFormat", "FullDebug")
					wr.TagWithBody("Optimization", "Disabled")
					wr.TagWithBody("OmitFramePointers", "false")
				} else {
					wr.TagWithBody("DebugInformationFormat", "None")
					wr.TagWithBody("Optimization", "Full")
					wr.TagWithBody("OmitFramePointers", "true")
				}

			} else {
				wr.TagWithBody("SDLCheck", "true")
				wr.TagWithBodyBool("MultiProcessorCompilation", proj.Settings.MultiThreadedBuild.Bool())

				if g.Workspace.MakeTarget.CompilerIsClang() {
					wr.TagWithBody("DebugInformationFormat", "None")
				} else {
					wr.TagWithBody("DebugInformationFormat", "ProgramDatabase")
				}

				if config.IsDebug {
					wr.TagWithBody("Optimization", "Disabled")
					wr.TagWithBody("RuntimeLibrary", "MultiThreadedDebugDLL")
					wr.TagWithBody("LinkIncremental", "true")
				} else {
					wr.TagWithBody("Optimization", "MaxSpeed")
					wr.TagWithBody("WholeProgramOptimization", "true")
					wr.TagWithBody("RuntimeLibrary", "MultiThreadedDLL")
					wr.TagWithBody("FunctionLevelLinking", "true")
					wr.TagWithBody("IntrinsicFunctions", "true")
					wr.TagWithBody("WholeProgramOptimization", "true")
					wr.TagWithBody("BasicRuntimeChecks", "Default")
				}
			}

			if proj.PchHeader != nil {
				wr.TagWithBody("PrecompiledHeader", "Use")

				pch := "$(ProjectDir)" + proj.PchHeader.Path
				wr.TagWithBody("PrecompiledHeaderFile", pch)
				wr.TagWithBody("ForcedIncludeFiles", pch)
			}

			if config.WarningLevel != "" {
				wr.TagWithBody("WarningLevel", config.WarningLevel)
			}

			if config.CppEnableModules {
				wr.TagWithBodyBool("EnableModules", config.CppEnableModules)
			}

			wr.TagWithBodyBool("TreatWarningAsError", config.WarningAsError)

			g.genConfigOption(wr, "DisableSpecificWarnings", config.DisableWarning.FinalDict, "")
			g.genConfigOption(wr, "PreprocessorDefinitions", config.CppDefines.FinalDict, "")
			g.genConfigOption(wr, "AdditionalIncludeDirectories", config.IncludeDirs.FinalDict, "")

			for key, i := range config.VisualStudioClCompile.Entries {
				wr.TagWithBody(key, config.VisualStudioClCompile.Values[i])
			}

			tag.Close()
		}
		{
			tag := wr.TagScope("Link")

			if proj.TypeIsDll() {
				wr.TagWithBody("ImportLibrary", config.OutputLib.Path)
			}

			if proj.TypeIsExeOrDll() {
				g.genConfigOption(wr, "AdditionalLibraryDirectories", config.LinkDirs.FinalDict, "")

				optName := "AdditionalDependencies"
				relativeTo := ""
				if g.Workspace.MakeTarget.OSIsLinux() {
					relativeTo = "$(RemoteRootDir)/"
				}
				tmp := config.LinkLibs.FinalDict.Concatenated("", ";")
				tmp += config.LinkFiles.FinalDict.Concatenated(relativeTo, ";")
				tmp += "%(" + optName + ")"
				wr.TagWithBody(optName, tmp)
			}

			if g.Workspace.MakeTarget.OSIsLinux() {
				wr.TagWithBody("VerboseOutput", "true")

				tmp := config.LinkFlags.FinalDict.Concatenated(" -Wl,", "")
				wr.TagWithBody("AdditionalOptions", tmp)

				if config.IsDebug {
					wr.TagWithBody("DebuggerSymbolInformation", "true")
				} else {
					wr.TagWithBody("DebuggerSymbolInformation", "OmitAllSymbolInformation")
				}

			} else {
				wr.TagWithBody("SubSystem", "Console")
				wr.TagWithBody("GenerateDebugInformation", "true")

				tmp := config.LinkFlags.FinalDict.Concatenated(" ", "")
				wr.TagWithBody("AdditionalOptions", tmp)

				if !config.IsDebug {
					wr.TagWithBodyBool("EnableCOMDATFolding", true)
					wr.TagWithBodyBool("OptimizeReferences", true)
				}
			}

			for key, i := range config.VisualStudioLink.Entries {
				wr.TagWithBody(key, config.VisualStudioLink.Values[i])
			}

			tag.Close()
		}

		if config.OutputTarget.Path != "" {
			tag := wr.TagScope("RemotePostBuildEvent")
			{
				cmd := "mkdir -p \"" + "$(RemoteRootDir)/" + PathDirname(config.OutputTarget.Path) + "\";"
				cmd += "cp -f \"$(RemoteProjectDir)/" + config.OutputTarget.Path + "\" \"$(RemoteRootDir)/" + config.OutputTarget.Path + "\""
				wr.TagWithBody("Command", cmd)
			}
			tag.Close()
		}

		tag.Close()
	}
}

func (g *MsDevGenerator) genConfigOption(wr *XmlWriter, name string, value *KeyValueDict, relativeTo string) {
	option := value.Concatenated(relativeTo, ";")
	option += "%(" + name + ")"
	wr.TagWithBody(name, option)
}

func (g *MsDevGenerator) genWorkspace(ws *ExtraWorkspace) {
	visualStudioSolutionFilepath := filepath.Join(g.Workspace.GenerateAbsPath, ws.Workspace.WorkspaceName+".sln")

	sb := NewLineWriter()

	sb.WriteManyLines(ws.MsDev.SlnHeader)

	{
		sb.WriteLines("", "# ---- projects ----", "")
		for _, proj := range ws.ProjectList.Values {
			sb.Write("Project(\"{8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942}\") = ")
			sb.WriteLine("\"" + proj.Name + "\", \"" + proj.Name + ".vcxproj\", \"" + proj.GenDataMsDev.UUID.String() + "\"")

			if len(proj.DependenciesInherit.Values) > 0 {
				if proj.TypeIsExeOrDll() {
					sb.WriteLine("\tProjectSection(ProjectDependencies) = postProject")
					for _, dp := range proj.DependenciesInherit.Values {
						sb.WriteLine("\t\t" + dp.GenDataMsDev.UUID.String() + " = " + dp.GenDataMsDev.UUID.String())
					}
					sb.WriteLine("\tEndProjectSection")
				}
			}
			sb.WriteLine("EndProject")
		}
	}
	{
		root := g.Workspace.ProjectGroups.Root
		//sb += "\n# ---- Groups ----\n"
		sb.WriteLines("", "# ---- Groups ----", "")
		for _, c := range g.Workspace.ProjectGroups.Values {
			if c == root {
				continue
			}
			c.MsDev.UUID = GenerateUUID()

			catName := PathBasename(c.Path, true)
			// sb += "Project(\"{2150E333-8FDC-42A3-9474-1A3956D46DE8}\") = \""
			// sb += catName + "\", \"" + catName + "\", \"" + c.MsDev.UUID.String() + "\"\n"
			// sb += "EndProject\n"
			sb.Write("Project(\"{2150E333-8FDC-42A3-9474-1A3956D46DE8}\") = \"")
			sb.Write(catName)
			sb.Write("\", \"")
			sb.Write(catName)
			sb.Write("\", \"")
			sb.Write(c.MsDev.UUID.String())
			sb.WriteLine("\"")
		}

		sb.WriteLines("", "# ----  (ProjectGroups) -> parent ----", "")
		sb.WriteLine("Global")
		sb.WriteLine("\tGlobalSection(NestedProjects) = preSolution")

		for _, c := range g.Workspace.ProjectGroups.Values {
			if c.Parent != nil && c.Parent != root {
				sb.Write("\t\t")
				sb.Write(c.MsDev.UUID.String())
				sb.Write(" = ")
				sb.WriteLine(c.Parent.MsDev.UUID.String())
			}

			for _, proj := range c.Projects {
				if proj == nil {
					continue
				}
				if !ws.HasProject(proj) {
					continue
				}
				sb.Write("\t\t")
				sb.Write(proj.GenDataMsDev.UUID.String())
				sb.Write(" = ")
				sb.WriteLine(c.MsDev.UUID.String())
			}
		}
		sb.WriteLine("\tEndGlobalSection")
		sb.WriteLine("EndGlobal")
	}

	sb.WriteToFile(visualStudioSolutionFilepath)
}

func (g *MsDevGenerator) genProjectFilters(proj *Project) {
	projectFiltersFilepath := filepath.Join(g.Workspace.GenerateAbsPath, proj.ProjectFilename+".vcxproj.filters")

	wr := NewXmlWriter()
	{
		wr.WriteHeader()

		tag := wr.TagScope("Project")
		wr.Attr("ToolsVersion", "4.0")
		wr.Attr("xmlns", "http://schemas.microsoft.com/developer/msbuild/2003")

		if proj.PchCpp != nil && len(proj.PchCpp.Path) > 0 {
			proj.VirtualFolders.AddFile(proj.PchCpp)
		}

		{
			tag := wr.TagScope("ItemGroup")
			for _, i := range proj.VirtualFolders.Folders {
				if i.Path == "." {
					continue
				}
				tag := wr.TagScope("Filter")
				winPath := PathWindowsPath(i.Path)
				wr.Attr("Include", winPath)
				tag.Close()
			}
			tag.Close()
		}

		{
			tag := wr.TagScope("ItemGroup")
			for _, i := range proj.VirtualFolders.Folders {
				for _, f := range i.Files {
					if f == nil {
						continue
					}

					var typeName string
					switch f.Type {
					case FileTypeIxx, FileTypeCSource, FileTypeCppSource:
						typeName = "ClCompile"
					default:
						typeName = "ClInclude"
					}

					tag := wr.TagScope(typeName)
					wr.Attr("Include", f.Path)
					if len(i.Path) > 0 {
						winPath := PathWindowsPath(i.Path)
						wr.TagWithBody("Filter", winPath)
					}
					tag.Close()
				}
			}
			tag.Close()
		}
		tag.Close()
	}

	wr.WriteToFile(projectFiltersFilepath)
}
