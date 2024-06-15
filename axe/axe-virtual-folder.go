package axe

import (
	"path/filepath"
	"sort"
)

type VirtualFolder struct {
	Path          string
	DiskPath      string
	Children      []*VirtualFolder
	Files         []*FileEntry
	Parent        *VirtualFolder
	GenData_xcode struct {
		UUID UUID
	}
}

func NewVirtualFolder() *VirtualFolder {
	c := &VirtualFolder{}
	c.Children = make([]*VirtualFolder, 0)
	c.Files = make([]*FileEntry, 0)
	c.GenData_xcode.UUID = GenerateUUID()
	return c
}

func (f *VirtualFolders) GetOrAddParent(baseDir, path string) *VirtualFolder {
	dir, name := filepath.Split(filepath.Clean(path))
	if len(name) == 0 {
		return f.Root
	}

	v := f.Folders[dir]
	if v == nil {
		v = NewVirtualFolder()
		v.Path = dir
		v.DiskPath = filepath.Join(baseDir, dir)

		p := f.GetOrAddParent(baseDir, dir)
		v.Parent = p
		p.Children = append(p.Children, v)

		f.Folders[dir] = v
	}

	return v
}

func (c *VirtualFolder) Sort() {
	sort.Slice(c.Children, func(i, j int) bool {
		return c.Children[i].Path < c.Children[j].Path
	})
	sort.Slice(c.Files, func(i, j int) bool {
		return c.Files[i].Path < c.Files[j].Path
	})

	for _, child := range c.Children {
		child.Sort()
	}
}

type VirtualFolders struct {
	Folders map[string]*VirtualFolder
	Root    *VirtualFolder
}

func NewVirtualFolders() *VirtualFolders {
	return &VirtualFolders{
		Folders: make(map[string]*VirtualFolder),
	}
}

// void VirtualFolderDict::add(const StrView& baseDir, FileEntry& file) {
// 	String rel;
// 	Path::getRel(rel, file.absPath(), baseDir);

// 	auto* p = getOrAddParent(baseDir, rel);
// 	file.parent = p;
// 	p->files.append(&file);
// }

func (f *VirtualFolders) AddFile(baseDir string, file *FileEntry) {
	rel := file.AbsPath
	p := f.GetOrAddParent(baseDir, rel)
	file.Parent = p
	p.Files = append(p.Files, file)
}
