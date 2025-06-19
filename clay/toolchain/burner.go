package toolchain

// A Burner is an interface that defines the methods required for preparing
// and burning a project to a USB device.
// For example, the Xtensa Espressif toolchain implements this interface
// to prepare the project for burning to an ESP32 device.
type Burner interface {
	SetupBuild(buildPath string)
	Build() error
	SetupBurn(buildPath string) error
	Burn() error
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Empty Burner

type EmptyBurner struct {
	// For some toolchains, the burner is empty, e.g. Darwin, Windows and Linux
}

func (cl *EmptyBurner) SetupBuild(buildPath string) {
}
func (cl *EmptyBurner) Build() error {
	return nil
}
func (cl *EmptyBurner) SetupBurn(buildPath string) error {
	return nil
}
func (cl *EmptyBurner) Burn() error {
	return nil
}
