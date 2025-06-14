package foundation

import (
	"fmt"
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

var jsonValueTypeStringMap = map[jsonValueType]string{
	JsonValueTypeError:  "Error",
	JsonValueTypeEmpty:  "Empty",
	JsonValueTypeObject: "Object",
	JsonValueTypeArray:  "Array",
	JsonValueTypeString: "String",
	JsonValueTypeNumber: "Number",
	JsonValueTypeBool:   "Bool",
	JsonValueTypeNull:   "Null",
	JsonValueTypeEnd:    "End",
}

func (valueType jsonValueType) String() string {
	return jsonValueTypeStringMap[valueType]
}

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

var jsonResultStringMap = map[jsonResult]string{
	JsonResultError:  "Error",
	JsonResultEmpty:  "Empty",
	JsonResultObject: "Object",
	JsonResultArray:  "Array",
	JsonResultString: "String",
	JsonResultNumber: "Number",
	JsonResultBool:   "Bool",
	JsonResultNull:   "Null",
	JsonResultEnd:    "End",
}

func (result jsonResult) String() string {
	return jsonResultStringMap[result]
}

type jsonField struct {
	Begin   int
	Length  int16
	Type    jsonValueType
	Padding int8
}

func (f jsonField) String() string {
	return fmt.Sprintf("Field{Begin: %d, Length: %d, Type: %s}", f.Begin, f.Length, f.Type.String())
}

func (f jsonField) IsEmpty() bool {
	return f.Type == JsonValueTypeEmpty || f.Length == 0
}

var JsonFieldEmpty = jsonField{Begin: 0, Length: 0, Type: JsonValueTypeEmpty}
var JsonFieldError = jsonField{Begin: 0, Length: 0, Type: JsonValueTypeError}

func newJsonField(begin int, length int16, valueType jsonValueType) jsonField {
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
	return f.Begin >= 0 && f.Length > 0 && (f.Begin+int(f.Length) < len(c.Json))
}

type JsonDecoder struct {
	Context *jsonContext
	Key     jsonField
	Value   jsonField
	Error   error
}

func NewJsonDecoder() *JsonDecoder {
	return &JsonDecoder{Context: newJsonContext()}
}

func (r *JsonDecoder) Begin(json string) bool {
	return r.Context.ParseBegin(json)
}

// ---------------------------------------------------------------------------
// ---------------------------------------------------------------------------

type JsonDecode func(decoder *JsonDecoder)

func (d *JsonDecoder) Decode(fields map[string]JsonDecode) error {
	for !d.ReadUntilObjectEnd() {
		fname := d.DecodeField()
		if fdecode, ok := fields[strings.ToLower(fname)]; ok {
			fdecode(d)
		}
	}
	return nil
}

func (d *JsonDecoder) DecodeField() string {
	return d.FieldStr(d.Key)
}
func (d *JsonDecoder) DecodeBool() bool {
	return d.ParseBool(d.Value)
}
func (d *JsonDecoder) DecodeInt32() int32 {
	return int32(d.ParseInt32(d.Value))
}
func (d *JsonDecoder) DecodeInt64() int64 {
	return d.ParseInt64(d.Value)
}
func (d *JsonDecoder) DecodeFloat32() float32 {
	return d.ParseFloat32(d.Value)
}
func (d *JsonDecoder) DecodeFloat64() float64 {
	return d.ParseFloat64(d.Value)
}
func (d *JsonDecoder) DecodeString() string {
	return d.ParseString(d.Value)
}

func (d *JsonDecoder) DecodeStringArray() (result []string) {
	result = make([]string, 0, 4)
	for !d.ReadUntilArrayEnd() {
		str := d.DecodeString()
		result = append(result, str)
	}
	return
}

func (d *JsonDecoder) DecodeStringMapString() (result map[string]string) {
	result = make(map[string]string, 4)
	for !d.ReadUntilObjectEnd() {
		key := d.DecodeField()
		value := d.DecodeString()
		result[key] = value
	}
	return result
}

// ---------------------------------------------------------------------------
// ---------------------------------------------------------------------------

func (r *JsonDecoder) FieldStr(f jsonField) string {
	return r.Context.Json[f.Begin : f.Begin+int(f.Length)]
}

func (r *JsonDecoder) ParseBool(field jsonField) (value bool) {
	if r.Context.isValidField(field) {
		value, r.Error = strconv.ParseBool(string(r.Context.Json[field.Begin : field.Begin+int(field.Length)]))
	}
	return
}

func (r *JsonDecoder) ParseFloat32(field jsonField) (result float32) {
	if r.Context.isValidField(field) {
		value := r.Context.Json[field.Begin : field.Begin+int(field.Length)]
		var r64 float64
		r64, r.Error = strconv.ParseFloat(value, 32)
		result = float32(r64)
	}
	return
}

func (r *JsonDecoder) ParseFloat64(field jsonField) (result float64) {
	if r.Context.isValidField(field) {
		json := r.Context.Json
		valueStr := json[field.Begin : field.Begin+int(field.Length)]
		result, r.Error = strconv.ParseFloat(valueStr, 64)
	} else {
		r.Error = fmt.Errorf("invalid '%s'", field.String())
	}
	return result
}

func (r *JsonDecoder) ParseInt32(field jsonField) (result int) {
	result = int(r.ParseInt64(field))
	return
}

func (r *JsonDecoder) ParseInt64(field jsonField) (result int64) {
	if r.Context.isValidField(field) {
		valueStr := r.Context.Json[field.Begin : field.Begin+int(field.Length)]
		result, r.Error = strconv.ParseInt(valueStr, 10, 64)
	} else {
		r.Error = fmt.Errorf("invalid '%s'", field.String())
	}
	return
}

func (r *JsonDecoder) ParseString(field jsonField) (result string) {
	if r.Context.IsEscapeString {
		var ok bool
		if result, ok = r.Context.getEscapedString(field); !ok {
			r.Error = fmt.Errorf("invalid string at %s", field.String())
		}
	} else {
		if r.Context.isValidField(field) {
			result = r.Context.Json[field.Begin : field.Begin+int(field.Length)]
		}
	}
	return
}

func (r *JsonDecoder) IsFieldName(field jsonField, name string) (result bool) {
	if r.Context.isValidField(field) {
		fieldName := r.Context.Json[field.Begin : field.Begin+int(field.Length)]
		result = strings.EqualFold(fieldName, name)
	} else {
		r.Error = fmt.Errorf("invalid '%s'", field.String())
	}
	return
}

func (r *JsonDecoder) ReadUntilObjectEnd() (ok bool) {
	ok = r.Read()
	return ok && r.Key.Type == JsonValueTypeObject && r.Value.Type == JsonValueTypeEnd
}

func (r *JsonDecoder) ReadUntilArrayEnd() (ok bool) {
	ok = r.Read()
	return ok && r.Key.Type == JsonValueTypeArray && r.Value.Type == JsonValueTypeEnd
}

func (r *JsonDecoder) Read() (ok bool) {
	if r.Context.StackIndex == jsonStackSize {
		r.Error = fmt.Errorf("invalid JSON, current position at %d", r.Context.Cursor)
		return false
	}

	r.Key = JsonFieldError
	r.Value = JsonFieldError

	// The stack should only contain 'JsonValueTypeObject' or 'JsonValueTypeArray'
	state := r.Context.Stack[r.Context.StackIndex]
	ok = true
	switch state {
	case JsonValueTypeObject:
		key, value, result := r.Context.parseObjectBody()
		switch result {
		case JsonResultEmpty: // End of object
			r.Key = jsonField{Begin: 0, Length: 0, Type: JsonValueTypeObject}
			r.Value = jsonField{Begin: 0, Length: 0, Type: JsonValueTypeEnd}
			r.Context.StackIndex++
		case JsonResultError:
			r.Error = fmt.Errorf("error parsing object at %d", r.Context.Cursor)
			ok = false
		default:
			r.Key = key
			r.Value = value
		}
	case JsonValueTypeArray:
		r.Key = jsonField{Begin: 0, Length: 0, Type: JsonValueTypeArray}
		value, result := r.Context.parseArrayBody()
		switch result {
		case JsonResultEmpty: // End of array
			r.Value = jsonField{Begin: 0, Length: 0, Type: JsonValueTypeEnd}
			r.Context.StackIndex++
		case JsonResultError:
			r.Error = fmt.Errorf("error parsing array at %d", r.Context.Cursor)
			ok = false
		default:
			r.Value = value
		}
	default:
		r.Error = fmt.Errorf("error reading at %d", r.Context.Cursor)
		ok = false
	}

	return ok
}

func (c *jsonContext) determineValueType() jsonValueType {
	if !c.skipWhiteSpace() {
		return JsonValueTypeError
	}
	switch c.Json[c.Cursor] {
	case '{':
		return JsonValueTypeObject
	case '[':
		return JsonValueTypeArray
	case '"':
		return JsonValueTypeString
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
		return JsonValueTypeNumber
	case 'f', 'F':
		return JsonValueTypeBool
	case 't', 'T':
		return JsonValueTypeBool
	case 'n':
		return JsonValueTypeNull
	default:
		return JsonValueTypeError
	}
}

func (c *jsonContext) ParseBegin(json string) bool {
	c.Json = json
	c.Cursor = 0

	if !c.skipWhiteSpace() {
		return false
	}

	jsonByte := c.Json[c.Cursor]
	if jsonByte == '}' || jsonByte == ',' || jsonByte == '"' {
		return false
	}

	state := c.determineValueType()
	switch state {
	case JsonValueTypeNumber, JsonValueTypeBool, JsonValueTypeString, JsonValueTypeNull:
		return true
	case JsonValueTypeArray:
		c.StackIndex--
		c.Stack[c.StackIndex] = JsonValueTypeObject
		c.Cursor++ // skip '['
		return true
	case JsonValueTypeObject:
		c.StackIndex--
		c.Stack[c.StackIndex] = JsonValueTypeObject
		c.Cursor++ // skip '{'
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
	if !c.skipWhiteSpace() {
		return JsonFieldError, JsonResultError
	}

	if c.Json[c.Cursor] == ',' {
		c.Cursor++
		if !c.skipWhiteSpace() {
			return JsonFieldError, JsonResultError
		}
	}

	if c.Json[c.Cursor] == ']' {
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
	for c.Cursor < len(c.Json) {
		b := c.Json[c.Cursor]
		if b >= '0' && b <= '9' || b == '-' || b == '+' || b == '.' || b == 'e' || b == 'E' {
			c.Cursor++ // Move to the next character
			continue
		}
		break
	}

	// If we reach here, it means we hit the end of the JSON string without finding a non-number character.
	span.Length = int16(c.Cursor - span.Begin)
	return span

}

func (c *jsonContext) parseBoolean() jsonField {
	span := jsonField{Begin: c.Cursor, Length: 0, Type: JsonValueTypeBool}

	// Scan characters and see if they match any of the following
	// t, T, true, True, TRUE
	// f, F, false, False, FALSE
	if !c.skipWhiteSpace() {
		return JsonFieldError
	}

	if end, ok := c.scanUntilDelimiter(); !ok {
		return JsonFieldError
	} else {
		span.Begin = c.Cursor
		span.Length = int16(end - c.Cursor)
		span.Type = JsonValueTypeBool
		c.Cursor += int(span.Length)
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

		length := int16(end - c.Cursor)
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
			c.Cursor += int(span.Length)
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
	for c.Cursor < len(c.Json) {
		switch c.Json[c.Cursor] {
		case ' ', '\t', '\n', '\r':
			c.Cursor++
		default:
			return true // Next character is not whitespace
		}
	}
	// Means we are at the end of the JSON content
	return false
}

func (c *jsonContext) skipWhiteSpaceUntil(until byte) int {
	for c.Cursor < len(c.Json) {
		b := c.Json[c.Cursor]
		switch b {
		case ' ', '\t', '\n', '\r':
			c.Cursor++
		case until:
			return 1
		default:
			return 0
		}
	}
	// If we reach here, it means we hit the end of the JSON content
	return -1
}

func (c *jsonContext) getString() jsonField {
	c.Cursor += 1 // skip '"'
	start := c.Cursor
	for c.Cursor < len(c.Json) {
		switch c.Json[c.Cursor] {
		case '"':
			c.Cursor++ // move past the closing quote
			return jsonField{Begin: start, Length: int16(c.Cursor-start) - 1, Type: JsonValueTypeString}
		case '\\':
			c.Cursor++ // skip escaped character
		default:
			c.Cursor++
		}
	}

	return JsonFieldError // error if we reach here without finding a closing quote
}

func (c *jsonContext) getEscapedString(f jsonField) (result string, ok bool) {
	c.EscapedStrings.Reset()

	index := f.Begin
	end := f.Begin + int(f.Length)
	for index < end {
		switch c.Json[index] {
		case '"':
			index++
			break
		case '\\':
			index++ // skip the escape character
			if index < end {
				switch c.Json[index] {
				case '"':
					c.EscapedStrings.WriteAscii('"')
					index++
				case '\\':
					c.EscapedStrings.WriteAscii('\\')
					index++
				case '/':
					c.EscapedStrings.WriteAscii('/')
					index++
				case '\'':
					c.EscapedStrings.WriteAscii('\'')
					index++
				case 'b':
					c.EscapedStrings.WriteAscii('\b')
					index++
				case 'f':
					c.EscapedStrings.WriteAscii('\f')
					index++
				case 'n':
					c.EscapedStrings.WriteAscii('\n')
					index++
				case 'r':
					c.EscapedStrings.WriteAscii('\r')
					index++
				case 't':
					c.EscapedStrings.WriteAscii('\t')
					index++
				case 'u':
					if index+4 <= end {
						var u rune
						if u, index = c.getUnicodeCodePoint(index + 1); u != 0 {
							c.EscapedStrings.WriteRune(u)
						}
					} else {
						return "", false // error in unicode escape
					}
				default:
					continue // unsupported, just keep
				}
			}
		default:
			c.EscapedStrings.WriteAscii(c.Json[index])
			index++
		}
	}

	result = c.EscapedStrings.String()
	return result, true
}

func (c *jsonContext) getUnicodeCodePoint(cursor int) (result rune, index int) {
	result = 0
	index = cursor
	for range 4 {
		result = (result << 4)
		switch c.Json[index] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			result = result | rune(c.Json[index]-'0')
		case 'A', 'B', 'C', 'D', 'E', 'F':
			result = result | rune(10+(c.Json[index]-'A'))
		case 'a', 'b', 'c', 'd', 'e', 'f':
			result = result | rune(10+(c.Json[index]-'a'))
		default:
			result = result | 0
		}
		index++
	}
	return
}
