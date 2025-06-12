package foundation

import (
	"strconv"
	"strings"
)

type jsonValueType int8

const (
	JsonValueTypeError  jsonValueType = -1
	JsonValueTypeEmpty  jsonValueType = 0
	JsonValueTypeObject jsonValueType = 1
	JsonValueTypeArray  jsonValueType = 2
	JsonValueTypeString jsonValueType = 3
	JsonValueTypeNumber jsonValueType = 4
	JsonValueTypeBool   jsonValueType = 5
	JsonValueTypeNull   jsonValueType = 6
	JsonValueTypeEnd    jsonValueType = 7
)

// Note: Should map 1:1 to jsonValueType
type jsonResult int8

const (
	JsonResultError  jsonResult = -1
	JsonResultEmpty  jsonResult = 0
	JsonResultObject jsonResult = 1
	JsonResultArray  jsonResult = 2
	JsonResultString jsonResult = 3
	JsonResultNumber jsonResult = 4
	JsonResultBool   jsonResult = 5
	JsonResultNull   jsonResult = 6
	JsonResultEnd    jsonResult = 7
)

type jsonField struct {
	Begin  int
	Length int
	Type   jsonValueType
}

func (f jsonField) IsEmpty() bool {
	return f.Type == JsonValueTypeEmpty || f.Length == 0
}

var JsonFieldEmpty = jsonField{Begin: 0, Length: 0, Type: JsonValueTypeEmpty}
var JsonFieldError = jsonField{Begin: 0, Length: 0, Type: JsonValueTypeError}

func newJsonField(begin, length int, valueType jsonValueType) jsonField {
	return jsonField{Begin: begin, Length: length, Type: valueType}
}

type jsonContext struct {
	Json           string
	Cursor         int
	IsEscapeString bool
	Stack          []jsonValueType
	StackIndex     int16
	EscapedStrings *StringBuilder
}

const jsonStackSize = int16(256)

func newJsonContext() *jsonContext {
	return &jsonContext{
		Json:           "",
		Cursor:         0,
		IsEscapeString: false,
		Stack:          make([]jsonValueType, jsonStackSize),
		StackIndex:     jsonStackSize,
		EscapedStrings: NewStringBuilder(),
	}
}

func (c *jsonContext) isValidField(f jsonField) bool {
	return f.Begin >= 0 && f.Length > 0 && (f.Begin+f.Length > len(c.Json))
}

type JsonReader struct {
	Context *jsonContext
	Key     jsonField
	Value   jsonField
}

func NewJsonReader() *JsonReader {
	return &JsonReader{Context: newJsonContext()}
}

func (r *JsonReader) Begin(json string) bool {
	r.Context = newJsonContext()
	return r.Context.ParseBegin()
}

func (r *JsonReader) FieldStr(f jsonField) string {
	return r.Context.Json[f.Begin : f.Begin+f.Length]
}

func (r *JsonReader) ParseBool(field jsonField) bool {
	if !r.Context.isValidField(field) {
		return false
	}
	value, err := strconv.ParseBool(string(r.Context.Json[field.Begin : field.Begin+field.Length]))
	if err != nil {
		return false
	}
	return value
}

func (r *JsonReader) ParseFloat(field jsonField) float64 {
	if !r.Context.isValidField(field) {
		return 0.0
	}

	json := r.Context.Json
	value := json[field.Begin : field.Begin+field.Length]
	result, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return 0.0
	}
	return result
}

func (r *JsonReader) ParseInt(field jsonField) int {
	if !r.Context.isValidField(field) {
		return 0
	}

	json := r.Context.Json
	value := json[field.Begin : field.Begin+field.Length]

	intValue, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0
	}
	return int(intValue)
}

func (r *JsonReader) ParseLong(field jsonField) int64 {
	if !r.Context.isValidField(field) {
		return 0
	}

	json := r.Context.Json
	value := json[field.Begin : field.Begin+field.Length]

	longValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return longValue
}

func (r *JsonReader) ParseString(field jsonField) string {
	if r.Context.IsEscapeString {
		return r.Context.getEscapedString(field)
	} else {
		if !r.Context.isValidField(field) {
			return ""
		}
		return r.Context.Json[field.Begin : field.Begin+field.Length]
	}
}

func (r *JsonReader) IsFieldName(f jsonField, name string) bool {
	if !r.Context.isValidField(f) {
		return false
	}
	fieldName := r.Context.Json[f.Begin : f.Begin+f.Length]
	return strings.EqualFold(fieldName, name)
}

func (r *JsonReader) ReadUntilObjectEnd() bool {
	if ok := r.Read(); ok {
		return r.Key.Type == JsonValueTypeObject && r.Value.Type == JsonValueTypeEnd
	}
	return false
}

func (r *JsonReader) ReadUntilArrayEnd() bool {
	if ok := r.Read(); ok {
		return r.Key.Type == JsonValueTypeArray && r.Value.Type == JsonValueTypeEnd
	}
	return false
}

func (r *JsonReader) Read() (ok bool) {
	if r.Context.StackIndex == jsonStackSize {
		return false
	}

	r.Key = JsonFieldError
	r.Value = JsonFieldError

	state := r.Context.Stack[r.Context.StackIndex]
	switch state {
	case JsonValueTypeObject:
		key, value, result := r.Context.parseObjectBody()
		ok = result >= 0
		switch result {
		case JsonResultEmpty:
			r.Key = jsonField{Begin: 0, Length: 0, Type: JsonValueTypeObject}
			r.Value = jsonField{Begin: 0, Length: 0, Type: JsonValueTypeEnd}
			r.Context.StackIndex++
		case JsonResultError:
			r.Context.StackIndex = 0
		default:
			r.Key = key
			r.Value = value
		}
	case JsonValueTypeArray:
		r.Key = jsonField{Begin: 0, Length: 0, Type: JsonValueTypeArray}
		value, result := r.Context.parseArrayBody()
		ok = result >= 0
		switch result {
		case JsonResultEmpty:
			// parsing array is done
			r.Value = jsonField{Begin: 0, Length: 0, Type: JsonValueTypeEnd}
			r.Context.StackIndex++
		case JsonResultError:
			r.Context.StackIndex = 0
		default:
			r.Value = value
		}
	case JsonValueTypeString, JsonValueTypeNumber, JsonValueTypeBool, JsonValueTypeNull:
		ok = true
	case JsonValueTypeError:
		ok = false
	default:
		ok = false
	}

	return ok
}

func (c *jsonContext) determineValueType() jsonValueType {
	c.skipWhiteSpace()
	json := c.Json
	if c.Cursor >= len(json) {
		return JsonValueTypeError
	}

	switch json[c.Cursor] {
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

func (context *jsonContext) ParseBegin() bool {
	context.skipWhiteSpace()
	if context.Cursor >= len(context.Json) {
		return false
	}

	jsonByte := context.Json[context.Cursor]
	if jsonByte == '}' || jsonByte == ',' || jsonByte == '"' {
		return false
	}

	state := context.determineValueType()
	switch state {
	case JsonValueTypeNumber, JsonValueTypeBool, JsonValueTypeString, JsonValueTypeNull:
		return true
	case JsonValueTypeArray:
		context.StackIndex--
		context.Stack[context.StackIndex] = JsonValueTypeObject
		context.Cursor++ // skip '['
		return true
	case JsonValueTypeObject:
		context.StackIndex--
		context.Stack[context.StackIndex] = JsonValueTypeObject
		context.Cursor++ // skip '{'
		return true
	default:
		return false
	}
}

func (c *jsonContext) parseObjectBody() (outKey jsonField, outValue jsonField, result jsonResult) {
	json := c.Json
	if !c.skipWhiteSpace() {
		return JsonFieldError, JsonFieldError, JsonResultError
	}

	if json[c.Cursor] == ',' {
		c.Cursor++
		if !c.skipWhiteSpace() {
			return JsonFieldError, JsonFieldError, JsonResultError
		}
	}

	if json[c.Cursor] == '}' {
		c.Cursor++
		return JsonFieldError, JsonFieldError, JsonResultEmpty
	}

	result = JsonResultError

	if json[c.Cursor] != '"' {
		// should be "
		outKey = newJsonField(c.Cursor, 1, JsonValueTypeError)
		outValue = JsonFieldEmpty
		return
	}

	outKey = c.getString() // get object key string

	if c.skipWhiteSpaceUntil(':') < 1 {
		outKey = newJsonField(c.Cursor, 1, JsonValueTypeError)
		outValue = JsonFieldEmpty
		return
	}

	c.Cursor++ // skip ':'
	state := c.determineValueType()
	result = jsonResult(state)
	switch state {
	case JsonValueTypeNumber:
		outValue = c.parseNumber()
	case JsonValueTypeBool:
		outValue = c.parseBoolean()
	case JsonValueTypeString:
		outValue = c.parseString()
	case JsonValueTypeNull:
		outValue = c.parseNull()
	case JsonValueTypeArray:
		if c.StackIndex == 0 {
			outKey = newJsonField(c.Cursor, 1, JsonValueTypeError)
			outValue = JsonFieldEmpty
			result = JsonResultError
		} else {
			c.StackIndex--
			c.Stack[c.StackIndex] = JsonValueTypeArray
			outValue = newJsonField(c.Cursor, 1, JsonValueTypeArray)
			c.Cursor++ // skip '['
		}
	case JsonValueTypeObject:
		if c.StackIndex == 0 {
			outKey = newJsonField(c.Cursor, 1, JsonValueTypeError)
			result = JsonResultError
		} else {
			c.StackIndex--
			c.Stack[c.StackIndex] = JsonValueTypeObject
			outValue = newJsonField(c.Cursor, 1, JsonValueTypeObject)
			c.Cursor++ // skip '{'
		}
	default:
		outKey = newJsonField(c.Cursor, 1, JsonValueTypeError)
		outValue = JsonFieldEmpty
		result = JsonResultError
	}
	return
}

func (c *jsonContext) parseArrayBody() (outValue jsonField, result jsonResult) {
	json := c.Json
	c.skipWhiteSpace()
	if c.Cursor >= len(json) {
		return JsonFieldError, JsonResultError
	}

	if json[c.Cursor] == ',' {
		c.Cursor++
		c.skipWhiteSpace()
	}

	if c.Cursor >= len(json) || json[c.Cursor] == ']' {
		c.Cursor++
		return JsonFieldEmpty, JsonResultEmpty
	}

	state := c.determineValueType()
	result = jsonResult(state)
	switch state {
	case JsonValueTypeNumber:
		outValue = c.parseNumber()
	case JsonValueTypeBool:
		outValue = c.parseBoolean()
	case JsonValueTypeString:
		outValue = c.parseString()
	case JsonValueTypeNull:
		outValue = c.parseNull()
	case JsonValueTypeArray:
		if c.StackIndex == 0 {
			outValue = JsonFieldError
			result = JsonResultError
		} else {
			c.StackIndex--
			c.Stack[c.StackIndex] = JsonValueTypeArray
			outValue = newJsonField(c.Cursor, 1, JsonValueTypeArray)
			c.Cursor++ // skip '['
		}
	case JsonValueTypeObject:
		if c.StackIndex == 0 {
			outValue = JsonFieldError
			result = JsonResultError
		} else {
			c.StackIndex--
			c.Stack[c.StackIndex] = JsonValueTypeObject
			outValue = newJsonField(c.Cursor, 1, JsonValueTypeObject)
			c.Cursor++ // skip '{'
		}
	default:
		outValue = JsonFieldEmpty
		result = JsonResultError
	}
	return
}

func (c *jsonContext) parseString() jsonField {
	return c.getString()
}

func (c *jsonContext) parseNumber() jsonField {
	span := jsonField{Begin: c.Cursor, Length: 0, Type: JsonValueTypeNumber}
	json := c.Json

	for c.Cursor < len(json) {
		switch json[c.Cursor] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '+', '.', 'e', 'E':
			c.Cursor++
			continue
		default:
			span.Length = c.Cursor - span.Begin
			return span
		}
	}

	// If we reach here, it means we hit the end of the JSON string without finding a non-number character.
	span.Length = c.Cursor - span.Begin
	return span

}

func (c *jsonContext) parseBoolean() jsonField {
	span := jsonField{Begin: c.Cursor, Length: 0, Type: JsonValueTypeBool}

	// Scan characters and see if they match any of the following
	// 1, t, T, true, True, TRUE
	// 0, f, F, false, False, FALSE
	// 1, y, Y, yes, Yes, YES
	// 0, n, N, no, No, NO
	if !c.skipWhiteSpace() {
		return JsonFieldError
	}

	if end, ok := c.scanUntilDelimiter(); !ok {
		return JsonFieldError
	} else {
		span.Begin = c.Cursor
		span.Length = 0
		span.Type = JsonValueTypeError

		length := end - c.Cursor
		if length == 1 {
			if jsonByte := c.Json[c.Cursor]; jsonByte == '1' || jsonByte == 't' || jsonByte == 'T' || jsonByte == 'y' || jsonByte == 'Y' {
				span.Length = length
			} else if jsonByte == '0' || jsonByte == 'f' || jsonByte == 'F' || jsonByte == 'n' || jsonByte == 'N' {
				span.Length = length
			}
		} else if length == 2 {
			if strings.EqualFold(string(c.Json[c.Cursor:end]), "no") {
				span.Length = length
			}
		} else if length == 3 {
			if strings.EqualFold(string(c.Json[c.Cursor:end]), "yes") {
				span.Length = length
			}
		} else if length == 4 {
			if strings.EqualFold(string(c.Json[c.Cursor:end]), "true") {
				span.Length = length
			}
		} else if length == 5 {
			if strings.EqualFold(string(c.Json[c.Cursor:end]), "false") {
				span.Length = length
			}
		}

		if span.Length > 0 {
			c.Cursor += span.Length
			span.Type = JsonValueTypeBool
		}
	}
	return span
}

func (c *jsonContext) parseNull() jsonField {
	span := jsonField{Begin: c.Cursor, Length: 0, Type: JsonValueTypeNull}

	if !c.skipWhiteSpace() {
		return JsonFieldError
	}

	if end, ok := c.scanUntilDelimiter(); !ok {
		return JsonFieldError
	} else {
		span.Begin = c.Cursor
		span.Length = 0
		span.Type = JsonValueTypeError

		length := end - c.Cursor
		if length == 4 {
			if strings.EqualFold(string(c.Json[c.Cursor:end]), "null") {
				span.Length = length
			}
		} else if length == 3 {
			if strings.EqualFold(string(c.Json[c.Cursor:end]), "nil") {
				span.Length = length
			}
		}

		if span.Length > 0 {
			c.Cursor += span.Length
			span.Type = JsonValueTypeNull
		}
	}
	return span
}

func (c *jsonContext) scanUntilDelimiter() (int, bool) {
	json := c.Json
	cursor := c.Cursor
	for cursor < len(json) {
		switch json[cursor] {
		case ' ', '\t', '\n', '\r', ',', ']', '}':
			return cursor, true
		default:
			cursor++
		}
	}
	// Means we are at the end of the string
	return c.Cursor, false
}

func (c *jsonContext) skipWhiteSpace() bool {
	json := c.Json
	for c.Cursor < len(json) {
		switch json[c.Cursor] {
		case ' ', '\t', '\n', '\r':
			c.Cursor++
		default:
			return true // Next character is not whitespace
		}
	}
	// Means we are at the end of the string
	return false
}

func (c *jsonContext) skipWhiteSpaceUntil(until byte) int {
	json := c.Json
	for c.Cursor < len(json) {
		switch json[c.Cursor] {
		case ' ', '\t', '\n', '\r':
			c.Cursor++
		default:
			if json[c.Cursor] == until {
				return 1
			}
			return 0
		}
	}
	// If we reach here, it means we hit the end of parsing
	return -1
}

func (c *jsonContext) getString() jsonField {
	start := c.Cursor + 1 // skip '"'
	json := c.Json

	for c.Cursor < len(json) {
		switch json[c.Cursor] {
		case '"':
			c.Cursor++ // move past the closing quote
			return jsonField{Begin: start, Length: c.Cursor - start - 1, Type: JsonValueTypeString}
		case '\\':
			c.Cursor++ // skip escaped character
			if c.Cursor >= len(json) {
				return JsonFieldError // error if we reach the end of the string
			}
		default:
			c.Cursor++
		}
	}

	return JsonFieldError // error if we reach here without finding a closing quote
}

func (context *jsonContext) getEscapedString(f jsonField) string {
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
					c, index = context.getUnicodeCodePoint(index)
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

func (context *jsonContext) getUnicodeCodePoint(index int) (rune, int) {
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
