package util

import "strings"

// SepListToArray converts a delimited list of items into an array of items
func SepListToArray(list string, sep string) []string {
	return strings.Split(list, sep)
}
