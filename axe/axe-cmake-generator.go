package axe

import (
	"path/filepath"
)

type CMakeGenerator struct {
	LastGenId  UUID
	Workspace  *Workspace
	VcxProjCpu string
}

func NewCMakeGenerator(ws *Workspace) *CMakeGenerator {
	g := &CMakeGenerator{
		LastGenId: GenerateUUID(),
		Workspace: ws,
	}
	g.init(ws)
	return g
}

func (g *CMakeGenerator) Generate() {

}
