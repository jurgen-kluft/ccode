package dep

import (
	"crypto/sha1"
	"testing"
	"time"
)

func TestDepTrackr(t *testing.T) {
	tracker := NewDepTrackr("testdb")

	hasher := sha1.New()

	itemData := []byte("test/test.cpp")
	hasher.Reset()
	hasher.Write(itemData)
	itemHash := hasher.Sum(nil)

	changeData := []byte(time.Now().String())
	hasher.Reset()
	hasher.Write(changeData)
	changeHash := hasher.Sum(nil)

	mainItem := ItemToAdd{
		IdDigest:     itemHash,
		IdData:       itemData,
		ChangeDigest: changeHash,
		ChangeData:   changeData,
		IdFlags:      33,
		ChangeFlags:  44,
	}

	depItemData := []byte("test/dependency.cpp")
	hasher.Reset()
	hasher.Write(depItemData)
	depItemHash := hasher.Sum(nil)

	depChangeData := []byte(time.Now().String())
	hasher.Reset()
	hasher.Write(depChangeData)
	depChangeHash := hasher.Sum(nil)

	depItem := ItemToAdd{
		IdDigest:     depItemHash,
		IdData:       depItemData,
		ChangeDigest: depChangeHash,
		ChangeData:   depChangeData,
		IdFlags:      55,
		ChangeFlags:  66,
	}

	// Test adding a dependency
	added := tracker.AddItem(mainItem, []ItemToAdd{depItem})
	if !added {
		t.Fatalf("Failed to add item")
	}

	// Query the main item
	mainItemState, err := tracker.QueryItem(mainItem.IdDigest, true, func(itemChangeFlags uint32, itemChangeData []byte, itemIdFlags uint32, itemIdData []byte) State {
		if itemChangeFlags != uint32(mainItem.ChangeFlags) || string(itemChangeData) != string(mainItem.ChangeData) ||
			itemIdFlags != uint32(mainItem.IdFlags) || string(itemIdData) != string(mainItem.IdData) {
			t.Fatalf("Main item data mismatch")
		}
		return StateUpToDate
	})

	if mainItemState != StateUpToDate || err != nil {
		t.Fatalf("Main item query failed: %v", err)
	}

}
