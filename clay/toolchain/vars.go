package toolchain

import (
	"fmt"
)

type Vars map[string][]string

func (v Vars) Set(key string, value string) {
	if _, ok := v[key]; !ok {
		v[key] = []string{}
	}
	v[key] = []string{value}
}

func (v Vars) Append(key string, value string) {
	if _, ok := v[key]; !ok {
		v[key] = []string{}
	}
	v[key] = append(v[key], value)
}

func (v Vars) GetOne(key string) string {
	if values, ok := v[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

func (v Vars) GetAll(key string) []string {
	if values, ok := v[key]; ok && len(values) > 0 {
		return values
	}
	return []string{}
}

func ResolveString(variable string, vars Vars) string {
	type pair struct {
		from int
		to   int
	}

	for true {
		// find [from:to] pairs of variables
		stack := make([]int16, 0, 4)
		list := make([]pair, 0)
		for i, c := range variable {
			if c == '{' {
				current := int16(len(list))
				stack = append(stack, current)
				list = append(list, pair{from: i, to: -1})
			} else if c == '}' {
				if len(list) > 0 {
					current := stack[len(stack)-1]
					list[current].to = i
					stack = stack[:len(stack)-1]
				}
			}
		}

		if len(list) == 0 {
			return variable
		}

		// See if we have an invalid pair, if so just return
		for _, p := range list {
			if p.to == -1 {
				fmt.Printf("Invalid variable pair in string: %s\n", variable)
				return variable // Return the original string if we have an invalid pair
			}
		}

		// resolve the variables, last to first, and assume all pairs are valid and closed
		replaced := 0
		for i := len(list) - 1; i >= 0; i-- {
			p := list[i]
			variableName := variable[p.from+1 : p.to]
			// Check if the variable exists in the vars map
			if values, ok := vars[variableName]; ok {
				// Replace the variable with its value (len(values) should be 1)
				value := values[0]
				variable = variable[:p.from] + value + variable[p.to+1:]
				replaced += 1
				// It this was a nested variable (has overlap with previous pair(s)),
				// we need to adjust the 'to' of the next pairs
				for j := i - 1; j >= 0; j-- {
					if list[j].to > p.from {
						list[j].to += len(value) - (p.to - p.from + 1)
					}
				}
			} else {
				// If the variable does not exist, we skip it.
				// This variable could have been a nested one, so we need to skip also
				// the overlapping pairs
				for j := i - 1; j >= 0; j-- {
					if list[j].to > p.from {
						i-- // Skip this pair
					} else {
						break // No more overlapping pairs
					}
				}
			}
		}

		if replaced == 0 {
			return variable
		}

	}

	return variable
}
