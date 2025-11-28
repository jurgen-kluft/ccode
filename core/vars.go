package corepkg

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// You can configure the variable pattern, e.g.:
// $(VARIABLE:option1:option2:...)
// {VARIABLE:option1:option2:...}

type Vars struct {
	Values     [][]string
	Keys       []string
	Format     VarsFormat
	KeyToIndex map[string]int
	Resolver   VarResolver
}

func (v *Vars) Clear() {
	v.Values = v.Values[:0]
	v.Keys = v.Keys[:0]
	v.KeyToIndex = map[string]int{}
}

func (v *Vars) Copy() *Vars {
	newVars := &Vars{
		Values:     make([][]string, len(v.Values)),
		Keys:       make([]string, len(v.Keys)),
		Format:     v.Format,
		KeyToIndex: make(map[string]int, len(v.KeyToIndex)),
		Resolver:   NewVarResolver(v.Format),
	}
	for i, vals := range v.Values {
		newVars.Values[i] = make([]string, len(vals))
		copy(newVars.Values[i], vals)
	}
	for i, key := range v.Keys {
		newVars.Keys[i] = key
		newVars.KeyToIndex[key] = i
	}
	return newVars
}

func (v *Vars) DecodeJson(decoder *JsonDecoder) {
	fields := map[string]JsonDecode{
		"keys":   func(decoder *JsonDecoder) { v.Keys = decoder.DecodeStringArray() },
		"values": func(decoder *JsonDecoder) { v.Values = decoder.DecodeStringArray2D() },
		"format": func(decoder *JsonDecoder) {
			formatStr := decoder.DecodeString()
			v.Format, _ = VarsFormatFromString(formatStr)
		},
	}
	decoder.Decode(fields)

	v.KeyToIndex = make(map[string]int, len(v.Keys))
	for i, key := range v.Keys {
		key = strings.ToLower(key)
		v.KeyToIndex[key] = i
	}
}

func (v *Vars) EncodeJson(fieldName string, encoder *JsonEncoder) {
	encoder.BeginObject(fieldName)
	{
		encoder.WriteFieldString("format", v.Format.String("{}"))
		if len(v.Keys) > 0 {
			encoder.WriteStringArray("keys", v.Keys)
			encoder.WriteStringArray2D("values", v.Values)
		}
	}
	encoder.EndObject()
}

func (v *Vars) String() string {
	sb := NewStringBuilder()
	for i, key := range v.Keys {
		value := v.Values[i]
		sb.WriteString(key)
		sb.WriteString(" := \n")
		//sb.WriteString(strings.Join(value, "\n"))
		for _, val := range value {
			sb.WriteString("    ")
			sb.WriteString(val)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

type VarsFormat int32

// int32('$')<<16 | int32('(')<<8 | int32(')')
// int32('{')<<8 | int32('}')
const (
	VarsFormatDollarParenthesis VarsFormat = 0x242829 // int32('$')<<16 | int32('(')<<8 | int32(')')
	VarsFormatCurlyBraces       VarsFormat = 0x007B7D // int32('{')<<8 | int32('}'
)

func (f VarsFormat) String(format string) string {
	switch f {
	case VarsFormatCurlyBraces:
		return "{}"
	case VarsFormatDollarParenthesis:
		return "$()"
	}
	return "{}"
}

func VarsFormatFromString(format string) (VarsFormat, error) {
	switch strings.ToLower(format) {
	case "{}":
		return VarsFormatCurlyBraces, nil
	case "$()":
		return VarsFormatDollarParenthesis, nil
	}
	return VarsFormatCurlyBraces, fmt.Errorf("unknown vars format: %s", format)
}

func NewVars(format VarsFormat) *Vars {
	vars := &Vars{
		Values:     make([][]string, 0, 4),
		Keys:       make([]string, 0, 4),
		Format:     format,
		KeyToIndex: make(map[string]int, 4),
		Resolver:   NewVarResolver(format),
	}
	return vars
}

func (v *Vars) ConvertToMap() map[string]string {
	simpleMap := make(map[string]string, len(v.Keys))
	for i, key := range v.Keys {
		values := v.Values[i]
		if len(values) > 0 {
			simpleMap[key] = values[0]
		}
	}
	return simpleMap
}

func (v *Vars) Join(other *Vars) {
	for i, key := range other.Keys {
		if _, ok := v.KeyToIndex[key]; !ok {
			v.KeyToIndex[key] = len(v.Values)
			v.Keys = append(v.Keys, key)
			v.Values = append(v.Values, other.Values[i])
		}
	}
}

func (v *Vars) JoinMap(other map[string]string) {
	for key, value := range other {
		key = strings.ToLower(key)
		if _, ok := v.KeyToIndex[key]; !ok {
			v.KeyToIndex[key] = len(v.Values)
			v.Keys = append(v.Keys, key)
			v.Values = append(v.Values, []string{value})
		}
	}
}

type SortKeyAndValue struct {
	Keys   []string
	Values [][]string
}

// Implement the sort.Interface for SortKeyAndValue
func (s SortKeyAndValue) Len() int {
	return len(s.Keys)
}

func (s SortKeyAndValue) Less(i, j int) bool {
	return strings.Compare(s.Keys[i], s.Keys[j]) < 0
}

func (s SortKeyAndValue) Swap(i, j int) {
	s.Keys[i], s.Keys[j] = s.Keys[j], s.Keys[i]
	s.Values[i], s.Values[j] = s.Values[j], s.Values[i]
}

func (v *Vars) SortByKey() {
	sort.Sort(SortKeyAndValue{Keys: v.Keys, Values: v.Values})
	newMap := make(map[string]int, len(v.KeyToIndex))
	for i, key := range v.Keys {
		newMap[key] = i
	}
	v.KeyToIndex = newMap
}

func (v *Vars) Set(key string, value ...string) {
	key = strings.ToLower(key)
	if i, ok := v.KeyToIndex[key]; !ok {
		v.KeyToIndex[key] = len(v.Values)
		v.Values = append(v.Values, value)
		v.Keys = append(v.Keys, key)
	} else {
		v.Values[i] = value
	}
}

func (v *Vars) SetMany(vars map[string][]string) {
	for key, value := range vars {
		key = strings.ToLower(key)
		if i, ok := v.KeyToIndex[key]; !ok {
			v.KeyToIndex[key] = len(v.Values)
			v.Values = append(v.Values, value)
			v.Keys = append(v.Keys, key)
		} else {
			v.Values[i] = value
		}
	}
}

func (v *Vars) Append(key string, value ...string) {
	key = strings.ToLower(key)
	if i, ok := v.KeyToIndex[key]; !ok {
		v.KeyToIndex[key] = len(v.Values)
		v.Values = append(v.Values, value)
		v.Keys = append(v.Keys, key)
	} else {
		v.Values[i] = append(v.Values[i], value...)
	}
}

func (v *Vars) Prepend(key string, value ...string) {
	key = strings.ToLower(key)
	if i, ok := v.KeyToIndex[key]; !ok {
		v.KeyToIndex[key] = len(v.Values)
		v.Values = append(v.Values, value)
		v.Keys = append(v.Keys, key)
	} else {
		v.Values[i] = append(value, v.Values[i]...)
	}
}

func (v *Vars) GetFirstOrEmpty(key string) string {
	key = strings.ToLower(key)
	if i, ok := v.KeyToIndex[key]; ok {
		values := v.Values[i]
		if len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func (v *Vars) GetFirst(key string) (string, bool) {
	key = strings.ToLower(key)
	if i, ok := v.KeyToIndex[key]; ok {
		values := v.Values[i]
		if len(values) > 0 {
			return values[0], true
		}
	}
	return "", false
}

func (v *Vars) Has(key string) bool {
	key = strings.ToLower(key)
	_, ok := v.KeyToIndex[key]
	return ok
}

func (v *Vars) Get(key string) ([]string, bool) {
	key = strings.ToLower(key)
	if i, ok := v.KeyToIndex[key]; ok {
		return v.Values[i], true
	}
	return []string{}, false
}

// Cull removes variables that have no values or nil
func (v *Vars) Cull() {
	newValues := make([][]string, 0, len(v.Values))
	newKeys := make([]string, 0, len(v.Keys))
	newKeyMap := make(map[string]int, len(v.KeyToIndex))
	for key, i := range v.KeyToIndex {
		if len(v.Values[i]) > 0 {
			newKeyMap[key] = len(newValues)
			newValues = append(newValues, v.Values[i])
			newKeys = append(newKeys, key)
		}
	}
	v.Values = newValues
	v.Keys = newKeys
	v.KeyToIndex = newKeyMap
}

func (v *Vars) Resolve() {
	for ki := range v.Values {
		newValues := make([]string, 0, len(v.Values[ki]))
		for _, value := range v.Values[ki] {
			_, resolvedValues := v.Resolver.Resolve(value, v)
			newValues = append(newValues, resolvedValues...)
		}
		v.Values[ki] = newValues
	}
}

func (v *Vars) ResolveString(str string, sep string, vars ...*Vars) string {
	_, resolvedValues := v.Resolver.Resolve(str, v, vars...)
	if len(resolvedValues) == 0 {
		return str
	}
	return strings.Join(resolvedValues, sep)
}

func (v *Vars) ResolveArray(strs []string, vars ...*Vars) []string {
	resolved := make([]string, 0, len(strs))
	for _, str := range strs {
		_, resolvedValues := v.Resolver.Resolve(str, v, vars...)
		resolved = append(resolved, resolvedValues...)
	}
	return resolved
}

func (v *Vars) FinalResolve(vars ...*Vars) {
	for ki := range v.Values {
		for true {
			resolvedCountTotal := 0
			newValues := make([]string, 0, len(v.Values[ki]))
			for _, value := range v.Values[ki] {
				resolvedCount, resolvedValues := v.Resolver.FinalResolve(value, v, vars...)
				newValues = append(newValues, resolvedValues...)
				resolvedCountTotal += resolvedCount
			}
			v.Values[ki] = newValues
			if resolvedCountTotal == 0 {
				break
			}
		}
	}
}

func (v *Vars) FinalResolveString(str string, sep string, vars ...*Vars) string {
	var resolvedCount int
	var resolvedValues []string
	resolvedCount = 1
	for resolvedCount > 0 {
		resolvedCount, resolvedValues = v.Resolver.FinalResolve(str, v, vars...)
		if resolvedCount > 0 && len(resolvedValues) > 0 {
			str = strings.Join(resolvedValues, sep)
		}
	}
	return str

}

func (v *Vars) FinalResolveArray(strs []string, vars ...*Vars) (result []string) {
	for true {
		resolvedCountTotal := 0
		resolved := make([]string, 0, len(strs))
		for _, str := range strs {
			resolvedNodes, resolvedValues := v.Resolver.FinalResolve(str, v, vars...)
			resolvedCountTotal += resolvedNodes
			resolved = append(resolved, resolvedValues...)
		}
		if resolvedCountTotal == 0 {
			result = resolved
			break
		}
		strs = resolved
	}
	return result
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

// func actionDelimitValue(values []string, param string) []string {
// 	for i, value := range values {
// 		values[i] = param + value + param
// 	}
// 	return values
// }

// func actionTrimValue(values []string, param string) []string {
// 	for i, value := range values {
// 		if strings.HasPrefix(value, param) {
// 			values[i] = value[len(param):]
// 		}
// 		if strings.HasSuffix(value, param) {
// 			values[i] = value[:len(value)-len(param)]
// 		}
// 	}
// 	return values
// }

// func actionTrimValueAny(values []string, param string) []string {
// 	for i, value := range values {
// 		runes := []rune(value)
// 		// Trim any character from the start
// 		for len(runes) > 0 {
// 			if strings.ContainsRune(param, runes[0]) {
// 				runes = runes[1:]
// 			} else {
// 				break
// 			}
// 		}
// 		// Trim any character from the end
// 		for len(runes) > 0 {
// 			if strings.ContainsRune(param, runes[len(runes)-1]) {
// 				runes = runes[:len(runes)-1]
// 			} else {
// 				break // No more characters to remove
// 			}
// 		}
// 		values[i] = string(runes)
// 	}
// 	return values
// }

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
	Resolve(text string, vars *Vars, varsArray ...*Vars) (int, []string)
	FinalResolve(text string, vars *Vars, varsArray ...*Vars) (int, []string)
}

// ---------------------------------------------------------------------------------------
type varResolver struct {
	strings           []varText   // list of strings
	options           []varOption // list of options, each option is a single character
	nodes             []varNode   // list of nodes
	source            []byte      // source text that we are parsing
	cursor            int         // current cursor in the text
	current           int         // current node
	stack             []int       // stack of nodes
	varOpenMask       uint16      // The read mask
	varOpen           uint16      // The varOpen(open) set, e.g. '$(' or '{'
	varClose          uint8       // The closing bracket, e.g. ')' or '}'
	varKeepUnresolved bool        // If != 0, keep unresolved variables as-is
}

func NewVarResolver(format VarsFormat) VarResolver {
	openMask := uint16(0xFFFF)
	if format&0x00FF0000 == 0 {
		openMask = 0x00FF
	}

	vr := &varResolver{
		strings:           make([]varText, 0, 16),
		options:           make([]varOption, 0, 16),
		nodes:             make([]varNode, 0, 32),
		source:            nil,
		cursor:            0,
		current:           0,
		stack:             make([]int, 0, 4),
		varOpen:           uint16(format >> 8),
		varClose:          uint8(format & 0xFF),
		varOpenMask:       openMask,
		varKeepUnresolved: false,
	}

	vr.reset()
	return vr
}

func (vr *varResolver) reset() {
	vr.strings = vr.strings[:0]
	vr.options = vr.options[:0]
	vr.nodes = vr.nodes[:1]                        // Reset the nodes, we will parse again
	vr.nodes[0] = newVarNode(PartTypeNone, -1, -1) // Reset the root node
	vr.source = []byte{}
	vr.cursor = 0
	vr.current = 0
	vr.stack = vr.stack[:0]
	vr.varKeepUnresolved = true
}

func (vr *varResolver) internalResolve(text string, vars *Vars, varsArray []*Vars, keepUnresolved bool) (numberOfNodes int, result []string) {
	vr.source = []byte(text)
	vr.varKeepUnresolved = keepUnresolved

	numberOfNodes = vr.parse()
	if numberOfNodes > 1 {
		result = vr.resolveNode(vars, varsArray, 0)
	} else {
		result = []string{text} // No variables found, return the original text
	}

	vr.reset()
	return numberOfNodes - 1, result
}

func (vr *varResolver) Resolve(text string, vars *Vars, varsArray ...*Vars) (int, []string) {
	return vr.internalResolve(text, vars, varsArray, true)
}

func (vr *varResolver) FinalResolve(text string, vars *Vars, varsArray ...*Vars) (int, []string) {
	return vr.internalResolve(text, vars, varsArray, false)
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
	// varOpen can be 1 or 2 characters, so we need to compare as well as adjust the cursor accordingly
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

func (vr *varResolver) resolveNode(vars *Vars, varsArray []*Vars, node int) []string {
	if node < 0 || node >= len(vr.nodes) {
		return []string{} // Invalid node index, return empty slice
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
			// to get the values from each Vars in varsArray and the main Vars
			for _, vn := range variableName {
				values := make([]string, 0, 8)
				for _, va := range varsArray {
					if va == nil {
						continue
					}
					if v, ok := va.Get(vn); ok {
						values = append(values, v...)
					}
				}
				if vars != nil {
					if v, ok := vars.Get(vn); ok {
						values = append(values, v...)
					}
				}

				if len(values) == 0 {
					// If we have no values, depending on the configuration we either use an empty string
					// or we keep this variable unresolved. We need to keep the full entry, including
					// the open and close brackets and any options.
					if !vr.varKeepUnresolved {
						continue // No values
					}
					if part.partIndex < 0 {
						if vr.varOpenMask == 0xFF {
							values = append(values, fmt.Sprintf("%c%s%c", byte(vr.varOpen), vn, byte(vr.varClose)))
						} else {
							values = append(values, fmt.Sprintf("%c%c%s%c", byte(vr.varOpen>>8), byte(vr.varOpen), vn, byte(vr.varClose)))
						}
					} else {
						value := ""
						if vr.varOpenMask == 0xFF {
							value = fmt.Sprintf("%c%s", byte(vr.varOpen), vn)
						} else {
							value = fmt.Sprintf("%c%c%s", byte(vr.varOpen>>8), byte(vr.varOpen), vn)
						}
						options := part.partIndex
						if options >= 0 {
							for o := options; o < varPartIndex(len(vr.options)); o++ {
								varOption := vr.options[o]
								if varOption.paramLen > 0 {
									value = fmt.Sprintf("%s:%c%s", value, varOption.option, string(vr.source[varOption.param:varOption.param+int(varOption.paramLen)]))
								} else {
									value = fmt.Sprintf("%s:%c", value, varOption.option)
								}
								if varOption.end == -1 {
									break
								}
							}
						}
						value += fmt.Sprintf("%c", byte(vr.varClose))
					}
				} else {
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
				}

				if len(values) == 1 {
					for ri := range results {
						results[ri] += values[0]
					}
				} else {
					// - if len(results) == 1 and len(results[0]) == 0, we can just set results to values
					// - if len(results) == 1 and len(results[0]) > 0, we append the joined values to results[0]
					// - if len(results) > 1, we need to create a new results array with all combinations
					if len(results) == 1 && len(results[0]) == 0 {
						results = values
					} else if len(results) == 1 {
						joinedValues := strings.Join(values, " ")
						for ri := range results {
							results[ri] += joinedValues
						}
					} else {
						newResults := make([]string, len(results)*len(values))
						nri := 0
						for ri := range results {
							for _, value := range values {
								newResults[nri] = results[ri] + value
								nri++
							}
						}
						results = newResults
					}
				}
			}
		case PartTypeVarNested:
			// Resolve the nested variable and append it to all results
			nestedResults := vr.resolveNode(vars, varsArray, int(part.partIndex))

			if len(nestedResults) == 1 {
				for ri := range variableName {
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
