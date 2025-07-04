package foundation

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

func TestInterpolateNew(t *testing.T) {
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
		{"Nested var with multiple values", "$(FOO)$(BAR)", []string{"StringA", "StringB", "StringC"}},
	}

	vars := NewVars()
	vars.Set("FOO", "String")
	vars.Set("BAR", "A", "B", "C")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewVarResolver()
			results := resolver.Resolve(tt.input, vars)

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

// TestInterpolateNew2, now variables are of the form '{VARIABLE}'
func TestInterpolateNew2(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"Simple variable", "{FOO}", []string{"String"}},
		{"Uppercase variable", "{FOO:u}", []string{"STRING"}},
		{"Lowercase variable", "{FOO:l}", []string{"string"}},
		{"Prefix with underscore", "{FOO:p__}", []string{"__String"}},
		{"Prefix and suffix with underscores", "{FOO:p__:s__}", []string{"__String__"}},
		{"Array variable", "{BAR}", []string{"A", "B", "C"}},
		{"Uppercase array variable", "{BAR:u}", []string{"A", "B", "C"}},
		{"Lowercase array variable", "{BAR:l}", []string{"a", "b", "c"}},
		{"Array with prefix", "{BAR:p__}", []string{"__A", "__B", "__C"}},
		{"Array with prefix and suffix", "{BAR:p__:s__:j!}", []string{"__A__!__B__!__C__"}},
		{"Array with custom separator", "{BAR:p\\::s!}", []string{":A!", ":B!", ":C!"}},
		{"Nested var with multiple values", "{FOO}{BAR}", []string{"StringA", "StringB", "StringC"}},
	}

	vars := NewVarsCustom(VarsFormatCurlyBraces)
	vars.Set("FOO", "String")
	vars.Set("BAR", "A", "B", "C")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewVarResolver()
			results := resolver.Resolve(tt.input, vars)

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

/*
## Nested Interpolation

Nested interpolation is possible, but should be used with care as it can be hard to
debug and understand. Here's an example of how the generic C toolchain inserts compiler
options dependening on what variant is currently active:

`$(CCOPTS_$(CURRENT_VARIANT:u))`

This works because the inner expansion will evalate `CURRENT_VARIANT` first (say, it
has the value `debug`). That value is then converted to upper-case and spliced into the
former which yields a new expression `$(CCOPTS_DEBUG)` which is then expanded in turn.

Used with care this is a powerful way of letting users customize variables per configuration
and then glue everything together with a simple template.
*/

func TestNestedInterpolationNew(t *testing.T) {

	vars := NewVars()
	vars.Set("CURRENT_VARIANT", "debug")
	vars.Set("CCOPTS_DEBUG", "-g -O0")

	resolver := NewVarResolver()
	result := resolver.Resolve("Test $(CCOPTS_$(CURRENT_VARIANT:u))", vars)

	if len(result) != 1 {
		t.Errorf("expected length %d, got %d", 1, len(result))
	}

	if result[0] != "Test -g -O0" {
		t.Errorf("expected %q, got %q", "Test -g -O0", result[0])
	}
}

// TestNestedInterpolationNew2, now variables are of the form '{VARIABLE}'
func TestNestedInterpolationNew2(t *testing.T) {

	vars := NewVarsCustom(VarsFormatCurlyBraces)
	vars.Set("CURRENT_VARIANT", "debug")
	vars.Set("CCOPTS_DEBUG", "-g -O0")

	resolver := NewVarResolver()
	result := resolver.Resolve("Test {CCOPTS_{CURRENT_VARIANT:u}}", vars)

	if len(result) != 1 {
		t.Errorf("expected length %d, got %d", 1, len(result))
	}

	if result[0] != "Test -g -O0" {
		t.Errorf("expected %q, got %q", "Test -g -O0", result[0])
	}
}

func TestNestedInterpolationExtensiveNew(t *testing.T) {
	vars := NewVars()
	vars.Set("CURRENT_VARIANT", "debug")
	vars.Set("CCOPTS_DEBUG", "-g -O0")
	vars.Set("CCOPTS_RELEASE", "-O2")
	vars.Set("CCOPTS_TEST", "-g -O0 -DTEST")
	vars.Set("CCOPTS_PRODUCTION", "-O2 -DNDEBUG")
	vars.Set("CCOPTS_CUSTOM", "-g -O0 -DCUSTOM")

	resolver := NewVarResolver()
	result := resolver.Resolve("Test $(CCOPTS_$(CURRENT_VARIANT:u):p__:s__)", vars)
	if len(result) != 1 {
		t.Errorf("expected length %d, got %d", 1, len(result))
	}
	if result[0] != "Test __-g -O0__" {
		t.Errorf("expected %q, got %q", "Test __-g -O0__", result[0])
	}
}
