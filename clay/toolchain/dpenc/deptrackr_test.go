package dpenc

import (
	"crypto/sha1"
	"path/filepath"
	"testing"
	"time"
)

func TestDepTrackrSimple(t *testing.T) {
	current := loadTrackr(filepath.Join("testdb", "deptrackr.simple"), "test deptrackr, v1.0.0")
	tracker := current.newTrackr()

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
	dependencyCount := 0
	mainItemState, err := tracker.QueryItem(mainItem.IdDigest, true, func(itemState State, itemIdFlags uint8, itemIdData []byte, itemChangeFlags uint8, itemChangeData []byte) State {
		if itemChangeFlags == mainItem.ChangeFlags && string(itemChangeData) == string(mainItem.ChangeData) &&
			itemIdFlags == mainItem.IdFlags && string(itemIdData) == string(mainItem.IdData) {
			return StateUpToDate
		}
		if itemChangeFlags == depItem.ChangeFlags && string(itemChangeData) == string(depItem.ChangeData) &&
			itemIdFlags == depItem.IdFlags && string(itemIdData) == string(depItem.IdData) {
			dependencyCount++
			return StateUpToDate
		}
		return StateOutOfDate
	})

	if dependencyCount != 1 {
		t.Fatalf("Expected 1 dependency, got %d", dependencyCount)
	}

	if mainItemState != StateUpToDate || err != nil {
		t.Fatalf("Main item query failed: %v", err)
	}
}

// Create a more elaborate test cast that includes multiple items where
// each item has more than 3 dependencies. Here we do not test for
// out-of-date items, but rather focus on the addition of multiple dependencies.
func TestDepTrackrMultipleDependencies(t *testing.T) {
	current := loadTrackr(filepath.Join("testdb", "deptrackr.multi"), "test deptrackr, v1.0.0")

	tracker := current.newTrackr()
	hasher := sha1.New()

	items := map[string]int{
		"test/main1.cpp": 1,
		"test/main2.cpp": 2,
		"test/main3.cpp": 3,
		"test/main4.cpp": 4,
		"test/main5.cpp": 5,
		"test/main6.cpp": 6,
	}

	dependencies := map[string]int{
		"test/dependency1.cpp": 100,
		"test/dependency2.cpp": 200,
		"test/dependency3.cpp": 300,
		"test/dependency4.cpp": 400,
		"test/dependency5.cpp": 500,
		"test/dependency6.cpp": 600,
	}

	for itemData, itemFlag := range items {
		hasher.Reset()
		hasher.Write([]byte(itemData))
		itemHash := hasher.Sum(nil)

		changeData := []byte(time.Now().String())
		hasher.Reset()
		hasher.Write(changeData)
		changeHash := hasher.Sum(nil)

		mainItem := ItemToAdd{
			IdDigest:     itemHash,
			IdData:       []byte(itemData),
			ChangeDigest: changeHash,
			ChangeData:   changeData,
			IdFlags:      uint8(itemFlag),
			ChangeFlags:  uint8(itemFlag),
		}

		var depItems []ItemToAdd
		for depData, depFlag := range dependencies {
			hasher.Reset()
			hasher.Write([]byte(depData))
			depItemHash := hasher.Sum(nil)

			depChangeData := []byte(time.Now().String())
			hasher.Reset()
			hasher.Write(depChangeData)
			depChangeHash := hasher.Sum(nil)

			depItem := ItemToAdd{
				IdDigest:     depItemHash,
				IdData:       []byte(depData),
				ChangeDigest: depChangeHash,
				ChangeData:   depChangeData,
				IdFlags:      uint8(depFlag),
				ChangeFlags:  uint8(depFlag),
			}
			depItems = append(depItems, depItem)
		}

		added := tracker.AddItem(mainItem, depItems)
		if !added {
			t.Fatalf("Failed to add item %s", itemData)
		}
	}

	// Query each main item to ensure all dependencies are added correctly
	for itemData, itemFlag := range items {
		hasher.Reset()
		hasher.Write([]byte(itemData))
		itemHash := hasher.Sum(nil)

		wrongState := false
		depCount := 0
		mainItemState, err := tracker.QueryItem(itemHash, true, func(itemState State, itemIdFlags uint8, itemIdData []byte, itemChangeFlags uint8, itemChangeData []byte) State {
			if itemState != StateNone {
				wrongState = true
			}
			if itemIdFlags == uint8(itemFlag) && string(itemIdData) == itemData {
				return StateUpToDate
			} else if itemIdFlags >= 100 {
				depCount++
				return StateUpToDate
			}
			return StateOutOfDate
		})

		if wrongState {
			t.Fatal("Expected state to be none")
		}

		if depCount != len(dependencies) {
			t.Fatalf("Expected %d dependencies, got %d", len(dependencies), depCount)
		}

		if mainItemState != StateUpToDate || err != nil {
			t.Fatalf("Main item %s query failed: %v", itemData, err)
		}
	}

}
