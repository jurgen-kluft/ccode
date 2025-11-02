package main

import (
	"log"

	winsdk "github.com/jurgen-kluft/ccode/windows-sdk"
)

func main() {
	sdks, err := winsdk.FindWindowsSDKs(winsdk.WinAppDesktop)
	if err != nil || sdks == nil {
		log.Fatalf("Failed to analyze Windows for its available SDKs: %v", err)
	}

	for _, sdk := range sdks {
		sdk.Print()
	}
}
