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
	Folders  map[string]*VirtualFolder
	Root     *VirtualFolder
}

func NewVirtualFolder() *VirtualFolder {
	c := &VirtualFolder{}
	c.Children = make([]*VirtualFolder, 0)
	c.Files = make([]*FileEntry, 0)
	c.GenData_xcode.UUID = GenerateUUID()
	return c
}

func (f *VirtualFolders) GetOrAddParent(filePath string) *VirtualFolder {
	dir, _ := PathUp(filePath)
	if len(dir) == 0 || dir == "." {
		return f.Root
	}

	v := f.Folders[dir]
	if v == nil {
		v = NewVirtualFolder()
		v.Path = dir
		v.DiskPath = filepath.Join(f.DiskPath, dir)

		p := f.GetOrAddParent(dir)
		p.Children = append(p.Children, v)
		v.Parent = p
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

func NewVirtualFolders(diskPath string) *VirtualFolders {
	v := &VirtualFolders{
		DiskPath: diskPath,
		Folders:  make(map[string]*VirtualFolder),
	}
	v.Root = NewVirtualFolder()
	v.Folders["."] = v.Root
	return v
}

func (f *VirtualFolders) AddFile(file *FileEntry) {
	p := f.GetOrAddParent(file.Path)
	file.Parent = p
	p.Files = append(p.Files, file)
}
