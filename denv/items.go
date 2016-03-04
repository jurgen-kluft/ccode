package denv

import (
	"path/filepath"
	"strings"
)

// DefineList holds a list of defines
type ItemsList string

func (l ItemsList) String() string {
	return string(l)
}

type Prefixer func(string, string) string

func NoPrefixer(item string, prefix string) string {
	return item
}

func PathPrefixer(item string, prefix string) string {
	if len(item) == 0 {
		return item
	}
	return filepath.Join(prefix, item)
}

func (l ItemsList) Prefix(prefix string, delimiter string, prefixer Prefixer) ItemsList {
	items := strings.Split(string(l), delimiter)
	for i, item := range items {
		items[i] = prefixer(item, prefix)
	}
	return ItemsList(strings.Join(items, delimiter))
}

type ItemsSet map[string]bool

// Join two ItemsSet together
func (d ItemsSet) Join(items ItemsSet) {
	for k, v := range items {
		d[k] = v
	}
}
