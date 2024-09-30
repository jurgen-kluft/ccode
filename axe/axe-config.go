package axe

import (
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/denv"
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
	if _, ok := p.Dict[config.String()]; !ok {
		p.Dict[config.String()] = len(p.Values)
		p.Values = append(p.Values, config)
		p.Keys = append(p.Keys, config.String())
	}
}

func (p *ConfigList) DefaultConfigName() string {
	if len(p.Values) > 0 {
		return p.Values[0].String()
	}
	return "Debug"
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

func (p *ConfigList) Has(t ConfigType) bool {
	_, ok := p.Dict[t.String()]
	return ok
}

func (p *ConfigList) CollectByWildcard(name string, list *ConfigList) {
	for _, p := range p.Values {
		if PathMatchWildcard(p.String(), name, true) {
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

func MakeFromDenvConfigType(cfgType denv.ConfigType) ConfigType {
	configType := ConfigTypeNone

	if cfgType.IsDebug() {
		configType = ConfigTypeDebug
	} else if cfgType.IsRelease() {
		configType = ConfigTypeRelease
	} else if cfgType.IsFinal() {
		configType = ConfigTypeFinal
	}

	if cfgType.IsProfile() {
		configType = ConfigTypeProfile
	}

	if cfgType.IsUnittest() {
		configType |= ConfigTypeTest
	}

	return configType
}

func (t ConfigType) IsDebug() bool {
	return t&ConfigTypeDebug != 0
}
func (t ConfigType) IsRelease() bool {
	return t&ConfigTypeRelease != 0
}
func (t ConfigType) IsFinal() bool {
	return t&ConfigTypeFinal != 0
}
func (t ConfigType) IsProfile() bool {
	return t&ConfigTypeProfile != 0
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
	Type      ConfigType
	Workspace *Workspace
	Project   *Project

	CppDefines     *VarSettings
	CppFlags       *VarSettings
	IncludeDirs    *PinnedPathSet
	IncludeFiles   *PinnedPathSet
	Library        *Library
	LinkFlags      *VarSettings
	DisableWarning *VarSettings

	XcodeSettings         *KeyValueDict
	VisualStudioClCompile *KeyValueDict
	VisualStudioLink      *KeyValueDict

	GenDataXcode struct {
		ProjectConfigUuid UUID
		TargetUuid        UUID
		TargetConfigUuid  UUID
	}

	Resolved *ConfigResolved
}

func NewConfig(t ConfigType, ws *Workspace, p *Project) *Config {
	c := &Config{}
	c.Type = t
	c.Workspace = ws
	c.Project = p

	c.CppDefines = NewVarDict("CppDefines")         // e.g. "DEBUG" "PROFILE"
	c.CppFlags = NewVarDict("CppFlags")             // e.g. "-g"
	c.IncludeDirs = NewPinnedPathSet()              // e.g. "source/main/include", "source/test/include"
	c.IncludeFiles = NewPinnedPathSet()             // e.g. "source/main/include/file.h", "source/test/include/file.h"
	c.Library = NewLibrary()                        // Holds the information of library directories and files
	c.LinkFlags = NewVarDict("LinkFlags")           // e.g. "-lstdc++"
	c.DisableWarning = NewVarDict("DisableWarning") // e.g. "unused-variable"

	c.XcodeSettings = NewKeyValueDict()
	c.VisualStudioClCompile = NewKeyValueDict()
	c.VisualStudioLink = NewKeyValueDict()

	c.GenDataXcode.ProjectConfigUuid = GenerateUUID()
	c.GenDataXcode.TargetUuid = GenerateUUID()
	c.GenDataXcode.TargetConfigUuid = GenerateUUID()

	c.InitXcodeSettings()
	c.InitVisualStudioSettings()

	return c
}

func (c *Config) String() string {
	return c.Type.String()
}

func (c *Config) AddIncludeDir(includeDir string) {
	c.IncludeDirs.AddOrSet(c.Project.ProjectAbsPath, includeDir)
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
		settings["MACOSX_DEPLOYMENT_TARGET"] = "10.15" // c++11 require 10.10+
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
	//	settings["CLANG_ENABLE_OBJC_ARC"] = "YES"
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

type ConfigResolved struct {
	OutputTarget *FileEntry
	OutputLib    *FileEntry
	BuildTmpDir  *FileEntry
}

func NewConfigResolved() *ConfigResolved {
	c := &ConfigResolved{}

	c.OutputTarget = NewFileEntry()
	c.OutputLib = NewFileEntry()
	c.BuildTmpDir = NewFileEntry()

	return c
}

func (c *Config) Copy() *Config {
	nc := NewConfig(c.Type, c.Workspace, c.Project)

	nc.CppDefines = c.CppDefines.Copy()
	nc.CppFlags = c.CppFlags.Copy()
	nc.IncludeDirs = c.IncludeDirs.Copy()
	nc.IncludeFiles = c.IncludeFiles.Copy()
	nc.Library = c.Library.Copy()
	nc.LinkFlags = c.LinkFlags.Copy()
	nc.DisableWarning = c.DisableWarning.Copy()

	nc.XcodeSettings = c.XcodeSettings.Copy()
	nc.VisualStudioClCompile = c.VisualStudioClCompile.Copy()
	nc.VisualStudioLink = c.VisualStudioLink.Copy()

	nc.GenDataXcode = c.GenDataXcode

	nc.Resolved = nil

	return nc
}

func (c *Config) BuildResolved(otherConfigs []*Config) *Config {
	configMerged := c.Copy()

	// Merge the settings from the other configs
	for _, otherConfig := range otherConfigs {
		configMerged.CppDefines.Merge(otherConfig.CppDefines)
		configMerged.CppFlags.Merge(otherConfig.CppFlags)
		configMerged.IncludeDirs.Merge(otherConfig.IncludeDirs)
		configMerged.IncludeFiles.Merge(otherConfig.IncludeFiles)
		configMerged.Library.Merge(otherConfig.Library)
		configMerged.LinkFlags.Merge(otherConfig.LinkFlags)

		configMerged.DisableWarning.Merge(otherConfig.DisableWarning)
		configMerged.XcodeSettings.Merge(otherConfig.XcodeSettings)
		configMerged.VisualStudioClCompile.Merge(otherConfig.VisualStudioClCompile)
		configMerged.VisualStudioLink.Merge(otherConfig.VisualStudioLink)

	}

	configResolved := NewConfigResolved()

	if configMerged.Project != nil {
		path := filepath.Join(c.Workspace.GenerateAbsPath, "build_tmp", c.String(), c.Project.Name)
		configResolved.BuildTmpDir = NewFileEntryInit(path, true)
	}

	// TODO Only do this if this project is ccore or has a dependency on ccore
	if c.Project != nil {
		configMerged.CppDefines.AddOrSet("CCORE_GEN_CPU", "CCORE_GEN_CPU_"+strings.ToUpper(c.Workspace.MakeTarget.ArchAsString()))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_OS", "CCORE_GEN_OS_"+strings.ToUpper(c.Workspace.MakeTarget.OSAsString()))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_COMPILER", "CCORE_GEN_COMPILER_"+strings.ToUpper(c.Workspace.MakeTarget.CompilerAsString()))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_GENERATOR", "CCORE_GEN_GENERATOR_"+strings.ToUpper(c.Workspace.Config.Dev.String()))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_CONFIG", "CCORE_GEN_CONFIG_"+strings.ToUpper(c.String()))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_PLATFORM_NAME", "CCORE_GEN_PLATFORM_NAME=\""+strings.ToUpper(c.Workspace.MakeTarget.OSAsString()+"\""))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_PROJECT", "CCORE_GEN_PROJECT_"+strings.ToUpper(c.Project.Name))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_TYPE", "CCORE_GEN_TYPE_"+strings.ToUpper(c.Project.Type.String()))
	}

	if c.Project != nil {
		BINDIR := filepath.Join(c.Workspace.GenerateAbsPath, "bin", c.Project.Name, c.String()+"_"+c.Workspace.MakeTarget.ArchAsString()+"_"+c.Workspace.Config.MsDev.PlatformToolset)
		LIBDIR := filepath.Join(c.Workspace.GenerateAbsPath, "lib", c.Project.Name, c.String()+"_"+c.Workspace.MakeTarget.ArchAsString()+"_"+c.Workspace.Config.MsDev.PlatformToolset)

		outputTarget := ""
		if c.Project.TypeIsExe() {
			executableFilename := c.Workspace.Config.ExeTargetPrefix + c.Project.Name + c.Workspace.Config.ExeTargetSuffix
			outputTarget = filepath.Join(BINDIR, executableFilename)
		} else if c.Project.TypeIsDll() {
			dllFilename := c.Workspace.Config.DllTargetPrefix + c.Project.Name + c.Workspace.Config.DllTargetSuffix
			outputTarget = filepath.Join(BINDIR, dllFilename)
			if c.Workspace.MakeTarget.OSIsWindows() {
				libFilename := c.Workspace.Config.LibTargetPrefix + c.Project.Name + c.Workspace.Config.LibTargetSuffix
				configResolved.OutputLib = NewFileEntryInit(filepath.Join(LIBDIR, libFilename), false)
			} else {
				configResolved.OutputLib = NewFileEntryInit(filepath.Join(BINDIR, dllFilename), false)
			}
		} else if c.Project.TypeIsLib() {
			libFilename := c.Workspace.Config.LibTargetPrefix + c.Project.Name + c.Workspace.Config.LibTargetSuffix
			outputTarget = filepath.Join(LIBDIR, libFilename)
			configResolved.OutputLib = NewFileEntryInit(outputTarget, false)
		}

		if len(outputTarget) > 0 {
			configResolved.OutputTarget = NewFileEntryInit(outputTarget, false)
		}
	}

	configMerged.Resolved = configResolved

	return configMerged
}
