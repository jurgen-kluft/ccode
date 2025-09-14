package dev

import "runtime"

// BuildTarget defines:
// - The target OS (Windows, Mac, Linux, iOS, Arduino)
// - The target architecture (x86, x64, arm32, arm64, esp32)
// - The bitness (32-bit, 64-bit)

// BuildTargetOs indicates the target OS to build for
type BuildTargetOs uint8

const (
	BuildTargetOsWindows BuildTargetOs = 0
	BuildTargetOsMac     BuildTargetOs = 1
	BuildTargetOsLinux   BuildTargetOs = 2
	BuildTargetOsiOS     BuildTargetOs = 3
	BuildTargetOsArduino BuildTargetOs = 4
	BuildTargetOsCount   BuildTargetOs = 5
)

// BuildTargetArch indicates the target architecture
type BuildTargetArch uint64

const (
	BuildTargetArchNone    BuildTargetArch = 0x0000
	BuildTargetArchMask    BuildTargetArch = 0xffff
	BuildTargetVariantMask BuildTargetArch = 0xffff
	BuildTargetArchX86     BuildTargetArch = (1 << 0)
	BuildTargetArchX64     BuildTargetArch = (1 << 1)
	BuildTargetArchArm32   BuildTargetArch = (1 << 2)
	BuildTargetArchArm64   BuildTargetArch = (1 << 3)
	BuildTargetArchEsp32   BuildTargetArch = (1 << 4)
)

func (arch BuildTargetArch) String() string {
	var full string
	var sep string
	if arch&BuildTargetArchX86 != 0 {
		full = "x86"
		sep = "|"
	}
	if arch&BuildTargetArchX64 != 0 {
		full = full + sep + "x64"
		sep = "|"
	}
	if arch&BuildTargetArchArm32 != 0 {
		full = full + sep + "arm32"
		sep = "|"
	}
	if arch&BuildTargetArchArm64 != 0 {
		full = full + sep + "arm64"
		sep = "|"
	}
	if arch&BuildTargetArchEsp32 != 0 {
		full = full + sep + "esp32"
		sep = "|"
	}
	return full
}

// BuildTarget indicates per OS the supported architectures
type BuildTarget struct {
	Targets [BuildTargetOsCount]BuildTargetArch
}

var BuildTargetWindowsX86 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{BuildTargetArchX86, BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchNone}}
var BuildTargetWindowsX64 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{BuildTargetArchX64, BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchNone}}
var BuildTargetMacX64 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{BuildTargetArchNone, BuildTargetArchX64, BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchNone}}
var BuildTargetMacArm64 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{BuildTargetArchNone, BuildTargetArchArm64, BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchNone}}
var BuildTargetLinuxX86 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchX86, BuildTargetArchNone, BuildTargetArchNone}}
var BuildTargetLinuxX64 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchX64, BuildTargetArchNone, BuildTargetArchNone}}
var BuildTargetLinuxArm32 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchArm32, BuildTargetArchNone, BuildTargetArchNone}}
var BuildTargetLinuxArm64 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchArm64, BuildTargetArchNone, BuildTargetArchNone}}
var BuildTargetAppleiOS = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchArm64, BuildTargetArchNone}}
var BuildTargetArduinoEsp32 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchNone, BuildTargetArchEsp32}}

var BuildTargetsAll = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchX86 | BuildTargetArchX64,
	BuildTargetArchX64 | BuildTargetArchArm64,
	BuildTargetArchX86 | BuildTargetArchX64 | BuildTargetArchArm32 | BuildTargetArchArm64,
	BuildTargetArchArm64,
	BuildTargetArchEsp32,
}}

var BuildTargetsDesktop = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchX86 | BuildTargetArchX64,
	BuildTargetArchX64 | BuildTargetArchArm64,
	BuildTargetArchX86 | BuildTargetArchX64 | BuildTargetArchArm32 | BuildTargetArchArm64,
	BuildTargetArchNone,
	BuildTargetArchNone,
}}

var BuildTargetsArduino = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchEsp32,
}}

var EmptyBuildTarget = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
}}

var CurrentBuildTarget = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
}}

func SetBuildTarget(os string, arch string) BuildTarget {
	// Set the build target based on the provided dev, os, and arch
	CurrentBuildTarget = EmptyBuildTarget

	if os == "arduino" {
		CurrentBuildTarget.Targets[BuildTargetOsArduino] |= BuildTargetArchEsp32
	} else if os == "windows" {
		if arch == "x86" {
			CurrentBuildTarget.Targets[BuildTargetOsWindows] |= BuildTargetArchX86
		} else {
			CurrentBuildTarget.Targets[BuildTargetOsWindows] |= BuildTargetArchX64
		}
	} else if os == "darwin" {
		if arch == "x64" {
			CurrentBuildTarget.Targets[BuildTargetOsMac] |= BuildTargetArchX64
		} else {
			CurrentBuildTarget.Targets[BuildTargetOsMac] |= BuildTargetArchArm64
		}
	} else if os == "linux" {
		if arch == "x64" {
			CurrentBuildTarget.Targets[BuildTargetOsLinux] |= BuildTargetArchX64
		} else {
			CurrentBuildTarget.Targets[BuildTargetOsLinux] |= BuildTargetArchArm64
		}
	}

	return CurrentBuildTarget
}

func GetBuildTarget() BuildTarget {
	return CurrentBuildTarget
}

func GetBuildTargetTargettingHost() BuildTarget {
	// Just looking at the host OS and arch
	b := EmptyBuildTarget
	if runtime.GOOS == "windows" {
		switch runtime.GOARCH {
		case "amd64", "x64":
			b = BuildTargetWindowsX64
		case "x86":
			b = BuildTargetWindowsX86
		}
	} else if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "amd64", "x64":
			b = BuildTargetMacX64
		case "arm64":
			b = BuildTargetMacArm64
		}
	} else if runtime.GOOS == "linux" {
		switch runtime.GOARCH {
		case "amd64", "x64":
			b = BuildTargetLinuxX64
		case "arm64":
			b = BuildTargetLinuxArm64
		case "x86":
			b = BuildTargetLinuxX86
		case "arm":
			b = BuildTargetLinuxArm32
		}
	}
	return b
}

func (pt BuildTarget) IsEqual(other BuildTarget) bool {
	for i := 0; i < len(pt.Targets); i++ {
		if pt.Targets[i] != other.Targets[i] {
			return false
		}
	}
	return true
}

func (bt BuildTarget) Contains(target BuildTarget) bool {
	for i := 0; i < len(bt.Targets); i++ {
		if (bt.Targets[i] & target.Targets[i]) != target.Targets[i] {
			return false
		}
	}
	return true
}

func (pt BuildTarget) Windows() bool {
	return pt.Targets[BuildTargetOsWindows] != BuildTargetArchNone
}

func (pt BuildTarget) Mac() bool {
	return pt.Targets[BuildTargetOsMac] != BuildTargetArchNone
}

func (pt BuildTarget) AppleiOS() bool {
	return pt.Targets[BuildTargetOsiOS] != BuildTargetArchNone
}

func (pt BuildTarget) Linux() bool {
	return pt.Targets[BuildTargetOsLinux] != BuildTargetArchNone
}

func (pt BuildTarget) Arduino() bool {
	return pt.Targets[BuildTargetOsArduino] != BuildTargetArchNone
}

func (pt BuildTarget) X64() bool {
	for i := 0; i < len(pt.Targets); i++ {
		if pt.Targets[i]&BuildTargetArchX64 != 0 {
			return true
		}
	}
	return false
}

func (pt BuildTarget) X86() bool {
	for i := 0; i < len(pt.Targets); i++ {
		if pt.Targets[i]&BuildTargetArchX86 != 0 {
			return true
		}
	}
	return false
}

func (pt BuildTarget) Arm64() bool {
	for i := 0; i < len(pt.Targets); i++ {
		if pt.Targets[i]&BuildTargetArchArm64 != 0 {
			return true
		}
	}
	return false
}

func (pt BuildTarget) Arm32() bool {
	for i := 0; i < len(pt.Targets); i++ {
		if pt.Targets[i]&BuildTargetArchArm32 != 0 {
			return true
		}
	}
	return false
}

// We know that only Arduino OS can have ESP32 architecture
func (pt BuildTarget) Esp32() bool {
	return pt.Targets[BuildTargetOsArduino]&BuildTargetArchEsp32 == BuildTargetArchEsp32
}

func (pt BuildTarget) OSAsString() string {
	switch {
	case pt.Windows():
		return "windows"
	case pt.Mac():
		return "darwin"
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

func (pt BuildTarget) ArchAsUcString() string {
	switch {
	case pt.X64():
		return "X64"
	case pt.X86():
		return "X86"
	case pt.Arm64():
		return "ARM64"
	case pt.Arm32():
		return "ARM32"
	case pt.Esp32():
		return "ESP32"
	default:
		return "UNKNOWN"
	}
}
func (pt BuildTarget) String() string {
	var full string
	for i := 0; i < int(BuildTargetOsCount); i++ {
		arch := pt.Targets[i]
		if arch != BuildTargetArchNone {
			os := ""
			switch i {
			case int(BuildTargetOsWindows):
				os = "windows"
			case int(BuildTargetOsMac):
				os = "mac"
			case int(BuildTargetOsLinux):
				os = "linux"
			case int(BuildTargetOsiOS):
				os = "ios"
			case int(BuildTargetOsArduino):
				os = "arduino"
			}
			if full == "" {
				full = os + "(" + arch.String() + ")"
			} else {
				full += ", " + os + "(" + arch.String() + ")"
			}
		}
	}
	return full
}

func BuildTargetFromString(os string, arch string) BuildTarget {
	// Set the build target based on the provided os and arch
	switch os {
	case "arduino":
		return BuildTargetArduinoEsp32
	case "windows":
		if arch == "x86" {
			return BuildTargetWindowsX86
		}
		return BuildTargetWindowsX64
	case "darwin":
		if arch == "x64" {
			return BuildTargetMacX64
		}
		return BuildTargetMacArm64
	case "linux":
		if arch == "arm64" {
			return BuildTargetLinuxArm64
		} else if arch == "arm32" {
			return BuildTargetLinuxArm32
		} else if arch == "x86" {
			return BuildTargetLinuxX86
		}
		return BuildTargetLinuxX64
	default:
		break
	}
	return EmptyBuildTarget
}
