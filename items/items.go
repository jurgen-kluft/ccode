package items

import (
	"strings"
)

// List holds a 'list' of @Items delimited by @Delimiter
type List struct {
	Items     []string
	Delimiter string
	Quote     string
}

// NewList will create a new list from the given string
func NewList(list, delimiter string, quote string) List {
	newlist := List{Items: []string{}, Delimiter: delimiter, Quote: quote}
	return newlist.Add(list)
}

// CopyList makes a copy of the incoming list and returns it
func CopyList(list List) List {
	items := make([]string, len(list.Items))
	copy(items, list.Items)
	return List{Items: items, Delimiter: list.Delimiter, Quote: list.Quote}
}

func (l List) String() string {
	if len(l.Items) == 0 {
		return ""
	}
	str := ""
	for i, item := range l.Items {
		if i == 0 {
			str = l.Quote + item + l.Quote
		} else {
			str = str + l.Delimiter + l.Quote + item + l.Quote
		}
	}
	return str
}

// Copy returns a copy of @l
func (l List) Copy() List {
	return CopyList(l)
}

// Add adds an item to the list
func (l List) Add(add string) List {
	add = strings.Trim(add, " "+l.Quote+l.Delimiter)
	if len(add) == 0 {
		return l
	}

	additems := strings.Split(add, l.Delimiter)
	for _, item := range additems {
		item = strings.Trim(item, " "+l.Quote+l.Delimiter)
		if len(item) > 0 {
			dup := false
			for _, litem := range l.Items {
				if litem == item {
					dup = true
					break
				}
			}
			if !dup {
				l.Items = append(l.Items, item)
			}
		}
	}
	return List{Items: l.Items, Delimiter: l.Delimiter, Quote: l.Quote}
}

// Merge will combine
func (l List) Merge(list List) List {
	items := list.String()
	return l.Add(items)
}

// Prefix will add a @prefix to every item in the @List
func (l List) Prefix(prefix string, prefixer Prefixer) List {
	currentitems := []string{}
	for _, item := range l.Items {
		item = strings.Trim(item, " "+l.Quote+l.Delimiter)
		if len(item) > 0 {
			item = prefixer(item, prefix)
			currentitems = append(currentitems, item)
		}
	}
	return List{Items: currentitems, Delimiter: l.Delimiter, Quote: l.Quote}
}

// ListToSet converts a List to a Set and removes duplicates
func ListToSet(list List) Set {
	set := Set{Items: []string{}, Delimiter: list.Delimiter, Quote: list.Quote}
	seen := map[string]bool{}
	for _, item := range list.Items {
		item = strings.Trim(item, " "+list.Quote+list.Delimiter)
		if len(item) > 0 {
			if _, dup := seen[item]; !dup {
				set = set.Add(item)
				seen[item] = true
			}
		}
	}
	return set
}
