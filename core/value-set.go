package corepkg

import (
	"strings"
)

type ValueSet struct {
	Values []string
	KeyMap map[string]int
}

func NewValueSet() *ValueSet {
	return &ValueSet{
		Values: make([]string, 0),
		KeyMap: make(map[string]int),
	}
}

func NewValueSet2(size int) *ValueSet {
	return &ValueSet{
		Values: make([]string, 0, size),
		KeyMap: make(map[string]int, size),
	}
}

func (kv *ValueSet) Add(value string) {
	// Convert key to lowercase for case-insensitive comparison
	lcValue := strings.ToLower(value)
	if index, exists := kv.KeyMap[lcValue]; !exists {
		index = len(kv.Values)
		kv.KeyMap[lcValue] = index
		kv.Values = append(kv.Values, value)
	} else {
		// If the key already exists, update the value
		kv.Values[index] = value
	}
}

func (kv *ValueSet) AddMany(value ...string) {
	for _, v := range value {
		kv.Add(v)
	}
}

func (kv *ValueSet) Has(value string) bool {
	// Convert value to lowercase for case-insensitive comparison
	lcValue := strings.ToLower(value)
	_, exists := kv.KeyMap[lcValue]
	return exists
}

func (kv *ValueSet) HasGet(value string) (string, bool) {
	// Convert value to lowercase for case-insensitive comparison
	lcValue := strings.ToLower(value)
	if index, exists := kv.KeyMap[lcValue]; exists {
		return kv.Values[index], true
	}
	return "", false
}

func (kv *ValueSet) Get(value string) string {
	// Convert value to lowercase for case-insensitive comparison
	lcValue := strings.ToLower(value)
	if index, exists := kv.KeyMap[lcValue]; exists {
		return kv.Values[index]
	}
	return ""
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

func (d *ValueSet) EncodeJson(encoder *JsonEncoder, key string) {
	encoder.WriteStringArray(key, d.Values)
}

func (d *ValueSet) DecodeJson(decoder *JsonDecoder) {
	d.Values = decoder.DecodeStringArray()
	for i, v := range d.Values {
		d.KeyMap[v] = i
	}
}
