package denv

import (
	"path/filepath"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
	"github.com/jurgen-kluft/ccode/dev"
)

// -----------------------------------------------------------------------------------------------------
// ConfigList
// -----------------------------------------------------------------------------------------------------

// -----------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------

type ConfigList struct {
	Dict   map[dev.BuildConfig]int
	Values []*Config
	Keys   []string
}

func NewConfigList() *ConfigList {
	return &ConfigList{
		Dict:   map[dev.BuildConfig]int{},
		Values: []*Config{},
		Keys:   []string{},
	}
}

func (p *ConfigList) Add(config *Config) {
	if _, ok := p.Dict[config.BuildConfig]; !ok {
		p.Dict[config.BuildConfig] = len(p.Values)
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

func (p *ConfigList) Get(t dev.BuildConfig) (*Config, bool) {
	if i, ok := p.Dict[t]; ok {
		return p.Values[i], true
	}
	return nil, false
}

func (p *ConfigList) Has(t dev.BuildConfig) bool {
	_, ok := p.Dict[t]
	return ok
}

func (p *ConfigList) CollectByWildcard(name string, list *ConfigList) {
	for _, p := range p.Values {
		if corepkg.PathMatchWildcard(p.String(), name, true) {
			list.Add(p)
		}
	}
}

// -----------------------------------------------------------------------------------------------------
// Config
// -----------------------------------------------------------------------------------------------------

type Config struct {
	BuildConfig       dev.BuildConfig
	Workspace         *Workspace
	Project           *Project
	CppDefines        *corepkg.KeyValueSet
	CppFlags          *corepkg.KeyValueSet
	IncludeDirs       *PinnedPathSet
	LibraryPaths      *PinnedPathSet
	LibraryFiles      *corepkg.ValueSet
	LibraryFrameworks *corepkg.ValueSet // MacOS specific
	LinkFlags         *corepkg.KeyValueSet
	DisableWarning    *corepkg.KeyValueSet

	XcodeSettings         *corepkg.KeyValueSet
	VisualStudioClCompile *corepkg.KeyValueSet
	VisualStudioLink      *corepkg.KeyValueSet

	GenDataXcode struct {
		ProjectConfigUuid corepkg.UUID
		TargetUuid        corepkg.UUID
		TargetConfigUuid  corepkg.UUID
	}

	Resolved *ConfigResolved
}

func NewConfig(t dev.BuildConfig, ws *Workspace, p *Project) *Config {
	c := &Config{}
	c.BuildConfig = t
	c.Workspace = ws
	c.Project = p

	c.CppDefines = corepkg.NewKeyValueSet() // e.g. "DEBUG" "PROFILE"
	c.CppFlags = corepkg.NewKeyValueSet()   // e.g. "-g"
	c.IncludeDirs = NewPinnedPathSet()      // e.g. "source/main/include", "source/test/include"

	c.LibraryFrameworks = corepkg.NewValueSet() // e.g. "Foundation", "Cocoa"
	c.LibraryPaths = NewPinnedPathSet()         // e.g. "source/main/lib", "source/test/lib"
	c.LibraryFiles = corepkg.NewValueSet()      // e.g. "libfoo.a", "libbar.a"

	c.LinkFlags = corepkg.NewKeyValueSet()      // e.g. "-lstdc++"
	c.DisableWarning = corepkg.NewKeyValueSet() // e.g. "unused-variable"

	c.XcodeSettings = corepkg.NewKeyValueSet()
	c.VisualStudioClCompile = corepkg.NewKeyValueSet()
	c.VisualStudioLink = corepkg.NewKeyValueSet()

	c.GenDataXcode.ProjectConfigUuid = corepkg.GenerateUUID()
	c.GenDataXcode.TargetUuid = corepkg.GenerateUUID()
	c.GenDataXcode.TargetConfigUuid = corepkg.GenerateUUID()

	c.InitTargetSettings()
	c.InitXcodeSettings()
	c.InitVisualStudioSettings()

	return c
}

func (c *Config) String() string {
	return c.BuildConfig.AsString()
}

// AddIncludeDir adds an include to the list of include directories
func (c *Config) AddIncludeDir(includeDir dev.PinnedPath) {
	c.IncludeDirs.AddOrSet(includeDir)
}

func (c *Config) AddLibrary(projectDirectory string, lib dev.PinnedFilepath) {
	c.LibraryPaths.AddOrSet(lib.Path)

	libfile := lib.Filename
	c.LibraryFiles.Add(libfile)
}

func (c *Config) AddFramework(framework string) {
	c.LibraryFrameworks.Add(framework)
}

func (c *Config) InitTargetSettings() {

	buildTarget := c.Workspace.BuildTarget

	if buildTarget.Windows() {
		c.CppDefines.AddOrSet("TARGET_PC", "TARGET_PC")
		c.CppDefines.AddOrSet("UNICODE", "UNICODE")
		c.CppDefines.AddOrSet("_UNICODE", "_UNICODE")
	} else if buildTarget.Linux() {
		c.CppDefines.AddOrSet("TARGET_LINUX", "TARGET_LINUX")
		c.CppDefines.AddOrSet("UNICODE", "UNICODE")
		c.CppDefines.AddOrSet("_UNICODE", "_UNICODE")
	} else if buildTarget.Mac() {
		c.CppDefines.AddOrSet("TARGET_MAC", "TARGET_MAC")
		c.CppDefines.AddOrSet("UNICODE", "UNICODE")
		c.CppDefines.AddOrSet("_UNICODE", "_UNICODE")

		c.LinkFlags.AddOrSet("-ObjC", "")

		for _, cocoa := range []string{"Foundation", "Cocoa", "Carbon", "Metal", "OpenGL", "IOKit", "AppKit", "CoreVideo", "QuartzCore"} {
			c.AddFramework(cocoa)
		}

		// func (l *Library2) Merge(other *Library2) {
		// 	l.Frameworks.Merge(other.Frameworks)
		// 	l.Files.Merge(other.Files)
		// 	l.Libs.Merge(other.Libs)
		// 	l.Dirs.Merge(other.Dirs)
		// }

	} else if buildTarget.Arduino() {
		c.CppDefines.AddOrSet("TARGET_ESP32", "TARGET_ESP32")
	}
}

func (c *Config) InitXcodeSettings() {

	settings := make(map[string]string)

	if c.Workspace.BuildTarget.Mac() {
		settings["SDKROOT"] = "iphoneos"
		settings["SUPPORTED_PLATFORMS"] = "iphonesimulator iphoneos"
		settings["IPHONEOS_DEPLOYMENT_TARGET"] = "10.1"
	} else if c.Workspace.BuildTarget.Mac() {
		settings["SDKROOT"] = "macosx"
		settings["SUPPORTED_PLATFORMS"] = "macosx"
		settings["MACOSX_DEPLOYMENT_TARGET"] = "10.15" // c++11 require 10.10+
	}

	if c.BuildConfig.IsDebug() {
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

	c.XcodeSettings = corepkg.NewKeyValueSet()
	for k, v := range settings {
		c.XcodeSettings.AddOrSet(k, v)
	}
}

func (c *Config) InitVisualStudioSettings() {
	c.VisualStudioClCompile = corepkg.NewKeyValueSet()
	c.VisualStudioClCompile.AddOrSet("MinimalRebuild", "false")
	c.VisualStudioClCompile.AddOrSet("ExceptionHandling", "false")
	c.VisualStudioClCompile.AddOrSet("CompileAs", "CompileAsCpp")
	c.VisualStudioClCompile.AddOrSet("EnableModules", "false")
	c.VisualStudioClCompile.AddOrSet("TreatWarningAsError", "true")
	c.VisualStudioClCompile.AddOrSet("WarningLevel", "Level3") // Level0, Level1, Level2, Level3, Level4

	if c.Workspace.Config.Dev.CompilerIsClang() {
		c.VisualStudioClCompile.AddOrSet("DebugInformationFormat", "None")
	} else {
		if c.BuildConfig.IsFinal() == false {
			c.VisualStudioClCompile.AddOrSet("DebugInformationFormat", "ProgramDatabase")
		}
	}

	if c.BuildConfig.IsDebug() {
		c.VisualStudioClCompile.AddOrSet("Optimization", "Disabled")
		c.VisualStudioClCompile.AddOrSet("OmitFramePointers", "false")
	} else {
		c.VisualStudioClCompile.AddOrSet("Optimization", "Full") // MinSpace, MaxSpeed
		c.VisualStudioClCompile.AddOrSet("OmitFramePointers", "true")
		c.VisualStudioClCompile.AddOrSet("FunctionLevelLinking", "true")
		c.VisualStudioClCompile.AddOrSet("IntrinsicFunctions", "true")
		c.VisualStudioClCompile.AddOrSet("WholeProgramOptimization", "true")
		c.VisualStudioClCompile.AddOrSet("BasicRuntimeChecks", "Default")
	}

	c.VisualStudioClCompile.AddOrSet("RuntimeLibrary", c.Workspace.Config.MsDev.RuntimeLibrary.String(c.BuildConfig.IsDebug()))
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
	nc := NewConfig(c.BuildConfig, c.Workspace, c.Project)

	nc.CppDefines = c.CppDefines.Copy()
	nc.CppFlags = c.CppFlags.Copy()
	nc.IncludeDirs = c.IncludeDirs.Copy()
	nc.LibraryFrameworks = c.LibraryFrameworks.Copy()
	nc.LibraryFiles = c.LibraryFiles.Copy()
	nc.LibraryPaths = c.LibraryPaths.Copy()
	nc.LibraryFiles = c.LibraryFiles.Copy()
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
		configMerged.LibraryFrameworks.Merge(otherConfig.LibraryFrameworks)
		configMerged.LibraryFiles.Merge(otherConfig.LibraryFiles)
		configMerged.LibraryPaths.Merge(otherConfig.LibraryPaths)
		configMerged.LibraryFiles.Merge(otherConfig.LibraryFiles)
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

	emitCoreDefines := false
	for _, deps := range c.Project.Dependencies.Values {
		if deps.Name == "ccore" {
			emitCoreDefines = true
			break
		}
	}
	if emitCoreDefines {
		configMerged.CppDefines.AddOrSet("CCORE_GEN_CPU", "CCORE_GEN_CPU_"+strings.ToUpper(c.Workspace.BuildTarget.ArchAsString()))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_OS", "CCORE_GEN_OS_"+strings.ToUpper(c.Workspace.BuildTarget.OSAsString()))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_COMPILER", "CCORE_GEN_COMPILER_"+strings.ToUpper(c.Workspace.Config.Dev.CompilerAsString()))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_GENERATOR", "CCORE_GEN_GENERATOR_"+strings.ToUpper(c.Workspace.Config.Dev.ToString()))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_CONFIG", "CCORE_GEN_CONFIG_"+strings.ToUpper(c.String()))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_PLATFORM_NAME", "CCORE_GEN_PLATFORM_NAME=\""+strings.ToUpper(c.Workspace.BuildTarget.OSAsString()+"\""))
		configMerged.CppDefines.AddOrSet("CCORE_GEN_PROJECT", "CCORE_GEN_PROJECT_"+strings.ToUpper(c.Project.Name))
		genType := strings.ToUpper(c.Project.BuildType.String())
		genType = strings.ReplaceAll(genType, " ", "_")
		configMerged.CppDefines.AddOrSet("CCORE_GEN_TYPE", "CCORE_GEN_TYPE_"+genType)
	}

	if c.Project != nil {
		BINDIR := filepath.Join(c.Workspace.GenerateAbsPath, "bin", c.Project.Name, c.String()+"_"+c.Workspace.BuildTarget.ArchAsString()+"_"+c.Workspace.Config.MsDev.PlatformToolset)
		LIBDIR := filepath.Join(c.Workspace.GenerateAbsPath, "lib", c.Project.Name, c.String()+"_"+c.Workspace.BuildTarget.ArchAsString()+"_"+c.Workspace.Config.MsDev.PlatformToolset)

		outputTarget := ""
		if c.Project.TypeIsExe() {
			executableFilename := c.Workspace.Config.ExeTargetPrefix + c.Project.Name + c.Workspace.Config.ExeTargetSuffix
			outputTarget = filepath.Join(BINDIR, executableFilename)
		} else if c.Project.TypeIsDll() {
			dllFilename := c.Workspace.Config.DllTargetPrefix + c.Project.Name + c.Workspace.Config.DllTargetSuffix
			outputTarget = filepath.Join(BINDIR, dllFilename)
			if c.Workspace.BuildTarget.Windows() {
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
