package axe

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func ObjectInfo(elfSizePath string, formType int, sortIndex int, objFiles []string) error {

	// Determine the output format string
	form := "%-38.38s %7s %7s %7s %7s %7s\n"
	if formType == 1 {
		form = "%s\t%s\t%s\t%s\t%s\t%s\n"
	}

	// Validate sort index range
	if sortIndex < 0 || sortIndex > 4 {
		return fmt.Errorf("sort index must be between 0 and 4 (inclusive)")
	}

	fmt.Printf(form, "File", "Flash", "RAM", "data", "rodata", "bss")

	if !strings.Contains(form, "\t") {
		fmt.Println(strings.Repeat("-", 78))
	}

	info := make(map[string][5]int)

	// Regex for matching object file paths and extracting the name
	reObjFile := regexp.MustCompile(`.+\/([\w\.]+)\.o$`)

	// Regexes for parsing elf_size output lines
	reFlash := regexp.MustCompile(`(?:\.irom0\.text|\.text|\.text1|\.data|\.rodata)\S*\s+([0-9]+).*`)
	reData := regexp.MustCompile(`^\.data\S*\s+([0-9]+).*`)     // Escaped the leading dot
	reRodata := regexp.MustCompile(`^\.rodata\S*\s+([0-9]+).*`) // Escaped the leading dot
	reBss := regexp.MustCompile(`^\.bss\S*\s+([0-9]+).*`)       // Escaped the leading dot

	for _, objFile := range objFiles {
		matches := reObjFile.FindStringSubmatch(objFile)
		if len(matches) < 2 {
			continue // Skip files that don't match the exact pattern (e.g., no path component or doesn't end in .o)
		}
		name := matches[1] // The captured group is at index 1

		currentInfo := [5]int{} // {Flash, RAM, data, rodata, bss} initialized to zeros

		// Execute the elf_size command for the current object file
		cmd := exec.Command(elfSizePath, "-A", objFile)
		output, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing '%s -A %s': %v\n", elfSizePath, objFile, err)
			continue // Skip processing this file
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if matches := reFlash.FindStringSubmatch(line); len(matches) > 1 {
				size, err := strconv.Atoi(matches[1])
				if err == nil { // Only add if conversion is successful
					currentInfo[0] += size
				}
			}
			if matches := reData.FindStringSubmatch(line); len(matches) > 1 {
				size, err := strconv.Atoi(matches[1])
				if err == nil {
					currentInfo[2] += size
				}
			}
			if matches := reRodata.FindStringSubmatch(line); len(matches) > 1 {
				size, err := strconv.Atoi(matches[1])
				if err == nil {
					currentInfo[3] += size
				}
			}
			if matches := reBss.FindStringSubmatch(line); len(matches) > 1 {
				size, err := strconv.Atoi(matches[1])
				if err == nil {
					currentInfo[4] += size
				}
			}
		}

		currentInfo[1] = currentInfo[2] + currentInfo[3] + currentInfo[4]

		// Store the accumulated info in the map
		info[name] = currentInfo
	}

	// Get the keys (object file names) from the map
	var names []string
	for name := range info {
		names = append(names, name)
	}

	// Sort the keys based on the values in the info map
	sort.Slice(names, func(i, j int) bool {
		nameI := names[i]
		nameJ := names[j]
		// Primary sort: descending by sortIndex
		if info[nameI][sortIndex] != info[nameJ][sortIndex] {
			return info[nameI][sortIndex] > info[nameJ][sortIndex] // i comes before j if info[i] > info[j]
		}
		// Secondary sort: descending by index 0 (Flash)
		return info[nameI][0] > info[nameJ][0] // i comes before j if info[i] > info[j]
	})

	// Print the sorted information
	for _, name := range names {
		fileInfo := info[name]
		fmt.Printf(form, name, fileInfo[0], fileInfo[1], fileInfo[2], fileInfo[3], fileInfo[4])
	}

	return nil
}
