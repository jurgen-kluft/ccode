package denv

import (
	"strings"

	"github.com/jurgen-kluft/ccode/dev"
)

// ----------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------

type PinnedPathSet struct {
	Entries map[string]int
	Values  []dev.PinnedPath
}

func NewPinnedPathSet() *PinnedPathSet {
	d := &PinnedPathSet{}
	d.Entries = make(map[string]int)
	d.Values = make([]dev.PinnedPath, 0)
	return d
}

func (d *PinnedPathSet) Merge(other *PinnedPathSet) {
	for _, value := range other.Values {
		d.AddOrSet(value)
	}
}

func (d *PinnedPathSet) Copy() *PinnedPathSet {
	c := NewPinnedPathSet()
	c.Merge(d)
	return c
}

func (d *PinnedPathSet) Extend(rhs *PinnedPathSet) {
	for _, fp := range rhs.Values {
		d.AddOrSet(fp)
	}
}

func (d *PinnedPathSet) UniqueExtend(rhs *PinnedPathSet) {
	for _, fp := range rhs.Values {
		fullpath := fp.String()
		if _, ok := d.Entries[fullpath]; !ok {
			d.AddOrSet(fp)
		}
	}
}

func (d *PinnedPathSet) AddOrSet(fp dev.PinnedPath) {
	fullpath := fp.String()
	i, ok := d.Entries[fullpath]
	if !ok {
		d.Entries[fullpath] = len(d.Values)
		d.Values = append(d.Values, fp)
	} else if strings.Compare(d.Values[i].String(), fullpath) != 0 {
		d.Values[i] = fp
	}
}

// Enumerate will call the enumerator function for each key-value pair in the dictionary.
//
//	'last' will be 0 for all but the last key-value pair, and 1 for the last key-value pair.
func (d *PinnedPathSet) Enumerate(enumerator func(i int, root string, base string, dir string, last int)) {
	n := (len(d.Values) - 1)
	for i, fp := range d.Values {
		root := fp.Root
		base := fp.Base
		dir := fp.Sub
		last := 0
		if i == n {
			last = 1
		}
		enumerator(i, root, base, dir, last)
	}
}

func (d *PinnedPathSet) Concatenated(prefix string, suffix string, modifier func(root, base, sub string) string) string {
	concat := ""
	for _, fp := range d.Values {
		newFullPath := modifier(fp.Root, fp.Base, fp.Sub)
		concat += prefix + newFullPath + suffix
	}
	return concat
}
