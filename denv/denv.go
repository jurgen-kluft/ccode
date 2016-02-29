package denv

// IDE is an enumeration for all possible IDE's that are supported
type IDE int

const (
	cVISUALSTUDIO IDE = 0x80000000
	VS2012        IDE = cVISUALSTUDIO | 2012
	VS2013        IDE = cVISUALSTUDIO | 2013
	VS2015        IDE = cVISUALSTUDIO | 2015
)
