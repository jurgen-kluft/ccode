package axe

import (
	"path/filepath"
	"sort"
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
	FileTypeIxx
	FileTypeMxx
)

type FileEntryXcodeConfig struct {
	UUID      UUID
	BuildUUID UUID
}

// -----------------------------------------------------------------------------------------------
// FileEntry
// -----------------------------------------------------------------------------------------------

type FileEntry struct {
	Path              string
	Type              FileType
	GenDataXcode      FileEntryXcodeConfig
	Parent            *VirtualDirectory
	ExcludedFromBuild bool
	Generated         bool
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

	ext := PathFileExtension(fe.Path)
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
	case ".m", ".mm":
		fe.Type = FileTypeObjC
	case ".mxx":
		fe.Type = FileTypeMxx
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

func (f *FileEntry) Is_IXX() bool {
	return f.Type == FileTypeIxx
}

// -----------------------------------------------------------------------------------------------
// FileEntryDict
// -----------------------------------------------------------------------------------------------

type FileEntryDict struct {
	Workspace *Workspace     // The workspace this dict belongs to
	Path      string         // All FileEntry objects are relative to this path
	Dict      map[string]int // Maps path to index in Keys/Values
	Keys      []string       // Keys
	Values    []*FileEntry   // Values
}

func NewFileEntryDict(ws *Workspace, path string) *FileEntryDict {
	return &FileEntryDict{
		Workspace: ws,
		Path:      path,
		Dict:      make(map[string]int),
		Keys:      []string{},
		Values:    []*FileEntry{},
	}
}

func (d *FileEntryDict) GetRelativePath(e *FileEntry, path string) string {
	return PathGetRel(filepath.Join(d.Path, e.Path), path)
}

func (d *FileEntryDict) Add(path string, isGenerated bool) *FileEntry {
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
