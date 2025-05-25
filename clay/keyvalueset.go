package clay

import "strings"

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

func (kv *KeyValueSet) Merge(other *KeyValueSet, overwrite bool) {
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
		} else if overwrite {
			kv.Values[i] = other.Values[index]
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
