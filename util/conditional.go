package util

func CRString(condition bool, first string, second string) string {
	if condition {
		return first
	}
	return second
}
