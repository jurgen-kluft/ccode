# Default target
# CONFIG = all

#-------------------------------------------------------------------------------
# Build type
#-------------------------------------------------------------------------------

# Checks if we're on OS-X to determine the build type
ifeq ($(findstring Darwin, $(shell uname)),)
    BUILD_TARGET  := linux
else
    BUILD_TARGET  := os-x
endif

# Host architecture
HOST_ARCH := $(shell uname -m)

#-------------------------------------------------------------------------------
# Commands
#-------------------------------------------------------------------------------

MAKE    := make -s
SHELL   := /bin/bash

# C/C++ compiler - Debug or Release
ifeq ($(CONFIG),debug)
_CC      = $(CC) $(FLAGS_WARN) -fPIC -$(FLAGS_OPTIM) $(FLAGS_OTHER) $(DIR_INCS) $(CPPDEFINES_DEBUG)  -g
else
_CC      = $(CC) $(FLAGS_WARN) -fPIC -$(FLAGS_OPTIM) $(FLAGS_OTHER) $(DIR_INCS) $(CPPDEFINES_RELEASE)
endif

# Linker - Debug and Release
ifeq ($(CONFIG),debug)
_LD      = $(LD) $(LIBS_DEBUG)
else
_LD      = $(LD) $(LIBS_RELEASE)
endif

#-------------------------------------------------------------------------------
# Tools
#-------------------------------------------------------------------------------

# Make version (version 4 allows parallel builds with output sync)
MAKE_VERSION_MAJOR  := $(shell echo $(MAKE_VERSION) | cut -f1 -d.)
MAKE_4              := $(shell [ $(MAKE_VERSION_MAJOR) -ge 4 ] && echo true)

#-------------------------------------------------------------------------------
# Paths
#-------------------------------------------------------------------------------

# Root build directory (debug or release)
#ifeq ($(findstring 1,$(DEBUG)),)
ifeq ($(CONFIG),debug)
	DIR_BUILD       := target/make/debug/
else
	DIR_BUILD       := target/make/release/
endif

# Relative build directories
DIR_BUILD_PRODUCTS  := $(DIR_BUILD)products/
DIR_BUILD_TEMP      := $(DIR_BUILD)intermediates/
DIR_BUILD_TESTS     := $(DIR)build/tests/

# Erases implicit rules
.SUFFIXES:

#-------------------------------------------------------------------------------
# Display
#-------------------------------------------------------------------------------

# Terminal colors
COLOR_NONE      := "\x1b[0m"
COLOR_GRAY      := "\x1b[30;01m"
COLOR_RED       := "\x1b[31;01m"
COLOR_GREEN     := "\x1b[32;01m"
COLOR_YELLOW    := "\x1b[33;01m"
COLOR_BLUE      := "\x1b[34;01m"
COLOR_PURPLE    := "\x1b[35;01m"
COLOR_CYAN      := "\x1b[36;01m"

#-------------------------------------------------------------------------------
# Functions
#-------------------------------------------------------------------------------

# Gets every C file in a specific source directory
#
# @1:   The directory with the source files
GET_C_FILES = $(foreach _DIR,$(1), $(wildcard $(_DIR)*$(EXT_C)) $(wildcard $(_DIR)**/*$(EXT_C)))

# Gets every C++ file in a specific source directory
#
# @1:   The directory with the source files
GET_CPP_FILES = $(foreach _DIR,$(1), $(wildcard $(_DIR)*$(EXT_CPP)) $(wildcard $(_DIR)**/*$(EXT_CPP)))

# Gets every Objective-C file in a specific source directory
#
# @1:   The directory with the source files
GET_M_FILES = $(foreach _DIR,$(1), $(wildcard $(_DIR)*$(EXT_M)) $(wildcard $(_DIR)**/*$(EXT_M)))

# Gets every Objective-C++ file in a specific source directory
#
# @1:   The directory with the source files
GET_MM_FILES = $(foreach _DIR,$(1), $(wildcard $(_DIR)*$(EXT_MM)) $(wildcard $(_DIR)**/*$(EXT_MM)))

# Gets an SDK value from Xcode
#
# @1:   The key for which to get the SDK value
XCODE_SDK_VALUE = "$(shell /usr/libexec/PlistBuddy -c "Print $(1)" /Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Info.plist)"

# Prints a message about a file
#
# @1:   The first message part
# @2:   The architecture
# @3:   The file
PRINT_FILE = $(call PRINT,$(1),$(2),$(subst /,/,$(subst ./,,$(dir $(3))))"$(COLOR_GRAY)"$(notdir $(3))"$(COLOR_NONE)")

# Prints a message
#
# @1:   The first message part
# @2:   The architecture
# @3:   The second message part
#ifeq ($(findstring 1,$(DEBUG)),)
ifeq ($(CONFIG),debug)
PRINT = "["$(COLOR_GREEN)" $(PRODUCT) "$(COLOR_NONE)"]> $(1) [ "$(COLOR_CYAN)"Debug - $(2)"$(COLOR_NONE)" ]: "$(COLOR_YELLOW)"$(3)"$(COLOR_NONE)
else
PRINT = "["$(COLOR_GREEN)" $(PRODUCT) "$(COLOR_NONE)"]> $(1) [ "$(COLOR_CYAN)"Release - $(2)"$(COLOR_NONE)" ]: "$(COLOR_YELLOW)"$(3)"$(COLOR_NONE)
endif

#-------------------------------------------------------------------------------
# Includes
#-------------------------------------------------------------------------------

__DIR__ := $(dir $(lastword $(MAKEFILE_LIST)))

ifeq ($(BUILD_TARGET),os-x)
    include $(__DIR__)/platform/osx.mk
else
    include $(__DIR__)/platform/linux.mk
endif
