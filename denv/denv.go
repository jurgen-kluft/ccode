package denv

// IDE is an enumeration for all possible IDE's that are supported
type IDE int

const (
	VISUALSTUDIO IDE = 0x80000000
	VS2012       IDE = VISUALSTUDIO | 2012
	VS2013       IDE = VISUALSTUDIO | 2013
	VS2015       IDE = VISUALSTUDIO | 2015
	CODELITE     IDE = 0x70000000
)
