package clay

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jurgen-kluft/ccode/clay/toolchain"
	"github.com/jurgen-kluft/ccode/foundation"
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
func BoardsOperation(esp32Toolchain *Esp32Toolchain, cpuName string, boardName string, opName string) (result []string, err error) {

	// var flashDefMatch string
	// if cpuName == "esp32" {
	// 	flashDefMatch = `\.build\.flash_size=(\S+)`
	// } else {
	// 	flashDefMatch = `\.menu\.(?:FlashSize|eesz)\.([^\.]+)=(.+)`
	// }

	// lwipDefMatch := `\.menu\.(?:LwIPVariant|ip)\.(\w+)=(.+)`

	if opName == "first" {
		result := esp32Toolchain.ListOfBoards[0].Name
		result = result[0:1]
	} else if opName == "check" {
		if i, ok := esp32Toolchain.NameToIndex[boardName]; ok {
			result = []string{esp32Toolchain.ListOfBoards[i].Name}
		} else {
			return []string{}, fmt.Errorf("Board %s not found in toolchain\n", boardName)
		}
	} else if opName == "first_flash" {
		if i, ok := esp32Toolchain.NameToIndex[boardName]; ok {
			for flashKey, flashValue := range esp32Toolchain.ListOfBoards[i].FlashSizes {
				result = append(result, fmt.Sprintf("%s %s", flashKey, flashValue))
				break
			}
		} else {
			return []string{}, fmt.Errorf("Board %s not found in toolchain\n", boardName)
		}
	} else if opName == "first_lwip" {
		// lwip configurations for board (a manual search shows there are none)
	} else if opName == "list_boards" {
		for i := 0; i < len(esp32Toolchain.ListOfBoards); i++ {
			result = append(result, fmt.Sprintf("%-20s %s", esp32Toolchain.ListOfBoards[i].Name, esp32Toolchain.ListOfBoards[i].Description))
		}
	} else if opName == "list_flash" {
		// memory configurations for board
	} else if opName == "list_lwip" {
		// lwip configurations for board (a manual search shows there are none)
	}

	return result, nil
}

func PrintAllFlashSizes(esp32Toolchain *Esp32Toolchain, cpuName string, boardName string) (err error) {

	// var flashDefMatch string
	// if cpuName == "esp32" {
	// 	flashDefMatch = `\.build\.flash_size=(\S+)`
	// } else {
	// 	flashDefMatch = `\.menu\.(?:FlashSize|eesz)\.([^\.]+)=(.+)`
	// }

	// Get the parsed board
	var board *toolchain.Esp32Board
	if i, ok := esp32Toolchain.NameToIndex[boardName]; ok {
		board = esp32Toolchain.ListOfBoards[i]

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
		foundation.LogPrintf("%-*s   %s\n", column1MaxLength, "----------", "-----------")
		foundation.LogPrintf("%-*s | %s\n", column1MaxLength, "Flash Size", "Description")
		for i := 0; i < len(column1); i++ {
			foundation.LogPrintf("%-*s | %s\n", column1MaxLength, column1[i], column2[i])
		}
	}
	return nil
}

func PrintAllBoardInfos(esp32Toolchain *Esp32Toolchain, boardName string, max int) error {

	// Print some info
	esp32Toolchain.PrintInfo()

	// First search in the board names
	names := make([]string, 0, len(esp32Toolchain.ListOfBoards))
	for _, board := range esp32Toolchain.ListOfBoards {
		names = append(names, board.Name)
	}

	cm := foundation.NewClosestMatch(names, []int{2})
	closest := cm.ClosestN(boardName, max)
	if len(closest) > 0 {
		for _, match := range closest {
			if board := esp32Toolchain.GetBoardByName(match); board != nil {
				foundation.LogPrintf("----------------------- " + board.Name + " -----------------------\n")
				foundation.LogPrintf("Board: %s\n", board.Name)
				foundation.LogPrintf("Description: %s\n", board.Description)
				foundation.LogPrint(board.Vars.String())
				foundation.LogPrintf("\n")
			}
		}
	}
	return nil
}

func GenerateAllBoards(esp32Toolchain *Esp32Toolchain) error {
	// Print some info
	esp32Toolchain.PrintInfo()

	// Generate the boards.txt file (custom format)
	file, err := os.Create("boards.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	boardList := make([]string, 0, len(esp32Toolchain.ListOfBoards))
	for _, board := range esp32Toolchain.ListOfBoards {
		boardList = append(boardList, board.Name)
	}
	sort.Strings(boardList)

	sb := foundation.NewStringBuilder()
	for _, boardName := range boardList {
		board := esp32Toolchain.GetBoardByName(boardName)
		if board != nil {
			sb.WriteLn("----board----")
			sb.WriteLn(board.Name)
			sb.WriteLn(board.Description)
			sb.WriteLn("variables=[")
			board.Vars.SortByKey()
			for i, varKey := range board.Vars.Keys {
				varValue := board.Vars.Values[i]
				sb.WriteString(varKey)
				sb.WriteString(":=")
				sb.WriteLn(strings.Join(varValue, " "))
			}
			sb.WriteLn("]")
			sb.WriteLn("flash_sizes=[")
			for flashKey, flashValue := range board.FlashSizes {
				sb.WriteString(flashKey)
				sb.WriteString(":=")
				sb.WriteLn(flashValue)
			}
			sb.WriteLn("]")

			file.WriteString(sb.String())
			sb.Reset()
		}
	}

	return nil
}

func LoadBoards(esp32Toolchain *Esp32Toolchain) error {

	file, err := os.Open("boards.txt")
	if err != nil {
		return foundation.LogError(err, "Failed to open boards.txt")
	}
	defer file.Close()

	// Read the file content line by line
	var currentBoard *toolchain.Esp32Board
	var currentSection string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(currentSection) == 0 {
			switch line {
			case "variables=[":
				currentSection = line
			case "flash_sizes=[":
				currentSection = line
			case "----board----":
				scanner.Scan()
				name := scanner.Text()
				scanner.Scan()
				description := scanner.Text()
				if currentBoard != nil {
					esp32Toolchain.ListOfBoards = append(esp32Toolchain.ListOfBoards, currentBoard)
					esp32Toolchain.NameToIndex[strings.ToLower(currentBoard.Name)] = len(esp32Toolchain.ListOfBoards) - 1
				}
				currentBoard = toolchain.NewBoard(name, description)
				currentBoard.SdkPath = esp32Toolchain.SdkPath
				currentBoard.Vars.Set("arduino.version", esp32Toolchain.Version)
				currentSection = ""
			}
		} else {
			if line == "]" {
				currentSection = ""
			} else {
				switch currentSection {
				case "variables=[":
					op := strings.Index(line, ":=")
					if op > 0 {
						key := line[0:op]
						value := line[op+2:]
						currentBoard.Vars.Set(key, value)
					}
				case "flash_sizes=[":
					op := strings.Index(line, ":=")
					if op > 0 {
						key := line[0:op]
						value := line[op+2:]
						currentBoard.FlashSizes[key] = value
					}
				}
			}
		}
	}

	if currentBoard != nil {
		esp32Toolchain.ListOfBoards = append(esp32Toolchain.ListOfBoards, currentBoard)
		esp32Toolchain.NameToIndex[strings.ToLower(currentBoard.Name)] = len(esp32Toolchain.ListOfBoards) - 1
	}

	return nil
}

func PrintAllMatchingBoards(esp32Toolchain *Esp32Toolchain, fuzzy string, max int) error {

	// First search in the board names
	names := make([]string, 0, len(esp32Toolchain.ListOfBoards))
	for _, board := range esp32Toolchain.ListOfBoards {
		names = append(names, board.Name)
	}

	cm := foundation.NewClosestMatch(names, []int{2})
	closest := cm.ClosestN(fuzzy, max)
	if len(closest) > 0 {

		// Create map of board name to board description
		boardMap := make(map[string]string)
		for _, board := range esp32Toolchain.ListOfBoards {
			boardMap[board.Name] = board.Description
		}

		longestName := 0
		for _, match := range closest {
			if len(match) > longestName {
				longestName = len(match)
			}
		}
		for _, match := range closest {
			foundation.LogPrintf("%-*s %s\n", longestName, match, boardMap[match])
		}
	}

	if len(closest) < max {

		// Now search in the board descriptions
		descriptions := make([]string, 0, len(esp32Toolchain.ListOfBoards))
		for _, board := range esp32Toolchain.ListOfBoards {
			descriptions = append(descriptions, board.Description)
		}
		cm = foundation.NewClosestMatch(descriptions, []int{2})
		closest = cm.ClosestN(fuzzy, max-len(closest))
		if len(closest) > 0 {

			// Create map of board name to board description
			boardMap := make(map[string]string)
			for _, board := range esp32Toolchain.ListOfBoards {
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
				foundation.LogPrintf("%-*s %s\n", longestName, boardName, match)
			}

		}
	}

	return nil
}
