package clay

import (
	"bufio"
	"fmt"
	"os"
	"sort"

	corepkg "github.com/jurgen-kluft/ccode/core"
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
func BoardsOperation(espressifToolchain *EspressifToolchain, cpuName string, boardName string, opName string) (result []string, err error) {

	// var flashDefMatch string
	// if cpuName == "esp32" {
	// 	flashDefMatch = `\.build\.flash_size=(\S+)`
	// } else {
	// 	flashDefMatch = `\.menu\.(?:FlashSize|eesz)\.([^\.]+)=(.+)`
	// }

	// lwipDefMatch := `\.menu\.(?:LwIPVariant|ip)\.(\w+)=(.+)`

	if opName == "first" {
		result := espressifToolchain.ListOfBoards[0].Name
		result = result[0:1]
	} else if opName == "check" {
		if i, ok := espressifToolchain.BoardNameToIndex[boardName]; ok {
			result = []string{espressifToolchain.ListOfBoards[i].Name}
		} else {
			return []string{}, fmt.Errorf("Board %s not found in toolchain\n", boardName)
		}
	} else if opName == "first_flash" {
		if _, ok := espressifToolchain.BoardNameToIndex[boardName]; ok {
			// TODO, iterate over the menu entries to find the first flash size
		} else {
			return []string{}, fmt.Errorf("Board %s not found in toolchain\n", boardName)
		}
	} else if opName == "first_lwip" {
		// lwip configurations for board (a manual search shows there are none)
	} else if opName == "list_boards" {
		for i := 0; i < len(espressifToolchain.ListOfBoards); i++ {
			result = append(result, fmt.Sprintf("%-20s %s", espressifToolchain.ListOfBoards[i].Name, espressifToolchain.ListOfBoards[i].Description))
		}
	} else if opName == "list_flash" {
		// memory configurations for board
	} else if opName == "list_lwip" {
		// lwip configurations for board (a manual search shows there are none)
	}

	return result, nil
}

func PrintAllFlashSizes(espressifToolchain *EspressifToolchain, arch string, boardName string) (err error) {

	// var flashDefMatch string
	// if cpuName == "esp32" {
	// 	flashDefMatch = `\.build\.flash_size=(\S+)`
	// } else {
	// 	flashDefMatch = `\.menu\.(?:FlashSize|eesz)\.([^\.]+)=(.+)`
	// }

	// Get the parsed board
	//var board *EspressifBoard
	if _, ok := espressifToolchain.BoardNameToIndex[boardName]; ok {
		//board = espressifToolchain.ListOfBoards[i]

		column1 := make([]string, 0)
		column2 := make([]string, 0)

		// Get the flash sizes
		// Iterate over the board menu entries to find the flash sizes

		// Sort the keys, column1
		sort.Strings(column1)

		// Now get the values
		for _, flashKey := range column1 {
			flashValue := flashKey
			column2 = append(column2, flashValue)
		}

		column1MaxLength := len("Flash Size")
		for _, val := range column1 {
			if len(val) > column1MaxLength {
				column1MaxLength = len(val)
			}
		}

		// Print the header
		corepkg.LogInfof("%-*s   %s\n", column1MaxLength, "----------", "-----------")
		corepkg.LogInfof("%-*s | %s\n", column1MaxLength, "Flash Size", "Description")
		for i := 0; i < len(column1); i++ {
			corepkg.LogInfof("%-*s | %s\n", column1MaxLength, column1[i], column2[i])
		}
	}
	return nil
}

func PrintAllBoardInfos(espressifToolchain *EspressifToolchain, boardName string, max int) error {

	// Print some info
	espressifToolchain.PrintInfo()

	// First search in the board names
	names := make([]string, 0, len(espressifToolchain.ListOfBoards))
	for _, board := range espressifToolchain.ListOfBoards {
		names = append(names, board.Name)
	}

	cm := corepkg.NewClosestMatch(names, []int{2})
	closest := cm.ClosestN(boardName, max)
	if len(closest) > 0 {
		for _, match := range closest {
			if board := espressifToolchain.GetBoardByName(match); board != nil {
				corepkg.LogInfo("----------------------- " + board.Name + " -----------------------")
				corepkg.LogInfof("Board: %s", board.Name)
				corepkg.LogInfof("Description: %s", board.Description)
				corepkg.LogInfo(board.Vars.String())
				corepkg.LogInfo()
			}
		}
	}
	return nil
}

func GenerateToolchainJson(espressifToolchain *EspressifToolchain, outputFilename string) error {
	file, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonEncoder := corepkg.NewJsonEncoder("    ")
	jsonEncoder.HintOutputSize(4 * 1024 * 1024)
	jsonEncoder.Begin()
	encodeJsonEspressifToolchain(jsonEncoder, "", espressifToolchain)
	json := jsonEncoder.End()

	file.WriteString(json)
	return nil
}

func LoadToolchainJson(espressifToolchain *EspressifToolchain, inputFilename string) error {
	file, err := os.Open(inputFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	size := stat.Size()
	buffer := make([]byte, size)

	reader := bufio.NewReader(file)
	_, err = reader.Read(buffer)
	if err != nil {
		return err
	}

	jsonDecoder := corepkg.NewJsonDecoder()
	jsonDecoder.Begin(string(buffer))
	if err := decodeJsonEspressifToolchain(espressifToolchain, jsonDecoder); err != nil {
		return err
	}

	return nil
}

func PrintAllMatchingBoards(espressifToolchain *EspressifToolchain, fuzzy string, max int) error {

	// First search in the board names
	names := make([]string, 0, len(espressifToolchain.ListOfBoards))
	for _, board := range espressifToolchain.ListOfBoards {
		names = append(names, board.Name)
	}

	cm := corepkg.NewClosestMatch(names, []int{2})
	closest := cm.ClosestN(fuzzy, max)
	if len(closest) > 0 {

		// Create map of board name to board description
		boardMap := make(map[string]string)
		for _, board := range espressifToolchain.ListOfBoards {
			boardMap[board.Name] = board.Description
		}

		longestName := 0
		for _, match := range closest {
			if len(match) > longestName {
				longestName = len(match)
			}
		}
		for _, match := range closest {
			corepkg.LogInfof("%-*s %s", longestName, match, boardMap[match])
		}
	}

	if len(closest) < max {

		// Now search in the board descriptions
		descriptions := make([]string, 0, len(espressifToolchain.ListOfBoards))
		for _, board := range espressifToolchain.ListOfBoards {
			descriptions = append(descriptions, board.Description)
		}
		cm = corepkg.NewClosestMatch(descriptions, []int{2})
		closest = cm.ClosestN(fuzzy, max-len(closest))
		if len(closest) > 0 {

			// Create map of board name to board description
			boardMap := make(map[string]string)
			for _, board := range espressifToolchain.ListOfBoards {
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
				corepkg.LogInfof("%-*s %s", longestName, boardName, match)
			}

		}
	}

	return nil
}
