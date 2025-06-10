package foundation

import (
	"strconv"
	"unicode/utf8"
)

// This custom StringBuilder is used to efficiently build a string using Write methods.
// The main purpose is to  minimize memory re-allocations and copies when building a string.
type StringBuilder struct {
	cursor int
	buf    []byte
	tmp    []byte
}

func NewStringBuilder() *StringBuilder {
	return &StringBuilder{
		cursor: 0,
		buf:    make([]byte, 8192),
		tmp:    make([]byte, 4),
	}
}

// String returns the accumulated string, but leaves the buffer owned by StringBuilder for reuse
func (b *StringBuilder) String() string {
	return string(b.buf[:b.cursor])
}

func (b *StringBuilder) Len() int { return b.cursor }
func (b *StringBuilder) Cap() int { return cap(b.buf) }

// Reset resets the [StringBuilder] to be empty.
func (b *StringBuilder) Reset() {
	b.cursor = 0
}

// grow copies the buffer to a new, larger buffer so that there
// are at least n bytes of capacity beyond len(b.buf).
func (b *StringBuilder) grow(n int) {
	newSize := 2*cap(b.buf) + n
	buf := make([]byte, newSize)
	copy(buf, b.buf)
	b.buf = buf
}

// Grow grows b's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to b
// without another allocation. If n is negative nothing happens.
func (b *StringBuilder) Grow(n int) {
	if n < 0 {
		return
	}
	if cap(b.buf)-b.cursor < n {
		b.grow(n)
	}
}

// Write appends the contents of p to b's buffer, and returns len(p)
func (b *StringBuilder) Write(p []byte) int {
	n := len(p)
	if cap(b.buf)-b.cursor < n {
		b.grow(n)
	}

	for _, s := range p {
		b.buf[b.cursor] = s
		b.cursor++
	}
	return len(p)
}

// WriteByte appends byte p to b's buffer
func (b *StringBuilder) WriteByte(p byte) int {
	if cap(b.buf)-b.cursor < 8 {
		b.grow(8)
	}

	b.buf[b.cursor] = p
	b.cursor++
	return 1
}

// WriteByteN appends the byte p to b's buffer n times.
func (b *StringBuilder) WriteByteN(p byte, n int) int {
	if cap(b.buf)-b.cursor < n {
		b.grow(n)
	}

	i := n
	for i > 0 {
		b.buf[b.cursor] = p
		b.cursor++
		i--
	}
	return n
}

// WriteFloat appends the string representation of f to b's buffer.
func (b *StringBuilder) WriteFloat(f float64, fmt byte, prec, bitSize int) int {
	b.tmp = b.tmp[:0]
	b.tmp = strconv.AppendFloat(b.tmp, f, fmt, prec, bitSize)
	b.Write(b.tmp)
	n := len(b.tmp)
	b.tmp = b.tmp[:4]
	return n
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to b's buffer.
func (b *StringBuilder) WriteRune(r rune) int {
	n := utf8.EncodeRune(b.tmp, r)
	if cap(b.buf)-b.cursor < n {
		b.grow(n)
	}
	for i := range n {
		b.buf[b.cursor] = b.tmp[i]
		b.cursor++
	}
	b.tmp = b.tmp[:4]
	return n
}

// WriteString appends the contents of s to b's buffer.
// It returns the length of s written to the buffer.
func (b *StringBuilder) WriteString(s string) int {
	n := len(s)
	if cap(b.buf)-b.cursor < n {
		b.grow(n)
	}
	for i := range n {
		b.buf[b.cursor] = s[i]
		b.cursor++
	}
	return n
}

// TODO:
// - implement WriteInt8, WriteInt16, WriteInt32, WriteInt64, WriteUint8, WriteUint16, WriteUint32, WriteUint64
// - add MapToString, SliceToString, Format, FormatComplex
