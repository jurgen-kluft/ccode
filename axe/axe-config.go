package axe

import (
	"log"
	"path/filepath"
	"strings"
)

// -----------------------------------------------------------------------------------------------------
// ConfigList
// -----------------------------------------------------------------------------------------------------

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type ConfigList struct {
	Dict   map[string]int
	Values []*Config
	Keys   []string
}

func NewConfigList() *ConfigList {
	return &ConfigList{
		Dict:   map[string]int{},
		Values: []*Config{},
		Keys:   []string{},
	}
}

func (p *ConfigList) Add(config *Config) {
	if _, ok := p.Dict[config.Type.String()]; !ok {
		p.Dict[config.Type.String()] = len(p.Values)
		p.Values = append(p.Values, config)
		p.Keys = append(p.Keys, config.Type.String())
	}
}

func (p *ConfigList) First() *Config {
	if len(p.Values) > 0 {
		return p.Values[0]
	}
	return nil
}

func (p *ConfigList) Get(t ConfigType) (*Config, bool) {
	if i, ok := p.Dict[t.String()]; ok {
		return p.Values[i], true
	}
	return nil, false
}

func (p *ConfigList) CollectByWildcard(name string, list *ConfigList) {
	for _, p := range p.Values {
		if PathMatchWildcard(p.Type.String(), name, true) {
			list.Add(p)
		}
	}
}

// -----------------------------------------------------------------------------------------------------
// ConfigType
// -----------------------------------------------------------------------------------------------------

type ConfigType int

const (
	ConfigTypeNone    ConfigType = 0
	ConfigTypeDebug   ConfigType = 1
	ConfigTypeRelease ConfigType = 2
	ConfigTypeFinal   ConfigType = 4
	ConfigTypeTest    ConfigType = 8
	ConfigTypeProfile ConfigType = 16
)

func (t ConfigType) IsDebug() bool {
	return t&ConfigTypeDebug != 0
}

func (t ConfigType) IsRelease() bool {
	return t&ConfigTypeRelease != 0
}

func (t ConfigType) IsProfile() bool {
	return t&ConfigTypeTest != 0
}

func (t ConfigType) IsTest() bool {
	return t&ConfigTypeTest != 0
}

func (t ConfigType) Tundra() string {
	switch t {
	case ConfigTypeDebug:
		return "*-*-debug-*"
	case ConfigTypeRelease:
		return "*-*-release-*"
	case ConfigTypeFinal:
		return "*-*-final-*"
	case ConfigTypeDebug | ConfigTypeTest:
		return "*-*-debug-test"
	case ConfigTypeRelease | ConfigTypeTest:
		return "*-*-release-test"
	case ConfigTypeFinal | ConfigTypeTest:
		return "*-*-final-test"
	case ConfigTypeDebug | ConfigTypeProfile:
		return "*-*-debug-profile"
	case ConfigTypeRelease | ConfigTypeProfile:
		return "*-*-release-profile"
	case ConfigTypeFinal | ConfigTypeProfile:
		return "*-*-final-profile"
	}
	return "*-*-debug-*"
}

func (t ConfigType) String() string {
	switch t {
	case ConfigTypeDebug:
		return "Debug"
	case ConfigTypeRelease:
		return "Release"
	case ConfigTypeFinal:
		return "Final"
	case ConfigTypeDebug | ConfigTypeTest:
		return "DebugTest"
	case ConfigTypeRelease | ConfigTypeTest:
		return "ReleaseTest"
	case ConfigTypeFinal | ConfigTypeTest:
		return "FinalTest"
	case ConfigTypeDebug | ConfigTypeProfile:
		return "DebugProfile"
	case ConfigTypeRelease | ConfigTypeProfile:
		return "ReleaseProfile"
	case ConfigTypeFinal | ConfigTypeProfile:
		return "FinalProfile"
	}
	return "Debug"
}

// -----------------------------------------------------------------------------------------------------
// Config
// -----------------------------------------------------------------------------------------------------

type Config struct {
	//Name         string
	Type         ConfigType
	Workspace    *Workspace
	CppStd       string
	Project      *Project
	OutputTarget *FileEntry
	OutputLib    *FileEntry
	BuildTmpDir  *FileEntry

	OutTargetDir string

	CppDefines     *VarSettings
	CppFlags       *VarSettings
	IncludeDirs    *PathSettings
	IncludeFiles   *PathSettings
	LinkDirs       *PathSettings
	LinkLibs       *VarSettings
	LinkFiles      *PathSettings
	LinkFlags      *VarSettings
	DisableWarning *VarSettings

	VarSettings  map[string]*VarSettings
	PathSettings map[string]*PathSettings

	XcodeSettings         *KeyValueDict
	VisualStudioClCompile *KeyValueDict
	VisualStudioLink      *KeyValueDict

	Resolved bool

	GenDataMakefile struct {
		CppObjDir string
	}

	GenDataXcode struct {
		ProjectConfigUuid UUID
		TargetUuid        UUID
		TargetConfigUuid  UUID
	}
}

func NewConfig(t ConfigType, ws *Workspace, p *Project) *Config {
	c := &Config{}
	//c.Name = name
	c.Type = t
	c.Workspace = ws
	c.Project = p
	c.CppStd = "c++14"

	proot := ""
	if p != nil {
		proot = p.ProjectAbsPath
	}

	c.CppDefines = NewVarDict("CppDefines")             // e.g. "DEBUG" "PROFILE"
	c.CppFlags = NewVarDict("CppFlags")                 // e.g. "-g"
	c.IncludeDirs = NewPathDict("IncludeDirs", proot)   // e.g. "source/main/include", "source/test/include"
	c.IncludeFiles = NewPathDict("IncludeFiles", proot) // e.g. "source/main/include/file.h", "source/test/include/file.h"
	c.LinkDirs = NewPathDict("LinkDirs", proot)         // e.g. "lib"
	c.LinkLibs = NewVarDict("LinkLibs")                 // These are just "name.lib" or "name.a" entries
	c.LinkFiles = NewPathDict("LinkFiles", proot)       // e.g. "link/name.o"
	c.LinkFlags = NewVarDict("LinkFlags")               // e.g. "-lstdc++"
	c.DisableWarning = NewVarDict("DisableWarning")     // e.g. "unused-variable"

	c.OutputTarget = NewFileEntry()
	c.OutputLib = NewFileEntry()
	c.BuildTmpDir = NewFileEntry()

	c.VarSettings = map[string]*VarSettings{
		c.CppDefines.Name:     c.CppDefines,
		c.CppFlags.Name:       c.CppFlags,
		c.LinkLibs.Name:       c.LinkLibs,
		c.LinkFlags.Name:      c.LinkFlags,
		c.DisableWarning.Name: c.DisableWarning,
	}

	c.PathSettings = map[string]*PathSettings{
		c.IncludeDirs.Name:  c.IncludeDirs,
		c.IncludeFiles.Name: c.IncludeFiles,
		c.LinkDirs.Name:     c.LinkDirs,
		c.LinkFiles.Name:    c.LinkFiles,
	}

	c.XcodeSettings = NewKeyValueDict()
	c.VisualStudioClCompile = NewKeyValueDict()
	c.VisualStudioLink = NewKeyValueDict()

	c.GenDataMakefile.CppObjDir = filepath.Join(c.Workspace.GenerateAbsPath, "build_tmp", c.Type.String())
	c.GenDataXcode.ProjectConfigUuid = GenerateUUID()
	c.GenDataXcode.TargetUuid = GenerateUUID()
	c.GenDataXcode.TargetConfigUuid = GenerateUUID()

	c.Resolved = false
	return c
}

func (c *Config) AddIncludeDir(includeDir string) {
	c.IncludeDirs.ValuesToAdd(includeDir)
}

func (c *Config) init(source *Config) {
	if c.Workspace == nil {
		log.Panic("Config hasn't been created with a valid Workspace")
	}

	if c.Project != nil {
		path := filepath.Join(c.Workspace.GenerateAbsPath, "build_tmp", c.Type.String(), c.Project.Name)
		c.BuildTmpDir = NewFileEntryInit(path, true)
	}

	if source != nil {
		c.CppStd = source.CppStd
		c.XcodeSettings.Merge(source.XcodeSettings)
		c.VisualStudioClCompile.Merge(source.VisualStudioClCompile)
		c.VisualStudioLink.Merge(source.VisualStudioLink)
	} else {
		c.InitXcodeSettings()
		c.InitVisualStudioSettings()
	}
}

func (c *Config) InitXcodeSettings() {

	settings := make(map[string]string)

	if c.Workspace.MakeTarget.OSIsIos() {
		settings["SDKROOT"] = "iphoneos"
		settings["SUPPORTED_PLATFORMS"] = "iphonesimulator iphoneos"
		settings["IPHONEOS_DEPLOYMENT_TARGET"] = "10.1"
	} else if c.Workspace.MakeTarget.OSIsMac() {
		settings["SDKROOT"] = "macosx"
		settings["SUPPORTED_PLATFORMS"] = "macosx"
		settings["MACOSX_DEPLOYMENT_TARGET"] = "10.14" // c++11 require 10.10+
	}

	if c.Type.IsDebug() {
		settings["DEBUG_INFORMATION_FORMAT"] = "dwarf"
		settings["GCC_GENERATE_DEBUGGING_SYMBOLS"] = "YES"

		// 0: None[-O0], 1: Fast[-O1],  2: Faster[-O2], 3: Fastest[-O3], s: Fastest, Smallest[-Os], Fastest, Aggressive Optimizations [-Ofast]
		settings["GCC_OPTIMIZATION_LEVEL"] = "0"
		settings["ONLY_ACTIVE_ARCH"] = "YES"
		settings["ENABLE_TESTABILITY"] = "YES"

	} else {
		settings["DEBUG_INFORMATION_FORMAT"] = "dwarf-with-dsym"
		settings["GCC_GENERATE_DEBUGGING_SYMBOLS"] = "NO"

		// 0: None[-O0], 1: Fast[-O1],  2: Faster[-O2], 3: Fastest[-O3], s: Fastest, Smallest[-Os], Fastest, Aggressive Optimizations [-Ofast]
		settings["GCC_OPTIMIZATION_LEVEL"] = "s"

		settings["ONLY_ACTIVE_ARCH"] = "NO"
		settings["ENABLE_TESTABILITY"] = "YES"
		settings["LLVM_LTO"] = "YES" //link time optimization
		settings["DEAD_CODE_STRIPPING"] = "YES"
		settings["STRIP_STYLE"] = "all"
	}

	settings["CODE_SIGN_IDENTITY"] = "-"
	settings["ALWAYS_SEARCH_USER_PATHS"] = "NO"
	settings["CLANG_ENABLE_OBJC_ARC"] = "YES"
	settings["GCC_SYMBOLS_PRIVATE_EXTERN"] = "YES"
	settings["ENABLE_STRICT_OBJC_MSGSEND"] = "YES"

	// clang warning flags
	settings["CLANG_ANALYZER_LOCALIZABILITY_NONLOCALIZED"] = "YES"
	settings["CLANG_WARN_BOOL_CONVERSION"] = "YES"
	settings["CLANG_WARN_CONSTANT_CONVERSION"] = "YES"
	settings["CLANG_WARN_EMPTY_BODY"] = "YES"
	settings["CLANG_WARN_ENUM_CONVERSION"] = "YES"
	settings["CLANG_WARN_INFINITE_RECURSION"] = "YES"
	settings["CLANG_WARN_INT_CONVERSION"] = "YES"
	settings["CLANG_WARN_SUSPICIOUS_MOVE"] = "YES"
	settings["CLANG_WARN_UNREACHABLE_CODE"] = "YES"
	settings["CLANG_WARN__DUPLICATE_METHOD_MATCH"] = "YES"
	settings["CLANG_WARN_IMPLICIT_SIGN_CONVERSION"] = "YES"
	settings["CLANG_WARN_ASSIGN_ENUM"] = "YES"
	settings["CLANG_WARN_SUSPICIOUS_IMPLICIT_CONVERSION"] = "YES"
	settings["CLANG_WARN_BLOCK_CAPTURE_AUTORELEASING"] = "YES"
	settings["CLANG_WARN_OBJC_IMPLICIT_RETAIN_SELF"] = "YES"
	settings["CLANG_WARN_DEPRECATED_OBJC_IMPLEMENTATIONS"] = "YES"
	settings["CLANG_WARN_RANGE_LOOP_ANALYSIS"] = "YES"
	settings["CLANG_WARN_STRICT_PROTOTYPES"] = "YES"
	settings["CLANG_WARN_COMMA"] = "YES"

	// gcc warning flags
	settings["GCC_WARN_FOUR_CHARACTER_CONSTANTS"] = "YES"
	settings["GCC_WARN_INITIALIZER_NOT_FULLY_BRACKETED"] = "YES"
	settings["GCC_WARN_ABOUT_MISSING_FIELD_INITIALIZERS"] = "YES"
	settings["GCC_WARN_SIGN_COMPARE"] = "YES"
	settings["GCC_TREAT_INCOMPATIBLE_POINTER_TYPE_WARNINGS_AS_ERRORS"] = "YES"
	settings["GCC_TREAT_IMPLICIT_FUNCTION_DECLARATIONS_AS_ERRORS"] = "YES"
	settings["GCC_WARN_UNUSED_LABEL"] = "YES"

	settings["GCC_WARN_64_TO_32_BIT_CONVERSION"] = "YES"
	settings["GCC_NO_COMMON_BLOCKS"] = "YES"
	settings["GCC_WARN_ABOUT_RETURN_TYPE"] = "YES"
	settings["GCC_WARN_UNDECLARED_SELECTOR"] = "YES"
	settings["GCC_WARN_UNINITIALIZED_AUTOS"] = "YES"
	settings["GCC_WARN_UNUSED_FUNCTION"] = "YES"
	settings["GCC_WARN_UNUSED_VARIABLE"] = "YES"

	c.XcodeSettings = NewKeyValueDict()
	for k, v := range settings {
		c.XcodeSettings.AddOrSet(k, v)
	}
}

func (c *Config) InitVisualStudioSettings() {
	c.VisualStudioClCompile = NewKeyValueDict()
	c.VisualStudioClCompile.AddOrSet("MinimalRebuild", "false")
	c.VisualStudioClCompile.AddOrSet("ExceptionHandling", "false")
	c.VisualStudioClCompile.AddOrSet("CompileAs", "CompileAsCpp")
	c.VisualStudioClCompile.AddOrSet("EnableModules", "false")
	c.VisualStudioClCompile.AddOrSet("TreatWarningAsError", "true")
	c.VisualStudioClCompile.AddOrSet("WarningLevel", "Level3") // Level0, Level1, Level2, Level3, Level4

	if c.Workspace.MakeTarget.CompilerIsClang() {
		c.VisualStudioClCompile.AddOrSet("DebugInformationFormat", "None")
	} else {
		c.VisualStudioClCompile.AddOrSet("DebugInformationFormat", "ProgramDatabase")
	}

	if c.Type.IsDebug() {
		c.VisualStudioClCompile.AddOrSet("Optimization", "Disabled")
		c.VisualStudioClCompile.AddOrSet("DebugInformationFormat", "ProgramDatabase")
		c.VisualStudioClCompile.AddOrSet("OmitFramePointers", "false")
	} else {
		c.VisualStudioClCompile.AddOrSet("Optimization", "Full") // MinSpace, MaxSpeed
		c.VisualStudioClCompile.AddOrSet("DebugInformationFormat", "None")
		c.VisualStudioClCompile.AddOrSet("OmitFramePointers", "true")
		c.VisualStudioClCompile.AddOrSet("FunctionLevelLinking", "true")
		c.VisualStudioClCompile.AddOrSet("IntrinsicFunctions", "true")
		c.VisualStudioClCompile.AddOrSet("WholeProgramOptimization", "true")
		c.VisualStudioClCompile.AddOrSet("BasicRuntimeChecks", "Default")
	}

	c.VisualStudioClCompile.AddOrSet("RuntimeLibrary", c.Workspace.Config.MsDev.RuntimeLibrary.String(c.Type.IsDebug()))
}

func (c *Config) inherit(rhs *Config) {
	c.resolve()

	for key, ps := range c.VarSettings {
		ps.inherit(rhs.VarSettings[key])
	}
	for key, ps := range c.PathSettings {
		ps.inherit(rhs.PathSettings[key])
	}

	if t := rhs.OutputLib.Path; len(t) > 0 {
		c.LinkFiles.InheritDict.AddOrSet(t, t)
	}

	c.XcodeSettings.UniqueExtend(rhs.XcodeSettings)
	c.VisualStudioClCompile.UniqueExtend(rhs.VisualStudioClCompile)
	c.VisualStudioLink.UniqueExtend(rhs.VisualStudioLink)
}

func (c *Config) computeFinal() {
	c.resolve()

	for _, p := range c.VarSettings {
		p.computeFinal()
	}
	for _, p := range c.PathSettings {
		p.computeFinal()
	}
}

func (c *Config) resolve() {
	if c.Resolved {
		return
	}
	c.Resolved = true

	if c.Project != nil {
		c.CppDefines.FinalDict.AddOrSet("CCORE_GEN_CPU", "CCORE_GEN_CPU_"+strings.ToUpper(c.Workspace.MakeTarget.ArchAsString()))
		c.CppDefines.FinalDict.AddOrSet("CCORE_GEN_OS", "CCORE_GEN_OS_"+strings.ToUpper(c.Workspace.MakeTarget.OSAsString()))
		c.CppDefines.FinalDict.AddOrSet("CCORE_GEN_COMPILER", "CCORE_GEN_COMPILER_"+strings.ToUpper(c.Workspace.MakeTarget.CompilerAsString()))
		c.CppDefines.FinalDict.AddOrSet("CCORE_GEN_GENERATOR", "CCORE_GEN_GENERATOR_"+strings.ToUpper(c.Workspace.Generator.String()))
		c.CppDefines.FinalDict.AddOrSet("CCORE_GEN_CONFIG", "CCORE_GEN_CONFIG_"+strings.ToUpper(c.Type.String()))
		c.CppDefines.FinalDict.AddOrSet("CCORE_GEN_PLATFORM_NAME", "CCORE_GEN_PLATFORM_NAME=\""+strings.ToUpper(c.Workspace.MakeTarget.OSAsString()+"\""))
		c.CppDefines.FinalDict.AddOrSet("CCORE_GEN_PROJECT", "CCORE_GEN_PROJECT_"+strings.ToUpper(c.Project.Name))
		c.CppDefines.FinalDict.AddOrSet("CCORE_GEN_TYPE", "CCORE_GEN_TYPE_"+strings.ToUpper(c.Project.Settings.Type.String()))

		BINDIR := filepath.Join(c.Workspace.GenerateAbsPath, "bin", c.Project.Name, c.Type.String()+"_"+c.Workspace.MakeTarget.ArchAsString()+"_"+c.Workspace.Config.MsDev.PlatformToolset)
		LIBDIR := filepath.Join(c.Workspace.GenerateAbsPath, "lib", c.Project.Name, c.Type.String()+"_"+c.Workspace.MakeTarget.ArchAsString()+"_"+c.Workspace.Config.MsDev.PlatformToolset)

		outputTarget := ""
		if c.Project.TypeIsExe() {
			executableFilename := c.Workspace.Config.ExeTargetPrefix + c.Project.Name + c.Workspace.Config.ExeTargetSuffix
			outputTarget = filepath.Join(BINDIR, executableFilename)
		} else if c.Project.TypeIsDll() {
			dllFilename := c.Workspace.Config.DllTargetPrefix + c.Project.Name + c.Workspace.Config.DllTargetSuffix
			outputTarget = filepath.Join(BINDIR, dllFilename)
			if c.Workspace.MakeTarget.OSIsWindows() {
				libFilename := c.Workspace.Config.LibTargetPrefix + c.Project.Name + c.Workspace.Config.LibTargetSuffix
				c.OutputLib = NewFileEntryInit(filepath.Join(LIBDIR, libFilename), false)
			} else {
				c.OutputLib = NewFileEntryInit(filepath.Join(BINDIR, dllFilename), false)
			}
		} else if c.Project.TypeIsLib() {
			libFilename := c.Workspace.Config.LibTargetPrefix + c.Project.Name + c.Workspace.Config.LibTargetSuffix
			outputTarget = filepath.Join(LIBDIR, libFilename)
			c.OutputLib = NewFileEntryInit(outputTarget, false)
		}

		if len(outputTarget) > 0 {
			c.OutputTarget = NewFileEntryInit(outputTarget, false)
		}
	}
}
