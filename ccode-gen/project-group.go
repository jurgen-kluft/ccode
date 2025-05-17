package ccode_gen

import (
	ccode_utils "github.com/jurgen-kluft/ccode/ccode-utils"
)

type ProjectGroup struct {
	Path     string
	Children []*ProjectGroup
	Projects []*Project
	Parent   *ProjectGroup
	MsDev    struct {
		UUID ccode_utils.UUID
	}
}

func NewProjectGroup(path string) *ProjectGroup {
	return &ProjectGroup{
		Path:     path,
		Children: make([]*ProjectGroup, 0),
		Projects: make([]*Project, 0),
	}
}

type ProjectGroups struct {
	Dict   map[string]int
	Values []*ProjectGroup
	Root   *ProjectGroup
}

func NewProjectGroups() *ProjectGroups {
	g := &ProjectGroups{
		Dict:   make(map[string]int),
		Values: make([]*ProjectGroup, 0),
		Root:   nil,
	}

	g.Root = NewProjectGroup("<group_root>")
	g.Dict["."] = 0
	g.Values = append(g.Values, g.Root)

	return g
}

func (d *ProjectGroups) Add(p *Project) *ProjectGroup {
	group := d.GetOrAddGroup(p.Settings.Group)
	p.Group = group
	group.Projects = append(group.Projects, p)
	return group
}

func (d *ProjectGroups) GetOrAddParent(path string) *ProjectGroup {
	parent, _ := ccode_utils.PathUp(path)
	if len(parent) == 0 || parent == "." {
		return d.Root
	}
	return d.GetOrAddGroup(parent)
}

func (d *ProjectGroups) GetOrAddGroup(path string) *ProjectGroup {
	if idx, ok := d.Dict[path]; ok {
		return d.Values[idx]
	}

	v := NewProjectGroup(path)

	idx := len(d.Values)
	d.Dict[path] = idx
	d.Values = append(d.Values, v)

	p := d.GetOrAddParent(path)
	v.Parent = p
	p.Children = append(p.Children, v)

	return v
}
