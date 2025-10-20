package cespressif

import (
	"bufio"
	"os"
	"sort"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

func (t *toolchain) loadJson(inputFilename string) error {
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
	if err := decodeJsonToolchain(t, jsonDecoder); err != nil {
		return err
	}

	return nil
}

func (t *toolchain) saveJson(outputFilename string) error {
	file, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonEncoder := corepkg.NewJsonEncoder("    ")
	jsonEncoder.HintOutputSize(4 * 1024 * 1024)
	jsonEncoder.Begin()
	encodeJsonToolchain(jsonEncoder, "", t)
	json := jsonEncoder.End()

	file.WriteString(json)
	return nil
}

func PrintAllFlashSizes(toolchain *toolchain, arch string, boardName string) (err error) {

	// var flashDefMatch string
	// if cpuName == "esp32" {
	// 	flashDefMatch = `\.build\.flash_size=(\S+)`
	// } else {
	// 	flashDefMatch = `\.menu\.(?:FlashSize|eesz)\.([^\.]+)=(.+)`
	// }

	if i, ok := toolchain.BoardNameToIndex[boardName]; ok {
		board := toolchain.ListOfBoards[i]

		column1 := make([]string, 0)
		column2 := make([]string, 0)

		// Get the flash sizes
		// Iterate over the board menu entries to find the flash sizes
		for _, e := range board.Menu.Entries {
			if e.Name == "eesz" {
				for _, se := range e.SubEntries {
					column2 = append(column2, se.Title)
					for pi, pk := range se.Keys {
						if pk == "build.flash_size" {
							column1 = append(column1, se.Values[pi])
							break
						}
					}
				}
			}
		}

		// Sort the keys, column1
		sort.Strings(column1)

		column1MaxLength := len("Flash Size")
		for _, val := range column1 {
			column1MaxLength = max(column1MaxLength, len(val))
		}

		// Print the header
		corepkg.LogInfof("%-*s   %s", column1MaxLength, "----------", "-----------")
		corepkg.LogInfof("%-*s | %s", column1MaxLength, "Flash Size", "Description")
		for i := 0; i < len(column1); i++ {
			corepkg.LogInfof("%-*s | %s", column1MaxLength, column1[i], column2[i])
		}
	}
	return nil
}

func PrintAllBoardInfos(toolchain *toolchain, boardName string, max int) error {

	// Print some info
	toolchain.PrintInfo()

	// First search in the board names
	names := make([]string, 0, len(toolchain.ListOfBoards))
	for _, board := range toolchain.ListOfBoards {
		names = append(names, board.Name)
	}

	cm := corepkg.NewClosestMatch(names, []int{2})
	closest := cm.ClosestN(boardName, max)
	if len(closest) > 0 {
		for _, match := range closest {
			if board := toolchain.GetBoardByName(match); board != nil {
				vars := corepkg.NewVars(corepkg.VarsFormatCurlyBraces)
				toolchain.ResolveVariablesForBoard(board, vars)
				corepkg.LogInfo("----------------------- " + board.Name + " -----------------------")
				corepkg.LogInfof("Board: %s", board.Name)
				corepkg.LogInfof("Description: %s", board.Description)
				for _, key := range vars.Keys {
					key = strings.ToLower(key)
					if strings.HasPrefix(key, "build.") || strings.HasPrefix(key, "upload.") {
						values := vars.Values[vars.KeyToIndex22[key]]
						corepkg.LogInfof("%s:%s", key, values)
					}
				}

				corepkg.LogInfo()
			}
		}
	}
	return nil
}

func PrintAllMatchingBoards(toolchain *toolchain, fuzzy string, listMax int) error {

	// First search in the board names
	names := make([]string, 0, len(toolchain.ListOfBoards))
	for _, board := range toolchain.ListOfBoards {
		names = append(names, board.Name)
	}

	boardNameMaxLen := 0
	listedBoards := make(map[string]bool)

	cm := corepkg.NewClosestMatch(names, []int{2})
	closest := cm.ClosestN(fuzzy, listMax)
	if len(closest) > 0 {

		// Create map of board name to board description
		boardMap := make(map[string]string)
		for _, board := range toolchain.ListOfBoards {
			boardMap[board.Name] = board.Description
		}

		for _, match := range closest {
			boardNameMaxLen = max(boardNameMaxLen, len(match)+8)
		}
		for _, match := range closest {
			listedBoards[match] = true
			corepkg.LogInfof("    %-*s %s", boardNameMaxLen, match, boardMap[match])
		}
	}

	if len(closest) < listMax {

		// Now search in the board descriptions
		descriptions := make([]string, 0, len(toolchain.ListOfBoards))
		for _, board := range toolchain.ListOfBoards {
			descriptions = append(descriptions, board.Description)
		}

		cm = corepkg.NewClosestMatch(descriptions, []int{2})
		closest = cm.ClosestN(fuzzy, listMax-len(closest))
		if len(closest) > 0 {

			// Create map of board description to board name
			boardMap := make(map[string]string)
			for _, board := range toolchain.ListOfBoards {
				boardMap[board.Description] = board.Name
			}

			for _, match := range closest {
				boardName := boardMap[match]
				boardNameMaxLen = max(boardNameMaxLen, len(boardName)+8)
			}
			for _, match := range closest {
				boardName := boardMap[match]
				if _, ok := listedBoards[boardName]; !ok {
					listedBoards[boardName] = true
					corepkg.LogInfof("    %-*s %s", boardNameMaxLen, boardName, match)
				}
			}
		}
	}

	return nil
}
