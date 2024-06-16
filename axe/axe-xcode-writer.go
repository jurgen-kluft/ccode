package axe

import (
	"fmt"
	"strings"
)

type LevelType int8

const (
	Object LevelType = iota
	Array
)

type XcodeWriter struct {
	Buffer        strings.Builder
	Levels        []int8
	NewlineNeeded bool
}

func NewXcodeWriter() *XcodeWriter {
	w := &XcodeWriter{}
	w.Buffer = strings.Builder{}
	w.Levels = make([]int8, 0)
	w.NewlineNeeded = true
	return w
}

func (w *XcodeWriter) StringBuilder() *strings.Builder {
	return &w.Buffer
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

func (o *ObjectScope) NewObjectScope() *ObjectScope {
	scope := &ObjectScope{o.W}
	o.W = nil
	return scope
}

func (w *XcodeWriter) NewObjectScope(name string) *ObjectScope {
	w.beginObject(name)
	return &ObjectScope{w}
}

func (w *XcodeWriter) beginObject(name string) {
	if len(name) > 0 {
		w.memberName(name)
	}
	w.Buffer.WriteString("{")
	w.Levels = append(w.Levels, int8(Object))
}

func (w *XcodeWriter) endObject() {
	w.newline(-1)
	w.Buffer.WriteString("}")

	if len(w.Levels) == 0 {
		panic("XCode XcodeWriter error endObject level")
	}
	if w.Levels[len(w.Levels)-1] != int8(Object) {
		panic("XCode XcodeWriter error endObject")
	}
	w.Levels = w.Levels[:len(w.Levels)-1]
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
	w.Buffer.WriteString("(")
	w.Levels = append(w.Levels, int8(Array))
}

func (w *XcodeWriter) endArray() {
	w.Buffer.WriteString(")")

	if len(w.Levels) == 0 {
		panic("XCode XcodeWriter error endArray level")
	}
	if w.Levels[len(w.Levels)-1] != int8(Array) {
		panic("XCode XcodeWriter error endArray")
	}
	w.Levels = w.Levels[:len(w.Levels)-1]
	w.writeTail()
}

// ------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------

func (w *XcodeWriter) commentBlock(s string) {
	w.Buffer.WriteString(" /* ")
	w.Buffer.WriteString(s)
	w.Buffer.WriteString(" */ ")
}

func (w *XcodeWriter) memberName(name string) {
	if len(w.Levels) == 0 || w.Levels[len(w.Levels)-1] != int8(Object) {
		panic("XCode XcodeWriter member must inside object scope")
	}

	w.newline(0)
	w.Buffer.WriteString(name)
	w.Buffer.WriteString(" = ")
}

func (w *XcodeWriter) member(name, value string) {
	w.memberName(name)
	w.write(value)
}

func (w *XcodeWriter) write(value string) {
	w.Buffer.WriteString(value)
}

func (w *XcodeWriter) writeTail() {
	if len(w.Levels) == 0 {
		return
	}
	if w.Levels[len(w.Levels)-1] == int8(Array) {
		w.Buffer.WriteString(",")
	}

	if w.Levels[len(w.Levels)-1] == int8(Object) {
		w.Buffer.WriteString(";")
	}
}

func (w *XcodeWriter) quoteString(v string) {
	w.Buffer.WriteString("\"")

	for _, ch := range v {
		if ch >= 0 && ch <= 0x1F {
			tmp := fmt.Sprintf("\\u%04x", ch)
			w.Buffer.WriteString(tmp)
			continue
		}

		switch ch {
		case '/':
			fallthrough
		case '\\':
			fallthrough
		case '"':
			w.Buffer.WriteString("\\")
			w.Buffer.WriteRune(ch)
		case '\b':
			w.Buffer.WriteString("\\b")
		case '\f':
			w.Buffer.WriteString("\\f")
		case '\n':
			w.Buffer.WriteString("\\n")
		case '\r':
			w.Buffer.WriteString("\\r")
		case '\t':
			w.Buffer.WriteString("\\t")
		default:
			w.Buffer.WriteRune(ch)
		}
	}
	w.Buffer.WriteString("\"")
}

func (w *XcodeWriter) newline(offset int) {
	if w.NewlineNeeded {
		w.Buffer.WriteString("\n")
		n := len(w.Levels) + offset
		for i := 0; i < n; i++ {
			w.Buffer.WriteString("  ")
		}
	}
}
