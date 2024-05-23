package vs

import (
	"fmt"
	"os"
	"strings"

	"github.com/jurgen-kluft/ccode/denv"
	"github.com/jurgen-kluft/ccode/items"
	"github.com/jurgen-kluft/ccode/uid"
	"github.com/jurgen-kluft/ccode/vars"

	"path/filepath"
)

// AddProjectVariables adds variables from the Project information
//   Example for 'xhash' project with 'cbase' as a dependency:
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

func mergeProjectDependency(p *denv.Project) {
	// Merge include directories, library directories, libraries and defines from dependencies
	//    Project.Platform.Config[debug].IncludeDirs <- Dep.Platform.Config[debug].IncludeDirs
	//    Project.Platform.Config[release].IncludeDirs <- Dep.Platform.Config[release].IncludeDirs
	//    ... etc.
	for _, config := range p.Platform.Configs {
		for _, dep := range p.Dependencies {
			depConfig := dep.GetConfig(config.Name)
			p.LibraryFiles = p.LibraryFiles.Merge(dep.LibraryFiles)
			config.LibraryFiles = config.LibraryFiles.Merge(p.LibraryFiles)
			if depConfig != nil {
				fmt.Println("Merging ", dep.Name, " into ", p.Name, " for ", config.Name)
				//fmt.Println("  Includes: ", config.IncludeDirs.String(), " <- ", depConfig.IncludeDirs.String())
				//fmt.Println("  Libraries: ", config.LibraryDirs.String(), " <- ", depConfig.LibraryDirs.String())
				//fmt.Println("  Defines: ", config.Defines.String(), " <- ", depConfig.Defines.String())
				//fmt.Println("  Libraries: ", config.LibraryFiles.String(), " <- ", depConfig.LibraryFiles.String())
				config.IncludeDirs = config.IncludeDirs.Merge(depConfig.IncludeDirs)
				config.LibraryDirs = config.LibraryDirs.Merge(depConfig.LibraryDirs)
				config.LibraryFiles = config.LibraryFiles.Merge(depConfig.LibraryFiles)
				config.Defines = config.Defines.Merge(depConfig.Defines)
			}
		}
	}
}

func initProjectVariables(p *denv.Project, v vars.Variables, r vars.Replacer) {
	p.MergeVars(v)
	p.ReplaceVars(v, r)
}

func addProjectVariables(p *denv.Project, v vars.Variables, r vars.Replacer, dev denv.DevEnum) {
	v.AddVar(p.Name+":GUID", p.GUID)
	v.AddVar(p.Name+":ROOT_DIR", p.PackagePath)

	//path, _ := filepath.Rel(p.ProjectPath, p.PackagePath)

	// Every output should be pointing into the current local /target folder
	v.AddVar(p.Name+":OUTDIR", filepath.Join("target", p.Name, "bin", "$(Configuration)_$(Platform)_$(PlatformToolset)")+"\\")
	v.AddVar(p.Name+":INTDIR", filepath.Join("target", p.Name, "obj", "$(Configuration)_$(Platform)_$(PlatformToolset)")+"\\")

	switch p.Type {
	case denv.StaticLibrary:
		v.AddVar(p.Name+":TYPE", "StaticLibrary")
	case denv.SharedLibrary:
		v.AddVar(p.Name+":TYPE", "SharedLibrary")
	case denv.Executable:
		v.AddVar(p.Name+":TYPE", "Application")
	}

	for _, config := range p.Platform.Configs {
		v.AddVar(fmt.Sprintf("%s:INCLUDE_DIRS[%s][%s]", p.Name, p.Platform.Name, config.Name), config.IncludeDirs.String())
		v.AddVar(fmt.Sprintf("%s:LIBRARY_DIRS[%s][%s]", p.Name, p.Platform.Name, config.Name), config.LibraryDirs.String())
		v.AddVar(fmt.Sprintf("%s:LIBRARY_FILES[%s][%s]", p.Name, p.Platform.Name, config.Name), config.LibraryFiles.String())
		v.AddVar(fmt.Sprintf("%s:DEFINES[%s][%s]", p.Name, p.Platform.Name, config.Name), config.Defines.String())
	}

	configitems := []string{"OPTIMIZATION", "DEBUG_INFO", "USE_DEBUG_LIBS"}
	getdefault := func(configname string, configitem int) string {
		var defaults []string
		if strings.Contains(strings.ToLower(configname), "debug") {
			defaults = []string{"Disabled", "true", "true"}
		} else {
			defaults = []string{"Full", "false", "false"}
		}
		return defaults[configitem]
	}

	for i, configitem := range configitems {
		for _, config := range p.Platform.Configs {
			value := getdefault(config.Name, i)
			v.AddVar(fmt.Sprintf("%s:%s[%s][%s]", p.Name, configitem, p.Platform.Name, config.Name), value)
		}
	}

	// based on 'dev' we can set the toolset
	switch dev {
	case denv.VS2015:
		v.AddVar(fmt.Sprintf("TOOLSET[%s]", p.Platform.Name), "v140")
	case denv.VS2017:
		v.AddVar(fmt.Sprintf("TOOLSET[%s]", p.Platform.Name), "v141")
	case denv.VS2019:
		v.AddVar(fmt.Sprintf("TOOLSET[%s]", p.Platform.Name), "v142")
	case denv.VS2022:
		v.AddVar(fmt.Sprintf("TOOLSET[%s]", p.Platform.Name), "v143")
	}
}

// setupProjectPaths will set correct paths for the main and dependency packages
// Note: This currently assumes that the dependency packages are in the vendor
//
//	folder relative to the main package.
//
// All project and workspace files will be written in the root of the main package
func setupProjectPaths(prj *denv.Project, deps []*denv.Project) {
	prj.PackagePath, _ = os.Getwd()
	prj.ProjectPath, _ = os.Getwd()
	fmt.Println("PACKAGE:" + prj.Name + " -  packagePath=" + prj.PackagePath + ", projectpath=" + prj.ProjectPath)

	for _, dep := range deps {
		dep.PackagePath = denv.Path(filepath.Join(prj.PackagePath, "..", dep.Name))
		dep.ProjectPath = prj.ProjectPath
		path, _ := filepath.Rel(dep.ProjectPath, dep.PackagePath)
		for _, config := range dep.Platform.Configs {
			config.IncludeDirs = config.IncludeDirs.Prefix(path, items.PathPrefixer)
			//config.LibraryDirs = config.LibraryDirs.Prefix(path, items.PathPrefixer)
		}
		fmt.Println("DEPENDENCY:" + dep.Name + " -  packagePath=" + dep.PackagePath + ", projectpath=" + dep.ProjectPath)
	}
}

// CPPprojectID is a GUID that is defining a C++ project in Visual Studio
var CPPprojectID = "8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942"

// generateVisualStudio2015Project generates a Visual Studio 2015 project (.vcxproj) file
func generateVisualStudioProject(prj *denv.Project, v vars.Variables, replacer vars.Replacer, writer denv.ProjectWriter) {
	var platform = prj.Platform

	writer.WriteLn(`<?xml version="1.0" encoding="utf-8"?>`)
	writer.WriteLn(`<Project DefaultTargets="Build" ToolsVersion="14.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">`)

	writer.WriteLn(`+<ItemGroup Label="ProjectConfigurations">`)
	for _, config := range platform.Configs {
		projconfig := []string{}
		projconfig = append(projconfig, `+<ProjectConfiguration Include="${CONFIG}|${PLATFORM}">`)
		projconfig = append(projconfig, `++<Configuration>${CONFIG}</Configuration>`)
		projconfig = append(projconfig, `++<Platform>${PLATFORM}</Platform>`)
		projconfig = append(projconfig, `+</ProjectConfiguration>`)

		replacer.ReplaceInLines("${PLATFORM}", platform.Name, projconfig)
		replacer.ReplaceInLines("${CONFIG}", config.Name, projconfig)
		writer.WriteLns(projconfig)
	}
	writer.WriteLn(`+</ItemGroup>`)

	{
		toolsets := []string{}
		toolsets = append(toolsets, `+<PropertyGroup Condition="'$(Platform)'=='${PLATFORM}'" Label="Configuration">`)
		toolsets = append(toolsets, `++<PlatformToolset>${TOOLSET[${PLATFORM}]}</PlatformToolset>`)
		toolsets = append(toolsets, `+</PropertyGroup>`)

		replacer.ReplaceInLines("${PLATFORM}", platform.Name, toolsets)
		v.ReplaceInLines(replacer, toolsets)
		writer.WriteLns(toolsets)
	}

	globals := []string{}
	globals = append(globals, `+<PropertyGroup Label="Globals">`)
	globals = append(globals, `++<ProjectGuid>${${Name}:GUID}</ProjectGuid>`)
	globals = append(globals, `++<PackageType>${${Name}:TYPE}</PackageType>`)
	globals = append(globals, `++<PackageSignature>$(Configuration)_$(Platform)_$(PlatformToolset)</PackageSignature>`)
	globals = append(globals, `+</PropertyGroup>`)

	replacer.ReplaceInLines("${Name}", prj.Name, globals)
	v.ReplaceInLines(replacer, globals)
	writer.WriteLns(globals)

	projects := []*denv.Project{}
	projects = append(projects, prj)
	projects = append(projects, prj.Dependencies...)

	writer.WriteLn(`+<ImportGroup Label="TargetDirs"/>`)
	writer.WriteLn(`+<Import Project="$(VCTargetsPath)\Microsoft.Cpp.Default.props"/>`)

	{
		for _, config := range platform.Configs {
			configuration := []string{}
			configuration = append(configuration, `+<PropertyGroup Condition="'$(Configuration)|$(Platform)'=='${CONFIG}|${PLATFORM}'" Label="Configuration">`)
			configuration = append(configuration, `++<ConfigurationType>$(PackageType)</ConfigurationType>`)
			configuration = append(configuration, `++<UseDebugLibraries>${USE_DEBUG_LIBS}</UseDebugLibraries>`)
			configuration = append(configuration, `++<CLRSupport>false</CLRSupport>`)
			configuration = append(configuration, `++<CharacterSet>NotSet</CharacterSet>`)
			configuration = append(configuration, `+</PropertyGroup>`)
			varkey := fmt.Sprintf("%s:USE_DEBUG_LIBS[%s][%s]", prj.Name, platform.Name, config.Name)
			usedebuglibs := v.GetVar(varkey)
			if len(usedebuglibs) > 0 {
				replacer.ReplaceInLines("${USE_DEBUG_LIBS}", usedebuglibs, configuration)
			}

			replacer.ReplaceInLines("${PLATFORM}", platform.Name, configuration)
			replacer.ReplaceInLines("${CONFIG}", config.Name, configuration)
			v.ReplaceInLines(replacer, configuration)
			writer.WriteLns(configuration)
		}
	}

	writer.WriteLn(`<Import Project="$(VCTargetsPath)\Microsoft.Cpp.props"/>`)

	{
		userprops := []string{}
		userprops = append(userprops, `+<ImportGroup Condition="'$(Platform)'=='${PLATFORM}'" Label="PropertySheets">`)
		userprops = append(userprops, `++<Import Condition="exists('$(UserRootDir)\Microsoft.Cpp.$(Platform).user.props')" Label="LocalAppDataPlatform" Project="$(UserRootDir)\Microsoft.Cpp.$(Platform).user.props"/>`)
		userprops = append(userprops, `+</ImportGroup>`)
		replacer.ReplaceInLines("${PLATFORM}", platform.Name, userprops)
		writer.WriteLns(userprops)
	}

	{
		platformprops := []string{}
		platformprops = append(platformprops, `+<PropertyGroup Condition="'$(Platform)'=='${PLATFORM}'">`)
		platformprops = append(platformprops, `++<LinkIncremental>true</LinkIncremental>`)
		platformprops = append(platformprops, `++<OutDir>${OUTDIR}</OutDir>`)
		platformprops = append(platformprops, `++<IntDir>${INTDIR}</IntDir>`)
		platformprops = append(platformprops, `<TargetName>${Name}_$(PackageSignature)</TargetName>`)
		platformprops = append(platformprops, `++<ExtensionsToDeleteOnClean>*.obj%3b*.d%3b*.map%3b*.lst%3b*.pch%3b$(TargetPath)</ExtensionsToDeleteOnClean>`)
		platformprops = append(platformprops, `++<GenerateManifest>false</GenerateManifest>`)
		platformprops = append(platformprops, `+</PropertyGroup>`)
		replacer.ReplaceInLines("${Name}", prj.Name, platformprops)
		replacer.ReplaceInLines("${PLATFORM}", platform.Name, platformprops)

		configitems := map[string]string{
			"INTDIR": "",
			"OUTDIR": "",
		}
		for configitem := range configitems {
			varkeystr := fmt.Sprintf("${%s}", configitem)
			varkey := fmt.Sprintf("%s:%s", prj.Name, configitem)
			varitem := v.GetVar(varkey)
			replacer.ReplaceInLines(varkeystr, varitem, platformprops)
		}

		writer.WriteLns(platformprops)
	}

	// includedirs := []string{}
	// for _, project := range projects {
	// 	includedir := "${${Name}:IncludeDir}"
	// 	includedir = replacer.ReplaceInLine("${Name}", project.Name, includedir)
	// 	includedir = v.ReplaceInLine(replacer, includedir)
	// 	includedirs = append(includedirs, includedir)
	// }

	{
		for _, config := range platform.Configs {
			compileandlink := []string{}
			compileandlink = append(compileandlink, `+<ItemDefinitionGroup Condition="'$(Configuration)|$(Platform)'=='${CONFIG}|${PLATFORM}'">`)
			compileandlink = append(compileandlink, `++<ClCompile>`)
			compileandlink = append(compileandlink, `+++<PreprocessorDefinitions>${DEFINES}</PreprocessorDefinitions>`)
			compileandlink = append(compileandlink, `+++<AdditionalIncludeDirectories>${INCLUDE_DIRS}</AdditionalIncludeDirectories>`)
			compileandlink = append(compileandlink, `+++<WarningLevel>Level3</WarningLevel>`)
			compileandlink = append(compileandlink, `+++<Optimization>${OPTIMIZATION}</Optimization>`)
			compileandlink = append(compileandlink, `+++<PrecompiledHeader>NotUsing</PrecompiledHeader>`)
			compileandlink = append(compileandlink, `+++<ExceptionHandling>${EXCEPTIONS}</ExceptionHandling>`)
			//compileandlink = append(compileandlink, `+++<ObjectFileName>$(IntDir)%(RelativeDir)</ObjectFileName>`)
			compileandlink = append(compileandlink, `+++<CompileAs>${COMPILE_AS}</CompileAs>`)
			compileandlink = append(compileandlink, `+++<MultiProcessorCompilation>true</MultiProcessorCompilation>`)
			compileandlink = append(compileandlink, `+++<MinimalRebuild>false</MinimalRebuild>`)
			compileandlink = append(compileandlink, `++</ClCompile>`)
			compileandlink = append(compileandlink, `++<Link>`)
			compileandlink = append(compileandlink, `+++<GenerateDebugInformation>${DEBUG_INFO}</GenerateDebugInformation>`)
			compileandlink = append(compileandlink, `+++<AdditionalDependencies>${LIBRARY_FILES}</AdditionalDependencies>`)
			compileandlink = append(compileandlink, `+++<AdditionalLibraryDirectories>${LIBRARY_DIRS}</AdditionalLibraryDirectories>`)
			compileandlink = append(compileandlink, `+++<SubSystem>Console</SubSystem>`)
			compileandlink = append(compileandlink, `++</Link>`)
			compileandlink = append(compileandlink, `++<Lib>`)
			compileandlink = append(compileandlink, `+++<OutputFile>$(OutDir)\$(TargetName)$(TargetExt)</OutputFile>`)
			compileandlink = append(compileandlink, `++</Lib>`)
			compileandlink = append(compileandlink, `+</ItemDefinitionGroup>`)

			libraries := []string{}
			for _, lib := range prj.LibraryFiles.Items {
				libraries = append(libraries, lib)
			}
			for _, depproject := range prj.Dependencies {
				if depproject.HasConfig(platform.Name, config.Name) {
					depconfig := depproject.GetConfig(config.Name)
					if depconfig != nil {
						libraries = append(libraries, depconfig.LibraryFile)
					}
				}
			}
			replacer.InsertInLines("${LIBRARY_FILES}", strings.Join(libraries, ";"), ";", compileandlink)

			configitems := map[string]string{
				"DEFINES":       "%(PreprocessorDefinitions)",
				"INCLUDE_DIRS":  "%(AdditionalIncludeDirectories)",
				"LIBRARY_DIRS":  "%(AdditionalLibraryDirectories)",
				"LIBRARY_FILES": "%(AdditionalLibraryFiles)",
			}

			for configitem, defaults := range configitems {
				varkeystr := fmt.Sprintf("${%s}", configitem)
				varlist := items.NewList("", ";", "")
				for _, project := range projects {
					varkey := fmt.Sprintf("%s:%s[%s][%s]", project.Name, configitem, platform.Name, config.Name)
					varitem := v.GetVar(varkey)
					if len(varitem) > 0 {
						varlist = varlist.Add(varitem)
					}
				}
				varset := items.ListToSet(varlist)
				varset = varset.Add(defaults)
				replacer.InsertInLines(varkeystr, varset.String(), ";", compileandlink)
				replacer.ReplaceInLines(varkeystr, "", compileandlink)
			}

			compileandlinkReplacer := func(key string, value string) {
				key = vars.MakeVarKey(key)
				replacer.ReplaceInLines(key, value, compileandlink)
			}

			fullReplaceVar("OPTIMIZATION", prj.Name, platform.Name, config.Name, v, compileandlinkReplacer)
			fullReplaceVar("EXCEPTIONS", prj.Name, platform.Name, config.Name, v, compileandlinkReplacer)
			fullReplaceVarWithDefault("COMPILE_AS", "Default", prj.Name, platform.Name, config.Name, v, compileandlinkReplacer)
			fullReplaceVar("DEBUG_INFO", prj.Name, platform.Name, config.Name, v, compileandlinkReplacer)

			replacer.ReplaceInLines("${Name}", prj.Name, compileandlink)
			replacer.ReplaceInLines("${PLATFORM}", platform.Name, compileandlink)
			replacer.ReplaceInLines("${CONFIG}", config.Name, compileandlink)
			v.ReplaceInLines(replacer, compileandlink)
			writer.WriteLns(compileandlink)
		}
	}

	relpath, _ := filepath.Rel(prj.ProjectPath, prj.PackagePath)

	if len(prj.SrcFiles.Files) > 0 {
		writer.WriteLn("+<ItemGroup>")
		for _, srcfile := range prj.SrcFiles.Files {
			srcfile = filepath.Join(relpath, srcfile)
			if strings.HasSuffix(srcfile, ".c") {
				writer.WriteLn(fmt.Sprintf("++<ClCompile Include=\"%s\">", srcfile))
				writer.WriteLn(`+++<CompileAs>CompileAsC</CompileAs>`)
				writer.WriteLn("++</ClCompile>")
			} else if strings.HasSuffix(srcfile, ".cpp") {
				writer.WriteLn(fmt.Sprintf("++<ClCompile Include=\"%s\" />", srcfile))
			} else {
				writer.WriteLn(fmt.Sprintf("++<None Include=\"%s\" />", srcfile))
			}
		}
		writer.WriteLn("+</ItemGroup>")
	}

	if len(prj.HdrFiles.Files) > 0 {
		writer.WriteLn("+<ItemGroup>")
		for _, hdrfile := range prj.HdrFiles.Files {
			hdrfile = filepath.Join(relpath, hdrfile)
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

func fullReplaceVar(varname string, prjname string, platform string, config string, v vars.Variables, replacer func(name, value string)) bool {
	value := v.GetVar(fmt.Sprintf("%s:%s[%s][%s]", prjname, varname, platform, config))
	if len(value) > 0 {
		replacer(varname, value)
	} else {
		value = v.GetVar(fmt.Sprintf("%s:%s", prjname, varname))
		if len(value) > 0 {
			replacer(varname, value)
		} else {
			return false
		}
	}
	return true
}

func fullReplaceVarWithDefault(varname string, vardefaultvalue string, prjname string, platform string, config string, v vars.Variables, replacer func(name, value string)) {
	if !fullReplaceVar(varname, prjname, platform, config, v, replacer) {
		replacer(varname, vardefaultvalue)
	}
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
		for {
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
func generateVisualStudioProjectFilters(prj *denv.Project, writer denv.ProjectWriter, dev denv.DevEnum) {
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

	relpath, _ := filepath.Rel(prj.ProjectPath, prj.PackagePath)

	writer.WriteLn("+<ItemGroup>")
	for hdrfile, vhdrfile := range includes {
		hdrfile = filepath.Join(relpath, hdrfile)
		writer.WriteLn(fmt.Sprintf("++<ClInclude Include=\"%s\">", hdrfile))
		writer.WriteLn(fmt.Sprintf("+++<Filter>%s</Filter>", vhdrfile))
		writer.WriteLn("++</ClInclude>")
	}
	writer.WriteLn("+</ItemGroup>")

	writer.WriteLn("+<ItemGroup>")
	for srcfile, vsrcfile := range cpp {
		srcfile = filepath.Join(relpath, srcfile)
		writer.WriteLn(fmt.Sprintf("++<ClCompile Include=\"%s\">", srcfile))
		writer.WriteLn(fmt.Sprintf("+++<Filter>%s</Filter>", vsrcfile))
		writer.WriteLn("++</ClCompile>")
	}
	writer.WriteLn("+</ItemGroup>")

	writer.WriteLn("</Project>")
}

type strStack []string

func (s strStack) Empty() bool    { return len(s) == 0 }
func (s strStack) Peek() string   { return s[len(s)-1] }
func (s *strStack) Push(i string) { (*s) = append((*s), i) }
func (s *strStack) Pop() string {
	d := (*s)[len(*s)-1]
	(*s) = (*s)[:len(*s)-1]
	return d
}

// GenerateVisualStudio2015Solution generates a Visual Studio 2015 solution (.sln) file
// for the current project. It also generates the project file for the current projects
// as well as the project files for the dependencies.
func GenerateVisualStudioSolution(p__ *denv.Project, dev denv.DevEnum) {

	mainprj := p__

	writer := &denv.ProjectTextWriter{}
	slnfilepath := filepath.Join(mainprj.ProjectPath, mainprj.Name+".sln")
	if writer.Open(slnfilepath) != nil {
		fmt.Printf("Error opening file '%s'", slnfilepath)
		return
	}

	if dev == denv.VS2015 {
		writer.WriteLn("Microsoft Visual Studio Solution File, Format Version 12.00")
		writer.WriteLn("# Visual Studio 14")
		writer.WriteLn("VisualStudioVersion = 14.0.24720.0")
		writer.WriteLn("MinimumVisualStudioVersion = 10.0.40219.1")
	} else if dev == denv.VS2017 {
		writer.WriteLn("Microsoft Visual Studio Solution File, Format Version 12.00")
		writer.WriteLn("# Visual Studio 15")
		writer.WriteLn("VisualStudioVersion = 15.0.28307.1234")
		writer.WriteLn("MinimumVisualStudioVersion = 10.0.40219.1")
	} else if dev == denv.VS2019 {
		writer.WriteLn("Microsoft Visual Studio Solution File, Format Version 12.00")
		writer.WriteLn("# Visual Studio Version 17")
		writer.WriteLn("VisualStudioVersion = 17.5.33414.496")
		writer.WriteLn("MinimumVisualStudioVersion = 10.0.40219.1")
	} else if dev == denv.VS2022 {
		writer.WriteLn("Microsoft Visual Studio Solution File, Format Version 12.00")
		writer.WriteLn("# Visual Studio Version 17")
		writer.WriteLn("VisualStudioVersion = 17.5.33414.496")
		writer.WriteLn("MinimumVisualStudioVersion = 10.0.40219.1")
	}

	// Write Projects and their dependency information
	//
	//          Project("{8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942}") = "chash", "source\main\cpp\xhash.vcxproj", "{04AB9C6F-0B84-4111-A772-53C03F5CB3C2}"
	//          	ProjectSection(ProjectDependencies) = postProject
	//          		{B83DA73D-6E7B-458D-A6C7-87013421D360} = {B83DA73D-6E7B-458D-A6C7-87013421D360}
	//          	EndProjectSection
	//          EndProject
	//

	// Add dependency projects (dependency tree)
	depmap := map[string]*denv.Project{}
	depmap[mainprj.Name] = mainprj
	depstack := &strStack{mainprj.Name}
	for !depstack.Empty() {
		prjname := depstack.Pop()
		prj := depmap[prjname]
		for _, dep := range prj.Dependencies {
			if _, ok := depmap[dep.Name]; !ok {
				depstack.Push(dep.Name)
				depmap[dep.Name] = dep
			}
		}
	}
	delete(depmap, mainprj.Name)

	// Make every project unique (there are duplicates in the tree)
	for _, prj := range depmap {
		for i, dep := range prj.Dependencies {
			prj.Dependencies[i] = depmap[dep.Name]
		}
	}

	dependencies := []*denv.Project{}
	for _, dep := range depmap {
		dependencies = append(dependencies, dep)
	}

	// -----------------------------------------------------------------------------------------------------------
	// Every project has dependencies, however nested dependencies should also be added to the top level project
	depstack.Push(mainprj.Name)
	for _, dep := range dependencies {
		depstack.Push(dep.Name)
	}

	for !depstack.Empty() {
		currentProjectName := depstack.Pop()
		var currentProject *denv.Project
		if currentProjectName == mainprj.Name {
			currentProject = mainprj
		} else {
			currentProject = depmap[currentProjectName]
		}

		// Build the full map of dependencies for this project
		currentProjectDepMap := map[string]*denv.Project{}

		// Initialize the stack with the dependencies of this project
		currentDepStack := &strStack{}
		for _, dep := range currentProject.Dependencies {
			currentDepStack.Push(dep.Name)
		}

		// Now handle the map of current dependencies until it is empty
		for !currentDepStack.Empty() {
			currentDepName := currentDepStack.Pop()
			currentDep := depmap[currentDepName]

			if _, ok := currentProjectDepMap[currentDepName]; !ok {
				// This dependency is not yet a dependency of the current project, add it
				currentProjectDepMap[currentDepName] = currentDep

				// Iterate over the dependencies of this dependency project and schedule a dependency when
				// it has not been recognized as being part of the dependencies of currentProject
				for _, dep := range currentDep.Dependencies {
					if _, ok := currentProjectDepMap[dep.Name]; !ok {
						currentDepStack.Push(dep.Name)
					}
				}
			}
		}

		// Empty the list of dependencies of this project, and then add all of the dependencies that where found
		if len(currentProject.Dependencies) < len(currentProjectDepMap) {
			currentProject.Dependencies = currentProject.Dependencies[:0]
			for _, dep := range currentProjectDepMap {
				currentProject.Dependencies = append(currentProject.Dependencies, dep)
			}
		}
	}
	// -----------------------------------------------------------------------------------------------------------

	variables := vars.NewVars()
	replacer := vars.NewReplacer()

	for _, prj := range dependencies {
		initProjectVariables(prj, variables, replacer)
	}
	initProjectVariables(mainprj, variables, replacer)
	setupProjectPaths(mainprj, dependencies)

	for _, prj := range dependencies {
		mergeProjectDependency(prj)
	}
	mergeProjectDependency(mainprj)

	// Main project
	projects := []*denv.Project{mainprj}
	projects = append(projects, dependencies...)

	for _, prj := range projects {
		addProjectVariables(prj, variables, replacer, dev)
	}

	variables.Print()

	// Glob all the source and header files for every project
	for _, prj := range projects {
		fmt.Println("GLOBBING: " + prj.Name + " : " + prj.PackagePath + " : ignore(" + strings.Join(prj.Platform.FilePatternsToIgnore, ", ") + ")")
		prj.SrcFiles.GlobFiles(prj.PackagePath, prj.Platform.FilePatternsToIgnore)
		prj.HdrFiles.GlobFiles(prj.PackagePath, prj.Platform.FilePatternsToIgnore)
	}

	// Generate all the projects
	for _, prj := range projects {
		// Generate the project file
		prjwriter := &denv.ProjectTextWriter{}
		prjwriter.Open(filepath.Join(prj.ProjectPath, prj.Name+".vcxproj"))
		generateVisualStudioProject(prj, variables, replacer, prjwriter)
		prjwriter.Close()

		// Generate the project filters file
		prjwriter = &denv.ProjectTextWriter{}
		prjwriter.Open(filepath.Join(prj.ProjectPath, prj.Name+".vcxproj.filters"))
		generateVisualStudioProjectFilters(prj, prjwriter, dev)
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
		var pp = project.Platform
		{
			for _, config := range pp.Configs {
				configstr := fmt.Sprintf("%s|%s", config.Name, pp.Name)
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
		var pp = project.Platform
		{
			for _, config := range pp.Configs {
				configplatform := fmt.Sprintf("%s|%s", config.Name, pp.Name)
				activecfg := fmt.Sprintf("++{%s}.%s|%s.ActiveCfg = %s|%s", project.GUID, config.Name, configplatform, config.Name, configplatform)
				buildcfg := fmt.Sprintf("++{%s}.%s|%s.Build.0 = %s|%s", project.GUID, config.Name, configplatform, config.Name, configplatform)
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
