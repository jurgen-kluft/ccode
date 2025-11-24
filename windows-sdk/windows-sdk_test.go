package winsdk

import (
	"os"
	"testing"
)

func DoesDirExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func TestWindowsSDK(t *testing.T) {
	winSDKs, err := Find(WinAppDesktop)
	if err != nil || winSDKs == nil {
		t.Fatalf("Failed to create Windows SDK: %v", err)
	}

	for _, sdk := range winSDKs {
		if !DoesDirExist(sdk.Dir) {
			t.Fatalf("Windows SDK path does not exist: %s", sdk.Dir)
		}
		for _, incPath := range sdk.Layout.Includes {
			if !DoesDirExist(incPath) {
				t.Fatalf("Windows SDK include path does not exist: %s", incPath)
			}
		}
		for _, libPath := range sdk.Layout.Libs {
			if !DoesDirExist(libPath) {
				t.Fatalf("Windows SDK lib path does not exist: %s", libPath)
			}
		}
		sdk.Print()
	}
}
