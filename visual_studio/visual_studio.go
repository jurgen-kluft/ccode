package vs

import (
	"fmt"
	"github.com/jurgen-kluft/xcode/denv"
)

func IsVisualStudio(ide string) bool {
	_, err := GetVisualStudio(ide)
	return err == nil
}

func GetVisualStudio(ide string) (denv.IDE, error) {
	if ide == "VS2015" {
		return denv.VS2015, nil
	} else if ide == "VS2013" {
		return denv.VS2013, nil
	} else if ide == "VS2012" {
		return denv.VS2012, nil
	}
	return denv.VS2015, fmt.Errorf("Not a Visual Studio IDE token")
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

type VariableReplacer interface {
	Replace(variable string, replace string, body string) string // Replaces occurences of @variable with @replace and thus will remove the variable from @body
	Insert(variable string, insert string, body string) string   //Inserts @variable at places where @variable occurs without removing the variable in @body
}

type Variables interface {
	AddVar(key string, value string)
	GetVar(key string) (string, error)
	Replace(replacer VariableReplacer, body string) string
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

func GenerateVisualStudio2015Project(projectname string, sourcefiles []string, headerfiles []string, platforms []string, configs []string, depprojectnames []string, vars Variables, replacer VariableReplacer, writer ProjectWriter) {

	writer.Write(`<Project DefaultTargets="Build" ToolsVersion="4.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">`)

	projconfigbegin := `
    <ItemGroup Label="ProjectConfigurations">`
	writer.Write(projconfigbegin)

	for _, platform := range platforms {
		for _, config := range configs {
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

	for _, platform := range platforms {
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

	globals = replacer.Replace("${Name}", projectname, globals)
	globals = vars.Replace(replacer, globals)
	writer.Write(globals)

	for _, depprojectname := range depprojectnames {
		projdirs := `
    <PropertyGroup Label="Directories">
        <__${Name}:RootDir__>${${Name}:RootDir}</__${Name}:RootDir__>
        <__${Name}:LibraryDir__>${${Name}:LibraryDir}target\${Name}\outdir\${Name}_$(PackageSignature)\</__${Name}:LibraryDir__>
        <__${Name}:IncludeDir__>${${Name}:IncludeDir}source\main\include\</__${Name}:IncludeDir__>
        <__${Name}:TestIncludeDir__>${${Name}:TestIncludeDir}source\test\include\</__${Name}:TestIncludeDir__>
    </PropertyGroup>`
		projdirs = replacer.Replace("${Name}", depprojectname, projdirs)
		projdirs = vars.Replace(replacer, projdirs)
		writer.Write(projdirs)
	}

	targetdirs := `
    <ImportGroup Label="TargetDirs"/>`
	writer.Write(targetdirs)

	defaultprops := `
    <Import Project="$(VCTargetsPath)\Microsoft.Cpp.Default.props"/>`
	writer.Write(defaultprops)

	for _, platform := range platforms {
		for _, config := range configs {
			configuration := `
    <PropertyGroup Condition="'$(Configuration)|$(Platform)'=='${CONFIG}|${PLATFORM}'" Label="Configuration">
        <ConfigurationType>StaticLibrary</ConfigurationType>
        <UseDebugLibraries>true</UseDebugLibraries>
        <CLRSupport>false</CLRSupport>
        <CharacterSet>NotSet</CharacterSet>
    </PropertyGroup>`
			configuration = replacer.Replace("${PLATFORM}", platform, configuration)
			configuration = replacer.Replace("${CONFIG}", config, configuration)
			configuration = vars.Replace(replacer, configuration)
			writer.Write(configuration)
		}
	}

	cppprops := `
    <Import Project="$(VCTargetsPath)\Microsoft.Cpp.props"/>`
	writer.Write(cppprops)

	userprops := `
    <ImportGroup Condition="'$(Platform)'=='${PLATFORM}'" Label="PropertySheets">
        <Import Condition="exists('$(UserRootDir)\Microsoft.Cpp.$(Platform).user.props')" Label="LocalAppDataPlatform" Project="$(UserRootDir)\Microsoft.Cpp.$(Platform).user.props"/>
    </ImportGroup>`
	writer.Write(userprops)

	for _, platform := range platforms {
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
	for _, depprojectname := range depprojectnames {
		depincludedirs := "${${Name}:IncludeDir}"
		depincludedirs = replacer.Replace("${Name}", depprojectname, depincludedirs)
		depincludedirs = vars.Replace(replacer, depincludedirs)
		includedirs = includedirs + depincludedirs
	}

	for _, platform := range platforms {
		for _, config := range configs {
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
				for _, depprojectname := range depprojectnames {
					varvalue, err := vars.GetVar(fmt.Sprintf("%s:%s[%s][%s]", depprojectname, varkey, config, platform))
					if err == nil {
						compileandlink = replacer.Insert(varkeystr, varvalue, compileandlink)
					}
				}
				compileandlink = replacer.Replace(varkeystr, "", compileandlink)
			}

			optimization, err := vars.GetVar(fmt.Sprintf("%s:OPTIMIZATION[%s]", projectname, config))
			if err == nil {
				compileandlink = replacer.Replace("${OPTIMIZATION}", optimization, compileandlink)
			}

			debuginfo, err := vars.GetVar(fmt.Sprintf("%s:DEBUGINFO[%s]", projectname, config))
			if err == nil {
				compileandlink = replacer.Replace("${DEBUGINFO}", debuginfo, compileandlink)
			}

			compileandlink = replacer.Replace("${PLATFORM}", platform, compileandlink)
			compileandlink = replacer.Replace("${CONFIG}", config, compileandlink)
			compileandlink = vars.Replace(replacer, compileandlink)
			writer.Write(compileandlink)
		}
	}

	writer.Write(`  <ItemGroup>`)
	for _, srcfile := range sourcefiles {
		clcompile := `
        <ClCompile Include="${FILE}"/>`
		clcompile = replacer.Replace("${FILE}", srcfile, clcompile)
		writer.Write(clcompile)
	}
	writer.Write(`  </ItemGroup>`)

	writer.Write(`  <ItemGroup>`)
	for _, hdrfile := range headerfiles {
		clinclude := `
        <ClInclude Include="${FILE}"/>`
		clinclude = replacer.Replace("${FILE}", hdrfile, clinclude)
		writer.Write(clinclude)
	}
	writer.Write(`  </ItemGroup>`)

	otherincludes := `
    <ItemGroup>
        <None Include=""/>
    </ItemGroup>`
	writer.Write(otherincludes)

	imports := `
    <Import Condition="'$(ConfigurationType)' == 'Makefile' and Exists('$(VCTargetsPath)\Platforms\$(Platform)\SCE.Makefile.$(Platform).targets')" Project="$(VCTargetsPath)\Platforms\$(Platform)\SCE.Makefile.$(Platform).targets"/>
    <Import Project="$(VCTargetsPath)\Microsoft.Cpp.targets"/>
    <ImportGroup Label="ExtensionTargets"/>`
	writer.Write(imports)

	writer.Write(`</Project>`)
}
