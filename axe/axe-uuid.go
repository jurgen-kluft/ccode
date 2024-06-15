package axe

import (
	"crypto/rand"
	"fmt"
)

type UUID [12]byte

// uniquely identified ID, 96 bits using a 24 hexadecimal text representation
func GenerateUUID() UUID {
	g := UUID{}
	if _, err := rand.Read(g[:]); err != nil {
		panic(err)
	}
	return g
}

func (u UUID) String() string {
	return fmt.Sprintf("%X%X%X", u[0:4], u[4:8], u[8:12])
}
