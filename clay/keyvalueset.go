package clay

import "strings"

type KeyValueSet struct {
	Values []string
	Keys   map[string]int
}

func NewKeyValueSet() *KeyValueSet {
	return &KeyValueSet{
		Values: make([]string, 0),
		Keys:   make(map[string]int),
	}
}

func (kv *KeyValueSet) Add(key string, value string) {
	// Convert key to lowercase for case-insensitive comparison
	lcKey := strings.ToLower(key)
	if index, exists := kv.Keys[lcKey]; !exists {
		index = len(kv.Values)
		kv.Keys[lcKey] = index
		kv.Values = append(kv.Values, value)
	} else {
		// If the key already exists, update the value
		kv.Values[index] = value
	}
}

func (kv *KeyValueSet) Has(key string) bool {
	// Convert key to lowercase for case-insensitive comparison
	lcKey := strings.ToLower(key)
	_, exists := kv.Keys[lcKey]
	return exists
}

func (kv *KeyValueSet) HasGet(key string) (string, bool) {
	// Convert key to lowercase for case-insensitive comparison
	lcKey := strings.ToLower(key)
	if index, exists := kv.Keys[lcKey]; exists {
		return kv.Values[index], true
	}
	return "", false
}

func (kv *KeyValueSet) Get(key string) string {
	// Convert key to lowercase for case-insensitive comparison
	lcKey := strings.ToLower(key)
	if index, exists := kv.Keys[lcKey]; exists {
		return kv.Values[index]
	}
	return ""
}
