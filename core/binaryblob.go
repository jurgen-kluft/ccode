package corepkg

import (
	"encoding/binary"
)

type BinaryBlob struct {
	data   []byte // The binary data buffer
	cursor int    // Current position in the buffer
}

func NewBinaryBlob(size int) *BinaryBlob {
	return &BinaryBlob{
		data:   make([]byte, size),
		cursor: 0,
	}
}

func BinaryBlobFromData(data []byte) *BinaryBlob {
	return &BinaryBlob{
		data:   data,
		cursor: 0,
	}
}

func (d *BinaryBlob) Data() []byte {
	return d.data[:d.cursor]
}

func (d *BinaryBlob) ResetCursor() {
	d.cursor = 0
}

func (d *BinaryBlob) WriteBytes(b []byte) {
	if d.cursor+len(b) <= len(d.data) {
		d.cursor += copy(d.data[d.cursor:], b)
	}
}

func (d *BinaryBlob) WriteInt(value int) {
	if d.cursor+4 <= len(d.data) {
		binary.LittleEndian.PutUint32(d.data[d.cursor:d.cursor+4], uint32(value))
		d.cursor += 4
	}
}

func (d *BinaryBlob) WriteInt32(value int32) {
	if d.cursor+4 <= len(d.data) {
		binary.LittleEndian.PutUint32(d.data[d.cursor:d.cursor+4], uint32(value))
		d.cursor += 4
	}
}

func (d *BinaryBlob) ReadNBytes(n int) []byte {
	if d.cursor+n > len(d.data) {
		return nil // or handle error
	}
	value := d.data[d.cursor : d.cursor+n]
	d.cursor += n
	return value
}

func (d *BinaryBlob) ReadInt() int {
	if d.cursor+4 > len(d.data) {
		return 0 // or handle error
	}
	value := binary.LittleEndian.Uint32(d.data[d.cursor : d.cursor+4])
	d.cursor += 4
	return int(value)
}

func (d *BinaryBlob) ReadInt32() int32 {
	if d.cursor+4 > len(d.data) {
		return 0 // or handle error
	}
	value := binary.LittleEndian.Uint32(d.data[d.cursor : d.cursor+4])
	d.cursor += 4
	return int32(value)
}
