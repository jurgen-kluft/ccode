package corepkg

import (
	"fmt"
	"math/bits"
)

type UUID uint64

func GenerateUUID() UUID {
	g := UUID(generator.Uint64())
	return g
}

type UUIDRandomGenerator struct {
	seed           uint64
	s0, s1, s2, s3 uint64
}

func NewUUIDRandomGenerator(seed uint64) UUIDRandomGenerator {
	var s0 uint64
	var s1 uint64
	var s2 uint64
	var s3 uint64
	s0, seed = __mix(seed)
	s1, seed = __mix(seed)
	s2, seed = __mix(seed)
	s3, seed = __mix(seed)
	return UUIDRandomGenerator{seed: seed, s0: s0, s1: s1, s2: s2, s3: s3}
}

var generator = &UUIDRandomGenerator{seed: 0xEED1A7370B428D0B, s0: 0x5C6121E067074000, s1: 0x7C6EB4D2F3A8C000, s2: 0x9C1E5D7CDC0C2000, s3: 0xFFD1A6E10C2D3000}

// String returns the UUID as a string depending on the generator type
// Visual Studio uses the format  {XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX}
// Xcode uses another format      XXXXXXXXXXXXXXXX
func (u UUID) ForVisualStudio() string {

	// Random number generation chain depends on what u is
	g := NewUUIDRandomGenerator(uint64(u))
	s1 := g.Uint64()
	s2 := g.Uint64()

	return fmt.Sprintf("{%08X-%04X-%04X-%04X-%08X%04X}", uint32(s1>>32), uint16(s1&0xFFFF), uint16(s1>>16), uint16(s2>>16), uint32(s2>>32), uint16(s2&0xFFFF))
}

func (u UUID) ForXCode() string {

	// Random number generation chain depends on what u is
	g := NewUUIDRandomGenerator(uint64(u))
	s1 := g.Uint64()
	s2 := g.Uint64()

	u1 := uint32(s1 >> 32)
	u2 := uint32(s1 & 0xFFFFFFFF)
	u3 := uint32(s2 >> 32)
	return fmt.Sprintf("%08X%08X%08X", u1, u2, u3)
}

func __mix(_seed uint64) (seed uint64, z uint64) {
	seed = _seed + uint64(0x9E3779B97F4A7C15)
	z = seed
	z = (z ^ (z >> 30)) * uint64(0xBF58476D1CE4E5B9)
	z = (z ^ (z >> 27)) * uint64(0x94D049BB133111EB)
	z = z ^ (z >> 31)
	return
}

func (x *UUIDRandomGenerator) Uint64() uint64 {
	s0, s1, s2, s3 := x.s0, x.s1, x.s2, x.s3
	x.s0 = s0 ^ s3 ^ s1
	x.s1 = s1 ^ s2 ^ s0
	x.s2 = s2 ^ s0 ^ (s1 << 17)
	x.s3 = bits.RotateLeft64(s3^s1, 45)
	return s0 + s3
}
