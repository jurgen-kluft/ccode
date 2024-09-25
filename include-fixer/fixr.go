package fixr

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type dirScanItem struct {
	dir    string
	filter func(filename string) bool
}

type renameItem struct {
	dir                string
	renameFunc         func(_filepath string) (bool, string)
	isSourceFileFilter func(filename string) bool
	isHeaderFileFilter func(filename string) bool
}

const (
	SourceFileType = 0
	HeaderFileType = 1
)

type fixerItem struct {
	dir      string
	fileType int
	filter2  func(filename string) bool
}

type DirScanner struct {
	duplicateMap map[string]bool
	scanners     []dirScanItem
}

func NewDirScanner() *DirScanner {
	return &DirScanner{duplicateMap: make(map[string]bool), scanners: make([]dirScanItem, 0)}
}

func (ds *DirScanner) Add(dir string, filter func(filename string) bool) {
	if _, ok := ds.duplicateMap[dir]; !ok {
		ds.scanners = append(ds.scanners, dirScanItem{dir: dir, filter: filter})
		ds.duplicateMap[dir] = true
	}
}

type Renamers struct {
	duplicateMap map[string]bool
	renamers     []renameItem
}

func NewRenamers() *Renamers {
	return &Renamers{duplicateMap: make(map[string]bool), renamers: make([]renameItem, 0)}
}

func (r *Renamers) Add(dir string, renameFunc func(_filepath string) (bool, string), isSourceFileFilter func(filename string) bool, isHeaderFileFilter func(filename string) bool) {
	if _, ok := r.duplicateMap[dir]; !ok {
		r.renamers = append(r.renamers, renameItem{dir: dir, renameFunc: renameFunc, isSourceFileFilter: isSourceFileFilter, isHeaderFileFilter: isHeaderFileFilter})
		r.duplicateMap[dir] = true
	}
}

type Fixers struct {
	duplicateMap map[string]bool
	fileFixers   []fixerItem
}

func NewFixers() *Fixers {
	return &Fixers{duplicateMap: make(map[string]bool), fileFixers: make([]fixerItem, 0)}
}

func (f *Fixers) AddSourceFileFilter(dir string, filter func(filename string) bool) {
	if _, ok := f.duplicateMap[dir]; !ok {
		f.fileFixers = append(f.fileFixers, fixerItem{dir: dir, fileType: SourceFileType, filter2: filter})
		f.duplicateMap[dir] = true
	}
}

func (f *Fixers) AddHeaderFileFilter(dir string, filter func(filename string) bool) {
	if _, ok := f.duplicateMap[dir]; !ok {
		f.fileFixers = append(f.fileFixers, fixerItem{dir: dir, fileType: HeaderFileType, filter2: filter})
		f.duplicateMap[dir] = true
	}
}

type ffile struct {
	name string // File name
	path string // Full path
	dir  *fdir
}

type fdir struct {
	name     string // Directory name
	path     string // Full path
	parent   *fdir
	children []*fdir
	files    []*ffile
}

func newFFile(name string, path string, dir *fdir) *ffile {
	return &ffile{name: name, path: path, dir: dir}
}

func newFDirRoot(path string) *fdir {
	d := &fdir{name: ".", path: path}
	d.parent = nil
	d.children = make([]*fdir, 0)
	d.files = make([]*ffile, 0)
	return d
}

func newFDir(name string, path string, parent *fdir) *fdir {
	d := &fdir{name: name, path: path}
	d.parent = parent
	d.children = make([]*fdir, 0)
	d.files = make([]*ffile, 0)
	parent.children = append(parent.children, d)
	return d
}

type FixrSetting int

const (
	DryRun  FixrSetting = 1 << 0
	Verbose             = 1 << 1
	Rename              = 1 << 2
)

type Fixr struct {
	Setting                FixrSetting
	rootStructure          *rootTree
	includeDirectiveConfig *IncludeDirectiveConfig
	includeGuardConfig     *IncludeGuardConfig
	renamedHeaderFiles     map[string]FileRename
}

func NewFixr(includeDirectiveConfig *IncludeDirectiveConfig, includeGuardConfig *IncludeGuardConfig) *Fixr {
	f := &Fixr{
		rootStructure: newRoot(),
	}

	f.includeDirectiveConfig = includeDirectiveConfig
	f.includeGuardConfig = includeGuardConfig
	f.renamedHeaderFiles = make(map[string]FileRename)

	return f
}

func (f *Fixr) Verbose() bool {
	return f.Setting&Verbose != 0
}
func (f *Fixr) DryRun() bool {
	return f.Setting&DryRun != 0
}
func (f *Fixr) Rename() bool {
	return f.Setting&Rename != 0
}

func (f *Fixr) matchAndFixIncludeDirective(lineNumber int, line string, _filepathOfFileBeingFixed string) (l string, modified bool) {
	matches := f.includeDirectiveConfig.IncludeDirectiveRegex.FindStringSubmatch(line)

	if matches == nil || len(matches) != 4 {
		return line, false
	}

	// We do not touch system includes, e.g. #include <iostream>
	if matches[1] == "\"" && matches[3] == "\"" {
		l, modified = f.correctIncludeDirective(lineNumber, line, matches[2], _filepathOfFileBeingFixed)
		return
	}

	return line, false
}

func (f *Fixr) matchAndFixIncludeGuard(lineNumber int, line string, nextline string, _filepathOfFileBeingFixed string) (l1 string, l2 string, status int) {
	cfg := f.includeGuardConfig

	_ifndef := cfg.IncludeGuardIfNotDefRegex.FindStringSubmatch(line)
	if _ifndef == nil || len(_ifndef) != 2 {
		return "", "", -1
	}

	_define := cfg.IncludeGuardDefineRegex.FindStringSubmatch(nextline)
	if _define == nil || len(_define) != 2 {
		return "", "", -1
	}

	status = 0
	if strings.Compare(_ifndef[1], _define[1]) != 0 {
		def := cfg.modifyDefine(_define[1], _filepathOfFileBeingFixed)
		l1 = "#ifndef " + def
		l2 = "#define " + def
		status = 1

		if f.Verbose() {
			fmt.Printf("fixer include guard in %s, line %d, %s -> %s\n", _filepathOfFileBeingFixed, lineNumber, line, l1)
		}
	}

	return l1, l2, status
}

// Scan is adding more directories and files into the full hierarchical structure of directories and files
func (f *Fixr) ProcessScanners(scanners *DirScanner) {
	for _, scanner := range scanners.scanners {
		err := f.rootStructure.scanDir(scanner.dir, scanner.filter)
		if err != nil {
			fmt.Println(err)
		}
	}
	f.rootStructure.finalize() // Finalize, will build matching database for fuzzy matching
}

// Rename will rename files in the directories that are passed in the Renamer slice
func (f *Fixr) ProcessRenamers(renamers *Renamers) {
	for _, renamer := range renamers.renamers {
		f.globAndRename(renamer.dir, renamer.renameFunc, renamer.isSourceFileFilter, renamer.isHeaderFileFilter)
	}
}

func (f *Fixr) ProcessFixers(fixers *Fixers) {
	for _, fxr := range fixers.fileFixers {

		sourceFilepaths := make([]string, 0, 512)
		headerFilepaths := make([]string, 0, 512)

		err := filepath.WalkDir(fxr.dir, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				path = filepath.Dir(path)
				relpath, _ := filepath.Rel(fxr.dir, path)
				_filepath := filepath.Join(relpath, d.Name())
				if fxr.filter2(_filepath) {
					if fxr.fileType == SourceFileType {
						sourceFilepaths = append(sourceFilepaths, _filepath)
					} else if fxr.fileType == HeaderFileType {
						headerFilepaths = append(headerFilepaths, _filepath)
					}
				}
			}
			return err
		})

		if err == nil {
			for _, _filepath := range sourceFilepaths {
				if err := f.fixSourceFile(fxr.dir, _filepath); err != nil {
					fmt.Println(err)
				}
			}

			for _, _filepath := range headerFilepaths {
				if err := f.fixHeaderFile(fxr.dir, _filepath); err != nil {
					fmt.Println(err)
				}
			}
		}
	}

}

func (f *Fixr) findBestMatchingHeaderFile(lineNumber int, includeDirective string, includeFilepath string, _filepathOfFileBeingFixed string) (string, bool) {
	found := f.rootStructure.files[strings.ToLower(includeFilepath)]
	if found != nil {
		return found.path, true
	}

	includeFilename := filepath.Base(includeFilepath)
	includeFilenameLowerCase := strings.ToLower(includeFilename)

	// Note, this can be optimized by building a map[string][]*ffile to map a filename to a
	// []*ffile that have the same filename but different path.
	files := map[string]*ffile{}
	for _, file := range f.rootStructure.files {
		if strings.EqualFold(file.name, includeFilenameLowerCase) {
			files[strings.ToLower(file.path)] = file
		}
	}

	if len(files) > 1 {
		// First check any of them if they have the same parent directory
		dirs := make(map[string]*fdir)
		for fk, fi := range files {
			dirs[fk] = fi.dir
		}
		includeFilePathIter := filepath.Dir(includeFilepath)
		for len(files) > 1 {
			pruneList := make([]string, 0, len(files))
			dn := filepath.Base(includeFilePathIter)
			for fk, fi := range files {
				fd := dirs[fk]
				if strings.Compare(dn, filepath.Base(fd.name)) != 0 {
					pruneList = append(pruneList, fi.path)
				}
			}
			for _, pi := range pruneList { // prune the map of files
				delete(files, pi)
				delete(dirs, pi)
			}

			// Go up one directory
			for dk, di := range dirs {
				dirs[dk] = di.parent
			}
			includeFilePathIter = ParentPath(includeFilePathIter)
		}

		for len(files) > 1 {
			possible := make([]string, len(files))
			for _, f := range files {
				possible = append(possible, f.path)
			}
			cm := NewClosestMatch(possible, []int{2, 3, 4}) // Find a fuzzy match for these files
			closest := cm.Closest(includeFilepath)
			_, cs := ClosestDistance(includeFilename, closest)
			if cs >= 0.9 {
				return closest, true
			}
		}
	}

	if len(files) == 1 {
		for _, fi := range files {
			return fi.path, true
		}
	}

	closest := f.rootStructure.fuzzyMatcher.Closest(includeFilename)
	_, cs := ClosestDistance(includeFilename, closest)
	if cs >= 0.9 {
		return closest, true
	}

	if f.Verbose() {
		fmt.Printf("fixer was unable to correct include directive in %s, line %d, %s\n", _filepathOfFileBeingFixed, lineNumber, includeDirective)
	}

	return includeFilepath, false
}

func (f *Fixr) correctIncludeDirective(lineNumber int, includeDirective string, includePart string, _filepathOfFileBeingFixed string) (string, bool) {

	if newIncludeFilePath, found := f.findBestMatchingHeaderFile(lineNumber, includeDirective, includePart, _filepathOfFileBeingFixed); found {
		newIncludeDirective := strings.Replace(includeDirective, includePart, newIncludeFilePath, -1)
		if strings.Compare(newIncludeDirective, includeDirective) != 0 {

			// TODO Handle the rename here ?
			//  -> f.renamedHeaderFiles
			if rename, found := f.renamedHeaderFiles[strings.ToLower(newIncludeDirective)]; found {
				newIncludeDirective = rename.NewFilePath
			}

			if f.Verbose() {
				fmt.Printf("fixer, include directive in %s, line %d, %s -> %s\n", _filepathOfFileBeingFixed, lineNumber, includeDirective, newIncludeDirective)
			}

			return newIncludeDirective, true
		}
	}
	return includeDirective, false
}

func skipCommentsAndEmptyLines(line int, lines []string) int {
	i := line
	for i < len(lines) {
		var line string

		skipped := false

		for i < len(lines) {
			line = lines[i]
			line = strings.TrimLeft(line, " \t")
			if line == "" {
				i++
				skipped = true
				continue
			}
			if strings.HasPrefix(line, "//") {
				i++
				skipped = true
				continue
			}
			if strings.HasPrefix(line, "/*") {
				for !strings.HasSuffix(line, "*/") {
					i++
					line = lines[i]
					line = strings.TrimLeft(line, " \t")
				}
				skipped = true
				continue
			}
			break
		}

		if !skipped {
			break
		}
	}

	return i
}

func (f *Fixr) fixHeaderFile(dirpath string, _filepath string) error {

	path := filepath.Join(dirpath, _filepath)
	lines, err := readLinesFromFile(path)
	if err != nil {
		return err
	}

	numCorrections := 0
	if f.includeDirectiveConfig != nil {
		for i, line := range lines {
			if l, modified := f.matchAndFixIncludeDirective(i+1, line, _filepath); modified {
				lines[i] = l
				numCorrections++
			}
		}
	}

	if f.includeGuardConfig != nil {
		i := 0

		// Determine if this header file is using include guards, some header files might
		// only have 'pragma once' and no include guards.
		i = skipCommentsAndEmptyLines(i, lines)
		onlyHasPragmaOnce := strings.HasPrefix(lines[i], "#pragma once")
		if !onlyHasPragmaOnce {
			i = 0
			for i < (len(lines) - 1) {
				l1, l2, status := f.matchAndFixIncludeGuard(i, lines[i], lines[i+1], _filepath)
				if status == 1 {
					lines[i] = l1
					lines[i+1] = l2
					numCorrections++
				}
				i += 1
			}
		} else {
			numCorrections += 1

			define := ""
			define = f.includeGuardConfig.modifyDefine(define, _filepath)

			if f.Verbose() {
				fmt.Printf("fixer, include guard set in %s, line %d as #define %s\n", _filepath, i, define)
			}

			lines[i] = "#ifndef " + define
			lineToWriteDefineGuard := i + 1

			// Last line of the file should be #endif
			lines = append(lines, "#endif")

			// Shift all lines down by one, first add one empty line to accommodate the shift
			lines = append(lines, "")
			for j := len(lines) - 1; j > i; j-- {
				lines[j] = lines[j-1]
			}

			lines[lineToWriteDefineGuard] = "#define " + define
		}
	}

	if numCorrections > 0 {
		if !f.DryRun() {
			if err = writeLinesToFile(path, lines); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *Fixr) fixSourceFile(dirpath string, _filepath string) error {

	path := filepath.Join(dirpath, _filepath)
	lines, err := readLinesFromFile(path)
	if err != nil {
		return err
	}

	numCorrections := 0
	if f.includeDirectiveConfig != nil {
		for i, line := range lines {
			if l, modified := f.matchAndFixIncludeDirective(i+1, line, _filepath); modified {
				lines[i] = l
				numCorrections++
			}
		}
	}

	if numCorrections > 0 {
		if !f.DryRun() {
			if err = writeLinesToFile(path, lines); err != nil {
				return err
			}
		}
	}
	return nil
}
