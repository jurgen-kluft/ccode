package dep

import (
	"encoding/binary"
	"os"
	"unsafe"
)

type DepTrackr struct {
	StoragePath          string   // Path to the directory where we store the database file
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
	N                    int32    // how many bits we take from the hash to index into the buckets (0-15)
	S                    int32    // size of a shard, this is the number of hashes per shardOffset, default is 512
	ShardOffsets         []int32  // This is an array of offsets into the Shards array, each offset corresponds to a shard, 0 means the shard doesn't exist yet
	ShardSizes           []int32  // This is an array of sizes of each shard, 0 means the shard is empty
	DirtyFlags           []uint8  // A bit per shard, indicates if the shard is dirty and needs to be sorted
	Shards               []int32  // A shard is a region of item-indices that belong to a shard
	EmptyShard           []int32  // A shard initialized to a size of S and full of zeros
	Deps                 []int32  // Array for each item to list their dependencies, this is a flat array of item indices
	Data                 []byte   // Here the data for Id and Change is stored, this is a flat array of bytes
}

func NewDepTrackr(storageDir string) *DepTrackr {

	reserve := 1024

	n := int32(10)  // how many bits we take from the hash to index into the buckets (0-15)
	s := int32(512) // size of a shard, this is the number of items per shard, default is 512

	d := &DepTrackr{
		StoragePath:          storageDir,
		HashSize:             20,                            // SHA1 hash size is 20 bytes
		ItemIdHash:           make([]byte, 0, reserve),      //
		ItemChangeHash:       make([]byte, 0, reserve),      //
		ItemIdFlags:          make([]uint16, 0, reserve),    //
		ItemChangeFlags:      make([]uint16, 0, reserve),    //
		ItemDepsStart:        make([]int32, 0, reserve),     //
		ItemDepsCount:        make([]int32, 0, reserve),     //
		ItemIdDataOffset:     make([]int32, 0, reserve),     //
		ItemIdDataSize:       make([]int32, 0, reserve),     //
		ItemChangeDataOffset: make([]int32, 0, reserve),     //
		ItemChangeDataSize:   make([]int32, 0, reserve),     //
		N:                    n,                             //
		S:                    s,                             //
		ShardOffsets:         make([]int32, 1<<n),           // initial capacity of 2^N shards
		ShardSizes:           make([]int32, 1<<n),           // initial capacity of 2^N shards
		DirtyFlags:           make([]uint8, (n+7)>>3),       // initial capacity of N bits for dirty flags (rounded up to the nearest byte)
		Shards:               make([]int32, 0, (1<<n)*s),    // initial capacity of 2^N shards where each shard has s elements
		EmptyShard:           make([]int32, s, s),           // an empty shard, used to copy when making a new shard in Shards
		Deps:                 make([]int32, 0, reserve),     //
		Data:                 make([]byte, 0, reserve*1024), //
	}

	// Set the content of an EmptyShard to -1
	for i, _ := range d.EmptyShard {
		d.EmptyShard[i] = NilIndex
	}

	return d
}

func (d *DepTrackr) NewDB() *DepTrackr {
	// Based on the current sizes of Items, Deps, DataNodes, and Data,
	// we will initialize the DepTrackr with empty slices and a new HashDB.
	nd := NewDepTrackr(d.StoragePath)
	nd.Deps = make([]int32, 0, cap(d.Deps))
	nd.Data = make([]byte, 0, cap(d.Data))

	return nd
}

// --------------------------------------------------------------------------
// Helper functions to cast loaded byte arrays to other types
func castByteArrayToInt32Array(s []byte) []int32 {
	if len(s) == 0 || len(s)&3 != 0 {
		return nil
	}
	return unsafe.Slice((*int32)(unsafe.Pointer(&s[0])), len(s)>>2)
}

func castInt32ArrayToByteArray(i []int32) []byte {
	if len(i) == 0 {
		return []byte{}
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&i[0])), len(i)<<2)
}

// --------------------------------------------------------------------------
func (d *DepTrackr) Save() error {
	dbFile, err := os.Create(d.StoragePath + "/deptrackr.point.db")
	if err != nil {
		return err
	}
	defer dbFile.Close()

	header := make([]byte, 0, 64)

	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.ItemIdHash))) // Number of items
	header = binary.LittleEndian.AppendUint32(header, uint32(d.N))               // Number of bits for the hash
	header = binary.LittleEndian.AppendUint32(header, uint32(d.S))               // Size of a shard
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.ShardOffsets)))
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.ShardSizes)))
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.DirtyFlags)))
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.Shards)))
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.Deps)))
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.Data)))

	// Write the header to the file
	if _, err := dbFile.Write(header); err != nil {
		return err
	}

	// Save Items
	itemsData := d.ItemIdHash
	if _, err := dbFile.Write(itemsData); err != nil {
		return err
	}

	// Write HashDB.ShardOffsets
	shardOffsets := castInt32ArrayToByteArray(d.ShardOffsets)
	if _, err := dbFile.Write(shardOffsets); err != nil {
		return err
	}

	// Write HashDB.ShardSizes
	shardSizes := castInt32ArrayToByteArray(d.ShardSizes)
	if _, err := dbFile.Write(shardSizes); err != nil {
		return err
	}

	// Write HashDB.DirtyFlags
	if _, err := dbFile.Write(d.DirtyFlags); err != nil {
		return err
	}

	// Write HashDB.Shards
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

func (d *DepTrackr) Load() error {

	return nil
}

// --------------------------------------------------------------------------
const (
	ItemFlagSourceFile = 1
	ItemFlagDependency = 2
	ItemFlagString     = 8
	ItemFlagUpToDate   = 128 // This is set when the item is up to date, otherwise it is out of date
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

func (d *DepTrackr) CompareDigest(a []byte, b []byte) int {
	for i := range d.HashSize {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func (d *DepTrackr) IsShardDirty(shardOffset int32) bool {
	return (d.DirtyFlags[shardOffset>>3] & (1 << (shardOffset & 7))) != 0 // check the dirty flag for the shard
}

func (d *DepTrackr) SetShardAsDirty(shardOffset int32) {
	d.DirtyFlags[shardOffset>>3] |= 1 << (shardOffset & 7) // set the dirty flag for the shard
}
func (d *DepTrackr) SetShardAsSorted(shardOffset int32) {
	d.DirtyFlags[shardOffset>>3] = d.DirtyFlags[shardOffset>>3] &^ (1 << (shardOffset & 7)) // clear the dirty flag for the shard
}

func (d *DepTrackr) InsertItemIntoDb(hash []byte, item int32) {
	indexOfShard := NilIndex
	if d.N < 8 {
		indexOfShard = int32(hash[0]) >> (8 - d.N) // use the first N bits of the hash
	} else if d.N < 16 {
		indexOfShard = int32(hash[0])<<8 | int32(hash[1]) // use the first N bits of the hash
		indexOfShard = indexOfShard >> (16 - d.N)         // shift to get the right index
	}

	if d.ShardOffsets[indexOfShard] == NilIndex {
		d.ShardSizes[indexOfShard] = 0                      // initialize the shard size
		d.ShardOffsets[indexOfShard] = int32(len(d.Shards)) // initialize the shard offset
		d.Shards = append(d.Shards, d.EmptyShard...)        // initialize the shard, all of them set to -1
	}
	d.AddItemToShard(d.ShardOffsets[indexOfShard], item)
}

func (d *DepTrackr) DoesItemExist(hash []byte) int32 {
	indexOfShard := NilIndex
	if d.N < 8 {
		indexOfShard = int32(hash[0]) >> (8 - d.N) // use the first N bits of the hash
	} else if d.N < 16 {
		indexOfShard = int32(hash[0])<<8 | int32(hash[1]) // use the first N bits of the hash
		indexOfShard = indexOfShard >> (16 - d.N)         // shift to get the right index
	}

	shardOffset := d.ShardOffsets[indexOfShard]
	if shardOffset == NilIndex {
		return NilIndex // shard doesn't exist, so the hash cannot exist
	}

	shardIsDirty := d.IsShardDirty(indexOfShard)
	if shardIsDirty {
		// Sort the indices based on the hashes
		// The indices are stored in d.Shards from [indexOfShard * d.S : (indexOfShard + 1) * d.S]

		// Mark the shard as sorted
		d.SetShardAsSorted(indexOfShard)
	}

	// Binary search for the hash in the sorted array
	indexOfHashInShard := int32(0)
	low, high := int32(0), d.ShardSizes[indexOfShard]-1
	for low <= high {
		mid := (low + high) / 2
		mido := mid * int32(d.HashSize)
		c := d.CompareDigest(d.ItemIdHash[mido:mido+int32(d.HashSize)], hash)
		if c == 0 {
			indexOfHashInShard = mid // found
			break
		} else if c < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	if indexOfHashInShard >= 0 {
		return d.Shards[shardOffset+indexOfHashInShard] // return the index of the item in the shard
	}
	return NilIndex
}

func (d *DepTrackr) AddItemToShard(shardOffset int32, item int32) {
	shardSize := d.ShardSizes[shardOffset]
	d.Shards[shardOffset+shardSize] = item    // add the new item index to the shard
	d.ShardSizes[shardOffset] = shardSize + 1 // increment the size of the shard
	d.SetShardAsDirty(shardOffset)
}

func (d *DepTrackr) AddItem(item ItemToAdd, deps []ItemToAdd) bool {

	existingItemIndex := d.DoesItemExist(item.IdDigest)
	if existingItemIndex != NilIndex {
		// This should not happen, as we are inserting a new item
		return false
	}

	itemIndex := int32(len(d.ItemIdHash))
	d.InsertItemIntoDb(item.IdDigest, itemIndex)

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

	// Insert dependencies
	// Note: dependencies as an Item are shared
	for _, dep := range deps {
		depItemIndex := int32(len(d.ItemIdHash))
		if existingItemIndex := d.DoesItemExist(dep.IdDigest); existingItemIndex != NilIndex {
			depItemIndex = existingItemIndex
		} else {
			// Insert the dependency item into the main arrays
			d.ItemIdHash = append(d.ItemIdHash, dep.IdDigest...)                            // add the dependency Id hash
			d.ItemChangeHash = append(d.ItemChangeHash, dep.ChangeDigest...)                // add the dependency Change hash
			d.ItemIdFlags = append(d.ItemIdFlags, dep.IdFlags)                              // add the dependency Id flags
			d.ItemChangeFlags = append(d.ItemChangeFlags, dep.ChangeFlags)                  // add the dependency Change flags
			d.ItemDepsStart = append(d.ItemDepsStart, int32(len(d.Deps)))                   // start of dependencies is 0
			d.ItemDepsCount = append(d.ItemDepsCount, int32(0))                             // count of dependencies is 0 for now
			d.ItemIdDataOffset = append(d.ItemIdDataOffset, int32(len(d.Data)))             // item Id data
			d.ItemIdDataSize = append(d.ItemIdDataSize, int32(len(dep.IdData)))             // item Id data
			d.Data = append(d.Data, dep.IdData...)                                          // add the Item Id data to the Data array
			d.ItemChangeDataOffset = append(d.ItemChangeDataOffset, int32(len(d.Data)))     // item Id data
			d.ItemChangeDataSize = append(d.ItemChangeDataSize, int32(len(dep.ChangeData))) // item Id data
			d.Data = append(d.Data, dep.ChangeData...)                                      // add the Item Change data to the Data array

			// Add the dependency item to the shard database, so that we can search it
			d.InsertItemIntoDb(dep.IdDigest, depItemIndex)
		}
		d.Deps = append(d.Deps, depItemIndex) // add the dependency item index
	}

	return true
}

func (d *DepTrackr) QueryItem(itemHash []byte, verifyAll bool, verifyCb VerifyItemFunc) (State, error) {

	itemIndex := d.DoesItemExist(itemHash)
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
	if itemState == StateOutOfDate && !verifyAll {
		return StateOutOfDate, nil
	}

	finalState := StateOutOfDate

	// Check the dependencies
	depStart := d.ItemDepsStart[itemIndex]
	depEnd := depStart + d.ItemDepsCount[itemIndex]
	for depStart < depEnd {

		depChangeFlags := d.ItemChangeFlags[d.Deps[depStart]]
		depChangeData := d.Data[d.ItemChangeDataOffset[d.Deps[depStart]] : d.ItemChangeDataOffset[d.Deps[depStart]]+d.ItemChangeDataSize[d.Deps[depStart]]]
		depIdFlags := d.ItemIdFlags[d.Deps[depStart]]
		depIdData := d.Data[d.ItemIdDataOffset[d.Deps[depStart]] : d.ItemIdDataOffset[d.Deps[depStart]]+d.ItemIdDataSize[d.Deps[depStart]]]

		depState := verifyCb(depChangeFlags, depChangeData, depIdFlags, depIdData)
		if depState == StateOutOfDate {
			if !verifyAll {
				return StateOutOfDate, nil // if we are not verifying all, we can return early
			}
			finalState = StateOutOfDate // at least one dependency is out of date
		}

		// If the dependency is ignored, we do not change the final state
		// but we still need to check other dependencies
		depStart++
	}

	return finalState, nil // item is up to date, but we need to check dependencies
}

// CopyItem copies an item from one DepTrackr to another.
func (src *DepTrackr) CopyItem(dst *DepTrackr, itemHash []byte) error {

	itemIndex := src.DoesItemExist(itemHash)
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
	dstItemIndex := int32(len(dst.ItemIdHash))                                        // index of the new item in the destination DepTrackr
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
	dst.InsertItemIntoDb(itemIdHash, dstItemIndex)

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

		dstDepItemIndex := int32(len(dst.ItemIdHash))                                     // index of the new dependency item in the destination DepTrackr
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
		dst.InsertItemIntoDb(depItemIdHash, dstDepItemIndex)

		dst.Deps = append(dst.Deps, dstDepItemIndex) // add the dependency item index to the destination DepTrackr
		srcDepItemIndex++
	}

	return nil
}
