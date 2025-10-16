package clay

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
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

type EspressifToolFunction struct {
	Function string        // e.g. upload, program, erase, bootloader
	Cmd      string        // .pattern
	CmdLine  string        //
	Args     []string      // .pattern_args
	Vars     *corepkg.Vars // A map of variables, e.g. 'upload.protocol=serial' or 'upload.params.verbose='
}

func decodeJsonEspressifToolFunction(decoder *corepkg.JsonDecoder) *EspressifToolFunction {
	function := ""
	cmd := ""
	cmdline := ""
	args := make([]string, 0, 10)
	vars := corepkg.NewVars(corepkg.VarsFormatCurlyBraces)
	fields := map[string]corepkg.JsonDecode{
		"function": func(decoder *corepkg.JsonDecoder) { function = decoder.DecodeString() },
		"cmd":      func(decoder *corepkg.JsonDecoder) { cmd = decoder.DecodeString() },
		"cmdline":  func(decoder *corepkg.JsonDecoder) { cmdline = decoder.DecodeString() },
		"args":     func(decoder *corepkg.JsonDecoder) { args = decoder.DecodeStringArray() },
		"vars":     func(decoder *corepkg.JsonDecoder) { vars.DecodeJson(decoder) },
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding Espressif tool function: %s", err.Error())
	}

	return &EspressifToolFunction{
		Function: function,
		Cmd:      cmd,
		CmdLine:  cmdline,
		Args:     args,
		Vars:     vars,
	}
}

func encodeJsonEspressifToolFunction(encoder *corepkg.JsonEncoder, key string, object *EspressifToolFunction) {
	encoder.BeginObject(key)
	{
		encoder.WriteField("function", object.Function)
		encoder.WriteField("cmd", object.Cmd)
		encoder.WriteField("cmdline", object.CmdLine)
		if len(object.Args) > 0 {
			encoder.BeginArray("args")
			{
				for _, arg := range object.Args {
					encoder.WriteArrayElement(arg)
				}
			}
			encoder.EndArray()
		}

		object.Vars.EncodeJson("vars", encoder)
	}
	encoder.EndObject()
}
func newEspressifToolFunction(function string) *EspressifToolFunction {
	return &EspressifToolFunction{
		Function: function,
		Cmd:      "",
		Args:     make([]string, 0),
		Vars:     corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
	}
}

// The following tools are taken:
// - esptool_py
// - esp_ota
type EspressifTool struct {
	Name      string
	Vars      *corepkg.Vars                     // A map of variables, e.g. 'runtime.os' or 'build.path'
	Functions map[string]*EspressifToolFunction // The list of functions for the tool, e.g. upload, program, erase, bootloader
}

func decodeJsonDecodeJsonEspressifTool(decoder *corepkg.JsonDecoder) *EspressifTool {
	name := ""
	vars := corepkg.NewVars(corepkg.VarsFormatCurlyBraces)
	funcs := make(map[string]*EspressifToolFunction)

	fields := map[string]corepkg.JsonDecode{
		"name": func(decoder *corepkg.JsonDecoder) { name = decoder.DecodeString() },
		"vars": func(decoder *corepkg.JsonDecoder) { vars.DecodeJson(decoder) },
		"functions": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				function := decodeJsonEspressifToolFunction(decoder)
				funcs[function.Function] = function
			})
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding Espressif tool: %s", err.Error())
	}

	return &EspressifTool{
		Name:      name,
		Vars:      vars,
		Functions: funcs,
	}
}

func encodeJsonEspressifTool(encoder *corepkg.JsonEncoder, key string, object *EspressifTool) {
	encoder.BeginObject(key)
	{
		encoder.WriteField("name", object.Name)
		object.Vars.EncodeJson("vars", encoder)
		if len(object.Functions) > 0 {
			encoder.BeginArray("functions")
			{
				for _, arg := range object.Functions {
					encodeJsonEspressifToolFunction(encoder, "", arg)
				}
			}
			encoder.EndArray()
		}
	}
	encoder.EndObject()
}

func NewEsp32Tool(name string) *EspressifTool {
	return &EspressifTool{
		Name:      name,
		Vars:      corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
		Functions: make(map[string]*EspressifToolFunction),
	}
}

type EspressifPlatformRecipe struct {
	Name string
	Cmd  string // e.g. C compiler command ('recipe.c.o.pattern')
}

func decodeJsonEspressifPlatformRecipe(decoder *corepkg.JsonDecoder) *EspressifPlatformRecipe {
	recipe := newEspressifPlatformRecipe()
	fields := map[string]corepkg.JsonDecode{
		"name": func(decoder *corepkg.JsonDecoder) { recipe.Name = decoder.DecodeString() },
		"cmd":  func(decoder *corepkg.JsonDecoder) { recipe.Cmd = decoder.DecodeString() },
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding Espressif platform recipe: %s", err.Error())
	}
	return recipe
}

func encodeJsonEspressifPlatformRecipe(encoder *corepkg.JsonEncoder, key string, object *EspressifPlatformRecipe) {
	encoder.BeginObject(key)
	{
		encoder.WriteField("name", object.Name)
		encoder.WriteField("cmd", object.Cmd)
	}
	encoder.EndObject()
}

func newEspressifPlatformRecipe() *EspressifPlatformRecipe {
	return &EspressifPlatformRecipe{
		Name: "",
		Cmd:  "",
	}
}

func (b *EspressifPlatformRecipe) Resolve(boardVars *corepkg.Vars) {
	b.Cmd = ResolveString(b.Cmd, boardVars)
}

type EspressifPlatform struct {
	Name    string                              // The name of the platform
	Version string                              // The version of the platform
	Vars    *corepkg.Vars                       // A map of variables, e.g. 'runtime.os' or 'build.path'
	Recipes map[string]*EspressifPlatformRecipe // The recipes
	Tools   map[string]*EspressifTool           // The list of tools (only 'tools.esptool_py' and 'esp_ota' for now)
}

func decodeJsonEspressifPlatform(decoder *corepkg.JsonDecoder) *EspressifPlatform {
	platform := NewPlatform()

	fields := map[string]corepkg.JsonDecode{
		"name":    func(decoder *corepkg.JsonDecoder) { platform.Name = decoder.DecodeString() },
		"version": func(decoder *corepkg.JsonDecoder) { platform.Version = decoder.DecodeString() },
		"vars":    func(decoder *corepkg.JsonDecoder) { platform.Vars.DecodeJson(decoder) },
		"recipes": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				recipe := decodeJsonEspressifPlatformRecipe(decoder)
				platform.Recipes[recipe.Name] = recipe
			})
		},
		"tools": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				tool := decodeJsonDecodeJsonEspressifTool(decoder)
				platform.Tools[tool.Name] = tool
			})
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding Espressif platform: %s", err.Error())
	}
	return platform
}

func encodeJsonEspressifPlatform(encoder *corepkg.JsonEncoder, key string, object *EspressifPlatform) {
	if object == nil {
		return
	}

	encoder.BeginObject(key)
	{
		encoder.WriteField("name", object.Name)
		encoder.WriteField("version", object.Version)
		object.Vars.EncodeJson("vars", encoder)

		// We write the map as an array, since the objects contain the map key
		if len(object.Recipes) > 0 {
			encoder.BeginArray("recipes")
			{
				for _, v := range object.Recipes {
					encodeJsonEspressifPlatformRecipe(encoder, "", v)
				}
			}
			encoder.EndArray()
		}

		// We write the map as an array, since the objects contain the map key
		if len(object.Tools) > 0 {
			encoder.BeginArray("tools")
			{
				for _, v := range object.Tools {
					encodeJsonEspressifTool(encoder, "", v)
				}
			}
			encoder.EndArray()
		}
	}
	encoder.EndObject()
}

func NewPlatform() *EspressifPlatform {
	platform := &EspressifPlatform{
		Name:    "",
		Version: "",
		Vars:    corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
		Recipes: make(map[string]*EspressifPlatformRecipe),
		Tools:   make(map[string]*EspressifTool),
	}

	return platform
}

type EspressifBoardMenuSubEntry struct {
	Name   string
	Title  string
	Keys   []string
	Values []string
}

func newEspressifBoardMenuSubEntry(name string) *EspressifBoardMenuSubEntry {
	return &EspressifBoardMenuSubEntry{
		Name:   name,
		Title:  "",
		Keys:   make([]string, 0, 10),
		Values: make([]string, 0, 10),
	}
}

func encodeJsonEspressifBoardMenuSubEntry(encoder *corepkg.JsonEncoder, key string, object *EspressifBoardMenuSubEntry) {
	if object == nil {
		return
	}

	encoder.BeginObject(key)
	{
		encoder.WriteField("name", object.Name)
		encoder.WriteField("title", object.Title)
		if len(object.Keys) > 0 {
			encoder.BeginArray("keys")
			{
				for _, key := range object.Keys {
					encoder.WriteArrayElement(key)
				}
			}
			encoder.EndArray()
		}
		if len(object.Values) > 0 {
			encoder.BeginArray("values")
			{
				for _, value := range object.Values {
					encoder.WriteArrayElement(value)
				}
			}
			encoder.EndArray()
		}
	}
	encoder.EndObject()
}

func decodeJsonEspressifBoardMenuSubEntry(decoder *corepkg.JsonDecoder) *EspressifBoardMenuSubEntry {
	name := ""
	title := ""
	keys := make([]string, 0, 10)
	values := make([]string, 0, 10)
	fields := map[string]corepkg.JsonDecode{
		"name":  func(decoder *corepkg.JsonDecoder) { name = decoder.DecodeString() },
		"title": func(decoder *corepkg.JsonDecoder) { title = decoder.DecodeString() },
		"keys":  func(decoder *corepkg.JsonDecoder) { keys = decoder.DecodeStringArray() },
		"values": func(decoder *corepkg.JsonDecoder) {
			values = decoder.DecodeStringArray()
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding Espressif board menu sub entry: %s", err.Error())
	}
	return &EspressifBoardMenuSubEntry{
		Name:   name,
		Title:  title,
		Keys:   keys,
		Values: values,
	}
}

type EspressifBoardMenuEntry struct {
	Name       string // ssl, mmu, FlashFreq, ...
	Current    *EspressifBoardMenuSubEntry
	SubEntries []*EspressifBoardMenuSubEntry
}

func newEspressifBoardMenuEntry(name string) *EspressifBoardMenuEntry {
	return &EspressifBoardMenuEntry{
		Name:       name,
		SubEntries: make([]*EspressifBoardMenuSubEntry, 0),
	}
}

func encodeJsonEspressifBoardMenuEntry(encoder *corepkg.JsonEncoder, key string, object *EspressifBoardMenuEntry) {
	if object == nil {
		return
	}
	encoder.BeginObject(key)
	{
		encoder.WriteField("name", object.Name)
		if len(object.SubEntries) > 0 {
			encoder.BeginArray("entries")
			{
				for _, entry := range object.SubEntries {
					encodeJsonEspressifBoardMenuSubEntry(encoder, "", entry)
				}
			}
			encoder.EndArray()
		}
	}
	encoder.EndObject()
}

func decodeJsonEspressifBoardMenuEntry(decoder *corepkg.JsonDecoder) *EspressifBoardMenuEntry {
	var name string
	var entries []*EspressifBoardMenuSubEntry
	fields := map[string]corepkg.JsonDecode{
		"name": func(decoder *corepkg.JsonDecoder) { name = decoder.DecodeString() },
		"entries": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				entry := decodeJsonEspressifBoardMenuSubEntry(decoder)
				entries = append(entries, entry)
			})
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding Espressif board menu entry: %s", err.Error())
	}
	return &EspressifBoardMenuEntry{
		Name:       name,
		Current:    nil,
		SubEntries: entries,
	}
}

func (e *EspressifBoardMenuEntry) Parse(entryItem string, key string, value string) {
	if e.Current == nil || e.Current.Name != entryItem {
		e.Current = newEspressifBoardMenuSubEntry(entryItem)
		e.SubEntries = append(e.SubEntries, e.Current)
	}
	if len(key) == 0 {
		e.Current.Title = value
	} else {
		e.Current.Keys = append(e.Current.Keys, key)
		e.Current.Values = append(e.Current.Values, value)
	}
}

func (m *EspressifBoardMenuEntry) RegisterVars(vars *corepkg.Vars) {
	if len(m.SubEntries) > 0 {
		keys := m.SubEntries[0].Keys
		values := m.SubEntries[0].Values
		for i, k := range keys {
			vars.Set(k, values[i])
		}
	}
}

type EspressifBoardMenu struct {
	Current *EspressifBoardMenuEntry
	Entries []*EspressifBoardMenuEntry
}

func newEspressifBoardMenu() *EspressifBoardMenu {
	return &EspressifBoardMenu{
		Current: nil,
		Entries: make([]*EspressifBoardMenuEntry, 0),
	}
}

func encodeJsonEspressifBoardMenu(encoder *corepkg.JsonEncoder, key string, object *EspressifBoardMenu) {
	if object == nil {
		return
	}

	encoder.BeginObject(key)
	{
		if len(object.Entries) > 0 {
			encoder.BeginArray("entries")
			{
				for _, entry := range object.Entries {
					encodeJsonEspressifBoardMenuEntry(encoder, "", entry)
				}
			}
			encoder.EndArray()
		}
	}
	encoder.EndObject()
}

func decodeJsonEspressifBoardMenu(decoder *corepkg.JsonDecoder) *EspressifBoardMenu {
	entries := make([]*EspressifBoardMenuEntry, 0)
	fields := map[string]corepkg.JsonDecode{
		"entries": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				entry := decodeJsonEspressifBoardMenuEntry(decoder)
				entries = append(entries, entry)
			})
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding Espressif board menu: %s", err.Error())
	}
	return &EspressifBoardMenu{
		Current: nil,
		Entries: entries,
	}
}

func (m *EspressifBoardMenu) Parse(entry string, subEntry string, key string, value string) {
	if m.Current == nil || m.Current.Name != entry {
		m.Current = newEspressifBoardMenuEntry(entry)
		m.Entries = append(m.Entries, m.Current)
	}
	m.Current.Parse(subEntry, key, value)
}

func (m *EspressifBoardMenu) RegisterVars(vars *corepkg.Vars) {
	// For each menu entry, we register the first variables
	for _, e := range m.Entries {
		e.RegisterVars(vars)
	}
}

type EspressifBoard struct {
	Name        string              // The name of the board
	Description string              // The description of the board
	SdkPath     string              // The path to the SDK
	Menu        *EspressifBoardMenu // Menu options for the board
	Vars        *corepkg.Vars       // Variables for the board
}

func newEspressifBoard(name string, description string) *EspressifBoard {
	return &EspressifBoard{
		Name:        name,
		Description: description,
		SdkPath:     "",
		Menu:        newEspressifBoardMenu(),
		Vars:        corepkg.NewVars(corepkg.VarsFormatCurlyBraces),
	}
}

type EspressifToolchain struct {
	Name             string             // The name of the toolchain
	Version          string             // The version of the toolchain
	SdkPath          string             // The path to the SDK
	ListOfBoards     []*EspressifBoard  // The list of boards
	Platform         *EspressifPlatform // The platform
	BoardNameToIndex map[string]int     // A map of board names to their index in the boards slice
}

func encodeJsonEspressifBoard(encoder *corepkg.JsonEncoder, object *EspressifBoard) {
	if object == nil {
		return
	}
	encoder.BeginObject("")
	{
		encoder.WriteField("name", object.Name)
		encoder.WriteField("descr", object.Description)
		encoder.WriteField("sdk", object.SdkPath)
		encodeJsonEspressifBoardMenu(encoder, "menu", object.Menu)
		object.Vars.EncodeJson("vars", encoder)
	}
	encoder.EndObject()
}

func decodeJsonEspressifBoard(decoder *corepkg.JsonDecoder) *EspressifBoard {
	var name string
	var descr string
	var sdk string
	var menu *EspressifBoardMenu
	vars := corepkg.NewVars(corepkg.VarsFormatCurlyBraces)

	fields := map[string]corepkg.JsonDecode{
		"name":  func(decoder *corepkg.JsonDecoder) { name = decoder.DecodeString() },
		"descr": func(decoder *corepkg.JsonDecoder) { descr = decoder.DecodeString() },
		"sdk":   func(decoder *corepkg.JsonDecoder) { sdk = decoder.DecodeString() },
		"menu":  func(decoder *corepkg.JsonDecoder) { menu = decodeJsonEspressifBoardMenu(decoder) },
		"vars":  func(decoder *corepkg.JsonDecoder) { vars.DecodeJson(decoder) },
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding Espressif board: %s", err.Error())
	}

	return &EspressifBoard{
		Name:        name,
		Description: descr,
		SdkPath:     sdk,
		Menu:        menu,
		Vars:        vars,
	}
}

func decodeJsonEspressifToolchain(toolchain *EspressifToolchain, decoder *corepkg.JsonDecoder) error {
	fields := map[string]corepkg.JsonDecode{
		"name":    func(decoder *corepkg.JsonDecoder) { toolchain.Name = decoder.DecodeString() },
		"version": func(decoder *corepkg.JsonDecoder) { toolchain.Version = decoder.DecodeString() },
		"sdk":     func(decoder *corepkg.JsonDecoder) { toolchain.SdkPath = decoder.DecodeString() },
		"boards": func(decoder *corepkg.JsonDecoder) {
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				board := decodeJsonEspressifBoard(decoder)
				toolchain.ListOfBoards = append(toolchain.ListOfBoards, board)
			})
		},
		"platform": func(decoder *corepkg.JsonDecoder) {
			toolchain.Platform = decodeJsonEspressifPlatform(decoder)
		},
	}
	if err := decoder.Decode(fields); err != nil {
		corepkg.LogErrorf(err, "error decoding Espressif toolchain: %s", err.Error())
	}

	// Construct the BoardNameToIndex map
	for i, board := range toolchain.ListOfBoards {
		toolchain.BoardNameToIndex[strings.ToLower(board.Name)] = i
	}

	return nil
}

func encodeJsonEspressifToolchain(encoder *corepkg.JsonEncoder, key string, object *EspressifToolchain) {
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
				encoder.BeginArray("boards")
				for _, item := range object.ListOfBoards {
					encodeJsonEspressifBoard(encoder, item)
				}
				encoder.EndArray()
			}
		}

		encodeJsonEspressifPlatform(encoder, "platform", object.Platform)
	}
	encoder.EndObject()
}

func NewEspressifToolchain(arch string) *EspressifToolchain {
	if arch == "esp32" {
		espSdkPath := ArduinoEspSdkPath("esp32")
		toolchain := &EspressifToolchain{
			Name:             "Espressif ESP32 Arduino",
			Version:          "3.2.0",
			SdkPath:          espSdkPath,
			ListOfBoards:     make([]*EspressifBoard, 0, 350),
			BoardNameToIndex: make(map[string]int),
			Platform:         NewPlatform(),
		}
		return toolchain
	} else if arch == "esp8266" {

		espSdkPath := ArduinoEspSdkPath("esp8266")
		toolchain := &EspressifToolchain{
			Name:             "Espressif ESP8266 Arduino",
			Version:          "3.2.0",
			SdkPath:          espSdkPath,
			ListOfBoards:     make([]*EspressifBoard, 0, 350),
			BoardNameToIndex: make(map[string]int),
			Platform:         NewPlatform(),
		}
		return toolchain
	}
	return nil
}

func (t *EspressifToolchain) PrintInfo() {
	corepkg.LogInfof("Toolchain: %s, version: %s", t.Name, t.Version)
	corepkg.LogInfof("SDK Path: %s", t.SdkPath)
	corepkg.LogInfof("Number of boards: %d", len(t.ListOfBoards))
	corepkg.LogInfof("Platform: %s", t.Platform.Name)
}

func (t *EspressifToolchain) GetBoardByName(name string) *EspressifBoard {
	if index, ok := t.BoardNameToIndex[strings.ToLower(name)]; ok {
		return t.ListOfBoards[index]
	}
	return nil
}

func ParseToolchainFiles(toolchain *EspressifToolchain) error {
	boardsFilepath := filepath.Join(toolchain.SdkPath, "boards.txt")
	platformFilepath := filepath.Join(toolchain.SdkPath, "platform.txt")

	// Read the boards.txt file
	err := toolchain.ParseBoardsFile(boardsFilepath)
	if err != nil {
		return err
	}

	// Read the platform.txt file
	err = toolchain.ParsePlatformFile(platformFilepath)
	if err != nil {
		return err
	}

	// Print basic info of the things just parsed
	toolchain.PrintInfo()

	return nil
}

func ResolveString(variable string, vars *corepkg.Vars) string {
	type pair struct {
		from int
		to   int
	}

	for true {
		// find [from:to] pairs of variables
		stack := make([]int16, 0, 4)
		list := make([]pair, 0)
		for i, c := range variable {
			switch c {
			case '{':
				current := int16(len(list))
				stack = append(stack, current)
				list = append(list, pair{from: i, to: -1})
			case '}':
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
				corepkg.LogWarningf("Invalid variable pair in string: %s\n", variable)
				return variable // Return the original string if we have an invalid pair
			}
		}

		// resolve the variables, last to first, and assume all pairs are valid and closed
		replaced := 0
		for i := len(list) - 1; i >= 0; i-- {
			p := list[i]
			variableName := variable[p.from+1 : p.to]
			// Check if the variable exists in the vars map
			if vars.Has(variableName) {
				value := vars.GetFirstOrEmpty(variableName)
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

func ResolveStringList(variableList []string, vars *corepkg.Vars) []string {
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

func (t *EspressifToolchain) ResolveVariables(board *EspressifBoard, buildPath string) error {

	boardVars := corepkg.NewVars(corepkg.VarsFormatCurlyBraces)
	boardVars.Set("runtime.os", runtime.GOOS)
	boardVars.Set("runtime.platform.path", t.SdkPath)
	boardVars.Set("runtime.ide.version", "10607")
	boardVars.Set("board.name", board.Name)

	// Add all recipes to the board
	for _, recipe := range t.Platform.Recipes {
		boardVars.Set(recipe.Name, recipe.Cmd)
	}

	boardVars.Join(board.Vars)
	board.Menu.RegisterVars(boardVars)
	boardVars.Join(t.Platform.Vars)

	// For tools we can resolve some of the variables that are local to the tool
	for _, tool := range t.Platform.Tools {
		//tool.Vars = ResolveStringList(tool.Vars, t.Platform.Vars)
		for _, function := range tool.Functions {
			//function.Vars = ResolveStringList(function.Vars, t.Platform.Vars)
			function.Cmd = ResolveString(function.Cmd, t.Platform.Vars)
			function.Args = ResolveStringList(function.Args, t.Platform.Vars)
		}
	}

	board.Vars = boardVars
	return nil
}

func (t *EspressifToolchain) ParseBoardsFile(boardsFile string) error {
	file, err := os.OpenFile(boardsFile, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var currentBoard *EspressifBoard

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
			currentBoard = newEspressifBoard(keyParts[0], value)
			currentBoard.SdkPath = t.SdkPath
			continue
		}

		if keyParts[0] == "menu" {
			if len(keyParts) == 2 {
				// This is a menu title, e.g. 'menu.baud=Upload Speed'
				menuTitles[strings.ToLower(keyParts[1])] = value
			}
		} else if keyParts[1] == "menu" {
			if len(keyParts) >= 4 {
				if len(keyParts) == 4 {
					currentBoard.Menu.Parse(keyParts[2], keyParts[3], "", value)
				} else {
					currentBoard.Menu.Parse(keyParts[2], keyParts[3], strings.Join(keyParts[4:], "."), value)
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

	return nil
}

func parseCmdAndArgs(cmd string) (string, []string) {
	// The arguments follow the first "cmd" part, also the arguments need to be split by ' '
	args := strings.SplitN(cmd, " ", -1)
	cmd = strings.TrimFunc(args[0], func(r rune) bool {
		return r == '"' || r == '\'' || r == ' '
	})
	if len(args) > 1 {
		return cmd, args[1:]
	}
	return cmd, []string{}
}

func parseArgs(args string) []string {
	// The arguments follow the first "cmd" part, also the arguments need to be split by ' '
	argsList := strings.Split(args, " ")
	for i := 0; i < len(argsList); i++ {
		argsList[i] = strings.TrimFunc(argsList[i], func(r rune) bool {
			return r == '"' || r == '\'' || r == ' '
		})
	}
	return argsList
}

func (t *EspressifToolchain) ParsePlatformFile(platformFile string) error {
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
		key = strings.ToLower(key)
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
			var recipe *EspressifPlatformRecipe
			if recipe, ok = t.Platform.Recipes[key]; ok {
				recipe.Name = key
				recipe.Cmd = value
			} else {
				recipe = newEspressifPlatformRecipe()
				recipe.Name = key
				recipe.Cmd = value
				t.Platform.Recipes[recipe.Name] = recipe
			}

		} else if keyParts[0] == "tools" {
			toolName := keyParts[1] // e.g. 'esptool', 'esptool_py', 'mkspiffs', 'mklittlefs'

			// Check if the tool already exists
			tool, exists := t.Platform.Tools[toolName]
			if !exists {
				tool = NewEsp32Tool(toolName)
				t.Platform.Tools[toolName] = tool
			}

			if keyParts[2] == "path" || keyParts[2] == "cmd" {
				// TODO still can have '.windows'
				tool.Vars.Set(keyParts[2], value)
			} else {
				isFunction := keyParts[2] == "upload" || keyParts[2] == "program" || keyParts[2] == "erase" || keyParts[2] == "bootloader"
				if isFunction {
					toolFunction := keyParts[2] // e.g. 'upload'

					// Check if the function already exists
					function, exists := tool.Functions[toolFunction]
					if !exists {
						function = newEspressifToolFunction(toolFunction)
						tool.Functions[toolFunction] = function
					}
					// Now we can set the variable based on the toolVar
					if keyParts[len(keyParts)-1] == "pattern" {
						function.Cmd, function.Args = parseCmdAndArgs(value)
					} else if keyParts[len(keyParts)-1] == "pattern_args" {
						function.Args = append(function.Args, parseArgs(value)...)
						//function.Vars[strings.Join(keyParts[2:], ".")] = function.Args
						function.Vars.Set(strings.Join(keyParts[3:], "."), value)
					} else {
						//function.Vars[strings.Join(keyParts[2:], ".")] = []string{value}
						function.Vars.Set(strings.Join(keyParts[3:], "."), value)
					}
				} else {

				}
			}
		}

		if !ignoreAsPlatformVar {
			// Add the variable to the platform variables
			t.Platform.Vars.Set(key, value)
		}
	}

	return nil
}
