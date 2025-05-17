package axe

import (
	"fmt"
	"path/filepath"
	"strings"

	ccode_utils "github.com/jurgen-kluft/ccode/utils"
)

type XcodeGenerator struct {
	Workspace *Workspace
}

func NewXcodeGenerator(ws *Workspace) *XcodeGenerator {
	return &XcodeGenerator{
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

	wr := ccode_utils.NewXmlWriter()
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

func (g *XcodeGenerator) genWorkspaceGroup(wr *ccode_utils.XmlWriter, group *ProjectGroup) {
	for _, c := range group.Children {
		tag := wr.TagScope("Group")
		{
			wr.Attr("location", "container:")
			wr.Attr("name", ccode_utils.PathFilename(c.Path, true))
			g.genWorkspaceGroup(wr, c)
		}
		tag.Close()
	}

	for _, proj := range group.Projects {
		tag := wr.TagScope("FileRef")
		{
			wr.Attr("location", "container:"+proj.Resolved.GenDataXcode.XcodeProj.Path)
		}
		tag.Close()
	}
}

func (g *XcodeGenerator) genProjectGenUuid(proj *Project) {
	proj.Resolved.GenDataXcode = NewXcodeProjectConfig()

	gd := proj.Resolved.GenDataXcode
	gd.XcodeProj.Init(filepath.Join(g.Workspace.GenerateAbsPath, proj.Name+".xcodeproj"), true)
	gd.PbxProj = filepath.Join(gd.XcodeProj.Path, "project.pbxproj")
	gd.Uuid = ccode_utils.GenerateUUID()
	gd.TargetUuid = ccode_utils.GenerateUUID()
	gd.TargetProductUuid = ccode_utils.GenerateUUID()
	gd.ConfigListUuid = ccode_utils.GenerateUUID()
	gd.TargetConfigListUuid = ccode_utils.GenerateUUID()
	gd.DependencyProxyUuid = ccode_utils.GenerateUUID()
	gd.DependencyTargetUuid = ccode_utils.GenerateUUID()
	gd.DependencyTargetProxyUuid = ccode_utils.GenerateUUID()

	for _, i := range proj.FileEntries.Dict {
		f := proj.FileEntries.Values[i]
		f.UUID = ccode_utils.GenerateUUID()
		f.BuildUUID = ccode_utils.GenerateUUID()
	}

	for _, i := range proj.ResourceEntries.Dict {
		f := proj.FileEntries.Values[i]
		f.UUID = ccode_utils.GenerateUUID()
		f.BuildUUID = ccode_utils.GenerateUUID()
	}

	for _, f := range proj.VirtualFolders.Folders {
		f.UUID = ccode_utils.GenerateUUID()
	}

	for _, config := range proj.Resolved.Configs.Values {
		config.GenDataXcode.ProjectConfigUuid = ccode_utils.GenerateUUID()
		config.GenDataXcode.TargetUuid = ccode_utils.GenerateUUID()
		config.GenDataXcode.TargetConfigUuid = ccode_utils.GenerateUUID()
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
		wr.member("rootObject", proj.Resolved.GenDataXcode.Uuid.ForXCode())
		scope.Close()
	}

	filename := proj.Resolved.GenDataXcode.PbxProj
	if err := wr.WriteToFile(filename); err != nil {
		return err
	}

	return nil
}

func (g *XcodeGenerator) genBuildFileReference(wr *XcodeWriter, f *FileEntry) {
	wr.newline(0)
	wr.commentBlock(f.Path)
	scope := wr.NewObjectScope(f.BuildUUID.ForXCode())
	{
		wr.member("isa", "PBXBuildFile")
		wr.member("fileRef", f.UUID.ForXCode())
	}
	scope.Close()
}

func (g *XcodeGenerator) genFileReference(wr *XcodeWriter, proj *Project, f *FileEntry) {
	wr.newline(0)
	wr.commentBlock(f.Path)

	scope := wr.NewObjectScope(f.UUID.ForXCode())
	basename := ccode_utils.PathFilename(f.Path, true)
	{
		wr.member("isa", "PBXFileReference")
		wr.member("name", g.quoteString(basename))
		wr.member("path", g.quoteString(basename))
		// if filepath.Ext(f.Path) == ".h" {
		// 	wr.member("path", g.quoteString(ccode_utils.PathGetRelativeTo(filepath.Join(proj.ProjectAbsPath, f.Path), proj.Resolved.GenDataXcode.PbxProj)))
		// } else {
		// 	wr.member("path", g.quoteString(ccode_utils.PathGetRelativeTo(filepath.Join(proj.ProjectAbsPath, f.Path), proj.Resolved.GenDataXcode.XcodeProj.Path)))
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
	for _, dp := range proj.Dependencies.Values {
		if !dp.Resolved.HasOutputTarget {
			continue
		}

		targetBasename := ccode_utils.PathFilename(dp.Resolved.GenDataXcode.XcodeProj.Path, true)

		wr.newline(0)
		wr.commentBlock(dp.Name)
		{
			wr.newline(0)
			wr.commentBlock("PBXContainerItemProxy for xcodeproject")
			scope := wr.NewObjectScope(dp.Resolved.GenDataXcode.DependencyProxyUuid.ForXCode())
			{
				wr.member("isa", "PBXContainerItemProxy")
				wr.member("containerPortal", dp.Resolved.GenDataXcode.Uuid.ForXCode())
				wr.member("proxyType", "2")
				wr.member("remoteInfo", g.quoteString(dp.Name))
			}
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("PBXContainerItemProxy for PBXTargetDependency")
			scope := wr.NewObjectScope(dp.Resolved.GenDataXcode.DependencyTargetProxyUuid.ForXCode())
			{
				wr.member("isa", "PBXContainerItemProxy")
				wr.member("containerPortal", dp.Resolved.GenDataXcode.Uuid.ForXCode())
				wr.member("proxyType", "1")
				wr.member("remoteInfo", g.quoteString(dp.Name))
			}
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("PBXTargetDependency")
			scope := wr.NewObjectScope(dp.Resolved.GenDataXcode.DependencyTargetUuid.ForXCode())
			{
				wr.member("isa", "PBXTargetDependency")
				wr.member("name", g.quoteString(dp.Name))
				wr.member("targetProxy", dp.Resolved.GenDataXcode.DependencyTargetProxyUuid.ForXCode())
			}
			scope.Close()
		}

		{
			wr.newline(0)
			wr.commentBlock("PBXFileReference")
			scope := wr.NewObjectScope(dp.Resolved.GenDataXcode.Uuid.ForXCode())
			{
				wr.member("isa", "PBXFileReference")
				wr.member("name", g.quoteString(targetBasename))
				wr.member("path", g.quoteString(dp.Resolved.GenDataXcode.XcodeProj.Path))
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
				for _, dp := range proj.Dependencies.Values {
					wr.write(dp.Resolved.GenDataXcode.Uuid.ForXCode())
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
				for _, i := range proj.ResourceEntries.Dict {
					f := proj.ResourceEntries.Values[i]
					wr.write(f.UUID.ForXCode())
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

	for _, i := range proj.ResourceEntries.Dict {
		f := proj.ResourceEntries.Values[i]
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

	for _, i := range proj.ResourceEntries.Dict {
		f := proj.ResourceEntries.Values[i]
		wr.newline(0)
		wr.commentBlock(f.Path)
		scope := wr.NewObjectScope(f.UUID.ForXCode())
		{
			basename := ccode_utils.PathFilename(f.Path, true)
			relPath := ccode_utils.PathGetRelativeTo(filepath.Join(proj.ProjectAbsPath, f.Path), proj.Resolved.GenDataXcode.XcodeProj.Path)
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
			if proj.Resolved.HasOutputTarget {
				wr.write(proj.Resolved.GenDataXcode.TargetProductUuid.ForXCode())
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
		scope := wr.NewObjectScope(v.UUID.ForXCode())
		{
			wr.member("isa", "PBXGroup")
			scope := wr.NewArrayScope("children")
			for _, c := range v.Children {
				wr.write(c.UUID.ForXCode())
			}
			for _, f := range v.Files {
				wr.write(f.UUID.ForXCode())
			}
			scope.Close()

			basename := ccode_utils.PathFilename(v.Path, true)
			wr.member("name", g.quoteString(basename))

			if v.Parent == root {
				wr.member("sourceTree", XcodeKSourceTreeProject)
				relPath := ccode_utils.PathGetRelativeTo(v.DiskPath, g.Workspace.GenerateAbsPath)
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
				wr.write(c.UUID.ForXCode())
			}
			for _, f := range root.Files {
				wr.write(f.UUID.ForXCode())
			}
			scope.Close()
		}
		wr.member("sourceTree", XcodeKSourceTreeProject)
		relPath := ccode_utils.PathGetRelativeTo(proj.GenerateAbsPath, g.Workspace.GenerateAbsPath)
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

	scope := wr.NewObjectScope(proj.Resolved.GenDataXcode.Uuid.ForXCode())
	{
		wr.member("isa", "PBXProject")
		{
			scope := wr.NewObjectScope("attributes")
			{
				scope := wr.NewObjectScope("TargetAttributes")
				{
					scope := wr.NewObjectScope(proj.Resolved.GenDataXcode.TargetUuid.ForXCode())
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
		wr.member("buildConfigurationList", proj.Resolved.GenDataXcode.ConfigListUuid.ForXCode())
		wr.member("mainGroup", XcodeMainGroupUUID)
		wr.member("productRefGroup", XcodeProductGroupUUID)
		wr.member("projectDirPath", g.quoteString(""))
		wr.member("projectRoot", g.quoteString(""))
		{
			scope := wr.NewArrayScope("targets")
			wr.write(proj.Resolved.GenDataXcode.TargetUuid.ForXCode())
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
				wr.write(f.BuildUUID.ForXCode())
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
			for _, i := range proj.ResourceEntries.Dict {
				f := proj.ResourceEntries.Values[i]
				wr.write(f.BuildUUID.ForXCode())
			}
			scope.Close()
		}
	}
	scope.Close()
}

func (g *XcodeGenerator) genProjectPBXNativeTarget(wr *XcodeWriter, proj *Project) {
	if !proj.Resolved.HasOutputTarget {
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
		scope := wr.NewObjectScope(proj.Resolved.GenDataXcode.TargetProductUuid.ForXCode())
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

			defaultConfig := proj.Resolved.Configs.First()
			targetBasename := ccode_utils.PathFilename(defaultConfig.Resolved.OutputTarget.Path, true)
			wr.member("path", g.quoteString(targetBasename))
			wr.member("sourceTree", "BUILT_PRODUCTS_DIR")
		}
		scope.Close()
	}

	{
		wr.newline(0)
		wr.commentBlock("--------- Target")
		scope := wr.NewObjectScope(proj.Resolved.GenDataXcode.TargetUuid.ForXCode())
		{
			wr.member("isa", "PBXNativeTarget")
			wr.member("name", g.quoteString(proj.Name))
			wr.member("productName", g.quoteString(proj.Name))
			wr.member("productReference", proj.Resolved.GenDataXcode.TargetProductUuid.ForXCode())
			wr.member("productType", productType)
			wr.member("buildConfigurationList", proj.Resolved.GenDataXcode.TargetConfigListUuid.ForXCode())
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
					for _, dp := range proj.Dependencies.Values {
						wr.write(dp.Resolved.GenDataXcode.DependencyTargetUuid.ForXCode())
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

	for _, config := range proj.Resolved.Configs.Values {
		wr.newline(0)
		wr.commentBlock("Project Config [" + config.String() + "]")
		scope := wr.NewObjectScope(config.GenDataXcode.ProjectConfigUuid.ForXCode())
		{
			wr.member("isa", "XCBuildConfiguration")
			{
				scope := wr.NewObjectScope("buildSettings")
				//cc_flags
				for i, value := range config.XcodeSettings.Values {
					wr.member(config.XcodeSettings.Keys[i], g.quoteString(value))
				}

				switch g.Workspace.Config.CppStd {
				case CppStd11:
					wr.member("CLANG_CXX_LANGUAGE_STANDARD", g.quoteString("c++11"))
				case CppStd14:
					wr.member("CLANG_CXX_LANGUAGE_STANDARD", g.quoteString("c++14"))
				case CppStd17:
					wr.member("CLANG_CXX_LANGUAGE_STANDARD", g.quoteString("c++17"))
				case CppStd20:
					wr.member("CLANG_CXX_LANGUAGE_STANDARD", g.quoteString("c++20"))
				case CppStdLatest:
					wr.member("CLANG_CXX_LANGUAGE_STANDARD", g.quoteString("c++latest"))
				}

				scope.Close()
			}
			wr.member("name", config.String())
		}
		scope.Close()
	}

	for _, config := range proj.Resolved.Configs.Values {
		wr.newline(0)
		wr.commentBlock("Target Config [" + config.String() + "]")
		scope := wr.NewObjectScope(config.GenDataXcode.TargetConfigUuid.ForXCode())
		{
			wr.member("isa", "XCBuildConfiguration")
			{
				scope := wr.NewObjectScope("buildSettings")

				//link_flags
				outputTarget := config.Resolved.OutputTarget.Path
				targetDir := filepath.Dir(outputTarget)
				targetBasename := ccode_utils.PathFilename(outputTarget, false)
				targetExt := strings.TrimLeft(filepath.Ext(outputTarget), ".")

				wr.member("PRODUCT_NAME", g.quoteString(targetBasename))
				wr.member("EXECUTABLE_PREFIX", g.quoteString(""))
				wr.member("EXECUTABLE_EXTENSION", g.quoteString(targetExt))
				wr.member("CONFIGURATION_BUILD_DIR", g.quoteString(targetDir))
				wr.member("CONFIGURATION_TEMP_DIR", g.quoteString(config.Resolved.BuildTmpDir.Path))

				{
					scope := wr.NewArrayScope("GCC_PREPROCESSOR_DEFINITIONS")
					for _, q := range config.CppDefines.Vars.Values {
						wr.newline(0)
						wr.write(g.quoteString(q))
					}
					scope.Close()
				}
				{
					scope := wr.NewArrayScope("HEADER_SEARCH_PATHS")
					for _, qk := range config.IncludeDirs.Values {
						wr.newline(0)
						headerSearchPath := qk.RelativeTo(g.Workspace.GenerateAbsPath)
						if filepath.IsAbs(headerSearchPath) {
							wr.write(g.quoteString2(headerSearchPath))
						} else {
							wr.write(g.quoteString2(headerSearchPath))
						}
					}
					scope.Close()
				}
				{
					scope := wr.NewArrayScope("OTHER_CFLAGS")
					for _, q := range config.CppFlags.Vars.Values {
						wr.newline(0)
						wr.write(g.quoteString(q))
					}
					scope.Close()
				}
				{
					scope := wr.NewArrayScope("OTHER_CPLUSPLUSFLAGS")
					for _, q := range config.CppFlags.Vars.Values {
						wr.newline(0)
						wr.write(g.quoteString(q))
					}
					scope.Close()
				}

				{
					scope := wr.NewArrayScope("OTHER_LDFLAGS")
					for _, q := range config.LinkFlags.Vars.Values {
						wr.newline(0)
						wr.write(g.quoteString(q))
					}
					// for _, q := range config.LinkDirs.FinalDict.Values {
					// 	wr.newline(0)
					// 	wr.write("-L" + g.quoteString(q))
					// }
					// for _, q := range config.LinkLibs.FinalDict.Values {
					// 	wr.newline(0)
					// 	wr.write("-l" + g.quoteString(q))
					// }
					// for _, q := range config.LinkFiles.FinalDict.Values {
					// 	wr.newline(0)
					// 	wr.write(g.quoteString2(q))
					// }
					scope.Close()
				}

				//-----------
				if proj.TypeIsExeOrDll() {
					wr.member("PRODUCT_BUNDLE_IDENTIFIER", g.quoteString(proj.Settings.Xcode.BundleIdentifier))
					wr.member("INFOPLIST_FILE", g.quoteString(proj.Resolved.GenDataXcode.InfoPlistFile))
				}

				if proj.Resolved.PchHeader != nil {
					wr.member("GCC_PREFIX_HEADER", g.quoteString(proj.Resolved.PchHeader.Path))
					wr.member("GCC_PRECOMPILE_PREFIX_HEADER", "YES")
				}

				for i, value := range config.XcodeSettings.Values {
					wr.member(config.XcodeSettings.Keys[i], g.quoteString(value))
				}

				scope.Close()
			}
			wr.member("name", config.String())
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
		scope := wr.NewObjectScope(proj.Resolved.GenDataXcode.ConfigListUuid.ForXCode())
		{
			wr.member("isa", "XCConfigurationList")
			{
				scope := wr.NewArrayScope("buildConfigurations")
				for _, config := range proj.Resolved.Configs.Values {
					wr.newline(0)
					wr.commentBlock(config.String())
					wr.newline(0)
					wr.write(config.GenDataXcode.ProjectConfigUuid.ForXCode())
				}
				scope.Close()
			}

			wr.member("defaultConfigurationIsVisible", "0")

			// TODO configurable ?
			defaultConfig := g.Workspace.ProjectList.Values[0].Resolved.Configs.First()
			wr.member("defaultConfigurationName", defaultConfig.String())
		}
		scope.Close()
	}

	{
		wr.newline(0)
		wr.commentBlock("Build configuration list for PBXNativeTarget")
		scope := wr.NewObjectScope(proj.Resolved.GenDataXcode.TargetConfigListUuid.ForXCode())
		{
			wr.member("isa", "XCConfigurationList")
			{
				scope := wr.NewArrayScope("buildConfigurations")
				for _, config := range proj.Resolved.Configs.Values {
					wr.newline(0)
					wr.commentBlock(config.String())
					wr.newline(0)
					wr.write(config.GenDataXcode.TargetConfigUuid.ForXCode())
				}
				scope.Close()
			}
			wr.member("defaultConfigurationIsVisible", "0")

			// TODO configurable ?
			defaultConfig := g.Workspace.ProjectList.Values[0].Resolved.Configs.First()
			wr.member("defaultConfigurationName", defaultConfig.String())
		}
		scope.Close()
	}
}

func (g *XcodeGenerator) genInfoPlistMacOSX(proj *Project) error {
	gd := proj.Resolved.GenDataXcode
	gd.InfoPlistFile = proj.Name + "_info.plist"

	wr := ccode_utils.NewXmlWriter()
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
