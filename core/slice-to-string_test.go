package corepkg

import (
	"testing"
)

func TestSliceToString(t *testing.T) {
	for name, test := range map[string]struct {
		separator      string
		data           []any
		expectedResult string
	}{
		"comma-separated slice": {
			separator:      ", ",
			data:           []any{11, 22, 33, 44, 55, 66, 77, 88, 99},
			expectedResult: "11, 22, 33, 44, 55, 66, 77, 88, 99",
		},
		"dash(kebab) line from slice": {
			separator:      "-",
			data:           []any{"str1", "str2", 101, "str3"},
			expectedResult: "str1-str2-101-str3",
		},
	} {
		t.Run(name, func(t *testing.T) {
			sb := NewStringBuilder()
			sb.SliceToString(&test.data, &test.separator)
			actualResult := sb.String()
			if actualResult != test.expectedResult {
				t.Errorf("expected '%s', got '%s'", test.expectedResult, actualResult)
			}
		})
	}
}

func TestSliceSameTypeToString(t *testing.T) {
	separator := ":"
	numericSlice := []int{100, 200, 400, 800}
	sb := NewStringBuilder()
	SliceSameTypeToString(sb, &numericSlice, &separator)
	result := sb.String()
	if result != "100:200:400:800" {
		t.Errorf("expected '100:200:400:800', got '%s'", result)
	}
}
