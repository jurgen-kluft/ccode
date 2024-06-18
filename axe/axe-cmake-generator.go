package axe

import ()

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

	return g
}

func (g *CMakeGenerator) Generate() {

}
