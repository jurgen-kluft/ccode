package ccode_utils

import (
	"testing"
)

/*
## Interpolation Examples

Assume there is an environment with the following bindings:

|     Key               |  Value
| --------------------- | ------------------------------------------------------------
|   `FOO`               |   `"String"`
|   `BAR`               |   `{ "A", "B", "C" }`


Then interpolating the following strings will give the associated result:

|   Expression          |   Resulting String
| --------------------- | ------------------------------------------------------------
|`$(FOO)`               |`String`
|`$(FOO:u)`             |`STRING`
|`$(FOO:l)`             |`string`
|`$(FOO:p__)`           |`__String`
|`$(FOO:p__:s__)`       |`__String__`
|`$(BAR)`               |`A B C`
|`$(BAR:u)`             |`A B C`
|`$(BAR:l)`             |`a b c`
|`$(BAR:p__)`           |`__A __B __C`
|`$(BAR:p__:s__:j!)`    |`__A__!__B__!__C__`
|`$(BAR:p\::s!)`        |`:A! :B! :C!`
|`$(BAR:AC)`            |`AC BC C`
*/

func TestInterpolate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"Simple variable", "$(FOO)", []string{"String"}},
		{"Uppercase variable", "$(FOO:u)", []string{"STRING"}},
		{"Lowercase variable", "$(FOO:l)", []string{"string"}},
		{"Prefix with underscore", "$(FOO:p__)", []string{"__String"}},
		{"Prefix and suffix with underscores", "$(FOO:p__:s__)", []string{"__String__"}},
		{"Array variable", "$(BAR)", []string{"A", "B", "C"}},
		{"Uppercase array variable", "$(BAR:u)", []string{"A", "B", "C"}},
		{"Lowercase array variable", "$(BAR:l)", []string{"a", "b", "c"}},
		{"Array with prefix", "$(BAR:p__)", []string{"__A", "__B", "__C"}},
		{"Array with prefix and suffix", "$(BAR:p__:s__:j!)", []string{"__A__!__B__!__C__"}},
		{"Array with custom separator", "$(BAR:p\\::s!)", []string{":A!", ":B!", ":C!"}},
	}

	vars := NewVars()
	vars.Set("FOO", "String")
	vars.Set("BAR", "A", "B", "C")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := vars.ResolveInterpolation(tt.input)

			if len(results) != len(tt.expected) {
				t.Errorf("expected length %q, got %q", len(tt.expected), len(results))
			}

			for i, result := range results {
				if result != tt.expected[i] {
					t.Errorf("expected %q, got %q", tt.expected[i], result)
				}
			}
		})
	}
}
