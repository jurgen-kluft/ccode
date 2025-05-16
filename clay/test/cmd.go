package main

import "github.com/jurgen-kluft/ccode/clay"

// Test clay package

func main() {
	esp32 := clay.NewTargetEsp32("build")
	if esp32 == nil {
		panic("Failed to create ESP32 target")
	}

	if err := esp32.Init(); err != nil {
		panic("Failed to initialize ESP32 compiler package: " + err.Error())
	}
	if err := esp32.Prebuild(); err != nil {
		panic("Failed to prebuild ESP32 compiler package: " + err.Error())
	}

	if err := esp32.Build(); err != nil {
		panic("Failed to build ESP32 compiler package: " + err.Error())
	}
}
