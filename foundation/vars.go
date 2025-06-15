package foundation

import (
	"os"
	"slices"
	"strconv"
	"strings"
)

// You can configure the variable pattern, e.g.:
// $(VARIABLE:option1:option2:...)
// {VARIABLE:option1:option2:...}

type Vars struct {
	Values [][]string
	Keys   map[string]int
	Leader byte // Leader byte, e.g. '$' or 0 for no varOpen
	Open   byte // Bracket byte, e.g. '(' or '{' for the start of the variable
	Close  byte // Bracket byte, e.g. '(' or '{' for the start of the variable
}

func NewVars() *Vars {
	return &Vars{
		Values: make([][]string, 0, 4),
		Keys:   make(map[string]int, 4),
		Leader: '$', // Default varOpen byte
		Open:   '(', // Default opening bracket
		Close:  ')', // Default closing bracket
	}
}

type VarsFormat int8

const (
	VarsFormatDollarParenthesis VarsFormat = iota
	VarsFormatCurlyBraces
)

func NewVarsCustom(format VarsFormat) *Vars {
	vars := &Vars{
		Values: make([][]string, 0, 4),
		Keys:   make(map[string]int, 4),
		Leader: 0, // Default varOpen byte
		Open:   0, // Default opening bracket
		Close:  0, // Default closing bracket
	}

	if format == VarsFormatDollarParenthesis {
		vars.Leader = '$'
		vars.Open = '('
		vars.Close = ')'
	} else if format == VarsFormatCurlyBraces {
		vars.Leader = 0
		vars.Open = '{'
		vars.Close = '}'
	}

	return vars
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

func (v *Vars) GetFirstOrEmpty(key string) string {
	if i, ok := v.Keys[key]; ok {
		values := v.Values[i]
		if len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func (v *Vars) GetFirst(key string) (string, bool) {
	if i, ok := v.Keys[key]; ok {
		values := v.Values[i]
		if len(values) > 0 {
			return values[0], true
		}
	}
	return "", false
}

func (v *Vars) Get(key string) ([]string, bool) {
	if i, ok := v.Keys[key]; ok {
		return v.Values[i], true
	}
	return []string{}, false
}

// Cull removes variables that have no values or nil
func (v *Vars) Cull() {
	newValues := make([][]string, 0, len(v.Values))
	newKeys := make(map[string]int, len(v.Keys))
	for key, i := range v.Keys {
		if len(v.Values[i]) > 0 {
			newKeys[key] = len(newValues)
			newValues = append(newValues, v.Values[i])
		}
	}
	v.Values = newValues
	v.Keys = newKeys
}

func (v *Vars) Resolve() {
	resolver := NewVarResolver()
	for ki, values := range v.Values {
		newValues := make([]string, 0, len(values))
		for _, value := range values {
			resolvedValues := resolver.Resolve(value, v)
			newValues = append(newValues, resolvedValues...)
		}
		v.Values[ki] = newValues
	}
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

func actionIndexValue(values []string, param string) []string {
	index, _ := strconv.ParseInt(param, 10, 32)
	if index < int64(len(values)) {
		return []string{values[index]}
	}
	return []string{""}
}

var varOptionActionMap = map[int8]func([]string, string) []string{
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
	'i': actionIndexValue,
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
	Resolve(text string, vars *Vars) []string
}

func NewVarResolver() VarResolver {
	vr := &varResolver{
		strings:  make([]varText, 0, 16),
		options:  make([]varOption, 0, 16),
		nodes:    make([]varNode, 0, 32),
		source:   nil,
		cursor:   0,
		current:  0,
		stack:    make([]int, 0, 4),
		varOpen:  uint16('$')<<8 | uint16('('), // Default varOpen+open
		varClose: ')',                          // Default closing
	}
	return vr
}

// ---------------------------------------------------------------------------------------
type varResolver struct {
	strings     []varText   // list of strings
	options     []varOption // list of options, each option is a single character
	nodes       []varNode   // list of nodes
	source      []byte      // source text that we are parsing
	cursor      int         // current cursor in the text
	current     int         // current node
	stack       []int       // stack of nodes
	varOpenMask uint16      // The read mask
	varOpen     uint16      // The varOpen(open) set, e.g. '$(' or '{'
	varClose    uint8       // The closing bracket, e.g. ')' or '}'
}

func (vr *varResolver) Resolve(text string, vars *Vars) []string {
	vr.strings = vr.strings[:0]
	vr.options = vr.options[:0]
	vr.nodes = vr.nodes[:1]                        // Reset the nodes, we will parse again
	vr.nodes[0] = newVarNode(PartTypeNone, -1, -1) // Reset the root node
	vr.source = []byte(text)
	vr.cursor = 0
	vr.current = 0
	vr.stack = vr.stack[:0]
	if vars.Leader == 0 {
		vr.varOpenMask = 0x00FF
		vr.varOpen = uint16(vars.Open) // '(' or '{'
	} else {
		vr.varOpenMask = 0xFFFF
		vr.varOpen = uint16(vars.Leader)<<8 | uint16(vars.Open) // e.g. '$(' or '${'
	}
	vr.varClose = uint8(vars.Close) // e.g. ')' or '}'
	vr.parse()
	return vr.resolveNode(vars, 0)
}

type varText struct {
	from int16
	to   int16 // from and to are indexes in the text slice
}

type varOption struct {
	param    int   // start of the option in varParseContext.text
	paramLen int16 // end of the option in varParseContext.text
	option   int8  // the option character, e.g. 'f', 'b', etc.
	end      int8  // end = 1 means this is the last option
}

func newVarOption(opt int8) varOption {
	return varOption{
		param:    -1, // -1 means no parameter
		paramLen: -1, // -1 means no parameter
		option:   opt,
		end:      0, // 0 means this is not the last option
	}
}

func (vr *varResolver) scanForVariable() int {
	cursor := vr.cursor
	c := uint16(0)
	for cursor < len(vr.source) {
		c = ((c << 8) | uint16(vr.source[cursor])) & vr.varOpenMask
		if vr.varOpen == c {
			return cursor - int((vr.varOpenMask>>8)&1) // Return the position of the variable start
		}
		cursor++
	}
	return cursor
}

func (vr *varResolver) scanInsideVariable() (lastChar uint16, cursor int) {
	cursor = vr.cursor
	lastChar = 0
	for cursor < len(vr.source) {
		read := vr.source[cursor]
		if read == ':' || read == vr.varClose {
			return uint16(read), cursor // Return ':' or ')' and the position
		}
		lastChar = ((lastChar << 8) | uint16(read)) & vr.varOpenMask
		if lastChar == vr.varOpen {
			cursor -= int((vr.varOpenMask >> 8) & 1)
			return
		}
		cursor++
	}
	return 0, cursor // No variable found, return 0 and the current cursor
}

// ScanOption scans the next option in the variable, returning the option character, and the index of
// the option parameter (if any) or -1.
func (vr *varResolver) parseOption() (option varOption) {
	if vr.source[vr.cursor] == vr.varClose {
		return varOption{-1, -1, 0, -1} // No option found, return 0 and -1 for the parameter
	}
	if vr.source[vr.cursor] == ':' {
		vr.cursor += 1
	}

	option.option = int8(vr.source[vr.cursor]) // The option character
	vr.cursor++                                // Move to the next character

	// Scan until we find a ')' or ':' to determine the end of the option and
	// the start-end of the parameter.
	option.param = vr.cursor
	for vr.cursor < len(vr.source) {
		c := vr.source[vr.cursor]
		if c == vr.varClose || c == ':' {
			option.paramLen = int16(vr.cursor - option.param) // Length of the parameter
			if c == vr.varClose {
				option.end = -1
			}
			break
		}
		vr.cursor++

		// The '\' character is used to tell our parser that the next character should not
		// be interpreted as a special character, so we skip it
		if c == '\\' {
			if vr.cursor < len(vr.source) {
				vr.cursor++ // Skip the next character
			}
		}
	}

	return option
}

type varPartIndex int16
type varPartType int16

const (
	PartTypeNone      varPartType = iota // 0 =
	PartTypeText                         // 1 = text
	PartTypeVarBegin                     // 2 = variable name or part of it
	PartTypeVarPart                      // 3 = variable name or part of it
	PartTypeVarNested                    // 4 = variable name or part of it
	PartTypeVarEnd                       // 5 = variable name or part of it
)

type varNode struct {
	partType  varPartType  // 0 = text, 1 = value, 2 = string, 3 = node, 4 = option, 5 = option parameter
	partIndex varPartIndex // index in the text, values, strings, options or nodes
	partNext  varPartIndex //
	partPrev  varPartIndex //
}

func newVarNode(partType varPartType, partIndex varPartIndex, partPrev varPartIndex) varNode {
	return varNode{
		partType:  partType,
		partIndex: partIndex,
		partNext:  -1, // No next part
		partPrev:  partPrev,
	}
}

func (vr *varResolver) addPart(node int, partType varPartType, partIndex varPartIndex) {
	if node >= 0 && node < len(vr.nodes) {
		if vr.nodes[node].partType == PartTypeNone {
			// If this is the first part, set the part type to the new part type
			vr.nodes[node] = newVarNode(partType, partIndex, varPartIndex(node)) // Set the part type and index
		} else {
			// Start adding new parts to the tail
			newPart := len(vr.nodes)                           // New part index
			oldTail := vr.nodes[node].partPrev                 // Old tail
			vr.nodes[oldTail].partNext = varPartIndex(newPart) // Old tail points to new part
			vr.nodes = append(vr.nodes, newVarNode(partType, partIndex, varPartIndex(oldTail)))
			vr.nodes[node].partPrev = varPartIndex(newPart) // Head points to new tail
		}
	}
}

func (vr *varResolver) parse() int {

	for vr.cursor < len(vr.source) {
		startVar := vr.scanForVariable()

		// Do we need to register any 'PartTypeText' for the current node
		if startVar > vr.cursor {
			vr.addPart(vr.current, PartTypeText, varPartIndex(len(vr.strings)))
			vr.strings = append(vr.strings, varText{from: int16(vr.cursor), to: int16(startVar)})
		}
		startVar = startVar + 1 + int((vr.varOpenMask>>8)&1)

		vr.cursor = startVar
		for vr.cursor < len(vr.source) {
			lastChar, cursor := vr.scanInsideVariable()
			if lastChar == vr.varOpen {
				// We reached a nested variable
				if startVar >= 0 {
					strIndex := -1
					if startVar < cursor {
						strIndex = len(vr.strings)
						vr.strings = append(vr.strings, varText{int16(startVar), int16(cursor)})
					}
					vr.addPart(vr.current, PartTypeVarBegin, varPartIndex(strIndex))
					startVar = -1
				} else if cursor > vr.cursor {
					strIndex := -1
					if vr.cursor < cursor {
						strIndex = len(vr.strings)
						vr.strings = append(vr.strings, varText{int16(vr.cursor), int16(cursor)})
					}
					vr.addPart(vr.current, PartTypeVarPart, varPartIndex(strIndex))
				}
				vr.cursor = cursor + 1 + int((vr.varOpenMask>>8)&1)

				// Create the new node here, since the below addPart will create a new node!
				nested := len(vr.nodes)
				vr.nodes = append(vr.nodes, newVarNode(PartTypeNone, -1, -1))

				// Current node needs a 'nested' part to be added and added to the stack
				vr.addPart(vr.current, PartTypeVarNested, varPartIndex(nested))
				vr.stack = append(vr.stack, vr.current)

				// Setup parsing this new nested variable
				startVar = vr.cursor
				vr.current = nested
			} else if lastChar == ':' || lastChar == uint16(vr.varClose) {
				if startVar >= 0 {
					// We haven't registered a PartTypeVarBegin yet
					vr.addPart(vr.current, PartTypeVarBegin, varPartIndex(len(vr.strings)))
					vr.strings = append(vr.strings, varText{int16(startVar), int16(cursor)})
					startVar = -1 // Reset startVar, as we have registered the variable begin
				} else {
					// We have to register a PartTypeVarPart if we have a variable name part
					if vr.cursor < cursor {
						vr.addPart(vr.current, PartTypeVarPart, varPartIndex(len(vr.strings)))
						vr.strings = append(vr.strings, varText{int16(vr.cursor), int16(cursor)})
					}
				}

				vr.cursor = cursor

				// Options are applied to the values obtained by using a VariableName which
				// acts as a key in the Vars map.
				if lastChar == ':' {
					vr.addPart(vr.current, PartTypeVarEnd, varPartIndex(len(vr.options)))
					option := newVarOption(0) // Create a new option
					for option.end == 0 {
						option = vr.parseOption()
						vr.options = append(vr.options, option)
					}
				} else {
					vr.addPart(vr.current, PartTypeVarEnd, varPartIndex(-1)) // No options, just end the variable
				}

				vr.cursor++ // Skip vr.varClose

				// Pop a node from the stack, to continue parsing inside the parent variable
				if len(vr.stack) == 0 {
					// So the stack is empty, this means that we should be back to the main/root node, and
					// we should go back to the top of this loop to start scanning for a variable.
					break
				}
				vr.current = vr.stack[len(vr.stack)-1]
				vr.stack = vr.stack[:len(vr.stack)-1]
			}
		}
	}

	return len(vr.nodes)
}

func (vr *varResolver) resolveNode(vars *Vars, node int) []string {
	if node < 0 || node >= len(vr.nodes) {
		return nil // Invalid node index, return empty slice
	}

	results := make([]string, 0, 8) // results of the resolution
	results = append(results, "")   // Start with an empty result

	variableName := make([]string, 1, 16)
	variableName[0] = ""

	partIter := node
	for partIter != -1 {
		part := vr.nodes[partIter]
		partIter = int(part.partNext)
		switch part.partType {
		case PartTypeNone:
			continue
		case PartTypeText:
			// Append the text to all results
			str := vr.strings[part.partIndex]
			text := string(vr.source[str.from:str.to])
			for i := range results {
				results[i] += text
			}
		case PartTypeVarPart:
			// Append the part of the variable name to the variableName
			for vi, vn := range variableName {
				str := vr.strings[part.partIndex]
				variableName[vi] = vn + string(vr.source[str.from:str.to])
			}
		case PartTypeVarBegin:
			// Prepare a new variable name
			variableName = variableName[:1]
			if part.partIndex >= 0 {
				str := vr.strings[part.partIndex]
				variableName[0] = string(vr.source[str.from:str.to])
			} else {
				variableName[0] = "" // Start with an empty variable name
			}
		case PartTypeVarEnd:
			// We now have a complete variable name, so we use it as a key
			// to get the values from the Vars map
			for _, vn := range variableName {
				values, _ := vars.Get(vn)
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
						if varOption.paramLen > 0 {
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
		}
	}

	return results // Return the list of results
}
