package toolchain

type Burner interface {
	SetupBuildArgs(userVars Vars, config string)
	Build() error

	SetupBurnArgs(userVars Vars, config string)
	Burn() error
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Empty Burner

type ToolchainEmptyBurner struct {
	// For some toolchains, the burner is empty as the toolchain does not handle
	// burning to a USB device.
}

// SetupBuildArgs sets up the build arguments for the burner.
func (cl *ToolchainEmptyBurner) SetupBuildArgs(userVars Vars, config string) {
	// Implement the logic to setup build arguments for the burner here
}

// Build builds the project and prepares it for burning.
func (cl *ToolchainEmptyBurner) Build() error {
	// Implement the build logic here
	return nil
}

// SetupBurnArgs sets up the burn arguments for the burner.
func (cl *ToolchainEmptyBurner) SetupBurnArgs(userVars Vars, config string) {
	// Implement the logic to setup burn arguments for the burner here
}

// Burn builds the project and prepares it for burning
func (cl *ToolchainEmptyBurner) Burn() error {
	// Implement the burn logic here
	return nil
}
