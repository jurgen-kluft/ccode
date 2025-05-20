package denv

import (
	"path"
	"path/filepath"

	cutils "github.com/jurgen-kluft/ccode/cutils"
)

type MsDevGenerator struct {
	Workspace  *Workspace
	VcxProjCpu string
}

func NewMsDevGenerator(ws *Workspace) *MsDevGenerator {
	g := &MsDevGenerator{
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
	if ws.BuildTarget == nil {
		ws.BuildTarget = NewDefaultMakeTarget(ws.Config.Dev, ws.Config.OS, ws.Config.Arch)
	}

	if ws.BuildTarget.ArchIsX64() {
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
	if proj.Workspace.Config.MsDev.PlatformToolset != "" {
		return proj.Workspace.Config.MsDev.PlatformToolset
	}
	return g.Workspace.Config.MsDev.PlatformToolset
}

func (g *MsDevGenerator) TargetPlatformVersion(proj *Project) string {
	if proj.Workspace.Config.MsDev.WindowsTargetPlatformVersion != "" {
		return proj.Workspace.Config.MsDev.WindowsTargetPlatformVersion
	}
	return g.Workspace.Config.MsDev.WindowsTargetPlatformVersion
}

func (g *MsDevGenerator) genProject(proj *Project) {
	projectFilepath := filepath.Join(g.Workspace.GenerateAbsPath, proj.ProjectFilename+".vcxproj")

	wr := cutils.NewXmlWriter()
	{
		wr.WriteHeader()
		tag := wr.TagScope("Project")

		wr.Attr("DefaultTargets", "Build")
		wr.Attr("ToolsVersion", proj.Workspace.Config.MsDev.ProjectTools)
		wr.Attr("xmlns", "http://schemas.microsoft.com/developer/msbuild/2003")

		{
			tag := wr.TagScope("ItemGroup")
			wr.Attr("Label", "ProjectConfigurations")

			for _, config := range proj.Resolved.Configs.Values {
				tag := wr.TagScope("ProjectConfiguration")
				wr.Attr("Include", config.String()+"|"+g.VcxProjCpu)

				wr.TagWithBody("Configuration", config.String())
				wr.TagWithBody("Platform", g.VcxProjCpu)
				tag.Close()
			}
			tag.Close()
		}
		{
			tag := wr.TagScope("PropertyGroup")
			wr.Attr("Label", "Globals")
			wr.TagWithBody("ProjectGuid", proj.Resolved.GenDataMsDev.UUID.ForVisualStudio())
			wr.TagWithBody("Keyword", "Win32Proj")
			wr.TagWithBody("RootNamespace", proj.Name)
			wr.TagWithBody("WindowsTargetPlatformVersion", proj.Workspace.Config.MsDev.WindowsTargetPlatformVersion)
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
			if proj.TypeIsLib() {
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

			if g.Workspace.BuildTarget.OSIsLinux() {
				wr.TagWithBody("PlatformToolset", "Remote_GCC_1_0")
				if g.Workspace.BuildTarget.CompilerIsGcc() {
					wr.TagWithBody("RemoteCCompileToolExe", "gcc")
					wr.TagWithBody("RemoteCppCompileToolExe", "g++")
				} else if g.Workspace.BuildTarget.CompilerIsClang() {
					wr.TagWithBody("RemoteCCompileToolExe", "clang")
					wr.TagWithBody("RemoteCppCompileToolExe", "clang++")
				} else {
					panic("Unsupported compiler")
				}

				relBuildDir := cutils.PathGetRelativeTo(g.Workspace.GenerateAbsPath, g.Workspace.WorkspaceAbsPath)

				remoteRootDir := "_vsForLinux/" + g.Workspace.BuildTarget.OSAsString() + "/" + relBuildDir
				wr.TagWithBody("RemoteRootDir", remoteRootDir)

				projectDir := "$(RemoteRootDir)/../ProjectDir_" + proj.Name
				wr.TagWithBody("RemoteProjectDir", projectDir)

				fileToCopy := ""
				for _, f := range proj.SrcFiles.Values {
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

		if g.Workspace.BuildTarget.OSIsLinux() {
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
		for _, config := range proj.Resolved.Configs.Values {
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

func (g *MsDevGenerator) genProjectFiles(wr *cutils.XmlWriter, proj *Project) {
	tag := wr.TagScope("ItemGroup")

	g.genProjectPch(wr, proj)

	for _, f := range proj.SrcFiles.Values {
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
		path := proj.SrcFiles.GetRelativePath(f, proj.Workspace.GenerateAbsPath)
		wr.Attr("Include", path)
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

func (g *MsDevGenerator) genProjectPch(wr *cutils.XmlWriter, proj *Project) {
	if proj.Resolved.PchHeader == nil {
		return
	}

	pchExt := ".cpp"

	filename := proj.Resolved.GeneratedFilesDir + proj.Name + "-precompiledHeader" + pchExt
	tmp := cutils.PathGetRelativeTo(proj.Resolved.PchHeader.Path, proj.Resolved.GeneratedFilesDir)

	proj.Resolved.PchHeader.Init(filename, true)
	code := "//-- Auto Generated File for Visual C++ precompiled header\n"
	code += "#include \"" + tmp + "\"\n"

	cutils.WriteTextToFile(filename, code)

	tag := wr.TagScope("ClCompile")
	{
		wr.Attr("Include", proj.Resolved.PchHeader.Path)
		wr.TagWithBody("PrecompiledHeader", "Create")
	}
	tag.Close()
}

func (g *MsDevGenerator) genProjectConfig(wr *cutils.XmlWriter, proj *Project, config *Config) {
	cond := "'$(Configuration)|$(Platform)'=='" + config.String() + "|" + g.VcxProjCpu + "'"
	{
		tag := wr.TagScope("PropertyGroup")
		wr.Attr("Condition", cond)

		outDir := cutils.PathDirname(config.Resolved.OutputTarget.Path)
		if g.Workspace.BuildTarget.OSIsLinux() {
			outDir = filepath.Join(outDir, g.Workspace.BuildTarget.OSAsString())
		}

		intDir := filepath.Join(g.Workspace.GenerateAbsPath, "obj", proj.Name, config.String()+"_"+g.Workspace.BuildTarget.ArchAsString()+"_"+g.Workspace.Config.MsDev.PlatformToolset)
		targetName := cutils.PathFilename(config.Resolved.OutputTarget.Path, false)
		targetExt := cutils.PathFileExtension(config.Resolved.OutputTarget.Path)

		// Visual Studio wants the following paths to end with a backslash
		wr.TagWithBody("OutDir", cutils.PathNormalize(cutils.PathGetRelativeTo(outDir, proj.GenerateAbsPath))+cutils.PathSlash())
		wr.TagWithBody("IntDir", cutils.PathNormalize(cutils.PathGetRelativeTo(intDir, proj.GenerateAbsPath))+cutils.PathSlash())
		wr.TagWithBody("TargetName", targetName)
		if targetExt != "" {
			wr.TagWithBody("TargetExt", targetExt)
		}

		tag.Close()
	}
	{
		tag := wr.TagScope("ItemDefinitionGroup")
		wr.Attr("Condition", cond)
		{
			tag := wr.TagScope("ClCompile")

			cppStd := "stdcpp14"
			switch g.Workspace.Config.CppStd {
			case CppStd11, CppStd14:
				cppStd = "stdcpp14"
			case CppStd17:
				cppStd = "stdcpp17"
			case CppStdLatest:
				cppStd = "stdcpplatest"
			}
			wr.TagWithBody("LanguageStandard", cppStd)

			cppAdvanced := g.Workspace.Config.CppAdvanced.VisualStudio()
			if cppAdvanced != "" {
				wr.TagWithBody("EnableEnhancedInstructionSet", cppAdvanced)
			}

			if g.Workspace.BuildTarget.OSIsLinux() {
				wr.TagWithBody("Verbose", "true")
			} else {
				wr.TagWithBody("SDLCheck", "true")
				wr.TagWithBodyBool("MultiProcessorCompilation", proj.Settings.MultiThreadedBuild.Bool())
			}

			if proj.Resolved.PchHeader != nil {
				wr.TagWithBody("PrecompiledHeader", "Use")
				pch := "$(ProjectDir)" + proj.Resolved.PchHeader.Path
				wr.TagWithBody("PrecompiledHeaderFile", pch)
				wr.TagWithBody("ForcedIncludeFiles", pch)
			} else {
				wr.TagWithBody("PrecompiledHeader", "NotUsing")
			}

			g.genConfigOptionFromKeyValueDict(wr, proj, "DisableSpecificWarnings", config.DisableWarning, false)
			g.genConfigOptionFromKeyValueDict(wr, proj, "PreprocessorDefinitions", config.CppDefines, false)
			g.genConfigOptionWithModifier(wr, "AdditionalIncludeDirectories", config.IncludeDirs, func(_root string, _path string) string {
				relpath := cutils.PathGetRelativeTo(path.Join(_root, _path), proj.Workspace.GenerateAbsPath)
				return relpath
			})

			for i, value := range config.VisualStudioClCompile.Values {
				wr.TagWithBody(config.VisualStudioClCompile.Keys[i], value)
			}

			tag.Close()
		}
		{
			tag := wr.TagScope("Link")

			if proj.TypeIsDll() {
				wr.TagWithBody("ImportLibrary", config.Resolved.OutputLib.Path)
			}

			// Based on the dependencies of this project we need to collect:
			// - Library directories
			// - Library files
			// - And also include the dependency project output as a file to link with

			linkDirs, linkFiles, linkLibs := proj.BuildLibraryInformation(DevVisualStudio, config, proj.Workspace.GenerateAbsPath)

			if proj.TypeIsExeOrDll() {
				g.genConfigOptionFromValueSet(wr, proj, "AdditionalLibraryDirectories", linkDirs, true)

				optName := "AdditionalDependencies"
				relativeTo := ""
				if g.Workspace.BuildTarget.OSIsLinux() {
					relativeTo = "$(RemoteRootDir)/"
				}
				tmp := linkLibs.Concatenated("", ";", func(s string) string { return s })
				tmp += linkFiles.Concatenated(relativeTo, ";", func(value string) string {
					path := cutils.PathGetRelativeTo(value, proj.Workspace.GenerateAbsPath)
					return path
				})
				tmp += "%(" + optName + ")"
				wr.TagWithBody(optName, tmp)
			}

			if g.Workspace.BuildTarget.OSIsLinux() {
				wr.TagWithBody("VerboseOutput", "true")

				tmp := config.LinkFlags.Concatenated(" -Wl,", "", func(string, s string) string { return s })
				wr.TagWithBody("AdditionalOptions", tmp)

				if config.Type.IsDebug() {
					wr.TagWithBody("DebuggerSymbolInformation", "true")
				} else {
					wr.TagWithBody("DebuggerSymbolInformation", "OmitAllSymbolInformation")
				}

			} else {
				wr.TagWithBody("SubSystem", "Console")
				wr.TagWithBody("GenerateDebugInformation", "true")

				tmp := config.LinkFlags.Concatenated(" ", "", func(string, s string) string { return s })
				if len(tmp) > 0 {
					wr.TagWithBody("AdditionalOptions", tmp)
				}

				if !config.Type.IsDebug() {
					wr.TagWithBodyBool("EnableCOMDATFolding", true)
					wr.TagWithBodyBool("OptimizeReferences", true)
				}
			}

			for i, value := range config.VisualStudioLink.Values {
				wr.TagWithBody(config.VisualStudioLink.Keys[i], value)
			}

			tag.Close()
		}

		if config.Resolved.OutputTarget.Path != "" {
			tag := wr.TagScope("RemotePostBuildEvent")
			{
				cmd := "mkdir -p \"" + "$(RemoteRootDir)/" + cutils.PathDirname(config.Resolved.OutputTarget.Path) + "\";"
				cmd += "cp -f \"$(RemoteProjectDir)/" + config.Resolved.OutputTarget.Path + "\" \"$(RemoteRootDir)/" + config.Resolved.OutputTarget.Path + "\""
				wr.TagWithBody("Command", cmd)
			}
			tag.Close()
		}

		tag.Close()
	}
}

func (g *MsDevGenerator) genConfigOptionFromKeyValueDict(wr *cutils.XmlWriter, proj *Project, name string, kv *KeyValueDict, treatAsPath bool) {
	option := kv.Concatenated("", ";", func(k string, v string) string {
		if treatAsPath {
			path := cutils.PathGetRelativeTo(v, proj.Workspace.GenerateAbsPath)
			return path
		}
		return v
	})
	option += "%(" + name + ")"
	wr.TagWithBody(name, option)
}

func (g *MsDevGenerator) genConfigOptionFromValueSet(wr *cutils.XmlWriter, proj *Project, name string, value *DevValueSet, treatAsPath bool) {
	option := value.Concatenated("", ";", func(v string) string {
		if treatAsPath {
			path := cutils.PathGetRelativeTo(v, proj.Workspace.GenerateAbsPath)
			return path
		}
		return v
	})
	option += "%(" + name + ")"
	wr.TagWithBody(name, option)
}

func (g *MsDevGenerator) genConfigOptionWithModifier(wr *cutils.XmlWriter, name string, value *PinnedPathSet, modifier func(string, string) string) {
	option := value.Concatenated("", ";", modifier)
	option += "%(" + name + ")"
	wr.TagWithBody(name, option)
}

func (g *MsDevGenerator) writeSolutionProject(proj *Project, sb *cutils.LineWriter) {
	sb.Write("Project(\"{8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942}\") = ")
	sb.WriteLine("\"" + proj.Name + "\", \"" + proj.Name + ".vcxproj\", \"" + proj.Resolved.GenDataMsDev.UUID.ForVisualStudio() + "\"")

	if len(proj.Dependencies.Values) > 0 {
		{
			sb.WriteLine("\tProjectSection(ProjectDependencies) = postProject")
			for _, dp := range proj.Dependencies.Values {
				sb.WriteLine("\t\t" + dp.Resolved.GenDataMsDev.UUID.ForVisualStudio() + " = " + dp.Resolved.GenDataMsDev.UUID.ForVisualStudio())
			}
			sb.WriteLine("\tEndProjectSection")
		}
	}
	sb.WriteLine("EndProject")
}

func (g *MsDevGenerator) genWorkspace(ws *ExtraWorkspace) {
	visualStudioSolutionFilepath := filepath.Join(g.Workspace.GenerateAbsPath, ws.Workspace.WorkspaceName+".sln")

	sb := cutils.NewLineWriter(cutils.IndentModeSpaces)

	sb.WriteManyLines(ws.MsDev.SlnHeader)

	{
		sb.WriteLines("", "# ---- projects ----", "")

		g.writeSolutionProject(ws.Workspace.StartupProject, sb)
		for _, proj := range ws.ProjectList.Values {
			if proj == ws.Workspace.StartupProject {
				continue
			}
			g.writeSolutionProject(proj, sb)
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
			c.MsDev.UUID = cutils.GenerateUUID()

			catName := cutils.PathFilename(c.Path, true)
			// sb += "Project(\"{2150E333-8FDC-42A3-9474-1A3956D46DE8}\") = \""
			// sb += catName + "\", \"" + catName + "\", \"" + c.MsDev.UUID.String() + "\"\n"
			// sb += "EndProject\n"
			sb.Write("Project(\"{2150E333-8FDC-42A3-9474-1A3956D46DE8}\") = \"")
			sb.Write(catName)
			sb.Write("\", \"")
			sb.Write(catName)
			sb.Write("\", \"")
			sb.Write(c.MsDev.UUID.ForVisualStudio())
			sb.WriteLine("\"")
			sb.WriteLine("EndProject")
		}

		sb.WriteLines("", "# ----  (ProjectGroups) -> parent ----", "")
		sb.WriteLine("Global")
		sb.WriteLine("\tGlobalSection(NestedProjects) = preSolution")

		for _, c := range g.Workspace.ProjectGroups.Values {
			if c.Parent != nil && c.Parent != root {
				sb.Write("\t\t")
				sb.Write(c.MsDev.UUID.ForVisualStudio())
				sb.Write(" = ")
				sb.WriteLine(c.Parent.MsDev.UUID.ForVisualStudio())
			}

			for _, proj := range c.Projects {
				if proj == nil {
					continue
				}
				if !ws.HasProject(proj) {
					continue
				}
				sb.Write("\t\t")
				sb.Write(proj.Resolved.GenDataMsDev.UUID.ForVisualStudio())
				sb.Write(" = ")
				sb.WriteLine(c.MsDev.UUID.ForVisualStudio())
			}
		}
		sb.WriteLine("\tEndGlobalSection")
		sb.WriteLine("EndGlobal")
	}

	sb.WriteToFile(visualStudioSolutionFilepath)
}

func (g *MsDevGenerator) genProjectFilters(proj *Project) {
	projectFiltersFilepath := filepath.Join(g.Workspace.GenerateAbsPath, proj.ProjectFilename+".vcxproj.filters")

	wr := cutils.NewXmlWriter()
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
				if len(i.Path) == 0 || i.Path == "." {
					continue
				}
				tag := wr.TagScope("Filter")
				winPath := cutils.PathWindowsPath(i.Path)
				wr.Attr("Include", winPath)
				tag.Close()
			}
			tag.Close()
		}

		{
			tag := wr.TagScope("ItemGroup")
			for _, vf := range proj.VirtualFolders.Folders {
				for _, f := range vf.Files {
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
					relPath := cutils.PathGetRelativeTo(filepath.Join(proj.VirtualFolders.DiskPath, f.Path), proj.Workspace.GenerateAbsPath)
					wr.Attr("Include", relPath)
					if len(vf.Path) > 0 {
						winPath := cutils.PathWindowsPath(vf.Path)
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
