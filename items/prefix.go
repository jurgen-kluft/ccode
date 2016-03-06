package items

import (
	"path/filepath"
)

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
