package util

import "strings"

// Seperate converts a delimited list of items into an array of items
func Seperate(list string, sep string) []string {
	return strings.Split(list, sep)
}

// Join will join every item from the array into a string using a given seperator
func Join(array []string, sep string) string {
	return strings.Join(array, sep)
}
