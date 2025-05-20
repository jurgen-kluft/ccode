package denv

// DevLibType defines the type of lib, there are 2 types of libraries, system and user
type DevLibType int

const (
	DevSystemLibrary DevLibType = 1 //
	DevUserLibrary   DevLibType = 2 //
	DevFramework     DevLibType = 4 //
)

type DevLib struct {
	Configs DevConfigType
	Type    DevLibType
	Files   []string
	Libs    []string
	Dir     string
}
