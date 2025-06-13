package foundation

import (
	"strings"
)

type JsonEncoder struct {
	indentation string // precomputed indentation string
	indentCount int
	indentLevel int
	stack       []int8
	stackIndex  int
	json        *StringBuilder
}

func NewJsonEncoder(indentation string) *JsonEncoder {
	maxIndentation := 64                                          // maximum indentation length
	setIndentation := strings.Repeat(indentation, maxIndentation) // spaces
	return &JsonEncoder{
		indentation: setIndentation,
		indentCount: len(indentation),
		indentLevel: 0,
		stack:       make([]int8, 64),
		stackIndex:  0,
		json:        NewStringBuilder(),
	}
}

func (e *JsonEncoder) Begin() {
	e.stackIndex = len(e.stack)
	e.stack = append(e.stack, 0)
	e.json.Reset()
}

func (e *JsonEncoder) End() string {
	return e.json.String()
}

func (e *JsonEncoder) BeginObject(key string) {
	e.incrementField()
	if key == "" {
		e.writeIndentedLine("{")
	} else {
		e.writeIndentedLine("\"", key, "\": {")
	}
	e.indent(1)
}

func (e *JsonEncoder) EndObject() {
	e.indent(-1)
	e.writeIndented("}")
}

func (e *JsonEncoder) BeginMap(key string) {
	e.BeginObject(key)
}

func (e *JsonEncoder) EndMap() {
	e.EndObject()
}

func (e *JsonEncoder) BeginArray(key string) {
	e.incrementField()
	if key == "" {
		e.writeIndentedLine("[")
	} else {
		e.writeIndentedLine("\"", key, "\": [")
	}
	e.indent(1)
}

func (e *JsonEncoder) EndArray() {
	e.indent(-1)
	e.writeIndented("]")
}

// WriteString writes a string value to the JSON output, for '"key": "value"' it requires the key already to be written.
func (e *JsonEncoder) WriteString(value string) {
	e.writeQuotedString("\"", escapeString(value))
}

func (e *JsonEncoder) WriteFieldString(key, value string) {
	e.incrementField()
	e.writeIndented("\"", key, "\": ")
	e.writeQuotedString("\"", escapeString(value))
}

func (e *JsonEncoder) WriteArrayElement(value any) {
	e.incrementField()

	// Figure out the type of value
	switch v := value.(type) {
	case string:
		e.writeIndented("\"", escapeString(v), "\"")
	case int8:
		e.writeInt(int64(v), 10)
	case int16:
		e.writeInt(int64(v), 10)
	case int32:
		e.writeInt(int64(v), 10)
	case int:
		e.writeInt(int64(v), 10)
	case int64:
		e.writeInt(v, 10)
	case uint8:
		e.writeInt(int64(v), 10)
	case uint16:
		e.writeInt(int64(v), 10)
	case uint32:
		e.writeInt(int64(v), 10)
	case uint:
		e.writeInt(int64(v), 10)
	case uint64:
		e.writeInt(int64(v), 10)
	case float32:
		e.writeFloat32(v)
	case float64:
		e.writeFloat64(v)
	case bool:
		e.writeBool(v)
	}
}

func (e *JsonEncoder) WriteMapElement(key string, value any) {
	e.WriteField(key, value)
}

func (e *JsonEncoder) WriteFieldInt32(key string, value int) {
	e.incrementField()
	e.writeIndented("\"", key, "\": ")
	e.writeInt(int64(value), 10)
}

func (e *JsonEncoder) WriteFieldInt64(key string, value int64) {
	e.incrementField()
	e.writeIndented("\"", key, "\": ")
	e.writeInt(value, 10)
}

func (e *JsonEncoder) WriteFieldFloat32(key string, value float32) {
	e.incrementField()
	e.writeIndented("\"", key, "\": ")
	e.writeFloat32(value)
}

func (e *JsonEncoder) WriteFieldFloat64(key string, value float64) {
	e.incrementField()
	e.writeIndented("\"", key, "\": ")
	e.writeFloat64(value)
}

func (e *JsonEncoder) WriteFieldBool(key string, value bool) {
	e.incrementField()
	e.writeIndented("\"", key, "\": ")
	e.writeBool(value)
}

func (e *JsonEncoder) WriteField(key string, value any) {
	e.incrementField()
	e.writeIndented("\"", key, "\": ")

	// Figure out the type of value
	switch v := value.(type) {
	case string:
		e.writeQuotedString("\"", escapeString(v))
	case int8:
		e.writeInt(int64(v), 10)
	case int16:
		e.writeInt(int64(v), 10)
	case int32:
		e.writeInt(int64(v), 10)
	case int:
		e.writeInt(int64(v), 10)
	case int64:
		e.writeInt(v, 10)
	case uint8:
		e.writeInt(int64(v), 10)
	case uint16:
		e.writeInt(int64(v), 10)
	case uint32:
		e.writeInt(int64(v), 10)
	case uint:
		e.writeInt(int64(v), 10)
	case uint64:
		e.writeInt(int64(v), 10)
	case float32:
		e.writeFloat32(v)
	case float64:
		e.writeFloat64(v)
	case bool:
		e.writeBool(v)
	}

}

func escapeString(str string) string {
	var sb strings.Builder
	for _, c := range str {
		switch c {
		case '"':
			sb.WriteString("\\\"")
		case '\\':
			sb.WriteString("\\\\")
		case '\b':
			sb.WriteString("\\b")
		case '\f':
			sb.WriteString("\\f")
		case '\n':
			sb.WriteString("\\n")
		case '\r':
			sb.WriteString("\\r")
		case '\t':
			sb.WriteString("\\t")
		default:
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

func (e *JsonEncoder) indent(level int) {
	if level > 0 {
		e.indentLevel++
		e.stackIndex--
		e.stack[e.stackIndex] = 0
	} else {
		e.json.WriteLn("")
		e.indentLevel--
		e.stackIndex++
	}
}

func (e *JsonEncoder) incrementField() {
	if e.stack[e.stackIndex] != 0 {
		e.json.WriteLn(",")
	}
	e.stack[e.stackIndex] = 1
}

func (e *JsonEncoder) writeIndented(str ...string) {
	e.json.WriteString(e.indentation[:e.indentLevel*e.indentCount])
	for _, s := range str {
		e.json.WriteString(s)
	}
}

func (e *JsonEncoder) writeIndentedLine(str ...string) {
	e.json.WriteString(e.indentation[:e.indentLevel*e.indentCount])
	for _, s := range str {
		e.json.WriteString(s)
	}
	e.json.WriteLn("")
}

func (e *JsonEncoder) writeString(value string) {
	e.json.WriteString(value)
}

func (e *JsonEncoder) writeQuotedString(quote, value string) {
	e.json.WriteString(quote)
	e.json.WriteString(value)
	e.json.WriteString(quote)
}

func (e *JsonEncoder) writeInt(value int64, base int) {
	e.json.WriteInt(value, base)
}

func (e *JsonEncoder) writeFloat32(value float32) {
	e.json.WriteFloat(float64(value), 'f', -1, 32)
}

func (e *JsonEncoder) writeFloat64(value float64) {
	e.json.WriteFloat(value, 'f', -1, 64)
}

func (e *JsonEncoder) writeBool(value bool) {
	if value {
		e.json.WriteString("true")
	} else {
		e.json.WriteString("false")
	}
}
