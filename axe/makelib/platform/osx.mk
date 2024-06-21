#-------------------------------------------------------------------------------
# File suffixes
#-------------------------------------------------------------------------------

# File extensions
EXT_O           := .o
EXT_LIB         := .a
EXT_DYLIB       := .dylib
EXT_FRAMEWORK   :=

#-------------------------------------------------------------------------------
# Tools
#-------------------------------------------------------------------------------

LD := ld
AR := ar

#-------------------------------------------------------------------------------
# Commands configuration
#-------------------------------------------------------------------------------

# Architecture specific flags for ld
LD_FLAGS_i386               :=
LD_FLAGS_x86_64             :=
LD_FLAGS_armv7              :=
LD_FLAGS_armv7s             :=
LD_FLAGS_arm64              :=

# Architecture specific flags for ar
AR_FLAGS_i386               := rcs
AR_FLAGS_x86_64             := rcs
AR_FLAGS_armv7              := rcs
AR_FLAGS_armv7s             := rcs
AR_FLAGS_arm64              := rcs

# Architecture specific flags for the C compiler
CC_FLAGS_i386               := -arch i386
CC_FLAGS_x86_64             := -arch x86_64
CC_FLAGS_armv7              := -arch armv7
CC_FLAGS_armv7s             := -arch armv7s
CC_FLAGS_arm64              := -arch arm64

# Architecture specific flags for the C compiler when creating a dynamic library
CC_FLAGS_DYLIB_i386         = -dynamiclib -install_name $(PREFIX_DYLIB)$(PRODUCT_DYLIB)$(EXT_DYLIB)
CC_FLAGS_DYLIB_x86_64       = -dynamiclib -install_name $(PREFIX_DYLIB)$(PRODUCT_DYLIB)$(EXT_DYLIB)

# Architecture specific flags for the C compiler when creating a Mac OS X framework
CC_FLAGS_FRAMEWORK_i386     = -dynamiclib -install_name $(PREFIX_FRAMEWORK)$(PRODUCT_FRAMEWORK)$(EXT_FRAMEWORK) -single_module -compatibility_version 1 -current_version 1
CC_FLAGS_FRAMEWORK_x86_64   = -dynamiclib -install_name $(PREFIX_FRAMEWORK)$(PRODUCT_FRAMEWORK)$(EXT_FRAMEWORK) -single_module -compatibility_version 1 -current_version 1
