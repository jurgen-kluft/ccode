package xcode

// class ProjectGroup {
// public:
// 	String path;
// 	Vector<ProjectGroup*> children;
// 	Vector<Project*> projects;
// 	ProjectGroup* parent {nullptr};

// 	struct GenData_vs2015 {
// 		String		uuid;
// 	};
// 	GenData_vs2015 genData_vs2015;
// };

type ProjectGroup struct {
	Path          string
	Children      []*ProjectGroup
	Projects      []*Project
	Parent        *ProjectGroup
	GenDataVs2015 struct {
		Uuid string
	}
}

func NewProjectGroup() *ProjectGroup {
	return &ProjectGroup{
		Children: make([]*ProjectGroup, 0),
		Projects: make([]*Project, 0),
	}
}

type ProjectGroupDict struct {
	Groups   map[string]int
	Projects []*ProjectGroup
	Root     *ProjectGroup
}

func NewProjectGroupDict(root *ProjectGroup) *ProjectGroupDict {
	return &ProjectGroupDict{
		Groups:   make(map[string]int),
		Projects: make([]*ProjectGroup, 0),
		Root:     root,
	}
}

func (d *ProjectGroupDict) GetOrAddParent(path string) *ProjectGroup {
	parent, sub := PathUp(path)
	if len(sub) == 0 {
		return d.Root
	}
	return d.GetOrAddGroup(parent)
}

func (d *ProjectGroupDict) GetOrAddGroup(group string) *ProjectGroup {
	if idx, ok := d.Groups[group]; ok {
		return d.Projects[idx]
	}
	idx := len(d.Projects)
	d.Groups[group] = idx
	d.Projects = append(d.Projects, NewProjectGroup())
	return d.Projects[idx]
}
