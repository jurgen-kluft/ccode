package corepkg

import (
	"slices"
	"strings"
	"testing"
)

const _separator = ", "

func MapToStringSortedByKey[K ~string | ~int | ~uint | ~int32 | ~int64 | ~uint32 | ~uint64, V any](
	data map[K]V, format string, separator string) string {
	if len(data) == 0 {
		return ""
	}

	keys := make([]K, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	// Sort keys to ensure consistent order
	slices.Sort(keys)

	mapStr := NewStringBuilder()
	for i, k := range keys {
		if i > 0 {
			mapStr.WriteString(separator)
		}
		KeyValueToStringAppend(k, data[k], format, mapStr)
	}

	return mapStr.String()
}

func TestMapToString(t *testing.T) {
	for name, test := range map[string]struct {
		str           string
		expectedParts []string
	}{
		"semicolon sep": {
			str: MapToStringSortedByKey(
				map[string]any{
					"connectTimeout": 1000,
					"useSsl":         true,
					"login":          "sa",
					"password":       "sa",
				},
				"{key} : {value}",
				_separator,
			),
			expectedParts: []string{
				"connectTimeout : 1000",
				"login : sa",
				"password : sa",
				"useSsl : true",
			},
		},
		"arrow sep": {
			str: MapToStringSortedByKey(
				map[int]any{
					1:  "value 1",
					2:  "value 2",
					-5: "value -5",
				},
				"{key} => {value}",
				_separator,
			),
			expectedParts: []string{
				"-5 => value -5",
				"1 => value 1",
				"2 => value 2",
			},
		},
		"only value": {
			str: MapToStringSortedByKey(
				map[uint64]any{
					1: "value 1",
					2: "value 2",
					5: "value 5",
				},
				"{value}",
				_separator,
			),
			expectedParts: []string{
				"value 1",
				"value 2",
				"value 5",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			actualParts := strings.Split(test.str, _separator)
			if len(test.expectedParts) != len(actualParts) {
				t.Errorf("expected %d parts, got %d", len(test.expectedParts), len(actualParts))
			}
			for i, part := range test.expectedParts {
				if part != actualParts[i] {
					t.Errorf("expected part %d to be '%s', got '%s'", i, part, actualParts[i])
				}
			}
		})
	}
}
