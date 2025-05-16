package axe

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------

type EspMakeGenerator struct {
	Workspace     *Workspace
	Verbose       bool
	TargetAbsPath string
	Libraries     []*Project
	Product       *Project
}

func NewEspMakeGenerator(ws *Workspace, verbose bool) *EspMakeGenerator {
	g := &EspMakeGenerator{
		Workspace: ws,
		Verbose:   verbose,
	}
	g.TargetAbsPath = ws.GenerateAbsPath

	// Add the libraries
	for _, p := range ws.ProjectList.Values {
		if p.TypeIsLib() || p.TypeIsDll() {
			g.Libraries = append(g.Libraries, p)
		} else if p.TypeIsExe() {
			g.Product = p
		}
	}

	return g
}

func (g *EspMakeGenerator) Generate() error {

	// TODO we do not need to generate this makefile separately, we
	//      can embed it in the main makefile
	mk := NewLineWriter(IndentModeTabs)
	if err := g.generateMakefile(mk); err != nil {
		return err
	}

	if err := mk.WriteToFile(filepath.Join(g.TargetAbsPath, "esp.make")); err != nil {
		return fmt.Errorf("Error writing ESP makefile: %s", err)
	}

	return nil
}

func (g *EspMakeGenerator) generateMakefile(mk *LineWriter) error {

	IS_ESP32 := true
	CPU_NAME := "esp32"
	BOARD_NAME := "esp32"
	CHIP := "esp32"

	//ESP_ROOT := os.Getenv("ESP_ROOT")
	ESP_ROOT := "/Users/obnosis5/sdk/Arduino/esp32"
	ARDUINO_ESP_ROOT := os.Getenv("ARDUINO_ESP_ROOT")

	//var ARDUINO_LIBS string
	var ESP_ARDUINO_VERSION *GitInfo
	var COMP_PATH string
	var MK_FS_MATCH string
	var MK_FS_PATH string
	var PYTHON3_PATH string

	FS_TYPE := "spiffs"
	if FS_TYPE_ENV, ok := os.LookupEnv("FS_TYPE"); ok {
		FS_TYPE = FS_TYPE_ENV
	}
	MK_FS_MATCH = "mk" + FS_TYPE

	if ESP_ROOT == "" || !DirExists(ESP_ROOT) {
		// Esp root not defined, find and use possible version in the Arduino IDE installation
		HOME := os.Getenv("HOME")
		ARDUINO_ROOT := filepath.Join(HOME, ".arduino15")
		if !DirExists(ARDUINO_ROOT) {
			ARDUINO_ROOT = filepath.Join(HOME, "Library", "Arduino15")
			if !DirExists(ARDUINO_ROOT) {
				return fmt.Errorf("No Arduino installation found in %s", ARDUINO_ROOT)
			}
		}

		ARDUINO_ESP_ROOT := filepath.Join(ARDUINO_ROOT, "packages", CHIP)

		var err error
		ESP_ROOT, err = FindDirMatching(filepath.Join(ARDUINO_ESP_ROOT, "hardware", CHIP), func(dir string) bool {
			return true
		})
		if err != nil {
			return fmt.Errorf("No installed version of %s Arduino found", CHIP)
		}

		// Something like '3.2.0'
		ESP_ARDUINO_VERSION = &GitInfo{}
		ESP_ARDUINO_VERSION.Version = filepath.Base(ESP_ROOT)

		// Find used version of compiler and tools
		COMP_PATH = filepath.Join(ARDUINO_ESP_ROOT, "tools")
		COMP_PATH, _ = FindDirMatching(COMP_PATH, func(dir string) bool {
			return strings.HasPrefix(dir, "xtensa-")
		})

		//		COMP_VERSION := ""
		COMP_PATH, _ = FindDirMatching(COMP_PATH, func(dir string) bool {
			//			COMP_VERSION = filepath.Base(dir)
			return true
		})

		// Validate the file system type
		//		MK_FS_VERSION := ""
		MK_FS_PATH, _ = FindDirMatching(filepath.Join(ARDUINO_ESP_ROOT, "tools", MK_FS_MATCH), func(dir string) bool {
			//			MK_FS_VERSION = filepath.Base(dir)
			return true
		})

		PYTHON3_PATH = filepath.Join(ARDUINO_ESP_ROOT, "tools", "python3")
		if DirExists(PYTHON3_PATH) {
			PYTHON3_PATH, _ = FindDirMatching(PYTHON3_PATH, func(dir string) bool {
				return true
			})
		} else {
			PYTHON3_PATH = ""
		}
	} else {
		if !DirExists(ESP_ROOT) {
			return fmt.Errorf("No ESP Arduino found")
		}

		ESP_ARDUINO_VERSION, _ = BuildGitInfo(ESP_ROOT, "git")

		// Validate the file system type
		MK_FS_PATH = filepath.Join(ESP_ROOT, "tools", MK_FS_MATCH)
		if !DirExists(MK_FS_PATH) {
			return fmt.Errorf("ESP Arduino, invalid filesystem %s, path not found!", MK_FS_PATH)
		}

		PYTHON3_PATH = filepath.Join(ESP_ROOT, "tools", "python3")
		if !DirExists(PYTHON3_PATH) {
			PYTHON3_PATH = ""
		}
	}

	ESP_ROOT_ABS, _ := filepath.Abs(ESP_ROOT)

	SKETCH_NAME := g.Product.Name

	mk.WriteLine(`#`)
	mk.WriteLine(`#           WARNING: This file is generated!`)
	mk.WriteLine(`#`)
	mk.WriteLine(`# Makefile for building Arduino projects for ESP8266 and ESP32`)
	mk.WriteLine()
	mk.WriteLine(`__START_TIME := $(shell date +%s)`)
	mk.WriteLine(`__THIS_FILE := `, g.Workspace.GenerateAbsPath+"/esp.make")
	mk.WriteLine(`__TOOLS_DIR := `, g.Workspace.GenerateAbsPath, `/tools`)
	mk.WriteLine()

	NUM_THREADS := fmt.Sprint(runtime.NumCPU())

	// # Operating system specfic settings
	OS_NAME := "linux"
	mk.WriteLine(`# Operating system specfic settings`)
	if runtime.GOOS == "darwin" {
		OS_NAME = "macosx"
		mk.WriteLine(`OS := `, OS_NAME)
		mk.WriteLine(`CONFIG_ROOT ?= $(HOME)/Library`)
		mk.WriteLine(`ARDUINO_ROOT ?= $(HOME)/Library/Arduino15`)
		mk.WriteLine(`UPLOAD_PORT_MATCH ?= /dev/tty.usb*`)
		mk.WriteLine(`CMD_LINE = $(shell ps $$PPID -o command | tail -1)`)
		mk.WriteLine(`OS_NAME = macosx`)
		mk.WriteLine()
	}

	// # Build threads, default is using all
	mk.WriteLine(`# Build threads, default is using all CPU cores`)
	mk.WriteLine(`BUILD_THREADS ?= `, NUM_THREADS)
	mk.WriteLine(`MAKEFLAGS += -j $(BUILD_THREADS)`)
	mk.WriteLine()

	if g.Verbose == false {
		mk.WriteLine(`# Build verbosity, silent by default`)
		mk.WriteLine(`MAKEFLAGS += --silent`)
		mk.WriteLine()
	}

	// # ESP chip family type
	mk.WriteLine(`# ESP chip family type`)
	mk.WriteLine(`CHIP ?= `, CHIP)
	mk.WriteLine(`UC_CHIP := `, strings.ToUpper(CHIP))
	if IS_ESP32 {
		mk.WriteLine(`IS_ESP32 := 1`)
	} else {
		mk.WriteLine(`IS_ESP32 := 0`)
	}
	mk.WriteLine()

	// # Serial flashing parameters
	mk.WriteLine(`# Serial flashing parameters`)
	mk.WriteLine(`UPLOAD_PORT_MATCH ?= /dev/ttyU*`)
	mk.WriteLine(`UPLOAD_PORT ?= $(shell ls -1tr $(UPLOAD_PORT_MATCH) 2>/dev/null | tail -1)`)
	mk.WriteLine()

	// # Monitor definitions
	mk.WriteLine(`# Monitor definitions`)
	mk.WriteLine(`MONITOR_SPEED ?= 115200`)
	mk.WriteLine(`MONITOR_PORT ?= $(UPLOAD_PORT)`)
	mk.WriteLine(`MONITOR_PAR ?= --rts=0 --dtr=0`)
	mk.WriteLine(`MONITOR_COM ?= $(if $(NO_PY_WRAP),python3,$(PY_WRAP)) -m serial.tools.miniterm $(MONITOR_PAR) $(MONITOR_PORT) $(MONITOR_SPEED)`)
	mk.WriteLine()

	// # OTA parameters
	mk.WriteLine(`# OTA parameters`)
	mk.WriteLine(`OTA_ADDR ?=`)
	mk.WriteLine(`OTA_HPORT ?=`)
	mk.WriteLine(`OTA_PORT ?= $(if $(IS_ESP32),3232,8266)`)
	mk.WriteLine(`OTA_PWD ?=`)
	mk.WriteLine(`OTA_ARGS = --progress --ip="$(OTA_ADDR)" --port="$(OTA_PORT)"`)
	mk.WriteLine(`ifneq ($(OTA_HPORT),)`)
	mk.WriteLine(`  OTA_ARGS += --host_port="$(OTA_HPORT)"`)
	mk.WriteLine(`endif`)
	mk.WriteLine(`ifneq ($(OTA_PWD),)`)
	mk.WriteLine(`  OTA_ARGS += --auth="$(OTA_PWD)"`)
	mk.WriteLine(`endif`)
	mk.WriteLine()

	mk.WriteLine(`# HTTP update parameters`)
	mk.WriteLine(`HTTP_ADDR ?=`)
	mk.WriteLine(`HTTP_URI ?= /update`)
	mk.WriteLine(`HTTP_PWD ?= user`)
	mk.WriteLine(`HTTP_USR ?= password`)
	mk.WriteLine(`HTTP_OPT ?= --progress-bar -o /dev/null`)
	mk.WriteLine()

	mk.WriteLine(`# Output directory`)
	mk.WriteLine(`BUILD_ROOT ?= ../target/espmake`)
	BUILD_DIR := g.Workspace.GenerateAbsPath + "/build"
	mk.WriteLine(`BUILD_DIR ?= `, BUILD_DIR)
	mk.WriteLine()

	mk.WriteLine(`# File system and corresponding disk directories`)
	mk.WriteLine(`FS_TYPE ?= `, FS_TYPE)
	mk.WriteLine(`MK_FS_MATCH = `, MK_FS_MATCH)

	mk.WriteLine(`FS_DIR ?= `, g.Workspace.WorkspaceAbsPath, `/data`)
	mk.WriteLine(`FS_RESTORE_DIR ?= $(BUILD_DIR)/file_system`)
	mk.WriteLine()

	mk.WriteLine(`ESP_ROOT := `, ESP_ROOT_ABS)
	mk.WriteLine(`ESP_LIBS = $(ESP_ROOT)/libraries`)
	mk.WriteLine(`SDK_ROOT = $(ESP_ROOT)/tools/sdk`)
	mk.WriteLine(`TOOLS_ROOT = $(ESP_ROOT)/tools`)
	mk.WriteLine()

	mk.WriteLine(`# Location defined, assume that it is a git clone`)
	mk.WriteLine(`ESP_ARDUINO_VERSION = `, ESP_ARDUINO_VERSION.Version)
	mk.WriteLine(`MK_FS_PATH ?= `, MK_FS_PATH)
	mk.WriteLine(`PYTHON3_PATH := `, PYTHON3_PATH)
	mk.WriteLine()

	mk.WriteLine(`# The esp8266 tools directory contains the python3 executable as well as some modules`)
	mk.WriteLine(`# Use these to avoid additional python installation requirements here`)
	mk.WriteLine(`PYTHON3_PATH := $(if $(PYTHON3_PATH),$(PYTHON3_PATH),$(dir $(shell which python3 2>/dev/null)))`)
	mk.WriteLine(`PY_WRAP = $(PYTHON3_PATH)/python3 $(__TOOLS_DIR)/py_wrap.py $(TOOLS_ROOT)`)
	mk.WriteLine(`NO_PY_WRAP ?= $(if $(IS_ESP32),1,)`)
	mk.WriteLine()

	// # Set possible default board variant and validate
	if result, err := BoardOp(filepath.Join(ESP_ROOT, "boards.txt"), CPU_NAME, BOARD_NAME, "check"); err != nil && len(result) == 0 {
		return fmt.Errorf("Invalid board: %s", err)
	}

	mk.WriteLine(`# Handle esptool variants`)
	mk.WriteLine(`MCU ?= $(CHIP)`)
	mk.WriteLine(`ESPTOOL_FILE = $(firstword $(wildcard $(ESP_ROOT)/tools/esptool/esptool.py) \`)
	mk.WriteLine(`                           $(wildcard $(ARDUINO_ESP_ROOT)/tools/esptool_py/*/esptool.py) \`)
	mk.WriteLine(`                           $(ESP_ROOT)/tools/esptool/esptool)`)
	mk.WriteLine(`ESPTOOL ?= $(if $(NO_PY_WRAP),$(ESPTOOL_FILE),$(PY_WRAP) esptool)`)
	mk.WriteLine(`ESPTOOL_COM ?= $(ESPTOOL) --baud=$(UPLOAD_SPEED) --port $(UPLOAD_PORT) --chip $(MCU)`)
	mk.WriteLine()

	if !IS_ESP32 {
		mk.WriteLine(`# esp8266, use esptool directly instead of via tools/upload.py in order to avoid speed restrictions currently implied there`)
		mk.WriteLine(`# UPLOAD_COM = $(ESPTOOL_COM) $(UPLOAD_RESET) write_flash 0x00000 $(BUILD_DIR)/$(MAIN_NAME).bin`)
		mk.WriteLine(`UPLOAD_COM = $(ESPTOOL_COM) $(UPLOAD_RESET) write_flash 0x00000 $(BUILD_DIR)/$(MAIN_NAME).bin`)
		mk.WriteLine(`FS_UPLOAD_COM = $(ESPTOOL_COM) $(UPLOAD_RESET) write_flash $(SPIFFS_START) $(FS_IMAGE)`)
	}
	mk.WriteLine()

	mk.WriteLine(`# Detect if the specified goal involves building or not`)
	mk.WriteLine(`GOALS := $(if $(MAKECMDGOALS),$(MAKECMDGOALS),all)`)
	mk.WriteLine(`BUILDING := $(if $(filter $(GOALS), monitor list_boards list_flash_defs list_lwip set_git_version install help tools_dir preproc info),,1)`)
	mk.WriteLine()

	// # Sketch (main program) selection
	// ifeq ($(BUILDING),)
	//   SKETCH = /dev/null
	// endif
	// ifdef DEMO
	//   SKETCH := $(if $(IS_ESP32),$(ESP_LIBS)/WiFi/examples/WiFiScan/WiFiScan.ino,$(ESP_LIBS)/ESP8266WiFi/examples/WiFiScan/WiFiScan.ino)
	// else
	//   SKETCH ?= $(wildcard *.ino *.pde)
	// endif
	// SKETCH := $(realpath $(wildcard $(SKETCH)))
	// ifeq ($(SKETCH),)
	//   $(error No sketch specified or found. Use "DEMO=1" for testing)
	// endif
	// ifeq ($(wildcard $(SKETCH)),)
	//   $(error Sketch $(SKETCH) not found)
	// endif
	// SRC_GIT_VERSION := $(call git_description,$(dir $(SKETCH)))

	//SKETCH_GIT_INFO, _ := BuildGitInfo(g.Workspace.WorkspaceAbsPath, "git")

	mk.WriteLine(`# Main output definitions`)
	mk.WriteLine(`SKETCH_NAME := `, SKETCH_NAME)
	mk.WriteLine(`MAIN_NAME ?= $(SKETCH_NAME)`)
	mk.WriteLine(`MAIN_EXE ?= $(BUILD_DIR)/$(MAIN_NAME).bin`)
	mk.WriteLine(`FS_IMAGE ?= $(BUILD_DIR)/FS.bin`)
	mk.WriteLine()

	mk.WriteLine(`# Build file extensions`)
	mk.WriteLine(`OBJ_EXT = .o`)
	mk.WriteLine(`DEP_EXT = .d`)
	mk.WriteLine()

	mk.WriteLine(`# Special tool definitions`)
	mk.WriteLine(`OTA_TOOL ?= python3 $(TOOLS_ROOT)/espota.py`)
	mk.WriteLine(`HTTP_TOOL ?= curl`)
	mk.WriteLine()

	mk.WriteLine(`# Core source files`)
	mk.WriteLine(`CORE_DIR = $(ESP_ROOT)/cores/$(CHIP)`)
	// TODO we can collect the core source files here and list them
	mk.WriteLine(`CORE_SRC := $(call find_files,S|c|cpp,$(CORE_DIR))`)
	mk.WriteLine(`CORE_OBJ := $(patsubst %,$(BUILD_DIR)/%$(OBJ_EXT),$(notdir $(CORE_SRC)))`)
	mk.WriteLine(`CORE_LIB = $(BUILD_DIR)/arduino.ar`)
	mk.WriteLine(`USER_OBJ_LIB = $(BUILD_DIR)/user_obj.ar`)
	mk.WriteLine()

	// # Find project specific source files and include directories
	USER_INC_DIRS := make([]string, 0)
	USER_SRC_FILES := make([]string, 0)
	USER_SRC := ""
	USER_LIBS := ""

	for _, proj := range g.Workspace.ProjectList.Values {
		for _, f := range proj.FileEntries.Values {
			path := proj.FileEntries.GetAbsPath(f)
			switch f.Type {
			case FileTypeCSource, FileTypeCppSource:
				USER_SRC_FILES = append(USER_SRC_FILES, path)
				USER_SRC += path + " "
			default:
				// ignore
			}
		}
	}

	USER_INC_DIRS_DONE := make(map[string]bool)

	// Collect all the project include directories as absolute paths
	for _, prj := range g.Workspace.ProjectList.Values {
		for _, config := range prj.Resolved.Configs.Values {
			if config.Type.IsRelease() {
				config.IncludeDirs.Enumerate(func(i int, base, dir string, last int) {
					include := filepath.Join(base, dir)
					if _, ok := USER_INC_DIRS_DONE[include]; !ok {
						if DirExists(include) {
							USER_INC_DIRS = append(USER_INC_DIRS, include)
						}
						USER_INC_DIRS_DONE[include] = true
					}
				})
			}
		}
	}

	mk.WriteLine(`USER_SRC  = `, USER_SRC)
	mk.WriteLine(`USER_LIBS  = `, USER_LIBS)

	mk.WriteLine(`USER_OBJ := $(patsubst %,$(BUILD_DIR)/%$(OBJ_EXT),$(notdir $(USER_SRC)))`)
	mk.WriteLine(`USER_DIRS := $(sort $(dir $(USER_SRC)))`)
	mk.WriteLine()

	FLASH_DEF, _ := BoardOp(filepath.Join(ESP_ROOT, "boards.txt"), CPU_NAME, BOARD_NAME, "first_flash")
	LWIP_VARIANT, _ := BoardOp(filepath.Join(ESP_ROOT, "boards.txt"), CPU_NAME, BOARD_NAME, "first_lwip")

	// Generate the Arduino makefile
	var err error
	var arduinoMakefileContent []string
	if arduinoMakefileContent, err = GenerateArduinoMake(ESP_ROOT, ARDUINO_ESP_ROOT, BOARD_NAME, FLASH_DEF, OS_NAME, LWIP_VARIANT); err != nil {
		return fmt.Errorf("Error parsing Arduino configuration: %s", err)
	}

	// Write the Arduino makefile
	arduinoMake := NewLineWriter(IndentModeSpaces)
	arduinoMake.WriteManyLines(arduinoMakefileContent)
	if err := arduinoMake.WriteToFile(filepath.Join(BUILD_DIR, "arduino.mk")); err != nil {
		return fmt.Errorf("Error writing Arduino makefile: %s", err)
	}

	// This makefile should know where the Arduino makefile is
	mk.WriteLine(`ARDUINO_MK = $(BUILD_DIR)/arduino.mk`)

	mk.WriteLine(`ifneq ($(MAKECMDGOALS),clean)`)
	mk.WriteLine(`-include $(ARDUINO_MK)`)
	mk.WriteLine(`endif`)
	mk.WriteLine()

	// # Compilation directories and path
	INCLUDE_DIRS := make([]string, 0)
	INCLUDE_DIRS = append(INCLUDE_DIRS, "$(CORE_DIR)")
	INCLUDE_DIRS = append(INCLUDE_DIRS, "$(ESP_ROOT)/variants/$(INCLUDE_VARIANT)")
	// For generated header files
	INCLUDE_DIRS = append(INCLUDE_DIRS, BUILD_DIR)

	C_INCLUDES := ""
	for _, dir := range INCLUDE_DIRS {
		C_INCLUDES = C_INCLUDES + fmt.Sprintf("-I%s ", dir)
	}
	for _, dir := range USER_INC_DIRS {
		C_INCLUDES = C_INCLUDES + fmt.Sprintf("-I%s ", dir)
	}
	mk.WriteLine(`C_INCLUDES := `, C_INCLUDES)

	// This is to give make directories to search
	mk.WriteLine(`VPATH += $(shell find $(CORE_DIR) -type d) $(USER_DIRS)`)

	// # Automatically generated build information data source file
	// # Makes the build date and git descriptions at the time of actual build event
	// # available as string constants in the program

	g.GenerateBuildInfo(BUILD_DIR, ESP_ROOT, "buildinfo.h", "buildinfo.c++")
	BUILD_INFO_H := "$(BUILD_DIR)/buildinfo.h"
	BUILD_INFO_CPP := "$(BUILD_DIR)/buildinfo.c++"
	BUILD_INFO_OBJ := "$(BUILD_INFO_CPP)$(OBJ_EXT)"
	mk.WriteLine(`BUILD_INFO_H := `, BUILD_INFO_H)
	mk.WriteLine(`BUILD_INFO_CPP := `, BUILD_INFO_CPP)
	mk.WriteLine(`BUILD_INFO_OBJ := `, BUILD_INFO_OBJ)
	mk.WriteLine()

	// # Use ccache if it is available and not explicitly disabled (USE_CCACHE=0)
	C_COM_PREFIX := ""
	CPP_COM_PREFIX := ""
	mk.WriteLine(`C_COM_PREFIX = `, C_COM_PREFIX)
	mk.WriteLine(`CPP_COM_PREFIX = `, CPP_COM_PREFIX)
	mk.WriteLine()

	// # Generated header files
	// GEN_H_FILES += $(BUILD_INFO_H)
	GEN_H_FILES := make([]string, 0)
	GEN_H_FILES = append(GEN_H_FILES, BUILD_INFO_H)

	// # Special handling needed for esp32 build_opt.h
	// # Golang, copy an existing 'build_opt.h' that may exist in the include directory
	//   of this package
	// ifneq ($(IS_ESP32),)
	// BUILD_OPT_NAME = build_opt.h
	// BUILD_OPT_SRC = $(firstword $(wildcard $(dir $(SKETCH))/$(BUILD_OPT_NAME) /dev/null))
	// BUILD_OPT_DST = $(BUILD_DIR)/$(BUILD_OPT_NAME)
	// GEN_H_FILES += $(BUILD_OPT_DST)
	// $(BUILD_OPT_DST): $(BUILD_OPT_SRC) | $(BUILD_DIR)
	// 	cp $(BUILD_OPT_SRC) $(BUILD_OPT_DST)
	// 	touch $(BUILD_DIR)/file_opts
	// endif
	BUILD_OPT_NAME := "build_opt.h"
	BUILD_OPT_DST := "$(BUILD_DIR)/" + BUILD_OPT_NAME
	if !IS_ESP32 {
		GEN_H_FILES = append(GEN_H_FILES, BUILD_OPT_DST)
	}

	mk.WriteLine(`# Build output root directory`)
	mk.WriteLine(`$(BUILD_DIR):`)
	mk.WriteLine(`	mkdir -p $(BUILD_DIR)`)
	mk.WriteLine()

	mk.WriteLine(`# Build rules for the different source file types`)
	mk.WriteLine(`$(BUILD_DIR)/%.cpp$(OBJ_EXT): %.cpp $(ARDUINO_MK) | $(GEN_H_FILES)`)
	mk.WriteLine(`	@echo  $(<F)`)
	mk.WriteLine(`	$(CPP_COM) $(CPP_EXTRA) $($(<F)_CFLAGS) $(abspath $<) -o $@`)
	mk.WriteLine()

	mk.WriteLine(`$(BUILD_DIR)/%.c$(OBJ_EXT): %.c $(ARDUINO_MK) | $(GEN_H_FILES)`)
	mk.WriteLine(`	@echo  $(<F)`)
	mk.WriteLine(`	$(C_COM) $(C_EXTRA) $($(<F)_CFLAGS) $(abspath $<) -o $@`)
	mk.WriteLine()

	mk.WriteLine(`$(BUILD_DIR)/%.S$(OBJ_EXT): %.S $(ARDUINO_MK) | $(GEN_H_FILES)`)
	mk.WriteLine(`	@echo  $(<F)`)
	mk.WriteLine(`	$(S_COM) $(S_EXTRA) $(abspath $<) -o $@`)
	mk.WriteLine()

	mk.WriteLine(`$(CORE_LIB): $(CORE_OBJ)`)
	mk.WriteLine(`	@echo Creating core archive`)
	mk.WriteLine(`	rm -f $@`)
	mk.WriteLine(`	$(CORE_LIB_COM) $^`)
	mk.WriteLine()

	mk.WriteLine(`$(USER_OBJ_LIB): $(USER_OBJ)`)
	mk.WriteLine(`	@echo Creating object archive`)
	mk.WriteLine(`	rm -f $@`)
	mk.WriteLine(`	$(LIB_COM) $@ $^`)
	mk.WriteLine()

	mk.WriteLine(`# Putting the object files in a library minimizes the memory usage in the executable`)
	mk.WriteLine(`ifneq ($(NO_USER_OBJ_LIB),)`)
	mk.WriteLine(`  USER_OBJ_DEP = $(USER_OBJ)`)
	mk.WriteLine(`else`)
	mk.WriteLine(`  USER_OBJ_DEP = $(USER_OBJ_LIB)`)
	mk.WriteLine(`endif`)
	mk.WriteLine()

	mk.WriteLine(`# Linking the executable`)
	mk.WriteLine(`$(MAIN_EXE): $(CORE_LIB) $(USER_LIBS) $(USER_OBJ_DEP)`)
	mk.WriteLine(`	@echo Linking $(MAIN_EXE)`)
	mk.WriteLine(`	$(PRELINK)`)
	mk.WriteLine(`	@echo "  Versions: $(SRC_GIT_VERSION), $(ESP_ARDUINO_VERSION)"`)
	mk.WriteLine(`	$(CPP_COM) $(BUILD_INFO_CPP) -o $(BUILD_INFO_OBJ)`)
	mk.WriteLine(`	$(LD_COM) $(LD_EXTRA)`)
	mk.WriteLine(`	$(GEN_PART_COM)`)
	mk.WriteLine(`	$(OBJCOPY)`)
	mk.WriteLine(`	$(SIZE_COM) | perl $(__TOOLS_DIR)/mem_use.pl "$(MEM_FLASH)" "$(MEM_RAM)"`)
	if len(LWIP_VARIANT) > 0 {
		mk.WriteLine(`	@printf "LwIPVariant: `, LWIP_VARIANT[0], `"`)
	}
	if len(FLASH_DEF) > 0 {
		mk.WriteLine(`	@printf "Flash size: `, FLASH_DEF[0], `"`)
	}

	mk.WriteLine(`	@perl -e 'print "Build complete. Elapsed time: ", time()-$(__START_TIME),  " seconds\n\n"'`)
	mk.WriteLine()

	mk.WriteLine(`# Flashing operations`)
	mk.WriteLine(`CHECK_PORT := $(if $(UPLOAD_PORT),\`)
	mk.WriteLine(`                   @echo === Using upload port: $(UPLOAD_PORT) @ $(UPLOAD_SPEED),\`)
	mk.WriteLine(`                   @echo "*** Upload port not found or defined" && exit 1)`)
	mk.WriteLine(`upload flash: all`)
	mk.WriteLine(`	$(CHECK_PORT)`)
	mk.WriteLine(`	$(UPLOAD_COM)`)
	mk.WriteLine()

	mk.WriteLine(`ota: all`)
	mk.WriteLine(`ifeq ($(OTA_ADDR),)`)
	mk.WriteLine(`	@echo == Error: Address of device must be specified via OTA_ADDR`)
	mk.WriteLine(`	exit 1`)
	mk.WriteLine(`endif`)
	mk.WriteLine(`	$(OTA_PRE_COM)`)
	mk.WriteLine(`	$(OTA_TOOL) $(OTA_ARGS) --file="$(MAIN_EXE)"`)
	mk.WriteLine()

	mk.WriteLine(`http: all`)
	mk.WriteLine(`ifeq ($(HTTP_ADDR),)`)
	mk.WriteLine(`	@echo == Error: Address of device must be specified via HTTP_ADDR`)
	mk.WriteLine(`	exit 1`)
	mk.WriteLine(`endif`)
	mk.WriteLine(`	$(HTTP_TOOL) $(HTTP_OPT) -F image=@$(MAIN_EXE) --user $(HTTP_USR):$(HTTP_PWD) http://$(HTTP_ADDR)$(HTTP_URI)`)
	mk.WriteLine(`	@echo "\n"`)
	mk.WriteLine()

	mk.WriteLine(`$(FS_IMAGE): $(ARDUINO_MK) $(shell find $(FS_DIR)/ 2>/dev/null)`)
	mk.WriteLine(`ifeq ($(SPIFFS_SIZE),)`)
	mk.WriteLine(`	@echo == Error: No file system specified in FLASH_DEF`)
	mk.WriteLine(`	exit 1`)
	mk.WriteLine(`endif`)
	mk.WriteLine(`	@echo Generating file system image: $(FS_IMAGE)`)
	mk.WriteLine(`	$(MK_FS_COM)`)
	mk.WriteLine()

	mk.WriteLine(`fs: $(FS_IMAGE)`)
	mk.WriteLine()

	mk.WriteLine(`upload_fs flash_fs: $(FS_IMAGE)`)
	mk.WriteLine(`	$(CHECK_PORT)`)
	mk.WriteLine(`	$(FS_UPLOAD_COM)`)
	mk.WriteLine()

	mk.WriteLine(`ota_fs: $(FS_IMAGE)`)
	mk.WriteLine(`ifeq ($(OTA_ADDR),)`)
	mk.WriteLine(`	@echo == Error: Address of device must be specified via OTA_ADDR`)
	mk.WriteLine(`	exit 1`)
	mk.WriteLine(`endif`)
	mk.WriteLine(`	$(OTA_TOOL) $(OTA_ARGS) --spiffs --file="$(FS_IMAGE)"`)
	mk.WriteLine()

	mk.WriteLine(`run: flash`)
	mk.WriteLine(`	$(MONITOR_COM)`)
	mk.WriteLine()

	mk.WriteLine(`monitor:`)
	mk.WriteLine(`ifeq ($(MONITOR_PORT),)`)
	mk.WriteLine(`	@echo "*** Monitor port not found or defined" && exit 1`)
	mk.WriteLine(`endif`)
	mk.WriteLine(`	$(MONITOR_COM)`)
	mk.WriteLine()

	mk.WriteLine(`FLASH_FILE ?= $(BUILD_DIR)/esp_flash.bin`)
	mk.WriteLine(`dump_flash:`)
	mk.WriteLine(`	$(CHECK_PORT)`)
	mk.WriteLine(`	@echo Dumping flash memory to file: $(FLASH_FILE)`)
	mk.WriteLine(`	$(ESPTOOL_COM) read_flash 0 $(shell perl -e 'shift =~ /(\d+)([MK])/ || die "Invalid memory size\n";$$mem_size=$$1*1024;$$mem_size*=1024 if $$2 eq "M";print $$mem_size;' $(FLASH_DEF)) $(FLASH_FILE)`)
	mk.WriteLine()

	mk.WriteLine(`dump_fs:`)
	mk.WriteLine(`	$(CHECK_PORT)`)
	mk.WriteLine(`	@echo Dumping flash file system to directory: $(FS_RESTORE_DIR)`)
	mk.WriteLine(`	-$(ESPTOOL_COM) read_flash $(SPIFFS_START) $(SPIFFS_SIZE) $(FS_IMAGE)`)
	mk.WriteLine(`	mkdir -p $(FS_RESTORE_DIR)`)
	mk.WriteLine(`	@echo`)
	mk.WriteLine(`	@echo == Files ==`)
	mk.WriteLine(`	$(RESTORE_FS_COM)`)
	mk.WriteLine()

	mk.WriteLine(`restore_flash:`)
	mk.WriteLine(`	$(CHECK_PORT)`)
	mk.WriteLine(`	@echo Restoring flash memory from file: $(FLASH_FILE)`)
	mk.WriteLine(`	$(ESPTOOL_COM) $(UPLOAD_RESET) write_flash 0 $(FLASH_FILE)`)
	mk.WriteLine()

	mk.WriteLine(`erase_flash:`)
	mk.WriteLine(`	$(CHECK_PORT)`)
	mk.WriteLine(`	$(ESPTOOL_COM) erase_flash`)
	mk.WriteLine()

	mk.WriteLine(`# Building library instead of executable`)
	mk.WriteLine(`LIB_OUT_FILE ?= $(BUILD_DIR)/$(MAIN_NAME).a`)
	mk.WriteLine(`.PHONY: lib`)
	mk.WriteLine(`lib: $(LIB_OUT_FILE)`)
	mk.WriteLine(`$(LIB_OUT_FILE): $(filter-out $(BUILD_DIR)/$(MAIN_NAME).cpp$(OBJ_EXT),$(USER_OBJ))`)
	mk.WriteLine(`	@echo Building library $(LIB_OUT_FILE)`)
	mk.WriteLine(`	rm -f $(LIB_OUT_FILE)`)
	mk.WriteLine(`	$(LIB_COM) $(LIB_OUT_FILE) $^`)
	mk.WriteLine()

	mk.WriteLine(`# Miscellaneous operations`)
	mk.WriteLine(`clean:`)
	mk.WriteLine(`	@echo Removing all build files`)
	mk.WriteLine(`	rm -rf "$(BUILD_DIR)" $(FILES_TO_CLEAN)`)
	mk.WriteLine()

	BOARD_LIST, _ := BoardOp(filepath.Join(ESP_ROOT_ABS, "boards.txt"), CPU_NAME, BOARD_NAME, "list_boards")
	if len(BOARD_LIST) > 0 {
		mk.WriteLine(`list_boards:`)
		for _, board := range BOARD_LIST {
			mk.WriteLine(`	@echo "`, board, `"`)
		}
	} else {
		mk.WriteLine(`list_boards:`)
		mk.WriteLine(`	@echo none`)
	}
	mk.WriteLine()

	FLASH_LIST, _ := BoardOp(filepath.Join(ESP_ROOT_ABS, "boards.txt"), CPU_NAME, BOARD_NAME, "list_flash")
	if len(FLASH_LIST) > 0 {
		mk.WriteLine(`list_flash_defs:`)
		for _, flash := range FLASH_LIST {
			mk.WriteLine(`	@echo "`, flash, `"`)
		}
	} else {
		mk.WriteLine(`list_flash_defs:`)
		mk.WriteLine(`	@echo none`)
	}
	mk.WriteLine()

	LWIP_LIST, _ := BoardOp(filepath.Join(ESP_ROOT_ABS, "boards.txt"), CPU_NAME, BOARD_NAME, "list_lwip")
	if len(LWIP_LIST) > 0 {
		mk.WriteLine(`list_lwip:`)
		for _, lwip := range LWIP_LIST {
			mk.WriteLine(`	@echo "`, lwip, `"`)
		}
	} else {
		mk.WriteLine(`list_lwip:`)
		mk.WriteLine(`	@echo none`)
	}
	mk.WriteLine()

	mk.WriteLine(`list_lib:`)
	mk.WriteLine(`	@echo Include Directories`)
	for _, dir := range INCLUDE_DIRS {
		mk.WriteLine(fmt.Sprintf("	@echo %s", dir))
	}
	for _, dir := range USER_INC_DIRS {
		mk.WriteLine(fmt.Sprintf("	@echo %s", dir))
	}
	mk.WriteLine(`	@echo Source Files`)
	for _, src := range USER_SRC_FILES {
		mk.WriteLine(fmt.Sprintf("	@echo %s", src))
	}

	mk.WriteLine(`# Just return the path of the tools directory `)
	mk.WriteLine(`tools_dir:`)
	mk.WriteLine(`	@echo $(__TOOLS_DIR)`)
	mk.WriteLine()

	mk.WriteLine(`# Show ram memory usage per variable`)
	mk.WriteLine(`ram_usage: $(MAIN_EXE)`)
	mk.WriteLine(`	$(shell find $(TOOLS_ROOT) | grep 'gcc-nm') -Clrtd --size-sort $(BUILD_DIR)/$(MAIN_NAME).elf | grep -i ' [b] '`)
	mk.WriteLine()

	mk.WriteLine(`# Show ram and flash usage per object files used in the build`)
	mk.WriteLine(`OBJ_INFO_FORM ?= 0`)
	mk.WriteLine(`OBJ_INFO_SORT ?= 1`)
	mk.WriteLine(`obj_info: $(MAIN_EXE)`)
	mk.WriteLine(`	perl $(__TOOLS_DIR)/obj_info.pl "$(shell find $(TOOLS_ROOT) | grep 'elf-size$$')" "$(OBJ_INFO_FORM)" "$(OBJ_INFO_SORT)" $(BUILD_DIR)/*.o`)
	mk.WriteLine()

	mk.WriteLine(`# Analyze crash log`)
	mk.WriteLine(`crash: $(MAIN_EXE)`)
	mk.WriteLine(`	perl $(__TOOLS_DIR)/crash_tool.pl $(ESP_ROOT) $(BUILD_DIR)/$(MAIN_NAME).elf`)
	mk.WriteLine()

	// # Run compiler preprocessor to get full expanded source for a file
	// preproc:
	// ifeq ($(SRC_FILE),)
	// 	$(error SRC_FILE must be defined)
	// endif
	// 	$(CPP_COM) -E $(SRC_FILE)

	mk.WriteLine(`# Main default rule, build the executable`)
	mk.WriteLine(`.PHONY: all`)
	mk.WriteLine(`all: $(BUILD_DIR) $(ARDUINO_MK) prebuild $(MAIN_EXE)`)
	mk.WriteLine()

	mk.WriteLine(`# Prebuild is currently only mandatory for esp32`)
	mk.WriteLine(`USE_PREBUILD ?= $(if $(IS_ESP32),1,)`)
	mk.WriteLine(`prebuild:`)
	mk.WriteLine(`ifneq ($(USE_PREBUILD),)`)
	mk.WriteLine(`	$(PREBUILD)`)
	mk.WriteLine(`endif`)
	mk.WriteLine()

	mk.WriteLine(`# Show installation information`)
	mk.WriteLine(`info:`)
	mk.WriteLine(`	echo == Build info`)
	mk.WriteLine(`	echo "  CHIP:        $(CHIP)"`)
	mk.WriteLine(`	echo "  MCU:         $(MCU)"`)
	mk.WriteLine(`	echo "  ESP_ROOT:    `, ESP_ROOT, `"`)
	mk.WriteLine(`	echo "  Version:     $(ESP_ARDUINO_VERSION)"`)
	mk.WriteLine(`	echo "  Threads:     `, NUM_THREADS, `"`)
	mk.WriteLine(`	echo "  Upload port: $(UPLOAD_PORT)"`)
	mk.WriteLine()

	mk.WriteLine(`# Include all available dependencies from the previous compilation`)
	mk.WriteLine(`-include $(wildcard $(BUILD_DIR)/*$(DEP_EXT))`)
	mk.WriteLine()

	mk.WriteLine(`DEFAULT_GOAL ?= all`)
	mk.WriteLine(`.DEFAULT_GOAL := $(DEFAULT_GOAL)`)
	mk.WriteLine()

	return nil
}

func (g *EspMakeGenerator) GenerateBuildInfo(buildDir string, espRootDir string, headerFile string, sourceFile string) error {
	mk := NewLineWriter(IndentModeTabs)
	mk.WriteLine(`// This file is generated by the build system`)

	BUILD_DATE := time.Now().Format("2006-01-02")
	BUILD_TIME := time.Now().Format("15:04:05")

	gitVersion, _ := BuildGitInfo(g.Workspace.WorkspaceAbsPath, "git")
	gitEspVersion, _ := BuildGitInfo(espRootDir, "git")

	mk.WriteLine(`#include <buildinfo.h>`)
	mk.WriteLine()
	mk.WriteLine(`__BuildInfo_t__ __BuildInfo__ = {`)
	mk.WriteLine(`    "`, BUILD_DATE, `",`)
	mk.WriteLine(`    "`, BUILD_TIME, `",`)
	mk.WriteLine(`    "`, gitVersion.Version, `",`)
	mk.WriteLine(`    "`, gitEspVersion.Version, `"`)
	mk.WriteLine(`};`)
	mk.WriteLine()

	if err := mk.WriteToFile(filepath.Join(buildDir, sourceFile)); err != nil {
		return fmt.Errorf("Error writing build info source file: %s", err)
	}
	mk.Clear()

	mk.WriteLine(`// This file is generated by the build system`)

	mk.WriteLine()
	mk.WriteLine(`typedef struct {`)
	mk.WriteLine(`    const char *date;`)
	mk.WriteLine(`    const char *time;`)
	mk.WriteLine(`    const char *src_version;`)
	mk.WriteLine(`    const char *env_version;`)
	mk.WriteLine(`} __BuildInfo_t__;`)
	mk.WriteLine()

	if err := mk.WriteToFile(filepath.Join(buildDir, headerFile)); err != nil {
		return fmt.Errorf("Error writing build info header file: %s", err)
	}

	return nil
}
