package clay

import (
	"path/filepath"
	"time"

	"github.com/jurgen-kluft/ccode/clay/toolchain"
	"github.com/jurgen-kluft/ccode/clay/toolchain/deptrackr"
	"github.com/jurgen-kluft/ccode/denv"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

// Project represents a C/C++ project that can be built using the Clay build system.
// A project can be a library or an executable.
type Project struct {
	Toolchain    toolchain.Environment // Build environment for this project
	DevProject   *denv.DevProject      // Development environment project (if any)
	Config       []*denv.DevConfig     // Build configurations
	SourceFiles  []SourceFile          // C/C++ Source files for the library
	Dependencies []*Project            // Libraries that this project depends on
	Frameworks   []string              // Frameworks to link against (for macOS)
}

func NewProjectFromDevProject(devPrj *denv.DevProject, configs []*denv.DevConfig) *Project {
	return &Project{
		DevProject:   devPrj,
		Config:       configs,
		Toolchain:    nil,
		SourceFiles:  []SourceFile{},
		Dependencies: []*Project{},
	}
}

type SourceFile struct {
	SrcAbsPath string // Absolute path to the source file
	SrcRelPath string // Relative path of the source file (based on where it was globbed from)
}

func (p *Project) IsExecutable() bool {
	return p.DevProject.BuildType.IsExecutable()
}

func (p *Project) GetOutputFilepath(buildPath, filename string) string {
	return filepath.Join(buildPath, p.DevProject.Name, filename)
}

func (p *Project) GetBuildPath(buildPath string) string {
	return filepath.Join(buildPath, p.DevProject.Name)
}

func (p *Project) CanBuildFor(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) bool {
	if !p.DevProject.BuildTargets.HasOverlap(buildTarget) {
		return false
	}
	for _, cfg := range p.Config {
		if cfg.BuildConfig.Contains(buildConfig) {
			return true
		}
	}
	return false
}

func (p *Project) AddSourceFile(srcPath string, srcRelPath string) {
	p.SourceFiles = append(p.SourceFiles, SourceFile{
		SrcAbsPath: srcPath,
		SrcRelPath: srcRelPath,
	})
}

func (p *Project) GlobSourceFiles(isExcluded func(string) bool) {
	for _, srcDir := range p.DevProject.SourceDirs {
		srcPath := srcDir.Path.String()
		srcPath = corepkg.PathNormalize(srcPath)
		pattern := corepkg.PathNormalize(srcDir.Glob)

		dirFunc := func(rootPath, relPath string) bool {
			return true // We want to include all directories
		}

		dirpath := filepath.Join(srcPath)
		fileFunc := func(rootPath, relPath string) {
			if !isExcluded(relPath) {
				if match := corepkg.GlobMatching(relPath, pattern); match {
					p.AddSourceFile(filepath.Join(rootPath, relPath), relPath)
				}
			}
		}

		err := corepkg.FileEnumerate(dirpath, dirFunc, fileFunc)
		if err != nil {
			corepkg.LogErrorf(err, "failed to enumerate files in %q: %v", dirpath)
		}
	}
}

func (p *Project) AddLibrary(lib *Project) {
	p.Dependencies = append(p.Dependencies, lib)
}

func (p *Project) GetConfig(buildConfig denv.BuildConfig) *denv.DevConfig {
	for _, cfg := range p.Config {
		if cfg.BuildConfig.Contains(buildConfig) {
			return cfg
		}
	}
	return nil
}

type CompileContext struct {
	buildPath         string
	compiler          toolchain.Compiler
	depTrackr         deptrackr.FileTrackr
	srcFilesOutOfDate []SourceFile
	srcFilesUpToDate  []SourceFile
	absSrcFilepaths   []string
	objRelFilepaths   []string
}

func newCompileContext(buildPath string, compiler toolchain.Compiler, depTrackr deptrackr.FileTrackr, numSourceFiles int) *CompileContext {
	return &CompileContext{
		buildPath:         buildPath,
		depTrackr:         depTrackr,
		compiler:          compiler,
		srcFilesOutOfDate: make([]SourceFile, 0, numSourceFiles),
		srcFilesUpToDate:  make([]SourceFile, 0, numSourceFiles),
		absSrcFilepaths:   []string{},
		objRelFilepaths:   []string{},
	}
}

func (cc *CompileContext) queryItem(item string) bool {
	return cc.depTrackr.QueryItem(item)
}

func (cc *CompileContext) trackOutOfDateItem(item string, deps []string) {
	cc.depTrackr.AddItem(item, deps)
}

func (cc *CompileContext) trackUpToDateItem(item string) {
	cc.depTrackr.CopyItem(item)
}

func (cc *CompileContext) saveDependencyTrackr() error {
	_, err := cc.depTrackr.Save()
	return err
}

// collectFilesToCompile checks which source files are out-of-date and need to be recompiled.
// It returns the number of out-of-date source files.
func (cc *CompileContext) collectFilesToCompile(sourceFiles []SourceFile) int {
	for _, src := range sourceFiles {
		srcObjRelPath := filepath.Join(cc.buildPath, cc.compiler.ObjFilepath(src.SrcRelPath))
		if !cc.depTrackr.QueryItem(srcObjRelPath) {
			corepkg.DirMake(filepath.Dir(srcObjRelPath))
			cc.srcFilesOutOfDate = append(cc.srcFilesOutOfDate, src)
		} else {
			cc.srcFilesUpToDate = append(cc.srcFilesUpToDate, src)
		}
	}

	cc.absSrcFilepaths = make([]string, len(cc.srcFilesOutOfDate))
	cc.objRelFilepaths = make([]string, len(cc.srcFilesOutOfDate))
	for i, src := range cc.srcFilesOutOfDate {
		cc.absSrcFilepaths[i] = src.SrcAbsPath
		cc.objRelFilepaths[i] = filepath.Join(cc.buildPath, cc.compiler.ObjFilepath(src.SrcRelPath))
	}

	return len(cc.srcFilesOutOfDate)
}

func (cc *CompileContext) compile() error {

	// TODO, get back a list of files that where compiled successfully, use that to update the dependency tracker.
	// For now we recompile all files next time if an error occured.

	return cc.compiler.Compile(cc.absSrcFilepaths, cc.objRelFilepaths)
}

func (cc *CompileContext) updateDependencyTracker() {
	// Update the dependency tracker
	for _, src := range cc.srcFilesUpToDate {
		objRelFilepath := filepath.Join(cc.buildPath, cc.compiler.ObjFilepath(src.SrcRelPath))
		cc.depTrackr.CopyItem(objRelFilepath)
	}
	for _, src := range cc.srcFilesOutOfDate {
		objRelFilepath := filepath.Join(cc.buildPath, cc.compiler.ObjFilepath(src.SrcRelPath))
		depRelFilepath := filepath.Join(cc.buildPath, cc.compiler.DepFilepath(src.SrcRelPath))
		if mainItem, depItems, err := cc.depTrackr.ParseDependencyFile(src.SrcAbsPath, objRelFilepath, depRelFilepath); err == nil {
			cc.depTrackr.AddItem(mainItem, depItems)
		}
	}
}

func (p *Project) GetIncludesAndDefines(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) ([]string, []string) {
	// Collect all include directories from project and dependencies
	// Collect all defines from project and dependencies
	includes := corepkg.NewValueSet2(8)
	defines := corepkg.NewValueSet2(8)

	projects := make([]*Project, 0, len(p.Dependencies)+1)
	projects = append(projects, p)

	projectsAdded := map[string]int{}
	projectsAdded[p.DevProject.Name] = 0

	for len(projects) > 0 {
		// Pop a project from the list
		prj := projects[0]
		projects = projects[1:]

		// Add dependency projects to the list to process
		for _, dep := range prj.Dependencies {
			if dep.CanBuildFor(buildConfig, buildTarget) {
				if _, exists := projectsAdded[dep.DevProject.Name]; !exists {
					projects = append(projects, dep)
					projectsAdded[dep.DevProject.Name] = 1
				}
			}
		}

		prjConfig := prj.GetConfig(buildConfig)
		if prjConfig != nil {
			for _, incDir := range prjConfig.IncludeDirs {
				incPath := incDir.String()
				incPath = corepkg.PathNormalize(incPath)
				includes.Add(incPath)
			}
			for _, def := range prjConfig.Defines.Values {
				defines.Add(def)
			}
		}
	}
	return includes.Values, defines.Values
}

func (p *Project) Build(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget, buildPath string) (outOfDate int, err error) {
	config := p.GetConfig(buildConfig)

	includes, defines := p.GetIncludesAndDefines(buildConfig, buildTarget)
	projectBuildPath := p.GetBuildPath(buildPath)

	compiler := p.Toolchain.NewCompiler(buildConfig, buildTarget)
	compiler.SetupArgs(defines, includes)
	depTrackr := p.Toolchain.NewDependencyTracker(projectBuildPath)
	compilerContext := newCompileContext(projectBuildPath, compiler, depTrackr, len(p.SourceFiles))

	buildStartTime := time.Now()

	outOfDate = compilerContext.collectFilesToCompile(p.SourceFiles)
	if outOfDate > 0 {
		corepkg.LogInfof("Building project: %s, config: %s\n", p.DevProject.Name, config.BuildConfig.String())
		if err := compilerContext.compile(); err != nil {
			return outOfDate, err
		}
		compilerContext.updateDependencyTracker()
	}

	staticArchiver := p.Toolchain.NewArchiver(toolchain.ArchiverTypeStatic, buildConfig, buildTarget)

	if p.IsExecutable() {
		linker := p.Toolchain.NewLinker(buildConfig, buildTarget)
		linker.SetupArgs([]string{}, []string{})

		executableOutputFilepath := p.GetOutputFilepath(buildPath, linker.LinkedFilepath(p.DevProject.Name))

		if outOfDate > 0 || !compilerContext.queryItem(executableOutputFilepath) {
			if outOfDate == 0 {
				corepkg.LogInfof("Linking project: %s, config: %s\n", p.DevProject.Name, buildConfig.String())
				buildStartTime = time.Now()
				outOfDate += 1
			}

			// Project archive dependencies (only those matching the build config)
			archivesToLink := make([]string, 0, len(p.SourceFiles)+len(p.Dependencies))
			for _, dep := range p.Dependencies {
				if dep.CanBuildFor(buildConfig, buildTarget) {
					libAbsFilepath := dep.GetOutputFilepath(buildPath, staticArchiver.LibFilepath(dep.DevProject.Name))
					archivesToLink = append(archivesToLink, libAbsFilepath)
				}
			}

			// Link them all together into a single executable
			if err := linker.Link(compilerContext.objRelFilepaths, archivesToLink, executableOutputFilepath); err != nil {
				return outOfDate, err
			}

			compilerContext.trackOutOfDateItem(executableOutputFilepath, archivesToLink)
		} else {
			compilerContext.trackUpToDateItem(executableOutputFilepath)
		}

	} else {
		archiveOutputFilepath := p.GetOutputFilepath(buildPath, staticArchiver.LibFilepath(p.DevProject.Name))
		if outOfDate > 0 || !compilerContext.queryItem(archiveOutputFilepath) {
			if outOfDate == 0 {
				corepkg.LogInfof("Archiving project: %s, config: %s\n", p.DevProject.Name, buildConfig.String())
				buildStartTime = time.Now()
				outOfDate += 1
			}

			staticArchiver.SetupArgs()

			// Archive all object files into a static library using the static archiver
			if err := staticArchiver.Archive(compilerContext.objRelFilepaths, archiveOutputFilepath); err != nil {
				return outOfDate, err
			}

			compilerContext.trackOutOfDateItem(archiveOutputFilepath, compilerContext.objRelFilepaths)
		} else {
			compilerContext.trackUpToDateItem(archiveOutputFilepath)
		}
	}

	if err := compilerContext.saveDependencyTrackr(); err != nil {
		return outOfDate, err
	}

	if outOfDate > 0 {
		seconds := float64(time.Since(buildStartTime).Milliseconds()) / 1000.0
		corepkg.LogInfof("Building done ... (duration %.2f seconds)\n", seconds)
	}

	return outOfDate, nil
}

func (p *Project) Flash(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget, buildPath string) error {
	burner := p.Toolchain.NewBurner(buildConfig, buildTarget)

	buildPath = p.GetBuildPath(buildPath)

	burner.SetupBuild(buildPath)
	if err := burner.Build(); err != nil {
		return err
	}

	if err := burner.SetupBurn(buildPath); err != nil {
		return err
	}
	if err := burner.Burn(); err != nil {
		return err
	}

	return nil
}
