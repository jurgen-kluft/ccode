package axe

import (
	_ "embed"
	"path/filepath"
	"strings"
)

// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------

type MakeGenerator2 struct {
	Workspace     *Workspace
	TargetAbsPath string
	Libraries     []*Project
	Product       *Project
}

func NewMakeGenerator2(ws *Workspace) *MakeGenerator2 {
	g := &MakeGenerator2{
		Workspace: ws,
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

func (g *MakeGenerator2) Generate() error {
	if err := g.generateLibMake(); err != nil {
		return err
	}
	if err := g.generateMainMakefile(); err != nil {
		return err
	}

	for _, lib := range g.Libraries {
		g.generateProjectMakefile(lib, false)
	}
	g.generateProjectMakefile(g.Product, true)

	return nil
}

func (g *MakeGenerator2) generateMainMakefile() error {
	mk := NewLineWriter(IndentModeTabs)
	mk.WriteLine(`# each library make file is in its own directory`)

	// For each library define a variable, 'library'name = 'directory'name
	for _, lib := range g.Libraries {
		mk.WriteLine(`library_`, lib.Name, ` := `, lib.Name)
	}

	mk.NewLine()
	mk.WriteLine(`# this is the main project and we reference the static libraries`)

	// Product is dependend on zero or more libraries
	mk.Write("libraries := ")
	for _, lib := range g.Libraries {
		mk.Write(`$(library_`, lib.Name, `) `)
	}
	mk.NewLine()

	// mk.WriteLine(`app               := cbase_unittest`)
	mk.WriteLine(`app := `, g.Product.Name)

	mk.NewLine()
	mk.Write(`.PHONY: all `)
	for _, cfg := range g.Product.Resolved.Configs.Values {
		mk.Write(` `, strings.ToLower(cfg.String()))
	}
	mk.WriteLine(` clean $(app) $(libraries)`)
	mk.NewLine()

	mk.Write(`all: `)
	for _, cfg := range g.Product.Resolved.Configs.Values {
		mk.Write(` `, strings.ToLower(cfg.String()))
	}
	mk.NewLine()

	mk.NewLine()
	mk.WriteLine(`# here we list all the projects to make the clean target`)
	mk.WriteLine(`clean:`)
	for _, lib := range g.Libraries {
		mk.WriteILine(`+`, `@$(MAKE) --directory=`, lib.Name, ` clean`)
	}
	mk.WriteILine(`+`, `@$(MAKE) --directory=`, g.Product.Name, ` clean`)
	mk.NewLine()

	mk.WriteLine(`# here we list all the projects to make all the targets`)
	for _, cfg := range g.Product.Resolved.Configs.Values {
		mk.WriteLine(strings.ToLower(cfg.String()), `:`)
		for _, lib := range g.Libraries {
			mk.WriteILine(`+`, `@$(MAKE) --directory=`, lib.Name, ` CONFIG=`, strings.ToLower(cfg.String()))
		}
		mk.WriteILine(`+`, `@$(MAKE) --directory=`, g.Product.Name, ` CONFIG=`, strings.ToLower(cfg.String()))
		mk.NewLine()
	}

	mk.WriteLine(`# library -> depends on all other libraries`)
	mk.WriteLine(`$(app): $(libraries)`)

	// Write the dependencies of each library (this is for the build order)
	for _, lib := range g.Libraries {
		if lib.Dependencies.IsEmpty() == false {
			mk.Write(`$(library_`, lib.Name, `):`)
			for _, dep := range lib.Dependencies.Values {
				mk.Write(` $(library_`, dep.Name, `)`)
			}
			mk.NewLine()
		}
	}

	mk.WriteToFile(filepath.Join(g.TargetAbsPath, "Makefile"))
	return nil
}

func (g *MakeGenerator2) generateProjectMakefile(project *Project, isMain bool) {
	mk := NewLineWriter(IndentModeTabs)
	mk.SetTabStops(42, 80, 112)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# This file is generated by ccode axe`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Include Platform Common`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	mk.WriteLine(`include ../makelib/common.mk`)
	mk.NewLine()

	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Include Platform Specifics`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	if g.Workspace.MakeTarget.OSIsMac() {
		mk.WriteLine(`include ../makelib/platform/macos.mk`)
	} else if g.Workspace.MakeTarget.OSIsLinux() {
		mk.WriteLine(`include ../makelib/platform/linux.mk`)
	}
	mk.NewLine()

	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteAlignedLine(`PRODUCT`, TabStop(0), `:= `, project.Name)
	mk.WriteAlignedLine(`PRODUCT_LIB`, TabStop(0), `:= lib`, project.Name)
	mk.WriteAlignedLine(`PRODUCT_DYLIB`, TabStop(0), `:= lib`, project.Name)
	mk.WriteAlignedLine(`PRODUCT_FRAMEWORK`, TabStop(0), `:= `, project.Name)
	mk.WriteAlignedLine(`PRODUCT_FRAMEWORK`, TabStop(0), `:= `, project.Name)

	// All include directories
	rootPath := filepath.Join(g.TargetAbsPath, project.Name)
	for _, cfg := range project.Resolved.Configs.Values {
		mk.WriteAligned(`INCLUDES_CPP_`, strings.ToLower(cfg.String()), TabStop(0), `:=`)
		for _, inc := range cfg.IncludeDirs.Values {
			incpath := inc.RelativeTo(rootPath)
			mk.Write(` -I`, incpath)
		}
		mk.NewLine()
	}

	if g.Workspace.MakeTarget.CompilerIsClang() {
		mk.WriteAlignedLine(`CC`, TabStop(0), `:= clang++`)
	} else if g.Workspace.MakeTarget.CompilerIsGcc() {
		mk.WriteAlignedLine(`CC`, TabStop(0), `:= gcc`)
	}

	// The compiler preprocessor defines
	for _, cfg := range project.Resolved.Configs.Values {
		mk.WriteAligned(`DEFINES_CPP_`, strings.ToLower(cfg.String()), TabStop(0), `:=`)
		for _, define := range cfg.CppDefines.Vars.Values {
			mk.Write(` -D`, define)
		}
		mk.NewLine()
	}

	// The linker library directories per config
	// mk.WriteLine(`LIBS_DEBUG          := -L../ccore/target/make/debug/products/arm64 -lccore`)
	// mk.WriteLine(`LIBS_RELEASE        := -L../ccore/target/make/release/products/arm64 -lccore`)
	for _, cfg := range project.Resolved.Configs.Values {
		mk.WriteAligned(`LIBS_`, strings.ToLower(cfg.String()), TabStop(0), `:=`)
		// Library directories are tight up with how make lib is registering the output directories
		for _, dep := range project.Dependencies.Values {
			mk.Write(` -L../`, dep.Name, `/$(DIR_BUILD_PRODUCTS)`, ` -l`, dep.Name)
		}
		mk.NewLine()
	}

	// compiler warnings per config
	for _, cfg := range project.Resolved.Configs.Values {
		mk.WriteAligned(`FLAGS_WARN_`, strings.ToLower(cfg.String()), TabStop(0), `:=`)
		for _, warn := range cfg.DisableWarning.Vars.Values {
			mk.Write(` `, warn)
		}
		mk.NewLine()
	}

	// compiler flags per config
	for _, cfg := range project.Resolved.Configs.Values {
		mk.WriteAligned(`FLAGS_CPP_`, strings.ToLower(cfg.String()), TabStop(0), `:=`)
		for _, flag := range cfg.CppFlags.Vars.Values {
			mk.Write(` `, flag)
		}
		mk.NewLine()
	}

	mk.WriteAlignedLine(`FLAGS_STD_C`, TabStop(0), `:= c99`)

	switch g.Workspace.Config.CppStd {
	case CppStd11:
		mk.WriteAlignedLine(`FLAGS_STD_CPP`, TabStop(0), `:= c++11`)
	case CppStd14:
		mk.WriteAlignedLine(`FLAGS_STD_CPP`, TabStop(0), `:= c++14`)
	case CppStd17:
		mk.WriteAlignedLine(`FLAGS_STD_CPP`, TabStop(0), `:= c++17`)
	case CppStd20:
		mk.WriteAlignedLine(`FLAGS_STD_CPP`, TabStop(0), `:= c++20`)
	case CppStdLatest:
		mk.WriteAlignedLine(`FLAGS_STD_CPP`, TabStop(0), `:= c++latest`)
	}

	mk.WriteAlignedLine(`FLAGS_OTHER_`, TabStop(0), `:=`)
	mk.WriteAlignedLine(`FLAGS_C`, TabStop(0), `:=`)
	mk.WriteAlignedLine(`FLAGS_M`, TabStop(0), `:= -fobjc-arc`)
	mk.WriteAlignedLine(`FLAGS_MM`, TabStop(0), `:= -fobjc-arc`)

	mk.NewLine()
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()

	g.generateProjectTargets(project, isMain, mk)

	mk.WriteToFile(filepath.Join(g.TargetAbsPath, project.Name, "Makefile"))
}

func (g *MakeGenerator2) generateProjectTargets(project *Project, isMain bool, mk *LineWriter) {

	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Built-in targets`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(``)
	mk.WriteLine(`# Declaration for phony targets, to avoid problems with local files`)
	mk.WriteLine(`.PHONY: all clean debugtest releasetest products`)
	mk.WriteLine(`# Declaration for precious targets, to avoid cleaning of intermediate files`)
	mk.WriteLine(`.PRECIOUS: $(DIR_BUILD_TEMP)%$(PRODUCT)$(EXT_O) $(DIR_BUILD_TEMP)%$(EXT_C)$(EXT_O) $(DIR_BUILD_TEMP)%$(EXT_CPP)$(EXT_O) $(DIR_BUILD_TEMP)%$(EXT_M)$(EXT_O) $(DIR_BUILD_TEMP)%$(EXT_MM)$(EXT_O)`)
	mk.NewLine()

	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Common targets`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(``)
	mk.WriteLine(`# Main Target`)

	mk.Write(`all: `)
	for _, cfg := range g.Product.Resolved.Configs.Values {
		mk.Write(` `, strings.ToLower(cfg.String()))
	}
	mk.NewLine()
	mk.WriteILine(`+`, `@:`)
	mk.NewLine()

	for _, cfg := range g.Product.Resolved.Configs.Values {
		mk.WriteLine(strings.ToLower(cfg.String()), `: \`)
		mk.WriteILine(`+`, `_prepare_build_directories \`)
		if isMain {
			mk.WriteILine(`+`, `$(PRODUCT_LIB)$(EXT_LIB) \`)
			mk.WriteILine(`+`, `$(PRODUCT_DYLIB)$(EXT_DYLIB) \`)
			mk.WriteILine(`+`, `$(PRODUCT_FRAMEWORK)$(EXT_FRAMEWORK)`)
		} else {
			mk.WriteILine(`+`, `$(PRODUCT_LIB)$(EXT_LIB)`)
		}
		mk.NewLine()
	}

	mk.WriteLine(`# Cleans all build files`)
	mk.WriteLine(`clean:`)
	for _, cfg := range g.Product.Resolved.Configs.Values {
		mk.WriteILine(`+`, `@$(MAKE) _clean_`, strings.ToLower(cfg.String()), ` CONFIG=`, strings.ToLower(cfg.String()))
	}
	mk.NewLine()

	mk.WriteLine(`# Cleans config specific files`)
	mk.WriteLine(`_clean_%:`)
	mk.WriteILine(`+`, `@echo -e $(call PRINT,Cleaning,$*,Cleaning all intermediate files)`)
	mk.WriteILine(`+`, `@rm -rf $(DIR_BUILD_TEMP)`)
	mk.WriteILine(`+`, `@echo -e $(call PRINT,Cleaning,$*,Cleaning all product files)`)
	mk.WriteILine(`+`, `@rm -rf $(DIR_BUILD_PRODUCTS)`)
	mk.NewLine()

	mk.WriteLine(`# Generate all the output directories once`)
	mk.WriteLine(`_prepare_build_directories:`)
	mk.WriteILine(`+`, `@echo -e $(call PRINT,preparing directories for,$(PRODUCT))`)
	mk.WriteILine(`+`, `@mkdir -p $(DIR_BUILD_PRODUCTS)`)
	mk.NewLine()

	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Targets`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()

	// We need to collect the directories that are needed for the source file compilation
	// Example: @mkdir -p $(DIR_BUILD_TEMP)ccore/source/main/cpp/
	directories := project.VirtualFolders.GetAllLeafDirectories()
	for _, dir := range directories {
		for _, f := range dir.Files {
			if !f.ExcludedFromBuild && (f.Is_C_or_CPP() || f.Is_ObjC()) {
				path := PathGetRelativeTo(dir.DiskPath, PathParent(project.ProjectAbsPath))
				mk.WriteILine(`+`, `@mkdir -p $(DIR_BUILD_TEMP)`, path)
				break
			}
		}
	}
	mk.NewLine()

	lineEnds := []string{"  \\", ""}

	isObjCFileNotExcluded := func(f *FileEntry) bool { return f.Is_ObjC() && !f.ExcludedFromBuild }
	isObjCppFileNotExcluded := func(f *FileEntry) bool { return f.Is_ObjCpp() && !f.ExcludedFromBuild }
	isCFileNotExcluded := func(f *FileEntry) bool { return f.Is_C() && !f.ExcludedFromBuild }
	isCppFileNotExcluded := func(f *FileEntry) bool { return f.Is_CPP() && !f.ExcludedFromBuild }

	// ---------- Framework       -------------------------------------------------------------------------

	mk.WriteLine(`# Framework target`)
	mk.WriteLine(`$(PRODUCT_FRAMEWORK)$(EXT_FRAMEWORK):     \`)

	project.FileEntries.Enumerate(isCppFileNotExcluded, func(i int, key string, value *FileEntry, last int) {
		path := PathGetRelativeTo(filepath.Join(project.ProjectAbsPath, value.Path), PathParent(project.ProjectAbsPath))
		mk.WriteILine(`+`, `$(DIR_BUILD_TEMP)`, path, `$(EXT_O)`, lineEnds[last])
	})

	mk.WriteILine(`+`, `@rm -rf $@`)
	mk.WriteILine(`+`, `@echo -e $(call PRINT,$(notdir $@),$(TARGET_ARCH),Linking the $(TARGET_ARCH) binary)`)
	mk.WriteILine(`+`, `@$(CC) $(LIBS_$(CONFIG)) $(CC_FLAGS_FRAMEWORK_$(TARGET_ARCH)) $(CC_FLAGS_$(TARGET_ARCH)) $(FLAGS_WARN_$(CONFIG)) -fPIC $(FLAGS_OTHER_$(CONFIG)) $(INCLUDES_CPP_$(CONFIG)) $(DEFINES_CPP_$(CONFIG)) -o $(DIR_BUILD_PRODUCTS)$@ $^`)

	mk.WriteLine(``)

	// ---------- Dynamic Library -------------------------------------------------------------------------

	mk.WriteLine(`# The dynamic library`)
	mk.WriteLine(`$(PRODUCT_DYLIB)$(EXT_DYLIB):     \`)

	// Example: $(DIR_BUILD_TEMP)ccore/source/main/cpp/c_allocator.cpp.o     \
	project.FileEntries.Enumerate(isCppFileNotExcluded, func(i int, key string, value *FileEntry, last int) {
		path := PathGetRelativeTo(filepath.Join(project.ProjectAbsPath, value.Path), PathParent(project.ProjectAbsPath))
		mk.WriteILine(`+`, `$(DIR_BUILD_TEMP)`, path, `$(EXT_O)`, lineEnds[last])
	})

	mk.WriteILine(`+`, `@echo -e $(call PRINT,$(notdir $@),$(TARGET_ARCH),Linking the $(TARGET_ARCH) dynamic binary)`)
	mk.WriteILine(`+`, `@mkdir -p $(DIR_BUILD_PRODUCTS)`)
	mk.WriteILine(`+`, `@$(CC) $(LIBS_$(CONFIG)) $(CC_FLAGS_DYLIB_$(TARGET_ARCH)) $(CC_FLAGS_$(TARGET_ARCH)) -o $(DIR_BUILD_PRODUCTS)$@ $^`)
	mk.WriteLine(``)

	// ---------- Static Library -------------------------------------------------------------------------

	mk.WriteLine(`# The static library`)
	mk.WriteLine(`$(PRODUCT_LIB)$(EXT_LIB):    \`)

	// Example: $(DIR_BUILD_TEMP)ccore/source/main/cpp/c_allocator.cpp.o     \
	project.FileEntries.Enumerate(isCppFileNotExcluded, func(i int, key string, value *FileEntry, last int) {
		path := PathGetRelativeTo(filepath.Join(project.ProjectAbsPath, value.Path), PathParent(project.ProjectAbsPath))
		mk.WriteILine(`+`, `$(DIR_BUILD_TEMP)`, path, `$(EXT_O)`, lineEnds[last])
	})

	mk.WriteILine(`+`, `@echo -e $(call PRINT,$(notdir $@),$(TARGET_ARCH),Linking the $(TARGET_ARCH) static binary)`)
	mk.WriteILine(`+`, `@mkdir -p $(DIR_BUILD_PRODUCTS)`)
	mk.WriteILine(`+`, `@$(AR) $(AR_FLAGS_$(TARGET_ARCH)) $(DIR_BUILD_PRODUCTS)$@ $^`)
	mk.WriteLine(``)

	// ---------- Source Files -------------------------------------------------------------------------

	mk.WriteLine(`# All the source file, object file and dependency file generation`)

	// ----- C

	project.FileEntries.Enumerate(isCFileNotExcluded, func(i int, key string, f *FileEntry, last int) {
		srcfile := PathGetRelativeTo(filepath.Join(project.ProjectAbsPath, f.Path), filepath.Join(g.TargetAbsPath, project.Name))
		buildfile := PathGetRelativeTo(filepath.Join(project.ProjectAbsPath, f.Path), PathParent(project.ProjectAbsPath))
		mk.WriteLine(`-include $(DIR_BUILD_TEMP)`, buildfile+`.d`)
		mk.WriteLine(`$(DIR_BUILD_TEMP)`, buildfile+`.o: `, srcfile)
		mk.WriteILine(`+`, `@echo -e $(call PRINT,compiling C,`, buildfile, `)`)
		mk.WriteILine(`+`, `@$(CC) $(CC_FLAGS_$(_ARCH)) -fPIC -std=$(FLAGS_STD_C) $(FLAGS_C) $(FLAGS_WARN_$(CONFIG)) $(FLAGS_OTHER_$(CONFIG)) $(INCLUDES_CPP_$(CONFIG)) $(DEFINES_CPP_$(CONFIG)) -o $@ -c $< -MT $@ -MMD -MP`)
		mk.NewLine()
	})

	// ----- C++

	project.FileEntries.Enumerate(isCppFileNotExcluded, func(i int, key string, f *FileEntry, last int) {
		srcfile := PathGetRelativeTo(filepath.Join(project.ProjectAbsPath, f.Path), filepath.Join(g.TargetAbsPath, project.Name))
		buildfile := PathGetRelativeTo(filepath.Join(project.ProjectAbsPath, f.Path), PathParent(project.ProjectAbsPath))
		mk.WriteLine(`-include $(DIR_BUILD_TEMP)`, buildfile+`.d`)
		mk.WriteLine(`$(DIR_BUILD_TEMP)`, buildfile+`.o: `, srcfile)
		mk.WriteILine(`+`, `@echo -e $(call PRINT,compiling C++,`, buildfile, `)`)
		mk.WriteILine(`+`, `@$(CC) $(CC_FLAGS_$(_ARCH)) -fPIC -std=$(FLAGS_STD_CPP) $(FLAGS_CPP) $(FLAGS_WARN_$(CONFIG)) $(FLAGS_OTHER_$(CONFIG)) $(INCLUDES_CPP_$(CONFIG)) $(DEFINES_CPP_$(CONFIG)) -o $@ -c $< -MT $@ -MMD -MP`)
		mk.NewLine()
	})

	// ----- Objective-C

	project.FileEntries.Enumerate(isObjCFileNotExcluded, func(i int, key string, f *FileEntry, last int) {
		srcfile := PathGetRelativeTo(filepath.Join(project.ProjectAbsPath, f.Path), filepath.Join(g.TargetAbsPath, project.Name))
		buildfile := PathGetRelativeTo(filepath.Join(project.ProjectAbsPath, f.Path), PathParent(project.ProjectAbsPath))
		mk.WriteLine(`-include $(DIR_BUILD_TEMP)`, buildfile+`.d`)
		mk.WriteLine(`$(DIR_BUILD_TEMP)`, buildfile+`.o: `, srcfile)
		mk.WriteILine(`+`, `@echo -e $(call PRINT,compiling objective-c,`, buildfile, `)`)
		mk.WriteILine(`+`, `@$(CC) $(CC_FLAGS_$(_ARCH)) -fPIC -std=$(FLAGS_STD_C) $(FLAGS_M) $(FLAGS_WARN_$(CONFIG)) $(FLAGS_OTHER_$(CONFIG)) $(INCLUDES_CPP_$(CONFIG)) $(DEFINES_CPP_$(CONFIG)) -o $@ -c $< -MT $@ -MMD -MP`)
		mk.NewLine()
	})

	// ----- Objective-C++

	project.FileEntries.Enumerate(isObjCppFileNotExcluded, func(i int, key string, f *FileEntry, last int) {
		srcfile := PathGetRelativeTo(filepath.Join(project.ProjectAbsPath, f.Path), filepath.Join(g.TargetAbsPath, project.Name))
		buildfile := PathGetRelativeTo(filepath.Join(project.ProjectAbsPath, f.Path), PathParent(project.ProjectAbsPath))
		mk.WriteLine(`-include $(DIR_BUILD_TEMP)`, buildfile+`.d`)
		mk.WriteLine(`$(DIR_BUILD_TEMP)`, buildfile+`.o: `, srcfile)
		mk.WriteILine(`+`, `@echo -e $(call PRINT,compiling objective-c++,`, buildfile, `)`)
		mk.WriteILine(`+`, `@$(CC) $(CC_FLAGS_$(_ARCH)) -fPIC -std=$(FLAGS_STD_CPP) $(FLAGS_MM) $(FLAGS_WARN_$(CONFIG)) $(FLAGS_OTHER_$(CONFIG)) $(INCLUDES_CPP_$(CONFIG)) $(DEFINES_CPP_$(CONFIG)) -o $@ -c $< -MT $@ -MMD -MP`)
		mk.NewLine()
	})

	mk.NewLine()
}

func (g *MakeGenerator2) generateLibMake() error {
	if err := g.generateLibMakeCommonMk(); err != nil {
		return err
	}
	if g.Workspace.MakeTarget.OSIsMac() {
		if err := g.generateLibMakePlatformMac(); err != nil {
			return err
		}
	}
	if g.Workspace.MakeTarget.OSIsLinux() {
		if err := g.generateLibMakePlatformLinux(); err != nil {
			return err
		}
	}
	return nil
}

func (g *MakeGenerator2) generateLibMakeCommonMk() error {
	mk := NewLineWriter(IndentModeTabs)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Build `)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()

	mk.WriteLine(`# Target architecture`)
	mk.WriteLine(`TARGET_ARCH := `, g.Workspace.MakeTarget.ArchAsString())
	mk.NewLine()

	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Commands`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	mk.WriteLine(`MAKE    := make -s`) // make or gmake
	mk.WriteLine(`SHELL   := /bin/bash`)
	mk.NewLine()

	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Paths`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	mk.WriteLine(`# Root build directory (debug or release)`)
	mk.WriteLine(`DIR_BUILD           := target/make/$(CONFIG)/`)

	mk.NewLine()
	mk.WriteLine(`# Relative build directories`)
	mk.WriteLine(`DIR_BUILD_PRODUCTS  := $(DIR_BUILD)products/$(TARGET_ARCH)/`)
	mk.WriteLine(`DIR_BUILD_TEMP      := $(DIR_BUILD)intermediates/$(TARGET_ARCH)/`)
	mk.NewLine()
	mk.WriteLine(`# Erases implicit rules`)
	mk.WriteLine(`.SUFFIXES:`)
	mk.NewLine()
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Display`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	mk.WriteLine(`# Terminal colors`)
	mk.WriteLine(`COLOR_NONE      := "\x1b[0m"`)
	mk.WriteLine(`COLOR_GRAY      := "\x1b[30;01m"`)
	mk.WriteLine(`COLOR_RED       := "\x1b[31;01m"`)
	mk.WriteLine(`COLOR_GREEN     := "\x1b[32;01m"`)
	mk.WriteLine(`COLOR_YELLOW    := "\x1b[33;01m"`)
	mk.WriteLine(`COLOR_BLUE      := "\x1b[34;01m"`)
	mk.WriteLine(`COLOR_PURPLE    := "\x1b[35;01m"`)
	mk.WriteLine(`COLOR_CYAN      := "\x1b[36;01m"`)
	mk.NewLine()
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Functions`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	mk.WriteLine(`# Prints a message about a file`)
	mk.WriteLine(`#`)
	mk.WriteLine(`# @1:   The first message part`)
	mk.WriteLine(`# @2:   The architecture`)
	mk.WriteLine(`# @3:   The file`)
	mk.WriteLine(`PRINT_FILE = $(call PRINT,$(1),$(2),$(subst /,/,$(subst ./,,$(dir $(3))))"$(COLOR_GRAY)"$(notdir $(3))"$(COLOR_NONE)")`)
	mk.NewLine()
	mk.WriteLine(`# Prints a message`)
	mk.WriteLine(`#`)
	mk.WriteLine(`# @1:   The first message part`)
	mk.WriteLine(`# @2:   The architecture`)
	mk.WriteLine(`# @3:   The second message part`)
	mk.WriteLine(`PRINT = "["$(COLOR_GREEN)" $(PRODUCT) "$(COLOR_NONE)"]> $(1) [ "$(COLOR_CYAN)$(CONFIG)" - $(2)"$(COLOR_NONE)" ]: "$(COLOR_YELLOW)"$(3)"$(COLOR_NONE)`)
	mk.NewLine()
	if err := mk.WriteToFile(filepath.Join(g.TargetAbsPath, "makelib", "common.mk")); err != nil {
		return err
	}
	return nil
}

func (g *MakeGenerator2) generateLibMakePlatformMac() error {
	mk := NewLineWriter(IndentModeTabs)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# File suffixes`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	mk.WriteLine(`# File extensions`)
	mk.WriteLine(`EXT_O           := .o`)
	mk.WriteLine(`EXT_LIB         := .a`)
	mk.WriteLine(`EXT_DYLIB       := .dylib`)
	mk.WriteLine(`EXT_FRAMEWORK   :=`)
	mk.WriteLine(`EXT_C           := .c`)
	mk.WriteLine(`EXT_CPP         := .cpp`)
	mk.WriteLine(`EXT_M           := .m`)
	mk.WriteLine(`EXT_MM          := .mm`)
	mk.WriteLine(`EXT_H           := .h`)
	mk.NewLine()
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Tools`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	mk.WriteLine(`LD := ld`)
	mk.WriteLine(`AR := ar`)
	mk.NewLine()
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Commands configuration`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	mk.WriteLine(`# Architecture specific flags for ld`)
	mk.WriteLine(`LD_FLAGS_x86_64             :=`)
	mk.WriteLine(`LD_FLAGS_arm64              :=`)
	mk.NewLine()
	mk.WriteLine(`# Architecture specific flags for ar`)
	mk.WriteLine(`AR_FLAGS_x86_64             := rcs`)
	mk.WriteLine(`AR_FLAGS_arm64              := rcs`)
	mk.NewLine()
	mk.WriteLine(`# Architecture specific flags for the C compiler`)
	mk.WriteLine(`CC_FLAGS_x86_64             := -arch x86_64`)
	mk.WriteLine(`CC_FLAGS_arm64              := -arch arm64`)
	mk.NewLine()
	mk.WriteLine(`# Architecture specific flags for the C compiler when creating a dynamic library`)
	mk.WriteLine(`CC_FLAGS_DYLIB_x86_64       = -dynamiclib -install_name $(PREFIX_DYLIB)$(PRODUCT_DYLIB)$(EXT_DYLIB)`)
	mk.WriteLine(`CC_FLAGS_DYLIB_arm64        = -dynamiclib -install_name $(PREFIX_DYLIB)$(PRODUCT_DYLIB)$(EXT_DYLIB)`)
	mk.NewLine()
	mk.WriteLine(`# Architecture specific flags for the C compiler when creating a Mac OS X framework`)
	mk.WriteLine(`CC_FLAGS_FRAMEWORK_x86_64   = -dynamiclib -install_name $(PREFIX_FRAMEWORK)$(PRODUCT_FRAMEWORK)$(EXT_FRAMEWORK) -single_module -compatibility_version 1 -current_version 1`)
	mk.WriteLine(`CC_FLAGS_FRAMEWORK_arm64    = -dynamiclib -install_name $(PREFIX_FRAMEWORK)$(PRODUCT_FRAMEWORK)$(EXT_FRAMEWORK) -compatibility_version 1 -current_version 1`)
	mk.NewLine()
	if err := mk.WriteToFile(filepath.Join(g.TargetAbsPath, "makelib", "platform", "macos.mk")); err != nil {
		return err
	}
	return nil
}

func (g *MakeGenerator2) generateLibMakePlatformLinux() error {
	mk := NewLineWriter(IndentModeTabs)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# File suffixes`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	mk.WriteLine(`# File extensions`)
	mk.WriteLine(`EXT_O          := .o`)
	mk.WriteLine(`EXT_LIB        := .a`)
	mk.WriteLine(`EXT_DYLIB      := .so`)
	mk.WriteLine(`EXT_FRAMEWORK  :=`)
	mk.WriteLine(`EXT_C          := .c`)
	mk.WriteLine(`EXT_CPP        := .cpp`)
	mk.WriteLine(`EXT_M          := .m`)
	mk.WriteLine(`EXT_MM         := .mm`)
	mk.WriteLine(`EXT_H          := .h`)
	mk.NewLine()
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Tools`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	mk.WriteLine(`LD := ld`)
	mk.WriteLine(`AR := ar`)
	mk.NewLine()
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.WriteLine(`# Commands configuration`)
	mk.WriteLine(`#-------------------------------------------------------------------------------`)
	mk.NewLine()
	mk.WriteLine(`# Architecture specific flags for ld`)
	mk.WriteLine(`# LD_FLAGS_$(TARGET_ARCH)       := -m elf_$(TARGET_ARCH)`)
	mk.WriteLine(`LD_FLAGS_$(TARGET_ARCH)       := `)
	mk.NewLine()
	mk.WriteLine(`# Architecture specific flags for ar`)
	mk.WriteLine(`AR_FLAGS_$(TARGET_ARCH)       := rcs`)
	mk.NewLine()
	mk.WriteLine(`# Architecture specific flags for the C compiler`)
	mk.WriteLine(`CC_FLAGS_$(TARGET_ARCH)       :=`)
	mk.NewLine()
	mk.WriteLine(`# Architecture specific flags for the C compiler when creating a dynamic library`)
	mk.WriteLine(`CC_FLAGS_DYLIB_$(TARGET_ARCH) = -shared -Wl,-soname,$(PRODUCT_DYLIB)$(EXT_DYLIB)`)
	mk.NewLine()
	if err := mk.WriteToFile(filepath.Join(g.TargetAbsPath, "makelib", "platform", "linux.mk")); err != nil {
		return err
	}
	return nil
}
