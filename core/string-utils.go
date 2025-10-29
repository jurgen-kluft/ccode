package corepkg

func StrTrimDelimiters(s string, c byte) string {
	if len(s) >= 2 && s[0] == c && s[len(s)-1] == c {
		return s[1 : len(s)-1]
	}
	return s
}

func StrTrimQuotes(s string) string {
	return StrTrimDelimiters(s, '"')
}
