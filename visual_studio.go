package xcode

import (
	"fmt"
)

type VisualStudio interface {
	GenerateSolution(path string, pkg Package)
	GenerateProject(path string, pkg Package)
	GenerateFilters(path string, pkg Package)
}

type VisualStudio2015 struct {
}

type VisualStudioVersion int

const (
	VS2012 VisualStudioVersion = 2012
	VS2013 VisualStudioVersion = 2013
	VS2015 VisualStudioVersion = 2015
)

func NewVisualStudioGenerator(version VisualStudioVersion) (VisualStudio, error) {
	switch version {
	case VS2015:
		return &VisualStudio2015{}, nil
	}
	return nil, fmt.Errorf("Wrong visual studio version")
}

func (vs *VisualStudio2015) GenerateSolution(path string, pkg Package) {

	// Write out the sln

}

func (vs *VisualStudio2015) GenerateProject(path string, pkg Package) {

	// Write out the vcxproject

}

func (vs *VisualStudio2015) GenerateFilters(path string, pkg Package) {

	// Write out the vcxproject

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
