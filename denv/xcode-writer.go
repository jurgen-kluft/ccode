package denv

import (
	"fmt"

	"github.com/jurgen-kluft/ccode/foundation"
)

type LevelType int8

const (
	Object LevelType = iota
	Array
)

type XcodeWriter struct {
	buffer        *foundation.StringBuilder
	lines         []string
	levels        []int8
	newlineNeeded bool
}

func NewXcodeWriter() *XcodeWriter {
	w := &XcodeWriter{}
	w.buffer = foundation.NewStringBuilder()
	w.lines = make([]string, 0, 2048)
	w.levels = make([]int8, 0, 16)
	w.newlineNeeded = true
	return w
}

func (w *XcodeWriter) WriteToFile(filename string) error {
	w.finalize()
	return foundation.WriteLinesToFile(filename, w.lines)
}

// ------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------

type ObjectScope struct {
	W *XcodeWriter
}

func (o *ObjectScope) Close() {
	if o.W != nil {
		o.W.endObject()
	}
}

func (w *XcodeWriter) NewObjectScope(name string) *ObjectScope {
	if len(name) == 0 {
		w.buffer.WriteString("{")
		w.levels = append(w.levels, int8(Object))
	} else {
		w.beginObject(name)
	}
	return &ObjectScope{w}
}

func (w *XcodeWriter) beginObject(name string) {
	if len(name) > 0 {
		w.memberName(name)
	}
	w.buffer.WriteString("{")
	w.levels = append(w.levels, int8(Object))
}

func (w *XcodeWriter) endObject() {
	w.newline(-1)
	w.buffer.WriteString("}")

	if len(w.levels) == 0 {
		panic("XCode XcodeWriter error endObject level")
	}
	if w.levels[len(w.levels)-1] != int8(Object) {
		panic("XCode XcodeWriter error endObject")
	}
	w.levels = w.levels[:len(w.levels)-1]
	w.writeTail()
}

// ------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------

type ArrayScope struct {
	W *XcodeWriter
}

func (a ArrayScope) Close() {
	if a.W != nil {
		a.W.endArray()
	}
}

func (a ArrayScope) NewArrayScope() *ArrayScope {
	scope := &ArrayScope{a.W}
	a.W = nil
	return scope
}

func (w *XcodeWriter) NewArrayScope(name string) *ArrayScope {
	w.beginArray(name)
	return &ArrayScope{w}
}

func (w *XcodeWriter) beginArray(name string) {
	if len(name) > 0 {
		w.memberName(name)
	}
	w.buffer.WriteString("(")
	w.levels = append(w.levels, int8(Array))
}

func (w *XcodeWriter) endArray() {
	w.buffer.WriteString(")")

	if len(w.levels) == 0 {
		panic("XCode XcodeWriter error endArray level")
	}
	if w.levels[len(w.levels)-1] != int8(Array) {
		panic("XCode XcodeWriter error endArray")
	}
	w.levels = w.levels[:len(w.levels)-1]
	w.writeTail()
}

// ------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------

func (w *XcodeWriter) commentBlock(s string) {
	w.buffer.WriteString(" /* ")
	w.buffer.WriteString(s)
	w.buffer.WriteString(" */ ")
}

func (w *XcodeWriter) memberName(name string) {
	if len(w.levels) == 0 || w.levels[len(w.levels)-1] != int8(Object) {
		panic("XCode XcodeWriter member must inside object scope")
	}

	w.newline(0)
	w.buffer.WriteString(name)
	w.buffer.WriteString(" = ")
}

func (w *XcodeWriter) member(name, value string) {
	w.memberName(name)
	w.write(value)
}

func (w *XcodeWriter) write(value string) {
	w.buffer.WriteString(value)
	w.writeTail()
}

func (w *XcodeWriter) writeTail() {
	if len(w.levels) == 0 {
		return
	}
	if w.levels[len(w.levels)-1] == int8(Array) {
		w.buffer.WriteString(",")
	}

	if w.levels[len(w.levels)-1] == int8(Object) {
		w.buffer.WriteString(";")
	}
}

func (w *XcodeWriter) quoteString(v string) {
	w.buffer.WriteString("\"")

	for _, ch := range v {
		if ch >= 0 && ch <= 0x1F {
			tmp := fmt.Sprintf("\\u%04x", ch)
			w.buffer.WriteString(tmp)
			continue
		}

		switch ch {
		case '/':
			fallthrough
		case '\\':
			fallthrough
		case '"':
			w.buffer.WriteString("\\")
			w.buffer.WriteRune(ch)
		case '\b':
			w.buffer.WriteString("\\b")
		case '\f':
			w.buffer.WriteString("\\f")
		case '\n':
			w.buffer.WriteString("\\n")
		case '\r':
			w.buffer.WriteString("\\r")
		case '\t':
			w.buffer.WriteString("\\t")
		default:
			w.buffer.WriteRune(ch)
		}
	}
	w.buffer.WriteString("\"")
}

func (w *XcodeWriter) finalize() {
	w.newlineNeeded = true
	w.newline(0)
}

func (w *XcodeWriter) newline(offset int) {
	if w.newlineNeeded {

		w.lines = append(w.lines, w.buffer.String())
		w.buffer.Reset()

		n := len(w.levels) + offset
		for i := 0; i < n; i++ {
			w.buffer.WriteString("  ")
		}
	}
}
