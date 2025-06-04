package ccode

import (
	"path/filepath"
	"strings"

	base "github.com/jurgen-kluft/ccode/base"
	"github.com/jurgen-kluft/ccode/denv"
	fixr "github.com/jurgen-kluft/ccode/include-fixer"
)

var DefaultHeaderFileExtensions = map[string]bool{".h": true, ".hpp": true, ".inl": true, ".inc": true}
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

func NewDefaultFixrConfig(prjname string, setting fixr.FixrSetting) *FixrConfig {
	return &FixrConfig{
		Setting:                setting,
		RenamePolicy:           NoFileNamingPolicy,
		IncludeGuardConfig:     fixr.NewIncludeGuardConfig(prjname),
		IncludeDirectiveConfig: fixr.NewIncludeDirectiveConfig(),
		HeaderFileFilter:       DefaultHeaderFileFilter,
		SourceFileFilter:       DefaultSourceFileFilter,
	}
}

func NewCCoreFixrConfig(prjname string, setting fixr.FixrSetting) *FixrConfig {
	return &FixrConfig{
		Setting:                setting,
		RenamePolicy:           CCoreFileNamingPolicy,
		IncludeGuardConfig:     fixr.NewIncludeGuardConfig(prjname),
		IncludeDirectiveConfig: fixr.NewIncludeDirectiveConfig(),
		HeaderFileFilter:       DefaultHeaderFileFilter,
		SourceFileFilter:       DefaultSourceFileFilter,
	}
}

func IncludeFixer(pkg *denv.Package, cfg *FixrConfig) {

	libraries := pkg.Libraries()      // All libraries, including dependencies
	mainProjects := pkg.Executables() // Main Unittest, App or Cli

	projects := make([]*denv.DevProject, 0, len(libraries)+len(mainProjects))
	projects = append(projects, libraries...)    // All libraries, including dependencies
	projects = append(projects, mainProjects...) // Main Unittest, App or Cli

	renamers := fixr.NewRenamers()
	scanners := fixr.NewDirScanner()
	fixers := fixr.NewFixers()

	// So we need the list of unique include directories of all the projects
	for _, p := range projects {
		for _, inc := range p.CollectIncludeDirs() {
			includePath := filepath.Join(inc.Root, inc.Base, inc.Sub)
			scanners.Add(includePath, cfg.HeaderFileFilter)
		}
	}

	// Then we need the source and include directories of the main application(s) and main library
	for _, mainProject := range mainProjects {
		//mainProjectPath := filepath.Join(basePath, mainProject.PackageURL)
		for _, sp := range mainProject.CollectSourceDirs() {
			sourcePath := filepath.Join(sp.Path.Root, sp.Path.Base, sp.Path.Sub)
			renamers.Add(sourcePath, cfg.RenamePolicy, cfg.SourceFileFilter, cfg.SourceFileFilter)
			fixers.AddSourceFileFilter(sourcePath, cfg.SourceFileFilter)
		}
		for _, inc := range mainProject.CollectIncludeDirs() {
			includePath := filepath.Join(inc.Root, inc.Base, inc.Sub)
			renamers.Add(includePath, cfg.RenamePolicy, cfg.SourceFileFilter, cfg.HeaderFileFilter)
			fixers.AddHeaderFileFilter(includePath, cfg.HeaderFileFilter)
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

func Generate(pkg *denv.Package, dryrun bool, verbose bool) {
	var setting fixr.FixrSetting
	if dryrun {
		setting |= fixr.DryRun
	}
	if verbose {
		setting |= fixr.Verbose
	}
	config := NewDefaultFixrConfig(pkg.RepoName, setting)
	IncludeFixer(pkg, config)
	base.Generate(pkg)
}

func GenerateFiles(pkg *denv.Package) {
	base.GenerateFiles(pkg)
}

func GenerateGitIgnore() {
	base.GenerateGitIgnore()
}

func GenerateTestMainCpp(has_ccore, has_cbase bool) {
	base.GenerateTestMainCpp(has_ccore, has_cbase)
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
