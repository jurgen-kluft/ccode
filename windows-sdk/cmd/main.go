package main

import (
	"fmt"
	"log"

	winsdk "github.com/jurgen-kluft/ccode/windows-sdk"
)

func main() {
	win, err := winsdk.FindWindowsSDK(winsdk.WinAppDesktop)
	if err != nil || win == nil {
		log.Fatalf("Failed to analyze Windows for its available SDKs: %v", err)
	}

	// type WindowsSdkLayout struct {
	// 	bin      string
	// 	includes []string
	// 	libs     []string
	// }

	for _, winSDK := range win.Versions {
		fmt.Printf("Windows SDK Version: %s\n", winSDK.Version)
		fmt.Printf("Windows SDK Directory: %s\n", winSDK.Dir)

		fmt.Println("Bins:")
		for _, bin := range winSDK.Layout.Bin {
			fmt.Printf("  %s\n", bin)
		}

		fmt.Println("Includes:")
		for _, inc := range winSDK.Layout.Includes {
			fmt.Printf("  %s\n", inc)
		}

		fmt.Println("Libs:")
		for _, lib := range winSDK.Layout.Libs {
			fmt.Printf("  %s\n", lib)
		}

	}
}
