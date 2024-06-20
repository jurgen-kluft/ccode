package axe

import (
	"fmt"
	"path/filepath"
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

func (g *XcodeGenerator) Generate() {
	g.genWorkSpace()
}

func (g *XcodeGenerator) genWorkSpace() {
	xcodeWorkspace := filepath.Join(g.Workspace.GenerateAbsPath, g.Workspace.WorkspaceName+".xcworkspace")

	for _, proj := range g.Workspace.ProjectList.Values {
		fmt.Println("Project UUID generation: ", proj.Name)
		g.genProjectGenUuid(proj)
	}

	for _, proj := range g.Workspace.ProjectList.Values {
		fmt.Println("Project generation: ", proj.Name, " - ", proj.ProjectAbsPath)
		if err := g.genProject(proj); err != nil {
			fmt.Println("        Error : Project generation failed for ", proj.Name)
		}
	}

	wr := NewXmlWriter()
	{
		tag := wr.TagScope("Workspace")
		wr.Attr("version", "1.0")
		{
			g.genWorkspaceGroup(wr, g.Workspace.ProjectGroups.Root)
		}
		tag.Close()
	}

	wr.WriteToFile(filepath.Join(xcodeWorkspace, "contents.xcworkspacedata"))
}

func (g *XcodeGenerator) genWorkspaceGroup(wr *XmlWriter, group *ProjectGroup) {
	for _, c := range group.Children {
		tag := wr.TagScope("Group")
		{
			wr.Attr("location", "container:")
			wr.Attr("name", PathFilename(c.Path, true))
			g.genWorkspaceGroup(wr, c)
		}
		tag.Close()
	}

	for _, proj := range group.Projects {
		tag := wr.TagScope("FileRef")
		{
			wr.Attr("location", "container:"+proj.GenDataXcode.XcodeProj.Path)
		}
		tag.Close()
	}
}

func (g *XcodeGenerator) genProjectGenUuid(proj *Project) {
	proj.GenDataXcode = NewXcodeProjectConfig()

	gd := proj.GenDataXcode
	gd.XcodeProj.Init(filepath.Join(g.Workspace.GenerateAbsPath, proj.Name+".xcodeproj"), true)
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

	for _, config := range proj.Configs.Values {
		config.GenDataXcode.ProjectConfigUuid = GenerateUUID()
		config.GenDataXcode.TargetUuid = GenerateUUID()
		config.GenDataXcode.TargetConfigUuid = GenerateUUID()
	}
}

func (g *XcodeGenerator) genProject(proj *Project) error {
	if proj.TypeIsExeOrDll() {
		if g.Workspace.MakeTarget.OSIsIos() {
			//g.GenInfoPlistIOS(proj)
		} else {
			if err := g.genInfoPlistMacOSX(proj); err != nil {
				return err
			}
		}
	}

	wr := NewXcodeWriter()
	wr.write("// !$*UTF8*$!")
	wr.newline(0)
	{
		scope := wr.NewObjectScope("")
		wr.member("archiveVersion", "1")
		{
			scope := wr.NewObjectScope("classes")
			scope.Close()
		}
		wr.member("objectVersion", "46")
		{
			scope := wr.NewObjectScope("objects")
			g.genProjectPBXBuildFile(wr, proj)
			g.genProjectDependencies(wr, proj)
			g.genProjectPBXGroup(wr, proj)
			g.genProjectPBXProject(wr, proj)
			g.genProjectPBXSourcesBuildPhase(wr, proj)
			g.genProjectPBXResourcesBuildPhase(wr, proj)
			g.genProjectPBXNativeTarget(wr, proj)
			g.genProjectXCBuildConfiguration(wr, proj)
			g.genProjectXCConfigurationList(wr, proj)
			scope.Close()
		}
		wr.member("rootObject", proj.GenDataXcode.Uuid.String(g.Workspace.Generator))
		scope.Close()
	}

	filename := proj.GenDataXcode.PbxProj
	if err := wr.WriteToFile(filename); err != nil {
		return err
	}

	return nil
}

func (g *XcodeGenerator) genBuildFileReference(wr *XcodeWriter, f *FileEntry) {
	wr.newline(0)
	wr.commentBlock(f.Path)
	scope := wr.NewObjectScope(f.GenDataXcode.BuildUUID.String(g.Workspace.Generator))
	{
		wr.member("isa", "PBXBuildFile")
		wr.member("fileRef", f.GenDataXcode.UUID.String(g.Workspace.Generator))
	}
	scope.Close()
}

func (g *XcodeGenerator) genFileReference(wr *XcodeWriter, proj *Project, f *FileEntry) {
	wr.newline(0)
	wr.commentBlock(f.Path)

	scope := wr.NewObjectScope(f.GenDataXcode.UUID.String(g.Workspace.Generator))
	basename := PathFilename(f.Path, true)
	{
		wr.member("isa", "PBXFileReference")
		wr.member("name", g.quoteString(basename))
		wr.member("path", g.quoteString(basename))
		// if filepath.Ext(f.Path) == ".h" {
		// 	wr.member("path", g.quoteString(PathGetRel(filepath.Join(proj.ProjectAbsPath, f.Path), proj.GenDataXcode.PbxProj)))
		// } else {
		// 	wr.member("path", g.quoteString(PathGetRel(filepath.Join(proj.ProjectAbsPath, f.Path), proj.GenDataXcode.XcodeProj.Path)))
		// }

		wr.member("sourceTree", XcodeKSourceTreeGroup)

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

		targetBasename := PathFilename(dp.GenDataXcode.XcodeProj.Path, true)

		wr.newline(0)
		wr.commentBlock(dp.Name)
		{
			wr.newline(0)
			wr.commentBlock("PBXContainerItemProxy for xcodeproject")
			scope := wr.NewObjectScope(dp.GenDataXcode.DependencyProxyUuid.String(g.Workspace.Generator))
			{
				wr.member("isa", "PBXContainerItemProxy")
				wr.member("containerPortal", dp.GenDataXcode.Uuid.String(g.Workspace.Generator))
				wr.member("proxyType", "2")
				wr.member("remoteInfo", g.quoteString(dp.Name))
			}
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("PBXContainerItemProxy for PBXTargetDependency")
			scope := wr.NewObjectScope(dp.GenDataXcode.DependencyTargetProxyUuid.String(g.Workspace.Generator))
			{
				wr.member("isa", "PBXContainerItemProxy")
				wr.member("containerPortal", dp.GenDataXcode.Uuid.String(g.Workspace.Generator))
				wr.member("proxyType", "1")
				wr.member("remoteInfo", g.quoteString(dp.Name))
			}
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("PBXTargetDependency")
			scope := wr.NewObjectScope(dp.GenDataXcode.DependencyTargetUuid.String(g.Workspace.Generator))
			{
				wr.member("isa", "PBXTargetDependency")
				wr.member("name", g.quoteString(dp.Name))
				wr.member("targetProxy", dp.GenDataXcode.DependencyTargetProxyUuid.String(g.Workspace.Generator))
			}
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("PBXFileReference")
			scope := wr.NewObjectScope(dp.GenDataXcode.Uuid.String(g.Workspace.Generator))
			{
				wr.member("isa", "PBXFileReference")
				wr.member("name", g.quoteString(targetBasename))
				wr.member("path", g.quoteString(dp.GenDataXcode.XcodeProj.Path))
				wr.member("sourceTree", XcodeKSourceTreeAbsolute)
			}
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("------ Folder dependencies")
			scope := wr.NewObjectScope(XcodeDependenciesGroupUUID)
			wr.member("isa", "PBXGroup")
			{
				scope := wr.NewArrayScope("children")
				for _, dp := range proj.DependenciesInherit.Values {
					wr.write(dp.GenDataXcode.Uuid.String(g.Workspace.Generator))
				}
				scope.Close()
			}
			wr.member("sourceTree", XcodeKSourceTreeGroup)
			wr.member("name", "_dependencies_")
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("------ Folder resources")
			scope := wr.NewObjectScope(XcodeResourcesGroupUUID)
			wr.member("isa", "PBXGroup")
			{
				scope := wr.NewArrayScope("children")
				for _, i := range proj.ResourceDirs.Dict {
					f := proj.ResourceDirs.Values[i]
					wr.write(f.GenDataXcode.UUID.String(g.Workspace.Generator))
				}
				scope.Close()
			}
			wr.member("sourceTree", XcodeKSourceTreeGroup)
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
		scope := wr.NewObjectScope(f.GenDataXcode.UUID.String(g.Workspace.Generator))
		{
			basename := PathFilename(f.Path, true)
			relPath := PathGetRel(filepath.Join(proj.ProjectAbsPath, f.Path), proj.GenDataXcode.XcodeProj.Path)
			if filepath.Ext(f.Path) == ".h" {
				relPath = "../" + relPath
			}
			wr.member("isa", "PBXFileReference")
			wr.member("name", g.quoteString(basename))
			wr.member("path", g.quoteString(basename))
			wr.member("sourceTree", XcodeKSourceTreeProject)
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
		scope := wr.NewObjectScope(XcodeBuildPhaseFrameworksUUID)
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
		scope := wr.NewObjectScope(XcodeProductGroupUUID)
		{
			wr.member("isa", "PBXGroup")
			scope := wr.NewArrayScope("children")
			if proj.HasOutputTarget {
				wr.write(proj.GenDataXcode.TargetProductUuid.String(g.Workspace.Generator))
			}
			scope.Close()
			wr.member("sourceTree", XcodeKSourceTreeGroup)
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
		scope := wr.NewObjectScope(v.GenData_xcode.UUID.String(g.Workspace.Generator))
		{
			wr.member("isa", "PBXGroup")
			scope := wr.NewArrayScope("children")
			for _, c := range v.Children {
				wr.write(c.GenData_xcode.UUID.String(g.Workspace.Generator))
			}
			for _, f := range v.Files {
				wr.write(f.GenDataXcode.UUID.String(g.Workspace.Generator))
			}
			scope.Close()

			basename := PathFilename(v.Path, true)
			wr.member("name", g.quoteString(basename))

			if v.Parent == root {
				wr.member("sourceTree", XcodeKSourceTreeProject)
				relPath := PathGetRel(v.DiskPath, g.Workspace.GenerateAbsPath)
				wr.member("path", g.quoteString(relPath))
			} else {
				wr.member("sourceTree", XcodeKSourceTreeGroup)
				wr.member("path", g.quoteString(basename))
			}
		}
		scope.Close()
	}

	{
		scope := wr.NewObjectScope(XcodeMainGroupUUID)
		wr.member("isa", "PBXGroup")
		{
			scope := wr.NewArrayScope("children")
			wr.write(XcodeProductGroupUUID)
			wr.write(XcodeDependenciesGroupUUID)
			wr.write(XcodeResourcesGroupUUID)
			for _, c := range root.Children {
				wr.write(c.GenData_xcode.UUID.String(g.Workspace.Generator))
			}
			for _, f := range root.Files {
				wr.write(f.GenDataXcode.UUID.String(g.Workspace.Generator))
			}
			scope.Close()
		}
		wr.member("sourceTree", XcodeKSourceTreeProject)
		relPath := PathGetRel(proj.GenerateAbsPath, g.Workspace.GenerateAbsPath)
		wr.member("path", g.quoteString(relPath))
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

	scope := wr.NewObjectScope(proj.GenDataXcode.Uuid.String(g.Workspace.Generator))
	{
		wr.member("isa", "PBXProject")
		{
			scope := wr.NewObjectScope("attributes")
			{
				scope := wr.NewObjectScope("TargetAttributes")
				{
					scope := wr.NewObjectScope(proj.GenDataXcode.TargetUuid.String(g.Workspace.Generator))
					wr.member("CreatedOnToolsVersion", "7.3.1")
					wr.member("ProvisioningStyle", "Automatic")
					scope.Close()
				}
				scope.Close()
			}
			scope.Close()
		}
		wr.member("compatibilityVersion", g.quoteString("Xcode 3.2"))
		wr.member("developmentRegion", "en")
		wr.member("hasScannedForEncodings", "0")
		wr.member("knownRegions", "(Base, en,)")
		wr.member("buildConfigurationList", proj.GenDataXcode.ConfigListUuid.String(g.Workspace.Generator))
		wr.member("mainGroup", XcodeMainGroupUUID)
		wr.member("productRefGroup", XcodeProductGroupUUID)
		wr.member("projectDirPath", g.quoteString(""))
		wr.member("projectRoot", g.quoteString(""))
		{
			scope := wr.NewArrayScope("targets")
			wr.write(proj.GenDataXcode.TargetUuid.String(g.Workspace.Generator))
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

	scope := wr.NewObjectScope(XcodeBuildPhaseSourcesUUID)
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
				wr.write(f.GenDataXcode.BuildUUID.String(g.Workspace.Generator))
			}
			scope.Close()
		}
	}
	scope.Close()
}

func (g *XcodeGenerator) genProjectPBXResourcesBuildPhase(wr *XcodeWriter, proj *Project) {
	wr.newline(0)
	wr.commentBlock("------ PBXResourcesBuildPhase section")

	scope := wr.NewObjectScope(XcodeBuildPhaseResourcesUUID)
	{
		wr.member("isa", "PBXResourcesBuildPhase")
		wr.member("buildActionMask", "2147483647")
		wr.member("runOnlyForDeploymentPostprocessing", "0")
		{
			scope := wr.NewArrayScope("files")
			for _, i := range proj.ResourceDirs.Dict {
				f := proj.ResourceDirs.Values[i]
				wr.write(f.GenDataXcode.BuildUUID.String(g.Workspace.Generator))
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
			productType = XcodeProductTypeApplication.String()
		} else {
			productType = XcodeProductTypeTool.String()
		}
	} else if proj.TypeIsDll() {
		productType = string(XcodeProductTypeDynamicLib)
	} else if proj.TypeIsLib() {
		productType = string(XcodeProductTypeStaticLib)
	} else {
		panic("Unsupported project type")
	}

	wr.newline(0)
	wr.commentBlock("--------- PBXNativeTarget section")

	{
		wr.newline(0)
		wr.commentBlock("--------- File Reference")
		scope := wr.NewObjectScope(proj.GenDataXcode.TargetProductUuid.String(g.Workspace.Generator))
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

			defaultConfig := proj.Configs.First()
			targetBasename := PathFilename(defaultConfig.OutputTarget.Path, true)
			wr.member("path", g.quoteString(targetBasename))
			wr.member("sourceTree", "BUILT_PRODUCTS_DIR")
		}
		scope.Close()
	}

	{
		wr.newline(0)
		wr.commentBlock("--------- Target")
		scope := wr.NewObjectScope(proj.GenDataXcode.TargetUuid.String(g.Workspace.Generator))
		{
			wr.member("isa", "PBXNativeTarget")
			wr.member("name", g.quoteString(proj.Name))
			wr.member("productName", g.quoteString(proj.Name))
			wr.member("productReference", proj.GenDataXcode.TargetProductUuid.String(g.Workspace.Generator))
			wr.member("productType", productType)
			wr.member("buildConfigurationList", proj.GenDataXcode.TargetConfigListUuid.String(g.Workspace.Generator))
			{
				scope := wr.NewArrayScope("buildPhases")
				wr.write(XcodeBuildPhaseSourcesUUID)
				wr.write(XcodeBuildPhaseFrameworksUUID)
				wr.write(XcodeBuildPhaseHeadersUUID)
				wr.write(XcodeBuildPhaseResourcesUUID)
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
						wr.write(dp.GenDataXcode.DependencyTargetUuid.String(g.Workspace.Generator))
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

	for _, config := range proj.Configs.Values {
		wr.newline(0)
		wr.commentBlock("Project Config [" + config.Type.String() + "]")
		scope := wr.NewObjectScope(config.GenDataXcode.ProjectConfigUuid.String(g.Workspace.Generator))
		{
			wr.member("isa", "XCBuildConfiguration")
			{
				scope := wr.NewObjectScope("buildSettings")
				//cc_flags
				for key, i := range config.XcodeSettings.Entries {
					wr.member(key, g.quoteString(config.XcodeSettings.Values[i]))
				}
				wr.member("CLANG_CXX_LANGUAGE_STANDARD", g.quoteString(g.Workspace.Config.CppStd))
				scope.Close()
			}
			wr.member("name", config.Type.String())
		}
		scope.Close()
	}

	for _, config := range proj.Configs.Values {
		wr.newline(0)
		wr.commentBlock("Target Config [" + config.Type.String() + "]")
		scope := wr.NewObjectScope(config.GenDataXcode.TargetConfigUuid.String(g.Workspace.Generator))
		{
			wr.member("isa", "XCBuildConfiguration")
			{
				scope := wr.NewObjectScope("buildSettings")

				//link_flags
				outputTarget := config.OutputTarget.Path
				targetDir := filepath.Dir(outputTarget)
				targetBasename := PathFilename(outputTarget, false)
				targetExt := filepath.Ext(outputTarget)

				wr.member("PRODUCT_NAME", g.quoteString(targetBasename))
				wr.member("EXECUTABLE_PREFIX", g.quoteString(""))
				wr.member("EXECUTABLE_EXTENSION", g.quoteString(targetExt))
				wr.member("CONFIGURATION_BUILD_DIR", g.quoteString(targetDir))
				wr.member("CONFIGURATION_TEMP_DIR", g.quoteString(config.BuildTmpDir.Path))

				{
					scope := wr.NewArrayScope("GCC_PREPROCESSOR_DEFINITIONS")
					for _, q := range config.CppDefines.FinalDict.Values {
						wr.newline(0)
						wr.write(g.quoteString(q))
					}
					scope.Close()
				}
				{
					scope := wr.NewArrayScope("HEADER_SEARCH_PATHS")
					for qk, _ := range config.IncludeDirs.FinalDict.Entries {
						wr.newline(0)
						if filepath.IsAbs(qk) {
							wr.write(g.quoteString2(qk))
						} else {
							//relq := PathGetRel(filepath.Join(proj.ProjectAbsPath, q), g.Workspace.GenerateAbsPath)
							//wr.write(g.quoteString2(filepath.Join("$(PROJECT_DIR)/", q)))
							wr.write(g.quoteString2(qk))
						}
					}
					scope.Close()
				}
				{
					scope := wr.NewArrayScope("OTHER_CFLAGS")
					for _, q := range config.CppFlags.FinalDict.Values {
						wr.newline(0)
						wr.write(g.quoteString(q))
					}
					scope.Close()
				}
				{
					scope := wr.NewArrayScope("OTHER_CPLUSPLUSFLAGS")
					for _, q := range config.CppFlags.FinalDict.Values {
						wr.newline(0)
						wr.write(g.quoteString(q))
					}
					scope.Close()
				}

				{
					scope := wr.NewArrayScope("OTHER_LDFLAGS")
					for _, q := range config.LinkFlags.FinalDict.Values {
						wr.newline(0)
						wr.write(g.quoteString(q))
					}
					for _, q := range config.LinkDirs.FinalDict.Values {
						wr.newline(0)
						wr.write(g.quoteString(filepath.Join("-L", q)))
					}
					for _, q := range config.LinkLibs.FinalDict.Values {
						wr.newline(0)
						wr.write(g.quoteString(filepath.Join("-l", q)))
					}
					for _, q := range config.LinkFiles.FinalDict.Values {
						wr.newline(0)
						wr.write(g.quoteString2(q))
					}
					scope.Close()
				}

				//-----------
				if proj.TypeIsExeOrDll() {
					wr.member("PRODUCT_BUNDLE_IDENTIFIER", g.quoteString(proj.Settings.Xcode.BundleIdentifier))
					wr.member("INFOPLIST_FILE", g.quoteString(proj.GenDataXcode.InfoPlistFile))
				}

				if proj.PchHeader != nil {
					wr.member("GCC_PREFIX_HEADER", g.quoteString(proj.PchHeader.Path))
					wr.member("GCC_PRECOMPILE_PREFIX_HEADER", "YES")
				}

				for key, value := range config.XcodeSettings.Entries {
					wr.member(key, g.quoteString(config.XcodeSettings.Values[value]))
				}

				scope.Close()
			}
			wr.member("name", config.Type.String())
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
		scope := wr.NewObjectScope(proj.GenDataXcode.ConfigListUuid.String(g.Workspace.Generator))
		{
			wr.member("isa", "XCConfigurationList")
			{
				scope := wr.NewArrayScope("buildConfigurations")
				for _, config := range proj.Configs.Values {
					wr.newline(0)
					wr.commentBlock(config.Type.String())
					wr.newline(0)
					wr.write(config.GenDataXcode.ProjectConfigUuid.String(g.Workspace.Generator))
				}
				scope.Close()
			}

			wr.member("defaultConfigurationIsVisible", "0")

			defaultConfig := g.Workspace.Configs.First()
			wr.member("defaultConfigurationName", defaultConfig.Type.String())
		}
		scope.Close()
	}

	{
		wr.newline(0)
		wr.commentBlock("Build configuration list for PBXNativeTarget")
		scope := wr.NewObjectScope(proj.GenDataXcode.TargetConfigListUuid.String(g.Workspace.Generator))
		{
			wr.member("isa", "XCConfigurationList")
			{
				scope := wr.NewArrayScope("buildConfigurations")
				for _, config := range proj.Configs.Values {
					wr.newline(0)
					wr.commentBlock(config.Type.String())
					wr.newline(0)
					wr.write(config.GenDataXcode.TargetConfigUuid.String(g.Workspace.Generator))
				}
				scope.Close()
			}
			wr.member("defaultConfigurationIsVisible", "0")
			defaultConfig := g.Workspace.Configs.First()
			wr.member("defaultConfigurationName", defaultConfig.Type.String())
		}
		scope.Close()
	}
}

func (g *XcodeGenerator) genInfoPlistMacOSX(proj *Project) error {
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

	filename := filepath.Join(g.Workspace.GenerateAbsPath, gd.InfoPlistFile)
	if err := wr.WriteToFile(filename); err != nil {
		return err
	}
	return nil
}

func (g *XcodeGenerator) quoteString2(v string) string {
	return g.quoteString(g.quoteString(v))
}

func (g *XcodeGenerator) quoteString(v string) string {
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
