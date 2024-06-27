package ccode

import (
	"os"
	"path/filepath"

	base "github.com/jurgen-kluft/ccode/ccode-base"
	"github.com/jurgen-kluft/ccode/denv"
	fixr "github.com/jurgen-kluft/ccode/include-fixer"
)

type FixrConfig struct {
	RenamePolicy           func(_filepath string) (bool, string)
	IncludeGuardConfig     *fixr.IncludeGuardConfig
	IncludeDirectiveConfig *fixr.IncludeDirectiveConfig
}

func NewDefaultFixrConfig() *FixrConfig {
	return &FixrConfig{
		RenamePolicy:           func(_filepath string) (bool, string) { return false, _filepath },
		IncludeGuardConfig:     fixr.NewIncludeGuardConfig(),
		IncludeDirectiveConfig: fixr.NewIncludeDirectiveConfig(),
	}
}

func NewCCoreFixrConfig() *FixrConfig {
	return &FixrConfig{
		RenamePolicy:           fixr.CCoreFileNamingPolicy,
		IncludeGuardConfig:     fixr.NewIncludeGuardConfig(),
		IncludeDirectiveConfig: fixr.NewIncludeDirectiveConfig(),
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
			scanners.Add(includePath, fixr.DefaultHeaderFileFilter)
		}
	}

	// Then we need the source and include directories of the main application(s) and main library
	for _, mainProject := range mainProjects {
		mainProjectPath := filepath.Join(basePath, mainProject.PackageURL)
		for _, sp := range mainProject.SourceDirs {
			sourcePath := filepath.Join(mainProjectPath, sp)
			renamers.Add(sourcePath, fixr.CCoreFileNamingPolicy, fixr.DefaultSourceFileFilter)
			fixers.Add(sourcePath, fixr.DefaultSourceFileFilter)
		}
		for _, inc := range mainProject.IncludeDirs {
			includePath := filepath.Join(mainProjectPath, inc)
			renamers.Add(includePath, fixr.CCoreFileNamingPolicy, fixr.DefaultHeaderFileFilter)
			fixers.Add(includePath, fixr.DefaultHeaderFileFilter)
		}
	}

	// Create instance
	includeDirectiveConfig := fixr.NewIncludeDirectiveConfig()
	includeGuardConfig := fixr.NewIncludeGuardConfig()
	fixer := fixr.NewFixr(includeDirectiveConfig, includeGuardConfig)

	fixer.Rename(renamers)
	fixer.Scan(scanners)
	fixer.Fix(fixers)
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
