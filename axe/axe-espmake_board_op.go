package axe

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// opName: Any of the following:
// - first
// - check
// - first_flash
// - first_lwip
// - list_names
// - list_flash
// - list_lwip

// BoardOp performs a search operation on the board configuration file and returns nil if successful.
func BoardOp(filePath string, cpuName string, boardName string, opName string) (result []string, err error) {

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
	} else if opName == "list_names" {
		re := regexp.MustCompile(`^([\w\-]+)\.name=(.+)`)
		for scanner.Scan() {
			line := scanner.Text()
			match := re.FindStringSubmatch(line)
			if len(match) > 2 {
				result = append(result, fmt.Sprintf("%-20s %s", match[1], match[2]))
			}
		}
	} else if opName == "list_flash" {
		//fmt.Printf("=== Memory configurations for board: %s ===\n", boardName)
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
		//fmt.Printf("=== lwip configurations for board: %s ===\n", boardName)
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
