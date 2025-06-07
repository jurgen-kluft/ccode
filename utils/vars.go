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
							fmt.Printf("Unknown interpolation option '%v' as part of $(%s:%s)\n", option, string(runes[p.bvar:p.evar]), string(runes[p.bopt:p.eopt]))
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

func NewVarResolver() *varResolver {
	return &varResolver{
		text:    make([]string, 0, 16),
		values:  make([][]string, 0, 16),
		strings: make([]string, 0, 16),
		options: make([]uint8, 0, 16),
		nodes:   make([]varNode, 0, 16),
	}
}

func (vr *varResolver) Parse(text string) int {
	ctx := &varParseContext{
		text:    []rune(text),      // convert the text to runes for indexed access
		cursor:  0,                 // current cursor in the text
		current: 0,                 // current node index in the nodes slice
		stack:   make([]int, 0, 8), // stack of nodes for nested variables
	}
	vr.reset()
	vr.nodes = append(vr.nodes, newVarNode()) // Start with the root node
	return vr.parse(ctx)
}

func (vr *varResolver) Resolve(vars *Vars) []string {
	return vr.resolveNode(vars, 0)
}

// ---------------------------------------------------------------------------------------
type varResolver struct {
	text    []string   // pure (resolved) text
	values  [][]string // list of values
	strings []string   // list of strings
	options []uint8    // list of options
	nodes   []varNode  // list of nodes
}

func (vr *varResolver) reset() {
	vr.text = vr.text[:0]
	vr.values = vr.values[:0]
	vr.strings = vr.strings[:0]
	vr.options = vr.options[:0]
	vr.nodes = vr.nodes[:0] // Reset the nodes, we will parse again
}

type varParseContext struct {
	text    []rune
	cursor  int   // current cursor in the text
	current int   // current node
	stack   []int // stack of nodes
}

func (ctx *varParseContext) scanForVariable() int {
	cursor := ctx.cursor
	for cursor < len(ctx.text) {
		if ctx.text[cursor] == '$' && (cursor+1) < len(ctx.text) && ctx.text[cursor+1] == '(' {
			return cursor
		}
		cursor++
	}
	return cursor
}

func (ctx *varParseContext) scanInsideVariable() (lastChar uint8, cursor int) {
	cursor = ctx.cursor
	lastChar = 0
	for cursor < len(ctx.text) {
		c := ctx.text[cursor]
		if c == ')' || c == ':' {
			return uint8(c), cursor // Return ')' or ':' and the position
		} else if c == '$' && (cursor+1) < len(ctx.text) && ctx.text[cursor+1] == '(' {
			return uint8(c), cursor // Return '$' and the position
		}
		cursor++
	}
	return lastChar, cursor // No variable found, return 0 and the current cursor
}

// ScanOption scans the next option in the variable, returning the option character, and the index of
// the option parameter (if any) or -1.
func (vr *varResolver) scanOption(ctx *varParseContext) (option uint8, param int, cursor int) {
	cursor = ctx.cursor

	if ctx.text[cursor] == ')' {
		return 0, -1, cursor // No option found, return 0 and -1 for the parameter
	}

	option = uint8(ctx.text[cursor]) // The option character
	param = -1                       // The index of the option parameter, if any
	cursor++                         // Move to the next character

	// Scan until we find a ')' or ':' to determine the end of the option and
	// the start-end of the parameter.
	paramStart := cursor
	for cursor < len(ctx.text) {

		c := ctx.text[cursor]

		// The '\' character is used to tell our parser that the next character should not
		// be interpreted as a special character, so we skip it
		if c == '\\' {
			cursor++ // Skip the '\' character
			if cursor < len(ctx.text) {
				cursor++ // Skip the next character
			}
			continue
		}

		if c == ')' || c == ':' {
			if cursor > paramStart {
				param = len(vr.strings) // Store the parameter index
				vr.strings = append(vr.strings, string(ctx.text[paramStart:cursor]))
			}
			break
		}
		cursor++
	}

	return option, param, cursor
}

type varPartIndex int16
type varPartType int16

const (
	PartTypeNone        varPartType = iota // 0 =
	PartTypeText                           // 1 = text
	PartTypeValue                          // 2 = value
	PartTypeName                           // 3 = variable name or part of it
	PartTypeNode                           // 4 = node
	PartTypeOption                         // 5 = option (e.g. f, b, etc.)
	PartTypeOptionParam                    // 6 = option parameter (e.g. :fparam, :bparam, etc.)
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

func (vr *varResolver) parse(ctx *varParseContext) int {

	for ctx.cursor < len(ctx.text) {
		start := ctx.scanForVariable()

		// Do we need to register any 'PartTypeText' for the current node
		if start > ctx.cursor {
			vr.nodes[ctx.current].parts = append(vr.nodes[ctx.current].parts, varPart{
				partType:  PartTypeText,
				partIndex: varPartIndex(len(vr.text)),
			})
			vr.text = append(vr.text, string(ctx.text[ctx.cursor:start]))
		}

		start += 2
		if start < len(ctx.text) {
			ctx.cursor = start

		continue_parsing_inside_variable:

			lastChar, cursor := ctx.scanInsideVariable()
			if lastChar == ')' {
				// We reached the end of a variable
				if cursor > ctx.cursor {
					variableName := string(ctx.text[ctx.cursor:cursor])
					vr.nodes[ctx.current].parts = append(vr.nodes[ctx.current].parts, varPart{
						partType:  PartTypeName,
						partIndex: varPartIndex(len(vr.strings)),
					})
					vr.strings = append(vr.strings, variableName)
				}
				ctx.cursor = cursor + 1 // Move the cursor to the end of the variable, right after ')'

				// Pop a node from the stack, to continue parsing inside the parent variable
				if len(ctx.stack) > 0 {
					ctx.current = ctx.stack[len(ctx.stack)-1]
					ctx.stack = ctx.stack[:len(ctx.stack)-1]
					goto continue_parsing_inside_variable
				}
				// So the stack is empty, this means that we should be back to the main/root node, and
				// we should go back to the top of this loop to start scanning for a variable.
			} else if lastChar == '$' {
				// We reached another variable
				if cursor > ctx.cursor {
					variableName := string(ctx.text[ctx.cursor:cursor])
					vr.nodes[ctx.current].parts = append(vr.nodes[ctx.current].parts, varPart{
						partType:  PartTypeName,
						partIndex: varPartIndex(len(vr.strings)),
					})
					vr.strings = append(vr.strings, variableName)
				}
				ctx.cursor = cursor + 2 // Move the cursor to after the '($'

				// Now we need to create a new node for this new variable
				ctx.stack = append(ctx.stack, ctx.current) // Push current node to stack

				// Current node needs a 'node' part to be added
				vr.nodes[ctx.current].parts = append(vr.nodes[ctx.current].parts, varPart{
					partType:  PartTypeNode,
					partIndex: varPartIndex(len(vr.nodes)),
				})

				ctx.current = len(vr.nodes)
				vr.nodes = append(vr.nodes, newVarNode())

				goto continue_parsing_inside_variable
			} else if lastChar == ':' {
				// We reached an option, so we need to register the variable name
				variableName := string(ctx.text[ctx.cursor:cursor]) // -1 to put the end on the ':'
				vr.nodes[ctx.current].parts = append(vr.nodes[ctx.current].parts, varPart{
					partType:  PartTypeName,
					partIndex: varPartIndex(len(vr.strings)),
				})
				vr.strings = append(vr.strings, variableName)

				ctx.cursor = cursor + 1
				for {
					var option uint8
					var param int
					option, param, cursor = vr.scanOption(ctx)
					if option == 0 {
						ctx.cursor = cursor + 1 // Move right after ')'
						break
					}

					// We have an option, so we need to register it
					vr.nodes[ctx.current].parts = append(vr.nodes[ctx.current].parts, varPart{
						partType:  PartTypeOption,
						partIndex: varPartIndex(len(vr.options)),
					})
					vr.options = append(vr.options, option)

					// Do we have an option parameter, if so we need to register it
					if param >= 0 {
						vr.nodes[ctx.current].parts = append(vr.nodes[ctx.current].parts, varPart{
							partType:  PartTypeOptionParam,
							partIndex: varPartIndex(len(vr.strings)),
						})
						vr.strings = append(vr.strings, string(ctx.text[param:cursor]))
					}
					ctx.cursor = cursor // Move the cursor to the end of the option
				}

				// Duplicate, any way to logically merge it with the other case ?
				// Pop a node from the stack, to continue parsing inside the parent variable
				if len(ctx.stack) > 0 {
					ctx.current = ctx.stack[len(ctx.stack)-1]
					ctx.stack = ctx.stack[:len(ctx.stack)-1]
					goto continue_parsing_inside_variable
				}
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

	resolved := len(parts)
	for resolved > 0 {
		currentResolved := resolved // Keep track of how many parts we are resolving in this iteration
		for pi, part := range parts {
			switch part.partType {
			case PartTypeText:
				// Append the text to all results
				text := vr.text[part.partIndex]
				for i := range results {
					results[i] += text
				}
				parts[pi].partType = PartTypeNone // Mark as processed
				resolved--                        // Decrease the resolved count
			case PartTypeValue:
				// Append the value to all results
				values := vr.values[part.partIndex]
				if len(values) == 0 {
					continue
				}
				newResults := make([]string, 0, len(results)*len(values))
				for _, result := range results {
					for _, value := range values {
						newResults = append(newResults, result+value)
					}
				}
				results = newResults
				parts[pi].partType = PartTypeNone // Mark as processed
				resolved--                        // Decrease the resolved count
			case PartTypeName:
				// Resolve the variable name and append it to all results
				variableName := vr.strings[part.partIndex]
				values := vars.GetAll(variableName)
				if len(values) == 0 {
					continue
				}

				newResults := make([]string, len(results)*len(values))
				nri := 0
				for _, result := range results {
					for _, value := range values {
						newResults[nri] = result + value
						nri++
					}
				}
				results = newResults
				parts[pi].partType = PartTypeNone // Mark as processed
				resolved--                        // Decrease the resolved count
			case PartTypeNode:
				// Resolve the nested variable and append it to all results
				nestedResults := vr.resolveNode(vars, int(part.partIndex))
				if len(nestedResults) == 0 {
					parts[pi].partType = PartTypeNone // Mark as processed
					continue                          // No nested results, skip
				}

				// We should change this part into a 'value' part, where the index
				// is pointing to vr.values
				parts[pi].partType = PartTypeValue                 // Change to value type
				parts[pi].partIndex = varPartIndex(len(vr.values)) // Set the index to the new values
				vr.values = append(vr.values, nestedResults)       // Add the nested results to the values

			case PartTypeOption:
				option := vr.options[part.partIndex]
				if option == 'j' { // Join values
					if len(results) > 1 {
						results = []string{strings.Join(results, " ")} // Join with space by default
					}
					parts[pi].partType = PartTypeNone // Mark as processed
					resolved--                        // Decrease the resolved count
					continue                          // Skip further processing for 'j' option
				}

				popi := int(part.partIndex + 1) // The next part should be the part option parameter
				param := ""                     // The parameter for the option, if any, default is empty
				if popi < len(parts) && parts[popi].partType == PartTypeOptionParam {
					param = vr.strings[parts[popi].partIndex]
					parts[popi].partType = PartTypeNone // Mark as processed
				}
				for ri, result := range results {
					switch option {
					case 'f':
						results[ri] = actionForwardSlashes(result)
					case 'b':
						results[ri] = actionBackwardSlashes(result)
					case 'n':
						results[ri] = actionNativeSlashes(result)
					case 'u':
						results[ri] = actionUpperCase(result)
					case 'l':
						results[ri] = actionLowerCase(result)
					case 'B':
						results[ri] = actionBaseName(result)
					case 'F':
						results[ri] = actionFileName(result)
					case 'D':
						results[ri] = actionDirName(result)
					case 'p':
						results[ri] = actionPrefix(result, param)
					case 's':
						results[ri] = actionSuffix(result, param)
					case 'P':
						results[ri] = actionPrefixIfNotExists(result, param)
					case 'S':
						results[ri] = actionSuffixIfNotExists(result, param)
					default:
						fmt.Printf("Unknown interpolation option '%c' as part of $(%s:%s)\n", option, vr.strings[parts[part.partIndex].partIndex], param)
						results[ri] = "?"
					}
				}
				parts[pi].partType = PartTypeNone // Mark as processed
				resolved--                        // Decrease the resolved count
			case PartTypeOptionParam:
				// This should not happen, as option parameters are always followed by a parameter and are marked as processed
				// in the PartTypeOption case. If we reach here, it means there is an unexpected option parameter.
				fmt.Printf("Unexpected option parameter in variable parts: %s\n", vr.strings[part.partIndex])
				parts[pi].partType = PartTypeNone // Mark as processed
				resolved--                        // Decrease the resolved count
			case PartTypeNone:
				// Do nothing
			}
		}

		// If we did not resolve any parts, we should break the loop to avoid an infinite loop
		if currentResolved == resolved {
			break
		}
	}

	return results // Return the list of results
}
