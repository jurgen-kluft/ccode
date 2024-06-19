package axe

type GeneratorType string

const (
	GeneratorMsDev  GeneratorType = "msdev"
	GeneratorTundra GeneratorType = "tundra"
	GeneratorXcode  GeneratorType = "xcode"
)

func (g GeneratorType) String() string {
	return string(g)
}
