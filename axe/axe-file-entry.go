package axe

import "sort"

// enum class FileType {
// 	None,
// 	cpp_header, // c or cpp header
// 	cpp_source,
// 	c_source,
// 	cu_header,	// cuda header
// 	cu_source,	// cuda source
// 	ixx, // cpp modules
// 	mxx, // cpp module implementation
// };

type FileType int

const (
	FileTypeNone FileType = iota
	FileTypeCppHeader
	FileTypeCppSource
	FileTypeCSource
	FileTypeCuHeader
	FileTypeCuSource
	FileTypeIxx
	FileTypeMxx
)

type FileEntryXcodeConfig struct {
	UUID      UUID
	BuildUUID UUID
}

type FileEntry struct {
	Path              string
	Type              FileType
	GenDataXcode      FileEntryXcodeConfig
	Parent            *VirtualFolder
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

	ext := PathExtension(fe.Path)
	switch ext {
	case "h", "hpp":
		fe.Type = FileTypeCppHeader
		fe.ExcludedFromBuild = true
	case "cpp", "cc", "cxx":
		fe.Type = FileTypeCppSource
	case "c":
		fe.Type = FileTypeCSource
	case "cuh":
		fe.Type = FileTypeCuHeader
	case "cu":
		fe.Type = FileTypeCuSource
	case "ixx":
		fe.Type = FileTypeIxx
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

func (f *FileEntry) Is_IXX() bool {
	return f.Type == FileTypeIxx
}

type FileEntryDict struct {
	Workspace *Workspace
	Dict      map[string]int
	Keys      []string
	Values    []*FileEntry
}

func NewFileEntryDict(ws *Workspace) *FileEntryDict {
	return &FileEntryDict{
		Workspace: ws,
		Dict:      make(map[string]int),
		Keys:      []string{},
		Values:    []*FileEntry{},
	}
}

func (d *FileEntryDict) Add(path string, isGenerated bool) *FileEntry {
	key := path
	if e, ok := d.Dict[key]; ok {
		return d.Values[e]
	}

	e := NewFileEntryInit(key, isGenerated)
	d.Dict[key] = len(d.Values) - 1
	d.Keys = append(d.Keys, key)
	d.Values = append(d.Values, e)
	return e
}

func (d *FileEntryDict) SortByKey() {
	sort.Strings(d.Keys)
	newList := []*FileEntry{}
	for _, k := range d.Keys {
		newList = append(newList, d.Values[d.Dict[k]])
	}
	d.Values = newList
}

// Custom sort for []FileEntry
type EntrySort struct {
	d *FileEntryDict
	i []int
}

func (s EntrySort) Len() int {
	return len(s.i)
}

func (s EntrySort) Swap(i, j int) {
	s.i[i], s.i[j] = s.i[j], s.i[i]
}

func (s EntrySort) Less(i, j int) bool {
	return s.d.Values[s.i[i]].Path < s.d.Values[s.i[j]].Path
}

func (d *FileEntryDict) SortByEntry() {

	// Create a Values of indexes
	indexes := make([]int, len(d.Values))
	for i := range indexes {
		indexes[i] = i
	}

	// Sort the Values through the custom sort
	sort.Sort(EntrySort{d, indexes})

	// Create a new Values of Entries
	newList := make([]*FileEntry, len(d.Values))
	for i, v := range indexes {
		newList[i] = d.Values[v]
	}

	// Replace the old Values with the new one
	d.Values = newList
}
