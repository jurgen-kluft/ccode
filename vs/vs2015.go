package vs

import (
	"fmt"
	"os"
	"strings"

	"github.com/jurgen-kluft/xcode/denv"
	"github.com/jurgen-kluft/xcode/items"
	"github.com/jurgen-kluft/xcode/uid"
	"github.com/jurgen-kluft/xcode/vars"

	"path/filepath"
)

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
func addProjectVariables(p *denv.Project, isdep bool, v vars.Variables) {
	v.AddVar(p.Name+":GUID", p.GUID)
	v.AddVar(p.Name+":ROOT_DIR", p.PackagePath)

	path, _ := filepath.Rel(p.ProjectPath, p.PackagePath)

	v.AddVar(p.Name+":OUTDIR", filepath.Join(path, "target", p.Name, "bin", "$(Configuration)_$(Platform)_$(PlatformToolset)"))
	v.AddVar(p.Name+":INTDIR", filepath.Join(path, "target", p.Name, "obj", "$(Configuration)_$(Platform)_$(PlatformToolset)"))

	for _, platform := range p.Platforms {
		for _, config := range p.Configs {
			includes := config.IncludeDirs.Prefix(path, items.PathPrefixer)
			includes = includes.Add("%(AdditionalIncludeDirectories)")
			v.AddVar(fmt.Sprintf("%s:INCLUDE_DIRS[%s][%s]", p.Name, platform, config.Name), includes.String())

			libdirs := config.LibraryDirs.Prefix(path, items.PathPrefixer)
			libdirs = libdirs.Add("%(AdditionalLibraryDirectories)")
			v.AddVar(fmt.Sprintf("%s:LIBRARY_DIRS[%s][%s]", p.Name, platform, config.Name), libdirs.String())

			libfiles := config.LibraryFiles.Add("%(AdditionalLibraryFiles)")
			v.AddVar(fmt.Sprintf("%s:LIBRARY_FILES[%s][%s]", p.Name, platform, config.Name), libfiles.String())

			defines := config.Defines.Add("%(PreprocessorDefinitions)")
			v.AddVar(fmt.Sprintf("%s:DEFINES[%s][%s]", p.Name, platform, config.Name), defines.String())
		}
	}

	configitems := []string{"OPTIMIZATION", "DEBUG_INFO", "USE_DEBUG_LIBS"}
	getdefault := func(configname string, configitem int) string {
		defaults := []string{}
		if strings.Contains(strings.ToLower(configname), "debug") {
			defaults = []string{"Disabled", "true", "true"}
		} else {
			defaults = []string{"Full", "false", "false"}
		}
		return defaults[configitem]
	}

	for i, configitem := range configitems {
		for _, platform := range p.Platforms {
			for _, config := range p.Configs {
				value := getdefault(config.Name, i)
				v.AddVar(fmt.Sprintf("%s:%s[%s][%s]", p.Name, configitem, platform, config.Name), value)
			}
		}
	}

	for _, platform := range p.Platforms {
		v.AddVar(fmt.Sprintf("TOOLSET[%s]", platform), "v140")
	}
}

// setupProjectPaths will set correct paths for the main and dependency packages
// Note: This currently assumes that the dependency packages are in the vendor
//       folder relative to the main package.
// All project and workspace files will be written in the root of the main package
func setupProjectPaths(p *denv.Project) {
	p.PackagePath, _ = os.Getwd()
	p.ProjectPath, _ = os.Getwd()
	for _, d := range p.Dependencies {
		d.PackagePath = filepath.Join(p.PackagePath, "vendor", d.PackageURL, d.Name)
		d.ProjectPath = p.ProjectPath
	}
}

// CPPprojectID is a GUID that is defining a C++ project in Visual Studio
var CPPprojectID = "8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942"

// generateVisualStudio2015Project generates a Visual Studio 2015 project (.vcxproj) file
func generateVisualStudio2015Project(prj *denv.Project, vars vars.Variables, replacer vars.Replacer, writer denv.ProjectWriter) {

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

	projects := []*denv.Project{}
	projects = append(projects, prj)
	for _, dep := range prj.Dependencies {
		projects = append(projects, dep)
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
			varkey := fmt.Sprintf("%s:USE_DEBUG_LIBS[%s][%s]", prj.Name, platform, config.Name)
			usedebuglibs, err := vars.GetVar(varkey)
			if err == nil {
				replacer.ReplaceInLines("${USE_DEBUG_LIBS}", usedebuglibs, configuration)
			} else {
				//fmt.Println("ERROR: could not find variable " + varkey)
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
		platformprops = append(platformprops, `++<OutDir>${OutDir}</OutDir>`)
		platformprops = append(platformprops, `++<IntDir>${IntDir}</IntDir>`)
		platformprops = append(platformprops, `++<TargetName>$(Config)_$(Platform)_$(ToolSet)</TargetName>`)
		platformprops = append(platformprops, `++<ExtensionsToDeleteOnClean>*.obj%3b*.d%3b*.map%3b*.lst%3b*.pch%3b$(TargetPath)</ExtensionsToDeleteOnClean>`)
		platformprops = append(platformprops, `++<GenerateManifest>false</GenerateManifest>`)
		platformprops = append(platformprops, `+</PropertyGroup>`)
		replacer.ReplaceInLines("${PLATFORM}", platform, platformprops)
		writer.WriteLns(platformprops)
	}

	includedirs := []string{}
	for _, project := range projects {
		includedir := "${${Name}:IncludeDir}"
		includedir = replacer.ReplaceInLine("${Name}", project.Name, includedir)
		includedir = vars.ReplaceInLine(replacer, includedir)
		includedirs = append(includedirs, includedir)
	}

	for _, platform := range prj.Platforms {
		for _, config := range prj.Configs {
			compileandlink := []string{}
			compileandlink = append(compileandlink, `+<ItemDefinitionGroup Condition="'$(Configuration)|$(Platform)'=='${CONFIG}|${PLATFORM}'">`)
			compileandlink = append(compileandlink, `++<ClCompile>`)
			compileandlink = append(compileandlink, `+++<PreprocessorDefinitions>${DEFINES}</PreprocessorDefinitions>`)
			compileandlink = append(compileandlink, `+++<AdditionalIncludeDirectories>${INCLUDE_DIRS}</AdditionalIncludeDirectories>`)
			compileandlink = append(compileandlink, `+++<WarningLevel>Level3</WarningLevel>`)
			compileandlink = append(compileandlink, `+++<Optimization>${OPTIMIZATION}</Optimization>`)
			compileandlink = append(compileandlink, `+++<PrecompiledHeader>NotUsing</PrecompiledHeader>`)
			compileandlink = append(compileandlink, `+++<ExceptionHandling>false</ExceptionHandling>`)
			compileandlink = append(compileandlink, `++</ClCompile>`)
			compileandlink = append(compileandlink, `++<Link>`)
			compileandlink = append(compileandlink, `+++<GenerateDebugInformation>${DEBUG_INFO}</GenerateDebugInformation>`)
			compileandlink = append(compileandlink, `+++<AdditionalDependencies>${LIBRARY_FILES};%(AdditionalDependencies)</AdditionalDependencies>`)
			compileandlink = append(compileandlink, `+++<AdditionalLibraryDirectories>${LIBRARY_DIRS}</AdditionalLibraryDirectories>`)
			compileandlink = append(compileandlink, `++</Link>`)
			compileandlink = append(compileandlink, `++<Lib>`)
			compileandlink = append(compileandlink, `+++<OutputFile>$(OutDir)\$(TargetName)$(TargetExt)</OutputFile>`)
			compileandlink = append(compileandlink, `++</Lib>`)
			compileandlink = append(compileandlink, `+</ItemDefinitionGroup>`)

			libraries := []string{}
			for _, depproject := range prj.Dependencies {
				if depproject.HasConfig(config.Name) {
					depconfig := depproject.Configs[config.Name]
					libraries = append(libraries, depconfig.LibraryFile)
				}
			}
			replacer.InsertInLines("${LIBRARY_FILES}", strings.Join(libraries, ";"), compileandlink)

			configitems := []string{"DEFINES", "INCLUDE_DIRS", "LIBRARY_DIRS", "LIBRARY_FILES"}
			for _, configitem := range configitems {
				varkeystr := fmt.Sprintf("${%s}", configitem)
				varlist := items.NewList("", ";")
				for _, project := range projects {
					varkey := fmt.Sprintf("%s:%s[%s][%s]", project.Name, configitem, platform, config.Name)
					varitem, err := vars.GetVar(varkey)
					if err == nil {
						varlist = varlist.Add(varitem)
					} else {
						//fmt.Println("ERROR: could not find variable " + varkey)
					}
				}
				varset := items.ListToSet(varlist)
				replacer.InsertInLines(varkeystr, varset.String(), compileandlink)
				replacer.ReplaceInLines(varkeystr, "", compileandlink)
			}

			optimization, err := vars.GetVar(fmt.Sprintf("%s:OPTIMIZATION[%s][%s]", prj.Name, platform, config.Name))
			if err == nil {
				replacer.ReplaceInLines("${OPTIMIZATION}", optimization, compileandlink)
			}

			debuginfo, err := vars.GetVar(fmt.Sprintf("%s:DEBUG_INFO[%s][%s]", prj.Name, platform, config.Name))
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

func generateFilters(prjguid string, files []string) (items map[string]string, filters map[string]string) {
	filters = make(map[string]string)
	items = make(map[string]string)
	for _, hdrfile := range files {
		dirpath := filepath.Dir(hdrfile)
		guid := uid.GetGUID(dirpath)
		filters[dirpath] = guid
		items[hdrfile] = dirpath

		// We need to add every 'depth' of the path
		for true {
			//fmt.Printf("dir:\"%s\" --> \"%s\" | ", dirpath, filepath.Dir(dirpath))
			dirpath = filepath.Dir(dirpath)
			if dirpath == "." || dirpath == "/" {
				break
			}
			// Generate a specific GUID for this entry
			guid = uid.GetGUID(prjguid + dirpath)
			filters[dirpath] = guid
		}
		//fmt.Println("")
	}
	return
}

// generateVisualStudio2015ProjectFilters generates a Visual Studio 2015 project filters (.vcxproj.filters) file
func generateVisualStudio2015ProjectFilters(prj *denv.Project, writer denv.ProjectWriter) {
	writer.WriteLn(`<?xml version="1.0" encoding="utf-8"?>`)
	writer.WriteLn(`<Project ToolsVersion="4.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">`)

	includes, includefilters := generateFilters(prj.Name+prj.GUID, prj.HdrFiles.Files)
	cpp, cppfilters := generateFilters(prj.Name+prj.GUID, prj.SrcFiles.Files)

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
// for the current project. It also generates the project file for the current projects
// as well as the project files for the dependencies.
func GenerateVisualStudio2015Solution(p *denv.Project) {

	setupProjectPaths(p)

	writer := &denv.ProjectTextWriter{}
	slnfilepath := filepath.Join(p.ProjectPath, p.Name+".sln")
	if writer.Open(slnfilepath) != nil {
		fmt.Printf("Error opening file '%s'", slnfilepath)
		return
	}

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

	projects := []*denv.Project{p}
	for _, dep := range p.Dependencies {
		projects = append(projects, dep)
	}

	variables := vars.NewVars()
	replacer := vars.NewReplacer()

	// Main project
	addProjectVariables(p, false, variables)
	p.ReplaceVars(variables, replacer)

	// And dependency projects
	for _, prj := range p.Dependencies {
		addProjectVariables(prj, true, variables)
		prj.ReplaceVars(variables, replacer)
	}
	variables.Print()

	// Glob all the source and header files for every project
	for _, prj := range projects {
		prj.SrcFiles.GlobFiles(prj.PackagePath)
		prj.HdrFiles.GlobFiles(prj.PackagePath)
	}

	// Generate all the projects
	for _, prj := range projects {
		// Generate the project file
		prjwriter := &denv.ProjectTextWriter{}
		prjwriter.Open(filepath.Join(prj.ProjectPath, prj.Name+".vcxproj"))
		generateVisualStudio2015Project(prj, variables, replacer, prjwriter)
		prjwriter.Close()

		// Generate the project filters file
		prjwriter = &denv.ProjectTextWriter{}
		prjwriter.Open(filepath.Join(prj.ProjectPath, prj.Name+".vcxproj.filters"))
		generateVisualStudio2015ProjectFilters(prj, prjwriter)
		prjwriter.Close()
	}

	for _, prj := range projects {
		projectbeginfmt := "Project(\"{%s}\") = \"%s\", \"%s\", \"{%s}\""
		projectbegin := fmt.Sprintf(projectbeginfmt, CPPprojectID, prj.Name, denv.Path(filepath.Join(prj.ProjectPath, prj.Name+".vcxproj")), prj.GUID)
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

	for _, project := range projects {
		for _, config := range project.Configs {
			for _, platform := range project.Platforms {
				configstr := fmt.Sprintf("%s|%s", config.Name, platform)
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
	for _, project := range projects {
		for _, config := range project.Configs {
			for _, platform := range project.Platforms {
				configplatform := fmt.Sprintf("%s|%s", config.Name, platform)
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
	writer.Close()
}
