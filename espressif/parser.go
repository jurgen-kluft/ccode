package cespressif

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

type ToolFunction struct {
	Function string            // e.g. upload, program, erase, bootloader
	Cmd      string            // .pattern
	CmdLine  string            //
	Args     []string          // .pattern_args
	Vars     map[string]string // A map of variables, e.g. 'runtime.os' or 'build.path'
}

func newToolFunction() *ToolFunction {
	return &ToolFunction{Args: make([]string, 0), Vars: make(map[string]string)}
}

func decodeJsonToolFunction(decoder *corepkg.JsonDecoder) *ToolFunction {
	tf := newToolFunction()
	fields := map[string]corepkg.JsonDecode{
		"function": func(decoder *corepkg.JsonDecoder) { tf.Function = decoder.DecodeString() },
		"cmd":      func(decoder *corepkg.JsonDecoder) { tf.Cmd = decoder.DecodeString() },
		"cmdline":  func(decoder *corepkg.JsonDecoder) { tf.CmdLine = decoder.DecodeString() },
		"args":     func(decoder *corepkg.JsonDecoder) { tf.Args = decoder.DecodeStringArray() },
		"vars":     func(decoder *corepkg.JsonDecoder) { tf.Vars = decoder.DecodeStringMapString() },
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding tool function: %s", err.Error())
	}
	return tf
}

func encodeJsonToolFunction(encoder *corepkg.JsonEncoder, key string, object *ToolFunction) {
	encoder.BeginObject(key)
	{
		encoder.WriteField("function", object.Function)
		encoder.WriteField("cmd", object.Cmd)
		encoder.WriteField("cmdline", object.CmdLine)
		if len(object.Args) > 0 {
			encoder.WriteStringArray("args", object.Args)
		}
		if len(object.Vars) > 0 {
			encoder.WriteStringMapString("vars", object.Vars)
		}
	}
	encoder.EndObject()
}

type tool struct {
	Name      string
	Vars      map[string]string        // A map of variables, e.g. 'runtime.os' or 'build.path'
	Functions map[string]*ToolFunction // The list of functions for the tool, e.g. upload, program, erase, bootloader
}

func newTool() *tool {
	return &tool{Vars: make(map[string]string), Functions: make(map[string]*ToolFunction)}
}

func decodeJsonDecodeJsonTool(decoder *corepkg.JsonDecoder) *tool {
	t := newTool()
	fields := map[string]corepkg.JsonDecode{
		"name": func(decoder *corepkg.JsonDecoder) { t.Name = decoder.DecodeString() },
		"vars": func(decoder *corepkg.JsonDecoder) { t.Vars = decoder.DecodeStringMapString() },
		"functions": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				function := decodeJsonToolFunction(decoder)
				t.Functions[function.Function] = function
			})
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding tool: %s", err.Error())
	}
	return t
}

func encodeJsonTool(encoder *corepkg.JsonEncoder, key string, object *tool) {
	encoder.BeginObject(key)
	{
		encoder.WriteField("name", object.Name)
		if len(object.Vars) > 0 {
			encoder.WriteStringMapString("vars", object.Vars)
		}
		if len(object.Functions) > 0 {
			encoder.WriteAsArray("functions", func(encoder *corepkg.JsonEncoder) {
				for _, arg := range object.Functions {
					encodeJsonToolFunction(encoder, "", arg)
				}
			})
		}
	}
	encoder.EndObject()
}

type platformRecipe struct {
	Name string
	Cmd  string
}

func newPlatformRecipe() *platformRecipe {
	return &platformRecipe{Name: "", Cmd: ""}
}

func decodeJsonPlatformRecipe(decoder *corepkg.JsonDecoder) *platformRecipe {
	recipe := newPlatformRecipe()
	fields := map[string]corepkg.JsonDecode{
		"name": func(decoder *corepkg.JsonDecoder) { recipe.Name = decoder.DecodeString() },
		"cmd":  func(decoder *corepkg.JsonDecoder) { recipe.Cmd = decoder.DecodeString() },
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding platform recipe: %s", err.Error())
	}
	return recipe
}

func encodeJsonPlatformRecipe(encoder *corepkg.JsonEncoder, key string, object *platformRecipe) {
	encoder.BeginObject(key)
	{
		encoder.WriteField("name", object.Name)
		encoder.WriteField("cmd", object.Cmd)
	}
	encoder.EndObject()
}

type platform struct {
	Name    string                     // The name of the platform
	Version string                     // The version of the platform
	Vars    map[string]string          // A map of variables, e.g. 'runtime.os' or 'build.path'
	Recipes map[string]*platformRecipe // The recipes
	Tools   map[string]*tool           // The list of tools (only 'tools.esptool_py' and 'esp_ota' for now)
}

func newPlatform() *platform {
	return &platform{Vars: make(map[string]string), Recipes: make(map[string]*platformRecipe), Tools: make(map[string]*tool)}
}

func decodeJsonPlatform(decoder *corepkg.JsonDecoder) *platform {
	platform := newPlatform()
	fields := map[string]corepkg.JsonDecode{
		"name":    func(decoder *corepkg.JsonDecoder) { platform.Name = decoder.DecodeString() },
		"version": func(decoder *corepkg.JsonDecoder) { platform.Version = decoder.DecodeString() },
		"vars":    func(decoder *corepkg.JsonDecoder) { platform.Vars = decoder.DecodeStringMapString() },
		"recipes": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				recipe := decodeJsonPlatformRecipe(decoder)
				platform.Recipes[recipe.Name] = recipe
			})
		},
		"tools": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				tool := decodeJsonDecodeJsonTool(decoder)
				platform.Tools[tool.Name] = tool
			})
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding platform: %s", err.Error())
	}
	return platform
}

func encodeJsonPlatform(encoder *corepkg.JsonEncoder, key string, object *platform) {
	if object == nil {
		return
	}
	encoder.BeginObject(key)
	{
		encoder.WriteField("name", object.Name)
		encoder.WriteField("version", object.Version)
		if len(object.Vars) > 0 {
			encoder.WriteStringMapString("vars", object.Vars)
		}
		if len(object.Recipes) > 0 {
			encoder.WriteAsArray("recipes", func(encoder *corepkg.JsonEncoder) {
				for _, v := range object.Recipes {
					encodeJsonPlatformRecipe(encoder, "", v)
				}
			})
		}
		if len(object.Tools) > 0 {
			encoder.WriteAsArray("tools", func(encoder *corepkg.JsonEncoder) {
				for _, v := range object.Tools {
					encodeJsonTool(encoder, "", v)
				}
			})
		}
	}
	encoder.EndObject()
}

type boardMenuSubEntry struct {
	Name   string
	Title  string
	Keys   []string
	Values []string
}

func newBoardMenuSubEntry() *boardMenuSubEntry {
	return &boardMenuSubEntry{Keys: make([]string, 0, 10), Values: make([]string, 0, 10)}
}

func encodeJsonBoardMenuSubEntry(encoder *corepkg.JsonEncoder, key string, object *boardMenuSubEntry) {
	if object == nil {
		return
	}

	encoder.BeginObject(key)
	{
		encoder.WriteField("name", object.Name)
		encoder.WriteField("title", object.Title)
		if len(object.Keys) > 0 {
			encoder.WriteStringArray("keys", object.Keys)
		}
		if len(object.Values) > 0 {
			encoder.WriteStringArray("values", object.Values)
		}
	}
	encoder.EndObject()
}

func decodeJsonBoardMenuSubEntry(decoder *corepkg.JsonDecoder) *boardMenuSubEntry {
	m := newBoardMenuSubEntry()
	fields := map[string]corepkg.JsonDecode{
		"name":  func(decoder *corepkg.JsonDecoder) { m.Name = decoder.DecodeString() },
		"title": func(decoder *corepkg.JsonDecoder) { m.Title = decoder.DecodeString() },
		"keys":  func(decoder *corepkg.JsonDecoder) { m.Keys = decoder.DecodeStringArray() },
		"values": func(decoder *corepkg.JsonDecoder) {
			m.Values = decoder.DecodeStringArray()
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding board menu sub entry: %s", err.Error())
	}
	return m
}

type boardMenuEntry struct {
	Name       string // ssl, mmu, FlashFreq, ...
	Current    *boardMenuSubEntry
	SubEntries []*boardMenuSubEntry
}

func newBoardMenuEntry() *boardMenuEntry {
	return &boardMenuEntry{Current: nil, SubEntries: make([]*boardMenuSubEntry, 0)}
}

func encodeJsonBoardMenuEntry(encoder *corepkg.JsonEncoder, key string, object *boardMenuEntry) {
	if object == nil {
		return
	}
	encoder.BeginObject(key)
	{
		encoder.WriteField("name", object.Name)
		if len(object.SubEntries) > 0 {
			encoder.WriteAsArray("entries", func(encoder *corepkg.JsonEncoder) {
				for _, entry := range object.SubEntries {
					encodeJsonBoardMenuSubEntry(encoder, "", entry)
				}
			})
		}
	}
	encoder.EndObject()
}

func decodeJsonBoardMenuEntry(decoder *corepkg.JsonDecoder) *boardMenuEntry {
	m := newBoardMenuEntry()
	fields := map[string]corepkg.JsonDecode{
		"name": func(decoder *corepkg.JsonDecoder) { m.Name = decoder.DecodeString() },
		"entries": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				entry := decodeJsonBoardMenuSubEntry(decoder)
				m.SubEntries = append(m.SubEntries, entry)
			})
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding board menu entry: %s", err.Error())
	}
	return m
}

func (e *boardMenuEntry) parse(entryItem string, key string, value string) {
	if e.Current == nil || e.Current.Name != entryItem {
		e.Current = newBoardMenuSubEntry()
		e.Current.Name = entryItem
		e.SubEntries = append(e.SubEntries, e.Current)
	}
	if len(key) == 0 {
		e.Current.Title = value
	} else {
		e.Current.Keys = append(e.Current.Keys, key)
		e.Current.Values = append(e.Current.Values, value)
	}
}

func (m *boardMenuEntry) RegisterVars(vars *corepkg.Vars) {
	if len(m.SubEntries) > 0 {
		keys := m.SubEntries[0].Keys
		values := m.SubEntries[0].Values
		for i, k := range keys {
			vars.Set(k, values[i])
		}
	}
}

type boardMenu struct {
	Current *boardMenuEntry
	Entries []*boardMenuEntry
}

func newBoardMenu() *boardMenu {
	return &boardMenu{
		Current: nil,
		Entries: make([]*boardMenuEntry, 0),
	}
}

func encodeJsonBoardMenu(encoder *corepkg.JsonEncoder, key string, object *boardMenu) {
	if object == nil {
		return
	}
	encoder.BeginObject(key)
	{
		if len(object.Entries) > 0 {
			encoder.WriteArray("entries", len(object.Entries), func(encoder *corepkg.JsonEncoder, i int) {
				encodeJsonBoardMenuEntry(encoder, "", object.Entries[i])
			})
		}
	}
	encoder.EndObject()
}

func decodeJsonBoardMenu(decoder *corepkg.JsonDecoder) *boardMenu {
	entries := make([]*boardMenuEntry, 0)
	fields := map[string]corepkg.JsonDecode{
		"entries": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				entry := decodeJsonBoardMenuEntry(decoder)
				entries = append(entries, entry)
			})
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding board menu: %s", err.Error())
	}
	return &boardMenu{Current: nil, Entries: entries}
}

func (m *boardMenu) parse(entry string, subEntry string, key string, value string) {
	if m.Current == nil || m.Current.Name != entry {
		m.Current = newBoardMenuEntry()
		m.Current.Name = entry
		m.Entries = append(m.Entries, m.Current)
	}
	m.Current.parse(subEntry, key, value)
}

func (m *boardMenu) RegisterVars(vars *corepkg.Vars) {
	// For each menu entry, we register the first variables
	for _, e := range m.Entries {
		e.RegisterVars(vars)
	}
}

type board struct {
	Name        string     // The name of the board
	Description string     // The description of the board
	SdkPath     string     // The path to the SDK
	Menu        *boardMenu // Menu options for the board
	Vars        *corepkg.Vars
}

func newBoard() *board {
	return &board{Menu: newBoardMenu(), Vars: corepkg.NewVars(corepkg.VarsFormatCurlyBraces)}
}

func newBoardName(name, description, sdkPath string) *board {
	return &board{Name: name, Description: description, SdkPath: sdkPath, Menu: newBoardMenu(), Vars: corepkg.NewVars(corepkg.VarsFormatCurlyBraces)}
}

type toolchain struct {
	Name             string         // The name of the toolchain
	Version          string         // The version of the toolchain
	SdkPath          string         // The path to the SDK
	BoardsFileTime   int64          // The modification time of the boards.txt file
	PlatformFileTime int64          // The modification time of the platform.txt file
	FormatVersion    string         // The format version of the toolchain json file
	ListOfBoards     []*board       // The list of boards
	Platform         *platform      // The platform
	BoardNameToIndex map[string]int // A map of board names to their index in the boards slice
}

func encodeJsonBoard(encoder *corepkg.JsonEncoder, object *board) {
	if object == nil {
		return
	}
	encoder.BeginObject("")
	{
		encoder.WriteField("name", object.Name)
		encoder.WriteField("descr", object.Description)
		encoder.WriteField("sdk", object.SdkPath)
		encodeJsonBoardMenu(encoder, "menu", object.Menu)
		object.Vars.EncodeJson("vars", encoder)
	}
	encoder.EndObject()
}

func decodeJsonBoard(decoder *corepkg.JsonDecoder) *board {
	b := newBoard()
	fields := map[string]corepkg.JsonDecode{
		"name":  func(decoder *corepkg.JsonDecoder) { b.Name = decoder.DecodeString() },
		"descr": func(decoder *corepkg.JsonDecoder) { b.Description = decoder.DecodeString() },
		"sdk":   func(decoder *corepkg.JsonDecoder) { b.SdkPath = decoder.DecodeString() },
		"menu":  func(decoder *corepkg.JsonDecoder) { b.Menu = decodeJsonBoardMenu(decoder) },
		"vars": func(decoder *corepkg.JsonDecoder) {
			b.Vars = corepkg.NewVars(corepkg.VarsFormatCurlyBraces)
			b.Vars.DecodeJson(decoder)
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding board: %s", err.Error())
	}
	return b
}

func decodeJsonToolchain(toolchain *toolchain, decoder *corepkg.JsonDecoder) error {
	fields := map[string]corepkg.JsonDecode{
		"name":               func(decoder *corepkg.JsonDecoder) { toolchain.Name = decoder.DecodeString() },
		"version":            func(decoder *corepkg.JsonDecoder) { toolchain.Version = decoder.DecodeString() },
		"sdk":                func(decoder *corepkg.JsonDecoder) { toolchain.SdkPath = decoder.DecodeString() },
		"board_file_time":    func(decoder *corepkg.JsonDecoder) { toolchain.BoardsFileTime = decoder.DecodeInt64() },
		"platform_file_time": func(decoder *corepkg.JsonDecoder) { toolchain.PlatformFileTime = decoder.DecodeInt64() },
		"format_version":     func(decoder *corepkg.JsonDecoder) { toolchain.FormatVersion = decoder.DecodeString() },
		"boards": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				board := decodeJsonBoard(decoder)
				toolchain.ListOfBoards = append(toolchain.ListOfBoards, board)
			})
		},
		"platform": func(decoder *corepkg.JsonDecoder) {
			toolchain.Platform = decodeJsonPlatform(decoder)
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding toolchain: %s", err.Error())
	}

	// Construct the BoardNameToIndex map
	for i, board := range toolchain.ListOfBoards {
		toolchain.BoardNameToIndex[strings.ToLower(board.Name)] = i
	}

	return nil
}

func encodeJsonToolchain(encoder *corepkg.JsonEncoder, key string, object *toolchain) {
	if object == nil {
		return
	}
	encoder.BeginObject(key)
	{
		encoder.WriteField("name", object.Name)
		encoder.WriteField("version", object.Version)
		encoder.WriteField("sdk", object.SdkPath)
		{
			if len(object.ListOfBoards) > 0 {
				encoder.WriteArray("boards", len(object.ListOfBoards), func(encoder *corepkg.JsonEncoder, i int) {
					encodeJsonBoard(encoder, object.ListOfBoards[i])
				})
			}
		}
		encoder.WriteField("board_file_time", object.BoardsFileTime)
		encoder.WriteField("platform_file_time", object.PlatformFileTime)
		encoder.WriteField("format_version", object.FormatVersion)
		encodeJsonPlatform(encoder, "platform", object.Platform)
	}
	encoder.EndObject()
}

func sdkPath(arch string) string {
	sdkPath := ""
	switch arch {
	case "esp32":
		sdkPath = "$HOME/sdk/arduino/esp32"
		if env := os.Getenv("ESP32_SDK"); env != "" {
			sdkPath = env
		}
	case "esp8266":
		sdkPath = "$HOME/sdk/arduino/esp8266"
		if env := os.Getenv("ESP8266_SDK"); env != "" {
			sdkPath = env
		}
	}
	sdkPath = os.ExpandEnv(sdkPath)
	return sdkPath
}

func NewToolchain(arch string) *toolchain {
	switch arch {
	case "esp32":
		espSdkPath := sdkPath("esp32")
		toolchain := &toolchain{
			Name:             "Espressif ESP32 Arduino",
			Version:          "3.2.0",
			SdkPath:          espSdkPath,
			ListOfBoards:     make([]*board, 0, 350),
			BoardNameToIndex: make(map[string]int),
			Platform:         newPlatform(),
		}
		return toolchain
	case "esp8266":
		espSdkPath := sdkPath("esp8266")
		toolchain := &toolchain{
			Name:             "Espressif ESP8266 Arduino",
			Version:          "3.2.0",
			SdkPath:          espSdkPath,
			ListOfBoards:     make([]*board, 0, 350),
			BoardNameToIndex: make(map[string]int),
			Platform:         newPlatform(),
		}
		return toolchain
	}
	return nil
}

func (t *toolchain) PrintInfo() {
	corepkg.LogInfof("Toolchain: %s, version: %s", t.Name, t.Version)
	corepkg.LogInfof("SDK Path: %s", t.SdkPath)
	corepkg.LogInfof("Number of boards: %d", len(t.ListOfBoards))
	corepkg.LogInfof("Platform: %s", t.Platform.Name)
}

func (t *toolchain) GetBoardByName(name string) *board {
	if index, ok := t.BoardNameToIndex[strings.ToLower(name)]; ok {
		return t.ListOfBoards[index]
	}
	return nil
}

var toolchainFormatVersion = "1.0.5"

func ParseToolchain(arch string) (toolchain *toolchain, err error) {
	// Can we figure out if we already have a esp32.json or esp8266.json file and if it is up to date?
	// If so, we can load that file instead of parsing the boards.txt and platform.txt files again
	toolchain = NewToolchain(arch)
	toolchain.FormatVersion = toolchainFormatVersion

	archJsonFilepath := filepath.Join(arch + ".json")
	err = toolchain.loadJson(archJsonFilepath)

	boardsFilepath := filepath.Join(toolchain.SdkPath, "boards.txt")
	platformFilepath := filepath.Join(toolchain.SdkPath, "platform.txt")

	if err == nil {
		// Check the modification times of the boards.txt and platform.txt files
		boardsFileInfo, err1 := os.Stat(boardsFilepath)
		platformFileInfo, err2 := os.Stat(platformFilepath)
		if err1 == nil && err2 == nil {
			if boardsFileInfo.ModTime().Unix() == toolchain.BoardsFileTime && platformFileInfo.ModTime().Unix() == toolchain.PlatformFileTime && toolchain.FormatVersion == toolchainFormatVersion {
				// The json file is up to date, we can return the toolchain
				return toolchain, nil
			}
		}
	}

	// Read the boards.txt file
	err = toolchain.parseBoardsFile(boardsFilepath)
	if err != nil {
		return nil, err
	}

	// Read the platform.txt file
	err = toolchain.parsePlatformFile(platformFilepath)
	if err != nil {
		return nil, err
	}

	// Save the toolchain to a json file for faster loading next time
	if err := toolchain.saveJson(archJsonFilepath); err != nil {
		corepkg.LogErrorf(err, "error saving toolchain to json file: %s", err.Error())
	}

	return toolchain, nil
}

func GetVars(toolchain *toolchain, boardName string, vars *corepkg.Vars) error {

	// get board by name
	board := toolchain.GetBoardByName(boardName)
	if board == nil {
		return corepkg.LogError(fmt.Errorf("board '%s' not found in toolchain '%s'", boardName, toolchain.Name), "board not found")
	}

	corepkg.LogInfof("Using Board '%s' (%s) found in toolchain '%s'", board.Name, board.Description, toolchain.Name)

	return toolchain.ResolveVariablesForBoard(board, vars)
}

func (t *toolchain) ResolveVariablesForBoard(board *board, vars *corepkg.Vars) error {

	vars.Set("runtime.os", runtime.GOOS)
	vars.Set("runtime.platform.path", t.SdkPath)
	vars.Set("runtime.ide.version", "10607")
	vars.Set("board.name", board.Name)

	// Add all recipes to the board
	for _, recipe := range t.Platform.Recipes {
		vars.Set(recipe.Name, recipe.Cmd)
	}

	vars.Join(board.Vars)
	board.Menu.RegisterVars(vars)
	vars.JoinMap(t.Platform.Vars)

	for _, key := range vars.Keys {
		key = strings.ToLower(key)
		if strings.HasPrefix(key, "tools.") || strings.HasPrefix(key, "compiler.") || strings.HasPrefix(key, "build.") || strings.HasPrefix(key, "recipe.") || strings.HasPrefix(key, "upload.") {
			oldValues := vars.Values[vars.KeyToIndex22[key]]
			newValues := make([]string, 0, len(oldValues))
			for _, value := range oldValues {
				// For a certain set of key types, we should (smartly) split the value by space
				// The following key types are known to have command-line values:
				// - tools.
				// - compiler.
				// - build.
				// - recipe.
				// We are going to split the value by space, but we need to take care of quoted values
				// e.g. tools.esptool_py.cmd="python" -m esptool ...
				// We will use a simple state machine to parse the value
				args := parseArgs(value, true)
				newValues = append(newValues, args...)
			}
			vars.Values[vars.KeyToIndex22[key]] = newValues
		}
	}

	return nil
}

func (t *toolchain) parseBoardsFile(boardsFile string) error {
	file, err := os.OpenFile(boardsFile, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var currentBoard *board

	menuTitles := make(map[string]string, 40)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty and comment lines
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		// Remove any comment at the end of the line
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		keyAndValue := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(keyAndValue[0])
		value := strings.TrimSpace(keyAndValue[1])

		keyParts := strings.Split(key, ".")

		// Check if the line is a board definition
		if len(keyParts) == 2 && keyParts[len(keyParts)-1] == "name" {
			if currentBoard != nil {
				t.ListOfBoards = append(t.ListOfBoards, currentBoard)
			}
			currentBoard = newBoardName(keyParts[0], value, t.SdkPath)
			continue
		}

		if keyParts[0] == "menu" {
			if len(keyParts) == 2 {
				// This is a menu title, e.g. 'menu.baud=Upload Speed'
				menuTitles[keyParts[1]] = value
			}
		} else if keyParts[1] == "menu" {
			if len(keyParts) >= 4 {
				if len(keyParts) == 4 {
					currentBoard.Menu.parse(keyParts[2], keyParts[3], "", value)
				} else {
					currentBoard.Menu.parse(keyParts[2], keyParts[3], strings.Join(keyParts[4:], "."), value)
				}
			}
		} else {
			key = strings.TrimPrefix(key, currentBoard.Name)
			key = strings.TrimPrefix(key, ".") // Remove the leading dot if present
			currentBoard.Vars.Set(key, value)
		}
	}

	if currentBoard != nil {
		t.ListOfBoards = append(t.ListOfBoards, currentBoard)
	}

	// Create a map of board names to their index in the boards slice
	for i, board := range t.ListOfBoards {
		t.BoardNameToIndex[strings.ToLower(board.Name)] = i
	}

	if fileInfo, err := os.Stat(boardsFile); err == nil {
		t.BoardsFileTime = fileInfo.ModTime().Unix()
	}

	return nil
}

func parseCmdAndArgs(cmdline string, removeEmptyEntries bool) (string, []string) {
	args := parseArgs(cmdline, removeEmptyEntries)
	if len(args) == 0 {
		return "", []string{}
	}
	return args[0], args[1:]
}

func parseArgs(cmdline string, removeEmptyEntries bool) []string {
	var args []string

	// Split the cmdline into arguments by ' ', taking care of quoted strings and brackets
	// Couple of rules:
	// - When encountering a '{'/'[, we ignore current state and look for the matching '}'/']'
	// - When encountering a "/'/`, we ignore current state and look for the matching "/'/`, but the above rule still applies
	state := make([]rune, 0)

	b := 0
	for e, c := range cmdline {
		if c == ' ' && len(state) == 0 {
			segment := strings.TrimSpace(cmdline[b:e])
			if len(segment) > 0 || !removeEmptyEntries {
				segment = corepkg.StrTrimDelimiters(segment, '"')
				//segment = corepkg.StrTrimDelimiters(segment, '\'')
				args = append(args, segment)
			}
			b = e + 1
		} else if c == '\\' {
			// Escape character, skip the next character
			e++
		} else {
			if len(state) == 0 {
				switch c {
				case '"', '\'':
					state = append(state, c)
				case '{':
					state = append(state, '}')
				case '[':
					state = append(state, ']')
				}
			} else {
				top := state[len(state)-1]
				if top == c {
					state = state[:len(state)-1]
				} else {
					switch c {
					case '"', '\'':
						state = append(state, c)
					case '{':
						state = append(state, '}')
					case '[':
						state = append(state, ']')
					}
				}
			}
		}
	}

	// Add the last segment
	segment := strings.TrimSpace(cmdline[b:])
	if len(segment) > 0 || !removeEmptyEntries {
		segment = corepkg.StrTrimDelimiters(segment, '"')
		//segment = corepkg.StrTrimDelimiters(segment, '\'')
		args = append(args, segment)
	}

	return args
}

func (t *toolchain) parsePlatformFile(platformFile string) error {
	file, err := os.OpenFile(platformFile, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	//createBootloader := t.Platform.Recipes["recipe.hooks.prebuild.4"]
	//createBootloader.Cmd = `"{tools.esptool_py.path}\{tools.esptool_py.cmd}" {recipe.hooks.prebuild.4.pattern_args} "{build.path}\{build.project_name}.bootloader.bin" "{compiler.sdk.path}\bin\bootloader_{build.boot}_{build.boot_freq}.elf"`
	//createBootloader.Cmd, createBootloader.CmdLine = ParseCmdAndArgs(createBootloader.Cmd)

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
			if runtime.GOOS != "windows" {
				continue
			}
			key = strings.TrimSuffix(key, ".windows")
			keyParts = keyParts[:len(keyParts)-1]
		}

		ignoreAsPlatformVar := false

		if len(keyParts) >= 3 && keyParts[0] == "recipe" {
			ignoreAsPlatformVar = true

			var ok bool
			var recipe *platformRecipe
			if recipe, ok = t.Platform.Recipes[key]; ok {
				recipe.Name = key
				recipe.Cmd = value
			} else {
				recipe = newPlatformRecipe()
				recipe.Name = key
				recipe.Cmd = value
				t.Platform.Recipes[recipe.Name] = recipe
			}

		} else if keyParts[0] == "tools" {
			toolName := keyParts[1] // e.g. 'esptool', 'esptool_py', 'mkspiffs', 'mklittlefs'

			// Check if the tool already exists
			tool, exists := t.Platform.Tools[toolName]
			if !exists {
				tool = newTool()
				tool.Name = toolName
				t.Platform.Tools[toolName] = tool
			}

			if keyParts[2] == "path" || keyParts[2] == "cmd" {
				// TODO still can have '.windows'
				tool.Vars[keyParts[2]] = value
			} else {
				isFunction := keyParts[2] == "upload" || keyParts[2] == "program" || keyParts[2] == "erase" || keyParts[2] == "bootloader"
				if isFunction {
					toolFunction := keyParts[2] // e.g. 'upload'

					// Check if the function already exists
					function, exists := tool.Functions[toolFunction]
					if !exists {
						function = newToolFunction()
						function.Function = toolFunction
						tool.Functions[toolFunction] = function
					}
					// Now we can set the variable based on the toolVar
					if keyParts[len(keyParts)-1] == "pattern" {
						function.Cmd, function.Args = parseCmdAndArgs(value, true)
					} else if keyParts[len(keyParts)-1] == "pattern_args" {
						function.Args = append(function.Args, parseArgs(value, true)...)
						function.Vars[strings.Join(keyParts[3:], ".")] = value
					} else {
						function.Vars[strings.Join(keyParts[3:], ".")] = value
					}
				} else {

				}
			}
		}

		if !ignoreAsPlatformVar {
			// Add the variable to the platform variables
			t.Platform.Vars[key] = value
		}
	}

	if fileInfo, err := os.Stat(platformFile); err == nil {
		t.PlatformFileTime = fileInfo.ModTime().Unix()
	}

	return nil
}
