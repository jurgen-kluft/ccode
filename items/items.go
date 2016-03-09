package items

import (
	"strings"
)

// List holds a 'list' of @Items delimited by @Delimiter
type List struct {
	Items     []string
	Delimiter string
}

// NewList will create a new list from the given string
func NewList(list, delimiter string) List {
	newlist := List{Items: []string{}, Delimiter: delimiter}
	return newlist.Add(list)
}

// CopyList makes a copy of the incoming list and returns it
func CopyList(list List) List {
	items := make([]string, len(list.Items))
	copy(items, list.Items)
	return List{Items: items, Delimiter: list.Delimiter}
}

func (l List) String() string {
	if len(l.Items) == 0 {
		return ""
	}
	return strings.Join(l.Items, l.Delimiter)
}

// Add adds an item to the list
func (l List) Add(add string) List {
	add = strings.Trim(add, " ;")
	if len(add) == 0 {
		return l
	}
	additems := strings.Split(add, l.Delimiter)
	currentitems := []string{}
	for _, item := range append(l.Items, additems...) {
		item = strings.Trim(item, " ;")
		if len(item) > 0 {
			currentitems = append(currentitems, item)
		}
	}
	return List{Items: currentitems, Delimiter: l.Delimiter}
}

// Prefix will add a @prefix to every item in the @List
func (l List) Prefix(prefix string, prefixer Prefixer) List {
	currentitems := []string{}
	for _, item := range l.Items {
		item = strings.Trim(item, " ;")
		if len(item) > 0 {
			item = prefixer(item, prefix)
			currentitems = append(currentitems, item)
		}
	}
	return List{Items: currentitems, Delimiter: l.Delimiter}
}

// ListToSet converts a List to a Set
func ListToSet(list List) Set {
	set := Set{Items: []string{}, Delimiter: list.Delimiter}
	for _, item := range list.Items {
		item = strings.Trim(item, " ;")
		if len(item) > 0 {
			set = set.Add(item)
		}
	}
	return set
}
