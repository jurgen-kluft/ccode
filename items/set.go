package items

import (
	"strings"
)

// Set holds a unique 'map' of @Items delimited by @Delimiter
type Set struct {
	Items     []string
	Delimiter string
	Quote     string
}

func (s Set) String() string {
	if len(s.Items) == 0 {
		return ""
	}
	str := ""
	for i, item := range s.Items {
		if i == 0 {
			str = s.Quote + item + s.Quote
		} else {
			str = str + s.Delimiter + s.Quote + item + s.Quote
		}
	}
	return str
}

func addUnique(items []string, itemtoadd string) []string {
	itemtoadd = strings.Trim(itemtoadd, " ;")
	if len(itemtoadd) > 0 {
		for _, item := range items {
			if strings.EqualFold(itemtoadd, item) {
				return items
			}
		}
		items = append(items, itemtoadd)
	}
	return items
}

// Join two Set's together
func (s Set) Join(set Set) {
	items := s.Items
	for _, item := range set.Items {
		items = addUnique(items, item)
	}
}

// Add adds an item to the Set only if it doesn't exist in the Set
func (s Set) Add(item string) Set {
	items := addUnique(s.Items, item)
	return Set{Items: items, Delimiter: s.Delimiter, Quote: s.Quote}
}

// SetToList converts a Set to a List
func SetToList(set Set) List {
	if len(set.Items) == 0 {
		return List{Items: []string{}, Delimiter: set.Delimiter, Quote: set.Quote}
	}
	listitems := []string{}
	copy(listitems, set.Items)
	return List{Items: listitems, Delimiter: set.Delimiter, Quote: set.Quote}
}
