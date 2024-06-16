package axe

import (
	"path/filepath"
	"sort"
)

type VirtualFolder struct {
	DiskPath      string
	Path          string
	Children      []*VirtualFolder
	Files         []*FileEntry
	Parent        *VirtualFolder
	GenData_xcode struct {
		UUID UUID
	}
}

type VirtualFolders struct {
	DiskPath string
	Map      map[string]int
	Folders  []*VirtualFolder
	Root     *VirtualFolder
}

func NewVirtualFolder() *VirtualFolder {
	c := &VirtualFolder{}
	c.Children = make([]*VirtualFolder, 0)
	c.Files = make([]*FileEntry, 0)
	c.GenData_xcode.UUID = GenerateUUID()
	return c
}

func (f *VirtualFolders) getOrAddParent(filePath string) *VirtualFolder {
	dir, _ := PathUp(filePath)
	if len(dir) == 0 || dir == "." {
		return f.Root
	}

	v, created := f.getOrCreateFolder(dir)
	if created {
		v.Path = dir
		v.DiskPath = filepath.Join(f.DiskPath, dir)
		p := f.getOrAddParent(dir)
		p.Children = append(p.Children, v)
		v.Parent = p
	}

	return v
}

func (c *VirtualFolder) SortByKey() {
	sort.Slice(c.Children, func(i, j int) bool {
		return c.Children[i].Path < c.Children[j].Path
	})
	sort.Slice(c.Files, func(i, j int) bool {
		return c.Files[i].Path < c.Files[j].Path
	})

	for _, child := range c.Children {
		child.SortByKey()
	}
}

func NewVirtualFolders(diskPath string) *VirtualFolders {
	v := &VirtualFolders{
		DiskPath: diskPath,
		Map:      make(map[string]int),
	}
	v.Root = NewVirtualFolder()
	v.Map["."] = len(v.Folders)
	v.Folders = []*VirtualFolder{v.Root}
	return v
}

func (f *VirtualFolders) getOrCreateFolder(key string) (*VirtualFolder, bool) {
	if v, ok := f.Map[key]; ok {
		return f.Folders[v], false
	}
	v := NewVirtualFolder()
	f.Map[key] = len(f.Folders)
	f.Folders = append(f.Folders, v)
	return v, true
}

func (f *VirtualFolders) AddFile(file *FileEntry) {
	p := f.getOrAddParent(file.Path)
	file.Parent = p
	p.Files = append(p.Files, file)
}

func (f *VirtualFolders) SortByKey() {

	if f.Root != nil {
		f.Root.SortByKey()
	}
}
