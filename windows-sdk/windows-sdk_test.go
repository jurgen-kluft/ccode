package winsdk

import (
	"testing"
)

func TestWindowsSDK(t *testing.T) {
	winSDKs, err := Find(WinAppDesktop)
	if err != nil || winSDKs == nil {
		t.Fatalf("Failed to create Windows SDK: %v", err)
	}

	for _, sdk := range winSDKs {
		sdk.Print()
	}
}
