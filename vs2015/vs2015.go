package vs2015

import (
	"fmt"
	"github.com/jurgen-kluft/xcode/denv"
	"github.com/jurgen-kluft/xcode/uid"
	"github.com/jurgen-kluft/xcode/vars"
	//	"path"
	"path/filepath"
)

func IsVisualStudio(ide string) bool {
	vs := GetVisualStudio(ide)
	return vs != -1
}

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

type VisualStudio2015 struct {
}

var gOSARCHToPlatform = map[string]string{
	"windows-x86":   "Win32",
	"windows-amd64": "x64",
}

type ProjectWriter interface {
	Write(string) error
}

// Required Variables:
//   Example for 'xhash' project with 'xbase' as a dependency:
//   - xhash:GUID
//   - xbase:GUID
//   - xhash:RootDir, xhash:LibraryDir, xhash:IncludeDir, xhash:TestIncludeDir
//   - xbase:RootDir, xbase:LibraryDir, xbase:IncludeDir, xbase:TestIncludeDir
//   - xhash:PP_DEFINES[Platform][Config], xhash:INCLUDE_DIRS[Platform][Config], xhash:LINK_WITH[Platform][Config]
//   - xbase:PP_DEFINES[Platform][Config], xbase:INCLUDE_DIRS[Platform][Config], xbase:LINK_WITH[Platform][Config]
//   - xhash:OPTIMIZATION[Platform][Config]
//   - xhash:DEBUG_INFO[Platform][Config]
//   - xhash:TOOL_SET[Platform]
//

var CPPprojectID string = "8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942"

func GenerateVisualStudio2015Project(prj denv.Project, vars vars.Variables, replacer vars.Replacer, writer ProjectWriter) {

	writer.Write(`<?xml version="1.0" encoding="utf-8"?>`)
	writer.Write("\n")
	writer.Write(`<Project DefaultTargets="Build" ToolsVersion="14.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">`)
	writer.Write("\n")

	projconfigbegin := `    <ItemGroup Label="ProjectConfigurations">`
	writer.Write(projconfigbegin)

	for _, platform := range prj.Platforms {
		for _, config := range prj.Configs {
			projconfig := `
    <ProjectConfiguration Include="${CONFIG}|${PLATFORM}">
        <Configuration>${CONFIG}</Configuration>
        <Platform>${PLATFORM}</Platform>
    </ProjectConfiguration>`

			projconfig = replacer.Replace("${PLATFORM}", platform, projconfig)
			projconfig = replacer.Replace("${CONFIG}", config, projconfig)
			writer.Write(projconfig)
		}
	}

	projconfigend := `
    </ItemGroup>`
	writer.Write(projconfigend)

	for _, platform := range prj.Platforms {
		toolsets := `
    <PropertyGroup Condition="'$(Platform)'=='${PLATFORM}'" Label="Configuration">
        <PlatformToolset>${TOOLSET[${PLATFORM}]}</PlatformToolset>
    </PropertyGroup>`

		toolsets = replacer.Replace("${PLATFORM}", platform, toolsets)
		toolsets = vars.Replace(replacer, toolsets)
		writer.Write(toolsets)
	}

	globals := `
    <PropertyGroup Label="Globals">
        <ProjectGuid>${${Name}:GUID}</ProjectGuid>
        <PackageSignature>$(Configuration)_$(Platform)_$(PlatformToolset)</PackageSignature>
    </PropertyGroup>`

	globals = replacer.Replace("${Name}", prj.Name, globals)
	globals = vars.Replace(replacer, globals)
	writer.Write(globals)

	for _, depproject := range append(prj.Dependencies, prj) {
		projdirs := `
    <PropertyGroup Label="Directories">
        <__${Name}RootDir__>${${Name}:RootDir}</__${Name}RootDir__>
        <__${Name}LibraryDir__>${${Name}:LibraryDir}target\${Name}\outdir\${Name}_$(PackageSignature)\</__${Name}LibraryDir__>
        <__${Name}IncludeDir__>${${Name}:IncludeDir}source\main\include\</__${Name}IncludeDir__>
        <__${Name}TestIncludeDir__>${${Name}:TestIncludeDir}source\test\include\</__${Name}TestIncludeDir__>
    </PropertyGroup>`
		projdirs = replacer.Replace("${Name}", depproject.Name, projdirs)
		projdirs = vars.Replace(replacer, projdirs)
		writer.Write(projdirs)
	}

	targetdirs := `
    <ImportGroup Label="TargetDirs"/>`
	writer.Write(targetdirs)

	defaultprops := `
    <Import Project="$(VCTargetsPath)\Microsoft.Cpp.Default.props"/>`
	writer.Write(defaultprops)

	for _, platform := range prj.Platforms {
		for _, config := range prj.Configs {
			configuration := `
    <PropertyGroup Condition="'$(Configuration)|$(Platform)'=='${CONFIG}|${PLATFORM}'" Label="Configuration">
        <ConfigurationType>StaticLibrary</ConfigurationType>
        <UseDebugLibraries>${USE_DEBUG_LIBS}</UseDebugLibraries>
        <CLRSupport>false</CLRSupport>
        <CharacterSet>NotSet</CharacterSet>
    </PropertyGroup>`
			usedebuglibs, err := vars.GetVar(fmt.Sprintf("%s:USE_DEBUG_LIBS[%s][%s]", prj.Name, platform, config))
			if err == nil {
				configuration = replacer.Replace("${USE_DEBUG_LIBS}", usedebuglibs, configuration)
			}

			configuration = replacer.Replace("${PLATFORM}", platform, configuration)
			configuration = replacer.Replace("${CONFIG}", config, configuration)
			configuration = vars.Replace(replacer, configuration)
			writer.Write(configuration)
		}
	}

	cppprops := `
    <Import Project="$(VCTargetsPath)\Microsoft.Cpp.props"/>`
	writer.Write(cppprops)

	for _, platform := range prj.Platforms {
		userprops := `
    <ImportGroup Condition="'$(Platform)'=='${PLATFORM}'" Label="PropertySheets">
        <Import Condition="exists('$(UserRootDir)\Microsoft.Cpp.$(Platform).user.props')" Label="LocalAppDataPlatform" Project="$(UserRootDir)\Microsoft.Cpp.$(Platform).user.props"/>
    </ImportGroup>`
		userprops = replacer.Replace("${PLATFORM}", platform, userprops)
		writer.Write(userprops)
	}

	for _, platform := range prj.Platforms {
		platformprops := `
    <PropertyGroup Condition="'$(Platform)'=='${PLATFORM}'">
        <LinkIncremental>true</LinkIncremental>
        <OutDir>$(SolutionDir)target\$(SolutionName)\outdir\$(ProjectName)_$(PackageSignature)\</OutDir>
        <IntDir>$(SolutionDir)target\$(SolutionName)\outdir\$(ProjectName)_$(PackageSignature)\</IntDir>
        <TargetName>$(ProjectName)_$(PackageSignature)</TargetName>
        <ExtensionsToDeleteOnClean>*.obj%3b*.d%3b*.map%3b*.lst%3b*.pch%3b$(TargetPath)</ExtensionsToDeleteOnClean>
        <GenerateManifest>false</GenerateManifest>
    </PropertyGroup>`
		platformprops = replacer.Replace("${PLATFORM}", platform, platformprops)
		writer.Write(platformprops)
	}

	includedirs := ""
	for _, depproject := range prj.Dependencies {
		depincludedirs := "${${Name}:IncludeDir}"
		depincludedirs = replacer.Replace("${Name}", depproject.Name, depincludedirs)
		depincludedirs = vars.Replace(replacer, depincludedirs)
		includedirs = includedirs + depincludedirs
	}

	for _, platform := range prj.Platforms {
		for _, config := range prj.Configs {
			compileandlink := `
    <ItemDefinitionGroup Condition="'$(Configuration)|$(Platform)'=='${CONFIG}|${PLATFORM}'">
        <ClCompile>
            <PreprocessorDefinitions>${PP_DEFINES}%(PreprocessorDefinitions)</PreprocessorDefinitions>
            <AdditionalIncludeDirectories>${INCLUDE_DIRS}%(AdditionalIncludeDirectories)</AdditionalIncludeDirectories>
            <WarningLevel>Level3</WarningLevel>
            <Optimization>${OPTIMIZATION}</Optimization>
            <PrecompiledHeader>NotUsing</PrecompiledHeader>
            <ExceptionHandling>false</ExceptionHandling>
        </ClCompile>
        <Link>
            <GenerateDebugInformation>${DEBUG_INFO}</GenerateDebugInformation>
            <AdditionalDependencies>${LINK_WITH}%(AdditionalDependencies)</AdditionalDependencies>
        </Link>
        <Lib>
            <OutputFile>$(OutDir)\$(TargetName)$(TargetExt)</OutputFile>
        </Lib>
    </ItemDefinitionGroup>`

			varkeys := []string{"PP_DEFINES", "INCLUDE_DIRS", "LINK_WITH"}
			for _, varkey := range varkeys {
				varkeystr := fmt.Sprintf("${%s}", varkey)
				for _, depprojectname := range prj.Dependencies {
					varvalue, err := vars.GetVar(fmt.Sprintf("%s:%s[%s][%s]", depprojectname, varkey, config, platform))
					if err == nil {
						compileandlink = replacer.Insert(varkeystr, varvalue, compileandlink)
					}
				}
				compileandlink = replacer.Replace(varkeystr, "", compileandlink)
			}

			optimization, err := vars.GetVar(fmt.Sprintf("%s:OPTIMIZATION[%s][%s]", prj.Name, platform, config))
			if err == nil {
				compileandlink = replacer.Replace("${OPTIMIZATION}", optimization, compileandlink)
			}

			debuginfo, err := vars.GetVar(fmt.Sprintf("%s:DEBUG_INFO[%s][%s]", prj.Name, platform, config))
			if err == nil {
				compileandlink = replacer.Replace("${DEBUG_INFO}", debuginfo, compileandlink)
			}

			compileandlink = replacer.Replace("${PLATFORM}", platform, compileandlink)
			compileandlink = replacer.Replace("${CONFIG}", config, compileandlink)
			compileandlink = vars.Replace(replacer, compileandlink)
			writer.Write(compileandlink)
		}
	}
	writer.Write("\n")

	writer.Write("\t<ItemGroup>\n")
	for _, srcfile := range prj.SrcFiles {
		clcompile := "\t\t<ClCompile Include=\"${FILE}\"/>\n"
		clcompile = replacer.Replace("${FILE}", srcfile, clcompile)
		writer.Write(clcompile)
	}
	writer.Write("\t</ItemGroup>\n")

	if len(prj.HdrFiles) > 0 {
		writer.Write("\t<ItemGroup>\n")
		for _, hdrfile := range prj.HdrFiles {
			clinclude := "\t\t<ClInclude Include=\"${FILE}\"/>\n"
			clinclude = replacer.Replace("${FILE}", hdrfile, clinclude)
			writer.Write(clinclude)
		}
		writer.Write("\t</ItemGroup>\n")
	}

	//writer.Write("\t<ItemGroup>\n")
	//writer.Write("\t\t<None Include=\"\"/>\n")
	//writer.Write("\t</ItemGroup>\n")

	imports := `
    <Import Condition="'$(ConfigurationType)' == 'Makefile' and Exists('$(VCTargetsPath)\Platforms\$(Platform)\SCE.Makefile.$(Platform).targets')" Project="$(VCTargetsPath)\Platforms\$(Platform)\SCE.Makefile.$(Platform).targets"/>
    <Import Project="$(VCTargetsPath)\Microsoft.Cpp.targets"/>
    <ImportGroup Label="ExtensionTargets"/>
`
	writer.Write(imports)

	writer.Write(`</Project>`)
}

func GenerateVisualStudio2015ProjectFilters(prj denv.Project, writer ProjectWriter) {
	writer.Write(`<?xml version="1.0" encoding="utf-8"?>`)
	writer.Write("\n")
	writer.Write(`<Project ToolsVersion="4.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">`)
	writer.Write("\n")

	includefilters := make(map[string]string)
	includes := make(map[string]string)
	for _, hdrfile := range prj.HdrFiles {
		dirpath := filepath.Dir(hdrfile)
		guid := uid.GetGUID(dirpath)
		includefilters[dirpath] = guid
		includes[hdrfile] = dirpath

		// We need to add every 'depth' of the path
		for true {
			fmt.Printf("dir:\"%s\" --> \"%s\" | ", dirpath, filepath.Dir(dirpath))
			dirpath = filepath.Dir(dirpath)
			if dirpath == "." {
				break
			}
			guid = uid.GetGUID(dirpath)
			includefilters[dirpath] = guid
		}
		fmt.Println("")
	}

	cppfilters := make(map[string]string)
	cpp := make(map[string]string)
	for _, srcfile := range prj.SrcFiles {
		dirpath := filepath.Dir(srcfile)
		guid := uid.GetGUID(dirpath)
		cppfilters[dirpath] = guid
		cpp[srcfile] = dirpath

		// We need to add every 'depth' of the path
		for true {
			fmt.Printf("dir:\"%s\" --> \"%s\" | ", dirpath, filepath.Dir(dirpath))
			dirpath = filepath.Dir(dirpath)
			if dirpath == "." {
				break
			}
			guid = uid.GetGUID(dirpath)
			cppfilters[dirpath] = guid
		}
		fmt.Println("")
	}

	writer.Write("\t<ItemGroup>\n")
	for k, v := range includefilters {
		writer.Write(fmt.Sprintf("\t\t<Filter Include=\"%s\">\n", k))
		writer.Write(fmt.Sprintf("\t\t\t<UniqueIdentifier>{%s}</UniqueIdentifier>\n", v))
		writer.Write("\t\t</Filter>\n")
	}
	for k := range cppfilters {
		writer.Write(fmt.Sprintf("\t\t<Filter Include=\"%s\">\n", k))
		writer.Write(fmt.Sprintf("\t\t\t<UniqueIdentifier>{%s}</UniqueIdentifier>\n", uid.GetGUID(k)))
		writer.Write("\t\t</Filter>\n")
	}
	writer.Write("\t</ItemGroup>\n")

	writer.Write("\t<ItemGroup>\n")
	for k, v := range includes {
		writer.Write(fmt.Sprintf("\t\t<ClInclude Include=\"%s\">\n", k))
		writer.Write(fmt.Sprintf("\t\t\t<Filter>%s</Filter>\n", v))
		writer.Write("\t\t</ClInclude>\n")
	}
	writer.Write("\t</ItemGroup>\n")

	writer.Write("\t<ItemGroup>\n")
	for k, v := range cpp {
		writer.Write(fmt.Sprintf("\t\t<ClCompile Include=\"%s\">\n", k))
		writer.Write(fmt.Sprintf("\t\t\t<Filter>%s</Filter>\n", v))
		writer.Write("\t\t</ClCompile>\n")
	}
	writer.Write("\t</ItemGroup>\n")

	writer.Write("</Project>\n")
}

func GenerateVisualStudio2015Solution(sln denv.Solution, writer ProjectWriter) {
	writer.Write("Microsoft Visual Studio Solution File, Format Version 12.00\n")
	writer.Write("# Visual Studio 14\n")
	writer.Write("VisualStudioVersion = 14.0.24720.0\n")
	writer.Write("MinimumVisualStudioVersion = 10.0.40219.1\n")

	// Write Projects and their dependency information
	//
	//          Project("{8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942}") = "xhash", "source\main\cpp\xhash.vcxproj", "{04AB9C6F-0B84-4111-A772-53C03F5CB3C2}"
	//          	ProjectSection(ProjectDependencies) = postProject
	//          		{B83DA73D-6E7B-458D-A6C7-87013421D360} = {B83DA73D-6E7B-458D-A6C7-87013421D360}
	//          	EndProjectSection
	//          EndProject
	//
	for _, prj := range sln.Projects {
		projectbeginfmt := "Project(\"{%s}\") = \"%s\", \"%s\\%s.vcxproj\", \"{%s}\"\n"
		projectbegin := fmt.Sprintf(projectbeginfmt, CPPprojectID, prj.Name, prj.Path, prj.Name, prj.GUID)
		writer.Write(projectbegin)
		if len(prj.Dependencies) > 0 {
			projectsessionbegin := "\tProjectSection(ProjectDependencies) = postProject\n"
			writer.Write(projectsessionbegin)
			for _, dep := range prj.Dependencies {
				projectdep := fmt.Sprintf("\t\t{%s} = {%s}\n", dep.GUID, dep.GUID)
				writer.Write(projectdep)
			}
			projectsessionend := "\tEndProjectSection\n"
			writer.Write(projectsessionend)
		}
		projectend := "EndProject\n"
		writer.Write(projectend)
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
	writer.Write("Global\n")
	writer.Write("\tGlobalSection(SolutionConfigurationPlatforms) = preSolution\n")
	for kconfig, vconfig := range configs {
		writer.Write(fmt.Sprintf("\t\t%s = %s\n", kconfig, vconfig))
	}
	writer.Write("\tEndGlobalSection\n")

	// ProjectConfigurationPlatforms
	writer.Write("\tGlobalSection(ProjectConfigurationPlatforms) = postSolution\n")
	for _, project := range sln.Projects {
		for _, config := range project.Configs {
			for _, platform := range project.Platforms {
				configplatform := fmt.Sprintf("%s|%s", config, platform)
				activecfg := fmt.Sprintf("{%s}.%s.ActiveCfg = %s\n", project.GUID, configplatform, configplatform)
				buildcfg := fmt.Sprintf("{%s}.%s.Buid.0 = %s\n", project.GUID, configplatform, configplatform)
				writer.Write(activecfg)
				writer.Write(buildcfg)
			}
		}
	}
	writer.Write("\tEndGlobalSection\n")

	// SolutionProperties
	writer.Write("\tGlobalSection(SolutionProperties) = preSolution\n")
	writer.Write("\t\tHideSolutionNode = FALSE\n")
	writer.Write("\tEndGlobalSection\n")

	writer.Write("EndGlobal\n")
}
