package ccode_utils

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

type Vars struct {
	Values [][]string
	Keys   map[string]int
}

func NewVars() *Vars {
	return &Vars{
		Values: make([][]string, 0, 4),
		Keys:   make(map[string]int, 4),
	}
}

func (v *Vars) Set(key string, value ...string) {
	if i, ok := v.Keys[key]; !ok {
		v.Keys[key] = len(v.Values)
		v.Values = append(v.Values, value)
	} else {
		v.Values[i] = value
	}
}

func (v *Vars) SetMany(vars map[string][]string) {
	for key, value := range vars {
		if i, ok := v.Keys[key]; !ok {
			v.Keys[key] = len(v.Values)
			v.Values = append(v.Values, value)
		} else {
			v.Values[i] = value
		}
	}
}

func (v *Vars) Append(key string, value ...string) {
	if i, ok := v.Keys[key]; !ok {
		v.Keys[key] = len(v.Values)
		v.Values = append(v.Values, value)
	} else {
		v.Values[i] = append(v.Values[i], value...)
	}
}

func (v *Vars) GetOne(key string) string {
	if i, ok := v.Keys[key]; ok {
		values := v.Values[i]
		if len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func (v *Vars) GetAll(key string) []string {
	if i, ok := v.Keys[key]; ok {
		return v.Values[i]
	}
	return nil
}

func (v *Vars) ResolveString(variable string) string {
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
			if vi, ok := v.Keys[variableName]; ok {
				// Replace the variable with its value (len(values) should be 1)
				values := v.Values[vi]
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

// ResolveInterpolation resolves the variable interpolation in the given text
// Note: See `vars.md` for the syntax of the variable interpolation.
func (v *Vars) ResolveInterpolation(text string) []string {
	type pair struct {
		bvar int // start of variable
		evar int // end of variable
		bopt int // start of option (-1 for no option)
		eopt int // end of option
	}

	// TODO These array's could be part of the Vars struct, so as to not allocate them each time
	stack := make([]int, 0, 4)
	list := make([]pair, 0, 4)

	finalResult := make([]string, 0, 4) // result of the interpolation

	toProcessIndex := 0
	toProcess := []string{text}
	for toProcessIndex < len(toProcess) {

		runes := []rune(toProcess[toProcessIndex])

		stack := stack[:0] // reset the stack
		list = list[:0]    // reset the list of pairs

		current := -1 // current index in the list of the pair we are working on
		for i := 0; i < len(runes); i++ {
			c := runes[i]
			if c == '$' && (i+1) < len(runes) && runes[i+1] == '(' {
			jump_label_parse_var:
				current = len(list)                                                           // start anew
				list = append(list, pair{bvar: i + 2, evar: i + 2, bopt: i + 2, eopt: i + 2}) // add the new pair
				i += 2                                                                        // skip the '$('

				for i < len(runes) {
					c := runes[i]
					if c == ':' || c == ')' {
						if c == ':' {
							list[current].evar = i // end of variable
							i++                    // skip ':'
							list[current].bopt = i // do initialize the start
							list[current].eopt = i // and end of the option
							for i < len(runes) {   // advance until we find ')'
								c = runes[i]
								if c == ')' {
									list[current].eopt = i
									break
								}
								i++
							}
						} else {
							list[current].evar = i
							list[current].bopt = i // no options, but we still set the start and
							list[current].eopt = i // end of the option to the end of the variable
						}

						if list[current].bvar != -1 && list[current].evar != -1 {
							if len(stack) > 0 {
								// We have nested vars on the stack, pop it and continue
								current = stack[len(stack)-1]
								stack = stack[:len(stack)-1]
							} else {
								current = -1
								break
							}
						}
					} else if c == '$' && (i+1) < len(runes) && runes[i+1] == '(' {
						stack = append(stack, current) // push current index to stack
						goto jump_label_parse_var      // start a new variable
					}
					i++
				}
			}
		}

		if len(list) == 0 {
			// This string has no variables, so just add it to the final result
			finalResult = append(finalResult, toProcess[toProcessIndex])
			toProcessIndex++
			continue
		}

		// The variable also contains normal text interleaved with variables, here we
		// replace each variable with its value.
		results := []string{""}

		rc := 0 // cursor in the runes
		for li, p := range list {
			if rc < (p.bvar - 2) { // There is some text before the variable, first add it every result entry
				text := string(runes[rc:(p.bvar - 2)])
				for j, _ := range results {
					results[j] += text
				}
			}

			variableName := string(runes[p.bvar:p.evar])
			if vi, ok := v.Keys[variableName]; ok {
				values := v.Values[vi]
				if len(values) == 0 {
					continue // skip empty values
				}

				// Apply options if any
				if p.bopt < p.eopt {
					join := ""

					// Ok, we are going to modify values, so clone it
					values = slices.Clone(values)

					options := runes[p.bopt:p.eopt]
					for len(options) > 0 {
						var option rune
						var param string
						options, option, param = consumeOption(options) // consume the option and its parameter

						for ii, value := range values {
							switch option {
							case 'f':
								values[ii] = actionForwardSlashes(value)
							case 'b':
								values[ii] = actionBackwardSlashes(value)
							case 'n':
								values[ii] = actionNativeSlashes(value)
							case 'u':
								values[ii] = actionUpperCase(value)
							case 'l':
								values[ii] = actionLowerCase(value)
							case 'B':
								values[ii] = actionBaseName(value)
							case 'F':
								values[ii] = actionFileName(value)
							case 'D':
								values[ii] = actionDirName(value)
							case 'p':
								values[ii] = actionPrefix(value, param)
							case 's':
								values[ii] = actionSuffix(value, param)
							case 'P':
								values[ii] = actionPrefixIfNotExists(value, param)
							case 'S':
								values[ii] = actionSuffixIfNotExists(value, param)
							case 'j':
								join = param
							default:
								fmt.Printf("Unknown interpolation option '%v' as part of $(%s:%s)\n", option, variableName, string(runes[p.bopt:p.eopt]))
								value = "?"
							}
						}
					}

					if len(join) > 0 {
						values = []string{actionJoinValues(values, join)}
					}
				}

				toProcess = append(toProcess, values...)

				// Two cases:
				// - one, this variable resulted in one value, so we can just continue with processing
				// - two, this variable resulted in multiple values, so we need to multiply the amount of results

				if len(values) > 1 {
					newResults := make([]string, 0, len(results)*len(values))
					for _, result := range results {
						for range values {
							newResults = append(newResults, result)
						}
					}
					results = newResults
				}
			} else {
				// This variable cannot be found
			}

			// If this is the last variable, we need to append the rest of the text
			if li == len(list)-1 && (p.eopt+1) < len(runes) {
				// There is some text after the variable, append it to the results
				text := string(runes[p.eopt+1:]) // everything after the variable
				for j := range results {
					results[j] += text
				}
			}
		}

		toProcessIndex++
	}

	return finalResult
}

// consumeOption consumes the option from runes and returns the option rune and option parameter
func consumeOption(runes []rune) ([]rune, rune, string) {
	if len(runes) == 0 {
		return runes, 0, ""
	}

	option := runes[0] // the first rune is the option
	cursor := 1

	w := 0
	for cursor < len(runes) {
		c := runes[cursor]
		if c == '\\' {
			cursor++ // Skip ('\' + next character)
			if cursor < len(runes) {
				c = runes[cursor]
			} else {
				break
			}
		} else if c == ':' {
			break
		}
		runes[w] = c
		w++
		cursor++
	}
	param := string(runes[:w])
	cursor = min(cursor, len(runes)) // Ensure we don't go out of bounds
	if (cursor + 1) < len(runes) {
		return runes[cursor+1:], option, param
	}
	return runes[0:0], option, param
}

func actionForwardSlashes(value string) string {
	return strings.ReplaceAll(value, "\\", "/")
}

func actionBackwardSlashes(value string) string {
	return strings.ReplaceAll(value, "/", "\\")
}

func actionNativeSlashes(value string) string {
	native := string(os.PathSeparator)
	nonnative := "/"
	if native == "/" {
		nonnative = "\\"
	}
	return strings.ReplaceAll(value, nonnative, native)
}
func actionUpperCase(value string) string {
	return strings.ToUpper(value)
}
func actionLowerCase(value string) string {
	return strings.ToLower(value)
}
func actionBaseName(value string) string {
	return PathFilename(value, false)
}
func actionFileName(value string) string {
	return PathFilename(value, true)
}
func actionDirName(value string) string {
	return PathDirname(value)
}

func actionPrefix(value string, prefix string) string {
	if len(value) == 0 {
		return value
	}
	return prefix + value
}

func actionSuffix(value string, suffix string) string {
	if len(value) == 0 {
		return value
	}
	return value + suffix
}

func actionPrefixIfNotExists(value string, prefix string) string {
	if strings.HasPrefix(value, prefix) {
		return value
	}
	return prefix + value
}

func actionSuffixIfNotExists(value string, suffix string) string {
	if strings.HasSuffix(value, suffix) {
		return value
	}
	return value + suffix
}

func actionJoinValues(values []string, sep string) string {
	return strings.Join(values, sep)
}
