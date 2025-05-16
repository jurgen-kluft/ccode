package axe

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Vars struct {
	originalKeys    []string // Original key (case sensitive)
	values          []string
	lowerKeyToIndex map[string]int //
}

func newVars() *Vars {
	v := Vars{
		originalKeys:    make([]string, 0),
		lowerKeyToIndex: make(map[string]int),
		values:          make([]string, 0),
	}
	return &v
}

func (v *Vars) Has(key string) bool {
	_, ok := v.lowerKeyToIndex[strings.ToLower(key)]
	return ok
}

func (v *Vars) HasGet(key string) (string, bool) {
	if index, ok := v.lowerKeyToIndex[strings.ToLower(key)]; ok {
		return v.values[index], true
	}
	return "", false
}

func (v *Vars) Get(key string) string {
	if index, ok := v.lowerKeyToIndex[strings.ToLower(key)]; ok {
		return v.values[index]
	}
	return ""
}

func (v *Vars) Set(key, value string) {
	lowerKey := strings.ToLower(key)
	if index, ok := v.lowerKeyToIndex[lowerKey]; ok {
		v.values[index] = value
	} else {
		index := len(v.originalKeys)
		v.lowerKeyToIndex[lowerKey] = index
		v.originalKeys = append(v.originalKeys, key)
		v.values = append(v.values, value)
	}
}

func (v *Vars) AppendToValue(key, value string, sep string) {
	if index, ok := v.lowerKeyToIndex[strings.ToLower(key)]; ok {
		v.values[index] += sep + value
	} else {
		v.Set(key, value)
	}
}

func (v *Vars) getKeysSorted() []string {
	keys := make([]string, 0, len(v.originalKeys))
	for key, _ := range v.lowerKeyToIndex {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (v *Vars) multiCom(match string) string {
	var result []string
	for index, name := range v.originalKeys {
		if matched, _ := regexp.MatchString("^"+match+"$", name); matched {
			result = append(result, v.values[index])
		}
	}
	sort.Strings(result)
	return strings.Join(result, " && \\\n")
}

type TextFile struct {
	lines []string
}

func newTextFile() *TextFile {
	return &TextFile{
		lines: make([]string, 0),
	}
}

func (text *TextFile) defVar(v *Vars, name, _var string) {
	text.writeLn(_var + " ?= " + v.Get(name))
	v.Set(name, "$("+_var+")")
}

func (text *TextFile) writeLn(lines ...string) {
	text.lines = append(text.lines, lines...)
}

func GenerateArduinoMake(espRoot string, ardEspRoot string, boardName string, flashSize []string, osType string, lwipVariant []string) (lines []string, err error) {

	filesToParse, err := filepath.Glob(espRoot + "/*.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	// This is not strictly necessary, but it makes the output deterministic.
	sort.Sort(sort.StringSlice(filesToParse))

	// Remove obvious non related files, e.g. CMakeLists.txt
	i := 0
	for i < len(filesToParse) {
		if filepath.Base(filesToParse[i]) == "CMakeLists.txt" {
			filesToParse = append(filesToParse[:i], filesToParse[i+1:]...)
		} else {
			i++
		}
	}

	// Some defaults
	vars := newVars()
	vars.Set("runtime.platform.path", espRoot)
	vars.Set("includes", "$(C_INCLUDES)")
	vars.Set("runtime.ide.version", "10605")
	vars.Set("runtime.ide.path", espRoot)
	vars.Set("build.arch", "$(UC_CHIP)")
	vars.Set("build.project_name", "$(MAIN_NAME)")
	vars.Set("build.path", "$(BUILD_DIR)")
	vars.Set("build.core.path", "$(BUILD_DIR)")
	vars.Set("object_files", "$^ $(BUILD_INFO_OBJ)")
	vars.Set("archive_file_path", "$(CORE_LIB)")
	vars.Set("build.sslflags", "$(SSL_FLAGS)")
	vars.Set("build.mmuflags", "$(MMU_FLAGS)")
	vars.Set("build.vtable_flags", "$(VTABLE_FLAGS)")
	vars.Set("build.source.path", "$(dir $(SKETCH))")
	vars.Set("build.variant.path", espRoot+"/variants/"+boardName)
	vars.Set("runtime.os", "$(OS)")
	vars.Set("build.fqbn", "generic")
	vars.Set("_id", boardName)

	var boardDefined bool

	var reFlashSize *regexp.Regexp
	if len(flashSize) > 0 {
		reFlashSize = regexp.MustCompile(fmt.Sprintf(`%s.menu.(?:FlashSize|eesz).%s\.`, boardName, flashSize))
	} else {
		return nil, fmt.Errorf("Flash size not defined for %s\n", boardName)
	}

	var reLwIPVariant *regexp.Regexp
	if len(lwipVariant) > 0 {
		reLwIPVariant = regexp.MustCompile(fmt.Sprintf(`%s.menu.(?:LwIPVariant|ip).%s\.`, boardName, lwipVariant))
	} else {
		reLwIPVariant = nil
	}

	reCpuFrequency := regexp.MustCompile(fmt.Sprintf(`%s.menu.CpuFrequency.[^\.]+\.`, boardName))
	reFlashFreq := regexp.MustCompile(fmt.Sprintf(`%s.menu.(?:FlashFreq|xtal).[^\.]+\.`, boardName))
	reUploadSpeed := regexp.MustCompile(fmt.Sprintf(`%s.menu.UploadSpeed.[^\.]+\.`, boardName))
	reBaud := regexp.MustCompile(fmt.Sprintf(`%s.menu.baud.[^\.]+\.`, boardName))
	reResetMethod := regexp.MustCompile(fmt.Sprintf(`%s.menu.ResetMethod.[^\.]+\.`, boardName))
	reFlashMode := regexp.MustCompile(fmt.Sprintf(`%s.menu.FlashMode.[^\.]+\.`, boardName))
	rePartitionScheme := regexp.MustCompile(fmt.Sprintf(`%s.menu.PartitionScheme.[^\.]+\.`, boardName))

	boardNameDotMenuDot := fmt.Sprintf("%s.menu.", boardName)
	boardNameDotName := fmt.Sprintf("%s.name", boardName)
	boardNameDot := fmt.Sprintf("%s.", boardName)
	dotOsType := fmt.Sprintf(".%s", osType)

	for _, fn := range filesToParse {
		var f *os.File
		f, err = os.Open(fn)
		if err != nil {
			fmt.Printf("Failed to open: %s\n", fn)
			return
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			line = strings.ReplaceAll(line, ".esptool_py.", ".esptool.")
			if matched, _ := regexp.MatchString(`^(\w[\w\-\.]+)=(.*)`, line); matched {
				parts := strings.SplitN(line, "=", 2)
				key, val := parts[0], parts[1]

				if !boardDefined && strings.HasPrefix(key, boardNameDotName) {
					boardDefined = true
				}

				// Truncation of some variable names is needed
				if strings.HasPrefix(key, boardNameDotMenuDot) {

					// Make sure that specified flash size is always used
					if reFlashSize.MatchString(key) {
						key = reFlashSize.ReplaceAllString(key, "")
						vars.Set(key, val)
					}
					key = reCpuFrequency.ReplaceAllString(key, "")
					key = reFlashFreq.ReplaceAllString(key, "")
					key = reUploadSpeed.ReplaceAllString(key, "")
					key = reBaud.ReplaceAllString(key, "")
					key = reResetMethod.ReplaceAllString(key, "")
					key = reFlashMode.ReplaceAllString(key, "")
					if reLwIPVariant != nil {
						key = reLwIPVariant.ReplaceAllString(key, "")
					}
					key = rePartitionScheme.ReplaceAllString(key, "")
				}

				if strings.HasPrefix(key, boardNameDot) {
					key = strings.TrimPrefix(key, boardNameDot)
					vars.Set(key, val)
				}

				if !vars.Has(key) {
					vars.Set(key, val)
				}

				if strings.HasSuffix(key, dotOsType) {
					// Extract the part before the dot and assign the value to that key.
					beforeDot := strings.TrimSuffix(key, dotOsType)
					vars.Set(beforeDot, vars.Get(key))
				}
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading file:", err)
		}
	}

	// Disable the new options handling as makeEspArduino already has this functionality
	vars.Set("build.opt.flags", "")

	if !vars.Has("runtime.tools.xtensa-esp32-elf-gcc.path") {
		vars.Set("runtime.tools.xtensa-esp32-elf-gcc.path", "{runtime.platform.path}/tools/xtensa-esp32-elf")
	}
	if !vars.Has("runtime.tools.xtensa-esp32s2-elf-gcc.path") {
		vars.Set("runtime.tools.xtensa-esp32s2-elf-gcc.path", "{runtime.platform.path}/tools/xtensa-esp32s2-elf")
	}
	if !vars.Has("runtime.tools.xtensa-esp32s3-elf-gcc.path") {
		vars.Set("runtime.tools.xtensa-esp32s3-elf-gcc.path", "{runtime.platform.path}/tools/xtensa-esp32s3-elf")
	}
	if !vars.Has("runtime.tools.riscv32-esp-elf-gcc.path") {
		vars.Set("runtime.tools.riscv32-esp-elf-gcc.path", "{runtime.platform.path}/tools/riscv32-esp-elf")
	}
	if !vars.Has("runtime.tools.xtensa-lx106-elf-gcc.path") {
		vars.Set("runtime.tools.xtensa-lx106-elf-gcc.path", "$(COMP_PATH)")
	}
	if !vars.Has("runtime.tools.python3.path") {
		vars.Set("runtime.tools.python3.path", "$(PYTHON3_PATH)")
	}
	if !vars.Has("upload.resetmethod") {
		vars.Set("upload.resetmethod", "--before default_reset --after hard_reset")
	}

	// fatal if boardDefined is false.
	if !boardDefined {
		return nil, fmt.Errorf("Unknown board %s", boardName)
	}

	text := newTextFile()

	text.writeLn(fmt.Sprintln("# Board definitions"))
	text.defVar(vars, "build.code_debug", "CORE_DEBUG_LEVEL")
	text.defVar(vars, "build.f_cpu", "F_CPU")
	text.defVar(vars, "build.flash_mode", "FLASH_MODE")
	text.defVar(vars, "build.cdc_on_boot", "CDC_ON_BOOT")
	text.defVar(vars, "build.flash_freq", "FLASH_SPEED")

	// Note: This call to defVar will overwrite the default set earlier if it was used.
	text.defVar(vars, "upload.resetmethod", "UPLOAD_RESET")
	text.defVar(vars, "upload.speed", "UPLOAD_SPEED")
	vars.Set("serial.port", "$(UPLOAD_PORT)")

	re := regexp.MustCompile(`\{(cmd|path)\}`)
	vars.Set("tools.esptool.upload.pattern", re.ReplaceAllString(vars.Get("tools.esptool.upload.pattern"), "{tools.esptool.$1}"))

	vars.AppendToValue("compiler.cpreprocessor.flags", "$(C_PRE_PROC_FLAGS)", " ")
	vars.AppendToValue("build.extra_flags", "$(BUILD_EXTRA_FLAGS)", " ")

	// Translate the initial assignment.
	vars.Set("tools.esptool.path", "$(dir $(ESPTOOL_FILE))")

	// Expand all variables
	keys := vars.getKeysSorted()

	// This matches {variable.name} where variable.name can contain letters, numbers, hyphens, and dots.
	varExpansionRegex := regexp.MustCompile(`\{([\w\-\.]+)\}`)

	for _, key := range keys {
		// The loop continues as long as the value contains '{'.
		// Inside the loop, replace the *first* match.
		// We simulate this by finding the first match, replacing it, and repeating until no '{' is left.
		keyValue := vars.Get(key)
		for strings.Contains(keyValue, "{") {
			// Find the first match of {variable.name} in the current value.
			match := varExpansionRegex.FindStringSubmatchIndex(keyValue)
			if match == nil {
				// Should not happen if strings.Contains("{") is true, but as a safeguard.
				break
			}

			// Extract the variable name from the first capture group.
			varName := keyValue[match[2]:match[3]]

			// Look up the variable name in the vars map.
			replacement, ok := vars.HasGet(varName)
			if !ok {
				// If the variable is not found, Perl might treat it as undef or "".
				// We'll replace it with an empty string.
				replacement = ""
			}

			// Manually build the new string by replacing the first matched pattern with the replacement.
			keyValue = keyValue[:match[0]] + replacement + keyValue[match[1]:]
			keyValue = strings.ReplaceAll(keyValue, `""`, "")
		}

		// After the while loop, the value for this key should have no '{' and no '""'.
		vars.Set(key, keyValue)

		// Checks if key is "compiler.path", if the value does NOT exist as a file/directory (-e),
		// and if the value does NOT contain "$(COMP_PATH".
		if key == "compiler.path" {
			compilerPath := vars.Get(key)
			containsCompPath := strings.Contains(compilerPath, "$(COMP_PATH")
			if !containsCompPath && !DirExists(compilerPath) {
				// Replaces the literal string $esp_root with $ard_esp_root.
				compilerPath = strings.ReplaceAll(compilerPath, espRoot, ardEspRoot)
				vars.Set(key, compilerPath)

				// Removes "/bin/" from the end of the string.
				compilerPath = strings.TrimSuffix(compilerPath, "/bin/")
				vars.Set(key, compilerPath)

				pattern := compilerPath + "*/*/bin/"
				matches, err := filepath.Glob(pattern)
				if err != nil {
					// Handle glob error. Assign empty string on error or no match.
					fmt.Fprintf(os.Stderr, "Error performing glob for pattern %q: %v", pattern, err)
					vars.Set(key, "")
				} else if len(matches) > 0 {
					vars.Set(key, matches[0]) // Assign the first match found
				} else {
					vars.Set(key, "") // Assign empty string if no matches found
				}
			}
		}

		// Removes " -o" followed by one or more whitespace characters at the end of the string.
		removeOflagRegex := regexp.MustCompile(` -o\s+$`)
		vars.Set(key, removeOflagRegex.ReplaceAllString(vars.Get(key), ""))

		// Finds patterns like -DVAR="value" and replaces them with -DVAR=\"value\".
		escapeQuotesRegex := regexp.MustCompile(`(-D\w+=)"([^"]+)"`)

		vars.Set(key, escapeQuotesRegex.ReplaceAllString(vars.Get(key), "$1\\\"$2\\\""))
	}

	text.defVar(vars, "compiler.warning_flags", "COMP_WARNINGS")

	// Print the makefile content
	var val string
	text.writeLn(fmt.Sprintf("MCU = %s", vars.Get("build.mcu")))
	text.writeLn(fmt.Sprintf("INCLUDE_VARIANT = %s", vars.Get("build.variant")))
	text.writeLn(fmt.Sprint("VTABLE_FLAGS?=-DVTABLES_IN_FLASH"))
	text.writeLn(fmt.Sprint("MMU_FLAGS?=-DMMU_IRAM_SIZE=0x8000 -DMMU_ICACHE_SIZE=0x8000"))
	text.writeLn(fmt.Sprint("SSL_FLAGS?="))
	text.writeLn(fmt.Sprintf("BOOT_LOADER?=%s/bootloaders/eboot/eboot.elf", espRoot))

	// Commands
	text.writeLn(fmt.Sprintf("C_COM=\\$(C_COM_PREFIX) %s", vars.Get("recipe.c.o.pattern")))
	text.writeLn(fmt.Sprintf("CPP_COM=\\$(CPP_COM_PREFIX) %s", vars.Get("recipe.cpp.o.pattern")))
	text.writeLn(fmt.Sprintf("S_COM=%s", vars.Get("recipe.S.o.pattern")))
	text.writeLn(fmt.Sprintf("LIB_COM=\"%s%s\" %s", vars.Get("compiler.path"), vars.Get("compiler.ar.cmd"), vars.Get("compiler.ar.flags")))
	text.writeLn(fmt.Sprintf("CORE_LIB_COM=%s", vars.Get("recipe.ar.pattern")))
	text.writeLn(fmt.Sprintf("LD_COM=%s", vars.Get("recipe.c.combine.pattern")))
	text.writeLn(fmt.Sprintf("PART_FILE?=%s/tools/partitions/default.csv", espRoot))

	val = vars.Get("recipe.objcopy.eep.pattern")
	if val == "" {
		val = vars.Get("recipe.objcopy.partitions.bin.pattern")
	}

	{
		// Regex to match a quoted string ending in .csv
		re1 := regexp.MustCompile(`\"([^"]+\\.csv)\"`)
		val = re1.ReplaceAllString(val, `$(PART_FILE)`)

		text.writeLn(fmt.Sprintf("GEN_PART_COM=%s", val))

		// First, assign the result of multiCom to val
		val = vars.multiCom(`recipe\.objcopy\.hex.*\.pattern`)

		// Then, perform the regex substitution on val
		// Regex to match a path ending in /bootloaders/eboot/eboot.elf
		re2 := regexp.MustCompile(`[^"]+\/bootloaders\/eboot\/eboot\.elf`)
		val = re2.ReplaceAllString(val, `$(BOOT_LOADER)`)
	}

	// If val is empty, assign the result of the second multiCom call
	if val == "" {
		val = vars.multiCom(`recipe\.objcopy\.bin.*\.pattern`)
	}

	text.writeLn(fmt.Sprintf("OBJCOPY=%s", val))
	text.writeLn(fmt.Sprintf("SIZE_COM=%s", vars.Get("recipe.size.pattern")))
	text.writeLn(fmt.Sprintf("UPLOAD_COM?=%s %s", vars.Get("tools.esptool.upload.pattern"), vars.Get("tools.esptool.upload.pattern_args")))

	if vars.Get("build.spiffs_start") != "" {
		text.writeLn(fmt.Sprintln("SPIFFS_START?=" + vars.Get("build.spiffs_start")))

		spiffsStartStr := vars.Get("build.spiffs_start")
		spiffsEndStr := vars.Get("build.spiffs_end")

		start, _ := strconv.ParseInt(spiffsStartStr, 16, 64) // Ignore error
		end, _ := strconv.ParseInt(spiffsEndStr, 16, 64)     // Ignore error

		size := end - start
		spiffsSize := fmt.Sprintf("0x%X", size) // Format as hex string with 0x prefix

		text.writeLn(fmt.Sprintln("SPIFFS_SIZE?=" + spiffsSize))

		// Check if the map key exists and the value is non-empty
	} else if vars.Get("build.partitions") != "" {
		text.writeLn(fmt.Sprint("COMMA=,"))
		text.writeLn(fmt.Sprint("SPIFFS_SPEC:=$(subst $(COMMA), ,$(shell grep spiffs $(PART_FILE)))"))
		text.writeLn(fmt.Sprint("SPIFFS_START:=$(word 4,$(SPIFFS_SPEC))"))
		text.writeLn(fmt.Sprint("SPIFFS_SIZE:=$(word 5,$(SPIFFS_SPEC))"))
	}

	if vars.Get("build.spiffs_blocksize") == "" {
		vars.Set("build.spiffs_blocksize", "4096")
	}

	text.writeLn(fmt.Sprint("SPIFFS_BLOCK_SIZE?=" + vars.Get("build.spiffs_blocksize")))
	text.writeLn(fmt.Sprint("MK_FS_COM?=\"$(MK_FS_PATH)\" -b $(SPIFFS_BLOCK_SIZE) -s $(SPIFFS_SIZE) -c $(FS_DIR) $(FS_IMAGE)"))
	text.writeLn(fmt.Sprint("RESTORE_FS_COM?=\"$(MK_FS_PATH)\" -b $(SPIFFS_BLOCK_SIZE) -s $(SPIFFS_SIZE) -u $(FS_RESTORE_DIR) $(FS_IMAGE)"))

	fs_upload_com := vars.Get("tools.esptool.upload.pattern") + " " + vars.Get("tools.esptool.upload.pattern_args")

	re1 := regexp.MustCompile(`(.+ -ca) .+`)
	fs_upload_com = re1.ReplaceAllString(fs_upload_com, `$1 $(SPIFFS_START) -cf $(FS_IMAGE)`)
	re2 := regexp.MustCompile(`(.+ --flash_size \S+) .+`)
	fs_upload_com = re2.ReplaceAllString(fs_upload_com, `$1 $(SPIFFS_START) $(FS_IMAGE)`)

	text.writeLn(fmt.Sprintf("FS_UPLOAD_COM?=%s", fs_upload_com))

	val = vars.multiCom(`recipe\.hooks*\.prebuild.*\.pattern`)
	re3 := regexp.MustCompile(`/usr/bin/env `)
	val = re3.ReplaceAllString(val, "")
	re4 := regexp.MustCompile(`bash -c "(.+)"`)
	val = re4.ReplaceAllString(val, `$1`)
	re5 := regexp.MustCompile("(#define .+0x)(`)")
	val = re5.ReplaceAllString(val, `"\\$1\"$2`)
    text.writeLn(fmt.Sprintf("PREBUILD=%s", val))

	text.writeLn(fmt.Sprintf("PRELINK=%s", vars.multiCom(`recipe\.hooks\.linking\.prelink.*\.pattern`)))
	text.writeLn(fmt.Sprintf("MEM_FLASH=%s", vars.Get("recipe.size.regex")))
	text.writeLn(fmt.Sprintf("MEM_RAM=%s", vars.Get("recipe.size.regex.data")))

	if len(flashSize) > 0 {
		flash_info := vars.Get("menu.FlashSize." + flashSize[0])
		if flash_info == "" {
			flash_info = vars.Get("menu.eesz." + flashSize[0])
		}
		text.writeLn(fmt.Sprintf("FLASH_INFO=%s", flash_info))
	} else {
		// default ?
		text.writeLn(fmt.Sprintf("FLASH_INFO=%s", "4MB"))
	}

	if len(lwipVariant) > 0 {
		lwip_info := vars.Get("menu.LwIPVariant." + lwipVariant[0])
		if lwip_info == "" {
			lwip_info = vars.Get("menu.ip." + lwipVariant[0])
		}
		text.writeLn(fmt.Sprintf("LWIP_INFO=%s", lwip_info))
	}

	// for testing purposes, print to stdout
	// for _, line := range text.Lines {
	// 	fmt.Println(line)
	// }

	return text.lines, nil
}
