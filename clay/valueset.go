package clay

import "strings"

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
