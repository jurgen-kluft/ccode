package dep

import (
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"
)

type DepTrackr struct {
	StoragePath string
	Items       []Item
	Deps        []int32
	DataNodes   []DataNode
	Data        []byte
	HashDB      *HashDB
}

func NewDepTrackr(storageDir string) *DepTrackr {
	// Note: We initialize many slices already with size = 1, this is because
	//       we identify index = 0 as nil
	return &DepTrackr{
		StoragePath: storageDir,
		Items:       make([]Item, 0, 1024),      // initial capacity of 1024 items
		Deps:        make([]int32, 0, 1024),     // initial capacity of 1024 dependencies
		DataNodes:   make([]DataNode, 0, 1024),  // initial capacity of 1024 data nodes
		Data:        make([]byte, 0, 1024*1024), // initial capacity of 1MB for data
		HashDB:      NewHashDB(10),              // default to 10 bits for the hash database
	}
}

func (d *DepTrackr) NewDB() *DepTrackr {
	// Based on the current sizes of Items, Deps, DataNodes, and Data,
	// we will initialize the DepTrackr with empty slices and a new HashDB.
	nd := NewDepTrackr(d.StoragePath)
	nd.Items = make([]Item, 0, cap(d.Items))
	nd.Deps = make([]int32, 0, cap(d.Deps))
	nd.DataNodes = make([]DataNode, 0, cap(d.DataNodes))
	nd.Data = make([]byte, 0, cap(d.Data))
	nd.HashDB = NewHashDB(d.HashDB.N)
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

func castByteArrayToItemArray(s []byte) []Item {
	sizeOfItem := unsafe.Sizeof(Item{})
	if len(s) == 0 || len(s)&(int(sizeOfItem)-1) != 0 {
		return nil
	}

	return unsafe.Slice((*Item)(unsafe.Pointer(&s[0])), len(s)/int(sizeOfItem))
}

func castItemArrayToByteArray(items []Item) []byte {
	n := len(items)
	if n == 0 {
		return []byte{}
	}
	sizeOfItem := unsafe.Sizeof(Item{})
	return unsafe.Slice((*byte)(unsafe.Pointer(&items[0])), n*int(sizeOfItem))
}

func castHashNodeArrayToByteArray(items []HashNode) []byte {
	n := len(items)
	if n == 0 {
		return []byte{}
	}
	sizeOfItem := unsafe.Sizeof(HashNode{})
	return unsafe.Slice((*byte)(unsafe.Pointer(&items[0])), n*int(sizeOfItem))
}

func castDataNodeArrayToByteArray(items []DataNode) []byte {
	n := len(items)
	if n == 0 {
		return []byte{}
	}
	sizeOfItem := unsafe.Sizeof(DataNode{})
	return unsafe.Slice((*byte)(unsafe.Pointer(&items[0])), n*int(sizeOfItem))
}

// --------------------------------------------------------------------------
func (d *DepTrackr) Save() error {
	dbFile, err := os.Create(d.StoragePath + "/deptrackr.point.db")
	if err != nil {
		return err
	}
	defer dbFile.Close()

	header := make([]byte, 0, 64)
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.Items)))     // Number of items
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.Deps)))      // Number of dependencies
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.DataNodes))) // Number of data nodes
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.Data)))      // Size of data
	header = binary.LittleEndian.AppendUint32(header, uint32(d.HashDB.N))       // Number of bits for the hash
	header = binary.LittleEndian.AppendUint32(header, uint32(d.HashDB.S))       // Size of a shard
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.HashDB.Hashes)))
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.HashDB.HashNodes)))
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.HashDB.ShardOffsets)))
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.HashDB.ShardSizes)))
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.HashDB.DirtyFlags)))
	header = binary.LittleEndian.AppendUint32(header, uint32(len(d.HashDB.Shards)))

	// Write the header to the file
	if _, err := dbFile.Write(header); err != nil {
		return err
	}

	// Save Items
	itemsData := castItemArrayToByteArray(d.Items)
	if _, err := dbFile.Write(itemsData); err != nil {
		return err
	}

	// Save Deps
	depsData := castInt32ArrayToByteArray(d.Deps)
	if _, err := dbFile.Write(depsData); err != nil {
		return err
	}

	// Save DataNodes
	dataNodesData := castDataNodeArrayToByteArray(d.DataNodes)
	if _, err := dbFile.Write(dataNodesData); err != nil {
		return err
	}

	// Save Data
	if _, err := dbFile.Write(d.Data); err != nil {
		return err
	}

	// Write HashDB.Hashes
	if _, err := dbFile.Write(d.HashDB.Hashes); err != nil {
		return err
	}

	// Write HashDB.HashNodes
	hashNodesData := castHashNodeArrayToByteArray(d.HashDB.HashNodes)
	if _, err := dbFile.Write(hashNodesData); err != nil {
		return err
	}

	// Write HashDB.ShardOffsets
	shardOffsets := castInt32ArrayToByteArray(d.HashDB.ShardOffsets)
	if _, err := dbFile.Write(shardOffsets); err != nil {
		return err
	}

	// Write HashDB.ShardSizes
	shardSizes := castInt32ArrayToByteArray(d.HashDB.ShardSizes)
	if _, err := dbFile.Write(shardSizes); err != nil {
		return err
	}

	// Write HashDB.DirtyFlags
	if _, err := dbFile.Write(d.HashDB.DirtyFlags); err != nil {
		return err
	}

	// Write HashDB.Shards
	shardsData := castInt32ArrayToByteArray(d.HashDB.Shards)
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
	IdDigest     []byte // SHA1 20 bytes
	IdData       []byte
	ChangeDigest []byte // SHA1 20 bytes
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

type VerifyItemFunc func(itemChangeFlags uint32, itemChangeData []byte, itemIdFlags uint32, itemIdData []byte) State

// -----------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------

const NilIndex = int32(-1)

type Item struct {
	IDHashNodeIndex     int32  // Hash, this is the ID of the item (filepath, label (e.g. 'MSVC C++ compiler cmd-line arguments))
	ChangeHashNodeIndex int32  // Hash, this identifies the 'change' (modification-time, file-size, file-content, command-line arguments, string, etc..)
	ArrayStart          uint32 // start of dependencies
	ArrayLength         uint32 // length of dependencies
}

type HashNode struct {
	HashOffset int32 // hash offset in the Digests array, 4 bytes
	DataIndex  int32 // data that gave us the hash, 4 bytes (0 means no data)
	ItemIndex  int32 // item that this hash belongs to, 4 bytes (help with swap remove)
	HashSize   int32 // We could support different hash sizes
}

type DataNode struct {
	Length int32
	Flags  uint32
	Offset uint32
}

func CompareDigest(a []byte, b []byte) int {
	for i := range 20 {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

type HashDB struct {
	N            int32      // how many bits we take from the hash to index into the buckets (0-15)
	S            int32      // size of a shard, this is the number of hashes per shardOffset, default is 512
	Hashes       []byte     // This is a byte array of 20 bytes SHA1 digests
	HashNodes    []HashNode // This is an array of HashNodes, each HashNode contains a hash offset, data index, item index, and a dummy value
	ShardOffsets []int32    // This is an array of offsets into the Shards array, each offset corresponds to a shard, 0 means the shard doesn't exist yet
	ShardSizes   []int32    // This is an array of sizes of each shard, 0 means the shard is empty
	DirtyFlags   []uint8    // A bit per shard, indicates if the shard is dirty and needs to be sorted
	Shards       []int32    // A shard is a region of hash-node-indices that belong to a shard
	EmptyShard   []int32    // A shard initialized to a size of S and full of zeros
}

func NewHashDB(n int32) *HashDB {
	if n < 0 || n > 15 {
		n = 10
	}

	s := int32(512) // maximum size of a shard

	hb := &HashDB{
		N:            n,
		S:            s,
		Hashes:       make([]byte, 0, 128*1024*20),  // initial capacity of 128KB for hashes (20 bytes each)
		HashNodes:    make([]HashNode, 0, 128*1024), // initial capacity of 128KB for hash nodes
		ShardOffsets: make([]int32, 1<<n),           // initial capacity of 2^N buckets
		ShardSizes:   make([]int32, 1<<n),           // initial capacity of 2^N buckets sizes
		DirtyFlags:   make([]uint8, (n+7)>>3),       // initial capacity of N bits for dirty flags (rounded up to the nearest byte)
		Shards:       make([]int32, 0, (1<<n)*s),    // initial capacity of 2^N shards where each shard has s elements
		EmptyShard:   make([]int32, s, s),           // an empty shard, used to copy when making a new shard in Shards
	}

	// Set the content of an EmptyShard to -1
	for i, _ := range hb.EmptyShard {
		hb.EmptyShard[i] = NilIndex
	}

	return hb
}

func (hb *HashDB) IsShardDirty(shardOffset int32) bool {
	return (hb.DirtyFlags[shardOffset>>3] & (1 << (shardOffset & 7))) != 0 // check the dirty flag for the shard
}

func (hb *HashDB) SetShardAsDirty(shardOffset int32) {
	hb.DirtyFlags[shardOffset>>3] |= 1 << (shardOffset & 7) // set the dirty flag for the shard
}
func (hb *HashDB) SetShardAsSorted(shardOffset int32) {
	hb.DirtyFlags[shardOffset>>3] = hb.DirtyFlags[shardOffset>>3] &^ (1 << (shardOffset & 7)) // clear the dirty flag for the shard
}

func (hb *HashDB) InsertHash(hash []byte, data int32, item int32, flags uint32) (int32, int32) {
	indexOfShard := NilIndex
	if hb.N < 8 {
		indexOfShard = int32(hash[0]) >> (8 - hb.N) // use the first N bits of the hash
	} else if hb.N < 16 {
		indexOfShard = int32(hash[0])<<8 | int32(hash[1]) // use the first N bits of the hash
		indexOfShard = indexOfShard >> (16 - hb.N)        // shift to get the right index
	}

	if hb.ShardOffsets[indexOfShard] == NilIndex {
		hb.ShardSizes[indexOfShard] = 0                       // initialize the shard size
		hb.ShardOffsets[indexOfShard] = int32(len(hb.Shards)) // initialize the shard offset
		hb.Shards = append(hb.Shards, hb.EmptyShard...)       // initialize the shard, all of them set to -1
	}
	hashIndex := hb.AddHashToShard(hb.ShardOffsets[indexOfShard], hash, data, item, flags)
	return indexOfShard, hashIndex
}

func (hb *HashDB) HashExists(hash []byte) (int32, int32) {
	indexOfShard := NilIndex
	if hb.N < 8 {
		indexOfShard = int32(hash[0]) >> (8 - hb.N) // use the first N bits of the hash
	} else if hb.N < 16 {
		indexOfShard = int32(hash[0])<<8 | int32(hash[1]) // use the first N bits of the hash
		indexOfShard = indexOfShard >> (16 - hb.N)        // shift to get the right index
	}

	shardOffset := hb.ShardOffsets[indexOfShard]
	if shardOffset == NilIndex {
		return NilIndex, NilIndex // shard doesn't exist, so the hash cannot exist
	}

	shardIsDirty := hb.IsShardDirty(indexOfShard)
	if shardIsDirty {
		// Sort the indices based on the hashes
		// The indices are stored in hb.Shards from [indexOfShard * hb.S : (indexOfShard + 1) * hb.S]

		// Mark the shard as sorted
		hb.SetShardAsSorted(indexOfShard)
	}

	// Binary search for the hash in the sorted array
	indexOfHashInShard := int32(0)
	low, high := int32(0), hb.ShardSizes[indexOfShard]-1
	for low <= high {
		mid := (low + high) / 2
		mido := hb.Hashes[hb.Shards[mid]]
		c := CompareDigest(hb.Hashes[mido:mido+20], hash)
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
		return int32(indexOfShard), indexOfHashInShard
	}
	return NilIndex, NilIndex
}

func (hb *HashDB) AddHashToShard(shardOffset int32, hash []byte, data int32, item int32, flags uint32) int32 {
	newHashOffset := int32(len(hb.Hashes))
	newHashNodeIndex := int32(len(hb.HashNodes))
	hb.HashNodes = append(hb.HashNodes, HashNode{HashOffset: newHashOffset, DataIndex: data, ItemIndex: item})
	hb.Hashes = append(hb.Hashes, hash...) // append the new hash to the hash byte array
	shardSize := hb.ShardSizes[shardOffset]
	hb.Shards[shardOffset+shardSize] = newHashNodeIndex // add the new hash node index to the shard
	hb.ShardSizes[shardOffset] = shardSize + 1          // increment the size of the shard
	hb.SetShardAsDirty(shardOffset)
	return newHashNodeIndex
}

func (d *DepTrackr) AddItem(item ItemToAdd, deps []ItemToAdd) bool {
	itemIndex := int32(len(d.Items))

	idHashShardIndex, idHashShardHashIndex := d.HashDB.HashExists(item.IdDigest)
	if idHashShardIndex > 0 {
		// This should not happen, as we are inserting a new item
		return false
	}
	idHashShardIndex, idHashShardHashIndex = d.HashDB.InsertHash(item.IdDigest, int32(len(d.DataNodes)), itemIndex, uint32(item.IdFlags))
	idHashNodeIndex := int32(d.HashDB.Shards[idHashShardIndex*d.HashDB.S+idHashShardHashIndex])

	changeHashNodeIndex := NilIndex
	if item.ChangeDigest != nil || len(item.ChangeDigest) > 0 {
		changeHashShardIndex, changeHashHashIndex := d.HashDB.HashExists(item.ChangeDigest)
		if changeHashShardIndex != NilIndex {
			// This should not happen, as we are inserting a new item
			return false
		}
		changeHashShardIndex, changeHashHashIndex = d.HashDB.InsertHash(item.ChangeDigest, int32(len(d.DataNodes)+1), itemIndex, uint32(item.ChangeFlags))
		changeHashNodeIndex = int32(d.HashDB.Shards[changeHashShardIndex*d.HashDB.S+changeHashHashIndex])
	} else {
		hashNode := HashNode{
			HashOffset: NilIndex,                    // No hash for the change, so we set it to 0
			DataIndex:  int32(len(d.DataNodes) + 1), // This will be the index of the DataNode for the change
			ItemIndex:  itemIndex,                   // This is the index of the item that this hash belongs to
			HashSize:   0,
		}
		d.HashDB.HashNodes = append(d.HashDB.HashNodes, hashNode) // add the new hash node
		changeHashNodeIndex = int32(len(d.HashDB.HashNodes) - 1)  // this is the index of the new hash node
	}

	// DataNode for the Item Id
	idDataNode := DataNode{
		Length: int32(len(item.IdData)),
		Flags:  uint32(item.IdFlags),
		Offset: uint32(len(d.Data)),
	}
	d.DataNodes = append(d.DataNodes, idDataNode) // add the DataNode for the Change
	d.Data = append(d.Data, item.IdData...)       // add the Item data to the Data array

	// DataNode for the Item Change
	changeDataNode := DataNode{
		Length: int32(len(item.ChangeData)),
		Flags:  uint32(item.ChangeFlags),
		Offset: uint32(len(d.Data)),
	}
	d.DataNodes = append(d.DataNodes, changeDataNode) // add the DataNode for the Change
	d.Data = append(d.Data, item.ChangeData...)       // add the Item data to the Data array

	depArrayStart := len(d.Deps) // this may be the start of the array of dependencies
	d.Items = append(d.Items, Item{
		IDHashNodeIndex:     idHashNodeIndex,
		ChangeHashNodeIndex: changeHashNodeIndex,
		ArrayStart:          uint32(depArrayStart),
		ArrayLength:         uint32(len(deps)),
	})

	// Insert dependencies
	// Note: dependencies as an Item are shared
	for _, dep := range deps {
		depItemIndex := int32(len(d.Items))

		depIdHashShardIndex, depIdHashHashIndex := d.HashDB.HashExists(dep.IdDigest)
		if depIdHashShardIndex > 0 {
			depHashNodeIndex := d.HashDB.Shards[depIdHashShardIndex*d.HashDB.S+depIdHashHashIndex]
			depItemIndex = d.HashDB.HashNodes[depHashNodeIndex].ItemIndex
		} else {

			// Need to build a new dependency item
			depIdHashShardIndex, depIdHashHashIndex = d.HashDB.InsertHash(dep.IdDigest, int32(len(d.DataNodes)), depItemIndex, uint32(dep.IdFlags))
			depIdHashNodeIndex := int32(d.HashDB.Shards[depIdHashShardIndex*d.HashDB.S+depIdHashHashIndex])

			depChangeHashBucketIndex, depChangeHashHashIndex := d.HashDB.HashExists(dep.IdDigest)
			if depChangeHashBucketIndex > 0 {
				// This should not happen, as we are inserting a new item
				return false
			}
			depChangeHashBucketIndex, depChangeHashHashIndex = d.HashDB.InsertHash(dep.ChangeDigest, int32(len(d.DataNodes)+1), depItemIndex, uint32(dep.ChangeFlags))
			depChangeHashNodeIndex := int32(d.HashDB.Shards[depChangeHashBucketIndex*d.HashDB.S+depChangeHashHashIndex])

			// DataNode for the Dependency Id
			depIdDataNode := DataNode{
				Length: int32(len(dep.IdData)),
				Flags:  uint32(dep.IdFlags),
				Offset: uint32(len(d.Data)),
			}
			d.DataNodes = append(d.DataNodes, depIdDataNode) // add the DataNode for the Dependency Id
			d.Data = append(d.Data, dep.IdData...)           // add the Dependency Id data to the Data array

			// DataNode for the Dependency Change
			depChangeDataNode := DataNode{
				Length: int32(len(dep.ChangeData)),
				Flags:  uint32(dep.ChangeFlags),
				Offset: uint32(len(d.Data)),
			}
			d.DataNodes = append(d.DataNodes, depChangeDataNode) // add the DataNode for the Dependency Change
			d.Data = append(d.Data, dep.ChangeData...)           // add the Dependency Change data to the Data array

			d.Items = append(d.Items, Item{
				IDHashNodeIndex:     depIdHashNodeIndex,
				ChangeHashNodeIndex: depChangeHashNodeIndex,
				ArrayStart:          0, // Dependencies do not have dependencies, so we set the start to 0
				ArrayLength:         0,
			})
		}

		d.Deps = append(d.Deps, depItemIndex) // add the dependency item index
	}

	return true
}

func (d *DepTrackr) QueryItem(itemHash []byte, verifyAll bool, verifyCb VerifyItemFunc) (State, error) {

	idHashShardIndex, idHashShardHashIndex := d.HashDB.HashExists(itemHash)
	if idHashShardIndex < 0 {
		return StateOutOfDate, nil // item not found
	}

	itemIdHashNodeIndex := d.HashDB.Shards[idHashShardIndex*d.HashDB.S+idHashShardHashIndex]
	itemIdHashNode := d.HashDB.HashNodes[itemIdHashNodeIndex]
	itemIndex := itemIdHashNode.ItemIndex
	itemDataIndex := itemIdHashNode.DataIndex

	changeHashNodeIndex := d.Items[itemIndex].ChangeHashNodeIndex
	changeHashNode := d.HashDB.HashNodes[changeHashNodeIndex]

	idData := d.DataNodes[itemDataIndex]
	idDataData := d.Data[idData.Offset : int32(idData.Offset)+idData.Length]
	idDataFlags := idData.Flags

	changeData := d.DataNodes[changeHashNode.DataIndex]

	changeDataData := d.Data[changeData.Offset : int32(changeData.Offset)+changeData.Length]
	changeDataFlags := changeData.Flags

	state := verifyCb(changeDataFlags, changeDataData, idDataFlags, idDataData)
	if state == StateOutOfDate && !verifyAll {
		return StateOutOfDate, nil // item is out of date (exit early)
	}

	finalState := state

	depArrayStart := d.Items[itemIndex].ArrayStart
	depArrayEnd := depArrayStart + d.Items[itemIndex].ArrayLength
	for depArrayStart < depArrayEnd {
		depItemIndex := d.Deps[depArrayStart]

		depItemDataHashNode := d.HashDB.HashNodes[d.Items[depItemIndex].ChangeHashNodeIndex]
		depItemDataNode := d.DataNodes[depItemDataHashNode.DataIndex]
		depItemDataData := d.Data[depItemDataNode.Offset : int32(depItemDataNode.Offset)+depItemDataNode.Length]
		depItemDataFlags := depItemDataNode.Flags

		depItemIdHashNodeIndex := d.Items[depItemIndex].IDHashNodeIndex
		depItemIdHashNode := d.HashDB.HashNodes[depItemIdHashNodeIndex]
		depItemIdData := d.DataNodes[depItemIdHashNode.DataIndex]
		depItemIdDataData := d.Data[depItemIdData.Offset : int32(depItemIdData.Offset)+depItemIdData.Length]
		depItemIdDataFlags := depItemIdData.Flags

		state = verifyCb(depItemDataFlags, depItemDataData, depItemIdDataFlags, depItemIdDataData)
		if state == StateOutOfDate {
			if !verifyAll {
				return StateOutOfDate, nil // dependency is out of date (exit early)
			}
			finalState = StateOutOfDate // remember the final state
		}

		depArrayStart += 1
	}

	return finalState, nil // item is up to date, but we need to check dependencies
}

// CopyItem copies an item from one DepTrackr to another.
func (src *DepTrackr) CopyItem(dst *DepTrackr, itemHash []byte) error {

	// Note: Item and other indices are not identical between the two DepTrackrs,
	// so we need to find the item in the source DepTrackr and add it to the destination DepTrackr.

	idHashShardIndex, idHashShardHashIndex := src.HashDB.HashExists(itemHash)
	if idHashShardIndex < 0 {
		return fmt.Errorf("item with hash %x doesn't exists in the source DepTrackr", itemHash)
	}

	srcItemHashNodeIndex := src.HashDB.Shards[idHashShardIndex*src.HashDB.S+idHashShardHashIndex]
	srcItemHashNode := &src.HashDB.HashNodes[srcItemHashNodeIndex]
	srcItem := &src.Items[srcItemHashNode.ItemIndex]
	srcItemDataNode := &src.DataNodes[srcItemHashNode.DataIndex]

	srcChangeHashNodeIndex := srcItem.ChangeHashNodeIndex
	srcChangeHashNode := &src.HashDB.HashNodes[srcChangeHashNodeIndex]
	srcChangeDataNode := &src.DataNodes[srcChangeHashNode.DataIndex]
	srcChangeHash := src.HashDB.Hashes[srcChangeHashNode.HashOffset : srcChangeHashNode.HashOffset+20]

	// Create a destination main item
	dstMainItemIndex := int32(len(dst.Items))
	dstMainIdHashShardIndex, dstMainIdHashShardHashIndex := dst.HashDB.HashExists(itemHash)
	if dstMainIdHashShardIndex != NilIndex {
		// The item already exists in the destination DepTrackr, so we just return
		return fmt.Errorf("item with hash %x already exists in the destination DepTrackr", itemHash)
	}

	dstMainIdHashShardIndex, dstMainIdHashShardHashIndex = dst.HashDB.InsertHash(itemHash, int32(len(dst.DataNodes)), dstMainItemIndex, srcItemDataNode.Flags)
	dstMainIdHashNodeIndex := int32(dst.HashDB.Shards[dstMainIdHashShardIndex*dst.HashDB.S+dstMainIdHashShardHashIndex])

	dstMainChangeHashShardIndex, dstMainChangeHashShardHashIndex := dst.HashDB.HashExists(srcChangeHash)
	dstMainChangeHashNodeIndex := src.HashDB.Shards[dstMainChangeHashShardIndex*dst.HashDB.S+dstMainChangeHashShardHashIndex]

	// DataNode for the Item Id
	dstMainIdDataNode := DataNode{
		Length: srcItemDataNode.Length,
		Flags:  srcItemDataNode.Flags,
		Offset: uint32(len(dst.Data)),
	}
	dst.DataNodes = append(dst.DataNodes, dstMainIdDataNode)                                                              // add the DataNode for the Item Id
	dst.Data = append(dst.Data, src.Data[srcItemDataNode.Offset:int32(srcItemDataNode.Offset)+srcItemDataNode.Length]...) // add the Item data to the Data array

	// DataNode for the Item Change
	dstMainChangeDataNode := DataNode{
		Length: srcChangeDataNode.Length,
		Flags:  srcChangeDataNode.Flags,
		Offset: uint32(len(dst.Data)),
	}
	dst.DataNodes = append(dst.DataNodes, dstMainChangeDataNode)                                                                // add the DataNode for the Item Change
	dst.Data = append(dst.Data, src.Data[srcChangeDataNode.Offset:int32(srcChangeDataNode.Offset)+srcChangeDataNode.Length]...) // add the Item change data to the Data array

	// Allocate the dependency array in the destination DepTrackr
	depArrayStart := len(dst.Deps)
	depArrayEnd := depArrayStart + int(srcItem.ArrayLength)

	// Create the main item in the destination DepTrackr
	dstMainItem := Item{
		IDHashNodeIndex:     dstMainIdHashNodeIndex,
		ChangeHashNodeIndex: dstMainChangeHashNodeIndex,
		ArrayStart:          uint32(depArrayStart),
		ArrayLength:         uint32(srcItem.ArrayLength),
	}
	dst.Items = append(dst.Items, dstMainItem) // add the main item to the destination DepTrackr

	// Now we need to copy the dependencies of the item from the source DepTrackr to the destination DepTrackr
	for depArrayStart < depArrayEnd {
		depSrcItemIndex := src.Deps[depArrayStart]
		depSrcItem := &src.Items[depSrcItemIndex]

		depSrcIdHashIndex := depSrcItem.IDHashNodeIndex
		depSrcIdHashNode := &src.HashDB.HashNodes[depSrcIdHashIndex]
		depSrcIdHash := src.HashDB.Hashes[depSrcIdHashNode.HashOffset : depSrcIdHashNode.HashOffset+20]

		depSrcIdDataNode := &src.DataNodes[depSrcIdHashNode.DataIndex]

		depSrcChangeHashIndex := depSrcItem.ChangeHashNodeIndex
		depSrcChangeHashNode := &src.HashDB.HashNodes[depSrcChangeHashIndex]
		depSrcChangeDataNode := &src.DataNodes[depSrcChangeHashNode.DataIndex]

		depDstIdHashShardIndex, depDstIdHashShardHashIndex := dst.HashDB.HashExists(depSrcIdHash)

		// The dep item either exists or it doesn't yet exist in which case we need to create it
		depDstItemIndex := int32(0)
		if depDstIdHashShardIndex == NilIndex {
			// The dependency item doesn't exist, so we need to create it
			depDstItemIndex = int32(len(dst.Items))

			depDstIdHashShardIndex, depDstIdHashShardHashIndex = dst.HashDB.InsertHash(depSrcIdHash, int32(len(dst.DataNodes)), depDstItemIndex, depSrcIdDataNode.Flags)
			depDstIdHashNodeIndex := int32(dst.HashDB.Shards[depDstIdHashShardIndex*dst.HashDB.S+depDstIdHashShardHashIndex])

			if depSrcIdHashNode.DataIndex != NilIndex {
				depSrcIdDataNode := &src.DataNodes[depSrcIdHashNode.DataIndex]
				depDstIdDataNodeIndex := int32(len(dst.DataNodes))
				dst.DataNodes = append(dst.DataNodes, DataNode{
					Length: depSrcIdDataNode.Length,
					Flags:  depSrcIdDataNode.Flags,
					Offset: uint32(len(dst.Data)),
				})
				dst.Data = append(dst.Data, src.Data[depSrcIdDataNode.Offset:int32(depSrcIdDataNode.Offset)+depSrcIdDataNode.Length]...)
				dst.HashDB.HashNodes[depDstIdHashNodeIndex].DataIndex = depDstIdDataNodeIndex // update the data index of the hash node
			}

			depSrcChangeHashIndex := depSrcItem.ChangeHashNodeIndex
			depSrcChangeHashNode := &src.HashDB.HashNodes[depSrcChangeHashIndex]

			depDstChangeDataNodeIndex := NilIndex
			if depSrcChangeHashNode.DataIndex != NilIndex {
				depSrcChangeDataNode := &src.DataNodes[depSrcChangeHashNode.DataIndex]
				depDstChangeDataNodeIndex = int32(len(dst.DataNodes))
				dst.DataNodes = append(dst.DataNodes, DataNode{
					Length: depSrcChangeDataNode.Length,
					Flags:  depSrcChangeDataNode.Flags,
					Offset: uint32(len(dst.Data)),
				})
				dst.Data = append(dst.Data, src.Data[depSrcChangeDataNode.Offset:int32(depSrcChangeDataNode.Offset)+depSrcChangeDataNode.Length]...)
			}

			depDstChangeHashNodeIndex := NilIndex
			if depSrcChangeHashNode.HashSize > 0 {
				depSrcChangeHash := src.HashDB.Hashes[depSrcChangeHashNode.HashOffset : depSrcChangeHashNode.HashOffset+depSrcChangeHashNode.HashSize]
				depDstChangeHashShardIndex, depDstChangeHashShardHashIndex := dst.HashDB.InsertHash(depSrcChangeHash, depDstChangeDataNodeIndex, depDstItemIndex, depSrcChangeDataNode.Flags)
				depDstChangeHashNodeIndex = int32(dst.HashDB.Shards[depDstChangeHashShardIndex*dst.HashDB.S+depDstChangeHashShardHashIndex])
			} else {
				depDstChangeHashNodeIndex = int32(len(dst.HashDB.HashNodes))
				dst.HashDB.HashNodes = append(dst.HashDB.HashNodes, HashNode{
					HashOffset: 0, // No hash for the change, so we set it to 0
					DataIndex:  depDstChangeDataNodeIndex,
					ItemIndex:  depDstItemIndex, // This is the index of the item that this hash belongs to
					HashSize:   0,               // Dummy value, not used
				})
			}

			dst.Items = append(dst.Items, Item{
				IDHashNodeIndex:     depDstIdHashNodeIndex,
				ChangeHashNodeIndex: depDstChangeHashNodeIndex,
				ArrayStart:          0, // Dependencies do not have dependencies, so we set the start to 0
				ArrayLength:         0, // No dependencies for the dependency item
			})

		} else {
			depDstHashNodeIndex := dst.HashDB.Shards[depDstIdHashShardIndex*dst.HashDB.S+depDstIdHashShardHashIndex]
			depDstHashNode := &dst.HashDB.HashNodes[depDstHashNodeIndex]
			depDstItemIndex = depDstHashNode.ItemIndex
		}

		dst.Deps = append(dst.Deps, depDstItemIndex) // add the dependency item index

		depArrayStart += 1
	}

	return nil
}
