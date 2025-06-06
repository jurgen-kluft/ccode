package dpenc

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"os"
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

type State int8

const (
	StateNone      State = 0
	StateUpToDate  State = 1
	StateOutOfDate State = 2
	StateError     State = 3
)

type trackr struct {
	hasher               hash.Hash
	readonly             bool        // If true, the database is read-only, we cannot add items
	scratchBuffer        *BinaryData // A temporary byte buffer for hashing and other operations, not saved to disk
	storageFilepath      string      // Filepath where we store the database file
	signature            string      // max 32 characters signature, e.g. ".d deptracker v1.0.0"
	hashSize             int32       // Size of the hash, this is 20 bytes for SHA1
	itemState            []State     // Hash, this is the state of the item that is modified during a query
	ItemIdHash           []byte      // Hash, this is the ID of the item (filepath, label (e.g. 'MSVC C++ compiler cmd-line arguments))
	ItemChangeHash       []byte      // Hash, this identifies the 'change' (modification-time, file-size, file-content, command-line arguments, string, etc..)
	ItemIdFlags          []uint8     //
	ItemChangeFlags      []uint8     //
	ItemDepsStart        []int32     // Item, start of dependencies
	ItemDepsCount        []int32     // Item, count of dependencies
	ItemIdDataOffset     []int32     // data for Id (-1 means no data)
	ItemIdDataSize       []int32     // (size=0 means no data)
	ItemExtraDataOffset  []int32     // extra data for item
	ItemExtraDataSize    []uint8     // (size=0 means no data)
	ItemChangeDataOffset []int32     // data for Change (-1 means no data)
	ItemChangeDataSize   []uint8     // (size=0 means no data)
	Deps                 []int32     // Array for each item to list their dependencies, this is a flat array of item indices
	Data                 []byte      // Here any data (id, change, extra) is stored, this is a flat array of bytes
	N                    int32       // how many bits we take from the hash to index into the shards (0-15)
	S                    int32       // size of a shard, this is the number of items per shard, default is 512
	ShardOffsets         []int32     // This is an array of offsets into the Shards array, each offset corresponds to a shard, 0 means the shard doesn't exist yet
	ShardSizes           []int16     // This is an array of sizes of each shard, 0 means the shard is empty
	DirtyFlags           []uint8     // A bit per shard, indicates if the shard is dirty (unsorted) and needs to be sorted (excluded from load/save)
	Shards               []int32     // An array of shards, a shard is a region of item-indices
	EmptyShard           []int32     // A shard initialized to a size of S and full of NillIndex
}

func constructTrackr(storageFilepath string, signature string, numItems int, dataSize int) *trackr {
	hs := int32(20) // SHA1 hash size is 20 bytes
	n := int32(10)  // how many bits we take from the hash to index into the buckets (0-15)
	s := int32(512) // size of a shard, this is the number of items per shard, default is 512

	d := &trackr{
		hasher:               sha1.New(),                        // Create a new SHA1 hasher
		readonly:             false,                             //
		scratchBuffer:        NewBinaryData(256),                // A temporary byte buffer for hashing and other operations, not saved to disk
		storageFilepath:      storageFilepath,                   //
		signature:            signature,                         // copy the signature from the source
		hashSize:             hs,                                //
		itemState:            make([]State, 0, numItems),        // State of the item, this is used to track if the item is up to date or not
		ItemIdHash:           make([]byte, 0, numItems*int(hs)), //
		ItemChangeHash:       make([]byte, 0, numItems*int(hs)), //
		ItemIdFlags:          make([]uint8, 0, numItems),        //
		ItemChangeFlags:      make([]uint8, 0, numItems),        //
		ItemDepsStart:        make([]int32, 0, numItems),        //
		ItemDepsCount:        make([]int32, 0, numItems),        //
		ItemIdDataOffset:     make([]int32, 0, numItems),        //
		ItemIdDataSize:       make([]int32, 0, numItems),        //
		ItemExtraDataOffset:  make([]int32, 0, numItems),        //
		ItemExtraDataSize:    make([]uint8, 0, numItems),        //
		ItemChangeDataOffset: make([]int32, 0, numItems),        //
		ItemChangeDataSize:   make([]uint8, 0, numItems),        //
		N:                    n,                                 //
		S:                    s,                                 //
		ShardOffsets:         make([]int32, 1<<n),               // initial capacity of 2^N shards
		ShardSizes:           make([]int16, 1<<n),               // initial capacity of 2^N shards
		DirtyFlags:           nil,                               // initial capacity of N bits for dirty flags (rounded up to the nearest byte)
		Shards:               make([]int32, 0, (1<<n)*s),        // initial capacity of 2^N shards where each shard has s elements
		EmptyShard:           make([]int32, s),                  // an empty shard, used to copy when making a new shard in Shards
		Deps:                 make([]int32, 0, numItems),        // Array for each item to list their dependencies, this is a flat array of item indices
		Data:                 make([]byte, 0, dataSize),         // Data for Id and Change, this is a flat array of bytes
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
	numItems := max(1024, len(src.ItemIdFlags)+len(src.ItemIdFlags)/8)
	dataSize := len(src.Data) + (len(src.Data) / 2)
	d := constructTrackr(src.storageFilepath, src.signature, numItems, dataSize)
	d.readonly = false
	d.DirtyFlags = make([]uint8, (((1 << d.N) + 7) >> 3))
	return d
}

// --------------------------------------------------------------------------
// Helper functions to cast to byte array (used for load/save)
func castInt16ArrayToByteArray(i []int16) []byte {
	if len(i) == 0 {
		return []byte{}
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&i[0])), len(i)<<1)
}

func castInt32ArrayToByteArray(i []int32) []byte {
	if len(i) == 0 {
		return []byte{}
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&i[0])), len(i)<<2)
}

func castByteArrayToInt16Array(b []byte) (array []int16, err error) {
	if len(b)&1 != 0 {
		err = fmt.Errorf("byte array length is not a multiple of 2, cannot cast to int16 array")
	} else {
		array = unsafe.Slice((*int16)(unsafe.Pointer(&b[0])), len(b)>>1)
	}
	return
}

func castByteArrayToInt32Array(b []byte) (array []int32, err error) {
	if len(b)&3 != 0 {
		err = fmt.Errorf("byte array length is not a multiple of 4, cannot cast to int32 array")
	} else {
		array = unsafe.Slice((*int32)(unsafe.Pointer(&b[0])), len(b)>>2)
	}
	return
}

func (d *trackr) getHashOfSignature(signature string) []byte {
	d.hasher.Reset()
	d.hasher.Write([]byte(d.signature)) // trackr signature
	d.hasher.Write([]byte(signature))   // array signature
	return d.hasher.Sum(nil)
}

func (d *trackr) writeArray(signature string, byteArray []byte, compress bool, f *os.File) error {
	arrayCompressedSizeInBytes := 0
	arrayOriginalSizeInBytes := len(byteArray)
	byteArrayToWrite := byteArray

	if compress {
		compressed := bytes.NewBuffer(make([]byte, 0, len(byteArray)))
		compressor := zlib.NewWriter(compressed)
		n, err1 := compressor.Write(byteArray)
		err2 := compressor.Close()
		if err1 == nil && n == len(byteArray) && err2 == nil {
			byteArrayToWrite = compressed.Bytes()
			arrayCompressedSizeInBytes = len(byteArrayToWrite)
		}
	}

	signatureHash := d.getHashOfSignature(signature)

	d.scratchBuffer.reset()
	d.scratchBuffer.writeBytes(signatureHash[:8])        // First 8 bytes of the signature hash
	d.scratchBuffer.writeInt(arrayOriginalSizeInBytes)   // Original size of the array
	d.scratchBuffer.writeInt(arrayCompressedSizeInBytes) // Compressed size of the array
	d.scratchBuffer.writeInt(0)                          // Reserved
	d.scratchBuffer.writeInt(0)                          // Reserved
	d.scratchBuffer.writeBytes(signatureHash[(20 - 8):]) // Last 8 bytes of the signature hash
	if err := d.scratchBuffer.writeToFile(f); err != nil {
		return err
	}

	if numBytesWritten, err := f.Write(byteArrayToWrite); err != nil || numBytesWritten != len(byteArrayToWrite) {
		return fmt.Errorf("failed to write compressed int32 array for signature '%s': %w", signature, err)
	}
	return nil
}

func (d *trackr) readArray(signature string, f *os.File) (byteArray []byte, err error) {
	headerSize := 32
	if err := d.scratchBuffer.readFromFile(headerSize, f); err != nil {
		return nil, err
	}

	signatureHash := d.getHashOfSignature(signature)
	firstSignatureHash := d.scratchBuffer.readNBytes(8)
	arrayOriginalSizeInBytes := d.scratchBuffer.readInt()
	arrayCompressedSizeInBytes := d.scratchBuffer.readInt()
	_ = d.scratchBuffer.readInt() // Reserved, not used
	_ = d.scratchBuffer.readInt() // Reserved, not used
	lastSignatureHash := d.scratchBuffer.readNBytes(8)
	if bytes.Compare(signatureHash[:8], firstSignatureHash) != 0 || bytes.Compare(signatureHash[(20-8):], lastSignatureHash) != 0 {
		return nil, fmt.Errorf("signature mismatch for '%s'", signature)
	}

	numBytesToRead := arrayOriginalSizeInBytes
	if arrayCompressedSizeInBytes > 0 {
		numBytesToRead = arrayCompressedSizeInBytes
	}

	byteArray = make([]byte, numBytesToRead)
	if numBytesRead, err := f.Read(byteArray); err != nil || numBytesRead != numBytesToRead {
		return nil, fmt.Errorf("failed to read byte array for signature '%s': %w", signature, err)
	}

	if arrayCompressedSizeInBytes > 0 { // Decompress the data
		decompressor, err := zlib.NewReader(bytes.NewReader(byteArray))
		if err != nil {
			return nil, fmt.Errorf("failed to create decompressor for signature '%s': %w", signature, err)
		}

		decompressedByteArray := make([]byte, arrayOriginalSizeInBytes)
		totalDecompressedNumBytes := 0

		var err1 error
		for totalDecompressedNumBytes < arrayOriginalSizeInBytes {
			var decompressedNumBytes int
			decompressedNumBytes, err1 = decompressor.Read(decompressedByteArray[totalDecompressedNumBytes:])
			totalDecompressedNumBytes += decompressedNumBytes
			if err1 == io.EOF {
				err1 = nil
				break
			} else if err1 != nil {
				break
			}
		}
		err2 := decompressor.Close()

		if err1 != nil || err2 != nil || totalDecompressedNumBytes != arrayOriginalSizeInBytes {
			return nil, fmt.Errorf("failed to decompress byte array for signature '%s': %w", signature, err1)
		}
		byteArray = decompressedByteArray
	}

	return byteArray, nil
}

func (d *trackr) writeByteArray(signature string, array []byte, f *os.File) error {
	compressed := false
	return d.writeArray(signature, array, compressed, f)
}

func (d *trackr) writeByteArrayCompressed(signature string, array []byte, f *os.File) error {
	compressed := true
	return d.writeArray(signature, array, compressed, f)
}

func (d *trackr) readByteArray(signature string, f *os.File) (byteArray []byte, err error) {
	byteArray, err = d.readArray(signature, f)
	return byteArray, err
}

func (d *trackr) writeInt32Array(signature string, array []int32, f *os.File) error {
	compressed := false
	return d.writeArray(signature, castInt32ArrayToByteArray(array), compressed, f)
}

func (d *trackr) writeInt32ArrayCompressed(signature string, array []int32, f *os.File) error {
	compressed := true
	return d.writeArray(signature, castInt32ArrayToByteArray(array), compressed, f)
}

func (d *trackr) readInt32Array(signature string, f *os.File) (arrayData []int32, err error) {
	var byteArray []byte
	if byteArray, err = d.readArray(signature, f); err == nil {
		arrayData, err = castByteArrayToInt32Array(byteArray)
	}
	return arrayData, err
}

func (d *trackr) writeInt16Array(signature string, array []int16, f *os.File) error {
	compressed := false
	return d.writeArray(signature, castInt16ArrayToByteArray(array), compressed, f)
}

// func (d *trackr) writeInt16ArrayCompressed(signature string, array []int16, f *os.File) error {
// 	compressed := true
// 	return d.writeArray(signature, castInt16ArrayToByteArray(array), compressed, f)
// }

func (d *trackr) readInt16Array(signature string, f *os.File) (arrayData []int16, err error) {
	var byteArray []byte
	if byteArray, err = d.readArray(signature, f); err == nil {
		arrayData, err = castByteArrayToInt16Array(byteArray)
	}
	return arrayData, err
}

// --------------------------------------------------------------------------
func (d *trackr) validate() error {

	// most of the array's have a deterministic size based on the number of items.
	// we can thus verify the size of those arrays against the number of items.
	hashSize := int(d.hashSize)
	numItems := len(d.ItemIdFlags)

	if len(d.ItemIdHash) != numItems*hashSize {
		return fmt.Errorf("ItemIdHash size mismatch: expected %d, got %d", numItems*hashSize, len(d.ItemIdHash))
	}
	if len(d.ItemChangeHash) != numItems*hashSize {
		return fmt.Errorf("ItemChangeHash size mismatch: expected %d, got %d", numItems*hashSize, len(d.ItemChangeHash))
	}
	if len(d.ItemChangeFlags) != numItems {
		return fmt.Errorf("ItemChangeFlags size mismatch: expected %d, got %d", numItems, len(d.ItemChangeFlags))
	}
	if len(d.ItemDepsStart) != numItems {
		return fmt.Errorf("ItemDepsStart size mismatch: expected %d, got %d", numItems, len(d.ItemDepsStart))
	}
	if len(d.ItemDepsCount) != numItems {
		return fmt.Errorf("ItemDepsCount size mismatch: expected %d, got %d", numItems, len(d.ItemDepsCount))
	}
	if len(d.ItemIdDataOffset) != numItems {
		return fmt.Errorf("ItemIdDataOffset size mismatch: expected %d, got %d", numItems, len(d.ItemIdDataOffset))
	}
	if len(d.ItemIdDataSize) != numItems {
		return fmt.Errorf("ItemIdDataSize size mismatch: expected %d, got %d", numItems, len(d.ItemIdDataSize))
	}
	if len(d.ItemExtraDataOffset) != numItems {
		return fmt.Errorf("ItemExtraDataOffset size mismatch: expected %d, got %d", numItems, len(d.ItemExtraDataOffset))
	}
	if len(d.ItemExtraDataSize) != numItems {
		return fmt.Errorf("ItemExtraDataSize size mismatch: expected %d, got %d", numItems, len(d.ItemExtraDataSize))
	}
	if len(d.ItemChangeDataOffset) != numItems {
		return fmt.Errorf("ItemChangeDataOffset size mismatch: expected %d, got %d", numItems, len(d.ItemChangeDataOffset))
	}
	if len(d.ItemChangeDataSize) != numItems {
		return fmt.Errorf("ItemChangeDataSize size mismatch: expected %d, got %d", numItems, len(d.ItemChangeDataSize))
	}

	shardCount := 1 << d.N // Number of shards is 2^N
	if len(d.ShardOffsets) != shardCount {
		return fmt.Errorf("ShardOffsets size mismatch: expected %d, got %d", shardCount, len(d.ShardOffsets))
	}
	if len(d.ShardSizes) != shardCount {
		return fmt.Errorf("ShardSizes size mismatch: expected %d, got %d", shardCount, len(d.ShardSizes))
	}
	// if len(d.DirtyFlags) != ((shardCount + 7) >> 3) {
	// 	return fmt.Errorf("DirtyFlags size mismatch: expected %d, got %d", (shardCount+7)>>3, len(d.DirtyFlags))
	// }

	return nil
}

// --------------------------------------------------------------------------
func (d *trackr) save() error {
	if d.readonly {
		return fmt.Errorf("a read-only tracker cannot be saved")
	}

	dbFile, err := os.Create(d.storageFilepath + ".point.db")
	if err != nil {
		return err
	}
	defer dbFile.Close()

	numItems := len(d.ItemIdFlags)

	// Write the header, 32 bytes
	d.hasher.Reset()
	d.hasher.Write([]byte(d.signature))
	signatureHash := d.hasher.Sum(nil)
	d.scratchBuffer.reset()
	d.scratchBuffer.writeBytes(signatureHash[:10]) // First 10 bytes of the signature hash
	d.scratchBuffer.writeInt(numItems)             // Number of items
	d.scratchBuffer.writeInt32(d.N)                // Number of bits for the hash
	d.scratchBuffer.writeInt32(d.S)                // Size of a shard
	d.scratchBuffer.writeBytes(signatureHash[10:]) // Last 10 bytes of the signature hash

	if err := d.scratchBuffer.writeToFile(dbFile); err != nil {
		return err
	}

	if err := d.validate(); err != nil {
		return err
	}

	if err := d.writeByteArray("item.id.hash.array", d.ItemIdHash, dbFile); err != nil {
		return err
	}
	if err := d.writeByteArray("item.change.hash.array", d.ItemChangeHash, dbFile); err != nil {
		return err
	}
	if err := d.writeByteArray("item.id.flags.array", d.ItemIdFlags, dbFile); err != nil {
		return err
	}
	if err := d.writeByteArray("item.change.flags.array", d.ItemChangeFlags, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.deps.start.array", d.ItemDepsStart, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.deps.count.array", d.ItemDepsCount, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.id.data.offset.array", d.ItemIdDataOffset, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.id.data.size.array", d.ItemIdDataSize, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.extra.data.offset.array", d.ItemExtraDataOffset, dbFile); err != nil {
		return err
	}
	if err := d.writeByteArray("item.extra.data.size.array", d.ItemExtraDataSize, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.change.data.offset.array", d.ItemChangeDataOffset, dbFile); err != nil {
		return err
	}
	if err := d.writeByteArray("item.change.data.size.array", d.ItemChangeDataSize, dbFile); err != nil {
		return err
	}
	if err := d.writeInt32Array("item.deps.array", d.Deps, dbFile); err != nil {
		return err
	}
	if err := d.writeByteArrayCompressed("item.data.array", d.Data, dbFile); err != nil {
		return err
	}

	// Here we iterate over the bytes of dirty flags and only sort a shard
	// if a dirty flag is set for that shard.
	for i := int32(0); i < int32(len(d.DirtyFlags)); i++ {
		if d.DirtyFlags[i] == 0 {
			continue
		}
		indexOfShard := (i << 3)
		b := d.DirtyFlags[i]
		for b != 0 {
			if (b & 1) != 0 {
				d.sortShard(indexOfShard)
			}
			indexOfShard++
			b = b >> 1
		}
		d.DirtyFlags[i] = 0
	}

	if err := d.writeInt32Array("shard.offsets.array", d.ShardOffsets, dbFile); err != nil {
		return err
	}
	if err := d.writeInt16Array("shard.sizes.array", d.ShardSizes, dbFile); err != nil {
		return err
	}
	// if err := d.writeByteArray("shard.dirty.flags.array", d.DirtyFlags, dbFile); err != nil {
	// 	return err
	// }
	if err := d.writeInt32ArrayCompressed("shard.items.array", d.Shards, dbFile); err != nil {
		return err
	}

	// Flush the file to ensure all data is written

	if err := dbFile.Sync(); err != nil {
		return err
	}

	return nil
}

// newDefaultTracker creates a new (nearly empty) tracker when loading fails
func newDefaultTracker(storageFilepath string, signature string) *trackr {
	d := constructTrackr(storageFilepath, signature, 8, 8)
	d.readonly = true // Set the trackr to read-only mode
	return d
}

func loadTrackr(storageFilepath string, signature string) *trackr {
	// If "/name.main.db" exists and "/name.point.db" then
	// delete "/name.main.db" and rename "/name.point.db" to "/name.main.db".
	var err error
	pointDbFilepath := storageFilepath + ".point.db"
	mainDbFilepath := storageFilepath + ".main.db"
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
		return newDefaultTracker(storageFilepath, signature)
	}

	// Open the database file
	dbFile, err := os.Open(mainDbFilepath)
	if err != nil {
		// No database exists on disk, so we just create an empty one
		d := newDefaultTracker(storageFilepath, signature)
		return d
	}
	defer dbFile.Close()

	d := newDefaultTracker(storageFilepath, signature)

	d.hasher.Reset()
	d.hasher.Write([]byte(signature))
	signatureHash := d.hasher.Sum(nil)

	headerSize := 32

	// Read the header
	if err := d.scratchBuffer.readFromFile(headerSize, dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}

	// The first 10 bytes is the first 10 bytes of the SHA1 of the signature
	readSignatureHash := d.scratchBuffer.readNBytes(10)
	if bytes.Compare(signatureHash[:10], readSignatureHash) != 0 {
		return newDefaultTracker(storageFilepath, signature)
	}

	numItems := d.scratchBuffer.readInt()
	shardN := d.scratchBuffer.readInt32()
	shardS := d.scratchBuffer.readInt32()

	// The last 10 bytes are the last 10 bytes of the SHA1 of the signature
	readSignatureHash = d.scratchBuffer.readNBytes(10)
	if bytes.Compare(signatureHash[10:], readSignatureHash) != 0 {
		return newDefaultTracker(storageFilepath, signature)
	}

	var newItemIdHash []byte
	var newItemChangeHash []byte
	var newItemIdFlags []uint8
	var newItemChangeFlags []uint8
	var newItemDepsStart []int32
	var newItemDepsCount []int32
	var newItemIdDataOffset []int32
	var newItemIdDataSize []int32
	var newItemExtraDataOffset []int32
	var newItemExtraDataSize []uint8
	var newItemChangeDataOffset []int32
	var newItemChangeDataSize []uint8
	var newShardOffsets []int32
	var newShardSizes []int16
	var newShards []int32
	var newDeps []int32
	var newData []byte

	if newItemIdHash, err = d.readByteArray("item.id.hash.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newItemChangeHash, err = d.readByteArray("item.change.hash.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newItemIdFlags, err = d.readByteArray("item.id.flags.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newItemChangeFlags, err = d.readByteArray("item.change.flags.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newItemDepsStart, err = d.readInt32Array("item.deps.start.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newItemDepsCount, err = d.readInt32Array("item.deps.count.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newItemIdDataOffset, err = d.readInt32Array("item.id.data.offset.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newItemIdDataSize, err = d.readInt32Array("item.id.data.size.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newItemExtraDataOffset, err = d.readInt32Array("item.extra.data.offset.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newItemExtraDataSize, err = d.readByteArray("item.extra.data.size.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newItemChangeDataOffset, err = d.readInt32Array("item.change.data.offset.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newItemChangeDataSize, err = d.readByteArray("item.change.data.size.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newDeps, err = d.readInt32Array("item.deps.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newData, err = d.readByteArray("item.data.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}

	if newShardOffsets, err = d.readInt32Array("shard.offsets.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	if newShardSizes, err = d.readInt16Array("shard.sizes.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}
	// if newDirtyFlags, err = d.readByteArray("shard.dirty.flags.array", dbFile); err != nil {
	// 	return newDefaultTracker(storageFilepath, signature)
	// }
	if newShards, err = d.readInt32Array("shard.items.array", dbFile); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}

	// Initialize the trackr with the loaded data
	d.readonly = true                                // Set the trackr to read-only mode after loading
	d.itemState = make([]State, numItems, numItems)  // State of the item (used in Query, not loaded/saved)
	d.ItemIdHash = newItemIdHash                     // Item Id hash
	d.ItemChangeHash = newItemChangeHash             // Item Change hash
	d.ItemIdFlags = newItemIdFlags                   // Item Id flags
	d.ItemChangeFlags = newItemChangeFlags           // Item Change flags
	d.ItemDepsStart = newItemDepsStart               // Item, start of dependencies
	d.ItemDepsCount = newItemDepsCount               // Item, count of dependencies
	d.ItemIdDataOffset = newItemIdDataOffset         // Item Id data offset
	d.ItemIdDataSize = newItemIdDataSize             // Item Id data size
	d.ItemExtraDataOffset = newItemExtraDataOffset   // Item extra data offset
	d.ItemExtraDataSize = newItemExtraDataSize       // Item extra data size
	d.ItemChangeDataOffset = newItemChangeDataOffset // Item Change data offset
	d.ItemChangeDataSize = newItemChangeDataSize     // Item Change data size
	d.N = shardN                                     // Number of bits for the hash
	d.S = shardS                                     // Size of a shard
	d.ShardOffsets = newShardOffsets                 // Shard offsets
	d.ShardSizes = newShardSizes                     // Shard sizes
	d.DirtyFlags = nil                               // Dirty flags, a bit per shard
	d.Shards = newShards                             // Shards, an array of item indices
	d.Deps = newDeps                                 // Dependencies
	d.Data = newData                                 // Data for Id and Change

	if err := d.validate(); err != nil {
		return newDefaultTracker(storageFilepath, signature)
	}

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

type VerifyItemFunc func(itemState State, itemIdFlags uint8, itemIdData []byte, itemChangeFlags uint8, itemChangeData []byte) State
type VerifyItemExtraFunc func(itemState State, itemIdFlags uint8, itemIdData []byte, itemExtraData []byte, itemChangeFlags uint8, itemChangeData []byte) State

type verifyItemIndexFunc func(itemIndex int32) State

// -----------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------

const NilIndex = int32(-1)

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

	shardSize := int32(d.ShardSizes[indexOfShard])
	shardOffset := d.ShardOffsets[indexOfShard]
	d.Shards[shardOffset+shardSize] = item                   // add the new item index to the shard
	d.ShardSizes[indexOfShard] += 1                          // increment the size of the shard
	d.DirtyFlags[indexOfShard>>3] |= 1 << (indexOfShard & 7) // set the dirty flag for the shard
}

func (d *trackr) sortShard(indexOfShard int32) {
	shardStart := d.ShardOffsets[indexOfShard]
	if shardStart != NilIndex {
		shardSize := int32(d.ShardSizes[indexOfShard])
		if shardSize > 1 {
			slices.SortFunc(d.Shards[shardStart:shardStart+shardSize], func(i, j int32) int {
				io := i * d.hashSize
				jo := j * d.hashSize
				ioe := io + d.hashSize
				for io < ioe {
					ib := d.ItemIdHash[io]
					jb := d.ItemIdHash[jo]
					if ib < jb {
						return -1
					} else if ib > jb {
						return 1
					}
					io++
					jo++
				}
				return 0
			})
		}
	}
}

func (d *trackr) DoesItemExistInDb(hash []byte) int32 {
	indexOfShard := int32(hash[0])<<8 | int32(hash[1]) // use the first N bits of the hash
	indexOfShard = indexOfShard >> (16 - d.N)          // shift to get the right index

	shardOffset := d.ShardOffsets[indexOfShard]
	if shardOffset == NilIndex {
		return NilIndex // shard doesn't exist, so the hash cannot exist
	}

	shardSize := d.ShardSizes[indexOfShard]
	if !d.readonly && shardSize > 1 && d.isShardUnsorted(indexOfShard) {
		d.sortShard(indexOfShard)
		// Mark the shard as sorted
		d.DirtyFlags[indexOfShard>>3] = d.DirtyFlags[indexOfShard>>3] &^ (1 << (indexOfShard & 7))
	}

	// Binary search for the hash in the sorted array
	indexOfHashInShard := NilIndex
	low, high := int32(0), int32(shardSize)-1
	for low <= high {
		mid := (low + high) / 2
		midItemIndex := d.Shards[shardOffset+mid]
		midItemHash := d.ItemIdHash[midItemIndex*d.hashSize : (midItemIndex+1)*d.hashSize]
		c := bytes.Compare(midItemHash, hash)
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

func (d *trackr) addItem(item ItemToAdd, itemExtraData []byte, deps []ItemToAdd) bool {
	existingItemIndex := d.DoesItemExistInDb(item.IdDigest)
	if existingItemIndex != NilIndex {
		// This should not happen, as we are inserting a new item
		return false
	}

	itemIndex := int32(len(d.ItemIdFlags))

	// Insert the item into the main arrays
	d.itemState = append(d.itemState, StateNone)          // default state is none
	d.ItemIdHash = append(d.ItemIdHash, item.IdDigest...) // add the item Id hash
	if len(item.ChangeDigest) == 20 {
		d.ItemChangeHash = append(d.ItemChangeHash, item.ChangeDigest...)
	} else {
		d.ItemChangeHash = append(d.ItemChangeHash, item.IdDigest...)
	}
	d.ItemIdFlags = append(d.ItemIdFlags, item.IdFlags)                  // add the item Id flags
	d.ItemChangeFlags = append(d.ItemChangeFlags, item.ChangeFlags)      // add the item Change flags
	d.ItemDepsStart = append(d.ItemDepsStart, int32(len(d.Deps)))        // start of dependencies is 0
	d.ItemDepsCount = append(d.ItemDepsCount, int32(len(deps)))          // count of dependencies
	d.ItemIdDataOffset = append(d.ItemIdDataOffset, int32(len(d.Data)))  // item Id data
	d.ItemIdDataSize = append(d.ItemIdDataSize, int32(len(item.IdData))) // item Id data
	d.Data = append(d.Data, item.IdData...)                              // add the Item Id data to the Data array
	if len(itemExtraData) > 0 {
		d.ItemExtraDataOffset = append(d.ItemExtraDataOffset, int32(len(d.Data)))    // item extra data
		d.ItemExtraDataSize = append(d.ItemExtraDataSize, uint8(len(itemExtraData))) // item extra data
		d.Data = append(d.Data, itemExtraData...)                                    // add the Item extra data to the Data array
	} else {
		d.ItemExtraDataOffset = append(d.ItemExtraDataOffset, 0) // item extra data
		d.ItemExtraDataSize = append(d.ItemExtraDataSize, 0)     // item extra data
	}
	d.ItemChangeDataOffset = append(d.ItemChangeDataOffset, int32(len(d.Data)))      // item Id data
	d.ItemChangeDataSize = append(d.ItemChangeDataSize, uint8(len(item.ChangeData))) // item Id data
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
			d.itemState = append(d.itemState, StateNone)         // default state is none
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
			d.ItemExtraDataOffset = append(d.ItemExtraDataOffset, 0)                        // item extra data
			d.ItemExtraDataSize = append(d.ItemExtraDataSize, 0)                            // item extra data
			d.ItemChangeDataOffset = append(d.ItemChangeDataOffset, int32(len(d.Data)))     // item Id data
			d.ItemChangeDataSize = append(d.ItemChangeDataSize, uint8(len(dep.ChangeData))) // item Id data
			d.Data = append(d.Data, dep.ChangeData...)                                      // add the Item Change data to the Data array

			// Add the dependency item to the shard database, so that we can search it
			d.insertItemIntoDb(dep.IdDigest, depItemIndex)

			d.Deps = append(d.Deps, depItemIndex) // add the dependency item index
		}
	}

	return true
}

func (d *trackr) AddItem(item ItemToAdd, deps []ItemToAdd) bool {
	return d.addItem(item, nil, deps) // no extra data
}
func (d *trackr) AddItemWithExtraData(item ItemToAdd, itemExtraData []byte, deps []ItemToAdd) bool {
	return d.addItem(item, itemExtraData, deps) // no extra data
}

// ---------------------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------
// QueryItem queries an item in the database and verifies its state using the provided callback function.

func (d *trackr) queryItem(itemHash []byte, verifyAll bool, verifyCb verifyItemIndexFunc) (State, error) {
	itemIndex := d.DoesItemExistInDb(itemHash)
	if itemIndex == NilIndex {
		return StateOutOfDate, nil // item does not exist, so it is out of date
	}

	// If the item is already marked as up to date, we can return immediately
	itemState := d.itemState[itemIndex]
	if itemState == StateUpToDate || itemState == StateOutOfDate {
		return itemState, nil
	}

	itemState = verifyCb(itemIndex)
	d.itemState[itemIndex] = itemState

	outOfDateCount := 0
	if itemState == StateOutOfDate {
		outOfDateCount++
	}

	// Check the dependencies
	depStart := d.ItemDepsStart[itemIndex]
	depEnd := depStart + d.ItemDepsCount[itemIndex]
	for depStart < depEnd {
		depItemIndex := d.Deps[depStart]
		depState := verifyCb(depItemIndex)
		d.itemState[depItemIndex] = depState
		if depState == StateOutOfDate {
			outOfDateCount++
			if !verifyAll {
				break
			}
		}
		depStart++
	}

	if outOfDateCount == 0 {
		d.itemState[itemIndex] = StateUpToDate // mark the item as up to date
		return StateUpToDate, nil
	}

	// Mark the main item as out-of-date when any of the dependencies where out of date
	d.itemState[itemIndex] = StateOutOfDate
	return StateOutOfDate, nil
}

func (d *trackr) QueryItem(itemHash []byte, verifyAll bool, verifyCb VerifyItemFunc) (State, error) {
	return d.queryItem(itemHash, verifyAll, func(itemIndex int32) State {
		itemIdFlags := d.ItemIdFlags[itemIndex]
		itemIdDataOffset := d.ItemIdDataOffset[itemIndex]
		itemIdDataSize := d.ItemIdDataSize[itemIndex]
		itemIdData := d.Data[itemIdDataOffset : itemIdDataOffset+itemIdDataSize]
		itemChangeFlags := d.ItemChangeFlags[itemIndex]
		itemChangeDataOffset := d.ItemChangeDataOffset[itemIndex]
		itemChangeDataSize := d.ItemChangeDataSize[itemIndex]
		itemChangeData := d.Data[itemChangeDataOffset : itemChangeDataOffset+int32(itemChangeDataSize)]
		return verifyCb(d.itemState[itemIndex], itemIdFlags, itemIdData, itemChangeFlags, itemChangeData)
	})
}

func (d *trackr) QueryItemExtra(itemHash []byte, verifyAll bool, verifyCb VerifyItemExtraFunc) (State, error) {
	return d.queryItem(itemHash, verifyAll, func(itemIndex int32) State {
		itemIdFlags := d.ItemIdFlags[itemIndex]
		itemIdDataOffset := d.ItemIdDataOffset[itemIndex]
		itemIdDataSize := d.ItemIdDataSize[itemIndex]
		itemIdData := d.Data[itemIdDataOffset : itemIdDataOffset+itemIdDataSize]
		itemExtraDataOffset := d.ItemExtraDataOffset[itemIndex]
		itemExtraDataSize := d.ItemExtraDataSize[itemIndex]
		itemExtraData := d.Data[itemExtraDataOffset : itemExtraDataOffset+int32(itemExtraDataSize)]
		itemChangeFlags := d.ItemChangeFlags[itemIndex]
		itemChangeDataOffset := d.ItemChangeDataOffset[itemIndex]
		itemChangeDataSize := d.ItemChangeDataSize[itemIndex]
		itemChangeData := d.Data[itemChangeDataOffset : itemChangeDataOffset+int32(itemChangeDataSize)]
		return verifyCb(d.itemState[itemIndex], itemIdFlags, itemIdData, itemExtraData, itemChangeFlags, itemChangeData)
	})
}

// ---------------------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------
// CopyItem copies an item from one trackr to another, this is used when an item is up-to-data in the
// source trackr.

func (src *trackr) CopyItem(dst *trackr, itemHash []byte) error {
	itemIndex := src.DoesItemExistInDb(itemHash)
	if itemIndex == NilIndex {
		return nil // item does not exist, nothing to copy
	}

	// Copy all the item properties
	itemIdHash := src.ItemIdHash[itemIndex*src.hashSize : (itemIndex+1)*src.hashSize]
	itemChangeHash := src.ItemChangeHash[itemIndex*src.hashSize : (itemIndex+1)*src.hashSize]
	itemIdFlags := src.ItemIdFlags[itemIndex]
	itemChangeFlags := src.ItemChangeFlags[itemIndex]
	itemDepsCount := src.ItemDepsCount[itemIndex]
	itemIdDataOffset := src.ItemIdDataOffset[itemIndex]
	itemIdDataSize := src.ItemIdDataSize[itemIndex]
	itemIdData := src.Data[itemIdDataOffset : itemIdDataOffset+itemIdDataSize]
	itemExtraDataOffset := src.ItemExtraDataOffset[itemIndex]
	itemExtraDataSize := src.ItemExtraDataSize[itemIndex]
	itemExtraData := src.Data[itemExtraDataOffset : itemExtraDataOffset+int32(itemExtraDataSize)]
	itemChangeDataOffset := src.ItemChangeDataOffset[itemIndex]
	itemChangeDataSize := src.ItemChangeDataSize[itemIndex]
	itemChangeData := src.Data[itemChangeDataOffset : itemChangeDataOffset+int32(itemChangeDataSize)]

	// Add a new item in dst
	dstItemIndex := int32(len(dst.ItemIdFlags))                               // index of the new item in the destination trackr
	dst.ItemIdHash = append(dst.ItemIdHash, itemIdHash...)                    // add the item Id hash
	dst.ItemChangeHash = append(dst.ItemChangeHash, itemChangeHash...)        // add the item Change hash
	dst.ItemIdFlags = append(dst.ItemIdFlags, itemIdFlags)                    // add the item Id flags
	dst.ItemChangeFlags = append(dst.ItemChangeFlags, itemChangeFlags)        // add the item Change flags
	dst.ItemDepsStart = append(dst.ItemDepsStart, int32(len(dst.Deps)))       // start of dependencies is 0
	dst.ItemDepsCount = append(dst.ItemDepsCount, itemDepsCount)              // count of dependencies
	dst.ItemIdDataOffset = append(dst.ItemIdDataOffset, int32(len(dst.Data))) // item Id data
	dst.ItemIdDataSize = append(dst.ItemIdDataSize, itemIdDataSize)           // item Id data
	dst.Data = append(dst.Data, itemIdData...)                                // add the Item Id data to the Data array
	if len(itemExtraData) > 0 {
		dst.ItemExtraDataOffset = append(dst.ItemExtraDataOffset, int32(len(dst.Data)))  // item extra data
		dst.ItemExtraDataSize = append(dst.ItemExtraDataSize, uint8(len(itemExtraData))) // item extra data
		dst.Data = append(dst.Data, itemExtraData...)                                    // add the Item extra data to the Data array
	} else {
		dst.ItemExtraDataOffset = append(dst.ItemExtraDataOffset, 0) // item extra data
		dst.ItemExtraDataSize = append(dst.ItemExtraDataSize, 0)     // item extra data
	}
	dst.ItemChangeDataOffset = append(dst.ItemChangeDataOffset, int32(len(dst.Data))) // item Id data
	dst.ItemChangeDataSize = append(dst.ItemChangeDataSize, itemChangeDataSize)       // item Id data
	dst.Data = append(dst.Data, itemChangeData...)                                    // add the Item Change data to the Data array

	// Insert the item into the shard database
	dst.insertItemIntoDb(itemIdHash, dstItemIndex)

	srcDepItemCursor := src.ItemDepsStart[itemIndex]
	srcDepItemEnd := srcDepItemCursor + itemDepsCount
	for srcDepItemCursor < srcDepItemEnd {

		srcDepItemIndex := src.Deps[srcDepItemCursor] // index of the dependency item in the source trackr
		depItemIdHash := src.ItemIdHash[srcDepItemIndex*src.hashSize : (srcDepItemIndex+1)*src.hashSize]

		dstDepItemIndex := dst.DoesItemExistInDb(depItemIdHash)
		if dstDepItemIndex == NilIndex {
			depItemChangeHash := src.ItemChangeHash[srcDepItemIndex*src.hashSize : (srcDepItemIndex+1)*src.hashSize]
			depItemIdFlags := src.ItemIdFlags[srcDepItemIndex]
			depItemChangeFlags := src.ItemChangeFlags[srcDepItemIndex]
			depItemIdDataOffset := src.ItemIdDataOffset[srcDepItemIndex]
			depItemIdDataSize := src.ItemIdDataSize[srcDepItemIndex]
			depItemIdData := src.Data[depItemIdDataOffset : depItemIdDataOffset+depItemIdDataSize]
			// Dependency items do not have extra data, so we can skip it
			depItemChangeDataOffset := src.ItemChangeDataOffset[srcDepItemIndex]
			depItemChangeDataSize := src.ItemChangeDataSize[srcDepItemIndex]
			depItemChangeData := src.Data[depItemChangeDataOffset : depItemChangeDataOffset+int32(depItemChangeDataSize)]

			dstDepItemIndex = int32(len(dst.ItemIdFlags))                                     // index of the new dependency item in the destination trackr
			dst.ItemIdHash = append(dst.ItemIdHash, depItemIdHash...)                         // add the dependency Id hash
			dst.ItemChangeHash = append(dst.ItemChangeHash, depItemChangeHash...)             // add the dependency Change hash
			dst.ItemIdFlags = append(dst.ItemIdFlags, depItemIdFlags)                         // add the dependency Id flags
			dst.ItemChangeFlags = append(dst.ItemChangeFlags, depItemChangeFlags)             // add the dependency Change flags
			dst.ItemDepsStart = append(dst.ItemDepsStart, 0)                                  // start of dependencies is 0
			dst.ItemDepsCount = append(dst.ItemDepsCount, 0)                                  // count of dependencies is 0 for now
			dst.ItemIdDataOffset = append(dst.ItemIdDataOffset, int32(len(dst.Data)))         // item Id data
			dst.ItemIdDataSize = append(dst.ItemIdDataSize, depItemIdDataSize)                // item Id data
			dst.Data = append(dst.Data, depItemIdData...)                                     // add the Item Id data to the Data array
			dst.ItemExtraDataOffset = append(dst.ItemExtraDataOffset, 0)                      // item extra data
			dst.ItemExtraDataSize = append(dst.ItemExtraDataSize, 0)                          // item extra data
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
