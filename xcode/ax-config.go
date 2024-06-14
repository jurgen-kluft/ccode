package xcode

import (
	"path/filepath"
	"strings"
)

type ConfigEntry struct {
	AbsPath string
	Path    string
	IsAbs   bool
}

func NewConfigEntry(absPath string, isAbs bool, ws *Workspace) *ConfigEntry {
	e := &ConfigEntry{}
	e.Init(absPath, isAbs, ws)
	return e
}

func (e *ConfigEntry) Init(absPath string, isAbs bool, ws *Workspace) {
	e.IsAbs = isAbs
	e.AbsPath = absPath
	if isAbs {
		e.Path = absPath
	} else {
		e.Path = PathGetRel(e.AbsPath, ws.BuildDir)
	}
}

type ConfigEntryDict struct {
	entries map[string]int
	list    []*ConfigEntry
}

func NewConfigEntryDict() *ConfigEntryDict {
	d := &ConfigEntryDict{}
	d.entries = make(map[string]int)
	d.list = make([]*ConfigEntry, 0)
	return d
}

// add(const StrView& value, const StrView& fromDir)
// {
// String key;
// 	bool isAbs = true;

// 	if (!fromDir) {
// 		key = value;
// 	}else{
// 		isAbs = Path::isAbs(value);
// 		if (isAbs) {
// 			key = value;
// 		}else{
// 			Path::makeFullPath(key, fromDir, value);
// 		}
// 	}

// 	auto* e = _dict.find(key);
// 	if (!e) {
// 		e = _dict.add(key);
// 	}
// 	e->init(key, isAbs);
// 	return e;
// }

func (d *ConfigEntryDict) Add(value string, fromDir string, ws *Workspace) *ConfigEntry {
	key := ""
	isAbs := true

	if len(fromDir) == 0 {
		key = value
	} else {
		isAbs = PathIsAbs(fromDir)
		if isAbs {
			key = fromDir
		} else {
			key = PathMakeFullPath(fromDir, value)
		}
	}

	i, ok := d.entries[key]
	if !ok {
		i = len(d.list)
		d.entries[key] = i
		d.list = append(d.list, NewConfigEntry(key, isAbs, ws))
	} else {
		d.list[i].Init(key, isAbs, ws)
	}

	return d.list[i]
}

func (d *ConfigEntryDict) Extend(rhs *ConfigEntryDict) {
	for key, value := range rhs.entries {
		d.entries[key] = value
	}
}

func (d *ConfigEntryDict) UniqueExtend(rhs *ConfigEntryDict) {
	for key, value := range rhs.entries {
		if _, ok := d.entries[key]; !ok {
			d.entries[key] = value
		}
	}
}

type StringStringDict struct {
	entries map[string]int
	keys    []string
	list    []string
}

func NewStringStringDict() *StringStringDict {
	d := &StringStringDict{}
	d.entries = make(map[string]int)
	d.keys = make([]string, 0)
	d.list = make([]string, 0)
	return d
}

func (d *StringStringDict) Extend(rhs *StringStringDict) {
	for key, value := range rhs.entries {
		d.Add(key, rhs.list[value])
	}
}

func (d *StringStringDict) UniqueExtend(rhs *StringStringDict) {
	for key, value := range rhs.entries {
		if _, ok := d.entries[key]; !ok {
			d.entries[key] = value
		}
	}
}

func (d *StringStringDict) Add(key string, value string) {
	i, ok := d.entries[key]
	if !ok {
		d.entries[key] = len(d.list)
		d.keys = append(d.keys, key)
		d.list = append(d.list, value)
	} else {
		d.list[i] = value
	}
}

type ConfigSettings struct {
	Name            string
	AddDict         *ConfigEntryDict
	RemoveDict      *ConfigEntryDict
	LocalAddDict    *ConfigEntryDict
	LocalRemoveDict *ConfigEntryDict
	InheritDict     *ConfigEntryDict
	FinalDict       *ConfigEntryDict
	IsPath          bool
}

func (s *ConfigSettings) Inherit(rhs *ConfigSettings) {
	s.AddDict.Extend(rhs.AddDict)
}

func (s *ConfigSettings) ComputeFinal() {
	s.InheritDict.Extend(s.AddDict)
	s.FinalDict.Extend(s.InheritDict)
	s.FinalDict.Extend(s.LocalAddDict)
}

type Config struct {
	Name             string
	Workspace        *Workspace
	IsDebug          bool
	CppStd           string
	CppEnableModules bool
	WarningAsError   bool
	WarningLevel     string
	Project          *Project
	OutputTarget     *FileEntry
	OutputLib        *FileEntry
	BuildTmpDir      *FileEntry

	OutTargetDir    string
	ExeTargetPrefix string
	ExeTargetSuffix string
	DllTargetPrefix string
	DllTargetSuffix string
	LibTargetPrefix string
	LibTargetSuffix string

	CppDefines     *ConfigSettings
	CppFlags       *ConfigSettings
	IncludeDirs    *ConfigSettings
	IncludeFiles   *ConfigSettings
	LinkDirs       *ConfigSettings
	LinkLibs       *ConfigSettings
	LinkFiles      *ConfigSettings
	LinkFlags      *ConfigSettings
	DisableWarning *ConfigSettings

	Settings []*ConfigSettings

	XcodeSettings    *StringStringDict
	VstudioClCompile *StringStringDict
	VstudioLink      *StringStringDict

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

func (c *Config) registerConfigSetting(s *ConfigSettings) {
	c.Name = s.Name
	c.Settings = append(c.Settings, s)
}

func NewConfig(name string, ws *Workspace) *Config {
	c := &Config{}
	c.Name = name
	c.Workspace = ws

	c.CppDefines = &ConfigSettings{}
	c.IncludeDirs = &ConfigSettings{}
	c.IncludeFiles = &ConfigSettings{}
	c.LinkDirs = &ConfigSettings{}
	c.LinkLibs = &ConfigSettings{}
	c.LinkFiles = &ConfigSettings{}

	c.CppDefines.IsPath = true
	c.IncludeDirs.IsPath = true
	c.IncludeFiles.IsPath = true
	c.LinkDirs.IsPath = true
	c.LinkLibs.IsPath = false
	c.LinkFiles.IsPath = true

	c.CppStd = "c++14"

	c.registerConfigSetting(c.CppDefines)
	c.registerConfigSetting(c.CppFlags)
	c.registerConfigSetting(c.IncludeDirs)
	c.registerConfigSetting(c.IncludeFiles)
	c.registerConfigSetting(c.LinkDirs)
	c.registerConfigSetting(c.LinkLibs)
	c.registerConfigSetting(c.LinkFiles)

	if ws.MakeTarget.OSIsWindows() {
		c.ExeTargetSuffix = ".exe"
		c.DllTargetSuffix = ".dll"
	} else {
		c.ExeTargetSuffix = ""
		c.DllTargetSuffix = ".so"
	}

	if ws.MakeTarget.CompilerIsVc() {
		c.LibTargetSuffix = ".lib"
	} else {
		c.LibTargetPrefix = "lib"
		c.LibTargetSuffix = ".a"
	}

	c.Settings = make([]*ConfigSettings, 0)

	c.XcodeSettings = NewStringStringDict()
	c.VstudioClCompile = NewStringStringDict()
	c.VstudioLink = NewStringStringDict()

	c.Resolved = false

	c.GenDataMakefile.CppObjDir = filepath.Join(ws.BuildDir, "_build_tmp", name)

	c.GenDataXcode.ProjectConfigUuid = GenerateUUID()
	c.GenDataXcode.TargetUuid = GenerateUUID()
	c.GenDataXcode.TargetConfigUuid = GenerateUUID()

	return c
}

func NewDefaultConfig(ws *Workspace) *Config {
	return NewConfig("Default", ws)
}

func (c *Config) Init(proj *Project, source *Config, name string) {
	c.Project = proj
	c.Name = name
	if c.Name == "Debug" {
		c.IsDebug = true
	}

	if proj != nil {
		c.BuildTmpDir = NewFileEntryInit(filepath.Join(c.Workspace.BuildDir, "_build_tmp/", name, "/", proj.Name), false, true, c.Workspace)
	}

	if source != nil {
		c.CppStd = source.CppStd
		c.CppEnableModules = source.CppEnableModules
		c.WarningAsError = source.WarningAsError
		c.WarningLevel = source.WarningLevel
		c.XcodeSettings = source.XcodeSettings
		c.VstudioClCompile = source.VstudioClCompile
	} else {
		c.initXcodeSettings()
		c.initVstudioSettings()
	}
}

func (c *Config) initXcodeSettings() {

	settings := make(map[string]string)

	if c.Workspace.MakeTarget.OSIsIos() {
		settings["SDKROOT"] = "iphoneos"
		settings["SUPPORTED_PLATFORMS"] = "iphonesimulator iphoneos"
		settings["IPHONEOS_DEPLOYMENT_TARGET"] = "10.1"
	} else if c.Workspace.MakeTarget.OSIsMac() {
		settings["SDKROOT"] = "macosx"
		settings["SUPPORTED_PLATFORMS"] = "macosx"
		settings["MACOSX_DEPLOYMENT_TARGET"] = "10.10" // c++11 require 10.10+
	}

	if c.IsDebug {
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

	c.XcodeSettings = NewStringStringDict()
	for k, v := range settings {
		c.XcodeSettings.Add(k, v)
	}
}

func (c *Config) initVstudioSettings() {
	c.VstudioClCompile = NewStringStringDict()
	c.VstudioClCompile.Add("MinimalRebuild", "false")
}

func (c *Config) Inherit(rhs *Config) {
	c.resolve()

	for i, p := range c.Settings {
		p.Inherit(rhs.Settings[i])
	}

	if t := rhs.OutputLib.Path; len(t) > 0 {
		c.LinkFiles.InheritDict.Add(t, c.Workspace.BuildDir, c.Workspace)
	}

	c.XcodeSettings.UniqueExtend(rhs.XcodeSettings)
	c.VstudioClCompile.UniqueExtend(rhs.VstudioClCompile)
	c.VstudioLink.UniqueExtend(rhs.VstudioLink)
}

func (c *Config) ComputeFinal() {
	c.resolve()

	for _, p := range c.Settings {
		p.ComputeFinal()
	}
}

func (c *Config) resolve() {
	if c.Resolved {
		return
	}
	c.Resolved = true

	if c.Project == nil {
		return
	}

	defines := []string{}
	defines = append(defines, "AX_GEN_CPU_"+strings.ToUpper(c.Workspace.MakeTarget.ArchAsString()))
	defines = append(defines, "AX_GEN_OS_"+strings.ToUpper(c.Workspace.MakeTarget.OSAsString()))
	defines = append(defines, "AX_GEN_COMPILER_"+strings.ToUpper(c.Workspace.MakeTarget.CompilerAsString()))
	defines = append(defines, "AX_GEN_GENERATOR_"+strings.ToUpper(c.Workspace.Generator))
	defines = append(defines, "AX_GEN_CONFIG_"+strings.ToUpper(c.Name))
	defines = append(defines, "AX_GEN_PLATFORM_NAME=\""+strings.ToUpper(c.Workspace.PlatformName+"\""))
	defines = append(defines, "AX_GEN_PROJECT_"+strings.ToUpper(c.Project.Name))
	defines = append(defines, "AX_GEN_TYPE_"+strings.ToUpper(c.Project.Input.Type))

	for _, define := range defines {
		c.CppDefines.FinalDict.Add(define, "", c.Workspace)
	}

	tmp := ""
	if c.Project.TypeIsExe() {
		tmp = filepath.Join(c.Workspace.BuildDir, "bin", c.Name, c.OutTargetDir, c.ExeTargetPrefix, c.Project.Name, c.ExeTargetSuffix)
	} else if c.Project.TypeIsDll() {
		tmp = filepath.Join(c.Workspace.BuildDir, "bin", c.Name, c.OutTargetDir, c.DllTargetPrefix, c.Project.Name, c.DllTargetSuffix)
		if c.Workspace.MakeTarget.OSIsWindows() {
			c.OutputLib = NewFileEntryInit(filepath.Join(c.Workspace.BuildDir, "lib", c.Name, c.LibTargetPrefix, c.Project.Name, c.LibTargetSuffix), false, false, c.Workspace)
		} else {
			c.OutputLib = NewFileEntryInit(filepath.Join(c.Workspace.BuildDir, "bin", c.Name, c.OutTargetDir, c.DllTargetPrefix, c.Project.Name, c.DllTargetSuffix), false, false, c.Workspace)
		}
	} else if c.Project.TypeIsLib() {
		tmp = filepath.Join(c.Workspace.BuildDir, "lib", c.Name, c.OutTargetDir, c.LibTargetPrefix, c.Project.Name, c.LibTargetSuffix)
		c.OutputLib = NewFileEntryInit(tmp, false, false, c.Workspace)
	}

	if len(tmp) > 0 {
		c.OutputTarget = NewFileEntryInit(tmp, false, false, c.Workspace)
	}
}
