package ccode

import (
	"os"
	"path/filepath"
	"strings"

	base "github.com/jurgen-kluft/ccode/ccode-base"
	"github.com/jurgen-kluft/ccode/denv"
	fixr "github.com/jurgen-kluft/ccode/include-fixer"
)

var DefaultHeaderFileExtensions = map[string]bool{".h": true, ".hpp": true, ".inl": true}
var DefaultHeaderFileFilter = func(_filepath string) bool {
	_filepath = strings.ToLower(_filepath)
	ext := filepath.Ext(_filepath)
	_, ok := DefaultHeaderFileExtensions[ext]
	return ok
}

var DefaultSourceFileExtensions = map[string]bool{".cpp": true, ".c": true, ".cxx": true, ".mm": true, ".m": true}
var DefaultSourceFileFilter = func(_filepath string) bool {
	_filepath = strings.ToLower(_filepath)
	ext := filepath.Ext(_filepath)
	_, ok := DefaultSourceFileExtensions[ext]
	return ok
}

var NoFileNamingPolicy = func(_filepath string) (bool, string) {
	return false, _filepath
}

var CCoreFileNamingPolicy = func(_filepath string) (bool, string) {
	renamed := false
	if strings.HasSuffix(_filepath, ".hpp") {
		renamed = true
		_filepath = strings.TrimSuffix(_filepath, ".hpp") + ".h"
	} else if strings.HasSuffix(_filepath, ".cxx") {
		renamed = true
		_filepath = strings.TrimSuffix(_filepath, ".cxx") + ".cpp"
	}

	filename := filepath.Base(_filepath)
	if !strings.HasPrefix(filename, "c_") {
		renamed = true
		_filepath = strings.TrimSuffix(_filepath, filename) + "c_" + filename
	}

	return renamed, _filepath
}

type FixrConfig struct {
	Setting                fixr.FixrSetting
	RenamePolicy           func(_filepath string) (bool, string)
	IncludeGuardConfig     *fixr.IncludeGuardConfig
	IncludeDirectiveConfig *fixr.IncludeDirectiveConfig
	HeaderFileFilter       func(_filepath string) bool
	SourceFileFilter       func(_filepath string) bool
}

func NewDefaultFixrConfig(setting fixr.FixrSetting) *FixrConfig {
	return &FixrConfig{
		Setting:                setting,
		RenamePolicy:           NoFileNamingPolicy,
		IncludeGuardConfig:     nil,
		IncludeDirectiveConfig: fixr.NewIncludeDirectiveConfig(),
		HeaderFileFilter:       DefaultHeaderFileFilter,
		SourceFileFilter:       DefaultSourceFileFilter,
	}
}

func NewCCoreFixrConfig(setting fixr.FixrSetting) *FixrConfig {
	return &FixrConfig{
		Setting:                setting,
		RenamePolicy:           CCoreFileNamingPolicy,
		IncludeGuardConfig:     fixr.NewIncludeGuardConfig(),
		IncludeDirectiveConfig: fixr.NewIncludeDirectiveConfig(),
		HeaderFileFilter:       DefaultHeaderFileFilter,
		SourceFileFilter:       DefaultSourceFileFilter,
	}
}

func IncludeFixer(pkg *denv.Package, cfg *FixrConfig) {

	basePath := filepath.Join(os.Getenv("GOPATH"), "src")

	// Collect all projects, including dependencies
	libraries := pkg.Libraries()
	mainProjects := pkg.MainProjects()

	projects := make([]*denv.Project, 0, len(libraries)+len(mainProjects))
	projects = append(projects, libraries...)
	projects = append(projects, mainProjects...) // Main Unittest, App and/or Library

	renamers := fixr.NewRenamers()
	scanners := fixr.NewDirScanner()
	fixers := fixr.NewFixers()

	// So we need the list of unique include directories of all the projects
	for _, p := range projects {
		projectPath := filepath.Join(basePath, p.PackageURL)
		for _, inc := range p.IncludeDirs {
			includePath := filepath.Join(projectPath, inc)
			scanners.Add(includePath, cfg.HeaderFileFilter)
		}
	}

	// Then we need the source and include directories of the main application(s) and main library
	for _, mainProject := range mainProjects {
		mainProjectPath := filepath.Join(basePath, mainProject.PackageURL)
		for _, sp := range mainProject.SourceDirs {
			sourcePath := filepath.Join(mainProjectPath, sp)
			renamers.Add(sourcePath, cfg.RenamePolicy, cfg.SourceFileFilter, cfg.SourceFileFilter)
			fixers.Add(sourcePath, cfg.SourceFileFilter)
		}
		for _, inc := range mainProject.IncludeDirs {
			includePath := filepath.Join(mainProjectPath, inc)
			renamers.Add(includePath, cfg.RenamePolicy, cfg.SourceFileFilter, cfg.HeaderFileFilter)
			fixers.Add(includePath, cfg.HeaderFileFilter)
		}
	}

	// Create instance
	fixer := fixr.NewFixr(cfg.IncludeDirectiveConfig, cfg.IncludeGuardConfig)
	fixer.Setting = cfg.Setting

	if fixer.Rename() {
		fixer.ProcessRenamers(renamers)
	}
	fixer.ProcessScanners(scanners)
	fixer.ProcessFixers(fixers)
}

func Init() bool {
	return base.Init()
}

func Generate(pkg *denv.Package, config *FixrConfig) error {
	IncludeFixer(pkg, config)
	return base.Generate(pkg)
}

func GenerateFiles() {
	base.GenerateFiles()
}

func GenerateGitIgnore() {
	base.GenerateGitIgnore()
}

func GenerateTestMainCpp() {
	base.GenerateTestMainCpp()
}

func GenerateEmbedded() {
	base.GenerateEmbedded()
}

func GenerateClangFormat() {
	base.GenerateClangFormat()
}

func GenerateCppEnums(inputFile string, outputFile string) error {
	return base.GenerateCppEnums(inputFile, outputFile)
}
