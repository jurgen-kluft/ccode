package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"

	axe "github.com/jurgen-kluft/ccode/axe"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usage: espmake_mem_use <flashSectionsPattern> <ramSectionsPattern>")
		return
	}

	flashSectionsPattern := os.Args[1]
	ramSectionsPattern := os.Args[2]

	// The input text is coming in from stdin, drain it first
	scanner := bufio.NewScanner(os.Stdin)
	var textToParse []string
	for {
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		textToParse = append(textToParse, line)
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	// Now we have the text to parse, let's call the function
	// and pass the flash and ram regex patterns
	err := axe.MemUse(flashSectionsPattern, ramSectionsPattern, textToParse)
	if err != nil {
		fmt.Println(err)
	}
}
