package xcode

import (
	"path/filepath"
)

type MsDevGenerator struct {
	LastGenId  UUID
	Workspace  *Workspace
	VcxProjCpu string
}

func NewMsDevGenerator(ws *Workspace) *MsDevGenerator {
	return &MsDevGenerator{
		LastGenId: GenerateUUID(),
		Workspace: ws,
	}
}

func (g *MsDevGenerator) Init(ws *Workspace) {
	if ws.MakeTarget == nil {
		ws.MakeTarget = NewDefaultMakeTarget()
	}

	if ws.MakeTarget.ArchIsX64() {
		g.VcxProjCpu = "x64"
	} else {
		g.VcxProjCpu = "Win32"
	}

	for _, p := range ws.Projects {
		g.genProject(p)
		g.genVcxprojFilters(p)
	}

	g.genWorkspace(ws.MasterWorkspace)
	for _, ew := range ws.ExtraWorkspaces {
		g.genWorkspace(ew)
	}
}

func (g *MsDevGenerator) PlatformToolset(proj *Project) string {
	if proj.Input.MsDevPlatformToolset != "" {
		return proj.Input.MsDevPlatformToolset
	}
	return g.Workspace.Config.VisualcPlatformToolset
}

func (g *MsDevGenerator) TargetPlatformVersion(proj *Project) string {
	if proj.Input.MsDevWindowsTargetPlatformVersion != "" {
		return proj.Input.MsDevWindowsTargetPlatformVersion
	}
	return g.Workspace.Config.VisualcWindowsTargetPlatformVersion
}

func (g *MsDevGenerator) genProject(proj *Project) {
	proj.GenDataMsDev.VcxProj = filepath.Join(g.Workspace.BuildDir, proj.Name, ".vcxproj")

	proj.GenDataMsDev.UUID = GenerateUUID()

	wr := NewXmlWriter()
	{
		wr.writeHeader()
		tag := wr.TagScope("Project")

		wr.Attr("DefaultTargets", "Build")
		wr.Attr("ToolsVersion", proj.Input.MsDevProjectTools)
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

			if proj.Input.MsDevProjectTools != "" {
				wr.TagWithBody("WindowsTargetPlatformVersion", proj.Input.MsDevWindowsTargetPlatformVersion)
			} else {
				wr.TagWithBody("WindowsTargetPlatformVersion", g.TargetPlatformVersion(proj))
			}
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

				relBuildDir := PathGetRel(g.Workspace.BuildDir, g.Workspace.AxworkspaceDir)

				remoteRootDir := "_vsForLinux/" + g.Workspace.PlatformName + "/" + relBuildDir
				wr.TagWithBody("RemoteRootDir", remoteRootDir)

				projectDir := "$(RemoteRootDir)/../ProjectDir_" + proj.Name
				wr.TagWithBody("RemoteProjectDir", projectDir)

				fileToCopy := ""
				for _, i := range proj.FileEntries.Dict {
					f := proj.FileEntries.List[i]
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

	WriteTextFile(proj.GenDataMsDev.VcxProj, wr.Buffer.String())
}

func (g *MsDevGenerator) genProjectFiles(wr *XmlWriter, proj *Project) {
	tag := wr.TagScope("ItemGroup")

	g.genProjectPch(wr, proj)

	for _, i := range proj.FileEntries.Dict {
		var tagName string
		var excludedFromBuild bool

		f := proj.FileEntries.List[i]

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

		wr.TagWithBody("RemoteCopyFile", "false")
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

	filename := proj.GeneratedFileDir + proj.Name + "-precompiledHeader" + pchExt
	tmp := PathGetRel(proj.PchHeader.AbsPath, proj.GeneratedFileDir)

	proj.PchHeader.Init(filename, false, true, g.Workspace)
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
			outputTarget = "../" + g.Workspace.PlatformName + "/"
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
			case "c++11":
				cppStd = "stdcpp14"
			case "c++14":
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
				wr.TagWithBodyBool("MultiProcessorCompilation", proj.Input.MultiThreadedBuild.Bool())

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

			for key, i := range config.VstudioClCompile.entries {
				wr.TagWithBody(key, config.VstudioClCompile.list[i])
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
				tmp := ""
				for _, p := range config.LinkLibs.FinalDict.list {
					tmp += p.Path + ";"
				}
				for _, p := range config.LinkFiles.FinalDict.list {
					tmp += relativeTo + p.Path + ";"
				}
				tmp += "%(" + optName + ")"
				wr.TagWithBody(optName, tmp)
			}

			if g.Workspace.MakeTarget.OSIsLinux() {
				wr.TagWithBody("VerboseOutput", "true")

				tmp := ""
				for _, e := range config.LinkFlags.FinalDict.list {
					tmp += " -Wl," + e.Path
				}
				wr.TagWithBody("AdditionalOptions", tmp)

				if config.IsDebug {
					wr.TagWithBody("DebuggerSymbolInformation", "true")
				} else {
					wr.TagWithBody("DebuggerSymbolInformation", "OmitAllSymbolInformation")
				}

			} else {
				wr.TagWithBody("SubSystem", "Console")
				wr.TagWithBody("GenerateDebugInformation", "true")

				tmp := ""
				for _, i := range config.LinkFlags.FinalDict.entries {
					tmp += " " + config.LinkFlags.FinalDict.list[i].Path
				}
				wr.TagWithBody("AdditionalOptions", tmp)

				if !config.IsDebug {
					wr.TagWithBodyBool("EnableCOMDATFolding", true)
					wr.TagWithBodyBool("OptimizeReferences", true)
				}
			}

			for key, i := range config.VstudioLink.entries {
				wr.TagWithBody(key, config.VstudioLink.list[i])
			}

			tag.Close()
		}

		if config.OutputTarget.Path != "" {
			tag := wr.TagScope("RemotePostBuildEvent")

			dir := "$(RemoteRootDir)/" + PathDirname(config.OutputTarget.Path)

			cmd := "mkdir -p \"" + dir + "\";"
			cmd += "cp -f \"$(RemoteProjectDir)/" + config.OutputTarget.Path + "\" \"$(RemoteRootDir)/" + config.OutputTarget.Path + "\""
			wr.TagWithBody("Command", cmd)

			tag.Close()
		}

		tag.Close()
	}
}

func (g *MsDevGenerator) genConfigOption(wr *XmlWriter, name string, value *ConfigEntryDict, relativeTo string) {
	option := ""
	for _, p := range value.list {
		option += relativeTo + p.Path + ";"
	}
	option += "%(" + name + ")"
	wr.TagWithBody(name, option)
}

/*

void Generator_vs2015::gen_workspace(ExtraWorkspace& ws) {
	Log::info("gen_workspace ", ws.name);

	ws.genData_vs2015.sln.set(g_ws->buildDir, ws.name, ".sln");

	String o;
	o.append(slnFileHeader());

	{
		o.append("\n# ---- projects ----\n");
		for (auto& proj : ws.projects) {
			o.append("Project(\"{8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942}\") = ",
					 "\"", proj->name, "\", \"", proj->name, ".vcxproj\", \"", proj->genData_vs2015.uuid, "\"\n");

			if (proj->_dependencies_inherit) {
				if (proj->type_is_exe_or_dll()) {
					o.append("\tProjectSection(ProjectDependencies) = postProject\n");
					for (auto& dp : proj->_dependencies_inherit) {
						o.append("\t\t", dp->genData_vs2015.uuid, " = ", dp->genData_vs2015.uuid, "\n");
					}
					o.append("\tEndProjectSection\n");
				}
			}
			o.append("EndProject\n");
		}
	}
	{
		auto* root = g_ws->projectGroups.root;
		o.append("\n# ---- Groups ----\n");
		for (auto& c : g_ws->projectGroups.dict) {
			if (&c == root) continue;
			genUuid(c.genData_vs2015.uuid);

			auto catName = Path::basename(c.path, true);
			o.append("Project(\"{2150E333-8FDC-42A3-9474-1A3956D46DE8}\") = \"",
						catName, "\", \"", catName, "\", \"", c.genData_vs2015.uuid, "\"\n");
			o.append("EndProject\n");
		}

		o.append("\n# ----  (ProjectGroups) -> parent ----\n");
		o.append("Global\n");
		o.append("\tGlobalSection(NestedProjects) = preSolution\n");
		for (auto& c : g_ws->projectGroups.dict) {
			if (c.parent && c.parent != root) {
				o.append("\t\t", c.genData_vs2015.uuid, " = ", c.parent->genData_vs2015.uuid, "\n");
			}

			for (auto& proj : c.projects) {
				if (ws.projects.indexOf(proj) < 0) continue;
				o.append("\t\t", proj->genData_vs2015.uuid, " = ", c.genData_vs2015.uuid, "\n");
			}
		}
		o.append("\tEndGlobalSection\n");
		o.append("EndGlobal\n");
	}

	FileUtil::writeTextFile(ws.genData_vs2015.sln, o);
}
*/

func (g *MsDevGenerator) genWorkspace(ws *ExtraWorkspace) {
	ws.GenDataVs2015.Sln = filepath.Join(g.Workspace.BuildDir, ws.Name, ".sln")

	sb := ""
	sb += slnFileHeader2022

	{
		sb += "\n# ---- projects ----\n"
		for _, proj := range ws.Projects {
			sb += "Project(\"{8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942}\") = "
			sb += "\"" + proj.Name + "\", \"" + proj.Name + ".vcxproj\", \"" + proj.GenDataMsDev.UUID.String() + "\"\n"

			if len(proj.DependenciesInherit) > 0 {
				if proj.TypeIsExeOrDll() {
					sb += "\tProjectSection(ProjectDependencies) = postProject\n"
					for _, dp := range proj.DependenciesInherit {
						sb += "\t\t" + dp.GenDataMsDev.UUID.String() + " = " + dp.GenDataMsDev.UUID.String() + "\n"
					}
					sb += "\tEndProjectSection\n"
				}
			}
			sb += "EndProject\n"
		}
	}
	{
		root := g.Workspace.ProjectGroups.Root
		sb += "\n# ---- Groups ----\n"
		for _, i := range g.Workspace.ProjectGroups.Groups {
			c := g.Workspace.ProjectGroups.Projects[i]
			if c == root {
				continue
			}
			c.GenDataVs2015.Uuid = GenerateUUID()

			catName := PathBasename(c.Path, true)
			sb += "Project(\"{2150E333-8FDC-42A3-9474-1A3956D46DE8}\") = \""
			sb += catName + "\", \"" + catName + "\", \"" + c.GenDataVs2015.Uuid.String() + "\"\n"
			sb += "EndProject\n"
		}

		sb += "\n# ----  (ProjectGroups) -> parent ----\n"
		sb += "Global\n"
		sb += "\tGlobalSection(NestedProjects) = preSolution\n"
		for _, i := range g.Workspace.ProjectGroups.Groups {
			c := g.Workspace.ProjectGroups.Projects[i]
			if c.Parent != nil && c.Parent != root {
				sb += "\t\t" + c.GenDataVs2015.Uuid.String() + " = " + c.Parent.GenDataVs2015.Uuid.String() + "\n"
			}

			for _, proj := range c.Projects {
				if proj == nil {
					continue
				}
				if ws.IndexOfProject(proj) < 0 {
					continue
				}
				sb += "\t\t" + proj.GenDataMsDev.UUID.String() + " = " + c.GenDataVs2015.Uuid.String() + "\n"
			}
		}
		sb += "\tEndGlobalSection\n"
		sb += "EndGlobal\n"
	}

	WriteTextFile(ws.GenDataVs2015.Sln, sb)
}

/*

void Generator_vs2015::gen_vcxproj_filters(Project& proj) {
	XmlWriter wr;

	{
		wr.writeHeader();
		auto tag = wr.tagScope("Project");
		wr.attr("ToolsVersion", "4.0");
		wr.attr("xmlns", "http://schemas.microsoft.com/developer/msbuild/2003");

		if (proj.pch_cpp.path()) {
			auto& f = proj.pch_cpp;
			proj.virtualFolders.add(g_ws->buildDir, f);
		}

		//------------
		{
			auto tag = wr.tagScope("ItemGroup");
			for (auto& d : proj.virtualFolders.dict) {
				if (d.path == ".") continue;
				auto tag = wr.tagScope("Filter");
				String winPath;
				Path::windowsPath(winPath, d.path);
				wr.attr("Include", winPath);
			}
		}

		{
			auto tag = wr.tagScope("ItemGroup");
			for (auto& pair : proj.virtualFolders.dict.pairs()) {
				for (auto f : pair.value->files) {
					if (!f) continue;

					auto type = StrView("ClInclude");
					switch (f->type()) {
						case FileType::ixx:	ax_fallthrough
						case FileType::c_source: ax_fallthrough
						case FileType::cpp_source: {
							type = "ClCompile";
						}break;
						default: break;
					}

					auto tag = wr.tagScope(type);
					wr.attr("Include", f->path());
					if (pair.key) {
						String winPath;
						Path::windowsPath(winPath, pair.key);
						wr.tagWithBody("Filter", winPath);
					}
				}
			}
		}
	}

	String filename;
	filename.append(g_ws->buildDir, proj.name, ".vcxproj.filters");
	FileUtil::writeTextFile(filename, wr.buffer());
}
*/

func (g *MsDevGenerator) genVcxprojFilters(proj *Project) {
	wr := NewXmlWriter()
	{
		wr.writeHeader()

		tag := wr.TagScope("Project")
		wr.Attr("ToolsVersion", "4.0")
		wr.Attr("xmlns", "http://schemas.microsoft.com/developer/msbuild/2003")

		if proj.PchCpp.Path != "" {
			proj.VirtualFolders.AddFile(g.Workspace.BuildDir, proj.PchCpp)
		}

		//------------
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
					if i.Path != "" {
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

	filename := filepath.Join(g.Workspace.BuildDir, proj.Name, ".vcxproj.filters")
	WriteTextFile(filename, wr.Buffer.String())
}
