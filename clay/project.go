package clay

import (
	"path/filepath"
	"time"

	"github.com/jurgen-kluft/ccode/clay/toolchain"
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

func (p *Project) Build(buildConfig denv.BuildConfig, buildTarget denv.BuildTarget, buildPath string) (outOfDate int, err error) {
	config := p.GetConfig(buildConfig)

	// Collect all include directories from project and dependencies
	// Collect all defines from project and dependencies
	includes := corepkg.NewValueSet2(8)
	defines := corepkg.NewValueSet2(8)

	projects := make([]*Project, 0, len(p.Dependencies)+1)
	projects = append(projects, p)

	projectTypes := []string{"main", "dependency"}

	projectsAdded := map[string]int{}
	projectsAdded[p.DevProject.Name] = 0

	for len(projects) > 0 {
		// Pop a project from the list
		prj := projects[0]
		prjType := projectTypes[projectsAdded[prj.DevProject.Name]]
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
				// For dependency projects, only include public include dirs
				if prjType == "main" || !incDir.Private {
					includes.Add(incPath)
				}
			}
			for _, def := range prjConfig.Defines.Values {
				defines.Add(def)
			}
		}
	}

	compiler := p.Toolchain.NewCompiler(buildConfig, buildTarget)
	compiler.SetupArgs(defines.Values, includes.Values)

	projectBuildPath := p.GetBuildPath(buildPath)
	projectDepFileTrackr := p.Toolchain.NewDependencyTracker(projectBuildPath)

	srcFilesOutOfDate := make([]SourceFile, 0, len(p.SourceFiles))
	srcFilesUpToDate := make([]SourceFile, 0, len(p.SourceFiles))
	for _, src := range p.SourceFiles {
		srcObjRelPath := filepath.Join(projectBuildPath, compiler.ObjFilepath(src.SrcRelPath))
		if !projectDepFileTrackr.QueryItem(srcObjRelPath) {
			corepkg.DirMake(filepath.Dir(srcObjRelPath))
			srcFilesOutOfDate = append(srcFilesOutOfDate, src)
		} else {
			srcFilesUpToDate = append(srcFilesUpToDate, src)
		}
	}

	absSrcFilepaths := make([]string, len(srcFilesOutOfDate))
	objRelFilepaths := make([]string, len(srcFilesOutOfDate))
	for i, src := range srcFilesOutOfDate {
		absSrcFilepaths[i] = src.SrcAbsPath
		objRelFilepaths[i] = filepath.Join(projectBuildPath, compiler.ObjFilepath(src.SrcRelPath))
	}

	buildStartTime := time.Now()
	outOfDate = len(srcFilesOutOfDate)
	if outOfDate > 0 {
		corepkg.LogInfof("Building project: %s, config: %s\n", p.DevProject.Name, config.BuildConfig.String())
		buildStartTime = time.Now()

		// Give the compiler the array of out-of-date source files (input) and their object files (output)
		if err := compiler.Compile(absSrcFilepaths, objRelFilepaths); err != nil {
			return outOfDate, err
		}

		// Update the dependency tracker
		for _, src := range srcFilesUpToDate {
			objRelFilepath := filepath.Join(projectBuildPath, compiler.ObjFilepath(src.SrcRelPath))
			projectDepFileTrackr.CopyItem(objRelFilepath)
		}
		for _, src := range srcFilesOutOfDate {
			objRelFilepath := filepath.Join(projectBuildPath, compiler.ObjFilepath(src.SrcRelPath))
			depRelFilepath := filepath.Join(projectBuildPath, compiler.DepFilepath(src.SrcRelPath))
			if mainItem, depItems, err := projectDepFileTrackr.ParseDependencyFile(src.SrcAbsPath, objRelFilepath, depRelFilepath); err == nil {
				projectDepFileTrackr.AddItem(mainItem, depItems)
			}
		}
	}

	staticArchiver := p.Toolchain.NewArchiver(toolchain.ArchiverTypeStatic, buildConfig, buildTarget)

	if p.IsExecutable() {
		linker := p.Toolchain.NewLinker(buildConfig, buildTarget)
		linker.SetupArgs([]string{}, []string{})

		executableOutputFilepath := p.GetOutputFilepath(buildPath, linker.LinkedFilepath(p.DevProject.Name))

		if outOfDate > 0 || !projectDepFileTrackr.QueryItem(executableOutputFilepath) {
			if outOfDate == 0 {
				corepkg.LogInfof("Linking project: %s, config: %s\n", p.DevProject.Name, buildConfig.String())
				buildStartTime = time.Now()
				outOfDate += 1
			}

			// Project object files
			objsToLink := make([]string, 0, len(p.SourceFiles)+len(p.Dependencies))
			for _, src := range p.SourceFiles {
				srcObjRelPath := filepath.Join(projectBuildPath, compiler.ObjFilepath(src.SrcRelPath))
				objsToLink = append(objsToLink, srcObjRelPath)
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
			if err := linker.Link(objsToLink, archivesToLink, executableOutputFilepath); err != nil {
				return outOfDate, err
			}
			projectDepFileTrackr.AddItem(executableOutputFilepath, archivesToLink)

		} else {
			projectDepFileTrackr.CopyItem(executableOutputFilepath)
		}

	} else {
		archiveOutputFilepath := p.GetOutputFilepath(buildPath, staticArchiver.LibFilepath(p.DevProject.Name))
		if outOfDate > 0 || !projectDepFileTrackr.QueryItem(archiveOutputFilepath) {
			if outOfDate == 0 {
				corepkg.LogInfof("Archiving project: %s, config: %s\n", p.DevProject.Name, buildConfig.String())
				buildStartTime = time.Now()
				outOfDate += 1
			}

			staticArchiver.SetupArgs()

			objFilesToArchive := make([]string, 0, len(p.SourceFiles))
			for _, src := range p.SourceFiles {
				srcObjRelPath := filepath.Join(projectBuildPath, compiler.ObjFilepath(src.SrcRelPath))
				objFilesToArchive = append(objFilesToArchive, srcObjRelPath)
			}
			if err := staticArchiver.Archive(objFilesToArchive, archiveOutputFilepath); err != nil {
				return outOfDate, err
			}

			projectDepFileTrackr.AddItem(archiveOutputFilepath, objFilesToArchive)
		} else {
			projectDepFileTrackr.CopyItem(archiveOutputFilepath)
		}
	}

	_, err = projectDepFileTrackr.Save()
	if err != nil {
		return outOfDate, err
	}

	if outOfDate > 0 {
		corepkg.LogInfof("Building done ... (duration %s)\n", time.Since(buildStartTime).Round(time.Second))
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
