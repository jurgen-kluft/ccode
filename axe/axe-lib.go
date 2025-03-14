package axe

import (
	"github.com/jurgen-kluft/ccode/denv"
)

// -----------------------------------------------------------------------------------------------------
//
// -----------------------------------------------------------------------------------------------------

type Library struct {
	Frameworks *ValueSet // MacOS specific
	Files      *ValueSet
	Libs       *ValueSet
	Dirs       *PinnedPathSet
}

func NewLibrary() *Library {
	l := &Library{}
	l.Frameworks = NewValueSet()
	l.Files = NewValueSet()
	l.Libs = NewValueSet()
	l.Dirs = NewPinnedPathSet()
	return l
}

func (l *Library) Merge(other *Library) {
	l.Frameworks.Merge(other.Frameworks)
	l.Files.Merge(other.Files)
	l.Libs.Merge(other.Libs)
	l.Dirs.Merge(other.Dirs)
}

func (l *Library) Copy() *Library {
	nl := &Library{}
	nl.Frameworks = l.Frameworks.Copy()
	nl.Files = l.Files.Copy()
	nl.Libs = l.Libs.Copy()
	nl.Dirs = l.Dirs.Copy()
	return nl
}

func (l *Library) Add(projectDirectory string, lib *denv.Lib) {
	if lib.Type == denv.Framework {
		l.Frameworks.AddMany(lib.Files...)
		l.Frameworks.AddMany(lib.Libs...)
	} else {
		l.Files.AddMany(lib.Files...)
		l.Libs.AddMany(lib.Libs...)
		l.Dirs.AddOrSet(projectDirectory, lib.Dir)
	}
}
