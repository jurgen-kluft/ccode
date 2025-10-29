package corepkg

import (
	"runtime"
	"testing"
)

// Write a unittest for PathMatchWildcard and PathMatchWildcardOptimized

func TestPathMatchWildcard(t *testing.T) {
	tests := []struct {
		path     string
		wildcard string
		expected bool
	}{
		{"", "", true},
		{"aaaa", "*", true},
		{"", "a", false},
		{"a", "", false},
		{"a", "*", true},
		{"a", "a", true},
		{"a", "b", false},
		{"a", "?", true},
		{"a", "??", false},
		{"ab", "??", true},
		{"ab", "a*", true},
		{"ab", "a?", true},
		{"file.txt", "*.txt", true},
		{"file.doc.txt", "*.txt", false},
	}

	for _, test := range tests {
		if output := PathMatchWildcard(test.path, test.wildcard, true); output != test.expected {
			t.Errorf("PathMatchWildcard(%q, %q) = %v", test.path, test.wildcard, output)
		}
	}

	for _, test := range tests {
		if output := PathMatchWildcardOptimized(test.path, test.wildcard, true); output != test.expected {
			t.Errorf("PathMatchWildcard(%q, %q) = %v", test.path, test.wildcard, output)
		}
	}
}

func TestMatchCharCaseInsensitive(t *testing.T) {
	tests := []struct {
		a        rune
		b        rune
		expected bool
	}{
		{'a', 'a', true},
		{'a', 'A', true},
		{'A', 'a', true},
		{'A', 'A', true},
		{'a', 'b', false},
		{'b', 'a', false},
	}

	for _, test := range tests {
		if output := MatchCharCaseInsensitive(test.a, test.b); output != test.expected {
			t.Errorf("MatchCharCaseInsensitive(%q, %q) = %v", test.a, test.b, output)
		}
	}
}

func TestPathNormalize(t *testing.T) {
	win := runtime.GOOS == "windows"

	tests := []struct {
		path      string
		expected  string
		direction bool
	}{
		{"c:\\documents\\old", "c:/documents/old", !win},
	}

	for _, test := range tests {
		if test.direction {
			if output := PathNormalize(test.path); output != test.expected {
				t.Errorf("PathNormalize(%q) = %v", test.path, output)
			}
		} else {

			if output := PathNormalize(test.expected); output != test.path {
				t.Errorf("PathNormalize(%q) = %v", test.expected, output)
			}
		}
	}

}
