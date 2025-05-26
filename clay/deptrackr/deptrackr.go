package dep

type DepTrackr struct {
	StoragePath string

	Items  *ItemDb
	Nodes  *NodeDb
	Datas  *DataDb
	HashDB HashDB
}

func NewDepTrackr(storageDir string) *DepTrackr {
	return &DepTrackr{
		StoragePath: storageDir,
	}
}

func (d *DepTrackr) Load() error {

	return nil
}

func (d *DepTrackr) Save() error {

	return nil
}

type Digest struct {
	Hash [20]byte
}

type ItemToAdd struct {
	IdData     []byte
	IdDigest   Digest
	ItemData   []byte
	ItemDigest Digest
	Flags      uint32
}

type State int

const (
	StateUpToDate State = iota
	StateOutOfDate
)

type VerifyItemFunc func(itemFlags uint32, itemData []byte, itemDigest Digest) State

// -----------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------

type Item struct {
	ID       uint32 // Hash, this is the ID of the item (filepath, label (e.g. 'MSVC C++ compiler cmd-line arguments))
	Change   uint32 // Hash, this identifies the 'change' (modification-time, file-size, file-content, command-line arguments, string, etc..)
	ListHead uint32 // head of list of dependencies
}

type Node struct {
	Item uint32 // Item, this is the item this node holds
	Next uint32 // list, next
	Prev uint32 // list, prev
}

type Hash struct {
	Hash  Digest // hash value, 20 bytes
	Data  uint32 // data that gave us the hash, 4 bytes
	Item  uint32 // item that this hash belongs to, 4 bytes (help with swap remove)
	Flags uint32 // useful flags, 4 bytes (marked as modified)
}

type Data struct {
	Length uint32
	Data   []byte
}

func CompareDigest(a, b Digest) int {
	for i := 0; i < len(a.Hash); i++ {
		if a.Hash[i] < b.Hash[i] {
			return -1
		} else if a.Hash[i] > b.Hash[i] {
			return 1
		}
	}
	return 0 // equal
}

type ItemDb struct {
	Items []Item
}

type NodeDb struct {
	Nodes []Node
}

type DataDb struct {
	Datas []Data
}

type HashDB struct {
	N       int32         // how many bits we take from the hash to index into the buckets (0-15)
	Buckets []*HashBucket // array of buckets, each bucket is an array of hashes
}

func NewHashDB(n int32) *HashDB {
	if n < 0 || n > 15 {
		n = 10
	}
	return &HashDB{
		N:       n,
		Buckets: make([]*HashBucket, 1<<n), // 2^n buckets
	}
}

func (hb *HashDB) InsertHash(hash Digest, data uint32, item uint32, flags uint32) (int16, int16) {
	indexOfBucket := 0
	if hb.N < 8 {
		indexOfBucket = int(hash.Hash[0]) >> (8 - hb.N) // use the first N bits of the hash
	} else if hb.N < 16 {
		indexOfBucket = int(hash.Hash[0])<<8 | int(hash.Hash[1]) // use the first N bits of the hash
		indexOfBucket = indexOfBucket >> (16 - hb.N)             // shift to get the right index
	}

	if hb.Buckets[indexOfBucket] == nil {
		hb.Buckets[indexOfBucket] = NewHashBucket(16) // create a new bucket if it doesn't exist
	}
	hashIndex := hb.Buckets[indexOfBucket].InsertHash(hash, data, item, flags)
	return int16(indexOfBucket), hashIndex
}

func (hb *HashDB) HashExists(hash Digest) (int16, int16) {
	indexOfBucket := 0
	if hb.N < 8 {
		indexOfBucket = int(hash.Hash[0]) >> (8 - hb.N) // use the first N bits of the hash
	} else if hb.N < 16 {
		indexOfBucket = int(hash.Hash[0])<<8 | int(hash.Hash[1]) // use the first N bits of the hash
		indexOfBucket = indexOfBucket >> (16 - hb.N)             // shift to get the right index
	}
	bucket := hb.Buckets[indexOfBucket]
	if bucket == nil {
		return -1, -1 // bucket does not exist
	}
	indexOfHash := bucket.IndexOfHash(hash)
	return int16(indexOfBucket), indexOfHash
}

type HashBucket struct {
	Hashes []Hash  // array of hashes in this bucket (max 65535)
	Sorted []int16 // sorted array of indices into the Hashes array, used for fast lookup
	Dirty  bool
}

func NewHashBucket(reserved int) *HashBucket {
	return &HashBucket{
		Hashes: make([]Hash, 0, reserved),  //
		Sorted: make([]int16, 0, reserved), // sorted indices for fast lookup
	}
}

func (hb *HashBucket) InsertHash(hash Digest, data uint32, item uint32, flags uint32) int16 {
	newIndex := int16(len(hb.Hashes))
	hb.Hashes = append(hb.Hashes, Hash{Hash: hash, Data: data, Item: item, Flags: flags})
	hb.Sorted = append(hb.Sorted, int16(newIndex))
	hb.Dirty = true // mark the bucket as dirty
	return newIndex
}

func (hb *HashBucket) IndexOfHash(hash Digest) int16 {
	if hb.Dirty {
		// Sort the hashes if the bucket is dirty
		hb.Sorted = make([]int16, len(hb.Hashes))
		for i := range hb.Hashes {
			hb.Sorted[i] = int16(i)
		}
		// Sort the indices based on the hashes
		for i := 0; i < len(hb.Sorted)-1; i++ {
			for j := i + 1; j < len(hb.Sorted); j++ {
				if CompareDigest(hb.Hashes[hb.Sorted[i]].Hash, hb.Hashes[hb.Sorted[j]].Hash) > 0 {
					hb.Sorted[i], hb.Sorted[j] = hb.Sorted[j], hb.Sorted[i]
				}
			}
		}
		hb.Dirty = false // reset dirty flag
	}

	// Binary search for the hash in the sorted array
	low, high := 0, len(hb.Sorted)-1
	for low <= high {
		mid := (low + high) / 2
		if hb.Hashes[hb.Sorted[mid]].Hash == hash {
			return int16(mid) // found
		} else if CompareDigest(hb.Hashes[hb.Sorted[mid]].Hash, hash) < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return -1 // not found
}

func (d *DepTrackr) Insert(item ItemToAdd, deps []ItemToAdd) bool {
	itemIndex := uint32(len(d.Items.Items))

	idHashBucketIndex, idHashHashIndex := d.HashDB.HashExists(item.IdDigest)
	if idHashBucketIndex >= 0 {
		// This should not happen, as we are inserting a new item
		return false
	}
	idHashBucketIndex, idHashHashIndex = d.HashDB.InsertHash(item.IdDigest, uint32(len(d.Datas.Datas)), itemIndex, item.Flags)

	changeHashBucketIndex, changeHashHashIndex := d.HashDB.HashExists(item.ItemDigest)
	if changeHashBucketIndex >= 0 {
		// This should not happen, as we are inserting a new item
		return false
	}
	changeHashBucketIndex, changeHashHashIndex = d.HashDB.InsertHash(item.ItemDigest, uint32(len(d.Datas.Datas)), itemIndex, item.Flags)

	listHead := len(d.Nodes.Nodes) // this may be the head of the list of dependencies

	d.Items.Items = append(d.Items.Items, Item{
		ID:       uint32(idHashBucketIndex)<<16 | uint32(idHashHashIndex),
		Change:   uint32(changeHashBucketIndex)<<16 | uint32(changeHashHashIndex),
		ListHead: uint32(listHead),
	})

	// Insert dependencies
	// Note: dependencies as an Item are shared
	prevNodeIndex := uint32(0) // (nil) this will be used to link the nodes in the list
	for _, dep := range deps {
		depItemIndex := uint32(len(d.Items.Items))
		nodeIndex := uint32(len(d.Nodes.Nodes))

		depIdHashBucketIndex, depIdHashHashIndex := d.HashDB.HashExists(dep.IdDigest)
		if depIdHashBucketIndex >= 0 {
			depItemIndex = d.HashDB.Buckets[depIdHashBucketIndex].Hashes[depIdHashHashIndex].Item
		} else {
			depIdHashBucketIndex, depIdHashHashIndex = d.HashDB.InsertHash(dep.IdDigest, uint32(len(d.Datas.Datas)), depItemIndex, dep.Flags)
		}

		depChangeHashBucketIndex, depChangeHashHashIndex := d.HashDB.HashExists(dep.ItemDigest)
		if depChangeHashBucketIndex >= 0 {
			// This should not happen, as we are inserting a new item
			return false
		}

		// Do we need to create the new item ?
		if depItemIndex == uint32(len(d.Items.Items)) {
			d.Items.Items = append(d.Items.Items, Item{
				ID:       uint32(depIdHashBucketIndex)<<16 | uint32(depIdHashHashIndex),
				Change:   uint32(depChangeHashBucketIndex)<<16 | uint32(depChangeHashHashIndex),
				ListHead: 0xffffffff, // this is a dependency item
			})
		}

		// Create the node for the dependency
		d.Nodes.Nodes = append(d.Nodes.Nodes, Node{
			Item: depItemIndex,
			Next: 0,             // nil (will be set later)
			Prev: prevNodeIndex, // link to the previous node
		})
		if prevNodeIndex != 0 {
			// Link the previous node to this one
			d.Nodes.Nodes[prevNodeIndex].Next = nodeIndex
		}
		prevNodeIndex = nodeIndex // update the previous node index
	}

	return true
}

func (d *DepTrackr) QueryItem(item Digest, verifyAll bool, verifyCb VerifyItemFunc) (State, error) {
	idHashBucketIndex, idHashHashIndex := d.HashDB.HashExists(item)
	if idHashBucketIndex < 0 {
		return StateOutOfDate, nil // item not found
	}

	changeHashBucketIndex, changeHashHashIndex := d.HashDB.HashExists(item)
	if changeHashBucketIndex < 0 {
		return StateOutOfDate, nil // item not found
	}

	itemData := d.Datas.Datas[d.HashDB.Buckets[idHashBucketIndex].Hashes[idHashHashIndex].Data].Data
	itemFlags := d.HashDB.Buckets[changeHashBucketIndex].Hashes[changeHashHashIndex].Flags

	state := verifyCb(itemFlags, itemData, item)
	if state == StateUpToDate && !verifyAll {
		return StateUpToDate, nil // item is up to date
	}

	finalState := state

	itemIndex := d.HashDB.Buckets[idHashBucketIndex].Hashes[idHashHashIndex].Item

	listHead := d.Items.Items[itemIndex].ListHead
	for listHead != 0 {
		node := d.Nodes.Nodes[listHead]
		depItemData := d.Datas.Datas[d.HashDB.Buckets[idHashBucketIndex].Hashes[node.Item].Data].Data
		depItemFlags := d.HashDB.Buckets[changeHashBucketIndex].Hashes[node.Item].Flags

		state = verifyCb(depItemFlags, depItemData, item)
		if state == StateOutOfDate && !verifyAll {
			return StateOutOfDate, nil // dependency is out of date
		}

		if state == StateOutOfDate {
			finalState = StateOutOfDate // at least one dependency is out of date
		}

		listHead = node.Next // move to the next dependency
	}

	return finalState, nil // item is up to date, but we need to check dependencies
}
