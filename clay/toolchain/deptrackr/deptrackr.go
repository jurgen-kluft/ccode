package deptrackr

import (
	"encoding/binary"
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

type depTrackr struct {
	StoragePath          string   // Path to the directory where we store the database file
	Signature            string   // max 32 characters signature, e.g. ".d deptracker v1.0.0"
	HashSize             int32    // Size of the hash, this is 20 bytes for SHA1
	ItemIdHash           []byte   // Hash, this is the ID of the item (filepath, label (e.g. 'MSVC C++ compiler cmd-line arguments))
	ItemChangeHash       []byte   // Hash, this identifies the 'change' (modification-time, file-size, file-content, command-line arguments, string, etc..)
	ItemIdFlags          []uint16 //
	ItemChangeFlags      []uint16 //
	ItemDepsStart        []int32  // Item, start of dependencies
	ItemDepsCount        []int32  // Item, count of dependencies
	ItemIdDataOffset     []int32  // data for Id, 4 bytes (0 means no data)
	ItemIdDataSize       []int32  //
	ItemChangeDataOffset []int32  // data for Change, 4 bytes (0 means no data)
	ItemChangeDataSize   []int32  //
	Deps                 []int32  // Array for each item to list their dependencies, this is a flat array of item indices
	Data                 []byte   // Here the data for Id and Change is stored, this is a flat array of bytes
	N                    int32    // how many bits we take from the hash to index into the shards (0-15)
	S                    int32    // size of a shard, this is the number of items per shard, default is 512
	ShardOffsets         []int32  // This is an array of offsets into the Shards array, each offset corresponds to a shard, 0 means the shard doesn't exist yet
	ShardSizes           []int32  // This is an array of sizes of each shard, 0 means the shard is empty
	DirtyFlags           []uint8  // A bit per shard, indicates if the shard is dirty (unsorted) and needs to be sorted
	Shards               []int32  // An array of shards, a shard is a region of item-indices
	EmptyShard           []int32  // A shard initialized to a size of S and full of NillIndex
}

func newDepTrackr(storageDir string, signature string) *depTrackr {
	reserve := 1024 // A new database is set to reserve 1024 items
	hs := int32(20) // SHA1 hash size is 20 bytes
	n := int32(10)  // how many bits we take from the hash to index into the buckets (0-15)
	s := int32(512) // size of a shard, this is the number of items per shard, default is 512

	d := &depTrackr{
		StoragePath:          storageDir,                       //
		Signature:            signature,                        //
		HashSize:             hs,                               //
		ItemIdHash:           make([]byte, 0, reserve*int(hs)), //
		ItemChangeHash:       make([]byte, 0, reserve*int(hs)), //
		ItemIdFlags:          make([]uint16, 0, reserve),       //
		ItemChangeFlags:      make([]uint16, 0, reserve),       //
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
		EmptyShard:           make([]int32, s, s),              // an empty shard, used to copy when making a new shard in Shards
		Deps:                 make([]int32, 0, reserve*64),     //
		Data:                 make([]byte, 0, reserve*1024),    //
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

func (src *depTrackr) newDepTrackr() *depTrackr {
	reserve := len(src.ItemIdHash)
	reserve = reserve + (reserve / 10) // reserve 10% more space for the new DB

	hs := src.HashSize
	n := src.N
	s := src.S

	newTrackr := &depTrackr{
		StoragePath:          src.StoragePath,                                    //
		Signature:            src.Signature,                                      // copy the signature from the source
		HashSize:             hs,                                                 //
		ItemIdHash:           make([]byte, 0, reserve*int(hs)),                   //
		ItemChangeHash:       make([]byte, 0, reserve*int(hs)),                   //
		ItemIdFlags:          make([]uint16, 0, reserve),                         //
		ItemChangeFlags:      make([]uint16, 0, reserve),                         //
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

func castUint16ArrayToByteArray(i []uint16) []byte {
	if len(i) == 0 {
		return []byte{}
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&i[0])), len(i)<<2)
}

// --------------------------------------------------------------------------
func (d *depTrackr) save() error {
	dbFile, err := os.Create(d.StoragePath + "/deptrackr.point.db")
	if err != nil {
		return err
	}
	defer dbFile.Close()

	header := make([]byte, 0, 64)

	// Write the signature to the header, max 32 bytes
	signatureBytes := []byte(d.Signature)
	header = append(header, signatureBytes[:32]...)

	// Number of items, also indicates the size of any Item... array
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.ItemIdHash)))
	// The length of the deps and data arrays
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.Deps)))
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.Data)))

	// The shard database information
	header = binary.LittleEndian.AppendUint32(header, uint32(d.N)) // Number of bits for the hash
	header = binary.LittleEndian.AppendUint32(header, uint32(d.S)) // Size of a shard
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.Shards)))

	// The rest of the header fill it with zeros
	for len(header) < 64 {
		header = append(header, 0)
	}

	// Write the header to the file
	if _, err := dbFile.Write(header); err != nil {
		return err
	}

	// Save Items
	if _, err := dbFile.Write(d.ItemIdHash); err != nil {
		return err
	}
	if _, err := dbFile.Write(d.ItemChangeHash); err != nil {
		return err
	}
	if _, err := dbFile.Write(castUint16ArrayToByteArray(d.ItemIdFlags)); err != nil {
		return err
	}
	if _, err := dbFile.Write(castUint16ArrayToByteArray(d.ItemChangeFlags)); err != nil {
		return err
	}
	if _, err := dbFile.Write(castInt32ArrayToByteArray(d.ItemDepsStart)); err != nil {
		return err
	}
	if _, err := dbFile.Write(castInt32ArrayToByteArray(d.ItemDepsCount)); err != nil {
		return err
	}
	if _, err := dbFile.Write(castInt32ArrayToByteArray(d.ItemIdDataOffset)); err != nil {
		return err
	}
	if _, err := dbFile.Write(castInt32ArrayToByteArray(d.ItemIdDataSize)); err != nil {
		return err
	}
	if _, err := dbFile.Write(castInt32ArrayToByteArray(d.ItemChangeDataOffset)); err != nil {
		return err
	}
	if _, err := dbFile.Write(castInt32ArrayToByteArray(d.ItemChangeDataSize)); err != nil {
		return err
	}
	if _, err := dbFile.Write(castInt32ArrayToByteArray(d.Deps)); err != nil {
		return err
	}
	if _, err := dbFile.Write(d.Data); err != nil {
		return err
	}

	// Write shard database, ShardOffsets
	shardOffsets := castInt32ArrayToByteArray(d.ShardOffsets)
	if _, err := dbFile.Write(shardOffsets); err != nil {
		return err
	}

	// Write shard database, ShardSizes
	shardSizes := castInt32ArrayToByteArray(d.ShardSizes)
	if _, err := dbFile.Write(shardSizes); err != nil {
		return err
	}

	// Write shard database, DirtyFlags
	if _, err := dbFile.Write(d.DirtyFlags); err != nil {
		return err
	}

	// Write shard database, Shards
	shardsData := castInt32ArrayToByteArray(d.Shards)
	if _, err := dbFile.Write(shardsData); err != nil {
		return err
	}

	// Flush the file to ensure all data is written
	if err := dbFile.Sync(); err != nil {
		return err
	}

	return nil
}

func loadDepTrackr(storagePath string, signature string) *depTrackr {

	// If "/deptrackr.main.db" exists and "/deptrackr.point.db" then
	// delete "/deptrackr.main.db" and rename "/deptrackr.point.db" to "/deptrackr.main.db".
	var err error
	if _, err = os.Stat(storagePath + "/deptrackr.point.db"); err == nil {
		if _, err = os.Stat(storagePath + "/deptrackr.main.db"); err == nil {
			if err = os.Remove(storagePath + "/deptrackr.main.db"); err == nil {
				err = os.Rename(storagePath+"/deptrackr.point.db", storagePath+"/deptrackr.main.db")
			}
		} else if os.IsNotExist(err) {
			// If the main database does not exist, we just rename the point database to the main database
			err = os.Rename(storagePath+"/deptrackr.point.db", storagePath+"/deptrackr.main.db")
		}
	} else {
		_, err = os.Stat(storagePath + "/deptrackr.main.db")
	}

	// On any error, we just create a new database
	if err != nil {
		return newDepTrackr(storagePath, signature)
	}

	// Open the database file
	dbFile, err := os.Open(storagePath + "/deptrackr.main.db")
	if err != nil {
		// No database exists on disk, so we just create an empty one
		d := newDepTrackr(storagePath, signature)
		return d
	}
	defer dbFile.Close()

	// Read the header
	header := make([]byte, 64)
	if _, err := dbFile.Read(header); err != nil {
		return newDepTrackr(storagePath, signature)
	}

	loadedSignature := string(header[0:32])
	if loadedSignature != signature {
		// The signature does not match, so we create a new database
		return newDepTrackr(storagePath, signature)
	}

	numItems := int(binary.LittleEndian.Uint32(header[0:4]))
	depsArrayLen := int(binary.LittleEndian.Uint32(header[0:4]))
	dataArrayLen := int(binary.LittleEndian.Uint32(header[0:4]))

	newHashSize := int32(20)

	newItemIdHash := make([]byte, 0, numItems*int(newHashSize))
	if bytesRead, err := dbFile.Read(newItemIdHash); err != nil || bytesRead != numItems*int(newHashSize) {
		return newDepTrackr(storagePath, signature)
	}
	newItemChangeHash := make([]byte, numItems*int(newHashSize))
	if bytesRead, err := dbFile.Read(newItemChangeHash); err != nil || bytesRead != numItems*int(newHashSize) {
		return newDepTrackr(storagePath, signature)
	}

	newItemIdFlags := make([]uint16, numItems)
	itemIdFlagsBytes := castUint16ArrayToByteArray(newItemIdFlags)
	itemIdFlagsNumBytes := numItems * 2
	if bytesRead, err := dbFile.Read(itemIdFlagsBytes); err != nil || bytesRead != itemIdFlagsNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	newItemChangeFlags := make([]uint16, numItems)
	itemChangeFlagsBytes := castUint16ArrayToByteArray(newItemChangeFlags)
	itemChangeFlagsNumBytes := numItems * 2
	if bytesRead, err := dbFile.Read(itemChangeFlagsBytes); err != nil || bytesRead != itemChangeFlagsNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	newItemDepsStart := make([]int32, numItems)
	itemDepsStartBytes := castInt32ArrayToByteArray(newItemDepsStart)
	itemDepsStartNumBytes := numItems * 4
	if bytesRead, err := dbFile.Read(itemDepsStartBytes); err != nil || bytesRead != itemDepsStartNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	newItemDepsCount := make([]int32, numItems)
	itemDepsCountBytes := castInt32ArrayToByteArray(newItemDepsCount)
	itemDepsCountNumBytes := numItems * 4
	if bytesRead, err := dbFile.Read(itemDepsCountBytes); err != nil || bytesRead != itemDepsCountNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	newItemIdDataOffset := make([]int32, numItems)
	itemIdDataOffsetBytes := castInt32ArrayToByteArray(newItemIdDataOffset)
	itemIdDataOffsetNumBytes := numItems * 4
	if bytesRead, err := dbFile.Read(itemIdDataOffsetBytes); err != nil || bytesRead != itemIdDataOffsetNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	newItemIdDataSize := make([]int32, numItems)
	itemIdDataSizeBytes := castInt32ArrayToByteArray(newItemIdDataSize)
	itemIdDataSizeNumBytes := numItems * 4
	if bytesRead, err := dbFile.Read(itemIdDataSizeBytes); err != nil || bytesRead != itemIdDataSizeNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	newItemChangeDataOffset := make([]int32, numItems)
	itemChangeDataOffsetBytes := castInt32ArrayToByteArray(newItemChangeDataOffset)
	itemChangeDataOffsetNumBytes := numItems * 4
	if bytesRead, err := dbFile.Read(itemChangeDataOffsetBytes); err != nil || bytesRead != itemChangeDataOffsetNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	newItemChangeDataSize := make([]int32, numItems)
	itemChangeDataSizeBytes := castInt32ArrayToByteArray(newItemChangeDataSize)
	itemChangeDataSizeNumBytes := numItems * 4
	if bytesRead, err := dbFile.Read(itemChangeDataSizeBytes); err != nil || bytesRead != itemChangeDataSizeNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	newDeps := make([]int32, depsArrayLen)
	depsBytes := castInt32ArrayToByteArray(newDeps)
	depsNumBytes := depsArrayLen * 4
	if bytesRead, err := dbFile.Read(depsBytes); err != nil || bytesRead != depsNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	newData := make([]byte, dataArrayLen)
	if bytesRead, err := dbFile.Read(newData); err != nil || bytesRead != dataArrayLen {
		return newDepTrackr(storagePath, signature)
	}

	shardN := int32(binary.LittleEndian.Uint32(header[0:4]))
	shardS := int32(binary.LittleEndian.Uint32(header[0:4]))
	shardsArrayLen := int32(binary.LittleEndian.Uint32(header[0:4]))

	newShardOffsets := make([]int32, (1 << shardN))
	shardOffsetsBytes := castInt32ArrayToByteArray(newShardOffsets)
	shardOffsetsNumBytes := (1 << shardN) * 4 // 4 bytes per int32
	if bytesRead, err := dbFile.Read(shardOffsetsBytes); err != nil || bytesRead != shardOffsetsNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	newShardSizes := make([]int32, (1 << shardN))
	shardSizesBytes := castInt32ArrayToByteArray(newShardSizes)
	shardSizesNumBytes := (1 << shardN) * 4 // 4 bytes per int32
	if bytesRead, err := dbFile.Read(shardSizesBytes); err != nil || bytesRead != shardSizesNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	newDirtyFlags := make([]byte, (shardN+7)>>3) // N bits, rounded up to the nearest byte
	dirtyFlagsBytes := newDirtyFlags
	dirtyFlagsNumBytes := int((shardN + 7) >> 3) // N bits, rounded up to the nearest byte
	if bytesRead, err := dbFile.Read(dirtyFlagsBytes); err != nil || bytesRead != dirtyFlagsNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	newShards := make([]int32, (shardsArrayLen * 4))
	shardsDataBytes := castInt32ArrayToByteArray(newShards)
	shardsDataNumBytes := int(shardsArrayLen * 4) // 4 bytes per int32
	if bytesRead, err := dbFile.Read(shardsDataBytes); err != nil || bytesRead != shardsDataNumBytes {
		return newDepTrackr(storagePath, signature)
	}

	d := &depTrackr{
		StoragePath:          storagePath,                   // Path to the directory where we store the database file
		Signature:            signature,                     // Signature of the database
		HashSize:             newHashSize,                   // SHA1 hash size is 20 bytes
		ItemIdHash:           newItemIdHash,                 // Item Id hash
		ItemChangeHash:       newItemChangeHash,             // Item Change hash
		ItemIdFlags:          newItemIdFlags,                // Item Id flags
		ItemChangeFlags:      newItemChangeFlags,            // Item Change flags
		ItemDepsStart:        newItemDepsStart,              // Item, start of dependencies
		ItemDepsCount:        newItemDepsCount,              // Item, count of dependencies
		ItemIdDataOffset:     newItemIdDataOffset,           // Item Id data offset
		ItemIdDataSize:       newItemIdDataSize,             // Item Id data size
		ItemChangeDataOffset: newItemChangeDataOffset,       // Item Change data offset
		ItemChangeDataSize:   newItemChangeDataSize,         // Item Change data size
		Deps:                 newDeps,                       // Dependencies
		Data:                 newData,                       // Data for Id and Change
		N:                    shardN,                        // Number of bits for the hash
		S:                    shardS,                        // Size of a shard
		ShardOffsets:         newShardOffsets,               // Shard offsets
		ShardSizes:           newShardSizes,                 // Shard sizes
		DirtyFlags:           newDirtyFlags,                 // Dirty flags for shards
		Shards:               newShards,                     // Shards, an array of item indices
		EmptyShard:           make([]int32, shardS, shardS), // An empty shard, used to copy when making a new shard in Shards
	}

	// Set the first 8 elements of EmptyShard to -1
	pattern := []int32{NilIndex, NilIndex, NilIndex, NilIndex, NilIndex, NilIndex, NilIndex, NilIndex}
	copy(d.EmptyShard, pattern)
	// Fill the rest of the shard with the pattern, increasing the size of the pattern
	for j := len(pattern); j < len(d.EmptyShard); j *= 2 {
		copy(d.EmptyShard[j:], d.EmptyShard[:j])
	}

	return d
}

// --------------------------------------------------------------------------
const (
	ItemFlagSourceFile = 1
	ItemFlagDependency = 2
	ItemFlagString     = 8
	ItemFlagUpToDate   = 64  // This is set when the item is up to date
	ItemFlagOutOfDate  = 128 // This is set when the item is out of dat
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
	IdFlags      uint16
	ChangeFlags  uint16
}

type State int

const (
	StateUpToDate  State = 0
	StateOutOfDate State = 2
	StateError     State = 4
	StateIgnore    State = 8
)

type VerifyItemFunc func(itemChangeFlags uint16, itemChangeData []byte, itemIdFlags uint16, itemIdData []byte) State

// -----------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------

const NilIndex = int32(-1)

func (d *depTrackr) compareDigest(a []byte, b []byte) int {
	for i := range d.HashSize {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func (d *depTrackr) isShardUnsorted(shardIndex int32) bool {
	return (d.DirtyFlags[shardIndex>>3] & (1 << (shardIndex & 7))) != 0 // check the dirty flag for the shard
}

func (d *depTrackr) insertItemIntoDb(hash []byte, item int32) {
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

func (d *depTrackr) DoesItemExistInDb(hash []byte) int32 {
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
			itemIndexI := d.Shards[shardStart+i]
			itemIndexJ := d.Shards[shardStart+j]
			hashI := d.ItemIdHash[itemIndexI*d.HashSize : (itemIndexI+1)*d.HashSize]
			hashJ := d.ItemIdHash[itemIndexJ*d.HashSize : (itemIndexJ+1)*d.HashSize]
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

func (d *depTrackr) AddItem(item ItemToAdd, deps []ItemToAdd) bool {
	existingItemIndex := d.DoesItemExistInDb(item.IdDigest)
	if existingItemIndex != NilIndex {
		// This should not happen, as we are inserting a new item
		return false
	}

	itemIndex := int32(len(d.ItemIdFlags))

	// Insert the item into the main arrays
	d.ItemIdHash = append(d.ItemIdHash, item.IdDigest...)                            // add the item Id hash
	d.ItemChangeHash = append(d.ItemChangeHash, item.ChangeDigest...)                // add the item Change hash
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
			depItemIndex := int32(len(d.ItemIdFlags))
			d.ItemIdHash = append(d.ItemIdHash, dep.IdDigest...)                            // add the dependency Id hash
			d.ItemChangeHash = append(d.ItemChangeHash, dep.ChangeDigest...)                // add the dependency Change hash
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

func (d *depTrackr) QueryItem(itemHash []byte, verifyAll bool, verifyCb VerifyItemFunc) (State, error) {

	itemIndex := d.DoesItemExistInDb(itemHash)
	if itemIndex == NilIndex {
		return StateOutOfDate, nil // item does not exist, so it is out of date
	}

	// type VerifyItemFunc func(itemChangeFlags uint32, itemChangeData []byte, itemIdFlags uint32, itemIdData []byte) State
	itemChangeFlags := d.ItemChangeFlags[itemIndex]
	itemChangeData := d.Data[d.ItemChangeDataOffset[itemIndex] : d.ItemChangeDataOffset[itemIndex]+d.ItemChangeDataSize[itemIndex]]
	itemIdFlags := d.ItemIdFlags[itemIndex]
	itemIdData := d.Data[d.ItemIdDataOffset[itemIndex] : d.ItemIdDataOffset[itemIndex]+d.ItemIdDataSize[itemIndex]]

	// Check if the item is up to date
	itemState := verifyCb(itemChangeFlags, itemChangeData, itemIdFlags, itemIdData)

	outOfDateCount := 0
	if itemState == StateOutOfDate {
		d.ItemIdFlags[itemIndex] |= ItemFlagUpToDate // mark the item as out of date
		if !verifyAll {
			return StateOutOfDate, nil
		}
		outOfDateCount++
	} else {
		d.ItemIdFlags[itemIndex] |= ItemFlagOutOfDate // mark the item as out of date
	}

	// Check the dependencies
	depStart := d.ItemDepsStart[itemIndex]
	depEnd := depStart + d.ItemDepsCount[itemIndex]
	for depStart < depEnd {

		depItemIndex := d.Deps[depStart]

		depChangeFlags := d.ItemChangeFlags[depItemIndex]
		depChangeData := d.Data[d.ItemChangeDataOffset[depItemIndex] : d.ItemChangeDataOffset[depItemIndex]+d.ItemChangeDataSize[depItemIndex]]
		depIdFlags := d.ItemIdFlags[depItemIndex]
		depIdData := d.Data[d.ItemIdDataOffset[depItemIndex] : d.ItemIdDataOffset[depItemIndex]+d.ItemIdDataSize[depItemIndex]]

		depState := verifyCb(depChangeFlags, depChangeData, depIdFlags, depIdData)
		if depState == StateOutOfDate {
			d.ItemIdFlags[depItemIndex] |= ItemFlagUpToDate // mark the item as out of date
			outOfDateCount++
			if !verifyAll {
				break
			}
		} else {
			d.ItemIdFlags[depItemIndex] |= ItemFlagOutOfDate // mark the item as out of date
		}

		// If the dependency is ignored, we do not change the final state
		// but we still need to check other dependencies
		depStart++
	}

	if outOfDateCount == 0 {
		return StateUpToDate, nil // item is up to date, but we need to check dependencies
	}

	// Mark the main item as out-of-date when any of the dependencies where out of date
	d.ItemIdFlags[itemIndex] |= ItemFlagOutOfDate // mark the item as out of date
	return StateOutOfDate, nil
}

// CopyItem copies an item from one depTrackr to another.
func (src *depTrackr) CopyItem(dst *depTrackr, itemHash []byte) error {
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
	dstItemIndex := int32(len(dst.ItemIdFlags))                                       // index of the new item in the destination depTrackr
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

	srcDepItemIndex := src.ItemDepsStart[itemIndex]
	srcDepItemIndexEnd := srcDepItemIndex + itemDepsCount
	for srcDepItemIndex < srcDepItemIndexEnd {

		depItemIdHash := src.ItemIdHash[srcDepItemIndex*src.HashSize : (srcDepItemIndex+1)*src.HashSize]
		depItemChangeHash := src.ItemChangeHash[srcDepItemIndex*src.HashSize : (srcDepItemIndex+1)*src.HashSize]
		depItemIdFlags := src.ItemIdFlags[srcDepItemIndex]
		depItemChangeFlags := src.ItemChangeFlags[srcDepItemIndex]
		depItemIdDataOffset := src.ItemIdDataOffset[srcDepItemIndex]
		depItemIdDataSize := src.ItemIdDataSize[srcDepItemIndex]
		depItemIdData := src.Data[depItemIdDataOffset : depItemIdDataOffset+depItemIdDataSize]
		depItemChangeDataOffset := src.ItemChangeDataOffset[srcDepItemIndex]
		depItemChangeDataSize := src.ItemChangeDataSize[srcDepItemIndex]
		depItemChangeData := src.Data[depItemChangeDataOffset : depItemChangeDataOffset+depItemChangeDataSize]

		dstDepItemIndex := int32(len(dst.ItemIdFlags))                                    // index of the new dependency item in the destination depTrackr
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

		dst.Deps = append(dst.Deps, dstDepItemIndex) // add the dependency item index to the destination depTrackr
		srcDepItemIndex++
	}

	return nil
}
