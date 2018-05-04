package glob

import (
	"testing"
)

type XMatchTest struct {
	pattern string
	name string
	matches []string // a pattern and path to test the pattern on
	shouldMatch       bool     // true if the pattern should match the path
	expectedErr       error    // an expected error
	testOnDisk        bool     // true: test pattern against files in "test" directory
}

// Tests which contain escapes and symlinks will not work on Windows

var xmatchTests = []XMatchTest{
	{"files\\**\\*.h", "*.h", []string{"header.h"}, true, nil, true},
	{"files\\**\\*.cpp", "*.cpp", []string{"source.cpp"}, true, nil, true},
}


func TestXPathMatch(t *testing.T) {
	for idx, tt := range xmatchTests {
		// Even though we aren't actually matching paths on disk, we are using
		// PathMatch() which will use the system's separator. As a result, any
		// patterns that might cause problems on-disk need to also be avoided
		// here in this test.
		if tt.testOnDisk {
			testXPathMatchWith(t, idx, tt) 
		}
	}
}

func testXPathMatchWith(t *testing.T, idx int, tt XMatchTest) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("#%v. Match(%#q, %#q) panicked: %#v", idx, tt.pattern, tt.name, r)
		}
	}()

	matches, err := Glob(tt.pattern)
	if len(matches) != len(tt.matches) || err != tt.expectedErr {
		t.Errorf("#%v. Match(%#q, %#q) = %v, %v want %v, %v", idx, tt.pattern, tt.name, len(matches), err, tt.shouldMatch, tt.expectedErr)
	}

	t.Log("Matches ...")
	for _, m := range matches {
		t.Log(m)
	}
}

