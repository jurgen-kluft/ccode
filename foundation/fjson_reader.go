package foundation

import (
	"strconv"
	"strings"
)

type JsonValueType int

const (
	JsonValueTypeObject JsonValueType = iota
	JsonValueTypeArray
	JsonValueTypeString
	JsonValueTypeNumber
	JsonValueTypeBool
	JsonValueTypeNull
	JsonValueTypeEmpty
	JsonValueTypeError
	JsonValueTypeEnd
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
	Json           []byte
	Index          int
	IsEscapeString bool
	Stack          []JsonValueType
	StackIndex     int
	EscapedStrings StringBuilder
}

func NewJsonContext() *JsonContext {
	return &JsonContext{
		Json:           []byte{},
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

// public class JsonReader
//
//	{
//	    private Context _context;
//
//	    public bool Begin(string json)
//	    {
//	        _context = new Context(json);
//	        return ParseBegin(ref _context);
//	    }
func (r *JsonReader) Begin(json string) bool {
	return ParseBegin(NewJsonContext())
}

// public ReadOnlySpan<char> FieldStr(Field f)
//
//	{
//	    return _context.Json.Slice(f.Begin, f.Length);
//	}
func (r *JsonReader) FieldStr(f JsonField) []byte {
	return r.Context.Json[f.Begin : f.Begin+f.Length]
}

// public bool ParseBool(Field field)
//
//	{
//	    var json = _context.Json;
//	    if (bool.TryParse(json.Slice(field.Begin, field.Length), out var result))
//	    {
//	        return result;
//	    }
//	    return false;
//	}
func (r *JsonReader) ParseBool(field JsonField) bool {
	if !r.Context.IsValidField(field) {
		return false
	}

	json := r.Context.Json
	value := strings.ToLower(string(json[field.Begin : field.Begin+field.Length]))

	switch value {
	case "true":
		return true
	case "false":
		return false
	case "on":
		return true
	case "off":
		return false
	case "yes":
		return true
	case "no":
		return false
	case "1":
		return true
	case "0":
		return false
	default:
		return false
	}
}

//     public float ParseFloat(Field field)
//     {
//         var json = _context.Json;
//         if (float.TryParse(json.Slice(field.Begin, field.Length), out var result))
//         {
//             return result;
//         }
//         return 0.0f;
//     }

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

//     public int ParseInt(Field field)
//     {
//         var json = _context.Json;
//         if (int.TryParse(json.Slice(field.Begin, field.Length), out var result))
//         {
//             return result;
//         }
//         return 0;
//     }

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

// public long ParseLong(Field field)
//
//	{
//	    var json = _context.Json;
//	    if (long.TryParse(json.Slice(field.Begin, field.Length), out var result))
//	    {
//	        return result;
//	    }
//	    return 0;
//	}
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

//     public string ParseString(Field field)
//     {
//         if (_context.IsEscapeString)
//         {
//             return GetEscapedString(ref _context, field);
//         }
//         else
//         {
//             var json = _context.Json;
//             return json.Slice(field.Begin, field.Length).ToString();
//         }
//     }

func (r *JsonReader) ParseString(field JsonField) string {
	if r.Context.IsEscapeString {
		return GetEscapedString(r.Context, field)
	} else {
		if !r.Context.IsValidField(field) {
			return ""
		}
		json := r.Context.Json
		return string(json[field.Begin : field.Begin+field.Length])
	}
}

// public bool IsFieldName(Field f, string name)
//
//	{
//	    return name.AsSpan().CompareTo(_context.Json.Slice(f.Begin, f.Length), StringComparison.OrdinalIgnoreCase) == 0;
//	}
func (r *JsonReader) IsFieldName(f JsonField, name string) bool {
	if !r.Context.IsValidField(f) {
		return false
	}

	json := r.Context.Json
	fieldName := string(json[f.Begin : f.Begin+f.Length])
	return strings.EqualFold(fieldName, name)
}

// public bool IsObjectEnd(Field key, Field value)
//
//	{
//	    return key.Type == ValueType.Object && value.Type == ValueType.End;
//	}
func (r *JsonReader) IsObjectEnd(key JsonField, value JsonField) bool {
	return key.Type == JsonValueTypeObject && value.Type == JsonValueTypeEnd
}

// public bool ReadUntilObjectEnd(out Field key, out Field value)
//
//	{
//	    if (Read(out key, out value))
//	    {
//	        return IsObjectEnd(key, value);
//	    }
//	    return false;
//	}
func (r *JsonReader) ReadUntilObjectEnd() (JsonField, JsonField, bool) {
	if key, value, ok := r.Read(); ok {
		return key, value, r.IsObjectEnd(key, value)
	}
	return JsonFieldError, JsonFieldError, false
}

// public bool IsArrayEnd(Field key, Field value)
//
//	{
//	    return key.Type == ValueType.Array && value.Type == ValueType.End;
//	}
func (r *JsonReader) IsArrayEnd(key JsonField, value JsonField) bool {
	return key.Type == JsonValueTypeArray && value.Type == JsonValueTypeEnd
}

// public bool ReadUntilArrayEnd(out Field key, out Field value)
//
//	{
//	    if (Read(out key, out value))
//	    {
//	        return IsArrayEnd(key, value);
//	    }
//	    return false;
//	}
func (r *JsonReader) ReadUntilArrayEnd() (JsonField, JsonField, bool) {
	if key, value, ok := r.Read(); ok {
		return key, value, r.IsArrayEnd(key, value)
	}
	return JsonFieldError, JsonFieldError, false
}

//     public bool Read(out Field key, out Field value)
//     {
//         key = Field.Error;
//         value = Field.Error;
//         if (_context.StackIndex == _context.Stack.Length)
//         {
//             return false;
//         }
//         var state = _context.Stack[_context.StackIndex];
//         switch (state)
//         {
//             case ValueType.Object:
//                 switch (ParseObjectBody(ref _context, out key, out value))
//                 {
//                     case 0:
//                         key = new Field(0,0, ValueType.Object);
//                         value = new Field(0,0, ValueType.End);
//                         _context.StackIndex++;
//                         break;
//                     case 1:
//                         // a key value was parsed
//                         break;
//                     case 2:
//                         // a key was parsed with an object or array as value
//                         break;
//                     case -1:
//                         _context.StackIndex = 0;
//                         return false;
//                 }
//                 break;
//             case ValueType.Array:
//                 key = new Field(0,0, ValueType.Array);
//                 switch (ParseArrayBody(ref _context, out value))
//                 {
//                     case 0:
//                         // parsing array is done
//                         value = new Field(0,0, ValueType.End);
//                         _context.StackIndex++;
//                         break;
//                     case 1:
//                         // a simple element was parsed
//                         break;
//                     case 2:
//                         // the array element is an array or object
//                         break;
//                     case -1:
//                         key = new Field(0,0, ValueType.Error);
//                         value = new Field(0,0, ValueType.Error);
//                         _context.StackIndex = 0;
//                         return false;
//                 }
//                 break;
//             case ValueType.String:
//             case ValueType.Number:
//             case ValueType.Bool:
//             case ValueType.Null:
//                 break;
//             case ValueType.Error:
//                 return false;
//         }
//         return true;
//     }

func (r *JsonReader) Read() (key JsonField, value JsonField, ok bool) {
	if r.Context.StackIndex == len(r.Context.Stack) {
		return JsonFieldError, JsonFieldError, false
	}

	state := r.Context.Stack[r.Context.StackIndex]
	switch state {
	case JsonValueTypeObject:
		var result int
		key, value, result = ParseObjectBody(r.Context)
		switch result {
		case 0:
			key = JsonField{Begin: 0, Length: 0, Type: JsonValueTypeObject}
			value = JsonField{Begin: 0, Length: 0, Type: JsonValueTypeEnd}
			r.Context.StackIndex++
		case 1:
			// a key value was parsed
		case 2:
			// a key was parsed with an object or array as value
		case -1:
			r.Context.StackIndex = 0
			return JsonFieldError, JsonFieldError, false
		}
	case JsonValueTypeArray:
		key = JsonField{Begin: 0, Length: 0, Type: JsonValueTypeArray}
		var result int
		value, result = ParseArrayBody(r.Context)
		switch result {
		case 0:
			// parsing array is done
			value = JsonField{Begin: 0, Length: 0, Type: JsonValueTypeEnd}
			r.Context.StackIndex++
		case 1:
			// a simple element was parsed
		case 2:
			// the array element is an array or object
		case -1:
			r.Context.StackIndex = 0
			return JsonFieldError, JsonFieldError, false
		}
	case JsonValueTypeString, JsonValueTypeNumber, JsonValueTypeBool, JsonValueTypeNull:
	case JsonValueTypeError:
		return JsonFieldError, JsonFieldError, false
	default:
		return JsonFieldError, JsonFieldError, false
	}

	return key, value, true
}

// private static ValueType DetermineValueType(ref Context context)
//
//	{
//	    SkipWhiteSpace(ref context);
//	    var json = context.Json;
//	    switch (json[context.Index])
//	    {
//	        case '{':
//	            return ValueType.Object;
//	        case '[':
//	            return ValueType.Array;
//	        case '"':
//	            return ValueType.String;
//	        case  >= '0' and  <= '9':
//	        case '-':
//	            return ValueType.Number;
//	        case 'f':
//	        case 't':
//	            return ValueType.Bool;
//	        case 'n':
//	            return ValueType.Null;
//	    }
//	    return ValueType.Error;
//	}
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

// private static bool ParseBegin(ref Context context)
//
//	{
//	    var json = context.Json;
//	    SkipWhiteSpace(ref context);
//	    if (json[context.Index] == '}')
//	    {
//	        return false;
//	    }
//	    if (json[context.Index] == ',')
//	    {
//	        return false;
//	    }
//	    if (json[context.Index] == '"')
//	    {
//	        return false;
//	    }
//	    var state = DetermineValueType(ref context);
//	    switch (state)
//	    {
//	        case ValueType.Number:
//	        case ValueType.Bool:
//	        case ValueType.String:
//	        case ValueType.Null:
//	            break;
//	        case ValueType.Array:
//	            context.Stack[--context.StackIndex] = ValueType.Object;
//	            ++context.Index; // skip '['
//	            return true;
//	        case ValueType.Object:
//	            context.Stack[--context.StackIndex] = ValueType.Object;
//	            ++context.Index; // skip '{'
//	            return true;
//	    }
//	    return false;
//	}
func ParseBegin(context *JsonContext) bool {
	context.Index = 0
	context.StackIndex = len(context.Stack)

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

//     private static int ParseObjectBody(ref Context context, out Field outKey, out Field outValue)
//     {
//         var json = context.Json;
//         SkipWhiteSpace(ref context);
//         if (json[context.Index] == ',')
//         {
//             ++context.Index;
//             SkipWhiteSpace(ref context);
//         }
//         if (json[context.Index] == '}')
//         {
//             outKey = Field.Empty;
//             outValue = Field.Empty;
//             ++context.Index;
//             return 0;
//         }
//         if (json[context.Index] != '"')
//         {
//             // should be "
//             outKey = new Field(context.Index, 1, ValueType.Error);
//             outValue = Field.Empty;
//             return -1;
//         }
//         var key = GetString(ref context); // get object key string
//         if (!SkipWhiteSpaceUntil(ref context, ':'))
//         {
//             outKey = new Field(context.Index, 1, ValueType.Error);
//             outValue = Field.Empty;
//             return -1;
//         }
//         // skip ':'
//         ++context.Index;
//         outKey = key;
//         var state = DetermineValueType(ref context);
//         switch (state)
//         {
//             case ValueType.Number:
//                 outValue = ParseNumber(ref context);
//                 return 1;
//             case ValueType.Bool:
//                 outValue = ParseBoolean(ref context);
//                 return 1;
//             case ValueType.String:
//                 outValue = ParseString(ref context);
//                 return 1;
//             case ValueType.Null:
//                 outValue = ParseNull(ref context);
//                 return 1;
//             case ValueType.Array:
//                 if (context.StackIndex == 0)
//                 {
//                     outKey = new Field(context.Index, 1, ValueType.Error);
//                     outValue = Field.Empty;
//                     return -1;
//                 }
//                 context.Stack[--context.StackIndex] = ValueType.Array;
//                 outValue = new Field(context.Index, 1, ValueType.Array);
//                 ++context.Index; // skip '['
//                 return 2;
//             case ValueType.Object:
//                 if (context.StackIndex == 0)
//                 {
//                     outKey = new Field(context.Index, 1, ValueType.Error);
//                     outValue = Field.Empty;
//                     return -1;
//                 }
//                 context.Stack[--context.StackIndex] = ValueType.Object;
//                 outValue = new Field(context.Index, 1, ValueType.Object);
//                 ++context.Index; // skip '{'
//                 return 2;
//         }
//         outKey = new Field(context.Index, 1, ValueType.Error);
//         outValue = Field.Empty;
//         return -1;
//     }

func ParseObjectBody(c *JsonContext) (outKey JsonField, outValue JsonField, result int) {
	json := c.Json
	SkipWhiteSpace(c)
	if c.Index >= len(json) {
		return JsonFieldError, JsonFieldError, -1
	}

	if json[c.Index] == ',' {
		c.Index++
		SkipWhiteSpace(c)
	}

	if c.Index >= len(json) || json[c.Index] == '}' {
		c.Index++
		return JsonFieldEmpty, JsonFieldEmpty, 0
	}

	if json[c.Index] != '"' {
		// should be "
		key := NewJsonField(c.Index, 1, JsonValueTypeError)
		value := JsonFieldEmpty
		return key, value, -1
	}

	key := GetString(c) // get object key string
	if !SkipWhiteSpaceUntil(c, ':') {
		key := NewJsonField(c.Index, 1, JsonValueTypeError)
		value := JsonFieldEmpty
		return key, value, -1
	}

	c.Index++ // skip ':'
	outKey = key

	state := DetermineValueType(c)
	switch state {
	case JsonValueTypeNumber:
		outValue = ParseNumber(c)
		return outKey, outValue, 1
	case JsonValueTypeBool:
		outValue = ParseBoolean(c)
		return outKey, outValue, 1
	case JsonValueTypeString:
		outValue = ParseString(c)
		return outKey, outValue, 1
	case JsonValueTypeNull:
		outValue = ParseNull(c)
		return outKey, outValue, 1
	case JsonValueTypeArray:
		if c.StackIndex == 0 {
			outKey = NewJsonField(c.Index, 1, JsonValueTypeError)
			outValue = JsonFieldEmpty
			return outKey, outValue, -1
		}
		c.StackIndex--
		c.Stack[c.StackIndex] = JsonValueTypeArray
		outValue.Begin = c.Index + 1 // skip '['
		outValue.Length = 0          // will be set later when parsing the array body
		outValue.Type = JsonValueTypeArray
		c.Index++ // skip '['

		return outKey, outValue, 2
	case JsonValueTypeObject:
		if c.StackIndex == 0 {
			outKey = NewJsonField(c.Index, 1, JsonValueTypeError)
			outValue.Begin = c.Index + 1 // skip '{'
			outValue.Length = 0          // will be set later when parsing the object body
			outValue.Type = JsonValueTypeObject
			return outKey, outValue, -1
		}
		c.StackIndex--
		c.Stack[c.StackIndex] = JsonValueTypeObject
		outValue.Begin = c.Index + 1 // skip '{'
		outValue.Length = 0          // will be set later when parsing the object body
		outValue.Type = JsonValueTypeObject
		c.Index++ // skip '{'
		return outKey, outValue, 2
	default:
		outKey = NewJsonField(c.Index, 1, JsonValueTypeError)
		outValue = JsonFieldEmpty
		return outKey, outValue, -1
	}
}

//     /// <summary>
//     /// Parse JsonArray.
//     /// </summary>
//     private static int ParseArrayBody(ref Context context, out Field outValue)
//     {
//         var json = context.Json;
//         SkipWhiteSpace(ref context);
//         if (json[context.Index] == ',')
//         {
//             ++context.Index;
//             SkipWhiteSpace(ref context);
//         }
//         if (json[context.Index] == ']')
//         {
//             ++context.Index;
//             outValue = Field.Empty;
//             return 0;
//         }
//         var state = DetermineValueType(ref context);
//         switch (state)
//         {
//             case ValueType.Number:
//                 outValue = ParseNumber(ref context);
//                 return 1;
//             case ValueType.Bool:
//                 outValue = ParseBoolean(ref context);
//                 return 1;
//             case ValueType.String:
//                 outValue = ParseString(ref context);
//                 return 1;
//             case ValueType.Null:
//                 outValue = ParseNull(ref context);
//                 return 1;
//             case ValueType.Array:
//                 if (context.StackIndex == 0)
//                 {
//                     outValue = Field.Error;
//                     return -1;
//                 }
//                 context.Stack[--context.StackIndex] = ValueType.Array;
//                 outValue = new Field(context.Index, context.Index+1, ValueType.Array);
//                 ++context.Index; // skip '['
//                 return 2;
//             case ValueType.Object:
//                 if (context.StackIndex == 0)
//                 {
//                     outValue = Field.Error;
//                     return -1;
//                 }
//                 context.Stack[--context.StackIndex] = ValueType.Object;
//                 outValue = new Field(context.Index, context.Index+1, ValueType.Object);
//                 ++context.Index; // skip '{'
//                 return 2;
//         }
//         outValue = Field.Empty;
//         return -1;
//     }

func ParseArrayBody(c *JsonContext) (outValue JsonField, result int) {
	json := c.Json
	SkipWhiteSpace(c)
	if c.Index >= len(json) {
		return JsonFieldError, -1
	}

	if json[c.Index] == ',' {
		c.Index++
		SkipWhiteSpace(c)
	}

	if c.Index >= len(json) || json[c.Index] == ']' {
		c.Index++
		return JsonFieldEmpty, 0
	}

	state := DetermineValueType(c)
	switch state {
	case JsonValueTypeNumber:
		outValue = ParseNumber(c)
		return outValue, 1
	case JsonValueTypeBool:
		outValue = ParseBoolean(c)
		return outValue, 1
	case JsonValueTypeString:
		outValue = ParseString(c)
		return outValue, 1
	case JsonValueTypeNull:
		outValue = ParseNull(c)
		return outValue, 1
	case JsonValueTypeArray:
		if c.StackIndex == 0 {
			outValue = JsonFieldError
			return outValue, -1
		}
		c.StackIndex--
		c.Stack[c.StackIndex] = JsonValueTypeArray
		outValue.Begin = c.Index + 1 // skip '['
		outValue.Length = 0          // will be set later when parsing the array body
		outValue.Type = JsonValueTypeArray
		c.Index++ // skip '['

		return outValue, 2
	case JsonValueTypeObject:
		if c.StackIndex == 0 {
			outValue = JsonFieldError
			return outValue, -1
		}
		c.StackIndex--
		c.Stack[c.StackIndex] = JsonValueTypeObject
		outValue.Begin = c.Index + 1 // skip '{'
		outValue.Length = 0          // will be set later when parsing the object body
		outValue.Type = JsonValueTypeObject
		c.Index++ // skip '{'
		return outValue, 2
	default:
		outValue = JsonFieldEmpty
		return outValue, -1
	}
}

// private static Field ParseString(ref Context context)
//
//	{
//	    return GetString(ref context);
//	}
func ParseString(c *JsonContext) JsonField {
	return GetString(c)
}

// private static Field ParseNumber(ref Context context)
//
//	{
//	    var span = new Field(context.Index, context.Index, ValueType.Number);
//	    var json = context.Json;
//	    while (true)
//	    {
//	        switch (json[++context.Index])
//	        {
//	            case >= '0' and <= '9':
//	            case '-':
//	            case '+':
//	            case '.':
//	            case 'e':
//	            case 'E':
//	                continue;
//	        }
//	        break;
//	    }
//	    span.Length = context.Index - span.Begin;
//	    return span;
//	}
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

// private static Field ParseBoolean(ref Context context)
//
//	{
//	    var span = new Field() { Begin = context.Index, Length = 0, Type = ValueType.Bool };
//	    var json = context.Json;
//	    var asTrue = "true".AsSpan();
//	    if (asTrue.CompareTo(json.Slice(context.Index, 4), StringComparison.OrdinalIgnoreCase) == 0)
//	    {
//	        span.Length = 4;
//	        context.Index += 4;
//	        return span;
//	    }
//	    var asFalse = "false".AsSpan();
//	    if (asFalse.CompareTo(json.Slice(context.Index, 5), StringComparison.OrdinalIgnoreCase) == 0)
//	    {
//	        span.Length = 5;
//	        context.Index += 5;
//	        return span;
//	    }
//	    span.Type = ValueType.Error;
//	    return span;
//	}
func ParseBoolean(c *JsonContext) JsonField {
	span := JsonField{Begin: c.Index, Length: 0, Type: JsonValueTypeBool}
	json := c.Json

	asTrue := "true"
	if strings.EqualFold(string(json[c.Index:c.Index+4]), asTrue) {
		span.Length = 4
		c.Index += 4
		return span
	}

	asFalse := "false"
	if strings.EqualFold(string(json[c.Index:c.Index+5]), asFalse) {
		span.Length = 5
		c.Index += 5
		return span
	}

	span.Type = JsonValueTypeError
	return span
}

//     private static Field ParseNull(ref Context context)
//     {
//         var span = new Field() { Begin = context.Index, Length = 0, Type = ValueType.Null };
//         var json = context.Json;
//         var asNull = "null".AsSpan();
//         if (asNull.CompareTo(json.Slice(context.Index, 4), StringComparison.OrdinalIgnoreCase) == 0)
//         {
//             span.Length = 4;
//             context.Index += 4;
//             return span;
//         }
//         span.Type = ValueType.Error;
//         return span;
//     }

func ParseNull(c *JsonContext) JsonField {
	span := JsonField{Begin: c.Index, Length: 0, Type: JsonValueTypeNull}
	json := c.Json

	asNull := "null"
	if strings.EqualFold(string(json[c.Index:c.Index+4]), asNull) {
		span.Length = 4
		c.Index += 4
		return span
	}

	span.Type = JsonValueTypeError
	return span
}

//	    private static void SkipWhiteSpace(ref Context context)
//	    {
//	        var json = context.Json;
//	        while (true)
//	        {
//	            switch (json[context.Index])
//	            {
//	                case ' ' :
//	                case '\t':
//	                case '\n':
//	                case '\r':
//	                    ++context.Index;
//	                    continue;
//	            }
//		        // index point to non-whitespace
//		        break;
//		    }
//		}
func SkipWhiteSpace(c *JsonContext) {
	json := c.Json
	for c.Index < len(json) {
		switch json[c.Index] {
		case ' ', '\t', '\n', '\r':
			c.Index++
		default:
			return // index points to non-whitespace
		}
	}
}

// private static bool SkipWhiteSpaceUntil(ref Context context, char until)
//
//	{
//	    var json = context.Json;
//	    while (true)
//	    {
//	        switch (json[context.Index])
//	        {
//	            case ' ' :
//	            case '\t':
//	            case '\n':
//	            case '\r':
//	                ++context.Index;
//	                continue;
//	        }
//	        return json[context.Index] == until;
//	    }
//	}
func SkipWhiteSpaceUntil(c *JsonContext, until byte) bool {
	json := c.Json
	for c.Index < len(json) {
		switch json[c.Index] {
		case ' ', '\t', '\n', '\r':
			c.Index++
		default:
			return json[c.Index] == until
		}
	}
	return false
}

// private static Field GetString(ref Context context)
//
//	{
//	    // skip '"'
//	    var start = ++context.Index;
//	    var json = context.Json;
//	    while (true)
//	    {
//	        switch (json[context.Index++])
//	        {
//	            // check end '"'
//	            case '"':
//	                break;
//	            case '\\':
//	                // skip escaped quotes
//	                // the escape char may be '\"'，which will break while
//	                ++context.Index;
//	                continue;
//	            default:
//	                continue;
//	        }
//	        break;
//	    }
//	    // index after the string end '"' so -1
//	    return new Field(start, (context.Index - 1) - start, ValueType.String);
//	}
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

// private static string GetEscapedString(ref Context context, Field f)
//
//	{
//	    // skip '"'
//	    var str = new StringBuilder();
//	    var json = context.Json.Slice(f.Begin, f.Length);
//	    var index = 0;
//	    while (index < f.Length)
//	    {
//	        switch (json[index])
//	        {
//	            // check string end '"'
//	            case '"':
//	                index++;
//	                break;
//	            // check escaped char
//	            case '\\':
//	                {
//	                    char c;
//	                    index++;
//	                    switch (json[index++])
//	                    {
//	                        case '"':
//	                            c = '"';
//	                            break;
//	                        case '\\':
//	                            c = '\\';
//	                            break;
//	                        case '/':
//	                            c = '/';
//	                            break;
//	                        case '\'':
//	                            c = '\'';
//	                            break;
//	                        case 'b':
//	                            c = '\b';
//	                            break;
//	                        case 'f':
//	                            c = '\f';
//	                            break;
//	                        case 'n':
//	                            c = '\n';
//	                            break;
//	                        case 'r':
//	                            c = '\r';
//	                            break;
//	                        case 't':
//	                            c = '\t';
//	                            break;
//	                        case 'u':
//	                            c = GetUnicodeCodePoint(context, ref index);
//	                            break;
//	                        default:
//	                            // unsupported, just keep
//	                            continue;
//	                    }
//	                    str.Append(c);
//	                    continue;
//	                }
//	            default:
//	                // json[write++] = json[context.Index++];
//	                str.Append(json[index++]);
//	                continue;
//	        }
//	        break;
//	    }
//	    return str.ToString();
//	}
func GetEscapedString(context *JsonContext, f JsonField) string {
	str := strings.Builder{}
	json := context.Json[f.Begin : f.Begin+f.Length]
	index := 0

	for index < f.Length {
		switch json[index] {
		case '"':
			index++
			break
		case '\\':
			index++ // skip the escape character
			if index >= f.Length {
				return "" // error if we reach the end of the string
			}
			switch json[index] {
			case '"':
				str.WriteByte('"')
			case '\\':
				str.WriteByte('\\')
			case '/':
				str.WriteByte('/')
			case '\'':
				str.WriteByte('\'')
			case 'b':
				str.WriteByte('\b')
			case 'f':
				str.WriteByte('\f')
			case 'n':
				str.WriteByte('\n')
			case 'r':
				str.WriteByte('\r')
			case 't':
				str.WriteByte('\t')
			case 'u':
				c := GetUnicodeCodePoint(context, &index)
				str.WriteRune(c)
			default:
				continue // unsupported, just keep
			}
			index++
		default:
			str.WriteByte(json[index])
			index++
		}
	}

	return str.String()
}

//     private static char GetUnicodeCodePoint(Context context, ref int index)
//     {
//         var json = context.Json;
//         uint unicode = 0;
//         for (var i = 0; i < 4; ++i)
//         {
//             var c = json[index++];
//             var cp = c switch
//             {
//                 >= '0' and <= '9' => (byte)(c - '0'),
//                 >= 'A' and <= 'F' => (byte)(10 + (c - 'A')),
//                 >= 'a' and <= 'f' => (byte)(10 + (c - 'a')),
//                 _ => (0)
//             };
//             unicode = (uint)((uint)(unicode << 4) | (uint)(c & 0xF));
//         }
//         return (char)(unicode&0xFFFF);
//     }
// }

func GetUnicodeCodePoint(context *JsonContext, index *int) rune {
	json := context.Json
	var unicode uint32 = 0

	for i := 0; i < 4; i++ {
		if *index >= len(json) {
			return 0 // error if we reach the end of the string
		}
		c := json[*index]
		var cp byte
		switch {
		case c >= '0' && c <= '9':
			cp = byte(c - '0')
		case c >= 'A' && c <= 'F':
			cp = byte(10 + (c - 'A'))
		case c >= 'a' && c <= 'f':
			cp = byte(10 + (c - 'a'))
		default:
			return 0 // invalid character
		}
		unicode = (unicode << 4) | uint32(cp)
		*index++
	}

	return rune(unicode & 0xFFFF)
}
