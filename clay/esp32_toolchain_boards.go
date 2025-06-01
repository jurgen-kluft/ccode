package clay

import (
	"fmt"
	"sort"

	utils "github.com/jurgen-kluft/ccode/utils"
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
func BoardsOperation(espSdkPath string, cpuName string, boardName string, opName string) (result []string, err error) {

	// var flashDefMatch string
	// if cpuName == "esp32" {
	// 	flashDefMatch = `\.build\.flash_size=(\S+)`
	// } else {
	// 	flashDefMatch = `\.menu\.(?:FlashSize|eesz)\.([^\.]+)=(.+)`
	// }

	// lwipDefMatch := `\.menu\.(?:LwIPVariant|ip)\.(\w+)=(.+)`

	toolchain, err := ParseEsp32Toolchain(espSdkPath)
	if err != nil {
		return []string{}, err
	}

	if opName == "first" {
		result := toolchain.Boards[0].Name
		result = result[0:1]
	} else if opName == "check" {
		if i, ok := toolchain.NameToBoard[boardName]; ok {
			result = []string{toolchain.Boards[i].Name}
		} else {
			return []string{}, fmt.Errorf("Board %s not found in toolchain\n", boardName)
		}
	} else if opName == "first_flash" {
		if i, ok := toolchain.NameToBoard[boardName]; ok {
			for flashKey, flashValue := range toolchain.Boards[i].FlashSizes {
				result = append(result, fmt.Sprintf("%s %s", flashKey, flashValue))
				break
			}
		} else {
			return []string{}, fmt.Errorf("Board %s not found in toolchain\n", boardName)
		}
	} else if opName == "first_lwip" {
		// lwip configurations for board (a manual search shows there are none)
	} else if opName == "list_boards" {
		for i := 0; i < len(toolchain.Boards); i++ {
			result = append(result, fmt.Sprintf("%-20s %s", toolchain.Boards[i].Name, toolchain.Boards[i].Description))
		}
	} else if opName == "list_flash" {
		// memory configurations for board
	} else if opName == "list_lwip" {
		// lwip configurations for board (a manual search shows there are none)
	}

	return result, nil
}

func PrintAllFlashSizes(espSdkPath string, cpuName string, boardName string) (err error) {

	// var flashDefMatch string
	// if cpuName == "esp32" {
	// 	flashDefMatch = `\.build\.flash_size=(\S+)`
	// } else {
	// 	flashDefMatch = `\.menu\.(?:FlashSize|eesz)\.([^\.]+)=(.+)`
	// }

	toolchain, err := ParseEsp32Toolchain(espSdkPath)
	if err != nil {
		return err
	}

	// Get the board
	var board *Esp32Board
	if i, ok := toolchain.NameToBoard[boardName]; ok {
		board = toolchain.Boards[i]

		column1 := make([]string, 0)
		column2 := make([]string, 0)

		// Get the flash sizes
		for flashKey, _ := range board.FlashSizes {
			column1 = append(column1, flashKey)
		}

		// Sort the keys, column1
		sort.Strings(column1)

		// Now get the values
		for _, flashKey := range column1 {
			flashValue := board.FlashSizes[flashKey]
			column2 = append(column2, flashValue)
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
	}
	return nil
}

type Board struct {
	Name        string
	Description string
}

func PrintAllMatchingBoards(espSdkPath string, fuzzy string, max int) error {

	toolchain, err := ParseEsp32Toolchain(espSdkPath)
	if err != nil {
		return err
	}

	// First search in the board names
	names := make([]string, 0, len(toolchain.Boards))
	for _, board := range toolchain.Boards {
		names = append(names, board.Name)
	}

	cm := utils.NewClosestMatch(names, []int{2})
	closest := cm.ClosestN(fuzzy, max)
	if len(closest) > 0 {

		// Create map of board name to board description
		boardMap := make(map[string]string)
		for _, board := range toolchain.Boards {
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
		descriptions := make([]string, 0, len(toolchain.Boards))
		for _, board := range toolchain.Boards {
			descriptions = append(descriptions, board.Description)
		}
		cm = utils.NewClosestMatch(descriptions, []int{2})
		closest = cm.ClosestN(fuzzy, max-len(closest))
		if len(closest) > 0 {

			// Create map of board name to board description
			boardMap := make(map[string]string)
			for _, board := range toolchain.Boards {
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
