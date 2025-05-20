package denv

// ----------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------

type DevValueSet struct {
	Entries map[string]int
	Values  []string
}

func (d *DevValueSet) Merge(other *DevValueSet) {
	for _, value := range other.Values {
		d.Add(value)
	}
}

func NewDevValueSet() *DevValueSet {
	d := &DevValueSet{}
	d.Entries = make(map[string]int)
	d.Values = make([]string, 0)
	return d
}

func (d *DevValueSet) Extend(rhs *DevValueSet) {
	for _, value := range rhs.Values {
		d.Add(value)
	}
}

func (d *DevValueSet) UniqueExtend(rhs *DevValueSet) {
	for _, value := range rhs.Values {
		if _, ok := d.Entries[value]; !ok {
			d.Add(value)
		}
	}
}

func (d *DevValueSet) Add(value string) {
	i, ok := d.Entries[value]
	if !ok {
		d.Entries[value] = len(d.Values)
		d.Values = append(d.Values, value)
	} else if d.Values[i] != value {
		d.Values[i] = value
	}
}

func (d *DevValueSet) AddMany(values ...string) {
	for _, value := range values {
		d.Add(value)
	}
}

// Enumerate will call the enumerator function for each key-value pair in the dictionary.
//
//	'last' will be 0 for all but the last key-value pair, and 1 for the last key-value pair.
func (d *DevValueSet) Enumerate(enumerator func(i int, key string, value string, last int)) {
	for i, key := range d.Values {
		if i == len(d.Values)-1 {
			enumerator(i, key, d.Values[i], 1)
		} else {
			enumerator(i, key, d.Values[i], 0)
		}
	}
}

func (d *DevValueSet) Concatenated(prefix string, suffix string, valueModifier func(string) string) string {
	concat := ""
	for _, value := range d.Values {
		concat += prefix + valueModifier(value) + suffix
	}
	return concat
}
