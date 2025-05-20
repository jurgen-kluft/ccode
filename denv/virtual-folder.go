package denv

import (
	"path/filepath"
	"sort"

	cutils "github.com/jurgen-kluft/ccode/cutils"
)

type VirtualDirectory struct {
	DiskPath string
	Path     string
	Children []*VirtualDirectory
	Files    []*FileEntry
	Parent   *VirtualDirectory
	UUID     cutils.UUID
}

type VirtualDirectories struct {
	DiskPath string
	Map      map[string]int
	Folders  []*VirtualDirectory
	Root     *VirtualDirectory
}

func NewVirtualFolder() *VirtualDirectory {
	c := &VirtualDirectory{}
	c.Children = make([]*VirtualDirectory, 0)
	c.Files = make([]*FileEntry, 0)
	c.UUID = cutils.GenerateUUID()
	return c
}

func (f *VirtualDirectories) GetAllLeafDirectories() []*VirtualDirectory {
	var result []*VirtualDirectory
	for _, v := range f.Folders {
		if len(v.Children) == 0 && len(v.Files) > 0 {
			result = append(result, v)
		}
	}

	return result
}

func (f *VirtualDirectories) getOrAddParent(filePath string) *VirtualDirectory {
	dir, _ := cutils.PathUp(filePath)
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

func (c *VirtualDirectory) SortByKey() {
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

func NewVirtualFolders(diskPath string) *VirtualDirectories {
	v := &VirtualDirectories{
		DiskPath: diskPath,
		Map:      make(map[string]int),
	}
	v.Root = NewVirtualFolder()
	v.Map["."] = len(v.Folders)
	v.Folders = []*VirtualDirectory{v.Root}
	return v
}

func (f *VirtualDirectories) getOrCreateFolder(key string) (*VirtualDirectory, bool) {
	if v, ok := f.Map[key]; ok {
		return f.Folders[v], false
	}
	v := NewVirtualFolder()
	f.Map[key] = len(f.Folders)
	f.Folders = append(f.Folders, v)
	return v, true
}

func (f *VirtualDirectories) AddFile(file *FileEntry) {
	p := f.getOrAddParent(file.Path)
	file.Parent = p
	p.Files = append(p.Files, file)
}

func (f *VirtualDirectories) SortByKey() {

	if f.Root != nil {
		f.Root.SortByKey()
	}
}
