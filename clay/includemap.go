package clay

import (
	"fmt"
	"strings"
)

type IncludeDir struct {
	Prefix      bool   // This is a prefix include path
	IncludePath string // This is the include path
}

type IncludeMap struct {
	Values []IncludeDir
	Keys   map[string]int
}

func NewIncludeMap() *IncludeMap {
	return &IncludeMap{
		Values: make([]IncludeDir, 0),
		Keys:   make(map[string]int),
	}
}

func (kv *IncludeMap) Add(value string, prefix bool) {
	if prefix {
		fmt.Println("Warning: Adding '/' to prefix include path to ensure it can concatenate")
		if !strings.HasSuffix(value, "/") {
			value += "/"
		}
	}

	// Convert key to lowercase for case-insensitive comparison
	lcValue := strings.ToLower(value)
	if index, exists := kv.Keys[lcValue]; !exists {
		index = len(kv.Values)
		kv.Keys[lcValue] = index
		kv.Values = append(kv.Values, IncludeDir{
			Prefix:      prefix,
			IncludePath: value,
		})
	} else {
		// If the key already exists, update the value
		kv.Values[index] = IncludeDir{
			Prefix:      prefix,
			IncludePath: value,
		}
	}
}

func (kv *IncludeMap) Has(value string) bool {
	// Convert value to lowercase for case-insensitive comparison
	lcValue := strings.ToLower(value)
	_, exists := kv.Keys[lcValue]
	return exists
}

func (kv *IncludeMap) HasGet(value string) (IncludeDir, bool) {
	// Convert value to lowercase for case-insensitive comparison
	lcValue := strings.ToLower(value)
	if index, exists := kv.Keys[lcValue]; exists {
		return kv.Values[index], true
	}
	return IncludeDir{}, false
}

func (kv *IncludeMap) Get(key string) IncludeDir {
	// Convert key to lowercase for case-insensitive comparison
	lcValue := strings.ToLower(key)
	if index, exists := kv.Keys[lcValue]; exists {
		return kv.Values[index]
	}
	return IncludeDir{}
}
