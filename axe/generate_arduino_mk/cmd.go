package main

import (
	_ "embed"
	"fmt"
	"os"

	axe "github.com/jurgen-kluft/ccode/axe"
)

func main() {
	if len(os.Args) < 6 {
		fmt.Println("Usage: axe-espmake_generate_arduino_mk <esp_root> <ard_root> <board_name> <flash_size> <os_type> <lwip_variant>")
		return
	}

	// We can collect the files that are collected in 'ARDUINO_DESC' by scanning the
	// 'ESP_ROOT' directory for files with the '.txt' extension.
	// In Go we can use the walk package to do this.

	fmt.Println("Generate Arduino makefile ...")

	espRoot := os.Args[1]
	boardName := os.Args[3]
	ardEspRoot := os.Args[2] + "/packages/" + boardName
	flashSize := []string{os.Args[4]}
	osType := os.Args[5]
	lwipVariant := []string{os.Args[6]}

	result, err := axe.GenerateArduinoMake(espRoot, ardEspRoot, boardName, flashSize, osType, lwipVariant)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, line := range result {
		fmt.Println(line)
	}
}
