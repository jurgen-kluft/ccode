package axe

type ProjectGroup struct {
	Path     string
	Children []*ProjectGroup
	Projects []*Project
	Parent   *ProjectGroup
	MsDev    struct {
		UUID UUID
	}
}

func NewProjectGroup() *ProjectGroup {
	return &ProjectGroup{
		Children: make([]*ProjectGroup, 0),
		Projects: make([]*Project, 0),
	}
}

type ProjectGroups struct {
	Dict   map[string]int
	Values []*ProjectGroup
	Root   *ProjectGroup
}

func NewProjectGroups(root *ProjectGroup) *ProjectGroups {
	return &ProjectGroups{
		Dict:   make(map[string]int),
		Values: make([]*ProjectGroup, 0),
		Root:   root,
	}
}

func (d *ProjectGroups) Add(p *Project) *ProjectGroup {
	group := d.GetOrAddGroup(p.Settings.Group)
	p.Group = group
	group.Projects = append(group.Projects, p)
	return group
}

func (d *ProjectGroups) GetOrAddParent(path string) *ProjectGroup {
	parent, sub := PathUp(path)
	if len(sub) == 0 {
		return d.Root
	}
	return d.GetOrAddGroup(parent)
}

func (d *ProjectGroups) GetOrAddGroup(group string) *ProjectGroup {
	if idx, ok := d.Dict[group]; ok {
		return d.Values[idx]
	}
	idx := len(d.Values)
	d.Dict[group] = idx
	d.Values = append(d.Values, NewProjectGroup())
	return d.Values[idx]
}
