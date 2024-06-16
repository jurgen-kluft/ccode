package ide

import (
	"path/filepath"

	. "github.com/jurgen-kluft/ccode/axe"
)

const (
	xcode_build_phase_sources_uuid    = "B0000000B0000000B0000002"
	xcode_build_phase_frameworks_uuid = "B0000000B0000000B0000003"
	xcode_build_phase_headers_uuid    = "B0000000B0000000B0000004"
	xcode_build_phase_resources_uuid  = "B0000000B0000000B0000005"
	xcode_main_group_uuid             = "C0000000B0000000B0000001"
	xcode_product_group_uuid          = "C0000000B0000000B0000002"
	xcode_dependencies_group_uuid     = "C0000000B0000000B0000003"
	xcode_resources_group_uuid        = "C0000000B0000000B0000004"
	xcode_kSourceTreeProject          = "SOURCE_ROOT"
	xcode_kSourceTreeGroup            = "\"<group>\""
	xcode_kSourceTreeAbsolute         = "\"<absolute>\""
)

type XcodeGenerator struct {
	LastGenId UUID
	Workspace *Workspace
}

func NewXcodeGenerator(ws *Workspace) *XcodeGenerator {
	return &XcodeGenerator{
		LastGenId: GenerateUUID(),
		Workspace: ws,
	}
}

func (g *XcodeGenerator) Init(ws *Workspace) {
	if ws.MakeTarget == nil {
		ws.MakeTarget = NewDefaultMakeTarget()
	}
}

func (g *XcodeGenerator) QuoteString2(v string) string {
	return g.QuoteString(g.QuoteString(v))
}

func (g *XcodeGenerator) QuoteString(v string) string {
	o := "\""
	for _, ch := range v {
		switch ch {
		case '"':
			o += "\\\""
		case '\\':
			o += "\\\\"
		default:
			o += string(ch)
		}
	}
	o += "\""
	return o
}

func (g *XcodeGenerator) Generate() {
	g.genWorkSpace()
}

func (g *XcodeGenerator) genWorkSpace() {
	xcodeWorkspace := filepath.Join(g.Workspace.GenerateAbsPath, g.Workspace.WorkspaceName+".xcworkspace")

	for _, proj := range g.Workspace.ProjectList.Values {
		g.genProjectGenUuid(proj)
	}

	for _, proj := range g.Workspace.ProjectList.Values {
		g.genProject(proj)
	}

	wr := NewXmlWriter()
	{
		tag := wr.TagScope("Workspace")
		defer tag.Close()
		wr.Attr("version", "1.0")
		g.genWorkspaceGroup(wr, g.Workspace.ProjectGroups.Root)
	}

	wr.WriteToFile(filepath.Join(xcodeWorkspace, "contents.xcworkspacedata"))
}

func (g *XcodeGenerator) genWorkspaceGroup(wr *XmlWriter, group *ProjectGroup) {
	for _, c := range group.Children {
		tag := wr.TagScope("Group")
		wr.Attr("location", "container:")
		basename := PathBasename(c.Path, true)
		wr.Attr("name", basename)
		g.genWorkspaceGroup(wr, c)
		tag.Close()
	}

	for _, proj := range group.Projects {
		tag := wr.TagScope("FileRef")
		wr.Attr("location", "container:"+proj.GenDataXcode.XcodeProj.Path)
		tag.Close()
	}
}

func (g *XcodeGenerator) genProjectGenUuid(proj *Project) {
	gd := NewXcodeProjectConfig()
	gd.XcodeProj.Init(g.Workspace.GenerateAbsPath+proj.Name+".xcodeproj", true)
	gd.PbxProj = filepath.Join(gd.XcodeProj.Path, "project.pbxproj")
	gd.Uuid = GenerateUUID()
	gd.TargetUuid = GenerateUUID()
	gd.TargetProductUuid = GenerateUUID()
	gd.ConfigListUuid = GenerateUUID()
	gd.TargetConfigListUuid = GenerateUUID()
	gd.DependencyProxyUuid = GenerateUUID()
	gd.DependencyTargetUuid = GenerateUUID()
	gd.DependencyTargetProxyUuid = GenerateUUID()

	for _, i := range proj.FileEntries.Dict {
		f := proj.FileEntries.Values[i]
		f.GenDataXcode.UUID = GenerateUUID()
		f.GenDataXcode.BuildUUID = GenerateUUID()
	}

	for _, i := range proj.ResourceDirs.Dict {
		f := proj.FileEntries.Values[i]
		f.GenDataXcode.UUID = GenerateUUID()
		f.GenDataXcode.BuildUUID = GenerateUUID()
	}

	for _, f := range proj.VirtualFolders.Folders {
		f.GenData_xcode.UUID = GenerateUUID()
	}

	for _, config := range proj.Configs {
		config.GenDataXcode.ProjectConfigUuid = GenerateUUID()
		config.GenDataXcode.TargetUuid = GenerateUUID()
		config.GenDataXcode.TargetConfigUuid = GenerateUUID()
	}
}

func (g *XcodeGenerator) genProject(proj *Project) {
	if proj.TypeIsExeOrDll() {
		if g.Workspace.MakeTarget.OSIsIos() {
			//g.GenInfoPlistIOS(proj)
		} else {
			g.genInfoPlistMacOSX(proj)
		}
	}

	wr := NewXcodeWriter()
	wr.write("// !$*UTF8*$!")
	{
		scope := wr.NewObjectScope("")
		defer scope.Close()

		wr.member("archiveVersion", "1")
		{
			scope := wr.NewObjectScope("classes")
			defer scope.Close()
		}
		wr.member("objectVersion", "46")
		{
			scope := wr.NewObjectScope("objects")
			defer scope.Close()

			g.genProjectPBXBuildFile(wr, proj)
			g.genProjectDependencies(wr, proj)
			g.genProjectPBXGroup(wr, proj)
			g.genProjectPBXProject(wr, proj)
			g.genProjectPBXSourcesBuildPhase(wr, proj)
			g.genProjectPBXResourcesBuildPhase(wr, proj)
			g.genProjectPBXNativeTarget(wr, proj)
			g.genProjectXCBuildConfiguration(wr, proj)
			g.genProjectXCConfigurationList(wr, proj)
		}
		wr.member("rootObject", proj.GenDataXcode.Uuid.String())
	}

	filename := proj.GenDataXcode.PbxProj
	WriteTextFile(filename, wr.Buffer.String())
}

func (g *XcodeGenerator) genBuildFileReference(wr *XcodeWriter, f *FileEntry) {
	wr.newline(0)
	wr.commentBlock(f.Path)
	scope := wr.NewObjectScope(f.GenDataXcode.BuildUUID.String())
	{
		wr.member("isa", "PBXBuildFile")
		wr.member("fileRef", f.GenDataXcode.UUID.String())
	}
	scope.Close()
}

func (g *XcodeGenerator) genFileReference(wr *XcodeWriter, proj *Project, f *FileEntry) {
	wr.newline(0)
	wr.commentBlock(f.Path)

	scope := wr.NewObjectScope(f.GenDataXcode.UUID.String())
	basename := PathBasename(f.Path, true)
	{
		wr.member("isa", "PBXFileReference")
		wr.member("name", g.QuoteString(basename))
		wr.member("path", g.QuoteString(basename))
		wr.member("sourceTree", xcode_kSourceTreeGroup)

		explicitFileType := ""
		switch f.Type {
		case FileTypeCppSource:
			if proj.Settings.CppAsObjCpp {
				explicitFileType = "sourcecode.cpp.objcpp"
			} else {
				explicitFileType = "sourcecode.cpp.cpp"
			}
		case FileTypeCSource:
			if proj.Settings.CppAsObjCpp {
				explicitFileType = "sourcecode.c.objc"
			} else {
				explicitFileType = "sourcecode.c.c"
			}
		}

		if len(explicitFileType) > 0 {
			wr.member("explicitFileType", explicitFileType)
		}
	}
	scope.Close()
}

func (g *XcodeGenerator) genProjectDependencies(wr *XcodeWriter, proj *Project) {
	wr.newline(0)
	wr.commentBlock("----- project dependencies -----------------")
	for _, dp := range proj.DependenciesInherit.Values {
		if !dp.HasOutputTarget {
			continue
		}

		targetBasename := PathBasename(dp.GenDataXcode.XcodeProj.Path, true)

		wr.newline(0)
		wr.commentBlock(dp.Name)
		{
			wr.newline(0)
			wr.commentBlock("PBXContainerItemProxy for xcodeproject")
			scope := wr.NewObjectScope(dp.GenDataXcode.DependencyProxyUuid.String())
			{
				wr.member("isa", "PBXContainerItemProxy")
				wr.member("containerPortal", dp.GenDataXcode.Uuid.String())
				wr.member("proxyType", "2")
				wr.member("remoteInfo", g.QuoteString(dp.Name))
			}
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("PBXContainerItemProxy for PBXTargetDependency")
			scope := wr.NewObjectScope(dp.GenDataXcode.DependencyTargetProxyUuid.String())
			{
				wr.member("isa", "PBXContainerItemProxy")
				wr.member("containerPortal", dp.GenDataXcode.Uuid.String())
				wr.member("proxyType", "1")
				wr.member("remoteInfo", g.QuoteString(dp.Name))
			}
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("PBXTargetDependency")
			scope := wr.NewObjectScope(dp.GenDataXcode.DependencyTargetUuid.String())
			{
				wr.member("isa", "PBXTargetDependency")
				wr.member("name", g.QuoteString(dp.Name))
				wr.member("targetProxy", dp.GenDataXcode.DependencyTargetProxyUuid.String())
			}
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("PBXFileReference")
			scope := wr.NewObjectScope(dp.GenDataXcode.Uuid.String())
			{
				wr.member("isa", "PBXFileReference")
				wr.member("name", g.QuoteString(targetBasename))
				wr.member("path", g.QuoteString(dp.GenDataXcode.XcodeProj.Path))
				wr.member("sourceTree", xcode_kSourceTreeAbsolute)
			}
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("------ Folder dependencies")
			scope := wr.NewObjectScope(xcode_dependencies_group_uuid)
			wr.member("isa", "PBXGroup")
			{
				scope := wr.NewArrayScope("children")
				for _, dp := range proj.DependenciesInherit.Values {
					wr.write(dp.GenDataXcode.Uuid.String())
				}
				scope.Close()
			}
			wr.member("sourceTree", xcode_kSourceTreeGroup)
			wr.member("name", "_dependencies_")
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("------ Folder resources")
			scope := wr.NewObjectScope(xcode_resources_group_uuid)
			wr.member("isa", "PBXGroup")
			{
				scope := wr.NewArrayScope("children")
				for _, i := range proj.ResourceDirs.Dict {
					f := proj.ResourceDirs.Values[i]
					wr.write(f.GenDataXcode.UUID.String())
				}
				scope.Close()
			}
			wr.member("sourceTree", xcode_kSourceTreeGroup)
			wr.member("name", "_resources_")
			scope.Close()
		}
	}
}

func (g *XcodeGenerator) genProjectPBXBuildFile(wr *XcodeWriter, proj *Project) {
	wr.newline(0)
	wr.commentBlock("------ Begin PBXBuildFile section")

	for _, i := range proj.FileEntries.Dict {
		f := proj.FileEntries.Values[i]
		if f.ExcludedFromBuild {
			continue
		}
		g.genBuildFileReference(wr, f)
	}

	for _, i := range proj.ResourceDirs.Dict {
		f := proj.ResourceDirs.Values[i]
		g.genBuildFileReference(wr, f)
	}

	wr.newline(0)
	wr.commentBlock("------ End PBXBuildFile section")

	//-------
	wr.newline(0)
	wr.newline(0)
	wr.commentBlock("------ Begin PBXFileReference section")

	for _, i := range proj.FileEntries.Dict {
		f := proj.FileEntries.Values[i]
		g.genFileReference(wr, proj, f)
	}

	for _, i := range proj.ResourceDirs.Dict {
		f := proj.ResourceDirs.Values[i]
		wr.newline(0)
		wr.commentBlock(f.Path)
		scope := wr.NewObjectScope(f.GenDataXcode.UUID.String())
		{
			basename := PathBasename(f.Path, true)
			relPath := PathGetRel(f.Path, g.Workspace.GenerateAbsPath)

			wr.member("isa", "PBXFileReference")
			wr.member("name", g.QuoteString(basename))
			wr.member("path", relPath)
			wr.member("sourceTree", xcode_kSourceTreeProject)
		}
		scope.Close()
	}

	wr.newline(0)
	wr.commentBlock("------ End PBXFileReference section")

	//-------
	wr.newline(0)
	wr.newline(0)
	wr.commentBlock("----- PBXFrameworksBuildPhase -----------------")
	wr.newline(0)
	wr.commentBlock(" Product Group ")

	{
		scope := wr.NewObjectScope(xcode_build_phase_frameworks_uuid)
		wr.member("isa", "PBXFrameworksBuildPhase")
		{
			scope := wr.NewArrayScope("file")
			scope.Close()
		}
		wr.member("runOnlyForDeploymentPostprocessing", "0")
		scope.Close()
	}
}

func (g *XcodeGenerator) genProjectPBXGroup(wr *XcodeWriter, proj *Project) {
	wr.newline(0)
	wr.commentBlock("------ Begin PBXGroup section")

	{
		wr.newline(0)
		wr.commentBlock("------ Folder Product")
		scope := wr.NewObjectScope(xcode_product_group_uuid)
		{
			wr.member("isa", "PBXGroup")
			scope := wr.NewArrayScope("children")
			if proj.HasOutputTarget {
				wr.write(proj.GenDataXcode.TargetProductUuid.String())
			}
			scope.Close()
			wr.member("sourceTree", xcode_kSourceTreeGroup)
			wr.member("name", "_products_")
		}
		scope.Close()
	}

	root := proj.VirtualFolders.Root
	for _, v := range proj.VirtualFolders.Folders {
		if v == root {
			continue
		}

		wr.newline(0)
		wr.commentBlock(v.Path)
		scope := wr.NewObjectScope(v.GenData_xcode.UUID.String())
		{
			wr.member("isa", "PBXGroup")
			scope := wr.NewArrayScope("children")
			for _, c := range v.Children {
				wr.write(c.GenData_xcode.UUID.String())
			}
			for _, f := range v.Files {
				wr.write(f.GenDataXcode.UUID.String())
			}
			scope.Close()

			basename := PathBasename(v.Path, true)
			wr.member("name", g.QuoteString(basename))

			if v.Parent == root {
				wr.member("sourceTree", xcode_kSourceTreeProject)
				relPath := PathGetRel(v.DiskPath, g.Workspace.GenerateAbsPath)
				wr.member("path", g.QuoteString(relPath))
			} else {
				wr.member("sourceTree", xcode_kSourceTreeGroup)
				wr.member("path", g.QuoteString(basename))
			}
		}
		scope.Close()
	}

	{
		scope := wr.NewObjectScope(xcode_main_group_uuid)
		wr.member("isa", "PBXGroup")
		{
			scope := wr.NewArrayScope("children")
			wr.write(xcode_product_group_uuid)
			wr.write(xcode_dependencies_group_uuid)
			wr.write(xcode_resources_group_uuid)
			for _, c := range root.Children {
				wr.write(c.GenData_xcode.UUID.String())
			}
			for _, f := range root.Files {
				wr.write(f.GenDataXcode.UUID.String())
			}
			scope.Close()
		}
		wr.member("sourceTree", xcode_kSourceTreeProject)
		relPath := PathGetRel(proj.GenerateAbsPath, g.Workspace.GenerateAbsPath)
		wr.member("path", g.QuoteString(relPath))
		wr.member("name", "MainGroup")
		scope.Close()
	}

	wr.newline(0)
	wr.newline(0)
	wr.commentBlock("------ End PBXGroup section")
}

func (g *XcodeGenerator) genProjectPBXProject(wr *XcodeWriter, proj *Project) {
	wr.newline(0)
	wr.commentBlock("------ Begin PBXProject section")

	scope := wr.NewObjectScope(proj.GenDataXcode.Uuid.String())
	{
		wr.member("isa", "PBXProject")
		{
			scope := wr.NewObjectScope("attributes")
			{
				scope := wr.NewObjectScope("TargetAttributes")
				{
					scope := wr.NewObjectScope(proj.GenDataXcode.TargetUuid.String())
					wr.member("CreatedOnToolsVersion", "7.3.1")
					wr.member("ProvisioningStyle", "Automatic")
					scope.Close()
				}
				scope.Close()
			}
			scope.Close()
		}
		wr.member("compatibilityVersion", g.QuoteString("Xcode 3.2"))
		wr.member("developmentRegion", "en")
		wr.member("hasScannedForEncodings", "0")
		wr.member("knownRegions", "(Base, en,)")
		wr.member("buildConfigurationList", proj.GenDataXcode.ConfigListUuid.String())
		wr.member("mainGroup", xcode_main_group_uuid)
		wr.member("productRefGroup", xcode_product_group_uuid)
		wr.member("projectDirPath", g.QuoteString(""))
		wr.member("projectRoot", g.QuoteString(""))
		{
			scope := wr.NewArrayScope("targets")
			wr.write(proj.GenDataXcode.TargetUuid.String())
			scope.Close()
		}
	}
	scope.Close()

	wr.newline(0)
	wr.commentBlock("------ End PBXProject section")
}

func (g *XcodeGenerator) genProjectPBXSourcesBuildPhase(wr *XcodeWriter, proj *Project) {
	wr.newline(0)
	wr.commentBlock("------ PBXSourcesBuildPhase section")

	scope := wr.NewObjectScope(xcode_build_phase_sources_uuid)
	{
		wr.member("isa", "PBXSourcesBuildPhase")
		wr.member("buildActionMask", "2147483647")
		wr.member("runOnlyForDeploymentPostprocessing", "0")
		{
			scope := wr.NewArrayScope("files")
			for _, i := range proj.FileEntries.Dict {
				f := proj.FileEntries.Values[i]
				if f.ExcludedFromBuild {
					continue
				}
				wr.write(f.GenDataXcode.BuildUUID.String())
			}
			scope.Close()
		}
	}
	scope.Close()
}

func (g *XcodeGenerator) genProjectPBXResourcesBuildPhase(wr *XcodeWriter, proj *Project) {
	wr.newline(0)
	wr.commentBlock("------ PBXResourcesBuildPhase section")

	scope := wr.NewObjectScope(xcode_build_phase_resources_uuid)
	{
		wr.member("isa", "PBXResourcesBuildPhase")
		wr.member("buildActionMask", "2147483647")
		wr.member("runOnlyForDeploymentPostprocessing", "0")
		{
			scope := wr.NewArrayScope("files")
			for _, i := range proj.ResourceDirs.Dict {
				f := proj.ResourceDirs.Values[i]
				wr.write(f.GenDataXcode.BuildUUID.String())
			}
			scope.Close()
		}
	}
	scope.Close()
}

func (g *XcodeGenerator) genProjectPBXNativeTarget(wr *XcodeWriter, proj *Project) {
	if !proj.HasOutputTarget {
		return
	}

	productType := ""
	if proj.TypeIsExe() {
		if proj.Settings.IsGuiApp || g.Workspace.MakeTarget.OSIsIos() {
			productType = ProductTypeApplication.String()
		} else {
			productType = ProductTypeTool.String()
		}
	} else if proj.TypeIsDll() {
		productType = string(ProductTypeDynamicLib)
	} else if proj.TypeIsLib() {
		productType = string(ProductTypeStaticLib)
	} else {
		panic("Unsupported project type")
	}

	wr.newline(0)
	wr.commentBlock("--------- PBXNativeTarget section")

	{
		wr.newline(0)
		wr.commentBlock("--------- File Reference")
		scope := wr.NewObjectScope(proj.GenDataXcode.TargetProductUuid.String())
		{
			explicitFileType := ""
			if proj.TypeIsLib() {
				explicitFileType = "archive.ar"
			} else {
				explicitFileType = "compiled.mach-o.executable"
			}

			wr.member("isa", "PBXFileReference")
			wr.member("explicitFileType", explicitFileType)
			wr.member("includeInIndex", "0")

			targetBasename := PathBasename(proj.GetDefaultConfig().OutputTarget.Path, true)
			wr.member("path", g.QuoteString(targetBasename))
			wr.member("sourceTree", "BUILT_PRODUCTS_DIR")
		}
		scope.Close()
	}

	{
		wr.newline(0)
		wr.commentBlock("--------- Target")
		scope := wr.NewObjectScope(proj.GenDataXcode.TargetUuid.String())
		{
			wr.member("isa", "PBXNativeTarget")
			wr.member("name", g.QuoteString(proj.Name))
			wr.member("productName", g.QuoteString(proj.Name))
			wr.member("productReference", proj.GenDataXcode.TargetProductUuid.String())
			wr.member("productType", productType)
			wr.member("buildConfigurationList", proj.GenDataXcode.TargetConfigListUuid.String())
			{
				scope := wr.NewArrayScope("buildPhases")
				wr.write(xcode_build_phase_sources_uuid)
				wr.write(xcode_build_phase_frameworks_uuid)
				wr.write(xcode_build_phase_headers_uuid)
				wr.write(xcode_build_phase_resources_uuid)
				scope.Close()
			}
			{
				scope := wr.NewArrayScope("buildRules")
				scope.Close()
			}
			{
				scope := wr.NewArrayScope("dependencies")
				if proj.TypeIsExeOrDll() {
					for _, dp := range proj.DependenciesInherit.Values {
						wr.write(dp.GenDataXcode.DependencyTargetUuid.String())
					}
				}
				scope.Close()
			}
		}
		scope.Close()
	}
}

func (g *XcodeGenerator) genProjectXCBuildConfiguration(wr *XcodeWriter, proj *Project) {
	wr.newline(0)
	wr.commentBlock("----- XCBuildConfiguration ---------------")
	wr.newline(0)

	for _, config := range proj.Configs {
		wr.newline(0)
		wr.commentBlock("Project Conifg [" + config.Name + "]")
		scope := wr.NewObjectScope(config.GenDataXcode.ProjectConfigUuid.String())
		{
			wr.member("isa", "XCBuildConfiguration")
			{
				scope := wr.NewObjectScope("buildSettings")
				//cc_flags
				for key, i := range config.XcodeSettings.Entries {
					wr.member(key, g.QuoteString(config.XcodeSettings.Values[i]))
				}
				wr.member("CLANG_CXX_LANGUAGE_STANDARD", g.QuoteString(config.CppStd))
				scope.Close()
			}
			wr.member("name", config.Name)
		}
		scope.Close()
	}

	for _, config := range proj.Configs {
		wr.newline(0)
		wr.commentBlock("Target Conifg [" + config.Name + "]")
		scope := wr.NewObjectScope(config.GenDataXcode.TargetConfigUuid.String())
		{
			wr.member("isa", "XCBuildConfiguration")
			{
				scope := wr.NewObjectScope("buildSettings")
				//link_flags
				outputTarget := config.OutputTarget.Path
				targetDir := filepath.Dir(outputTarget)
				targetBasename := PathBasename(outputTarget, false)
				targetExt := filepath.Ext(outputTarget)

				wr.member("PRODUCT_NAME", g.QuoteString(targetBasename))
				wr.member("EXECUTABLE_PREFIX", g.QuoteString(""))
				wr.member("EXECUTABLE_EXTENSION", g.QuoteString(targetExt))
				wr.member("CONFIGURATION_BUILD_DIR", g.QuoteString(targetDir))
				wr.member("CONFIGURATION_TEMP_DIR", g.QuoteString(config.BuildTmpDir.Path))

				{
					scope := wr.NewArrayScope("GCC_PREPROCESSOR_DEFINITIONS")
					for _, q := range config.CppDefines.FinalDict.Values {
						wr.newline(0)
						wr.write(g.QuoteString(q))
					}
					scope.Close()
				}
				{
					scope := wr.NewArrayScope("HEADER_SEARCH_PATHS")
					for _, q := range config.IncludeDirs.FinalDict.Values {
						wr.newline(0)
						if filepath.IsAbs(q) {
							wr.write(g.QuoteString2(q))
						} else {
							wr.write(g.QuoteString2(filepath.Join("$(PROJECT_DIR)/", q)))
						}
					}
					scope.Close()
				}
				{
					scope := wr.NewArrayScope("OTHER_CFLAGS")
					for _, q := range config.CppFlags.FinalDict.Values {
						wr.newline(0)
						wr.write(g.QuoteString(q))
					}
					scope.Close()
				}
				{
					scope := wr.NewArrayScope("OTHER_CPLUSPLUSFLAGS")
					for _, q := range config.CppFlags.FinalDict.Values {
						wr.newline(0)
						wr.write(g.QuoteString(q))
					}
					scope.Close()
				}

				{
					scope := wr.NewArrayScope("OTHER_LDFLAGS")
					for _, q := range config.LinkFlags.FinalDict.Values {
						wr.newline(0)
						wr.write(g.QuoteString(q))
					}
					for _, q := range config.LinkDirs.FinalDict.Values {
						wr.newline(0)
						wr.write(g.QuoteString(filepath.Join("-L", q)))
					}
					for _, q := range config.LinkLibs.FinalDict.Values {
						wr.newline(0)
						wr.write(g.QuoteString(filepath.Join("-l", q)))
					}
					for _, q := range config.LinkFiles.FinalDict.Values {
						wr.newline(0)
						wr.write(g.QuoteString2(q))
					}
					scope.Close()
				}

				//-----------
				if proj.TypeIsExeOrDll() {
					wr.member("PRODUCT_BUNDLE_IDENTIFIER", g.QuoteString(proj.Settings.Xcode.BundleIdentifier))
					wr.member("INFOPLIST_FILE", g.QuoteString(proj.GenDataXcode.InfoPlistFile))
				}

				if proj.PchHeader != nil {
					wr.member("GCC_PREFIX_HEADER", g.QuoteString(proj.PchHeader.Path))
					wr.member("GCC_PRECOMPILE_PREFIX_HEADER", "YES")
				}

				for key, value := range config.XcodeSettings.Entries {
					wr.member(key, g.QuoteString(config.XcodeSettings.Values[value]))
				}

				scope.Close()
			}
			wr.member("name", config.Name)
			scope.Close()
		}
	}
}

func (g *XcodeGenerator) genProjectXCConfigurationList(wr *XcodeWriter, proj *Project) {
	wr.newline(0)
	wr.commentBlock("----- XCConfigurationList -----------------")
	wr.newline(0)

	{
		wr.newline(0)
		wr.commentBlock("Build configuration list for PBXProject")
		scope := wr.NewObjectScope(proj.GenDataXcode.ConfigListUuid.String())
		{
			wr.member("isa", "XCConfigurationList")
			{
				scope := wr.NewArrayScope("buildConfigurations")
				for _, config := range proj.Configs {
					wr.newline(0)
					wr.commentBlock(config.Name)
					wr.newline(0)
					wr.write(config.GenDataXcode.ProjectConfigUuid.String())
				}
				scope.Close()
			}
			wr.member("defaultConfigurationIsVisible", "0")
			wr.member("defaultConfigurationName", g.Workspace.DefaultConfigName())
		}
		scope.Close()
	}

	{
		wr.newline(0)
		wr.commentBlock("Build configuration list for PBXNativeTarget")
		scope := wr.NewObjectScope(proj.GenDataXcode.TargetConfigListUuid.String())
		{
			wr.member("isa", "XCConfigurationList")
			{
				scope := wr.NewArrayScope("buildConfigurations")
				for _, config := range proj.Configs {
					wr.newline(0)
					wr.commentBlock(config.Name)
					wr.newline(0)
					wr.write(config.GenDataXcode.TargetConfigUuid.String())
				}
				scope.Close()
			}
			wr.member("defaultConfigurationIsVisible", "0")
			wr.member("defaultConfigurationName", g.Workspace.DefaultConfigName())
		}
		scope.Close()
	}
}

func (g *XcodeGenerator) genInfoPlistMacOSX(proj *Project) {
	gd := proj.GenDataXcode
	gd.InfoPlistFile = proj.Name + "_info.plist"

	wr := NewXmlWriter()
	wr.WriteHeader()
	wr.WriteDocType("plist", "-//Apple//DTD PLIST 1.0//EN", "http://www.apple.com/DTDs/PropertyList-1.0.dtd")

	{
		tag := wr.TagScope("plist")
		wr.Attr("version", "1.0")
		{
			tag := wr.TagScope("dict")
			wr.TagWithBody("key", "CFBundleDevelopmentRegion")
			wr.TagWithBody("string", "en")

			wr.TagWithBody("key", "CFBundleExecutable")
			wr.TagWithBody("string", "$(EXECUTABLE_NAME)")

			wr.TagWithBody("key", "CFBundleIconFile")
			wr.TagWithBody("string", "")

			wr.TagWithBody("key", "CFBundleIdentifier")
			wr.TagWithBody("string", proj.Settings.Xcode.BundleIdentifier)

			wr.TagWithBody("key", "CFBundleInfoDictionaryVersion")
			wr.TagWithBody("string", "6.0")

			wr.TagWithBody("key", "CFBundleName")
			wr.TagWithBody("string", "$(PRODUCT_NAME)")

			wr.TagWithBody("key", "CFBundlePackageType")
			wr.TagWithBody("string", "APPL")

			wr.TagWithBody("key", "CFBundleShortVersionString")
			wr.TagWithBody("string", "1.0")

			wr.TagWithBody("key", "CFBundleVersion")
			wr.TagWithBody("string", "1")

			wr.TagWithBody("key", "LSMinimumSystemVersion")
			wr.TagWithBody("string", "$(MACOSX_DEPLOYMENT_TARGET)")

			wr.TagWithBody("key", "NSHumanReadableCopyright")
			wr.TagWithBody("string", "=== Copyright ===")

			wr.TagWithBody("key", "NSMainNibFile")
			wr.TagWithBody("string", "MainMenu")

			wr.TagWithBody("key", "NSPrincipalClass")
			wr.TagWithBody("string", "NSApplication")

			tag.Close()
		}

		tag.Close()
	}

	filename := g.Workspace.GenerateAbsPath + gd.InfoPlistFile
	wr.WriteToFile(filename)
}
