package dev

// ----------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------

type KeyValueSet struct {
	Entries map[string]int
	Keys    []string
	Values  []string
}

func (d *KeyValueSet) Merge(other *KeyValueSet) {
	for i, value := range other.Values {
		d.AddOrSet(other.Keys[i], value)
	}
}

func NewKeyValueSet() *KeyValueSet {
	d := &KeyValueSet{}
	d.Entries = make(map[string]int)
	d.Keys = make([]string, 0)
	d.Values = make([]string, 0)
	return d
}

func (d *KeyValueSet) Copy() *KeyValueSet {
	c := NewKeyValueSet()
	c.Merge(d)
	return c
}

func (d *KeyValueSet) Extend(rhs *KeyValueSet) {
	for i, value := range rhs.Values {
		d.AddOrSet(rhs.Keys[i], value)
	}
}

func (d *KeyValueSet) UniqueExtend(rhs *KeyValueSet) {
	for i, key := range rhs.Keys {
		if _, ok := d.Entries[key]; !ok {
			d.AddOrSet(key, rhs.Values[i])
		}
	}
}

func (d *KeyValueSet) AddOrSet(key string, value string) {
	i, ok := d.Entries[key]
	if !ok {
		d.Entries[key] = len(d.Values)
		d.Keys = append(d.Keys, key)
		d.Values = append(d.Values, value)
	} else if d.Values[i] != value {
		d.Values[i] = value
	}
}

func (s *KeyValueSet) ValuesToAdd(values ...string) {
	for _, value := range values {
		s.AddOrSet(value, value)
	}
}

// Enumerate will call the enumerator function for each key-value pair in the dictionary.
//
//	'last' will be 0 for all but the last key-value pair, and 1 for the last key-value pair.
func (d *KeyValueSet) Enumerate(enumerator func(i int, key string, value string, last int)) {
	for i, key := range d.Keys {
		if i == len(d.Keys)-1 {
			enumerator(i, key, d.Values[i], 1)
		} else {
			enumerator(i, key, d.Values[i], 0)
		}
	}
}

func (d *KeyValueSet) Concatenated(prefix string, suffix string, valueModifier func(string, string) string) string {
	concat := ""
	for i, value := range d.Values {
		concat += prefix + valueModifier(d.Keys[i], value) + suffix
	}
	return concat
}
