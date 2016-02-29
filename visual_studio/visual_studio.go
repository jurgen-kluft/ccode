package vs

import (
	"fmt"
	"github.com/jurgen-kluft/xcode"
)

type VisualStudio2015 struct {
}

var gOSARCHToPlatform = map[string]string{
	"windows-x86":   "Win32",
	"windows-amd64": "x64",
}

func NewGeneratorForVisualStudio(version IDE) (ProjectGenerator, error) {
	switch version {
	case VS2015:
		return &VisualStudio2015{}, nil
	}
	return nil, fmt.Errorf("Unsupported visual studio version")
}

func (vs *VisualStudio2015) Generate(path string, pkg Package) {

	// For every dependency and main {
	//     Write out the vcxproject
	//     Write out the vcxproject filters
	// }

	// Write out the sln

}

var strVisualStudio2015Project string = `
<Project DefaultTargets="Build" ToolsVersion="4.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
  <ItemGroup Label="ProjectConfigurations">
    <ProjectConfiguration Include="${CONFIG}|${PLATFORM}">
      <Configuration>${CONFIG}</Configuration>
      <Platform>${PLATFORM}</Platform>
    </ProjectConfiguration>
  </ItemGroup>
  <PropertyGroup Condition="'$(Platform)'=='${PLATFORM}'" Label="Configuration">
    <PlatformToolset>${${PLATFORM}_TOOLSET}</PlatformToolset>
  </PropertyGroup>
  <PropertyGroup Label="Globals">
    <ProjectGuid>${GUID}</ProjectGuid>
    <PackageSignature>$(Configuration)_$(Platform)_$(PlatformToolset)</PackageSignature>
  </PropertyGroup>
  <PropertyGroup Label="Directories">
  </PropertyGroup>
  <ImportGroup Label="TargetDirs"/>
  <Import Project="$(VCTargetsPath)\Microsoft.Cpp.Default.props"/>
  <PropertyGroup Condition="'$(Configuration)|$(Platform)'=='${CONFIG}|${PLATFORM}'" Label="Configuration">
    <ConfigurationType>StaticLibrary</ConfigurationType>
    <UseDebugLibraries>true</UseDebugLibraries>
    <CLRSupport>false</CLRSupport>
    <CharacterSet>NotSet</CharacterSet>
  </PropertyGroup>
  <Import Project="$(VCTargetsPath)\Microsoft.Cpp.props"/>
  <ImportGroup Condition="'$(Platform)'=='${PLATFORM}'" Label="PropertySheets">
    <Import Condition="exists('$(UserRootDir)\Microsoft.Cpp.$(Platform).user.props')" Label="LocalAppDataPlatform" Project="$(UserRootDir)\Microsoft.Cpp.$(Platform).user.props"/>
  </ImportGroup>
  <PropertyGroup Condition="'$(Platform)'=='${PLATFORM}'">
    <LinkIncremental>true</LinkIncremental>
    <OutDir>$(SolutionDir)target\$(SolutionName)\outdir\$(ProjectName)_$(PackageSignature)\</OutDir>
    <IntDir>$(SolutionDir)target\$(SolutionName)\outdir\$(ProjectName)_$(PackageSignature)\</IntDir>
    <TargetName>$(ProjectName)_$(PackageSignature)</TargetName>
    <ExtensionsToDeleteOnClean>*.obj%3b*.d%3b*.map%3b*.lst%3b*.pch%3b$(TargetPath)</ExtensionsToDeleteOnClean>
    <GenerateManifest>false</GenerateManifest>
  </PropertyGroup>
  <ItemDefinitionGroup Condition="'$(Configuration)|$(Platform)'=='${CONFIG}|${PLATFORM}'">
    <ClCompile>
      <WarningLevel>Level3</WarningLevel>
      <Optimization>Disabled</Optimization>
      <PreprocessorDefinitions>${PREPROCESSOR_DEFINES}%()</PreprocessorDefinitions>
      <PrecompiledHeader>NotUsing</PrecompiledHeader>
      <AdditionalIncludeDirectories>${INCLUDE_DIRS}%()</AdditionalIncludeDirectories>
      <ExceptionHandling>false</ExceptionHandling>
    </ClCompile>
    <Link>
      <GenerateDebugInformation>true</GenerateDebugInformation>
      <AdditionalDependencies>%()</AdditionalDependencies>
    </Link>
    <Lib>
      <OutputFile>$(OutDir)\$(TargetName)$(TargetExt)</OutputFile>
    </Lib>
  </ItemDefinitionGroup>
  <ItemGroup>
    <ClCompile Include=""/>
  </ItemGroup>
  <ItemGroup>
    <ClInclude Include=""/>
  </ItemGroup>
  <ItemGroup>
    <None Include=""/>
  </ItemGroup>
  <Import Condition="'$(ConfigurationType)' == 'Makefile' and Exists('$(VCTargetsPath)\Platforms\$(Platform)\SCE.Makefile.$(Platform).targets')" Project="$(VCTargetsPath)\Platforms\$(Platform)\SCE.Makefile.$(Platform).targets"/>
  <Import Project="$(VCTargetsPath)\Microsoft.Cpp.targets"/>
  <ImportGroup Label="ExtensionTargets"/>
</Project>
`
