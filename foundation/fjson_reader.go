package foundation

import (
	"strconv"
	"strings"
)

type JsonValueType int16

const (
	JsonValueTypeError  JsonValueType = -1
	JsonValueTypeEmpty  JsonValueType = 0
	JsonValueTypeObject JsonValueType = 1
	JsonValueTypeArray  JsonValueType = 2
	JsonValueTypeString JsonValueType = 3
	JsonValueTypeNumber JsonValueType = 4
	JsonValueTypeBool   JsonValueType = 5
	JsonValueTypeNull   JsonValueType = 6
	JsonValueTypeEnd    JsonValueType = 7
)

// Note: Should map 1:1 to JsonValueType
type JsonResult int16

const (
	JsonResultError  JsonResult = -1
	JsonResultEmpty  JsonResult = 0
	JsonResultObject JsonResult = 1
	JsonResultArray  JsonResult = 2
	JsonResultString JsonResult = 3
	JsonResultNumber JsonResult = 4
	JsonResultBool   JsonResult = 5
	JsonResultNull   JsonResult = 6
	JsonResultEnd    JsonResult = 7
)

type JsonField struct {
	Begin  int
	Length int
	Type   JsonValueType
}

func (f JsonField) IsEmpty() bool {
	return f.Type == JsonValueTypeEmpty || f.Length == 0
}

var JsonFieldEmpty = JsonField{Begin: 0, Length: 0, Type: JsonValueTypeEmpty}
var JsonFieldError = JsonField{Begin: 0, Length: 0, Type: JsonValueTypeError}

func NewJsonField(begin, length int, valueType JsonValueType) JsonField {
	return JsonField{Begin: begin, Length: length, Type: valueType}
}

type JsonContext struct {
	Json           string
	Index          int
	IsEscapeString bool
	Stack          []JsonValueType
	StackIndex     int
	EscapedStrings StringBuilder
}

func NewJsonContext() *JsonContext {
	return &JsonContext{
		Json:           "",
		Index:          0,
		IsEscapeString: false,
		Stack:          make([]JsonValueType, 64),
		StackIndex:     len(make([]JsonValueType, 64)),
		EscapedStrings: StringBuilder{},
	}
}

func (c *JsonContext) IsValidField(f JsonField) bool {
	return f.Begin >= 0 && f.Length > 0 && (f.Begin+f.Length > len(c.Json))
}

type JsonReader struct {
	Context *JsonContext
}

func NewJsonReader() *JsonReader {
	return &JsonReader{Context: NewJsonContext()}
}

func (r *JsonReader) Begin(json string) bool {
	return ParseBegin(NewJsonContext())
}

func (r *JsonReader) FieldStr(f JsonField) string {
	return r.Context.Json[f.Begin : f.Begin+f.Length]
}

func (r *JsonReader) ParseBool(field JsonField) bool {
	if !r.Context.IsValidField(field) {
		return false
	}
	value, err := strconv.ParseBool(string(r.Context.Json[field.Begin : field.Begin+field.Length]))
	if err != nil {
		return false
	}
	return value
}

func (r *JsonReader) ParseFloat(field JsonField) float64 {
	if !r.Context.IsValidField(field) {
		return 0.0
	}

	json := r.Context.Json
	value := string(json[field.Begin : field.Begin+field.Length])
	result, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return 0.0
	}
	return result
}

func (r *JsonReader) ParseInt(field JsonField) int {
	if !r.Context.IsValidField(field) {
		return 0
	}

	json := r.Context.Json
	value := string(json[field.Begin : field.Begin+field.Length])

	intValue, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0
	}
	return int(intValue)
}

func (r *JsonReader) ParseLong(field JsonField) int64 {
	if !r.Context.IsValidField(field) {
		return 0
	}

	json := r.Context.Json
	value := string(json[field.Begin : field.Begin+field.Length])

	longValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return longValue
}

func (r *JsonReader) ParseString(field JsonField) string {
	if r.Context.IsEscapeString {
		return GetEscapedString(r.Context, field)
	} else {
		if !r.Context.IsValidField(field) {
			return ""
		}
		return string(r.Context.Json[field.Begin : field.Begin+field.Length])
	}
}

func (r *JsonReader) IsFieldName(f JsonField, name string) bool {
	if !r.Context.IsValidField(f) {
		return false
	}
	fieldName := string(r.Context.Json[f.Begin : f.Begin+f.Length])
	return strings.EqualFold(fieldName, name)
}

func (r *JsonReader) IsObjectEnd(key JsonField, value JsonField) bool {
	return key.Type == JsonValueTypeObject && value.Type == JsonValueTypeEnd
}

func (r *JsonReader) ReadUntilObjectEnd() (JsonField, JsonField, bool) {
	if key, value, ok := r.Read(); ok {
		return key, value, r.IsObjectEnd(key, value)
	}
	return JsonFieldError, JsonFieldError, false
}

func (r *JsonReader) IsArrayEnd(key JsonField, value JsonField) bool {
	return key.Type == JsonValueTypeArray && value.Type == JsonValueTypeEnd
}

func (r *JsonReader) ReadUntilArrayEnd() (JsonField, JsonField, bool) {
	if key, value, ok := r.Read(); ok {
		return key, value, r.IsArrayEnd(key, value)
	}
	return JsonFieldError, JsonFieldError, false
}

func (r *JsonReader) Read() (key JsonField, value JsonField, ok bool) {
	if r.Context.StackIndex == len(r.Context.Stack) {
		return JsonFieldError, JsonFieldError, false
	}

	state := r.Context.Stack[r.Context.StackIndex]
	switch state {
	case JsonValueTypeObject:
		var result JsonResult
		key, value, result = ParseObjectBody(r.Context)
		ok = result >= 0
		switch result {
		case JsonResultEmpty:
			key = JsonField{Begin: 0, Length: 0, Type: JsonValueTypeObject}
			value = JsonField{Begin: 0, Length: 0, Type: JsonValueTypeEnd}
			r.Context.StackIndex++
		case JsonResultError:
			r.Context.StackIndex = 0
		}
	case JsonValueTypeArray:
		key = JsonField{Begin: 0, Length: 0, Type: JsonValueTypeArray}
		var result JsonResult
		value, result = ParseArrayBody(r.Context)
		ok = result >= 0
		switch result {
		case JsonResultEmpty:
			// parsing array is done
			value = JsonField{Begin: 0, Length: 0, Type: JsonValueTypeEnd}
			r.Context.StackIndex++
		case JsonResultError:
			r.Context.StackIndex = 0
		}
	case JsonValueTypeString, JsonValueTypeNumber, JsonValueTypeBool, JsonValueTypeNull:
		ok = true
	case JsonValueTypeError:
		ok = false
	default:
		ok = false
	}

	if !ok {
		return JsonFieldError, JsonFieldError, false
	}

	return key, value, ok
}

func DetermineValueType(c *JsonContext) JsonValueType {
	SkipWhiteSpace(c)
	json := c.Json
	if c.Index >= len(json) {
		return JsonValueTypeError
	}

	switch json[c.Index] {
	case '{':
		return JsonValueTypeObject
	case '[':
		return JsonValueTypeArray
	case '"':
		return JsonValueTypeString
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
		return JsonValueTypeNumber
	case 'f', 't':
		return JsonValueTypeBool
	case 'n':
		return JsonValueTypeNull
	default:
		return JsonValueTypeError
	}
}

func ParseBegin(context *JsonContext) bool {
	SkipWhiteSpace(context)
	if context.Index >= len(context.Json) {
		return false
	}

	jsonByte := context.Json[context.Index]
	if jsonByte == '}' || jsonByte == ',' || jsonByte == '"' {
		return false
	}

	state := DetermineValueType(context)
	switch state {
	case JsonValueTypeNumber, JsonValueTypeBool, JsonValueTypeString, JsonValueTypeNull:
		return true
	case JsonValueTypeArray:
		context.StackIndex--
		context.Stack[context.StackIndex] = JsonValueTypeObject
		context.Index++ // skip '['
		return true
	case JsonValueTypeObject:
		context.StackIndex--
		context.Stack[context.StackIndex] = JsonValueTypeObject
		context.Index++ // skip '{'
		return true
	default:
		return false
	}
}

func ParseObjectBody(c *JsonContext) (outKey JsonField, outValue JsonField, result JsonResult) {
	json := c.Json
	if !SkipWhiteSpace(c) {
		return JsonFieldError, JsonFieldError, JsonResultError
	}

	if json[c.Index] == ',' {
		c.Index++
		if !SkipWhiteSpace(c) {
			return JsonFieldError, JsonFieldError, JsonResultError
		}
	}

	if json[c.Index] == '}' {
		c.Index++
		return JsonFieldEmpty, JsonFieldEmpty, JsonResultEmpty
	}

	result = JsonResultError

	if json[c.Index] != '"' {
		// should be "
		outKey = NewJsonField(c.Index, 1, JsonValueTypeError)
		outValue = JsonFieldEmpty
		return outKey, outValue, result
	}

	outKey = GetString(c) // get object key string

	if SkipWhiteSpaceUntil(c, ':') < 1 {
		outKey = NewJsonField(c.Index, 1, JsonValueTypeError)
		outValue = JsonFieldEmpty
		return outKey, outValue, result
	}

	c.Index++ // skip ':'
	state := DetermineValueType(c)
	result = JsonResult(state)
	switch state {
	case JsonValueTypeNumber:
		outValue = ParseNumber(c)
	case JsonValueTypeBool:
		outValue = ParseBoolean(c)
	case JsonValueTypeString:
		outValue = ParseString(c)
	case JsonValueTypeNull:
		outValue = ParseNull(c)
	case JsonValueTypeArray:
		if c.StackIndex == 0 {
			outKey = NewJsonField(c.Index, 1, JsonValueTypeError)
			outValue = JsonFieldEmpty
			result = JsonResultError
		} else {
			c.StackIndex--
			c.Stack[c.StackIndex] = JsonValueTypeArray
			outValue = NewJsonField(c.Index, 1, JsonValueTypeArray)
			c.Index++ // skip '['
		}
	case JsonValueTypeObject:
		if c.StackIndex == 0 {
			outKey = NewJsonField(c.Index, 1, JsonValueTypeError)
			result = JsonResultError
		} else {
			c.StackIndex--
			c.Stack[c.StackIndex] = JsonValueTypeObject
			outValue = NewJsonField(c.Index, 1, JsonValueTypeObject)
			c.Index++ // skip '{'
		}
	default:
		outKey = NewJsonField(c.Index, 1, JsonValueTypeError)
		outValue = JsonFieldEmpty
		result = JsonResultError
	}
	return outKey, outValue, result
}

func ParseArrayBody(c *JsonContext) (outValue JsonField, result JsonResult) {
	json := c.Json
	SkipWhiteSpace(c)
	if c.Index >= len(json) {
		return JsonFieldError, JsonResultError
	}

	if json[c.Index] == ',' {
		c.Index++
		SkipWhiteSpace(c)
	}

	if c.Index >= len(json) || json[c.Index] == ']' {
		c.Index++
		return JsonFieldEmpty, JsonResultEmpty
	}

	state := DetermineValueType(c)
	result = JsonResult(state)
	switch state {
	case JsonValueTypeNumber:
		outValue = ParseNumber(c)
	case JsonValueTypeBool:
		outValue = ParseBoolean(c)
	case JsonValueTypeString:
		outValue = ParseString(c)
	case JsonValueTypeNull:
		outValue = ParseNull(c)
	case JsonValueTypeArray:
		if c.StackIndex == 0 {
			outValue = JsonFieldError
			result = JsonResultError
		} else {
			c.StackIndex--
			c.Stack[c.StackIndex] = JsonValueTypeArray
			outValue = NewJsonField(c.Index, 1, JsonValueTypeArray)
			c.Index++ // skip '['
		}
	case JsonValueTypeObject:
		if c.StackIndex == 0 {
			outValue = JsonFieldError
			result = JsonResultError
		} else {
			c.StackIndex--
			c.Stack[c.StackIndex] = JsonValueTypeObject
			outValue = NewJsonField(c.Index, 1, JsonValueTypeObject)
			c.Index++ // skip '{'
		}
	default:
		outValue = JsonFieldEmpty
		result = JsonResultError
	}
	return outValue, result
}

func ParseString(c *JsonContext) JsonField {
	return GetString(c)
}

func ParseNumber(c *JsonContext) JsonField {
	span := JsonField{Begin: c.Index, Length: 0, Type: JsonValueTypeNumber}
	json := c.Json

	for c.Index < len(json) {
		switch json[c.Index] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '+', '.', 'e', 'E':
			c.Index++
			continue
		default:
			span.Length = c.Index - span.Begin
			return span
		}
	}

	// If we reach here, it means we hit the end of the JSON string without finding a non-number character.
	span.Length = c.Index - span.Begin
	return span

}

func ParseBoolean(c *JsonContext) JsonField {
	span := JsonField{Begin: c.Index, Length: 0, Type: JsonValueTypeBool}

	// Scan characters and see if they match any of the following
	// 1, t, T, true, True, TRUE
	// 0, f, F, false, False, FALSE
	// 1, y, Y, yes, Yes, YES
	// 0, n, N, no, No, NO
	if !SkipWhiteSpace(c) {
		return JsonFieldError
	}

	if end, ok := ScanUntilDelimiter(c); !ok {
		return JsonFieldError
	} else {
		span.Begin = c.Index
		span.Length = 0
		span.Type = JsonValueTypeError

		length := end - c.Index
		if length == 1 {
			if jsonByte := c.Json[c.Index]; jsonByte == '1' || jsonByte == 't' || jsonByte == 'T' || jsonByte == 'y' || jsonByte == 'Y' {
				span.Length = length
			} else if jsonByte == '0' || jsonByte == 'f' || jsonByte == 'F' || jsonByte == 'n' || jsonByte == 'N' {
				span.Length = length
			}
		} else if length == 2 {
			if strings.EqualFold(string(c.Json[c.Index:end]), "no") {
				span.Length = length
			}
		} else if length == 3 {
			if strings.EqualFold(string(c.Json[c.Index:end]), "yes") {
				span.Length = length
			}
		} else if length == 4 {
			if strings.EqualFold(string(c.Json[c.Index:end]), "true") {
				span.Length = length
			}
		} else if length == 5 {
			if strings.EqualFold(string(c.Json[c.Index:end]), "false") {
				span.Length = length
			}
		}

		if span.Length > 0 {
			c.Index += span.Length
			span.Type = JsonValueTypeBool
		}
	}
	return span
}

func ParseNull(c *JsonContext) JsonField {
	span := JsonField{Begin: c.Index, Length: 0, Type: JsonValueTypeNull}

	if !SkipWhiteSpace(c) {
		return JsonFieldError
	}

	if end, ok := ScanUntilDelimiter(c); !ok {
		return JsonFieldError
	} else {
		span.Begin = c.Index
		span.Length = 0
		span.Type = JsonValueTypeError

		length := end - c.Index
		if length == 4 {
			if strings.EqualFold(string(c.Json[c.Index:end]), "null") {
				span.Length = length
			}
		} else if length == 3 {
			if strings.EqualFold(string(c.Json[c.Index:end]), "nil") {
				span.Length = length
			}
		}

		if span.Length > 0 {
			c.Index += span.Length
			span.Type = JsonValueTypeNull
		}
	}
	return span
}

func ScanUntilDelimiter(c *JsonContext) (int, bool) {
	json := c.Json
	cursor := c.Index
	for cursor < len(json) {
		switch json[cursor] {
		case ' ', '\t', '\n', '\r', ',', ']', '}':
			return cursor, true
		default:
			cursor++
		}
	}
	// Means we are at the end of the string
	return c.Index, false
}

func SkipWhiteSpace(c *JsonContext) bool {
	json := c.Json
	for c.Index < len(json) {
		switch json[c.Index] {
		case ' ', '\t', '\n', '\r':
			c.Index++
		default:
			return true // Next character is not whitespace
		}
	}
	// Means we are at the end of the string
	return false
}

func SkipWhiteSpaceUntil(c *JsonContext, until byte) int {
	json := c.Json
	for c.Index < len(json) {
		switch json[c.Index] {
		case ' ', '\t', '\n', '\r':
			c.Index++
		default:
			if json[c.Index] == until {
				return 1
			}
			return 0
		}
	}
	// If we reach here, it means we hit the end of parsing
	return -1
}

func GetString(c *JsonContext) JsonField {
	start := c.Index + 1 // skip '"'
	json := c.Json

	for c.Index < len(json) {
		switch json[c.Index] {
		case '"':
			c.Index++ // move past the closing quote
			return JsonField{Begin: start, Length: c.Index - start - 1, Type: JsonValueTypeString}
		case '\\':
			c.Index++ // skip escaped character
			if c.Index >= len(json) {
				return JsonFieldError // error if we reach the end of the string
			}
		default:
			c.Index++
		}
	}

	return JsonFieldError // error if we reach here without finding a closing quote
}

func GetEscapedString(context *JsonContext, f JsonField) string {
	context.EscapedStrings.Reset()

	json := context.Json[f.Begin : f.Begin+f.Length]
	index := 0

	for index < f.Length {
		switch json[index] {
		case '"':
			index++
			break
		case '\\':
			index++ // skip the escape character
			if index < f.Length {
				switch json[index] {
				case '"':
					context.EscapedStrings.WriteAscii('"')
				case '\\':
					context.EscapedStrings.WriteAscii('\\')
				case '/':
					context.EscapedStrings.WriteAscii('/')
				case '\'':
					context.EscapedStrings.WriteAscii('\'')
				case 'b':
					context.EscapedStrings.WriteAscii('\b')
				case 'f':
					context.EscapedStrings.WriteAscii('\f')
				case 'n':
					context.EscapedStrings.WriteAscii('\n')
				case 'r':
					context.EscapedStrings.WriteAscii('\r')
				case 't':
					context.EscapedStrings.WriteAscii('\t')
				case 'u':
					var c rune
					c, index = GetUnicodeCodePoint(context, index)
					context.EscapedStrings.WriteRune(c)
				default:
					continue // unsupported, just keep
				}
				index++
			}
		default:
			context.EscapedStrings.WriteAscii(json[index])
			index++
		}
	}

	return context.EscapedStrings.String()
}

func GetUnicodeCodePoint(context *JsonContext, index int) (rune, int) {
	var unicode int32 = 0

	for i := 0; i < 4; i++ {
		if index >= len(context.Json) {
			return 0, index // error if we reach the end of the string
		}
		c := context.Json[index]
		var cp byte
		switch {
		case c >= '0' && c <= '9':
			cp = byte(c - '0')
		case c >= 'A' && c <= 'F':
			cp = byte(10 + (c - 'A'))
		case c >= 'a' && c <= 'f':
			cp = byte(10 + (c - 'a'))
		default:
			return 0, index // invalid character
		}
		unicode = (unicode << 4) | int32(cp)
		index++
	}

	return rune(unicode & 0xFFFF), index
}
