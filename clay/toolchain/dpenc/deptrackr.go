package dpenc

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"slices"
	"unsafe"
)

// Golang prototype for a dependency tracker database, the core is the shard structure.
// Mostly this is about adding items to the database and should be very efficient to implement
// in a language like C or C++ where we have virtual memory and even mmapped file IO.
// A database loaded from disk is always read-only, we do not load, modify and save. Instead
// we load the database, then create a new one and build it anew utilizing the old one. This
// is a lot faster and simpler than loading, modifying and saving.
//
// TODO One thing that has not been implemented yet is the case when a shard gets full. We
//      can implement this by making the first int32 of a shard being an offset to the next
//      shard, so we can have a linked list of shards. This way we can have an unlimited
//      number of items in a shard, however performance will degrade when searching for a
//      hash in a shard that has many linked shards. This can only occur when the hash
//      keeps hitting this shard, so it is not a very common case.
//
// e.g.
// currentTrackr := deptrackr.Load("path/to/storage")
// futureTrackr := currentTrackr.NewDB() // create a new database based on the current one
//
// // Query item
// state, err := currentTrackr.QueryItem(itemHash, true, verifyCb)
// if state == deptrackr.StateUpToDate {
//     currentTrackr.CopyItem(futureTrackr, itemHash) // copy the item to the new database
// } else {
//     // Build a new item and add it to the new database
// }
//
// futureTrackr.Save()
//

type State int8

const (
	StateNone      State = 0
	StateUpToDate  State = 1
	StateOutOfDate State = 2
	StateError     State = 3
)

type trackr struct {
	hasher               hash.Hash
	scratchBuffer        []byte  // A temporary byte buffer for hashing and other operations, not saved to disk
	scratchBufferCursor  int     // Cursor for the scratch byte buffer, used to write data to it
	StoragePath          string  // Path to the directory where we store the database file
	Signature            string  // max 32 characters signature, e.g. ".d deptracker v1.0.0"
	HashSize             int32   // Size of the hash, this is 20 bytes for SHA1
	ItemState            []State // Hash, this is the state of the item that is modified during a query
	ItemIdHash           []byte  // Hash, this is the ID of the item (filepath, label (e.g. 'MSVC C++ compiler cmd-line arguments))
	ItemChangeHash       []byte  // Hash, this identifies the 'change' (modification-time, file-size, file-content, command-line arguments, string, etc..)
	ItemIdFlags          []uint8 //
	ItemChangeFlags      []uint8 //
	ItemDepsStart        []int32 // Item, start of dependencies
	ItemDepsCount        []int32 // Item, count of dependencies
	ItemIdDataOffset     []int32 // data for Id, 4 bytes (0 means no data)
	ItemIdDataSize       []int32 //
	ItemChangeDataOffset []int32 // data for Change, 4 bytes (0 means no data)
	ItemChangeDataSize   []int32 //
	Deps                 []int32 // Array for each item to list their dependencies, this is a flat array of item indices
	Data                 []byte  // Here the data for Id and Change is stored, this is a flat array of bytes
	N                    int32   // how many bits we take from the hash to index into the shards (0-15)
	S                    int32   // size of a shard, this is the number of items per shard, default is 512
	ShardOffsets         []int32 // This is an array of offsets into the Shards array, each offset corresponds to a shard, 0 means the shard doesn't exist yet
	ShardSizes           []int32 // This is an array of sizes of each shard, 0 means the shard is empty
	DirtyFlags           []uint8 // A bit per shard, indicates if the shard is dirty (unsorted) and needs to be sorted
	Shards               []int32 // An array of shards, a shard is a region of item-indices
	EmptyShard           []int32 // A shard initialized to a size of S and full of NillIndex
}

func newTrackr(storageDir string, signature string) *trackr {
	hs := int32(20) // SHA1 hash size is 20 bytes
	n := int32(10)  // how many bits we take from the hash to index into the buckets (0-15)
	s := int32(512) // size of a shard, this is the number of items per shard, default is 512

	reserve := 32

	d := &trackr{
		hasher:               sha1.New(),                       // Create a new SHA1 hasher
		scratchBuffer:        make([]byte, 1024),               // A temporary byte buffer for hashing and other operations, not saved to disk
		scratchBufferCursor:  0,                                // Cursor for the scratch byte buffer, used to write data to it
		StoragePath:          storageDir,                       //
		Signature:            signature,                        // copy the signature from the source
		HashSize:             hs,                               //
		ItemState:            make([]State, 0, reserve),        // State of the item, this is used to track if the item is up to date or not
		ItemIdHash:           make([]byte, 0, reserve*int(hs)), //
		ItemChangeHash:       make([]byte, 0, reserve*int(hs)), //
		ItemIdFlags:          make([]uint8, 0, reserve),        //
		ItemChangeFlags:      make([]uint8, 0, reserve),        //
		ItemDepsStart:        make([]int32, 0, reserve),        //
		ItemDepsCount:        make([]int32, 0, reserve),        //
		ItemIdDataOffset:     make([]int32, 0, reserve),        //
		ItemIdDataSize:       make([]int32, 0, reserve),        //
		ItemChangeDataOffset: make([]int32, 0, reserve),        //
		ItemChangeDataSize:   make([]int32, 0, reserve),        //
		N:                    n,                                //
		S:                    s,                                //
		ShardOffsets:         make([]int32, 1<<n),              // initial capacity of 2^N shards
		ShardSizes:           make([]int32, 1<<n),              // initial capacity of 2^N shards
		DirtyFlags:           make([]uint8, ((1<<n)+7)>>3),     // initial capacity of N bits for dirty flags (rounded up to the nearest byte)
		Shards:               make([]int32, 0, (1<<n)*s),       // initial capacity of 2^N shards where each shard has s elements
		EmptyShard:           make([]int32, s),                 // an empty shard, used to copy when making a new shard in Shards
		Deps:                 make([]int32, 0, reserve),        // Array for each item to list their dependencies, this is a flat array of item indices
		Data:                 make([]byte, 0, reserve),         // Data for Id and Change, this is a flat array of bytes
	}

	// Set the first 8 elements of EmptyShard to -1
	pattern := []int32{NilIndex, NilIndex, NilIndex, NilIndex, NilIndex, NilIndex, NilIndex, NilIndex}
	copy(d.EmptyShard, pattern)
	copy(d.ShardOffsets, pattern)

	// Fill the rest of the shard with the pattern, increasing the size of the pattern
	for j := len(pattern); j < len(d.EmptyShard); j *= 2 {
		copy(d.EmptyShard[j:], d.EmptyShard[:j])
	}

	for j := len(pattern); j < len(d.ShardOffsets); j *= 2 {
		copy(d.ShardOffsets[j:], d.ShardOffsets[:j])
	}

	return d
}

func (src *trackr) newTrackr() *trackr {
	reserve := max(1024, len(src.ItemIdHash)+len(src.ItemIdHash)/8)

	hs := src.HashSize
	n := src.N
	s := src.S

	newTrackr := &trackr{
		hasher:               sha1.New(),                                         // Create a new SHA1 hasher
		scratchBuffer:        make([]byte, 1024),                                 // A temporary byte buffer for hashing and other operations, not saved to disk
		scratchBufferCursor:  0,                                                  // Cursor for the scratch byte buffer, used to write data to it
		StoragePath:          src.StoragePath,                                    //
		Signature:            src.Signature,                                      // copy the signature from the source
		HashSize:             hs,                                                 //
		ItemState:            make([]State, 0, reserve),                          // State of the item, this is used to track if the item is up to date or not
		ItemIdHash:           make([]byte, 0, reserve*int(hs)),                   //
		ItemChangeHash:       make([]byte, 0, reserve*int(hs)),                   //
		ItemIdFlags:          make([]uint8, 0, reserve),                          //
		ItemChangeFlags:      make([]uint8, 0, reserve),                          //
		ItemDepsStart:        make([]int32, 0, reserve),                          //
		ItemDepsCount:        make([]int32, 0, reserve),                          //
		ItemIdDataOffset:     make([]int32, 0, reserve),                          //
		ItemIdDataSize:       make([]int32, 0, reserve),                          //
		ItemChangeDataOffset: make([]int32, 0, reserve),                          //
		ItemChangeDataSize:   make([]int32, 0, reserve),                          //
		N:                    n,                                                  //
		S:                    s,                                                  //
		ShardOffsets:         make([]int32, 1<<n),                                // initial capacity of 2^N shards
		ShardSizes:           make([]int32, 1<<n),                                // initial capacity of 2^N shards
		DirtyFlags:           make([]uint8, ((1<<n)+7)>>3),                       // initial capacity of N bits for dirty flags (rounded up to the nearest byte)
		Shards:               make([]int32, 0, (1<<n)*s),                         // initial capacity of 2^N shards where each shard has s elements
		EmptyShard:           src.EmptyShard,                                     // an empty shard, used to copy when making a new shard in Shards
		Deps:                 make([]int32, 0, len(src.Deps)+(len(src.Deps)/10)), //
		Data:                 make([]byte, 0, len(src.Data)+(len(src.Data)/10)),  //
	}

	pattern := []int32{NilIndex, NilIndex, NilIndex, NilIndex, NilIndex, NilIndex, NilIndex, NilIndex}
	copy(newTrackr.ShardOffsets, pattern)

	// Fill the rest of the shard with the pattern, increasing the size of the pattern
	for j := len(pattern); j < len(newTrackr.ShardOffsets); j *= 2 {
		copy(newTrackr.ShardOffsets[j:], newTrackr.ShardOffsets[:j])
	}

	return newTrackr
}

// --------------------------------------------------------------------------
// Helper functions to cast loaded byte arrays to other types
func castInt32ArrayToByteArray(i []int32) []byte {
	if len(i) == 0 {
		return []byte{}
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&i[0])), len(i)<<2)
}

func (d *trackr) writeBytesToScratch(b []byte) {
	d.scratchBufferCursor += copy(d.scratchBuffer[d.scratchBufferCursor:], b)
}

func (d *trackr) writeIntToScratch(value int) {
	binary.LittleEndian.PutUint32(d.scratchBuffer[d.scratchBufferCursor:d.scratchBufferCursor+4], uint32(value))
	d.scratchBufferCursor += 4
}

func (d *trackr) writeInt32ToScratch(value int32) {
	binary.LittleEndian.PutUint32(d.scratchBuffer[d.scratchBufferCursor:d.scratchBufferCursor+4], uint32(value))
	d.scratchBufferCursor += 4
}

func (d *trackr) readNBytesFromScratch(n int) []byte {
	if d.scratchBufferCursor+n > len(d.scratchBuffer) {
		return nil // or handle error
	}
	value := d.scratchBuffer[d.scratchBufferCursor : d.scratchBufferCursor+n]
	d.scratchBufferCursor += n
	return value
}

func (d *trackr) readIntFromScratch() int {
	if d.scratchBufferCursor+4 > len(d.scratchBuffer) {
		return 0 // or handle error
	}
	value := binary.LittleEndian.Uint32(d.scratchBuffer[d.scratchBufferCursor : d.scratchBufferCursor+4])
	d.scratchBufferCursor += 4
	return int(value)
}

func (d *trackr) readInt32FromScratch() int32 {
	if d.scratchBufferCursor+4 > len(d.scratchBuffer) {
		return 0 // or handle error
	}
	value := binary.LittleEndian.Uint32(d.scratchBuffer[d.scratchBufferCursor : d.scratchBufferCursor+4])
	d.scratchBufferCursor += 4
	return int32(value)
}

func (d *trackr) writeByteArray(signature string, numItems int, array []byte, sizeofElement int, f *os.File) error {
	if (numItems * sizeofElement) != len(array) {
		return fmt.Errorf("number of items*sizeof(item) (%d) does not match the length of the array (%d)", numItems*sizeofElement, len(array))
	}

	d.hasher.Reset()
	d.hasher.Write([]byte(signature))
	signatureHash := d.hasher.Sum(nil)

	d.scratchBufferCursor = 0
	d.writeBytesToScratch(signatureHash[:10])
	d.writeIntToScratch(numItems)
	d.writeIntToScratch(sizeofElement)
	d.writeIntToScratch(len(array))
	d.writeBytesToScratch(signatureHash[10:])
	if numBytesWritten, err := f.Write(d.scratchBuffer[:d.scratchBufferCursor]); err != nil || numBytesWritten != d.scratchBufferCursor {
		return fmt.Errorf("failed to write header for signature '%s': %w", signature, err)
	}
	if numBytesWritten, err := f.Write(array); err != nil || numBytesWritten != len(array) {
		return fmt.Errorf("failed to write byte array for signature '%s': %w", signature, err)
	}
	return nil
}

func (d *trackr) readByteArray(signature string, numItems int, itemSize int, f *os.File) ([]byte, error) {
	d.hasher.Reset()
	d.hasher.Write([]byte(signature))
	signatureHash := d.hasher.Sum(nil)

	headerSize := 32

	d.scratchBufferCursor = 0
	if numBytesRead, err := f.Read(d.scratchBuffer[:headerSize]); err != nil || numBytesRead != headerSize {
		return nil, fmt.Errorf("failed to read header for signature '%s': %w", signature, err)
	}

	readSignatureHash := d.readNBytesFromScratch(10)
	if bytes.Compare(signatureHash[:10], readSignatureHash) != 0 {
		return nil, fmt.Errorf("signature mismatch for '%s'", signature)
	}

	readNumItems := d.readIntFromScratch()
	if numItems != 0 && readNumItems != numItems {
		return nil, fmt.Errorf("number of items (%d) does not match the expected number (%d)", readNumItems, numItems)
	}

	sizeOfElement := d.readIntFromScratch()
	if sizeOfElement <= 0 || sizeOfElement != itemSize {
		return nil, fmt.Errorf("expected size of element to be greater than 0, got %d", sizeOfElement)
	}

	totalSize := d.readIntFromScratch()
	if totalSize != readNumItems*sizeOfElement {
		return nil, fmt.Errorf("total size (%d) does not match the expected size (%d)", totalSize, numItems*sizeOfElement)
	}

	readSignatureHash = d.readNBytesFromScratch(10)
	if bytes.Compare(signatureHash[10:], readSignatureHash) != 0 {
		return nil, fmt.Errorf("signature mismatch for '%s'", signature)
	}

	arrayData := make([]byte, totalSize)
	if numBytesRead, err := f.Read(arrayData); err != nil || numBytesRead != totalSize {
		return nil, fmt.Errorf("failed to read byte array for signature '%s': %w", signature, err)
	}

	return arrayData, nil
}

func (d *trackr) writeInt32Array(signature string, numItems int, array []int32, f *os.File) error {

	d.hasher.Reset()
	d.hasher.Write([]byte(signature))
	signatureHash := d.hasher.Sum(nil)

	d.scratchBufferCursor = 0
	d.writeBytesToScratch(signatureHash[:10])
	d.writeIntToScratch(numItems)
	d.writeIntToScratch(4)              // size of int32 is 4 bytes
	d.writeIntToScratch(len(array) * 4) // total size of the array in bytes
	d.writeBytesToScratch(signatureHash[10:])
	if numBytesWritten, err := f.Write(d.scratchBuffer[:d.scratchBufferCursor]); err != nil || numBytesWritten != d.scratchBufferCursor {
		return fmt.Errorf("failed to write header for signature '%s': %w", signature, err)
	}
	arrayBytes := castInt32ArrayToByteArray(array)
	if numBytesWritten, err := f.Write(arrayBytes); err != nil || numBytesWritten != len(arrayBytes) {
		return fmt.Errorf("failed to write int32 array for signature '%s': %w", signature, err)
	}
	return nil
}

func (d *trackr) readInt32Array(signature string, numItems int, f *os.File) ([]int32, error) {
	d.hasher.Reset()
	d.hasher.Write([]byte(signature))
	signatureHash := d.hasher.Sum(nil)

	headerSize := 32

	d.scratchBufferCursor = 0
	if numBytesRead, err := f.Read(d.scratchBuffer[:headerSize]); err != nil || numBytesRead != headerSize {
		return nil, fmt.Errorf("failed to read header for signature '%s': %w", signature, err)
	}

	readSignatureHash := d.readNBytesFromScratch(10)
	if bytes.Compare(signatureHash[:10], readSignatureHash) != 0 {
		return nil, fmt.Errorf("signature mismatch for '%s'", signature)
	}

	readNumElements := d.readIntFromScratch()
	if numItems != 0 && numItems != readNumElements {
		return nil, fmt.Errorf("number of items (%d) does not match the expected number (%d)", readNumElements, numItems)
	}

	sizeOfElement := d.readIntFromScratch()
	if sizeOfElement != 4 {
		return nil, fmt.Errorf("expected size of element to be 4 bytes, got %d", sizeOfElement)
	}

	totalSize := d.readIntFromScratch()
	if totalSize != readNumElements*sizeOfElement {
		return nil, fmt.Errorf("total size (%d) does not match the expected size (%d)", totalSize, numItems*sizeOfElement)
	}

	readSignatureHash = d.readNBytesFromScratch(10)
	if bytes.Compare(signatureHash[10:], readSignatureHash) != 0 {
		return nil, fmt.Errorf("signature mismatch for '%s'", signature)
	}

	arrayData := make([]int32, readNumElements)
	arrayBytes := castInt32ArrayToByteArray(arrayData)
	if numBytesRead, err := f.Read(arrayBytes); err != nil || numBytesRead != totalSize {
		return nil, fmt.Errorf("failed to read int32 array for signature '%s': %w", signature, err)
	}

	return arrayData, nil
}

// --------------------------------------------------------------------------
func (d *trackr) save() error {
	dbFile, err := os.Create(filepath.Join(d.StoragePath, "deptrackr.point.db"))
	if err != nil {
		return err
	}
	defer dbFile.Close()

	// header := newHeader()

	numItems := len(d.ItemIdFlags)

	d.hasher.Reset()
	d.hasher.Write([]byte(d.Signature))
	signatureHash := d.hasher.Sum(nil)

	// Write the header, 32 bytes
	d.scratchBufferCursor = 0
	d.writeBytesToScratch(signatureHash[:10]) // First 10 bytes of the signature hash
	d.writeIntToScratch(numItems)             // Number of items
	d.writeInt32ToScratch(d.N)                // Number of bits for the hash
	d.writeInt32ToScratch(d.S)                // Size of a shard
	d.writeBytesToScratch(signatureHash[10:]) // Last 10 bytes of the signature hash

	if numBytesWritten, err := dbFile.Write(d.scratchBuffer[:d.scratchBufferCursor]); err != nil || numBytesWritten != d.scratchBufferCursor {
		return fmt.Errorf("failed to write DB header: %w", err)
	}

	if err := d.writeByteArray("item.id.hash.array(v1.0)", numItems, d.ItemIdHash, int(d.HashSize), dbFile); err != nil {
		return err
	}
	if err := d.writeByteArray("item.change.hash.array(v1.0)", numItems, d.ItemChangeHash, int(d.HashSize), dbFile); err != nil {
		return err
	}
	if err := d.writeByteArray("item.id.flags.array(v1.0)", numItems, d.ItemIdFlags, 1, dbFile); err != nil {
		return err
	}
	if err := d.writeByteArray("item.change.flags.array(v1.0)", numItems, d.ItemChangeFlags, 1, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.deps.start.array(v1.0)", numItems, d.ItemDepsStart, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.deps.count.array(v1.0)", numItems, d.ItemDepsCount, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.id.data.offset.array(v1.0)", numItems, d.ItemIdDataOffset, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.id.data.size.array(v1.0)", numItems, d.ItemIdDataSize, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.change.data.offset.array(v1.0)", numItems, d.ItemChangeDataOffset, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.change.data.size.array(v1.0)", numItems, d.ItemChangeDataSize, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.deps.array(v1.0)", len(d.Deps), d.Deps, dbFile); err != nil {
		return err
	}
	if err := d.writeByteArray("item.data.array(v1.0)", len(d.Data), d.Data, 1, dbFile); err != nil {
		return err
	}

	// Write shard database

	if err := d.writeInt32Array("shard.offsets.array(v1.0)", len(d.ShardOffsets), d.ShardOffsets, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("shard.sizes.array(v1.0)", len(d.ShardSizes), d.ShardSizes, dbFile); err != nil {
		return err
	}
	if err := d.writeByteArray("shard.dirty.flags.array(v1.0)", len(d.DirtyFlags), d.DirtyFlags, 1, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("shard.items.array(v1.0)", len(d.Shards), d.Shards, dbFile); err != nil {
		return err
	}

	// Flush the file to ensure all data is written

	if err := dbFile.Sync(); err != nil {
		return err
	}

	return nil
}

func loadTrackr(storagePath string, signature string) *trackr {

	// If "/deptrackr.main.db" exists and "/deptrackr.point.db" then
	// delete "/deptrackr.main.db" and rename "/deptrackr.point.db" to "/deptrackr.main.db".
	var err error
	pointDbFilepath := filepath.Join(storagePath, "deptrackr.point.db")
	mainDbFilepath := filepath.Join(storagePath, "deptrackr.main.db")
	if _, err = os.Stat(pointDbFilepath); err == nil {
		if _, err = os.Stat(mainDbFilepath); err == nil {
			if err = os.Remove(mainDbFilepath); err == nil {
				err = os.Rename(pointDbFilepath, mainDbFilepath)
			}
		} else if os.IsNotExist(err) {
			// If the main database does not exist, we just rename the point database to the main database
			err = os.Rename(pointDbFilepath, mainDbFilepath)
		}
	} else {
		_, err = os.Stat(mainDbFilepath)
	}

	// On any error, we just create a new database
	if err != nil {
		return newTrackr(storagePath, signature)
	}

	// Open the database file
	dbFile, err := os.Open(mainDbFilepath)
	if err != nil {
		// No database exists on disk, so we just create an empty one
		d := newTrackr(storagePath, signature)
		return d
	}
	defer dbFile.Close()

	d := newTrackr(storagePath, signature)

	d.hasher.Reset()
	d.hasher.Write([]byte(signature))
	signatureHash := d.hasher.Sum(nil)

	headerSize := 32

	// Read the header
	d.scratchBufferCursor = 0
	if numBytesRead, err := dbFile.Read(d.scratchBuffer[:headerSize]); err != nil || numBytesRead != headerSize {
		return newTrackr(storagePath, signature)
	}

	// The first 10 bytes is the first 10 bytes of the SHA1 of the signature
	readSignatureHash := d.readNBytesFromScratch(10)
	if bytes.Compare(signatureHash[:10], readSignatureHash) != 0 {
		return newTrackr(storagePath, signature)
	}

	numItems := d.readIntFromScratch()
	shardN := d.readInt32FromScratch()
	shardS := d.readInt32FromScratch()

	// The last 10 bytes are the last 10 bytes of the SHA1 of the signature
	readSignatureHash = d.readNBytesFromScratch(10)
	if bytes.Compare(signatureHash[10:], readSignatureHash) != 0 {
		return newTrackr(storagePath, signature)
	}

	var newItemIdHash []byte
	var newItemChangeHash []byte
	var newItemIdFlags []uint8
	var newItemChangeFlags []uint8
	var newItemDepsStart []int32
	var newItemDepsCount []int32
	var newItemIdDataOffset []int32
	var newItemIdDataSize []int32
	var newItemChangeDataOffset []int32
	var newItemChangeDataSize []int32
	var newShardOffsets []int32
	var newShardSizes []int32
	var newDirtyFlags []uint8
	var newShards []int32
	var newDeps []int32
	var newData []byte

	if newItemIdHash, err = d.readByteArray("item.id.hash.array(v1.0)", numItems, int(d.HashSize), dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newItemChangeHash, err = d.readByteArray("item.change.hash.array(v1.0)", numItems, int(d.HashSize), dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newItemIdFlags, err = d.readByteArray("item.id.flags.array(v1.0)", numItems, 1, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newItemChangeFlags, err = d.readByteArray("item.change.flags.array(v1.0)", numItems, 1, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newItemDepsStart, err = d.readInt32Array("item.deps.start.array(v1.0)", numItems, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newItemDepsCount, err = d.readInt32Array("item.deps.count.array(v1.0)", numItems, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newItemIdDataOffset, err = d.readInt32Array("item.id.data.offset.array(v1.0)", numItems, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newItemIdDataSize, err = d.readInt32Array("item.id.data.size.array(v1.0)", numItems, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newItemChangeDataOffset, err = d.readInt32Array("item.change.data.offset.array(v1.0)", numItems, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newItemChangeDataSize, err = d.readInt32Array("item.change.data.size.array(v1.0)", numItems, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newDeps, err = d.readInt32Array("item.deps.array(v1.0)", 0, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newData, err = d.readByteArray("item.data.array(v1.0)", 0, 1, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}

	if newShardOffsets, err = d.readInt32Array("shard.offsets.array(v1.0)", 1<<shardN, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newShardSizes, err = d.readInt32Array("shard.sizes.array(v1.0)", 1<<shardN, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newDirtyFlags, err = d.readByteArray("shard.dirty.flags.array(v1.0)", ((1<<shardN)+7)>>3, 1, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}
	if newShards, err = d.readInt32Array("shard.items.array(v1.0)", 0, dbFile); err != nil {
		return newTrackr(storagePath, signature)
	}

	// Initialize the trackr with the loaded data
	d.ItemState = make([]State, numItems, numItems)  // State of the item (used in Query, not loaded/saved)
	d.ItemIdHash = newItemIdHash                     // Item Id hash
	d.ItemChangeHash = newItemChangeHash             // Item Change hash
	d.ItemIdFlags = newItemIdFlags                   // Item Id flags
	d.ItemChangeFlags = newItemChangeFlags           // Item Change flags
	d.ItemDepsStart = newItemDepsStart               // Item, start of dependencies
	d.ItemDepsCount = newItemDepsCount               // Item, count of dependencies
	d.ItemIdDataOffset = newItemIdDataOffset         // Item Id data offset
	d.ItemIdDataSize = newItemIdDataSize             // Item Id data size
	d.ItemChangeDataOffset = newItemChangeDataOffset // Item Change data offset
	d.ItemChangeDataSize = newItemChangeDataSize     // Item Change data size
	d.N = shardN                                     // Number of bits for the hash
	d.S = shardS                                     // Size of a shard
	d.ShardOffsets = newShardOffsets                 // Shard offsets
	d.ShardSizes = newShardSizes                     // Shard sizes
	d.DirtyFlags = newDirtyFlags                     // Dirty flags for shards
	d.Shards = newShards                             // Shards, an array of item indices
	d.Deps = newDeps                                 // Dependencies
	d.Data = newData                                 // Data for Id and Change

	return d
}

// --------------------------------------------------------------------------
const (
	ItemFlagSourceFile = 1
	ItemFlagDependency = 2
	ItemFlagString     = 3
)

const (
	ChangeFlagModTime = 1
	ChangeFlagString  = 2
)

type ItemToAdd struct {
	IdDigest     []byte
	IdData       []byte
	ChangeDigest []byte
	ChangeData   []byte
	IdFlags      uint8
	ChangeFlags  uint8
}

type VerifyItemFunc func(itemState State, itemChangeFlags uint8, itemChangeData []byte, itemIdFlags uint8, itemIdData []byte) State

// -----------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------

const NilIndex = int32(-1)

func (d *trackr) compareDigest(a []byte, b []byte) int {
	for i := range d.HashSize {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func (d *trackr) isShardUnsorted(shardIndex int32) bool {
	return (d.DirtyFlags[shardIndex>>3] & (1 << (shardIndex & 7))) != 0 // check the dirty flag for the shard
}

func (d *trackr) insertItemIntoDb(hash []byte, item int32) {
	indexOfShard := int32(hash[0])<<8 | int32(hash[1]) // use the first N bits of the hash
	indexOfShard = indexOfShard >> (16 - d.N)          // shift to get the right index

	if d.ShardOffsets[indexOfShard] == NilIndex {
		d.ShardSizes[indexOfShard] = 0                      // initialize the shard size
		d.ShardOffsets[indexOfShard] = int32(len(d.Shards)) // initialize the shard offset
		d.Shards = append(d.Shards, d.EmptyShard...)        // initialize the shard, all of them set to -1
	}

	shardSize := d.ShardSizes[indexOfShard]
	shardOffset := d.ShardOffsets[indexOfShard]
	d.Shards[shardOffset+shardSize] = item                   // add the new item index to the shard
	d.ShardSizes[indexOfShard] = shardSize + 1               // increment the size of the shard
	d.DirtyFlags[indexOfShard>>3] |= 1 << (indexOfShard & 7) // set the dirty flag for the shard
}

func (d *trackr) DoesItemExistInDb(hash []byte) int32 {
	indexOfShard := int32(hash[0])<<8 | int32(hash[1]) // use the first N bits of the hash
	indexOfShard = indexOfShard >> (16 - d.N)          // shift to get the right index

	shardOffset := d.ShardOffsets[indexOfShard]
	if shardOffset == NilIndex {
		return NilIndex // shard doesn't exist, so the hash cannot exist
	}

	shardSize := d.ShardSizes[indexOfShard]
	if shardSize > 1 && d.isShardUnsorted(indexOfShard) {
		shardStart := shardOffset
		shardEnd := shardStart + shardSize
		slices.SortFunc(d.Shards[shardStart:shardEnd], func(i, j int32) int {
			hashI := d.ItemIdHash[i*d.HashSize : (i+1)*d.HashSize]
			hashJ := d.ItemIdHash[j*d.HashSize : (j+1)*d.HashSize]
			return d.compareDigest(hashI, hashJ) // sort in ascending order
		})
		// Mark the shard as sorted
		d.DirtyFlags[indexOfShard>>3] = d.DirtyFlags[indexOfShard>>3] &^ (1 << (indexOfShard & 7))
	}

	// Binary search for the hash in the sorted array
	indexOfHashInShard := NilIndex
	low, high := int32(0), shardSize-1
	for low <= high {
		mid := (low + high) / 2
		midItemIndex := d.Shards[shardOffset+mid]
		midItemHash := d.ItemIdHash[midItemIndex*d.HashSize : (midItemIndex+1)*d.HashSize]
		c := d.compareDigest(midItemHash, hash)
		if c == 0 {
			indexOfHashInShard = mid // found
			break
		} else if c < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	if indexOfHashInShard == NilIndex {
		return NilIndex
	}
	return d.Shards[shardOffset+indexOfHashInShard] // return the index of the item in the shard
}

func (d *trackr) AddItem(item ItemToAdd, deps []ItemToAdd) bool {
	existingItemIndex := d.DoesItemExistInDb(item.IdDigest)
	if existingItemIndex != NilIndex {
		// This should not happen, as we are inserting a new item
		return false
	}

	itemIndex := int32(len(d.ItemIdFlags))

	// Insert the item into the main arrays
	d.ItemState = append(d.ItemState, StateNone)          // default state is none
	d.ItemIdHash = append(d.ItemIdHash, item.IdDigest...) // add the item Id hash
	if len(item.ChangeDigest) == 20 {
		d.ItemChangeHash = append(d.ItemChangeHash, item.ChangeDigest...)
	} else {
		d.ItemChangeHash = append(d.ItemChangeHash, item.IdDigest...)
	}
	d.ItemIdFlags = append(d.ItemIdFlags, item.IdFlags)                              // add the item Id flags
	d.ItemChangeFlags = append(d.ItemChangeFlags, item.ChangeFlags)                  // add the item Change flags
	d.ItemDepsStart = append(d.ItemDepsStart, int32(len(d.Deps)))                    // start of dependencies is 0
	d.ItemDepsCount = append(d.ItemDepsCount, int32(len(deps)))                      // count of dependencies
	d.ItemIdDataOffset = append(d.ItemIdDataOffset, int32(len(d.Data)))              // item Id data
	d.ItemIdDataSize = append(d.ItemIdDataSize, int32(len(item.IdData)))             // item Id data
	d.Data = append(d.Data, item.IdData...)                                          // add the Item Id data to the Data array
	d.ItemChangeDataOffset = append(d.ItemChangeDataOffset, int32(len(d.Data)))      // item Id data
	d.ItemChangeDataSize = append(d.ItemChangeDataSize, int32(len(item.ChangeData))) // item Id data
	d.Data = append(d.Data, item.ChangeData...)                                      // add the Item Change data to the Data array

	// Add the item to the shard database, so that we can search it
	d.insertItemIntoDb(item.IdDigest, itemIndex)

	// Insert dependencies
	// Note: dependencies as an Item are shared
	for _, dep := range deps {
		if existingItemIndex := d.DoesItemExistInDb(dep.IdDigest); existingItemIndex != NilIndex {
			d.Deps = append(d.Deps, existingItemIndex) // add the dependency item index
		} else {
			// Insert the dependency item into the main arrays
			depItemIndex := int32(len(d.ItemIdFlags))            //
			d.ItemState = append(d.ItemState, StateNone)         // default state is none
			d.ItemIdHash = append(d.ItemIdHash, dep.IdDigest...) // add the dependency Id hash
			if len(item.ChangeDigest) == 20 {                    //
				d.ItemChangeHash = append(d.ItemChangeHash, dep.ChangeDigest...)
			} else {
				d.ItemChangeHash = append(d.ItemChangeHash, dep.IdDigest...)
			}
			d.ItemIdFlags = append(d.ItemIdFlags, dep.IdFlags)                              // add the dependency Id flags
			d.ItemChangeFlags = append(d.ItemChangeFlags, dep.ChangeFlags)                  // add the dependency Change flags
			d.ItemDepsStart = append(d.ItemDepsStart, int32(0))                             // start of dependencies is 0
			d.ItemDepsCount = append(d.ItemDepsCount, int32(0))                             // count of dependencies is 0 for now
			d.ItemIdDataOffset = append(d.ItemIdDataOffset, int32(len(d.Data)))             // item Id data
			d.ItemIdDataSize = append(d.ItemIdDataSize, int32(len(dep.IdData)))             // item Id data
			d.Data = append(d.Data, dep.IdData...)                                          // add the Item Id data to the Data array
			d.ItemChangeDataOffset = append(d.ItemChangeDataOffset, int32(len(d.Data)))     // item Id data
			d.ItemChangeDataSize = append(d.ItemChangeDataSize, int32(len(dep.ChangeData))) // item Id data
			d.Data = append(d.Data, dep.ChangeData...)                                      // add the Item Change data to the Data array

			// Add the dependency item to the shard database, so that we can search it
			d.insertItemIntoDb(dep.IdDigest, depItemIndex)

			d.Deps = append(d.Deps, depItemIndex) // add the dependency item index
		}
	}

	return true
}

func (d *trackr) QueryItem(itemHash []byte, verifyAll bool, verifyCb VerifyItemFunc) (State, error) {

	itemIndex := d.DoesItemExistInDb(itemHash)
	if itemIndex == NilIndex {
		return StateOutOfDate, nil // item does not exist, so it is out of date
	}

	// If the item is already marked as up to date, we can return immediately
	itemState := d.ItemState[itemIndex]
	if itemState == StateUpToDate || itemState == StateOutOfDate {
		return itemState, nil
	}

	itemChangeFlags := d.ItemChangeFlags[itemIndex]
	itemChangeData := d.Data[d.ItemChangeDataOffset[itemIndex] : d.ItemChangeDataOffset[itemIndex]+d.ItemChangeDataSize[itemIndex]]
	itemIdFlags := d.ItemIdFlags[itemIndex]
	itemIdData := d.Data[d.ItemIdDataOffset[itemIndex] : d.ItemIdDataOffset[itemIndex]+d.ItemIdDataSize[itemIndex]]
	itemState = verifyCb(itemState, itemChangeFlags, itemChangeData, itemIdFlags, itemIdData)

	d.ItemState[itemIndex] = itemState

	outOfDateCount := 0
	if itemState == StateOutOfDate {
		outOfDateCount++
	}

	// Check the dependencies
	depStart := d.ItemDepsStart[itemIndex]
	depEnd := depStart + d.ItemDepsCount[itemIndex]
	for depStart < depEnd {
		depItemIndex := d.Deps[depStart]
		depState := d.ItemState[depItemIndex]
		depChangeFlags := d.ItemChangeFlags[depItemIndex]
		depChangeData := d.Data[d.ItemChangeDataOffset[depItemIndex] : d.ItemChangeDataOffset[depItemIndex]+d.ItemChangeDataSize[depItemIndex]]
		depIdFlags := d.ItemIdFlags[depItemIndex]
		depIdData := d.Data[d.ItemIdDataOffset[depItemIndex] : d.ItemIdDataOffset[depItemIndex]+d.ItemIdDataSize[depItemIndex]]
		depState = verifyCb(depState, depChangeFlags, depChangeData, depIdFlags, depIdData)

		d.ItemState[depItemIndex] = depState

		if depState == StateOutOfDate {
			outOfDateCount++
			if !verifyAll {
				break
			}
		}
		depStart++
	}

	if outOfDateCount == 0 {
		d.ItemState[itemIndex] = StateUpToDate // mark the item as up to date
		return StateUpToDate, nil
	}

	// Mark the main item as out-of-date when any of the dependencies where out of date
	d.ItemState[itemIndex] = StateOutOfDate
	return StateOutOfDate, nil
}

// CopyItem copies an item from one trackr to another.
func (src *trackr) CopyItem(dst *trackr, itemHash []byte) error {
	itemIndex := src.DoesItemExistInDb(itemHash)
	if itemIndex == NilIndex {
		return nil // item does not exist, nothing to copy
	}

	// Copy all the item properties
	itemIdHash := src.ItemIdHash[itemIndex*src.HashSize : (itemIndex+1)*src.HashSize]
	itemChangeHash := src.ItemChangeHash[itemIndex*src.HashSize : (itemIndex+1)*src.HashSize]
	itemIdFlags := src.ItemIdFlags[itemIndex]
	itemChangeFlags := src.ItemChangeFlags[itemIndex]
	itemDepsCount := src.ItemDepsCount[itemIndex]
	itemIdDataOffset := src.ItemIdDataOffset[itemIndex]
	itemIdDataSize := src.ItemIdDataSize[itemIndex]
	itemIdData := src.Data[itemIdDataOffset : itemIdDataOffset+itemIdDataSize]
	itemChangeDataOffset := src.ItemChangeDataOffset[itemIndex]
	itemChangeDataSize := src.ItemChangeDataSize[itemIndex]
	itemChangeData := src.Data[itemChangeDataOffset : itemChangeDataOffset+itemChangeDataSize]

	// Add a new item in dst
	dstItemIndex := int32(len(dst.ItemIdFlags))                                       // index of the new item in the destination trackr
	dst.ItemIdHash = append(dst.ItemIdHash, itemIdHash...)                            // add the item Id hash
	dst.ItemChangeHash = append(dst.ItemChangeHash, itemChangeHash...)                // add the item Change hash
	dst.ItemIdFlags = append(dst.ItemIdFlags, itemIdFlags)                            // add the item Id flags
	dst.ItemChangeFlags = append(dst.ItemChangeFlags, itemChangeFlags)                // add the item Change flags
	dst.ItemDepsStart = append(dst.ItemDepsStart, int32(len(dst.Deps)))               // start of dependencies is 0
	dst.ItemDepsCount = append(dst.ItemDepsCount, itemDepsCount)                      // count of dependencies
	dst.ItemIdDataOffset = append(dst.ItemIdDataOffset, int32(len(dst.Data)))         // item Id data
	dst.ItemIdDataSize = append(dst.ItemIdDataSize, itemIdDataSize)                   // item Id data
	dst.Data = append(dst.Data, itemIdData...)                                        // add the Item Id data to the Data array
	dst.ItemChangeDataOffset = append(dst.ItemChangeDataOffset, int32(len(dst.Data))) // item Id data
	dst.ItemChangeDataSize = append(dst.ItemChangeDataSize, itemChangeDataSize)       // item Id data
	dst.Data = append(dst.Data, itemChangeData...)                                    // add the Item Change data to the Data array

	// Insert the item into the shard database
	dst.insertItemIntoDb(itemIdHash, dstItemIndex)

	srcDepItemCursor := src.ItemDepsStart[itemIndex]
	srcDepItemEnd := srcDepItemCursor + itemDepsCount
	for srcDepItemCursor < srcDepItemEnd {

		srcDepItemIndex := src.Deps[srcDepItemCursor] // index of the dependency item in the source trackr
		depItemIdHash := src.ItemIdHash[srcDepItemIndex*src.HashSize : (srcDepItemIndex+1)*src.HashSize]

        dstDepItemIndex := dst.DoesItemExistInDb(depItemIdHash)
        if dstDepItemIndex == NilIndex {
            depItemChangeHash := src.ItemChangeHash[srcDepItemIndex*src.HashSize : (srcDepItemIndex+1)*src.HashSize]
            depItemIdFlags := src.ItemIdFlags[srcDepItemIndex]
            depItemChangeFlags := src.ItemChangeFlags[srcDepItemIndex]
            depItemIdDataOffset := src.ItemIdDataOffset[srcDepItemIndex]
            depItemIdDataSize := src.ItemIdDataSize[srcDepItemIndex]
            depItemIdData := src.Data[depItemIdDataOffset : depItemIdDataOffset+depItemIdDataSize]
            depItemChangeDataOffset := src.ItemChangeDataOffset[srcDepItemIndex]
            depItemChangeDataSize := src.ItemChangeDataSize[srcDepItemIndex]
            depItemChangeData := src.Data[depItemChangeDataOffset : depItemChangeDataOffset+depItemChangeDataSize]

            dstDepItemIndex = int32(len(dst.ItemIdFlags))                                    // index of the new dependency item in the destination trackr
            dst.ItemIdHash = append(dst.ItemIdHash, depItemIdHash...)                         // add the dependency Id hash
            dst.ItemChangeHash = append(dst.ItemChangeHash, depItemChangeHash...)             // add the dependency Change hash
            dst.ItemIdFlags = append(dst.ItemIdFlags, depItemIdFlags)                         // add the dependency Id flags
            dst.ItemChangeFlags = append(dst.ItemChangeFlags, depItemChangeFlags)             // add the dependency Change flags
            dst.ItemDepsStart = append(dst.ItemDepsStart, 0)                                  // start of dependencies is 0
            dst.ItemDepsCount = append(dst.ItemDepsCount, 0)                                  // count of dependencies is 0 for now
            dst.ItemIdDataOffset = append(dst.ItemIdDataOffset, int32(len(dst.Data)))         // item Id data
            dst.ItemIdDataSize = append(dst.ItemIdDataSize, depItemIdDataSize)                // item Id data
            dst.Data = append(dst.Data, depItemIdData...)                                     // add the Item Id data to the Data array
            dst.ItemChangeDataOffset = append(dst.ItemChangeDataOffset, int32(len(dst.Data))) // item Id data
            dst.ItemChangeDataSize = append(dst.ItemChangeDataSize, depItemChangeDataSize)    // item Id data
            dst.Data = append(dst.Data, depItemChangeData...)                                 // add the Item Change data to the Data array

            // Insert the dependency item into the shard database
            dst.insertItemIntoDb(depItemIdHash, dstDepItemIndex)
        }

        // add the dependency item index to the destination trackr
		dst.Deps = append(dst.Deps, dstDepItemIndex)

        srcDepItemCursor++
	}

	return nil
}
