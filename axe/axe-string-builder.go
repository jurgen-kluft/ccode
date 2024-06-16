package axe

import (
	"unicode/utf8"
)

// This custom stringBuilder is used to efficiently build a string using Write methods.
// The main purpose is to  minimize memory re-allocations and copies when building a string.
type stringBuilder struct {
	cursor int
	buf    []byte
	tmp    []byte
}

func NewStringBuilder() *stringBuilder {
	return &stringBuilder{
		cursor: 0,
		buf:    make([]byte, 8192),
		tmp:    make([]byte, 4),
	}
}

// String returns the accumulated string, but leaves the buffer owned by stringBuilder for reuse
func (b *stringBuilder) String() string {
	return string(b.buf[:b.cursor])
}

func (b *stringBuilder) Len() int { return b.cursor }
func (b *stringBuilder) Cap() int { return cap(b.buf) }

// Reset resets the [stringBuilder] to be empty.
func (b *stringBuilder) Reset() {
	b.cursor = 0
}

// grow copies the buffer to a new, larger buffer so that there
// are at least n bytes of capacity beyond len(b.buf).
func (b *stringBuilder) grow(n int) {
	newSize := 2*cap(b.buf) + n
	buf := make([]byte, newSize)
	copy(buf, b.buf)
	b.buf = buf
}

// Grow grows b's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to b
// without another allocation. If n is negative nothing happens.
func (b *stringBuilder) Grow(n int) {
	if n < 0 {
		return
	}
	if cap(b.buf)-b.cursor < n {
		b.grow(n)
	}
}

// Write appends the contents of p to b's buffer.
// Write always returns len(p)
func (b *stringBuilder) Write(p []byte) int {
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

// WriteRune appends the UTF-8 encoding of Unicode code point r to b's buffer.
// It returns the length of r and a nil error.
func (b *stringBuilder) WriteRune(r rune) int {
	n := utf8.EncodeRune(b.tmp, r)
	if cap(b.buf)-b.cursor < n {
		b.grow(n)
	}

	for i := 0; i < n; i++ {
		b.buf = append(b.buf, b.tmp[i])
	}
	b.cursor += n
	return len(b.buf) - n
}

// WriteString appends the contents of s to b's buffer.
// It returns the length of s written to the buffer.
func (b *stringBuilder) WriteString(s string) int {
	n := len(s)
	if cap(b.buf)-b.cursor < n {
		b.grow(n)
	}
	for i := 0; i < n; i++ {
		b.buf[b.cursor] = s[i]
		b.cursor++
	}
	return n
}
