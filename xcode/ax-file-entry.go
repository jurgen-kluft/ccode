package xcode

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

// class FileEntry {
// public:
// 	void init(const StrView& absPath, bool isAbs, bool isGenerated);
// 	bool excludedFromBuild {true};

// 	const String&	absPath		() const { return _absPath; }
// 	const String&	path		() const { return _path; }

// 	FileType		type		() const { return _type; }
// 	bool			type_is_c	() const { return _type == FileType::c_source; }
// 	bool			type_is_cpp	() const { return _type == FileType::cpp_source; }
// 	bool			type_is_c_or_cpp() const { return type_is_c() || type_is_cpp(); }
// 	bool			type_is_ixx	() const { return _type == FileType::ixx; }

// 	struct GenData_xcode {
// 		String		uuid;
// 		String		buildUuid;
// 	};
// 	GenData_xcode genData_xcode;

// 	VirtualFolder* parent {nullptr};
// 	bool	generated = {false};

// 	explicit operator bool() const { return _absPath.operator bool(); }

// private:
// 	String		_absPath;
// 	String		_path; // absPath or relative path depends on user input, and relative path always relative to _build_dir
// 	FileType	_type {FileType::None};
// };

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
	Dict      map[string]*FileEntry
}

func NewFileEntryDict(ws *Workspace) *FileEntryDict {
	return &FileEntryDict{
		Workspace: ws,
		Dict:      make(map[string]*FileEntry),
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
		return e
	}

	e := NewFileEntryInit(key, isAbs, isGenerated, d.Workspace)
	d.Dict[key] = e
	return e
}
