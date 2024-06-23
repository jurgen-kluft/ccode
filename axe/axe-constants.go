package axe

type GeneratorType string

const (
	GeneratorMsDev  GeneratorType = "msdev"
	GeneratorTundra GeneratorType = "tundra"
	GeneratorMake   GeneratorType = "make"
	GeneratorCMake  GeneratorType = "cmake"
	GeneratorXcode  GeneratorType = "xcode"
)

func (g GeneratorType) String() string {
	return string(g)
}