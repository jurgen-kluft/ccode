package axe

import (
	"path/filepath"
)

// ----------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------

type ValueSet struct {
	Entries map[string]int
	Values  []string
}

func (d *ValueSet) Merge(other *ValueSet) {
	for _, value := range other.Values {
		d.Add(value)
	}
}

func NewValueSet() *ValueSet {
	d := &ValueSet{}
	d.Entries = make(map[string]int)
	d.Values = make([]string, 0)
	return d
}

func (d *ValueSet) Extend(rhs *ValueSet) {
	for _, value := range rhs.Values {
		d.Add(value)
	}
}

func (d *ValueSet) UniqueExtend(rhs *ValueSet) {
	for _, value := range rhs.Values {
		if _, ok := d.Entries[value]; !ok {
			d.Add(value)
		}
	}
}

func (d *ValueSet) Add(value string) {
	i, ok := d.Entries[value]
	if !ok {
		d.Entries[value] = len(d.Values)
		d.Values = append(d.Values, value)
	} else if d.Values[i] != value {
		d.Values[i] = value
	}
}

// Enumerate will call the enumerator function for each key-value pair in the dictionary.
//
//	'last' will be 0 for all but the last key-value pair, and 1 for the last key-value pair.
func (d *ValueSet) Enumerate(enumerator func(i int, key string, value string, last int)) {
	for i, key := range d.Values {
		if i == len(d.Values)-1 {
			enumerator(i, key, d.Values[i], 1)
		} else {
			enumerator(i, key, d.Values[i], 0)
		}
	}
}

func (d *ValueSet) Concatenated(prefix string, suffix string, valueModifier func(string) string) string {
	concat := ""
	for _, value := range d.Values {
		concat += prefix + valueModifier(value) + suffix
	}
	return concat
}

// ----------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------

type KeyValueDict struct {
	Entries map[string]int
	Keys    []string
	Values  []string
}

func (d *KeyValueDict) Merge(other *KeyValueDict) {
	for i, value := range other.Values {
		d.AddOrSet(other.Keys[i], value)
	}
}

func NewKeyValueDict() *KeyValueDict {
	d := &KeyValueDict{}
	d.Entries = make(map[string]int)
	d.Keys = make([]string, 0)
	d.Values = make([]string, 0)
	return d
}

func (d *KeyValueDict) Extend(rhs *KeyValueDict) {
	for i, value := range rhs.Values {
		d.AddOrSet(rhs.Keys[i], value)
	}
}

func (d *KeyValueDict) UniqueExtend(rhs *KeyValueDict) {
	for i, key := range rhs.Keys {
		if _, ok := d.Entries[key]; !ok {
			d.AddOrSet(key, rhs.Values[i])
		}
	}
}

func (d *KeyValueDict) AddOrSet(key string, value string) {
	i, ok := d.Entries[key]
	if !ok {
		d.Entries[key] = len(d.Values)
		d.Keys = append(d.Keys, key)
		d.Values = append(d.Values, value)
	} else if d.Values[i] != value {
		d.Values[i] = value
	}
}

// Enumerate will call the enumerator function for each key-value pair in the dictionary.
//
//	'last' will be 0 for all but the last key-value pair, and 1 for the last key-value pair.
func (d *KeyValueDict) Enumerate(enumerator func(i int, key string, value string, last int)) {
	for i, key := range d.Keys {
		if i == len(d.Keys)-1 {
			enumerator(i, key, d.Values[i], 1)
		} else {
			enumerator(i, key, d.Values[i], 0)
		}
	}
}

func (d *KeyValueDict) Concatenated(prefix string, suffix string, valueModifier func(string, string) string) string {
	concat := ""
	for i, value := range d.Values {
		concat += prefix + valueModifier(d.Keys[i], value) + suffix
	}
	return concat
}

// ----------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------

type VarSettings struct {
	Name            string
	InheritDict     *KeyValueDict
	AddDict         *KeyValueDict
	RemoveDict      *KeyValueDict
	LocalAddDict    *KeyValueDict
	LocalRemoveDict *KeyValueDict
	FinalDict       *KeyValueDict
}

func NewVarDict(name string) *VarSettings {
	s := &VarSettings{}
	s.Name = name
	s.InheritDict = NewKeyValueDict()
	s.AddDict = NewKeyValueDict()
	s.RemoveDict = NewKeyValueDict()
	s.LocalAddDict = NewKeyValueDict()
	s.LocalRemoveDict = NewKeyValueDict()
	s.FinalDict = NewKeyValueDict()
	return s
}

func (s *VarSettings) ValuesToAdd(values ...string) {
	for _, value := range values {
		s.AddDict.AddOrSet(value, value)
	}
}

func (s *VarSettings) inherit(rhs *VarSettings) {
	s.InheritDict.Extend(rhs.InheritDict)
}

func (s *VarSettings) computeFinal() {
	s.InheritDict.Extend(s.AddDict)
	s.FinalDict.Extend(s.InheritDict)
	//s.FinalDict.Extend(s.LocalAddDict)
}

// ----------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------

/*
NOTE At the moment, remove, local_add, and local_remove are not used. The purpose of these dictionaries is to
provide a means to specify values that are not going to be inherited by other dictionaries.

How 'inheritance' works:
- The 'inherit' dictionary is the dictionary that is inherited from the parent dictionary.
- The 'add' dictionary is the dictionary where the user can add values to the dictionary.
- The 'remove' dictionary is the dictionary where the user can specify values to remove from the dictionary.
- The 'local_add' dictionary is the dictionary where the user can add values that are not going to be inherited.
- The 'local_remove' dictionary is the dictionary where the user can specify values to remove but that are not going to be inherited.
- The 'final' dictionary is the dictionary that is computed from the 'inherit', 'add', 'local_add' dictionaries,
  and the 'remove' and 'local_remove' dictionaries are applied to the 'final' dictionary.

*/

type PathSettings struct {
	Name            string
	Root            string
	InheritDict     *KeyValueDict
	AddDict         *KeyValueDict
	RemoveDict      *KeyValueDict
	LocalAddDict    *KeyValueDict
	LocalRemoveDict *KeyValueDict
	FinalDict       *KeyValueDict
}

func NewPathDict(name string, root string) *PathSettings {
	s := &PathSettings{}
	s.Name = name
	s.Root = root
	s.InheritDict = NewKeyValueDict()
	s.AddDict = NewKeyValueDict()
	s.RemoveDict = NewKeyValueDict()
	s.LocalAddDict = NewKeyValueDict()
	s.LocalRemoveDict = NewKeyValueDict()
	s.FinalDict = NewKeyValueDict()
	return s
}

func (s *PathSettings) ValuesToAdd(values ...string) {
	for _, value := range values {
		s.AddDict.AddOrSet(filepath.Join(s.Root, value), value)
	}
}

func (s *PathSettings) inherit(rhs *PathSettings) {
	s.InheritDict.Extend(rhs.InheritDict)
}

func (s *PathSettings) computeFinal() {
	s.InheritDict.Extend(s.AddDict)
	s.FinalDict.Extend(s.InheritDict)
	//s.FinalDict.Extend(s.LocalAddDict)
}
