package denv

import "runtime"

// BuildTarget defines the type of build target
type BuildTarget uint32
type BuildTargets []BuildTarget

const (
	BuildTargetNotSupported BuildTarget = 0x00000000

	BuildTargetXXBitMask BuildTarget = 0x000000ff
	BuildTarget32Bit     BuildTarget = 0x00000001
	BuildTarget64Bit     BuildTarget = 0x00000002

	BuildTargetOsMask    BuildTarget = 0x00000ff00
	BuildTargetOsWindows BuildTarget = 0x000000100
	BuildTargetOsMac     BuildTarget = 0x000000200
	BuildTargetOsLinux   BuildTarget = 0x000000400
	BuildTargetOsiOS     BuildTarget = 0x000000800
	BuildTargetOsArduino BuildTarget = 0x000001000

	BuildTargetArchMask  BuildTarget = 0x00ff0000
	BuildTargetArchX86   BuildTarget = 0x00010000 | BuildTarget32Bit
	BuildTargetArchX64   BuildTarget = 0x00020000 | BuildTarget64Bit
	BuildTargetArchAmd64 BuildTarget = 0x00040000 | BuildTarget64Bit
	BuildTargetArchArm64 BuildTarget = 0x00080000 | BuildTarget64Bit
	BuildTargetArchArm32 BuildTarget = 0x00100000 | BuildTarget32Bit
	BuildTargetArchEsp32 BuildTarget = 0x00200000 | BuildTarget32Bit

	BuildTargetMask         BuildTarget = BuildTargetXXBitMask | BuildTargetOsMask | BuildTargetArchMask
	BuildTargetWindowsX86   BuildTarget = BuildTargetOsWindows | BuildTargetArchX86
	BuildTargetWindowsX64   BuildTarget = BuildTargetOsWindows | BuildTargetArchX64
	BuildTargetWindowsAmd64 BuildTarget = BuildTargetOsWindows | BuildTargetArchAmd64
	BuildTargetMacX64       BuildTarget = BuildTargetOsMac | BuildTargetArchX64
	BuildTargetMacArm64     BuildTarget = BuildTargetOsMac | BuildTargetArchArm64
	BuildTargetLinuxX86     BuildTarget = BuildTargetOsLinux | BuildTargetArchX86
	BuildTargetLinuxX64     BuildTarget = BuildTargetOsLinux | BuildTargetArchX64
	BuildTargetLinuxArm64   BuildTarget = BuildTargetOsLinux | BuildTargetArchArm64
	BuildTargetLinuxArm32   BuildTarget = BuildTargetOsLinux | BuildTargetArchArm32
	BuildTargetAppleiOS     BuildTarget = BuildTargetOsiOS | BuildTargetArchArm64
	BuildTargetArduinoEsp32 BuildTarget = BuildTargetOsArduino | BuildTargetArchEsp32
)

var BuildTargetsAll = []BuildTarget{
	BuildTargetWindowsX64,
	BuildTargetMacX64,
	BuildTargetMacArm64,
	BuildTargetLinuxX64,
	BuildTargetLinuxArm64,
	BuildTargetAppleiOS,
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

var CurrentBuildTarget BuildTarget = BuildTargetNotSupported

func SetBuildTarget(os string, arch string) BuildTarget {

	// Set the build target based on the provided dev, os, and arch
	CurrentBuildTarget = BuildTargetNotSupported

	if os == "arduino" {
		CurrentBuildTarget = BuildTargetArduinoEsp32
	} else if os == "windows" && arch == "x64" {
		CurrentBuildTarget = BuildTargetWindowsX64
	} else if os == "mac" && arch == "x64" {
		CurrentBuildTarget = BuildTargetMacX64
	} else if os == "mac" && arch == "arm64" {
		CurrentBuildTarget = BuildTargetMacArm64
	} else if os == "linux" && arch == "x64" {
		CurrentBuildTarget = BuildTargetLinuxX64
	} else if os == "linux" && arch == "arm64" {
		CurrentBuildTarget = BuildTargetLinuxArm64
	}

	return CurrentBuildTarget
}

func GetBuildTarget() BuildTarget {
	return CurrentBuildTarget
}

func GetBuildTargetTargettingHost() BuildTarget {
	// Just looking at the host OS and arch
	switch {
	case runtime.GOOS == "windows" && runtime.GOARCH == "amd64":
		return BuildTargetWindowsAmd64
	case runtime.GOOS == "windows" && runtime.GOARCH == "x64":
		return BuildTargetWindowsX64
	case runtime.GOOS == "windows" && runtime.GOARCH == "x86":
		return BuildTargetWindowsX86
	case runtime.GOOS == "darwin" && runtime.GOARCH == "amd64":
		return BuildTargetMacX64
	case runtime.GOOS == "darwin" && runtime.GOARCH == "arm64":
		return BuildTargetMacArm64
	case runtime.GOOS == "linux" && runtime.GOARCH == "amd64":
		return BuildTargetLinuxX64
	case runtime.GOOS == "linux" && runtime.GOARCH == "arm64":
		return BuildTargetLinuxArm64
	case runtime.GOOS == "linux" && runtime.GOARCH == "x86":
		return BuildTargetLinuxX86
	case runtime.GOOS == "linux" && runtime.GOARCH == "arm":
		return BuildTargetLinuxArm32
	default:
		return BuildTargetNotSupported
	}
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
	return pt&BuildTargetOsMask == BuildTargetOsWindows
}

func (pt BuildTarget) Mac() bool {
	return pt&BuildTargetOsMask == BuildTargetOsMac
}

func (pt BuildTarget) AppleiOS() bool {
	return pt&BuildTargetOsMask == BuildTargetOsiOS
}

func (pt BuildTarget) Linux() bool {
	return pt&BuildTargetOsMask == BuildTargetOsLinux
}

func (pt BuildTarget) Arduino() bool {
	return pt&BuildTargetOsMask == BuildTargetOsArduino
}

func (pt BuildTarget) X64() bool {
	return pt&BuildTargetArchMask == BuildTargetArchX64
}

func (pt BuildTarget) Amd64() bool {
	return pt&BuildTargetArchMask == BuildTargetArchAmd64
}

func (pt BuildTarget) X86() bool {
	return pt&BuildTargetArchMask == BuildTargetArchX86
}

func (pt BuildTarget) Arm64() bool {
	return pt&BuildTargetArchMask == BuildTargetArchArm64
}

func (pt BuildTarget) Arm32() bool {
	return pt&BuildTargetArchMask == BuildTargetArchArm32
}

func (pt BuildTarget) Esp32() bool {
	return pt&BuildTargetArchMask == BuildTargetArchEsp32
}

func (pt BuildTarget) OSAsString() string {
	switch {
	case pt.Windows():
		return "windows"
	case pt.Mac():
		return "mac"
	case pt.Linux():
		return "linux"
	case pt.AppleiOS():
		return "ios"
	case pt.Arduino():
		return "arduino"
	default:
		return "unknown"
	}
}

func (pt BuildTarget) ArchAsString() string {
	switch {
	case pt.X64():
		return "x64"
	case pt.X86():
		return "x86"
	case pt.Arm64():
		return "arm64"
	case pt.Arm32():
		return "arm32"
	case pt.Esp32():
		return "esp32"
	default:
		return "unknown"
	}
}
