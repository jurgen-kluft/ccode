package vs2015

import (
	"fmt"

	"github.com/jurgen-kluft/xcode/denv"
	"github.com/jurgen-kluft/xcode/uid"
	"github.com/jurgen-kluft/xcode/util"
	"github.com/jurgen-kluft/xcode/vars"

	"path"
	"path/filepath"
)

// IsVisualStudio returns true if @ide is matching a Visual Studio identifier
func IsVisualStudio(ide string) bool {
	vs := GetVisualStudio(ide)
	return vs != -1
}

// GetVisualStudio returns the IDE type from a given string
func GetVisualStudio(ide string) denv.IDE {
	if ide == "VS2015" {
		return denv.VS2015
	} else if ide == "VS2013" {
		return denv.VS2013
	} else if ide == "VS2012" {
		return denv.VS2012
	}
	return -1
}

// AddProjectVariables adds variables from the Project information
//   Example for 'xhash' project with 'xbase' as a dependency:
//   - xhash:GUID
//   - xhash:ROOT_DIR
//   - xhash:LIBRARY_DIRS[Platform][Config]
//   - xhash:INCLUDE_DIRS[Platform][Config]
//   - xhash:LINK_WITH[Platform][Config]
//   - xhash:DEFINES[Platform][Config], xhash:INCLUDE_DIRS[Platform][Config], xhash:LINK_WITH[Platform][Config]
//   - xhash:OPTIMIZATION[Platform][Config]
//   - xhash:DEBUG_INFO[Platform][Config]
//   - xhash:TOOL_SET[Platform]
//
func AddProjectVariables(p *denv.Project, v vars.Variables) {
	v.AddVar(p.Name+":GUID", p.GUID)
	v.AddVar(p.Name+":ROOT_DIR", p.Path)
	for _, platform := range p.Platforms {
		for _, config := range p.Configs {
			v.AddVar(fmt.Sprintf("%s:LIBRARY_DIRS[%s][%s]", p.Name, platform, config.Name), util.Join(config.LibraryDirs, ";"))
			v.AddVar(fmt.Sprintf("%s:INCLUDE_DIRS[%s][%s]", p.Name, platform, config.Name), util.Join(config.IncludeDirs, ";"))
			v.AddVar(fmt.Sprintf("%s:LINK_WITH[%s][%s]", p.Name, platform, config.Name), util.Join(config.LinkWith, ";"))
		}
	}
}

// SetProjectPath will set correct paths for the dependencies
func SetProjectPath(p *denv.Project) {
	p.Path = ""
	for _, d := range p.Dependencies {
		d.Path = path.Join("vendor", d.Path)
	}
}

// CPPprojectID is a GUID that is defining a C++ project in Visual Studio
var CPPprojectID = "8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942"

// GenerateVisualStudio2015Project generates a Visual Studio 2015 project (.vcxproj) file
func GenerateVisualStudio2015Project(prj *denv.Project, vars vars.Variables, replacer vars.Replacer, writer denv.ProjectWriter) {

	writer.WriteLn(`<?xml version="1.0" encoding="utf-8"?>`)
	writer.WriteLn(`<Project DefaultTargets="Build" ToolsVersion="14.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">`)

	writer.WriteLn(`+<ItemGroup Label="ProjectConfigurations">`)
	for _, platform := range prj.Platforms {
		for _, config := range prj.Configs {
			projconfig := []string{}
			projconfig = append(projconfig, `+<ProjectConfiguration Include="${CONFIG}|${PLATFORM}">`)
			projconfig = append(projconfig, `++<Configuration>${CONFIG}</Configuration>`)
			projconfig = append(projconfig, `++<Platform>${PLATFORM}</Platform>`)
			projconfig = append(projconfig, `+</ProjectConfiguration>`)

			replacer.ReplaceInLines("${PLATFORM}", platform, projconfig)
			replacer.ReplaceInLines("${CONFIG}", config.Name, projconfig)
			writer.WriteLns(projconfig)
		}
	}
	writer.WriteLn(`+</ItemGroup>`)

	for _, platform := range prj.Platforms {
		toolsets := []string{}
		toolsets = append(toolsets, `+<PropertyGroup Condition="'$(Platform)'=='${PLATFORM}'" Label="Configuration">`)
		toolsets = append(toolsets, `++<PlatformToolset>${TOOLSET[${PLATFORM}]}</PlatformToolset>`)
		toolsets = append(toolsets, `+</PropertyGroup>`)

		replacer.ReplaceInLines("${PLATFORM}", platform, toolsets)
		vars.ReplaceInLines(replacer, toolsets)
		writer.WriteLns(toolsets)
	}

	globals := []string{}
	globals = append(globals, `+<PropertyGroup Label="Globals">`)
	globals = append(globals, `++<ProjectGuid>${${Name}:GUID}</ProjectGuid>`)
	globals = append(globals, `++<PackageSignature>$(Configuration)_$(Platform)_$(PlatformToolset)</PackageSignature>`)
	globals = append(globals, `+</PropertyGroup>`)

	replacer.ReplaceInLines("${Name}", prj.Name, globals)
	vars.ReplaceInLines(replacer, globals)
	writer.WriteLns(globals)

	for _, depproject := range append(prj.Dependencies, prj) {
		projdirs := []string{}
		projdirs = append(projdirs, `+<PropertyGroup Label="Directories">`)
		projdirs = append(projdirs, `++<__${Name}RootDir__>${${Name}:RootDir}</__${Name}RootDir__>`)
		projdirs = append(projdirs, `++<__${Name}LibraryDir__>${${Name}:LibraryDir}target\${Name}\outdir\${Name}_$(PackageSignature)\</__${Name}LibraryDir__>`)
		projdirs = append(projdirs, `++<__${Name}IncludeDir__>${${Name}:IncludeDir}source\main\include\</__${Name}IncludeDir__>`)
		projdirs = append(projdirs, `++<__${Name}TestIncludeDir__>${${Name}:TestIncludeDir}source+est\include\</__${Name}TestIncludeDir__>`)
		projdirs = append(projdirs, `+</PropertyGroup>`)
		replacer.ReplaceInLines("${Name}", depproject.Name, projdirs)
		vars.ReplaceInLines(replacer, projdirs)
		writer.WriteLns(projdirs)
	}

	writer.WriteLn(`+<ImportGroup Label="TargetDirs"/>`)
	writer.WriteLn(`+<Import Project="$(VCTargetsPath)\Microsoft.Cpp.Default.props"/>`)

	for _, platform := range prj.Platforms {
		for _, config := range prj.Configs {
			configuration := []string{}
			configuration = append(configuration, `+<PropertyGroup Condition="'$(Configuration)|$(Platform)'=='${CONFIG}|${PLATFORM}'" Label="Configuration">`)
			configuration = append(configuration, `++<ConfigurationType>StaticLibrary</ConfigurationType>`)
			configuration = append(configuration, `++<UseDebugLibraries>${USE_DEBUG_LIBS}</UseDebugLibraries>`)
			configuration = append(configuration, `++<CLRSupport>false</CLRSupport>`)
			configuration = append(configuration, `++<CharacterSet>NotSet</CharacterSet>`)
			configuration = append(configuration, `+</PropertyGroup>`)
			usedebuglibs, err := vars.GetVar(fmt.Sprintf("%s:USE_DEBUG_LIBS[%s][%s]", prj.Name, platform, config))
			if err == nil {
				replacer.ReplaceInLines("${USE_DEBUG_LIBS}", usedebuglibs, configuration)
			}

			replacer.ReplaceInLines("${PLATFORM}", platform, configuration)
			replacer.ReplaceInLines("${CONFIG}", config.Name, configuration)
			vars.ReplaceInLines(replacer, configuration)
			writer.WriteLns(configuration)
		}
	}

	writer.WriteLn(`<Import Project="$(VCTargetsPath)\Microsoft.Cpp.props"/>`)

	for _, platform := range prj.Platforms {
		userprops := []string{}
		userprops = append(userprops, `+<ImportGroup Condition="'$(Platform)'=='${PLATFORM}'" Label="PropertySheets">`)
		userprops = append(userprops, `++<Import Condition="exists('$(UserRootDir)\Microsoft.Cpp.$(Platform).user.props')" Label="LocalAppDataPlatform" Project="$(UserRootDir)\Microsoft.Cpp.$(Platform).user.props"/>`)
		userprops = append(userprops, `+</ImportGroup>`)
		replacer.ReplaceInLines("${PLATFORM}", platform, userprops)
		writer.WriteLns(userprops)
	}

	for _, platform := range prj.Platforms {
		platformprops := []string{}
		platformprops = append(platformprops, `+<PropertyGroup Condition="'$(Platform)'=='${PLATFORM}'">`)
		platformprops = append(platformprops, `++<LinkIncremental>true</LinkIncremental>`)
		platformprops = append(platformprops, `++<OutDir>$(SolutionDir)target\$(ProjectName)_$(Config)_$(Platform)_$(ToolSet)\</OutDir>`)
		platformprops = append(platformprops, `++<IntDir>$(SolutionDir)target\$(ProjectName)_$(Config)_$(Platform)\</IntDir>`)
		platformprops = append(platformprops, `++<TargetName>$(ProjectName)_$(Config)_$(Platform)_$(ToolSet)</TargetName>`)
		platformprops = append(platformprops, `++<ExtensionsToDeleteOnClean>*.obj%3b*.d%3b*.map%3b*.lst%3b*.pch%3b$(TargetPath)</ExtensionsToDeleteOnClean>`)
		platformprops = append(platformprops, `++<GenerateManifest>false</GenerateManifest>`)
		platformprops = append(platformprops, `+</PropertyGroup>`)
		replacer.ReplaceInLines("${PLATFORM}", platform, platformprops)
		writer.WriteLns(platformprops)
	}

	includedirs := []string{}
	for _, depproject := range prj.Dependencies {
		depincludedirs := "${${Name}:IncludeDir}"
		depincludedirs = replacer.ReplaceInLine("${Name}", depproject.Name, depincludedirs)
		depincludedirs = vars.ReplaceInLine(replacer, depincludedirs)
		includedirs = append(includedirs, depincludedirs)
	}

	for _, platform := range prj.Platforms {
		for _, config := range prj.Configs {
			compileandlink := []string{}
			compileandlink = append(compileandlink, `+<ItemDefinitionGroup Condition="'$(Configuration)|$(Platform)'=='${CONFIG}|${PLATFORM}'">`)
			compileandlink = append(compileandlink, `++<ClCompile>`)
			compileandlink = append(compileandlink, `+++<PreprocessorDefinitions>${DEFINES}%(PreprocessorDefinitions)</PreprocessorDefinitions>`)
			compileandlink = append(compileandlink, `+++<AdditionalIncludeDirectories>${INCLUDE_DIRS}%(AdditionalIncludeDirectories)</AdditionalIncludeDirectories>`)
			compileandlink = append(compileandlink, `+++<WarningLevel>Level3</WarningLevel>`)
			compileandlink = append(compileandlink, `+++<Optimization>${OPTIMIZATION}</Optimization>`)
			compileandlink = append(compileandlink, `+++<PrecompiledHeader>NotUsing</PrecompiledHeader>`)
			compileandlink = append(compileandlink, `+++<ExceptionHandling>false</ExceptionHandling>`)
			compileandlink = append(compileandlink, `++</ClCompile>`)
			compileandlink = append(compileandlink, `++<Link>`)
			compileandlink = append(compileandlink, `+++<GenerateDebugInformation>${DEBUG_INFO}</GenerateDebugInformation>`)
			compileandlink = append(compileandlink, `+++<AdditionalDependencies>${LINK_WITH}%(AdditionalDependencies)</AdditionalDependencies>`)
			compileandlink = append(compileandlink, `++</Link>`)
			compileandlink = append(compileandlink, `++<Lib>`)
			compileandlink = append(compileandlink, `+++<OutputFile>$(OutDir)\$(TargetName)$(TargetExt)</OutputFile>`)
			compileandlink = append(compileandlink, `++</Lib>`)
			compileandlink = append(compileandlink, `+</ItemDefinitionGroup>`)

			varkeys := []string{"DEFINES", "INCLUDE_DIRS", "LINK_WITH"}
			for _, varkey := range varkeys {
				varkeystr := fmt.Sprintf("${%s}", varkey)
				for _, depprojectname := range append(prj.Dependencies, prj) {
					varvalue, err := vars.GetVar(fmt.Sprintf("%s:%s[%s][%s]", depprojectname, varkey, config, platform))
					if err == nil {
						replacer.InsertInLines(varkeystr, varvalue, compileandlink)
					}
				}
				replacer.ReplaceInLines(varkeystr, "", compileandlink)
			}

			optimization, err := vars.GetVar(fmt.Sprintf("%s:OPTIMIZATION[%s][%s]", prj.Name, platform, config))
			if err == nil {
				replacer.ReplaceInLines("${OPTIMIZATION}", optimization, compileandlink)
			}

			debuginfo, err := vars.GetVar(fmt.Sprintf("%s:DEBUG_INFO[%s][%s]", prj.Name, platform, config))
			if err == nil {
				replacer.ReplaceInLines("${DEBUG_INFO}", debuginfo, compileandlink)
			}

			replacer.ReplaceInLines("${PLATFORM}", platform, compileandlink)
			replacer.ReplaceInLines("${CONFIG}", config.Name, compileandlink)
			vars.ReplaceInLines(replacer, compileandlink)
			writer.WriteLns(compileandlink)
		}
	}

	writer.WriteLn("+<ItemGroup>")
	for _, srcfile := range prj.SrcFiles.Files {
		clcompile := "++<ClCompile Include=\"${FILE}\"/>"
		clcompile = replacer.ReplaceInLine("${FILE}", srcfile, clcompile)
		writer.WriteLn(clcompile)
	}
	writer.WriteLn("+</ItemGroup>")

	if len(prj.HdrFiles.Files) > 0 {
		writer.WriteLn("+<ItemGroup>")
		for _, hdrfile := range prj.HdrFiles.Files {
			clinclude := "++<ClInclude Include=\"${FILE}\"/>"
			clinclude = replacer.ReplaceInLine("${FILE}", hdrfile, clinclude)
			writer.WriteLn(clinclude)
		}
		writer.WriteLn("+</ItemGroup>")
	}

	//writer.WriteLn("+<ItemGroup>")
	//writer.WriteLn("++<None Include=\"\"/>")
	//writer.WriteLn("+</ItemGroup>")

	writer.WriteLn(`+<Import Condition="'$(ConfigurationType)' == 'Makefile' and Exists('$(VCTargetsPath)\Platforms\$(Platform)\SCE.Makefile.$(Platform).targets')" Project="$(VCTargetsPath)\Platforms\$(Platform)\SCE.Makefile.$(Platform).targets"/>`)
	writer.WriteLn(`+<Import Project="$(VCTargetsPath)\Microsoft.Cpp.targets"/>`)
	writer.WriteLn(`+<ImportGroup Label="ExtensionTargets"/>`)

	writer.WriteLn(`</Project>`)
}

// GenerateVisualStudio2015ProjectFilters generates a Visual Studio 2015 project filters (.vcxproj.filters) file
func GenerateVisualStudio2015ProjectFilters(prj *denv.Project, writer denv.ProjectWriter) {
	writer.WriteLn(`<?xml version="1.0" encoding="utf-8"?>`)
	writer.WriteLn(`<Project ToolsVersion="4.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">`)

	includefilters := make(map[string]string)
	includes := make(map[string]string)
	for _, hdrfile := range prj.HdrFiles.Files {
		dirpath := filepath.Dir(hdrfile)
		guid := uid.GetGUID(dirpath)
		includefilters[dirpath] = guid
		includes[hdrfile] = dirpath

		// We need to add every 'depth' of the path
		for true {
			//fmt.Printf("dir:\"%s\" --> \"%s\" | ", dirpath, filepath.Dir(dirpath))
			dirpath = filepath.Dir(dirpath)
			if dirpath == "." {
				break
			}
			guid = uid.GetGUID(dirpath)
			includefilters[dirpath] = guid
		}
		//fmt.Println("")
	}

	cppfilters := make(map[string]string)
	cpp := make(map[string]string)
	for _, srcfile := range prj.SrcFiles.Files {
		dirpath := filepath.Dir(srcfile)
		guid := uid.GetGUID(dirpath)
		cppfilters[dirpath] = guid
		cpp[srcfile] = dirpath

		// We need to add every 'depth' of the path
		for true {
			//fmt.Printf("dir:\"%s\" --> \"%s\" | ", dirpath, filepath.Dir(dirpath))
			dirpath = filepath.Dir(dirpath)
			if dirpath == "." {
				break
			}
			guid = uid.GetGUID(dirpath)
			cppfilters[dirpath] = guid
		}
		//fmt.Println("")
	}

	writer.WriteLn("+<ItemGroup>")
	for k, v := range includefilters {
		writer.WriteLn(fmt.Sprintf("++<Filter Include=\"%s\">", k))
		writer.WriteLn(fmt.Sprintf("+++<UniqueIdentifier>{%s}</UniqueIdentifier>", v))
		writer.WriteLn("++</Filter>")
	}
	for k := range cppfilters {
		writer.WriteLn(fmt.Sprintf("++<Filter Include=\"%s\">", k))
		writer.WriteLn(fmt.Sprintf("+++<UniqueIdentifier>{%s}</UniqueIdentifier>", uid.GetGUID(k)))
		writer.WriteLn("++</Filter>")
	}
	writer.WriteLn("+</ItemGroup>")

	writer.WriteLn("+<ItemGroup>")
	for k, v := range includes {
		writer.WriteLn(fmt.Sprintf("++<ClInclude Include=\"%s\">", k))
		writer.WriteLn(fmt.Sprintf("+++<Filter>%s</Filter>", v))
		writer.WriteLn("++</ClInclude>")
	}
	writer.WriteLn("+</ItemGroup>")

	writer.WriteLn("+<ItemGroup>")
	for k, v := range cpp {
		writer.WriteLn(fmt.Sprintf("++<ClCompile Include=\"%s\">", k))
		writer.WriteLn(fmt.Sprintf("+++<Filter>%s</Filter>", v))
		writer.WriteLn("++</ClCompile>")
	}
	writer.WriteLn("+</ItemGroup>")

	writer.WriteLn("</Project>")
}

// GenerateVisualStudio2015Solution generates a Visual Studio 2015 solution (.sln) file
func GenerateVisualStudio2015Solution(sln denv.Solution, writer denv.ProjectWriter) {
	writer.WriteLn("Microsoft Visual Studio Solution File, Format Version 12.00")
	writer.WriteLn("# Visual Studio 14")
	writer.WriteLn("VisualStudioVersion = 14.0.24720.0")
	writer.WriteLn("MinimumVisualStudioVersion = 10.0.40219.1")

	// Write Projects and their dependency information
	//
	//          Project("{8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942}") = "xhash", "source\main\cpp\xhash.vcxproj", "{04AB9C6F-0B84-4111-A772-53C03F5CB3C2}"
	//          	ProjectSection(ProjectDependencies) = postProject
	//          		{B83DA73D-6E7B-458D-A6C7-87013421D360} = {B83DA73D-6E7B-458D-A6C7-87013421D360}
	//          	EndProjectSection
	//          EndProject
	//
	for _, prj := range sln.Projects {
		projectbeginfmt := "Project(\"{%s}\") = \"%s\", \"%s\\%s.vcxproj\", \"{%s}\""
		projectbegin := fmt.Sprintf(projectbeginfmt, CPPprojectID, prj.Name, prj.Path, prj.Name, prj.GUID)
		writer.WriteLn(projectbegin)
		if len(prj.Dependencies) > 0 {
			projectsessionbegin := "+ProjectSection(ProjectDependencies) = postProject"
			writer.WriteLn(projectsessionbegin)
			for _, dep := range prj.Dependencies {
				projectdep := fmt.Sprintf("++{%s} = {%s}", dep.GUID, dep.GUID)
				writer.WriteLn(projectdep)
			}
			projectsessionend := "+EndProjectSection"
			writer.WriteLn(projectsessionend)
		}
		projectend := "EndProject"
		writer.WriteLn(projectend)
	}

	// Global
	//        GlobalSection(SolutionConfigurationPlatforms) = preSolution
	//            DevDebug|x64 = DevDebug|x64
	//            DevFinal|x64 = DevFinal|x64
	//            DevRelease|x64 = DevRelease|x64
	//            TestDebug|x64 = TestDebug|x64
	//            TestRelease|x64 = TestRelease|x64
	//        EndGlobalSection
	configs := make(map[string]string)

	for _, project := range sln.Projects {
		for _, config := range project.Configs {
			for _, platform := range project.Platforms {
				configstr := fmt.Sprintf("%s|%s", config, platform)
				configs[configstr] = configstr
			}
		}
	}

	// SolutionConfigurationPlatforms
	writer.WriteLn("Global")
	writer.WriteLn("+GlobalSection(SolutionConfigurationPlatforms) = preSolution")
	for kconfig, vconfig := range configs {
		writer.WriteLn(fmt.Sprintf("++%s = %s", kconfig, vconfig))
	}
	writer.WriteLn("+EndGlobalSection")

	// ProjectConfigurationPlatforms
	writer.WriteLn("+GlobalSection(ProjectConfigurationPlatforms) = postSolution")
	for _, project := range sln.Projects {
		for _, config := range project.Configs {
			for _, platform := range project.Platforms {
				configplatform := fmt.Sprintf("%s|%s", config, platform)
				activecfg := fmt.Sprintf("++{%s}.%s.ActiveCfg = %s", project.GUID, configplatform, configplatform)
				buildcfg := fmt.Sprintf("++{%s}.%s.Buid.0 = %s", project.GUID, configplatform, configplatform)
				writer.WriteLn(activecfg)
				writer.WriteLn(buildcfg)
			}
		}
	}
	writer.WriteLn("+EndGlobalSection")

	// SolutionProperties
	writer.WriteLn("+GlobalSection(SolutionProperties) = preSolution")
	writer.WriteLn("++HideSolutionNode = FALSE")
	writer.WriteLn("+EndGlobalSection")

	writer.WriteLn("EndGlobal")
}
