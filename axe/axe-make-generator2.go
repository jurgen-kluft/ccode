package axe

import (
	_ "embed"
	"io"
	"log"
	"os"
)

//go:embed makelib/targets.mk
var makelib_targets_mk string

//go:embed makelib/common.mk
var makelib_common_mk string

//go:embed makelib/platform/osx.mk
var makelib_platform_osx_mk string

//go:embed makelib/platform/linux.mk
var makelib_platform_linux_mk string

// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------

func (g *MakeGenerator2) createMakeLib() {
	g.createMakefile("target/make/makelib/targets.mk", makelib_targets_mk, true)
	g.createMakefile("target/make/makelib/common.mk", makelib_common_mk, true)
	g.createMakefile("target/make/makelib/platform/osx.mk", makelib_platform_osx_mk, true)
	g.createMakefile("target/make/makelib/platform/linux.mk", makelib_platform_linux_mk, true)
}

func (g *MakeGenerator2) createMakefile(filepath string, content string, overwrite bool) {
	_, err := os.Stat(filepath)
	if err == nil && !overwrite {
		return
	}
	f, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = io.WriteString(f, content)
	if err != nil {
		log.Fatal(err)
	}
}

// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------

type MakeGenerator2 struct {
	LastGenId  UUID
	Workspace  *Workspace
	VcxProjCpu string
}

func NewMakeGenerator2(ws *Workspace) *MakeGenerator2 {
	g := &MakeGenerator2{
		LastGenId: GenerateUUID(),
		Workspace: ws,
	}

	return g
}

func (g *MakeGenerator2) Generate() {
	g.createMakeLib()
	g.generateMainMakefile()
}

func (g *MakeGenerator2) generateMainMakefile() {
	/*
	   Main makefile structure (example) for cbase unittest with dependencies on the static libraries cbase, ccore, cunittest where cbase depends on ccore.

	   # each library make file is in its own directory
	   # these are the main project dependencies
	   library_ccore     := ccore
	   library_cbase     := cbase
	   library_cunittest := cunittest

	   # this is the main project and we reference the static libraries
	   libraries         := $(library_ccore) $(library_cbase) $(library_cunittest)
	   app               := cbase_unittest

	   PHONY: all debug release clean $(app) $(libraries)

	   all: debug release

	   # here we list all the projects to make the clean target
	   clean:
	   @$(MAKE) --directory=ccore clean
	   @$(MAKE) --directory=cbase clean
	   @$(MAKE) --directory=cunittest clean
	   @$(MAKE) --directory=cbase_unittest clean

	   # here we list all the projects to make the debug target
	   debug:
	   @$(MAKE) --directory=ccore CONFIG=debug
	   @$(MAKE) --directory=cbase CONFIG=debug
	   @$(MAKE) --directory=cunittest CONFIG=debug
	   @$(MAKE) --directory=cbase_unittest CONFIG=debug

	   # here we list all the projects to make the release target
	   release:
	   @$(MAKE) --directory=ccore CONFIG=release
	   @$(MAKE) --directory=cbase CONFIG=release
	   @$(MAKE) --directory=cunittest CONFIG=release
	   @$(MAKE) --directory=cbase_unittest CONFIG=release

	   # library -> depends on other libraries
	   $(app): $(libraries)
	   $(library_cbase): $(library_ccore)
	*/

	// create the main makefile

}

func (g *MakeGenerator2) generateLibMakefile(project *Project) {

}

func (g *MakeGenerator2) generateExeMakefile(project *Project) {

}
