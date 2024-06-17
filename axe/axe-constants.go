package axe

type GeneratorType string

const (
	GeneratorMsDev GeneratorType = "msdev"
	GeneratorXcode GeneratorType = "xcode"
)

func (g GeneratorType) String() string {
	return string(g)
}
