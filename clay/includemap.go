package clay

import (
	"strings"
)

type IncludeMap struct {
	Values []string
	Keys   map[string]int
}

func NewIncludeMap() *IncludeMap {
	return &IncludeMap{
		Values: make([]string, 0),
		Keys:   make(map[string]int),
	}
}

func (kv *IncludeMap) Add(value string) {
	// Convert key to lowercase for case-insensitive comparison
	lcValue := strings.ToLower(value)
	if index, exists := kv.Keys[lcValue]; !exists {
		index = len(kv.Values)
		kv.Keys[lcValue] = index
		kv.Values = append(kv.Values, value)
	} else {
		// If the key already exists, update the value
		kv.Values[index] = value
	}
}

func (kv *IncludeMap) Has(value string) bool {
	// Convert value to lowercase for case-insensitive comparison
	lcValue := strings.ToLower(value)
	_, exists := kv.Keys[lcValue]
	return exists
}

func (kv *IncludeMap) HasGet(value string) (string, bool) {
	// Convert value to lowercase for case-insensitive comparison
	lcValue := strings.ToLower(value)
	if index, exists := kv.Keys[lcValue]; exists {
		return kv.Values[index], true
	}
	return "", false
}

func (kv *IncludeMap) Get(key string) string {
	// Convert key to lowercase for case-insensitive comparison
	lcValue := strings.ToLower(key)
	if index, exists := kv.Keys[lcValue]; exists {
		return kv.Values[index]
	}
	return ""
}
