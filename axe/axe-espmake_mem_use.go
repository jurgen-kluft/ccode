package axe

import (
	_ "embed"
	"fmt"
	"regexp"
	"strconv"
)

// These patterns come from parsing the Arduino description files.
// MEM_FLASH -> flashSectionsPattern
// MEM_RAM -> ramSectionsPattern

func MemUse(flashSectionsPattern, ramSectionsPattern string, textToParse []string) (err error) {

	// Compile regexes
	flashRegex, err := regexp.Compile(flashSectionsPattern)
	if err != nil {
		return fmt.Errorf("Error compiling flash regex '%s': %v", flashSectionsPattern, err)
	}

	ramRegex, err := regexp.Compile(ramSectionsPattern)
	if err != nil {
		return fmt.Errorf("Error compiling ram regex '%s': %v", ramSectionsPattern, err)
	}

	var flashTot int64 = 0
	var ramTot int64 = 0

	for _, line := range textToParse {

		flashMatches := flashRegex.FindStringSubmatch(line)
		if len(flashMatches) > 1 { // Check if regex matched and captured at least one group
			// Attempt to convert the first captured group ($1) to an integer
			val, err := strconv.ParseInt(flashMatches[1], 10, 64)
			if err == nil {
				flashTot += val
			}
			// If err != nil, the captured group wasn't a number.
		}

		ramMatches := ramRegex.FindStringSubmatch(line)
		if len(ramMatches) > 1 { // Check if regex matched and captured at least one group
			// Attempt to convert the first captured group ($1) to an integer
			val, err := strconv.ParseInt(ramMatches[1], 10, 64)
			if err == nil {
				ramTot += val
			}
			// If err != nil, the captured group wasn't a number.
		}
	}

	fmt.Println("\nMemory summary")
	fmt.Printf("  %-6s %6d bytes\n  %-6s %6d bytes\n\n", "RAM:", ramTot, "Flash:", flashTot)

	return nil
}
