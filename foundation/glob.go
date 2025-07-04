package foundation

import (
	"os"
	"path"
	"strings"
	"unicode/utf8"
)

func GlobMatching(filepath string, glob string) bool {
	if match, err := PathMatch(glob, filepath); match == true && err == nil {
		return match
	}
	return false
}

var ErrBadPattern = path.ErrBadPattern

// Split a path on the given separator, respecting escaping.
func splitPathOnSeparator(path string, separator rune) []string {
	// if the separator is '\\', then we can just split...
	if separator == '\\' {
		return strings.Split(path, string(separator))
	}

	// otherwise, we need to be careful of situations where the separator was escaped
	cnt := strings.Count(path, string(separator))
	if cnt == 0 {
		return []string{path}
	}
	ret := make([]string, cnt+1)
	pathlen := len(path)
	separatorLen := utf8.RuneLen(separator)
	idx := 0
	for start := 0; start < pathlen; {
		end := indexRuneWithEscaping(path[start:], separator)
		if end == -1 {
			end = pathlen
		} else {
			end += start
		}
		ret[idx] = path[start:end]
		start = end + separatorLen
		idx++
	}
	return ret[:idx]
}

// Find the first index of a rune in a string,
// ignoring any times the rune is escaped using "\".
func indexRuneWithEscaping(s string, r rune) int {
	end := strings.IndexRune(s, r)
	if end == -1 {
		return -1
	}
	if end > 0 && s[end-1] == '\\' {
		start := end + utf8.RuneLen(r)
		end = indexRuneWithEscaping(s[start:], r)
		if end != -1 {
			end += start
		}
	}
	return end
}

// Match returns true if name matches the shell file name pattern.
// The pattern syntax is:
//
//	pattern:
//	  { term }
//	term:
//	  '*'         matches any sequence of non-path-separators
//	            '**'        matches any sequence of characters, including
//	                        path separators.
//	  '?'         matches any single non-path-separator character
//	  '[' [ '^' ] { character-range } ']'
//	        character class (must be non-empty)
//	  '{' { term } [ ',' { term } ... ] '}'
//	  c           matches character c (c != '*', '?', '\\', '[')
//	  '\\' c      matches character c
//
//	character-range:
//	  c           matches character c (c != '\\', '-', ']')
//	  '\\' c      matches character c
//	  lo '-' hi   matches character c for lo <= c <= hi
//
// Match requires pattern to match all of name, not just a substring.
// The path-separator defaults to the '/' character. The only possible
// returned error is ErrBadPattern, when pattern is malformed.
//
// Note: this is meant as a drop-in replacement for path.Match() which
// always uses '/' as the path separator. If you want to support systems
// which use a different path separator (such as Windows), what you want
// is the PathMatch() function below.
func Match(pattern, name string) (bool, error) {
	return matchWithSeparator(pattern, name, '/')
}

// PathMatch is like Match except that it uses your system's path separator.
// For most systems, this will be '/'. However, for Windows, it would be '\\'.
// Note that for systems where the path separator is '\\', escaping is
// disabled.
//
// Note: this is meant as a drop-in replacement for filepath.Match().
func PathMatch(pattern, name string) (bool, error) {
	pattern = strings.ToLower(pattern)
	name = strings.ToLower(name)
	return matchWithSeparator(pattern, name, os.PathSeparator)
}

// Match returns true if name matches the shell file name pattern.
// The pattern syntax is:
//
//	pattern:
//	  { term }
//	term:
//	  '*'         matches any sequence of non-path-separators
//	            '**'        matches any sequence of characters, including
//	                        path separators.
//	  '?'         matches any single non-path-separator character
//	  '[' [ '^' ] { character-range } ']'
//	        character class (must be non-empty)
//	  '{' { term } [ ',' { term } ... ] '}'
//	  c           matches character c (c != '*', '?', '\\', '[')
//	  '\\' c      matches character c
//
//	character-range:
//	  c           matches character c (c != '\\', '-', ']')
//	  '\\' c      matches character c, unless separator is '\\'
//	  lo '-' hi   matches character c for lo <= c <= hi
//
// Match requires pattern to match all of name, not just a substring.
// The only possible returned error is ErrBadPattern, when pattern
// is malformed.
func matchWithSeparator(pattern, name string, separator rune) (bool, error) {
	patternComponents := splitPathOnSeparator(pattern, separator)
	nameComponents := splitPathOnSeparator(name, separator)
	return doMatching(patternComponents, nameComponents)
}

func doMatching(patternComponents, nameComponents []string) (matched bool, err error) {
	// check for some base-cases
	patternLen, nameLen := len(patternComponents), len(nameComponents)
	if patternLen == 0 && nameLen == 0 {
		return true, nil
	}
	if patternLen == 0 || nameLen == 0 {
		return false, nil
	}

	patIdx, nameIdx := 0, 0
	for patIdx < patternLen && nameIdx < nameLen {
		if patternComponents[patIdx] == "**" {
			// if our last pattern component is a doublestar, we're done -
			// doublestar will match any remaining name components, if any.
			if patIdx++; patIdx >= patternLen {
				return true, nil
			}

			// otherwise, try matching remaining components
			for ; nameIdx < nameLen; nameIdx++ {
				if m, _ := doMatching(patternComponents[patIdx:], nameComponents[nameIdx:]); m {
					return true, nil
				}
			}
			return false, nil
		} else {
			// try matching components
			matched, err = matchComponent(patternComponents[patIdx], nameComponents[nameIdx])
			if !matched || err != nil {
				return
			}
		}
		patIdx++
		nameIdx++
	}
	return patIdx >= patternLen && nameIdx >= nameLen, nil
}

// Attempt to match a single pattern component with a path component
func matchComponent(pattern, name string) (bool, error) {
	// check some base cases
	patternLen, nameLen := len(pattern), len(name)
	if patternLen == 0 && nameLen == 0 {
		return true, nil
	}
	if patternLen == 0 {
		return false, nil
	}
	if nameLen == 0 && pattern != "*" {
		return false, nil
	}

	// check for matches one rune at a time
	patIdx, nameIdx := 0, 0
	for patIdx < patternLen && nameIdx < nameLen {
		patRune, patAdj := utf8.DecodeRuneInString(pattern[patIdx:])
		nameRune, nameAdj := utf8.DecodeRuneInString(name[nameIdx:])
		if patRune == '\\' {
			// handle escaped runes
			patIdx += patAdj
			patRune, patAdj = utf8.DecodeRuneInString(pattern[patIdx:])
			if patRune == utf8.RuneError {
				return false, ErrBadPattern
			} else if patRune == nameRune {
				patIdx += patAdj
				nameIdx += nameAdj
			} else if runeEqualFold(patRune, nameRune) {
				patIdx += patAdj
				nameIdx += nameAdj
			} else {
				return false, nil
			}
		} else if patRune == '*' {
			// handle stars
			if patIdx += patAdj; patIdx >= patternLen {
				// a star at the end of a pattern will always
				// match the rest of the path
				return true, nil
			}
			// check if we can make any matches
			for ; nameIdx < nameLen; nameIdx += nameAdj {
				if m, _ := matchComponent(pattern[patIdx:], name[nameIdx:]); m {
					return true, nil
				}
			}
			return false, nil
		} else if patRune == '[' {
			// handle character sets
			patIdx += patAdj
			endClass := indexRuneWithEscaping(pattern[patIdx:], ']')
			if endClass == -1 {
				return false, ErrBadPattern
			}
			endClass += patIdx
			classRunes := []rune(pattern[patIdx:endClass])
			classRunesLen := len(classRunes)
			if classRunesLen > 0 {
				classIdx := 0
				matchClass := false
				if classRunes[0] == '^' {
					classIdx++
				}
				for classIdx < classRunesLen {
					low := classRunes[classIdx]
					if low == '-' {
						return false, ErrBadPattern
					}
					classIdx++
					if low == '\\' {
						if classIdx < classRunesLen {
							low = classRunes[classIdx]
							classIdx++
						} else {
							return false, ErrBadPattern
						}
					}
					high := low
					if classIdx < classRunesLen && classRunes[classIdx] == '-' {
						// we have a range of runes
						if classIdx++; classIdx >= classRunesLen {
							return false, ErrBadPattern
						}
						high = classRunes[classIdx]
						if high == '-' {
							return false, ErrBadPattern
						}
						classIdx++
						if high == '\\' {
							if classIdx < classRunesLen {
								high = classRunes[classIdx]
								classIdx++
							} else {
								return false, ErrBadPattern
							}
						}
					}
					if low <= nameRune && nameRune <= high {
						matchClass = true
					}
				}
				if matchClass == (classRunes[0] == '^') {
					return false, nil
				}
			} else {
				return false, ErrBadPattern
			}
			patIdx = endClass + 1
			nameIdx += nameAdj
		} else if patRune == '{' {
			// handle alternatives such as {alt1,alt2,...}
			patIdx += patAdj
			endOptions := indexRuneWithEscaping(pattern[patIdx:], '}')
			if endOptions == -1 {
				return false, ErrBadPattern
			}
			endOptions += patIdx
			options := splitPathOnSeparator(pattern[patIdx:endOptions], ',')
			patIdx = endOptions + 1
			for _, o := range options {
				m, e := matchComponent(o+pattern[patIdx:], name[nameIdx:])
				if e != nil {
					return false, e
				}
				if m {
					return true, nil
				}
			}
			return false, nil
		} else if patRune == '?' || patRune == nameRune {
			// handle single-rune wildcard
			patIdx += patAdj
			nameIdx += nameAdj
		} else if runeEqualFold(patRune, nameRune) {
			patIdx += patAdj
			nameIdx += nameAdj
		} else {
			return false, nil
		}
	}
	if patIdx >= patternLen && nameIdx >= nameLen {
		return true, nil
	}
	if nameIdx >= nameLen && pattern[patIdx:] == "*" || pattern[patIdx:] == "**" {
		return true, nil
	}
	return false, nil
}

func runeEqualFold(a, b rune) bool {
	if a >= 'a' && a <= 'z' {
		a += (a - 'a') + 'A'
	}
	if b >= 'a' && b <= 'z' {
		b += (b - 'a') + 'A'
	}
	return a == b
}
