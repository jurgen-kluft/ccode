package axe

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

func (d *ValueSet) Copy() *ValueSet {
	c := NewValueSet()
	c.Merge(d)
	return c
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

func (d *ValueSet) AddMany(values ...string) {
	for _, value := range values {
		d.Add(value)
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

func (d *KeyValueDict) Copy() *KeyValueDict {
	c := NewKeyValueDict()
	c.Merge(d)
	return c
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
	Name string
	Vars *KeyValueDict
}

func NewVarDict(name string) *VarSettings {
	s := &VarSettings{}
	s.Name = name
	s.Vars = NewKeyValueDict()
	return s
}

func (s *VarSettings) Merge(other *VarSettings) {
	s.Vars.Merge(other.Vars)
}

func (s *VarSettings) Copy() *VarSettings {
	c := NewVarDict(s.Name)
	c.Vars.Merge(s.Vars)
	return c
}

func (s *VarSettings) ValuesToAdd(values ...string) {
	for _, value := range values {
		s.Vars.AddOrSet(value, value)
	}
}

func (s *VarSettings) AddOrSet(key string, value string) {
	s.Vars.AddOrSet(key, value)
}
