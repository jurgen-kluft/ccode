package clay

import (
	"bufio"
	"os"
	"runtime"
	"strings"

	"github.com/jurgen-kluft/ccode/foundation"
)

//
// Esp32 toolchain parser parses the following 2 files from the ESP32 SDK:
// - /sdk-path/boards.txt
// - /sdk-path/platform.txt

// Fundamentals:
//
// Variables are used and are identified as
//
//          {name.of.variable}
//
// The dot is used in a hierarchical sense.
//
// Variables can be nested, for example:
//  - {build.extra_flags.{build.mcu}}
//
// The top hierachy is named 'toolchain' and the first 2 variables are:
//   - name=
//   - version=
//   - the name is the name of the toolchain (e.g. ESP32 Arduino)
//   - the version is the version of the toolchain (e.g. 3.2.0)
//
// Parsing, a line can be:
// - empty
// - a comment (starts with #)
// - a variable
//
// Some other lines can be ignored, these are the patterns:
// - menu.xxxxxx
// - {boardname}.menu.xxxx etc..

// The first file to parse is 'boards.txt', this file contains the board definitions.
// There are around 340 boards in the boards.txt file and the values of the variables
// also use other variables, for example these are the ones used in 'boards.txt':
//  - {runtime.platform.path}
//  - {serial.port}
//  - {build.path}
//  - {build.board}
//  - {build.boot}
//  - {build.mcu}
//  - {build.psram}
//  - {build.variant}
//  - {build.variant.path}
//  - {build.band}
//  - {build.einksize}
// Remember, they are hierarchical, so the above {build.xxx} variables mean the
// ones for the board that you are compiling for.

// In the platform.txt, there are 224 variables and the most top one is
// {runtime.os} is the OS that you are running on (windows, linux, darwin)
// {runtime.platform.path}, this is the path to the SDK.
// {runtime.ide.version} the version of the IDE that you are using (e.g. 3.2.0)
//
// Some other very important variables are:
//   - {build.path} - the path to the build folder
//   - {build.tarch} - the target architecture (e.g. xtensa or riscv32)
//   - {build.mcu} - the MCU name (esp32, esp32s2, esp32s3, etc.)
//   - {build.variant} - the variant name

// For a board, we strip away the first part of the variable:
// e.g: 'esp32wrover.name=ESP32 Wrover Module' becomes 'name=ESP32 Wrover Module'

// A 'new' board is recognized by a line looking like this:
//   xxxxxxx.name=ESP32 Wrover Module

type Esp32Board struct {
	Name        string            // The name of the board
	Description string            // The description of the board
	FlashSizes  map[string]string // The list of flash sizes
	Vars        *KeyValueSet
}

func NewBoard(name string, description string) *Esp32Board {
	return &Esp32Board{
		Name:        name,
		Description: description,
		FlashSizes:  make(map[string]string),
		Vars:        NewKeyValueSet(),
	}
}

type Esp32ToolFunction struct {
	Function string              // e.g. upload, program, erase, bootloader
	Cmd      string              // .pattern
	CmdLine  string              //
	Args     []string            // .pattern_args
	Vars     map[string][]string // A map of variables, e.g. 'upload.protocol=serial' or 'upload.params.verbose='
}

func NewEsp32ToolFunction(function string) *Esp32ToolFunction {
	return &Esp32ToolFunction{
		Function: function,
		Cmd:      "",
		Args:     make([]string, 0),
		Vars:     make(map[string][]string),
	}
}

// The following tools are taken:
// - esptool_py
// - esp_ota
type Esp32Tool struct {
	Name      string
	Vars      map[string]string             // A map of variables, e.g. 'runtime.os' or 'build.path'
	Functions map[string]*Esp32ToolFunction // The list of functions for the tool, e.g. upload, program, erase, bootloader
}

func NewEsp32Tool(name string) *Esp32Tool {
	return &Esp32Tool{
		Name:      name,
		Vars:      make(map[string]string),
		Functions: make(map[string]*Esp32ToolFunction),
	}
}

type Esp32Platform struct {
	Name                    string                // The name of the platform
	Version                 string                // The version of the platform
	Vars                    *KeyValueSet          // A map of variables, e.g. 'runtime.os' or 'build.path'
	CCompilerCmd            string                // C compiler command ('recipe.c.o.pattern')
	CCompilerCmdLine        string                //
	CCompilerArgs           []string              // The arguments for the C compiler
	CppCompilerCmd          string                // C++ compiler command ('recipe.cpp.o.pattern')
	CppCompilerCmdLine      string                // C++ compiler command ('recipe.cpp.o.pattern')
	CppCompilerArgs         []string              // The arguments for the C++ compiler
	AssemblerCmd            string                // S Assembler command ('recipe.S.o.pattern')
	AssemblerCmdLine        string                // S Assembler command ('recipe.S.o.pattern')
	AssemblerArgs           []string              // The arguments for the assembler
	ArchiverCmd             string                // Archiver command ('recipe.ar.pattern')
	ArchiverCmdLine         string                // Archiver command ('recipe.ar.pattern')
	ArchiverArgs            []string              // The arguments for the archiver
	LinkerCmd               string                // Linker command ('recipe.c.combine.pattern')
	LinkerCmdLine           string                // Linker command ('recipe.c.combine.pattern')
	LinkerArgs              []string              // The arguments for the linker
	CreatePartitionsCmd     string                // CreatePartitions command ('recipe.objcopy.partitions.bin.pattern')
	CreatePartitionsCmdLine string                // CreatePartitions command ('recipe.objcopy.partitions.bin.pattern')
	CreatePartitionsArgs    []string              // The arguments for the create partitions command
	CreateBinCmd            string                // CreateBin command ('recipe.objcopy.bin.pattern')
	CreateBinCmdLine        string                // CreateBin command ('recipe.objcopy.bin.pattern')
	CreateBinArgs           []string              // The arguments for the create bin command
	CreatBootloaderCmd      string                // CreateBootloader command ('recipe.hooks.prebuild.4.pattern')
	CreatBootloaderCmdLine  string                // CreateBootloader command ('recipe.hooks.prebuild.4.pattern')
	CreatBootloaderArgs     []string              // The arguments for the create bootloader command
	CreateMergedBinCmd      string                // CreateMergedBin command ('recipe.hooks.objcopy.postobjcopy.3.pattern')
	CreateMergedBinCmdLine  string                // CreateMergedBin command ('recipe.hooks.objcopy.postobjcopy.3.pattern')
	CreateMergedBinArgs     []string              // The arguments for the create merged bin command
	ComputeSizeCmd          string                // ComputeSize command ('recipe.size.pattern')
	ComputeSizeCmdLine      string                // ComputeSize command ('recipe.size.pattern')
	ComputeSizeArgs         []string              // The arguments for the compute size command
	Tools                   map[string]*Esp32Tool // The list of tools (only 'tools.esptool_py' and 'esp_ota' for now)
}

func NewPlatform() *Esp32Platform {
	return &Esp32Platform{
		Name:                    "",
		Version:                 "",
		Vars:                    NewKeyValueSet(),
		CCompilerCmd:            "",
		CCompilerCmdLine:        "",
		CCompilerArgs:           make([]string, 0),
		CppCompilerCmd:          "",
		CppCompilerCmdLine:      "",
		CppCompilerArgs:         make([]string, 0),
		AssemblerCmd:            "",
		AssemblerCmdLine:        "",
		AssemblerArgs:           make([]string, 0),
		ArchiverCmd:             "",
		ArchiverCmdLine:         "",
		ArchiverArgs:            make([]string, 0),
		LinkerCmd:               "",
		LinkerCmdLine:           "",
		LinkerArgs:              make([]string, 0),
		CreatePartitionsCmd:     "",
		CreatePartitionsCmdLine: "",
		CreatePartitionsArgs:    make([]string, 0),
		CreateBinCmd:            "",
		CreateBinCmdLine:        "",
		CreateBinArgs:           make([]string, 0),
		CreateMergedBinCmd:      "",
		CreateMergedBinCmdLine:  "",
		CreateMergedBinArgs:     make([]string, 0),
		ComputeSizeCmd:          "",
		ComputeSizeCmdLine:      "",
		ComputeSizeArgs:         make([]string, 0),
		Tools:                   make(map[string]*Esp32Tool),
	}
}

type Esp32Toolchain struct {
	Name        string // The name of the toolchain
	Version     string // The version of the toolchain
	SdkPath     string
	Boards      []*Esp32Board  // The list of boards
	NameToBoard map[string]int // A map of board names to their index in the boards slice
	Platform    *Esp32Platform
}

func ParseEsp32Toolchain(espSdkPath string) (*Esp32Toolchain, error) {

	boardsFile := espSdkPath + "/boards.txt"
	platformFile := espSdkPath + "/platform.txt"

	// Create the toolchain object
	toolchain := &Esp32Toolchain{
		Name:        "",
		Version:     "",
		SdkPath:     espSdkPath,
		Boards:      make([]*Esp32Board, 0, 350),
		NameToBoard: make(map[string]int),
		Platform:    NewPlatform(),
	}

	// Read the boards.txt file
	err := toolchain.ParseEsp32Boards(boardsFile)
	if err != nil {
		return nil, err
	}

	// Read the platform.txt file
	err = toolchain.ParseEsp32Platform(platformFile)
	if err != nil {
		return nil, err
	}

	return toolchain, nil
}

func ResolveString(variable string, vars *KeyValueSet) string {
	type pair struct {
		from int
		to   int
	}

	for true {
		// find [from:to] pairs of variables
		stack := make([]int16, 0, 4)
		list := make([]pair, 0)
		for i, c := range variable {
			if c == '{' {
				current := int16(len(list))
				stack = append(stack, current)
				list = append(list, pair{from: i, to: -1})
			} else if c == '}' {
				if len(list) > 0 {
					current := stack[len(stack)-1]
					list[current].to = i
					stack = stack[:len(stack)-1]
				}
			}
		}

		if len(list) == 0 {
			return variable
		}

		// See if we have an invalid pair, if so just return
		for _, p := range list {
			if p.to == -1 {
				foundation.LogWarningf("Invalid variable pair in string: %s\n", variable)
				return variable // Return the original string if we have an invalid pair
			}
		}

		// resolve the variables, last to first, and assume all pairs are valid and closed
		replaced := 0
		for i := len(list) - 1; i >= 0; i-- {
			p := list[i]
			variableName := variable[p.from+1 : p.to]
			// Check if the variable exists in the vars map
			if value, ok := vars.HasGet(variableName); ok {
				// Replace the variable with its value
				variable = variable[:p.from] + value + variable[p.to+1:]
				replaced += 1
				// It this was a nested variable (has overlap with previous pair(s)),
				// we need to adjust the 'to' of the next pairs
				for j := i - 1; j >= 0; j-- {
					if list[j].to > p.from {
						list[j].to += len(value) - (p.to - p.from + 1)
					}
				}
			} else {
				// If the variable does not exist, we skip it.
				// This variable could have been a nested one, so we need to skip also
				// the overlapping pairs
				for j := i - 1; j >= 0; j-- {
					if list[j].to > p.from {
						i-- // Skip this pair
					} else {
						break // No more overlapping pairs
					}
				}
			}
		}

		if replaced == 0 {
			return variable
		}

	}

	return variable
}

func RemoveEmptyEntries(list []string) []string {
	for i, item := range list {
		if item == "" {
			newList := list[:i]
			for j := i + 1; j < len(list); j++ {
				if list[j] != "" {
					newList = append(newList, list[j])
				}
			}
			return newList
		}
	}
	return list
}

func ResolveStringList(variableList []string, vars *KeyValueSet) []string {
	for i, variable := range variableList {
		variableList[i] = ResolveString(variable, vars)
	}
	return variableList
}

func SplitCmdLineIntoArgs(cmdline string, removeEmptyEntries bool) []string {
	var args []string
	for len(cmdline) > 0 {
		i := 0
		for i < len(cmdline) && cmdline[i] == ' ' {
			i++
		}
		cmdline = cmdline[i:] // Remove leading spaces
		if cmdline[0] == '"' {
			// Find the closing quote
			endQuote := strings.Index(cmdline[1:], "\"")
			if endQuote == -1 {
				// No closing quote found, return the original string
				args = append(args, cmdline)
				break
			}
			// Add the argument without the quotes
			if removeEmptyEntries && endQuote == 0 {
				// If we are removing empty entries, skip this argument
				cmdline = cmdline[endQuote+2:] // Move past the closing quote and space
				continue
			}
			args = append(args, cmdline[1:endQuote+1])
			cmdline = cmdline[endQuote+2:] // Move past the closing quote and space
		} else {
			// Find the next space
			nextSpace := strings.Index(cmdline, " ")
			if nextSpace == -1 {
				args = append(args, cmdline)
				break
			}
			// Add the argument before the space
			if removeEmptyEntries && nextSpace == 0 {
				cmdline = cmdline[nextSpace+1:] // Move past the space
				continue
			}
			args = append(args, cmdline[:nextSpace])
			cmdline = cmdline[nextSpace+1:] // Move past the space
		}
	}
	return args
}

func (t *Esp32Toolchain) ResolveVariables(board string) error {

	globalVars := NewKeyValueSet()
	globalVars.Add("runtime.os", runtime.GOOS)
	globalVars.Add("runtime.platform.path", t.SdkPath)
	globalVars.Add("runtime.ide.version", t.Platform.Version)
	globalVars.Add("build.path", "build")

	if boardIndex, boardExists := t.NameToBoard[board]; !boardExists {
		return foundation.LogErrorf(os.ErrInvalid, "Invalid board name: %s", board)
	} else {
		board := t.Boards[boardIndex]
		for i, k := range board.Vars.Keys {
			globalVars.Add(k, board.Vars.Values[i])
		}
	}

	for i, _ := range t.Platform.Vars.Keys {
		v := t.Platform.Vars.Values[i]
		v = ResolveString(v, globalVars)
		t.Platform.Vars.Values[i] = v
	}

	globalVars.Merge(t.Platform.Vars, false)

	for i, _ := range globalVars.Keys {
		v := globalVars.Values[i]
		v = ResolveString(v, globalVars)
		globalVars.Values[i] = v
	}

	// For platform we can resolve some of the variables that are local to the platform
	t.Platform.CCompilerCmd = ResolveString(t.Platform.CCompilerCmd, globalVars)
	t.Platform.CCompilerCmdLine = ResolveString(t.Platform.CCompilerCmdLine, globalVars)
	t.Platform.CCompilerArgs = SplitCmdLineIntoArgs(t.Platform.CCompilerCmdLine, true)

	t.Platform.CppCompilerCmd = ResolveString(t.Platform.CppCompilerCmd, globalVars)
	t.Platform.CppCompilerCmdLine = ResolveString(t.Platform.CppCompilerCmdLine, globalVars)
	t.Platform.CppCompilerArgs = SplitCmdLineIntoArgs(t.Platform.CppCompilerCmdLine, true)

	t.Platform.AssemblerCmd = ResolveString(t.Platform.AssemblerCmd, globalVars)
	t.Platform.AssemblerCmdLine = ResolveString(t.Platform.AssemblerCmdLine, globalVars)
	t.Platform.AssemblerArgs = SplitCmdLineIntoArgs(t.Platform.AssemblerCmdLine, true)

	t.Platform.ArchiverCmd = ResolveString(t.Platform.ArchiverCmd, globalVars)
	t.Platform.ArchiverCmdLine = ResolveString(t.Platform.ArchiverCmdLine, globalVars)
	t.Platform.ArchiverArgs = SplitCmdLineIntoArgs(t.Platform.ArchiverCmdLine, true)

	t.Platform.LinkerCmd = ResolveString(t.Platform.LinkerCmd, globalVars)
	t.Platform.LinkerCmdLine = ResolveString(t.Platform.LinkerCmdLine, globalVars)
	t.Platform.LinkerArgs = SplitCmdLineIntoArgs(t.Platform.LinkerCmdLine, true)

	t.Platform.CreatePartitionsCmd = ResolveString(t.Platform.CreatePartitionsCmd, globalVars)
	t.Platform.CreatePartitionsCmdLine = ResolveString(t.Platform.CreatePartitionsCmdLine, globalVars)
	t.Platform.CreatePartitionsArgs = SplitCmdLineIntoArgs(t.Platform.CreatePartitionsCmdLine, true)

	t.Platform.CreateBinCmd = ResolveString(t.Platform.CreateBinCmd, globalVars)
	t.Platform.CreateBinCmdLine = ResolveString(t.Platform.CreateBinCmdLine, globalVars)
	t.Platform.CreateBinArgs = SplitCmdLineIntoArgs(t.Platform.CreateBinCmdLine, true)

	t.Platform.CreatBootloaderCmd = ResolveString(t.Platform.CreatBootloaderCmd, globalVars)
	t.Platform.CreatBootloaderCmdLine = ResolveString(t.Platform.CreatBootloaderCmdLine, globalVars)
	t.Platform.CreatBootloaderArgs = SplitCmdLineIntoArgs(t.Platform.CreatBootloaderCmdLine, true)

	t.Platform.CreateMergedBinCmd = ResolveString(t.Platform.CreateMergedBinCmd, globalVars)
	t.Platform.CreateMergedBinCmdLine = ResolveString(t.Platform.CreateMergedBinCmdLine, globalVars)
	t.Platform.CreateMergedBinArgs = SplitCmdLineIntoArgs(t.Platform.CreateMergedBinCmdLine, true)

	t.Platform.ComputeSizeCmd = ResolveString(t.Platform.ComputeSizeCmd, globalVars)
	t.Platform.ComputeSizeCmdLine = ResolveString(t.Platform.ComputeSizeCmdLine, globalVars)
	t.Platform.ComputeSizeArgs = SplitCmdLineIntoArgs(t.Platform.ComputeSizeCmdLine, true)

	// For tools we can resolve some of the variables that are local to the tool
	for _, tool := range t.Platform.Tools {
		//tool.Vars = ResolveStringList(tool.Vars, t.Platform.Vars)
		for _, function := range tool.Functions {
			//function.Vars = ResolveStringList(function.Vars, t.Platform.Vars)
			function.Cmd = ResolveString(function.Cmd, t.Platform.Vars)
			function.Args = ResolveStringList(function.Args, t.Platform.Vars)
		}
	}
	return nil
}

func (t *Esp32Toolchain) ParseEsp32Boards(boardsFile string) error {
	file, err := os.OpenFile(boardsFile, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var currentBoard *Esp32Board

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Skip lines that start with 'menu.'
		if strings.HasPrefix(line, "menu.") {
			continue
		}

		keyAndValue := strings.SplitN(line, "=", 2)
		key := keyAndValue[0]
		value := keyAndValue[1]
		keyParts := strings.Split(key, ".")

		// Check if the line is a board definition
		if len(keyParts) == 2 && keyParts[len(keyParts)-1] == "name" {
			if currentBoard != nil {
				t.Boards = append(t.Boards, currentBoard)
			}
			currentBoard = NewBoard(keyParts[0], value)
			continue
		}

		if currentBoard != nil {
			if keyParts[0] == currentBoard.Name && keyParts[1] == "menu" {
				if strings.EqualFold(keyParts[2], "flashsize") {
					flashsize := strings.Join(keyParts[3:], ".")
					currentBoard.FlashSizes[flashsize] = value
				}
			} else {
				key = strings.TrimPrefix(key, currentBoard.Name)
				key = strings.TrimPrefix(key, ".") // Remove the leading dot if present
				currentBoard.Vars.Add(key, value)
			}
		}
	}

	if currentBoard != nil {
		t.Boards = append(t.Boards, currentBoard)
	}

	// Create a map of board names to their index in the boards slice
	for i, board := range t.Boards {
		t.NameToBoard[board.Name] = i
	}

	return nil
}

func ParseCmdAndArgs(cmd string) (string, string) {
	// The arguments follow the first "cmd" part, also the arguments need to be split by ' '
	args := strings.SplitN(cmd, " ", 2)
	cmd = strings.TrimFunc(args[0], func(r rune) bool {
		return r == '"' || r == '\'' || r == ' '
	})
	if len(args) > 1 {
		return cmd, args[1]
	}
	return cmd, ""
}

func ParseArgs(args string) []string {
	// The arguments follow the first "cmd" part, also the arguments need to be split by ' '
	argsList := strings.Split(args, " ")
	for i := 0; i < len(argsList); i++ {
		argsList[i] = strings.TrimFunc(argsList[i], func(r rune) bool {
			return r == '"' || r == '\'' || r == ' '
		})
	}
	return argsList
}

func (t *Esp32Toolchain) ParseEsp32Platform(platformFile string) error {
	file, err := os.OpenFile(platformFile, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	t.Platform.CreatBootloaderCmd = `"{tools.esptool_py.path}\{tools.esptool_py.cmd}" {recipe.hooks.prebuild.4.pattern_args} "{build.path}\{build.project_name}.bootloader.bin" "{compiler.sdk.path}\bin\bootloader_{build.boot}_{build.boot_freq}.elf"`
	t.Platform.CreatBootloaderCmd, t.Platform.CreatBootloaderCmdLine = ParseCmdAndArgs(t.Platform.CreatBootloaderCmd)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Ignore 'debug.', 'debug_config.' and 'debug_script.' lines
		if strings.HasPrefix(line, "debug.") || strings.HasPrefix(line, "debug_config.") || strings.HasPrefix(line, "debug_script.") {
			continue
		}

		keyAndValue := strings.SplitN(line, "=", 2)
		key := keyAndValue[0]
		value := keyAndValue[1]
		keyParts := strings.Split(key, ".")

		if len(keyParts) == 1 && keyParts[0] == "name" {
			t.Platform.Name = value
			continue
		}

		if len(keyParts) == 1 && keyParts[0] == "version" {
			t.Platform.Version = value
			continue
		}

		if keyParts[len(keyParts)-1] == "windows" {
			// Ignored for now
			if runtime.GOOS != "windows" {
				continue
			}
			// key = strings.TrimSuffix(key, ".windows")
			// keyParts = keyParts[:len(keyParts)-1]
		}

		ignoreAsVar := false

		if len(keyParts) >= 3 && keyParts[0] == "recipe" {
			ignoreAsVar = true
			if strings.Compare(key, "recipe.c.o.pattern") == 0 {
				t.Platform.CCompilerCmd = value
				t.Platform.CCompilerCmd, t.Platform.CCompilerCmdLine = ParseCmdAndArgs(t.Platform.CCompilerCmd)
				continue
			}
			if strings.Compare(key, "recipe.cpp.o.pattern") == 0 {
				t.Platform.CppCompilerCmd = value
				t.Platform.CppCompilerCmd, t.Platform.CppCompilerCmdLine = ParseCmdAndArgs(t.Platform.CppCompilerCmd)
				continue
			}
			if strings.Compare(key, "recipe.S.o.pattern") == 0 {
				t.Platform.AssemblerCmd = value
				t.Platform.AssemblerCmd, t.Platform.AssemblerCmdLine = ParseCmdAndArgs(t.Platform.AssemblerCmd)
				continue
			}
			if strings.Compare(key, "recipe.ar.pattern") == 0 {
				t.Platform.ArchiverCmd = value
				t.Platform.ArchiverCmd, t.Platform.ArchiverCmdLine = ParseCmdAndArgs(t.Platform.ArchiverCmd)
				continue
			}
			if strings.Compare(key, "recipe.c.combine.pattern") == 0 {
				t.Platform.LinkerCmd = value
				t.Platform.LinkerCmd, t.Platform.LinkerCmdLine = ParseCmdAndArgs(t.Platform.LinkerCmd)
				continue
			}
			if strings.Compare(key, "recipe.objcopy.partitions.bin.pattern") == 0 {
				t.Platform.CreatePartitionsCmd = value
				t.Platform.CreatePartitionsCmd, t.Platform.CreatePartitionsCmdLine = ParseCmdAndArgs(t.Platform.CreatePartitionsCmd)
				continue
			}
			if strings.Compare(key, "recipe.objcopy.bin.pattern") == 0 {
				t.Platform.CreateBinCmd = value
				t.Platform.CreateBinCmd, t.Platform.CreateBinCmdLine = ParseCmdAndArgs(t.Platform.CreateBinCmd)
				continue
			}
			if strings.Compare(key, "recipe.hooks.objcopy.postobjcopy.3.pattern") == 0 {
				t.Platform.CreateMergedBinCmd = value
				t.Platform.CreateMergedBinCmd, t.Platform.CreateMergedBinCmdLine = ParseCmdAndArgs(t.Platform.CreateMergedBinCmd)
				continue
			}
			if strings.Compare(key, "recipe.size.pattern") == 0 {
				t.Platform.ComputeSizeCmd = value
				t.Platform.ComputeSizeCmd, t.Platform.ComputeSizeCmdLine = ParseCmdAndArgs(t.Platform.ComputeSizeCmd)
				continue
			}
			ignoreAsVar = false
		} else if keyParts[0] == "tools" && (keyParts[1] == "esptool_py" || keyParts[1] == "esp_ota" || keyParts[1] == "gen_esp32part" || keyParts[1] == "gen_insights_pkg") {

			toolName := keyParts[1] // e.g. 'esptool_py' or 'esp_ota'

			// Check if the tool already exists
			tool, exists := t.Platform.Tools[toolName]
			if !exists {
				tool = NewEsp32Tool(toolName)
				t.Platform.Tools[toolName] = tool
			}

			if keyParts[2] == "path" || keyParts[2] == "cmd" {
				tool.Vars[keyParts[2]] = value
			} else {
				isFunction := keyParts[2] == "upload" || keyParts[2] == "program" || keyParts[2] == "erase" || keyParts[2] == "bootloader"
				if isFunction {
					toolFunction := keyParts[2] // e.g. 'upload'

					// Check if the function already exists
					function, exists := tool.Functions[toolFunction]
					if !exists {
						function = NewEsp32ToolFunction(toolFunction)
						tool.Functions[toolFunction] = function
					}
					// Now we can set the variable based on the toolVar
					if keyParts[len(keyParts)-1] == "pattern" {
						function.Cmd, function.CmdLine = ParseCmdAndArgs(value)
					} else if keyParts[len(keyParts)-1] == "pattern_args" {
						function.Args = append(function.Args, ParseArgs(value)...)
						function.Vars[strings.Join(keyParts[2:], ".")] = function.Args
					} else {
						function.Vars[strings.Join(keyParts[2:], ".")] = []string{value}
					}
				} else {

				}
			}
		} else {
			// Special case for 'build.extra_flags.boardname=value'
			// strings.HasPrefix(key, "build.extra_flags.")
			if len(keyParts) == 3 && keyParts[0] == "build" && keyParts[1] == "extra_flags" {
				ignoreAsVar = true

				// The last part is a board name, we are going to
				// add 'build.extra_flags=value' to the matching board
				// in the toolchain, so remove the '.boardname'
				boardName := keyParts[2]
				if i, ok := t.NameToBoard[boardName]; ok {
					t.Boards[i].Vars.Add(keyParts[0]+"."+keyParts[1], value)
				}
			}
		}

		if !ignoreAsVar {
			// Add the variable to the platform variables
			t.Platform.Vars.Add(key, value)
		}
	}

	return nil
}
