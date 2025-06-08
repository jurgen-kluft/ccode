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
		bvar  int  // start of variable
		evar  int  // end of variable
		bopt  int  // start of option (-1 for no option)
		eopt  int  // end of option
		depth int8 // (nested variables) depth of the variable
		child int  // (nested variables) child index
		prev  int  // (nested variables) sibling previous
		next  int  // (nested variables) sibling next
	}

	finalResult := make([]string, 0, 16) // result of the interpolation

	// TODO These array's could be part of the Vars struct, so as to not allocate them each time
	stack := make([]int, 0, 8)
	list := make([]pair, 0, 8)

	toProcessIndex := 0
	toProcess := []string{text}
	for toProcessIndex < len(toProcess) {

		runes := []rune(toProcess[toProcessIndex])

		maxdepth := int8(0) // maximum depth encountered
		stack := stack[:0]  // reset the stack
		list = list[:0]     // reset the list of pairs

		previous := -1
		current := -1 // current index in the list of the pair we are working on
		child := -1   // child index, used for nested variables
		for i := 0; i < len(runes); i++ {
			c := runes[i]
			if c == '$' && (i+1) < len(runes) && runes[i+1] == '(' {

			jump_label_parse_var:
				current = len(list)                        // start anew
				maxdepth = max(maxdepth, int8(len(stack))) // update the maximum depth
				list = append(list, pair{                  // add a new pair with the start of the variable
					bvar:  i + 2,
					evar:  i + 2,
					bopt:  i + 2,
					eopt:  i + 2,
					depth: int8(len(stack)),
					child: child,
					prev:  previous,
					next:  -1,
				})
				if previous >= 0 {
					list[previous].next = current // set the next index for the previous variable
				}

				i += 2 // skip the '$('

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

						if len(stack) > 0 {
							// We have nested vars on the stack, pop it and continue
							current = stack[len(stack)-1]
							stack = stack[:len(stack)-1]
							child = -1                    // reset child index for the new variable
							previous = list[current].prev // take the previous from current for the next variable
						} else {
							// We just closed a variable, if we had a previous variable, we need to link it to the current one
							if previous >= 0 {
								list[previous].next = current
							}
							// no more nested vars, so we can setup the parameters for the new variable
							child = -1
							previous = current
							current = -1
							break
						}

					} else if c == '$' && (i+1) < len(runes) && runes[i+1] == '(' {
						stack = append(stack, current) // push current index to stack
						child = len(list)              // set the child index for the new variable
						previous = -1                  // previous index is reset for the new variable
						child = -1                     // reset child index for the new variable
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

		// Now we have a list of pairs, each pair is a variable with options, depth and list of siblings
		// Here we resolve the variables, starting from the deepest ones
		resolved := make([][]string, len(list)) // resolved values for each pair
		for d := maxdepth; d >= 0; d-- {
			for i, p := range list {
				if p.depth == d {

					// We could have the following text:
					// $(BEST_DAMN_$(SUBVAR)_SUPER_$(TRAILVAR)_EVER)
					// Then $(SUBVAR) is resolved as MOVIE
					// Then we have:
					//     $(BEST_DAMN_MOVIE_SUPER_$(TRAILVAR)_EVER)
					//    $(TRAILVAR) is resolved as STAR
					// Then we have:
					//    $(BEST_DAMN_MOVIE_SUPER_STAR_EVER)
					// Note: This is, of course, resolved as 'Bruce Willis'

					variableName := string(runes[p.bvar:p.evar])
					if vi, ok := v.Keys[variableName]; ok {
						resolved[i] = v.Values[vi] // use the values from the variable
					} else {
						resolved[i] = []string{string(runes[p.bvar-2 : p.eopt+1])} // variable not found, just put back the var
					}

					// Concatenate all children, and since they are at a lower depth they should already be resolved
					child := p.child
					if child >= 0 {
						bvar := p.bvar
						for child >= 0 {
							c := list[child]
							// Prefix resolved with any text that the variable might have
							text := string(runes[bvar:(c.bvar - 2)])
							if len(text) > 0 {
								for ri, rv := range resolved[i] {
									resolved[i][ri] = text + rv
								}
							}

							values := make([]string, 0, len(resolved[i])*len(resolved[child]))
							for _, r := range resolved[i] {
								for _, c := range resolved[child] {
									values = append(values, r+c)
								}
							}

							bvar = c.eopt + 1 // update the text start
							child = list[child].next
						}

						if bvar < p.evar {
							// There is some text before the end of the variable, append it to the resolved values
							text := string(runes[bvar:p.evar]) // everything after the variable
							for ri, rv := range resolved[i] {
								resolved[i][ri] = rv + text
							}
						}
					}
				}
			}
		}

		// The variable also contains normal text interleaved with variables, here we
		// replace each variable with its value.
		results := []string{""}

		rc := 0 // cursor in the runes
		for li, p := range list {
			// There can be some text before this variable, so here we
			// consume that text and append it to the results
			if rc < (p.bvar - 2) {
				text := string(runes[rc:(p.bvar - 2)])
				for j, _ := range results {
					results[j] += text
				}
			}

			values := resolved[li]

			// Apply options if any
			if p.bopt < p.eopt {
				// Ok, we are going to modify values, so clone it
				values = slices.Clone(values)

				options := runes[p.bopt:p.eopt]
				for len(options) > 0 {
					var option rune
					var param string
					options, option, param = consumeOption(options) // consume the option and its parameter

					switch option {
					case 'f':
						values = actionForwardSlashes(values, param)
					case 'b':
						values = actionBackwardSlashes(values, param)
					case 'n':
						values = actionNativeSlashes(values, param)
					case 'u':
						values = actionUpperCase(values, param)
					case 'l':
						values = actionLowerCase(values, param)
					case 'B':
						values = actionBaseName(values, param)
					case 'F':
						values = actionFileName(values, param)
					case 'D':
						values = actionDirName(values, param)
					case 'p':
						values = actionPrefix(values, param)
					case 's':
						values = actionSuffix(values, param)
					case 'P':
						values = actionPrefixIfNotExists(values, param)
					case 'S':
						values = actionSuffixIfNotExists(values, param)
					case 'j':
						values = actionJoinValues(values, param)
					default:
						fmt.Printf("Unknown interpolation option '%v' as part of $(%s:%s)\n", option, string(runes[p.bvar:p.evar]), string(runes[p.bopt:p.eopt]))
					}
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

func actionForwardSlashes(values []string, param string) []string {
	for i, value := range values {
		values[i] = strings.ReplaceAll(value, "\\", "/")
	}
	return values
}

func actionBackwardSlashes(values []string, param string) []string {
	for i, value := range values {
		values[i] = strings.ReplaceAll(value, "/", "\\")
	}
	return values
}

func actionNativeSlashes(values []string, param string) []string {
	native := string(os.PathSeparator)
	nonnative := "/"
	if native == "/" {
		nonnative = "\\"
	}
	for i, value := range values {
		values[i] = strings.ReplaceAll(value, nonnative, native)
	}
	return values
}
func actionUpperCase(values []string, param string) []string {
	for i, value := range values {
		values[i] = strings.ToUpper(value)
	}
	return values
}
func actionLowerCase(values []string, param string) []string {
	for i, value := range values {
		values[i] = strings.ToLower(value)
	}
	return values
}
func actionBaseName(values []string, param string) []string {
	for i, value := range values {
		values[i] = PathFilename(value, false)
	}
	return values
}
func actionFileName(values []string, param string) []string {
	for i, value := range values {
		values[i] = PathFilename(value, true)
	}
	return values
}
func actionDirName(values []string, param string) []string {
	for i, value := range values {
		values[i] = PathDirname(value)
	}
	return values
}

func actionDelimitValue(values []string, param string) []string {
	for i, value := range values {
		values[i] = param + value + param
	}
	return values
}

func actionTrimValue(values []string, param string) []string {
	for i, value := range values {
		if strings.HasPrefix(value, param) {
			values[i] = value[len(param):]
		}
		if strings.HasSuffix(value, param) {
			values[i] = value[:len(value)-len(param)]
		}
	}
	return values
}

func actionTrimValueAny(values []string, param string) []string {
	for i, value := range values {
		runes := []rune(value)
		// Trim any character from the start
		for len(runes) > 0 {
			if strings.ContainsRune(param, runes[0]) {
				runes = runes[1:]
			} else {
				break
			}
		}
		// Trim any character from the end
		for len(runes) > 0 {
			if strings.ContainsRune(param, runes[len(runes)-1]) {
				runes = runes[:len(runes)-1]
			} else {
				break // No more characters to remove
			}
		}
		values[i] = string(runes)
	}
	return values
}

func actionPrefix(values []string, prefix string) []string {
	for i, value := range values {
		values[i] = prefix + value
	}
	return values
}

func actionSuffix(values []string, suffix string) []string {
	for i, value := range values {
		values[i] = value + suffix
	}
	return values
}

func actionPrefixIfNotExists(values []string, prefix string) []string {
	for i, value := range values {
		if strings.HasPrefix(value, prefix) {
			values[i] = value
		} else {
			values[i] = prefix + value
		}
	}
	return values
}

func actionSuffixIfNotExists(values []string, suffix string) []string {
	for i, value := range values {
		if strings.HasSuffix(value, suffix) {
			values[i] = value
		} else {
			values[i] = value + suffix
		}
	}
	return values
}

func actionJoinValues(values []string, join string) []string {
	return []string{strings.Join(values, join)}
}

var varOptionActionMap = map[uint8]func([]string, string) []string{
	'f': actionForwardSlashes,
	'b': actionBackwardSlashes,
	'n': actionNativeSlashes,
	'u': actionUpperCase,
	'l': actionLowerCase,
	'B': actionBaseName,
	'F': actionFileName,
	'D': actionDirName,
	'p': actionPrefix,
	's': actionSuffix,
	'P': actionPrefixIfNotExists,
	'S': actionSuffixIfNotExists,
	'j': actionJoinValues,
}

// ---------------------------------------------------------------------------------------
// A parse/resolve approach

// We could have the following text:
// "Of course $(BEST_DAMN_$(SUBVAR)_SUPER_$(TRAILVAR)_EVER)"
// Then $(SUBVAR) is resolved as MOVIE
// Then we have:
//     $(BEST_DAMN_MOVIE_SUPER_$(TRAILVAR)_EVER)
//    $(TRAILVAR) is resolved as STAR
// Then we have:
//    "Of course $(BEST_DAMN_MOVIE_SUPER_STAR_EVER)""
// Note: This is, of course, resolved as 'Of course Bruce Willis'

type VarResolver interface {
	Parse(text string) int
	Resolve(vars *Vars) []string
}

func NewVarResolver() VarResolver {
	vr := &varResolver{
		strings: make([]varText, 0, 16),
		options: make([]varOption, 0, 16),
		nodes:   make([]varNode, 0, 16),
		source:  nil,
		cursor:  0,
		current: 0,
		stack:   make([]int, 0, 4),
	}
	return vr
}

// ---------------------------------------------------------------------------------------
type varResolver struct {
	strings []varText   // list of strings
	options []varOption // list of options, each option is a single character
	nodes   []varNode   // list of nodes
	source  []rune      // source text that we are parsing
	cursor  int         // current cursor in the text
	current int         // current node
	stack   []int       // stack of nodes
}

func (vr *varResolver) Parse(text string) int {
	vr.strings = vr.strings[:0]
	vr.options = vr.options[:0]
	vr.nodes = vr.nodes[:0]                   // Reset the nodes, we will parse again
	vr.nodes = append(vr.nodes, newVarNode()) // Start with the root node
	vr.source = []rune(text)
	vr.cursor = 0
	vr.current = 0
	vr.stack = vr.stack[:0]
	return vr.parse()
}

func (vr *varResolver) Resolve(vars *Vars) []string {
	return vr.resolveNode(vars, 0)
}

type varText struct {
	from int16
	to   int16 // from and to are indexes in the text slice
}

type varOption struct {
	param    int   // start of the option in varParseContext.text
	paramLen int16 // end of the option in varParseContext.text
	option   uint8 // the option character, e.g. 'f', 'b', etc.
	end      int8  // end = 1 means this is the last option
}

func newVarOption(opt uint8) varOption {
	return varOption{
		param:    -1, // -1 means no parameter
		paramLen: -1, // -1 means no parameter
		option:   opt,
		end:      0, // 0 means this is not the last option
	}
}

func (vr *varResolver) scanForVariable() int {
	cursor := vr.cursor
	for cursor < len(vr.source) {
		if vr.source[cursor] == '$' && (cursor+1) < len(vr.source) && vr.source[cursor+1] == '(' {
			return cursor
		}
		cursor++
	}
	return cursor
}

func (vr *varResolver) scanInsideVariable() (lastChar uint8, cursor int) {
	cursor = vr.cursor
	lastChar = 0
	for cursor < len(vr.source) {
		c := vr.source[cursor]
		if c == ')' || c == ':' {
			return uint8(c), cursor // Return ')' or ':' and the position
		} else if c == '$' && (cursor+1) < len(vr.source) && vr.source[cursor+1] == '(' {
			return uint8(c), cursor // Return '$' and the position
		}
		cursor++
	}
	return lastChar, cursor // No variable found, return 0 and the current cursor
}

// ScanOption scans the next option in the variable, returning the option character, and the index of
// the option parameter (if any) or -1.
func (vr *varResolver) scanOption() (option uint8, param int, cursor int) {
	cursor = vr.cursor
	if vr.source[cursor] == ')' {
		return 0, -1, cursor // No option found, return 0 and -1 for the parameter
	}
	if vr.source[cursor] == ':' {
		cursor += 1
	}

	option = uint8(vr.source[cursor]) // The option character
	param = -1                        // The start of the option parameter, if any
	cursor++                          // Move to the next character

	// Scan until we find a ')' or ':' to determine the end of the option and
	// the start-end of the parameter.
	param = cursor
	for cursor < len(vr.source) {
		c := vr.source[cursor]
		if c == ')' || c == ':' {
			break
		}
		cursor++

		// The '\' character is used to tell our parser that the next character should not
		// be interpreted as a special character, so we skip it
		if c == '\\' {
			if cursor < len(vr.source) {
				cursor++ // Skip the next character
			}
		}
	}

	return option, param, cursor
}

type varPartIndex int16
type varPartType int8
type varPartParam int8

const (
	PartTypeNone      varPartType = iota // 0 =
	PartTypeText                         // 1 = text
	PartTypeObsolete                     // 2 = obsolete, not used anymore
	PartTypeVarBegin                     // 3 = variable name or part of it
	PartTypeVarPart                      // 4 = variable name or part of it
	PartTypeVarNested                    // 6 = variable name or part of it
	PartTypeVarEnd                       // 5 = variable name or part of it
	PartTypeOption                       // 7 = option (e.g. f, b, etc.)
)

type varPart struct {
	partType  varPartType  // 0 = text, 1 = value, 2 = string, 3 = node, 4 = option, 5 = option parameter
	partIndex varPartIndex // index in the text, values, strings, options or nodes
}

type varNode struct {
	parts []varPart
}

func newVarNode() varNode {
	return varNode{
		parts: make([]varPart, 0, 8), // preallocate space for parts
	}
}

func (vr *varResolver) parse() int {

	for vr.cursor < len(vr.source) {
		startVar := vr.scanForVariable()

		// Do we need to register any 'PartTypeText' for the current node
		if startVar > vr.cursor {
			vr.nodes[vr.current].parts = append(vr.nodes[vr.current].parts, varPart{
				partType:  PartTypeText,
				partIndex: varPartIndex(len(vr.strings)),
			})
			vr.strings = append(vr.strings, varText{from: int16(vr.cursor), to: int16(startVar)})
		}

		startVar += 2
		if startVar < len(vr.source) {
			vr.cursor = startVar

		continue_parsing_inside_variable:

			lastChar, cursor := vr.scanInsideVariable()
			if lastChar == '$' {
				// We reached a nested variable
				if startVar >= 0 {
					strIndex := -1
					if startVar < cursor {
						strIndex = len(vr.strings)
						vr.strings = append(vr.strings, varText{int16(startVar), int16(cursor)})
					}
					vr.nodes[vr.current].parts = append(vr.nodes[vr.current].parts, varPart{
						partType:  PartTypeVarBegin,
						partIndex: varPartIndex(strIndex),
					})
					startVar = -1
				} else if cursor > vr.cursor {
					strIndex := -1
					if vr.cursor < cursor {
						strIndex = len(vr.strings)
						vr.strings = append(vr.strings, varText{int16(vr.cursor), int16(cursor)})
					}
					vr.nodes[vr.current].parts = append(vr.nodes[vr.current].parts, varPart{
						partType:  PartTypeVarPart,
						partIndex: varPartIndex(strIndex),
					})
				}
				vr.cursor = cursor + 2 // Move the cursor to after the '($'

				// Current node needs a 'nested' part to be added
				vr.nodes[vr.current].parts = append(vr.nodes[vr.current].parts, varPart{
					partType:  PartTypeVarNested,
					partIndex: varPartIndex(len(vr.nodes)),
				})

				// Now we need to create a new node for this new variable
				vr.stack = append(vr.stack, vr.current) // Push current node to stack

				// Nested node needs a PartTypeVarBegin
				startVar = vr.cursor
				vr.current = len(vr.nodes)
				vr.nodes = append(vr.nodes, newVarNode())

				goto continue_parsing_inside_variable
			} else if lastChar == ':' || lastChar == ')' {
				if startVar >= 0 {
					// We haven't registered a PartTypeVarBegin yet
					vr.nodes[vr.current].parts = append(vr.nodes[vr.current].parts, varPart{
						partType:  PartTypeVarBegin,
						partIndex: varPartIndex(len(vr.strings)),
					})
					vr.strings = append(vr.strings, varText{int16(startVar), int16(cursor)})
					startVar = -1 // Reset startVar, as we have registered the variable begin
				} else {
					// We have to register a PartTypeVarPart if we have a variable name part
					if vr.cursor < cursor {
						vr.nodes[vr.current].parts = append(vr.nodes[vr.current].parts, varPart{
							partType:  PartTypeVarPart,
							partIndex: varPartIndex(len(vr.strings)),
						})
						vr.strings = append(vr.strings, varText{int16(vr.cursor), int16(cursor)})
					}
				}

				vr.cursor = cursor

				// Options are applied to the values obtained by using a VariableName which
				// acts as a key in the Vars map.
				if lastChar == ':' {

					vr.nodes[vr.current].parts = append(vr.nodes[vr.current].parts, varPart{
						partType:  PartTypeVarEnd,
						partIndex: varPartIndex(len(vr.options)),
					})

					lastOptionIndex := -1
					for {
						var option uint8
						var param int
						option, param, cursor = vr.scanOption()
						if option == 0 {
							if lastOptionIndex >= 0 {
								vr.options[lastOptionIndex].end = -1 // Mark the last option as the end
							}
							break
						}

						// Do we have an option parameter, if so we need to obtain the index of it
						varOption := newVarOption(option)
						if param >= 0 && param < cursor {
							varOption.param = param
							varOption.paramLen = int16(cursor - param) // Length of the parameter
						}
						varOption.end = int8(len(vr.options))

						// Register the option
						lastOptionIndex = len(vr.options)
						vr.options = append(vr.options, varOption)
						vr.cursor = cursor
					}
				} else {
					vr.nodes[vr.current].parts = append(vr.nodes[vr.current].parts, varPart{
						partType:  PartTypeVarEnd,
						partIndex: varPartIndex(-1),
					})
				}

				vr.cursor = cursor + 1 // Skip ')'

				// Pop a node from the stack, to continue parsing inside the parent variable
				if len(vr.stack) > 0 {
					vr.current = vr.stack[len(vr.stack)-1]
					vr.stack = vr.stack[:len(vr.stack)-1]
					goto continue_parsing_inside_variable
				}
				// So the stack is empty, this means that we should be back to the main/root node, and
				// we should go back to the top of this loop to start scanning for a variable.
			}
		}
	}

	return len(vr.nodes)
}

func (vr *varResolver) resolveNode(vars *Vars, node int) []string {

	// Recursively resolve

	// When we resolve a variable to its value, the value
	// can again contain text with variables.

	// A node holds an array of parts, each part can be:
	// - text: a string of text
	// - name: a variable name or is a part of it
	// - value: a variable value, which can be a string or a list of strings
	// - node: a nested variable, which can have its own parts etc..
	// - option: an option for the variable, e.g. 'f', 'b', etc.
	// - option parameter: a parameter for the option, e.g. ':fparam', ':bparam', etc.

	if node < 0 || node >= len(vr.nodes) {
		return nil // Invalid node index, return empty slice
	}

	parts := vr.nodes[node].parts
	if len(parts) == 0 {
		return nil // No parts, return empty slice
	}

	results := make([]string, 0, 8) // results of the resolution
	results = append(results, "")   // Start with an empty result

	variableName := make([]string, 1, 16)
	variableName[0] = ""

	for pi, part := range parts {
		if part.partType == PartTypeNone {
			continue
		}

		switch part.partType {
		case PartTypeText:
			// Append the text to all results
			str := vr.strings[part.partIndex]
			text := string(vr.source[str.from:str.to])
			for i := range results {
				results[i] += text
			}
			parts[pi].partType = PartTypeNone // Mark as processed
		case PartTypeVarPart:
			// Append the part of the variable name to the variableName
			for vi, vn := range variableName {
				str := vr.strings[part.partIndex]
				variableName[vi] = vn + string(vr.source[str.from:str.to])
			}
			parts[pi].partType = PartTypeNone
		case PartTypeVarBegin:
			// Prepare a new variable name
			variableName = variableName[:1]
			if part.partIndex >= 0 {
				str := vr.strings[part.partIndex]
				variableName[0] = string(vr.source[str.from:str.to])
			} else {
				variableName[0] = "" // Start with an empty variable name
			}
			parts[pi].partType = PartTypeNone // Mark as processed

		case PartTypeVarEnd:
			// We now have a complete variable name, so we use it as a key
			// to get the values from the Vars map
			for _, vn := range variableName {
				values := vars.GetAll(vn)
				if len(values) == 0 {
					continue
				}

				// Apply options to the values
				options := part.partIndex
				if options >= 0 {
					// We are going to mutate the values, so we clone them first
					values = slices.Clone(values)
					for o := options; o < varPartIndex(len(vr.options)); o++ {
						varOption := vr.options[o]
						optionParam := ""
						if varOption.param >= 0 {
							optionParam = string(vr.source[varOption.param : varOption.param+int(varOption.paramLen)])
							optionParam = strings.ReplaceAll(optionParam, "\\", "")
						}

						if action, ok := varOptionActionMap[varOption.option]; ok {
							values = action(values, optionParam)
						}

						if varOption.end == -1 {
							break
						}
					}
				}
				if len(values) == 1 {
					// If we have only one value, we can just append it to all results
					for ri, _ := range results {
						results[ri] += values[0]
					}
				} else {
					// If we have multiple values, we need to multiply the results by the number of values
					// This means we need to create a new slice of results and append each value
					newResults := make([]string, len(results)*len(values))
					nri := 0
					for _, result := range results {
						for _, value := range values {
							newResults[nri] = result + value
							nri++
						}
					}
					results = newResults
				}
			}

			parts[pi].partType = PartTypeNone // Mark as processed
		case PartTypeVarNested:
			// Resolve the nested variable and append it to all results
			nestedResults := vr.resolveNode(vars, int(part.partIndex))

			if len(nestedResults) == 1 {
				for ri, _ := range variableName {
					variableName[ri] += nestedResults[0]
				}
			} else if len(nestedResults) > 1 {
				newVariableName := make([]string, len(results)*len(nestedResults))
				nri := 0
				for _, vn := range variableName {
					for _, nestedResult := range nestedResults {
						newVariableName[nri] = vn + nestedResult
						nri++
					}
				}
				variableName = newVariableName
			}
			parts[pi].partType = PartTypeNone // Mark as processed
		}
	}

	return results // Return the list of results
}
