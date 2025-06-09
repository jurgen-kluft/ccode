package ccode_utils

import (
	"strings"
	"testing"
)

const _separator = ", "

func TestMapToString(t *testing.T) {
	for name, test := range map[string]struct {
		str           string
		expectedParts []string
	}{
		"semicolon sep": {
			str: MapToString(
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
				"useSsl : true",
				"login : sa",
				"password : sa",
			},
		},
		"arrow sep": {
			str: MapToString(
				map[int]any{
					1:  "value 1",
					2:  "value 2",
					-5: "value -5",
				},
				"{key} => {value}",
				_separator,
			),
			expectedParts: []string{
				"1 => value 1",
				"2 => value 2",
				"-5 => value -5",
			},
		},
		"only value": {
			str: MapToString(
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
