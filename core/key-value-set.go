package corepkg

import (
	"slices"
	"strings"
)

type KeyValueSet struct {
	Values []string       // values corresponding to the keys
	Keys   []string       // original case and order of insertion
	KeyMap map[string]int // map for case-insensitive key lookup
}

func NewKeyValueSet() *KeyValueSet {
	return &KeyValueSet{
		Values: make([]string, 0),
		Keys:   make([]string, 0),
		KeyMap: make(map[string]int),
	}
}

func (kv *KeyValueSet) String() string {
	result := ""
	for i, key := range kv.Keys {
		result += key + "=" + kv.Values[i] + "\n"
	}
	return result
}

func (kv *KeyValueSet) SortByKeys() {
	values := make([]string, len(kv.Values))
	keys := make([]string, len(kv.Keys))
	keymap := make(map[string]int)

	slices.Sort(kv.Keys)
	for _, key := range kv.Keys {
		index := kv.KeyMap[strings.ToLower(key)]
		values = append(values, kv.Values[index])
		keys = append(keys, key)
		keymap[strings.ToLower(key)] = len(keys) - 1
	}

	kv.Values = values
	kv.Keys = keys
	kv.KeyMap = keymap
}

// Merge merges another KeyValueSet into this one, preserving the original case of keys.
// When a key exists in both sets, the value from the current set is overwritten by the value from the other set.
func (kv *KeyValueSet) Merge(other *KeyValueSet) {
	if other == nil {
		return
	}
	for index, key := range other.Keys {
		// Convert key to lowercase for case-insensitive comparison
		lcKey := strings.ToLower(key)
		if i, exists := kv.KeyMap[lcKey]; !exists {
			kv.KeyMap[lcKey] = len(kv.Keys)
			kv.Keys = append(kv.Keys, key) // keep original case
			kv.Values = append(kv.Values, other.Values[index])
		} else {
			kv.Values[i] = other.Values[index]
		}
	}
}

// Join merges another KeyValueSet into this one, adding only keys that do not already exist in this set.
func (kv *KeyValueSet) Join(other *KeyValueSet) {
	if other == nil {
		return
	}
	for index, key := range other.Keys {
		// Convert key to lowercase for case-insensitive comparison
		lcKey := strings.ToLower(key)
		if _, exists := kv.KeyMap[lcKey]; !exists {
			kv.KeyMap[lcKey] = len(kv.Keys)
			kv.Keys = append(kv.Keys, key) // keep original case
			kv.Values = append(kv.Values, other.Values[index])
		}
	}
}

func (kv *KeyValueSet) Add(key string, value string) {
	// Convert key to lowercase for case-insensitive comparison
	lcKey := strings.ToLower(key)
	if index, exists := kv.KeyMap[lcKey]; !exists {
		index = len(kv.Values)
		kv.KeyMap[lcKey] = index
		kv.Keys = append(kv.Keys, key) // keep original case
		kv.Values = append(kv.Values, value)
	} else {
		// If the key already exists, update the value
		kv.Values[index] = value
	}
}

func (kv *KeyValueSet) AddOrSet(key string, value string) {
	// Convert key to lowercase for case-insensitive comparison
	lcKey := strings.ToLower(key)
	if index, exists := kv.KeyMap[lcKey]; !exists {
		index = len(kv.Values)
		kv.KeyMap[lcKey] = index
		kv.Keys = append(kv.Keys, key) // keep original case
		kv.Values = append(kv.Values, value)
	} else {
		// If the key already exists, update the value
		kv.Values[index] = value
	}
}

func (s *KeyValueSet) ValuesToAdd(values ...string) {
	for _, value := range values {
		s.AddOrSet(value, value)
	}
}

func (kv *KeyValueSet) Has(key string) bool {
	// Convert key to lowercase for case-insensitive comparison
	lcKey := strings.ToLower(key)
	_, exists := kv.KeyMap[lcKey]
	return exists
}

func (kv *KeyValueSet) HasGet(key string) (string, bool) {
	// Convert key to lowercase for case-insensitive comparison
	lcKey := strings.ToLower(key)
	if index, exists := kv.KeyMap[lcKey]; exists {
		return kv.Values[index], true
	}
	return "", false
}

func (kv *KeyValueSet) Get(key string) string {
	// Convert key to lowercase for case-insensitive comparison
	lcKey := strings.ToLower(key)
	if index, exists := kv.KeyMap[lcKey]; exists {
		return kv.Values[index]
	}
	return ""
}

func (d *KeyValueSet) Copy() *KeyValueSet {
	c := NewKeyValueSet()
	c.Merge(d)
	return c
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
