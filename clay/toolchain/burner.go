package toolchain

// A Burner is an interface that defines the methods required for preparing
// and burning a project to a USB device.
// For example, the Xtensa Espressif toolchain implements this interface
// to prepare the project for burning to an ESP32 device.
type Burner interface {
	SetupBuildArgs()
	Build() error
	SetupBurnArgs()
	Burn() error
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Empty Burner

type ToolchainEmptyBurner struct {
	// For some toolchains, the burner is empty as the toolchain does not handle
	// burning to a USB device.
}

func (cl *ToolchainEmptyBurner) SetupBuildArgs() {
}
func (cl *ToolchainEmptyBurner) Build() error {
	return nil
}
func (cl *ToolchainEmptyBurner) SetupBurnArgs() {
}
func (cl *ToolchainEmptyBurner) Burn() error {
	return nil
}
