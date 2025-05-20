package clay

// Parse the 'boards.txt' file to extract board basic information.
// Then 'platform.txt', contains information for building
// - compiler, archive, linker
// - image, partitions, bootloader
// - flashing

// esp32:
// -esp32
// -s2
// -s3
// -c3

// compiler
// archive
// linker

// flash size
// flash frequency
// flash mode

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	cutils "github.com/jurgen-kluft/ccode/cutils"
)

// opName: Any of the following:
// - first
// - check
// - first_flash
// - first_lwip
// - list_names
// - list_flash
// - list_lwip

// BoardsOperation performs a search operation on the board configuration file and returns nil if successful.
func BoardsOperation(filePath string, cpuName string, boardName string, opName string) (result []string, err error) {

	var flashDefMatch string
	if cpuName == "esp32" {
		flashDefMatch = `\.build\.flash_size=(\S+)`
	} else {
		flashDefMatch = `\.menu\.(?:FlashSize|eesz)\.([^\.]+)=(.+)`
	}

	lwipDefMatch := `\.menu\.(?:LwIPVariant|ip)\.(\w+)=(.+)`

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return []string{}, fmt.Errorf("Failed to open: %s\n", filePath)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	if opName == "first" {
		// take the first occurrence of a board name
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, boardName+".name=") {
				result = strings.SplitN(line, "=", 2)
				result = result[0:1]
				break
			}
		}
	} else if opName == "check" {
		boardNameDotName := boardName + ".name"
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, boardNameDotName) {
				result = []string{"1"}
				break
			}
		}

	} else if opName == "first_flash" {
		re := regexp.MustCompile(regexp.QuoteMeta(boardName) + flashDefMatch)
		for scanner.Scan() {
			line := scanner.Text()
			match := re.FindStringSubmatch(line)
			if len(match) > 1 {
				result = []string{match[1]}
				break
			}
		}
	} else if opName == "first_lwip" {
		re := regexp.MustCompile(regexp.QuoteMeta(boardName) + lwipDefMatch)
		for scanner.Scan() {
			line := scanner.Text()
			match := re.FindStringSubmatch(line)
			if len(match) > 1 {
				result = []string{match[1]}
				break
			}
		}
	} else if opName == "list_boards" {
		re := regexp.MustCompile(`^([\w\-]+)\.name=(.+)`)
		for scanner.Scan() {
			line := scanner.Text()
			match := re.FindStringSubmatch(line)
			if len(match) > 2 {
				result = append(result, fmt.Sprintf("%-20s %s", match[1], match[2]))
			}
		}
	} else if opName == "list_flash" {
		// memory configurations for board
		re := regexp.MustCompile(regexp.QuoteMeta(boardName) + flashDefMatch)
		for scanner.Scan() {
			line := scanner.Text()
			match := re.FindStringSubmatch(line)
			if len(match) > 0 {
				val1 := ""
				if len(match) > 1 {
					val1 = match[1]
				}
				val2 := ""
				if len(match) > 2 {
					val2 = match[2]
				}
				//fmt.Printf("%-10s %s\n", val1, val2)
				result = append(result, fmt.Sprintf("%-10s %s", val1, val2))
			}
		}
	} else if opName == "list_lwip" {
		// lwip configurations for board
		re := regexp.MustCompile(regexp.QuoteMeta(boardName) + lwipDefMatch)
		for scanner.Scan() {
			line := scanner.Text()
			match := re.FindStringSubmatch(line)
			if len(match) > 2 { // Expects 2 capture groups for printing
				result = append(result, fmt.Sprintf("%-10s %s", match[1], match[2]))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return []string{}, fmt.Errorf("Error reading file: %s\n", err)
	}

	return result, nil
}

func PrintAllFlashSizes(filePath string, cpuName string, boardName string) (err error) {

	var flashDefMatch string
	if cpuName == "esp32" {
		flashDefMatch = `\.build\.flash_size=(\S+)`
	} else {
		flashDefMatch = `\.menu\.(?:FlashSize|eesz)\.([^\.]+)=(.+)`
	}

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("Failed to open: %s\n", filePath)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	column1 := make([]string, 0)
	column2 := make([]string, 0)

	re := regexp.MustCompile(regexp.QuoteMeta(boardName) + flashDefMatch)
	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)
		if len(match) > 0 {
			val1 := ""
			if len(match) > 1 {
				val1 = match[1]
			}
			val2 := ""
			if len(match) > 2 {
				val2 = match[2]
			}
			column1 = append(column1, val1)
			column2 = append(column2, val2)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error reading file: %s\n", err)
	}

	column1MaxLength := len("Flash Size")
	for _, val := range column1 {
		if len(val) > column1MaxLength {
			column1MaxLength = len(val)
		}
	}

	// Print the header
	fmt.Printf("%-*s   %s\n", column1MaxLength, "----------", "-----------")
	fmt.Printf("%-*s | %s\n", column1MaxLength, "Flash Size", "Description")
	for i := 0; i < len(column1); i++ {
		fmt.Printf("%-*s | %s\n", column1MaxLength, column1[i], column2[i])
	}

	return nil
}

type Board struct {
	Name        string
	Description string
}

func GenerateListOfAllBoards(filePath string) (result []Board, err error) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return []Board{}, fmt.Errorf("Failed to open: %s\n", filePath)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	re := regexp.MustCompile(`^([\w\-]+)\.name=(.+)`)
	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)
		if len(match) > 2 {
			result = append(result, Board{Name: match[1], Description: match[2]})
		}
	}

	if err := scanner.Err(); err != nil {
		return []Board{}, fmt.Errorf("Error reading file: %s\n", err)
	}

	return result, nil
}

func PrintAllMatchingBoards(boardsFilePath string, fuzzy string, max int) error {
	boards, err := GenerateListOfAllBoards(boardsFilePath)
	if err != nil {
		return fmt.Errorf("Failed to list boards: %v", err)
	}

	// First search in the board names
	names := make([]string, 0, len(boards))
	for _, board := range boards {
		names = append(names, board.Name)
	}

	cm := cutils.NewClosestMatch(names, []int{2})
	closest := cm.ClosestN(fuzzy, max)
	if len(closest) > 0 {

		// Create map of board name to board description
		boardMap := make(map[string]string)
		for _, board := range boards {
			boardMap[board.Name] = board.Description
		}

		longestName := 0
		for _, match := range closest {
			if len(match) > longestName {
				longestName = len(match)
			}
		}
		for _, match := range closest {
			fmt.Printf("%-*s %s\n", longestName, match, boardMap[match])
		}
	}

	if len(closest) < max {

		// Now search in the board descriptions
		descriptions := make([]string, 0, len(boards))
		for _, board := range boards {
			descriptions = append(descriptions, board.Description)
		}
		cm = cutils.NewClosestMatch(descriptions, []int{2})
		closest = cm.ClosestN(fuzzy, max-len(closest))
		if len(closest) > 0 {

			// Create map of board name to board description
			boardMap := make(map[string]string)
			for _, board := range boards {
				boardMap[board.Description] = board.Name
			}

			longestName := 0
			for _, match := range closest {
				boardName := boardMap[match]
				if len(boardName) > longestName {
					longestName = len(match)
				}
			}
			for _, match := range closest {
				boardName := boardMap[match]
				fmt.Printf("%-*s %s\n", longestName, boardName, match)
			}

		}
	}

	return nil
}
