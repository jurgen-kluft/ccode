package main

import (
	_ "embed"
	"fmt"
	"os"

	axe "github.com/jurgen-kluft/ccode/axe"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: espmake_board_op <file_path> <cpu_name> <board_name>")
		fmt.Println(`Example: espmake_board_op "/Users/user/sdk/arduino/esp32/boards.txt" "esp32" "" "check"`)
		return
	}

	// Command Line:
	// "/Users/obnosis5/sdk/arduino/esp32/boards.txt" "esp32" "esp32" "check"

	filePath := os.Args[1]
	cpuName := os.Args[2]
	boardName := os.Args[3]
	opName := os.Args[4]

	if boardName == "" {
		boardName = "esp32"
	}

	result, err := axe.BoardOp(filePath, cpuName, boardName, opName)
	if err != nil {
		fmt.Println(err)
	}

	for _, line := range result {
		fmt.Println(line)
	}
}
