package deptrackr

import (
	"path/filepath"
	"testing"
)

func TestLoadDotdDepTrackr(t *testing.T) {
	buildDir := "test_build_dir"
	d := LoadDepFileTrackr(filepath.Join(buildDir, "deptrackr.test"))

	if d == nil {
		t.Fatal("Expected deptrackr to be loaded, but got nil")
	}

	srcFilepath := "test_allocator.cpp"
	objFilepath := "test_allocator.obj"
	dotdFilepath := "test_allocator.cpp.d"
	mainItem, depItems, err := d.ParseDependencyFile(srcFilepath, objFilepath, dotdFilepath)
	if err != nil {
		t.Fatalf("Failed to parse .d file '%s', with error: %v", dotdFilepath, err)
	}
	err = d.AddItem(mainItem, depItems)
	if err != nil {
		t.Fatalf("Expected to add item without error, but got: %v", err)
	}
}
