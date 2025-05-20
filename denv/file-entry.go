package denv

import (
	"path/filepath"
	"sort"

	cutils "github.com/jurgen-kluft/ccode/cutils"
)

type FileType int

const (
	FileTypeNone FileType = iota
	FileTypeCppHeader
	FileTypeCppSource
	FileTypeCSource
	FileTypeCuHeader
	FileTypeCuSource
	FileTypeObjC
	FileTypeObjCpp
	FileTypeIxx
	FileTypeMxx
	FileTypeStaticLib
	FileTypeSharedLib
)

// -----------------------------------------------------------------------------------------------
// FileEntry
// -----------------------------------------------------------------------------------------------

type FileEntry struct {
	Path              string            // Relative to the workspace
	Type              FileType          // File type
	Parent            *VirtualDirectory // Parent directory
	ExcludedFromBuild bool              // Excluded from build
	Generated         bool              // Generated file
	UUID              cutils.UUID
	BuildUUID         cutils.UUID
}

func NewFileEntry() *FileEntry {
	return &FileEntry{Parent: nil, ExcludedFromBuild: true, Generated: false}
}

func NewFileEntryInit(path string, isGenerated bool) *FileEntry {
	fe := &FileEntry{Parent: nil, ExcludedFromBuild: true, Generated: false}
	fe.Init(path, isGenerated)
	return fe
}

func (fe *FileEntry) Init(path string, isGenerated bool) {

	fe.Path = path
	fe.ExcludedFromBuild = false

	ext := cutils.PathFileExtension(fe.Path)
	switch ext {
	case ".h", ".hpp", ".inl":
		fe.Type = FileTypeCppHeader
		fe.ExcludedFromBuild = true
	case ".cpp", ".cc", ".cxx":
		fe.Type = FileTypeCppSource
	case ".c":
		fe.Type = FileTypeCSource
	case ".cuh":
		fe.Type = FileTypeCuHeader
	case ".cu":
		fe.Type = FileTypeCuSource
	case ".ixx":
		fe.Type = FileTypeIxx
	case ".m":
		fe.Type = FileTypeObjC
	case ".mm":
		fe.Type = FileTypeObjCpp
	case ".mxx":
		fe.Type = FileTypeMxx
	case ".lib":
		fe.Type = FileTypeStaticLib
	case ".a":
		fe.Type = FileTypeStaticLib
	case ".dll":
		fe.Type = FileTypeSharedLib
	case ".so":
		fe.Type = FileTypeSharedLib
	default:
		fe.Type = FileTypeNone
	}
}

func (f *FileEntry) Is_C() bool {
	return f.Type == FileTypeCSource
}

func (f *FileEntry) Is_CPP() bool {
	return f.Type == FileTypeCppSource
}

func (f *FileEntry) Is_C_or_CPP() bool {
	return f.Is_C() || f.Is_CPP()
}

func (f *FileEntry) Is_ObjC() bool {
	return f.Type == FileTypeObjC
}

func (f *FileEntry) Is_SourceFile() bool {
	return f.Type == FileTypeCppSource || f.Type == FileTypeCSource || f.Type == FileTypeCuSource || f.Type == FileTypeObjC || f.Type == FileTypeObjCpp
}

func (f *FileEntry) Is_ObjCpp() bool {
	return f.Type == FileTypeObjCpp
}

func (f *FileEntry) Is_IXX() bool {
	return f.Type == FileTypeIxx
}

func (f *FileEntry) Is_StaticLib() bool {
	return f.Type == FileTypeStaticLib
}

func (f *FileEntry) Is_SharedLib() bool {
	return f.Type == FileTypeSharedLib
}

// -----------------------------------------------------------------------------------------------
// FileEntryDict
// -----------------------------------------------------------------------------------------------

type FileEntryDict struct {
	Path   string         // All FileEntry objects are relative to this path
	Dict   map[string]int // Maps path to index in Keys/Values
	Keys   []string       // Keys
	Values []*FileEntry   // Values
}

func NewFileEntryDict(path string) *FileEntryDict {
	return &FileEntryDict{
		Path:   path,
		Dict:   make(map[string]int),
		Keys:   []string{},
		Values: []*FileEntry{},
	}
}

func (d *FileEntryDict) GetAbsPath(e *FileEntry) string {
	if len(d.Path) == 0 {
		return e.Path
	}
	return filepath.Join(d.Path, e.Path)
}

func (d *FileEntryDict) GetRelativePath(e *FileEntry, path string) string {
	if len(d.Path) == 0 {
		return e.Path
	}
	return cutils.PathGetRelativeTo(filepath.Join(d.Path, e.Path), path)
}

func (d *FileEntryDict) add(path string, isGenerated bool) *FileEntry {
	key := path
	if e, ok := d.Dict[key]; ok {
		return d.Values[e]
	}

	e := NewFileEntryInit(key, isGenerated)
	d.Dict[key] = len(d.Values)
	d.Keys = append(d.Keys, key)
	d.Values = append(d.Values, e)
	return e
}

func (d *FileEntryDict) Add(path string) *FileEntry {
	return d.add(path, false)
}

func (d *FileEntryDict) AddGenerated(path string) *FileEntry {
	return d.add(path, true)
}

func (d *FileEntryDict) SortByKey() {
	sort.Strings(d.Keys)
	sortedValues := make([]*FileEntry, 0, len(d.Values))
	for _, k := range d.Keys {
		sortedValues = append(sortedValues, d.Values[d.Dict[k]])
	}
	d.Values = sortedValues
	for i, k := range d.Keys {
		d.Dict[k] = i
	}
}

func (d *FileEntryDict) Enumerate(includeInEnumeration func(value *FileEntry) bool, enumerator func(i int, key string, value *FileEntry, last int)) {

	// first calculate the number of items that are included in the enumeration
	count := 0
	for _, value := range d.Values {
		if includeInEnumeration(value) {
			count++
		}
	}

	// then enumerate them
	for i, value := range d.Values {
		if !includeInEnumeration(value) {
			continue
		}

		count -= 1
		if count == 0 {
			enumerator(i, d.Keys[i], value, 1)
		} else {
			enumerator(i, d.Keys[i], value, 0)
		}

	}
}
