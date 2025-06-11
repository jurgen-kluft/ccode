package msvc

import (
	"fmt"
	"testing"
)

func TestWindowsSDK(t *testing.T) {
	sdkDir, sdkVersion, err := findWindowsSDK("", Desktop)
	if err != nil {
		t.Fatalf("Failed to create Windows SDK: %v", err)
	}

	fmt.Printf("Windows SDK Directory: %s\n", sdkDir)
	fmt.Printf("Windows SDK Version: %s\n", sdkVersion)
}
