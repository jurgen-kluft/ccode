package dev

type ValueSet struct {
	Entries map[string]int
	Values  []string
}

func (d *ValueSet) Merge(other *ValueSet) {
	for _, value := range other.Values {
		d.Add(value)
	}
}

func (d *ValueSet) Copy() *ValueSet {
	c := NewValueSet()
	c.Merge(d)
	return c
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
