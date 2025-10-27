package corepkg

import (
	"testing"
)

// Write some unittest for the UUIDRandomGenerator

func TestUUIDRandomGeneration(t *testing.T) {
	g := NewUUIDRandomGenerator(0xEED1A7370B428D0B)

	r0 := g.Uint64()
	if r0 == 0 {
		t.Errorf("The first random number should be non-zero")
	}

	r1 := g.Uint64()
	if r1 == 0 || r1 == r0 {
		t.Errorf("The second random number should be different from the first one and non-zero")
	}
}

func TestUUID1(t *testing.T) {
	uuid := GenerateUUID()
	str := uuid.ForXCode()
	if len(str) != 24 || str != "DEBED0FBE6AF2E6F30796D40" {
		t.Errorf("The UUID string should be 24 characters long and equal to DEBED0FBE6AF2E6F30796D40, instead len=%d and string=%s", len(str), str)
	}

}
func TestUUID2(t *testing.T) {
	uuid := GenerateUUID()
	str := uuid.ForVisualStudio()
	if len(str) != 38 || str != "{AA5C6F25-A3A0-0D83-9810-AC1ABA981740}" {
		t.Errorf("The UUID string should be 32 characters long and equal to {AA5C6F25-A3A0-0D83-9810-AC1ABA981740}, instead len=%d and string=%s", len(str), str)
	}
}

func TestUUIDGenerationForXcode(t *testing.T) {

	numUUIDs := 10000
	uuidMap := make(map[string]bool)
	for i := 0; i < numUUIDs; i++ {
		uuid := GenerateUUID()
		str := uuid.ForXCode()
		if _, ok := uuidMap[str]; ok {
			t.Errorf("The UUID string should be unique, but it was not: %s", str)
		} else {
			uuidMap[str] = true
		}
	}
}

func TestUUIDGenerationForMsdev(t *testing.T) {

	numUUIDs := 10000
	uuidMap := make(map[string]bool)
	for i := 0; i < numUUIDs; i++ {
		uuid := GenerateUUID()
		str := uuid.ForVisualStudio()
		if _, ok := uuidMap[str]; ok {
			t.Errorf("The UUID string should be unique, but it was not: %s", str)
		} else {
			uuidMap[str] = true
		}
	}
}
