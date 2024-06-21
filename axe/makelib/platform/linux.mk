#-------------------------------------------------------------------------------
# File suffixes
#-------------------------------------------------------------------------------

# File extensions
EXT_O       := .o
EXT_LIB     := .a
EXT_DYLIB   := .so

#-------------------------------------------------------------------------------
# Tools
#-------------------------------------------------------------------------------

LD := ld
AR := ar

#-------------------------------------------------------------------------------
# Commands configuration
#-------------------------------------------------------------------------------

# Architecture specific flags for ld
LD_FLAGS_$(HOST_ARCH)       := -m elf_$(HOST_ARCH)

# Architecture specific flags for ar
AR_FLAGS_$(HOST_ARCH)       := rcs

# Architecture specific flags for the C compiler
CC_FLAGS_$(HOST_ARCH)       :=

# Architecture specific flags for the C compiler when creating a dynamic library
CC_FLAGS_DYLIB_$(HOST_ARCH) = -shared -Wl,-soname,$(PRODUCT_DYLIB)$(EXT_DYLIB)
