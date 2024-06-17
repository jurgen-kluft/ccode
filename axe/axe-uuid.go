package axe

import (
	"crypto/rand"
	"fmt"
)

type UUID [16]byte

func GenerateUUID() UUID {
	g := UUID{}
	if _, err := rand.Read(g[:]); err != nil {
		panic(err)
	}
	return g
}

// String returns the UUID as a string depending on the generator type
// Visual Studio uses the format  {XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX}
// Xcode uses another format      XXXXXXXXXXXXXXXX
func (u UUID) String(t GeneratorType) string {
	switch t {
	case GeneratorMsDev:
		return fmt.Sprintf("{%08X-%04X-%04X-%04X-%012X}", u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
	case GeneratorXcode:
		return fmt.Sprintf("%X%X%X", u[0:4], u[4:8], u[8:12])
	}
	return "Cannot generate UUID for an unknown generator type"
}
