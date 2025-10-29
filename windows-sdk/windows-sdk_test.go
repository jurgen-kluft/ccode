package winsdk

import (
	"fmt"
	"testing"
)

func TestWindowsSDK(t *testing.T) {
	winSDK, err := FindWindowsSDK(WinAppDesktop)
	if err != nil || winSDK == nil {
		t.Fatalf("Failed to create Windows SDK: %v", err)
	}

	fmt.Printf("Windows SDK Directory: %s\n", winSDK.Dir)
	for _, version := range winSDK.Versions {
		fmt.Printf("Available Windows SDK Version: %s\n", version)
	}
}
