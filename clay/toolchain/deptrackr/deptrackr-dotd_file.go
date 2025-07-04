package deptrackr

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"hash"
	"os"
	"strings"
	"time"
)

type FileTrackr interface {
	QueryItem(item string) bool
	AddItem(item string, deps []string) error

	QueryItemWithExtraData(item string, data []byte) bool
	AddItemWithExtraData(item string, data []byte, deps []string) error

	ParseDependencyFile(filepath string) (mainItem string, depItems []string, err error)
	CopyItem(item string)

	Save() (int, error) // 0=none save (up-to-date), 1=changes saved (out-of-date), -1=error
}

// The depFileTracker is using the output of compilers like gcc,
// clang, etc. It uses the .d files generated by these compilers.
type depFileTracker struct {
	current      *trackr
	currentState State
	future       *trackr
	hasher       hash.Hash
}

func LoadDepFileTrackr(storageFilepath string) FileTrackr {
	current := loadTrackr(storageFilepath, "file deptrackr, v1.0.4")
	tracker := current.newTrackr()
	return &depFileTracker{
		current:      current,
		currentState: StateUpToDate, // Start with an up-to-date state
		future:       tracker,
		hasher:       sha1.New(),
	}
}

func (d *depFileTracker) Save() (int, error) {
	if d.currentState == StateUpToDate {
		// If the current state is up to date, we do not need to save anything
		return 0, nil
	}
	err := d.future.save()
	return 1, err
}

// In short form, a .d file (items are separated by '\'):
// <object-file>: \ <source-file> \ <header-file> \ ...

func (d *depFileTracker) CopyItem(item string) {
	d.hasher.Reset()
	d.hasher.Write([]byte(item))
	itemHash := d.hasher.Sum(nil)
	d.current.CopyItem(d.future, itemHash)
}

func (d *depFileTracker) ParseDependencyFile(filepath string) (mainItem string, depItems []string, err error) {
	contentBytes, err := os.ReadFile(filepath)
	if err != nil {
		return "", []string{}, err
	}

	type part struct {
		from int
		to   int
	}

	// Parse the .d file content
	content := string(contentBytes)
	var parts []part
	current := part{from: 0, to: 0} // Start with an empty part
	startPos := 0
	for startPos < len(content) {
		endPos := strings.Index(content[startPos:], "\\")
		if endPos == -1 {
			endPos = len(content)
		} else {
			endPos += startPos
		}

		// If we encounter a backslash, we assume the next part starts
		// Figure out the 'to' index of the current part by stepping back
		// ignoring any spaces, tabs, or newlines before the backslash
		end := endPos - 1
		for end >= 0 && (content[end] == ' ' || content[end] == '\t' || content[end] == '\n' || content[end] == '\r' || content[end] == ':') {
			end-- // move back until we find a non-(space,tab,cr,ln,:) character
		}
		current.to = end + 1 // set the 'to' index of the last part
		parts = append(parts, current)

		// Now we need to find the beginning of the next part, but first
		// skip space, tab, and newline characters after a backslash
		begin := endPos + 1
		for begin < len(content) && (content[begin] == ' ' || content[begin] == '\t' || content[begin] == '\n' || content[begin] == '\r') {
			begin++ // move forward until we find a non-(space,tab,cr,ln) character
		}
		current = part{from: begin, to: begin} // Start a new part

		startPos = begin // Set the index to continue scanning
	}

	for i, p := range parts {
		if i == 0 {
			mainItem = content[p.from:p.to]
		} else {
			depItem := content[p.from:p.to]
			depItems = append(depItems, depItem)
		}
	}

	return mainItem, depItems, nil
}

// item = depfileAbsFilepath
func (d *depFileTracker) addItem(item string, extra []byte, deps []string) error {

	// ----------------------------------------------------------------
	d.hasher.Reset()
	d.hasher.Write([]byte(item))
	mainHash := d.hasher.Sum(nil)

	// For the 'change', we want the file modification time and hash it
	modTimeBytes := make([]byte, 8)
	fileInfo, err := os.Stat(item)
	if err != nil {
		binary.LittleEndian.PutUint64(modTimeBytes, uint64(time.Now().Unix())) // If the file does not exist, use the current time
	} else {
		binary.LittleEndian.PutUint64(modTimeBytes, uint64(fileInfo.ModTime().Unix()))
	}

	// We are adding a new item, so the trackr is marked as out of date
	d.currentState = StateOutOfDate

	itemToAdd := ItemToAdd{
		IdData:       []byte(item),
		IdDigest:     mainHash,
		IdFlags:      ItemFlagSourceFile,
		ChangeData:   bytes.Clone(modTimeBytes),
		ChangeDigest: nil, // mod-time is small enough, we do not need a hash
		ChangeFlags:  ChangeFlagModTime,
	}

	var depItems []ItemToAdd
	for _, depFilepath := range deps {
		// Make sure the digest of a dependency will be unique and
		// not identical when the same file is used as a main item
		d.hasher.Reset()
		d.hasher.Write([]byte{'d', 'e', 'p'})
		d.hasher.Write([]byte(depFilepath))
		depDigest := d.hasher.Sum(nil)

		// For the 'change', we want the file modification time and hash it
		fileInfo, err = os.Stat(depFilepath)
		if err != nil {
			binary.LittleEndian.PutUint64(modTimeBytes, uint64(time.Now().Unix()))
		} else {
			binary.LittleEndian.PutUint64(modTimeBytes, uint64(fileInfo.ModTime().Unix()))
		}

		depItemToAdd := ItemToAdd{
			IdDigest:     depDigest,
			IdData:       []byte(depFilepath),
			IdFlags:      ItemFlagDependency,
			ChangeDigest: nil, // mod-time is small enough, we do not need a hash
			ChangeData:   bytes.Clone(modTimeBytes),
			ChangeFlags:  ChangeFlagModTime,
		}
		depItems = append(depItems, depItemToAdd)
	}
	if len(extra) == 0 {
		d.future.AddItem(itemToAdd, depItems)
	} else {
		d.future.AddItemWithExtraData(itemToAdd, extra, depItems)
	}
	return nil
}

func (d *depFileTracker) AddItem(item string, deps []string) error {
	return d.addItem(item, nil, deps)
}

// QueryFile checks if the files are up to date, returning true if it is,
// or false if it is out of date or does not exist in the current state.
func (d *depFileTracker) QueryItem(item string) bool {
	d.hasher.Reset()
	d.hasher.Write([]byte(item))
	mainDigest := d.hasher.Sum(nil)

	modTimeBytes := make([]byte, 8) // 8 bytes for the file modification time

	state, err := d.current.QueryItem(mainDigest, true, func(itemState State, itemIdFlags uint8, itemIdData []byte, itemChangeFlags uint8, itemChangeData []byte) State {
		if itemIdFlags&ItemFlagSourceFile == ItemFlagSourceFile || itemIdFlags&ItemFlagDependency == ItemFlagDependency {
			// Items that have been gone through a query have been updated with their current state which
			// was either up to date or out of date. This avoids asking the filesystem for the state of the file
			// if it is already known to be up to date or out of date.
			// This is mainly relevant for dependency files, which can be shared between multiple main items.
			if itemState == StateNone {
				srcFileInfo, err := os.Stat(string(itemIdData))
				if err == nil {
					binary.LittleEndian.PutUint64(modTimeBytes, uint64(srcFileInfo.ModTime().Unix()))
					if bytes.Compare(modTimeBytes, itemChangeData) == 0 {
						return StateUpToDate
					}
				}
				return StateOutOfDate
			}
			return itemState
		} else {
			// TODO, handle other types of items (StringItem)
			return StateUpToDate
		}
	})

	if err != nil {
		fmt.Println("Error querying item:", err)
	}

	return state == StateUpToDate
}

func (d *depFileTracker) AddItemWithExtraData(item string, data []byte, deps []string) error {
	return d.addItem(item, data, deps)
}

func (d *depFileTracker) QueryItemWithExtraData(item string, data []byte) bool {
	d.hasher.Reset()
	d.hasher.Write([]byte(item))
	mainDigest := d.hasher.Sum(nil)

	modTimeBytes := make([]byte, 8) // 8 bytes for the file modification time

	state, err := d.current.QueryItemExtra(mainDigest, true, func(itemState State, itemIdFlags uint8, itemIdData []byte, itemExtraData []byte, itemChangeFlags uint8, itemChangeData []byte) State {
		if itemIdFlags&ItemFlagSourceFile == ItemFlagSourceFile || itemIdFlags&ItemFlagDependency == ItemFlagDependency {
			// Items that have been gone through a query have been updated with their current state which
			// was either up to date or out of date. This avoids asking the filesystem for the state of the file
			// if it is already known to be up to date or out of date.
			// This is mainly relevant for dependency files, which can be shared between multiple main items.
			if itemState == StateNone {
				// Check if the itemExtraData matches the item extra data we are querying
				// Note: dependency items do not have extra data (nil or zero size)
				if len(itemExtraData) == 0 || bytes.Equal(itemExtraData, data) {
					srcFileInfo, err := os.Stat(string(itemIdData))
					if err == nil {
						binary.LittleEndian.PutUint64(modTimeBytes, uint64(srcFileInfo.ModTime().Unix()))
						if bytes.Compare(modTimeBytes, itemChangeData) == 0 {
							return StateUpToDate
						}
					}
				}
				return StateOutOfDate
			}
			return itemState
		} else {
			// TODO, handle other types of items (StringItem)
			return StateUpToDate
		}
	})

	if err != nil {
		fmt.Println("Error querying item:", err)
	}

	return state == StateUpToDate
}
