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
	AbsPath           string
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

func NewFileEntryInit(absPath string, isAbs bool, isGenerated bool, ws *Workspace) *FileEntry {
	fe := &FileEntry{Parent: nil, ExcludedFromBuild: true, Generated: false}
	fe.Init(absPath, isAbs, isGenerated, ws)
	return fe
}

func (fe *FileEntry) Init(absPath string, isAbs bool, isGenerated bool, ws *Workspace) {

	if isAbs {
		fe.Path = absPath
	} else {
		fe.Path = PathGetRel(absPath, ws.BuildDir)
	}

	ext := PathExtension(fe.Path)

	fe.ExcludedFromBuild = false
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
	List      []*FileEntry
}

func NewFileEntryDict(ws *Workspace) *FileEntryDict {
	return &FileEntryDict{
		Workspace: ws,
		Dict:      make(map[string]int),
		Keys:      []string{},
		List:      []*FileEntry{},
	}
}

func (d *FileEntryDict) Add(path, fromDir string, isGenerated bool) *FileEntry {
	key := path
	isAbs := true

	if fromDir != "" {
		isAbs = PathIsAbs(path)
		if isAbs {
			key = path
		} else {
			key = PathMakeFullPath(fromDir, path)
		}
	}

	if e, ok := d.Dict[key]; ok {
		return d.List[e]
	}

	e := NewFileEntryInit(key, isAbs, isGenerated, d.Workspace)
	d.Dict[key] = len(d.List) - 1
	d.Keys = append(d.Keys, key)
	d.List = append(d.List, e)
	return e
}

func (d *FileEntryDict) SortByKey() {
	sort.Strings(d.Keys)
	newList := []*FileEntry{}
	for _, k := range d.Keys {
		newList = append(newList, d.List[d.Dict[k]])
	}
	d.List = newList
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
	return s.d.List[s.i[i]].Path < s.d.List[s.i[j]].Path
}

func (d *FileEntryDict) SortByEntry() {

	// Create a List of indexes
	indexes := make([]int, len(d.List))
	for i := range indexes {
		indexes[i] = i
	}

	// Sort the List through the custom sort
	sort.Sort(EntrySort{d, indexes})

	// Create a new List of Entries
	newList := make([]*FileEntry, len(d.List))
	for i, v := range indexes {
		newList[i] = d.List[v]
	}

	// Replace the old List with the new one
	d.List = newList
}
