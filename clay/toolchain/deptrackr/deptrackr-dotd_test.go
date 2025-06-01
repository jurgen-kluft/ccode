package deptrackr

import (
	"testing"
)

func TestLoadDotdDepTrackr(t *testing.T) {
	buildDir := "test_build_dir"
	d := LoadDotdDepTrackr(buildDir)

	if d == nil {
		t.Fatal("Expected deptrackr to be loaded, but got nil")
	}

	err := d.AddItem("test_allocator.cpp.d", []StringItem{})
	if err != nil {
		t.Fatalf("Expected to add item without error, but got: %v", err)
	}
}
