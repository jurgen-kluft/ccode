package denv

// BuildTarget defines the type of build target
type BuildTarget uint32
type BuildTargets []BuildTarget

const (
	BuildTargetArchX64   BuildTarget = 0x00010000
	BuildTargetArchX86   BuildTarget = 0x00020000
	BuildTargetArchArm64 BuildTarget = 0x00040000
	BuildTargetArchArm32 BuildTarget = 0x00080000
	BuildTargetArchEsp32 BuildTarget = 0x00100000

	BuildTargetOsWindows BuildTarget = 0x00000001
	BuildTargetOsMac     BuildTarget = 0x00000002
	BuildTargetOsLinux   BuildTarget = 0x00000004
	BuildTargetOsArduino BuildTarget = 0x00000008

	BuildTargetWindowsX64   BuildTarget = BuildTargetOsWindows | BuildTargetArchX64
	BuildTargetMacX64       BuildTarget = BuildTargetOsMac | BuildTargetArchX64
	BuildTargetMacArm64     BuildTarget = BuildTargetOsMac | BuildTargetArchArm64
	BuildTargetLinuxX64     BuildTarget = BuildTargetOsLinux | BuildTargetArchX64
	BuildTargetLinuxArm64   BuildTarget = BuildTargetOsLinux | BuildTargetArchArm64
	BuildTargetArduinoEsp32 BuildTarget = BuildTargetOsArduino | BuildTargetArchEsp32
)

var BuildTargetsAll = []BuildTarget{
	BuildTargetWindowsX64,
	BuildTargetMacX64,
	BuildTargetMacArm64,
	BuildTargetLinuxX64,
	BuildTargetLinuxArm64,
	BuildTargetArduinoEsp32,
}
var BuildTargetsDesktop = []BuildTarget{
	BuildTargetWindowsX64,
	BuildTargetMacX64,
	BuildTargetMacArm64,
	BuildTargetLinuxX64,
	BuildTargetLinuxArm64,
}

var BuildTargetsArduino = []BuildTarget{
	BuildTargetArduinoEsp32,
}

func (pt BuildTarget) IsEqual(other BuildTarget) bool {
	return pt&other == other
}

func (bt BuildTargets) Contains(target BuildTarget) bool {
	for _, t := range bt {
		if t.IsEqual(target) {
			return true
		}
	}
	return false
}

func (pt BuildTarget) Windows() bool {
	return pt&(BuildTargetOsWindows) != 0
}

func (pt BuildTarget) Mac() bool {
	return pt&(BuildTargetOsMac) != 0
}

func (pt BuildTarget) Linux() bool {
	return pt&(BuildTargetOsLinux) != 0
}

func (pt BuildTarget) Arduino() bool {
	return pt&BuildTargetOsArduino != 0
}

func (pt BuildTarget) X64() bool {
	return pt&(BuildTargetArchX64) != 0
}

func (pt BuildTarget) X86() bool {
	return pt&(BuildTargetArchX86) != 0
}

func (pt BuildTarget) Arm64() bool {
	return pt&(BuildTargetArchArm64) != 0
}

func (pt BuildTarget) Arm32() bool {
	return pt&(BuildTargetArchArm32) != 0
}

func (pt BuildTarget) Esp32() bool {
	return pt&(BuildTargetArchEsp32) != 0
}
