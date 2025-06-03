package dpenc

import (
	"encoding/binary"
	"fmt"
	"os"
)

type BinaryData struct {
	Data   []byte // The binary data buffer
	Cursor int    // Current position in the buffer
}

func NewBinaryData(size int) *BinaryData {
	return &BinaryData{
		Data:   make([]byte, size),
		Cursor: 0,
	}
}

func (d *BinaryData) reset() {
	d.Cursor = 0
}

func (d *BinaryData) writeToFile(f *os.File) error {
	n, err := f.Write(d.Data[:d.Cursor])
	if err == nil && n < d.Cursor {
		return fmt.Errorf("short write: %d bytes written, expected %d", n, d.Cursor)
	}
	return err
}

func (d *BinaryData) readFromFile(numBytes int, f *os.File) error {
	d.Cursor = 0
	if numBytes <= 0 || numBytes > len(d.Data) {
		return fmt.Errorf("invalid number of bytes to read: %d", numBytes)
	}
	n, err := f.Read(d.Data[:numBytes])
	if err != nil {
		return err
	}
	if n < numBytes {
		return fmt.Errorf("short read: %d bytes read, expected %d", n, numBytes)
	}
	return nil
}

func (d *BinaryData) writeBytes(b []byte) {
	if d.Cursor+len(b) <= len(d.Data) {
		d.Cursor += copy(d.Data[d.Cursor:], b)
	}
}

func (d *BinaryData) writeInt(value int) {
	if d.Cursor+4 <= len(d.Data) {
		binary.LittleEndian.PutUint32(d.Data[d.Cursor:d.Cursor+4], uint32(value))
		d.Cursor += 4
	}
}

func (d *BinaryData) writeInt32(value int32) {
	if d.Cursor+4 <= len(d.Data) {
		binary.LittleEndian.PutUint32(d.Data[d.Cursor:d.Cursor+4], uint32(value))
		d.Cursor += 4
	}
}

func (d *BinaryData) readNBytes(n int) []byte {
	if d.Cursor+n > len(d.Data) {
		return nil // or handle error
	}
	value := d.Data[d.Cursor : d.Cursor+n]
	d.Cursor += n
	return value
}

func (d *BinaryData) readInt() int {
	if d.Cursor+4 > len(d.Data) {
		return 0 // or handle error
	}
	value := binary.LittleEndian.Uint32(d.Data[d.Cursor : d.Cursor+4])
	d.Cursor += 4
	return int(value)
}

func (d *BinaryData) readInt32() int32 {
	if d.Cursor+4 > len(d.Data) {
		return 0 // or handle error
	}
	value := binary.LittleEndian.Uint32(d.Data[d.Cursor : d.Cursor+4])
	d.Cursor += 4
	return int32(value)
}
