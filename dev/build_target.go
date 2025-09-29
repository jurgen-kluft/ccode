package dev

import (
	"runtime"
	"strings"
)

// BuildTarget defines:
// - The target OS (Windows, Mac, Linux, iOS, Arduino)
// - The target architecture (x86, x64, arm32, arm64, esp32)
// - The bitness (32-bit, 64-bit)

// BuildTargetOs indicates the target OS to build for
type BuildTargetOs uint8

const (
	BuildTargetOsEmpty   BuildTargetOs = 0
	BuildTargetOsWindows BuildTargetOs = 1
	BuildTargetOsMac     BuildTargetOs = 2
	BuildTargetOsLinux   BuildTargetOs = 3
	BuildTargetOsiOS     BuildTargetOs = 4
	BuildTargetOsArduino BuildTargetOs = 5
	BuildTargetOsCount   BuildTargetOs = 6
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
	BuildTargetArchEsp8266 BuildTargetArch = (1 << 5)
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
	if arch&BuildTargetArchEsp8266 != 0 {
		full = full + sep + "esp8266"
		sep = "|"
	}
	return full
}

// BuildTarget indicates per OS the supported architectures
type BuildTarget struct {
	Targets [BuildTargetOsCount]BuildTargetArch
}

var BuildTargetWindowsX86 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchX86,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
}}
var BuildTargetWindowsX64 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchX64,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
}}
var BuildTargetMacX64 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchX64,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
}}
var BuildTargetMacArm64 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchArm64,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
}}
var BuildTargetLinuxX86 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchX86,
	BuildTargetArchNone,
	BuildTargetArchNone,
}}
var BuildTargetLinuxX64 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchX64,
	BuildTargetArchNone,
	BuildTargetArchNone,
}}
var BuildTargetLinuxArm32 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchArm32,
	BuildTargetArchNone,
	BuildTargetArchNone,
}}
var BuildTargetLinuxArm64 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchArm64,
	BuildTargetArchNone,
	BuildTargetArchNone,
}}
var BuildTargetAppleiOS = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchArm64,
	BuildTargetArchNone,
}}
var BuildTargetArduinoEsp32 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchEsp32,
}}
var BuildTargetArduinoEsp8266 = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchNone,
	BuildTargetArchEsp8266,
}}

var BuildTargetsAll = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
	BuildTargetArchX86 | BuildTargetArchX64,
	BuildTargetArchX64 | BuildTargetArchArm64,
	BuildTargetArchX86 | BuildTargetArchX64 | BuildTargetArchArm32 | BuildTargetArchArm64,
	BuildTargetArchArm64,
	BuildTargetArchEsp32 | BuildTargetArchEsp8266,
}}

var BuildTargetsDesktop = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
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
	BuildTargetArchNone,
	BuildTargetArchEsp32 | BuildTargetArchEsp8266,
}}

var EmptyBuildTarget = BuildTarget{Targets: [BuildTargetOsCount]BuildTargetArch{
	BuildTargetArchNone,
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
	BuildTargetArchNone,
}}

func SetBuildTarget(os string, arch string) BuildTarget {
	// Set the build target based on the provided dev, os, and arch
	CurrentBuildTarget = EmptyBuildTarget

	if os == "arduino" {
		CurrentBuildTarget.Targets[BuildTargetOsArduino] |= BuildTargetArchEsp32
		CurrentBuildTarget.Targets[BuildTargetOsArduino] |= BuildTargetArchEsp8266
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

func GetBuildTargetArchFromString(arch string) (b BuildTargetArch) {
	b = BuildTargetArchNone
	switch arch {
	case "amd64", "x64":
		b = BuildTargetArchX64
	case "x86":
		b = BuildTargetArchX86
	case "arm32":
		b = BuildTargetArchArm32
	case "arm64":
		b = BuildTargetArchArm64
	case "esp32":
		b = BuildTargetArchEsp32
	case "esp8266":
		b = BuildTargetArchEsp8266
	}
	return b
}

func GetBuildTargetOsFromString(os string) BuildTargetOs {
	switch os {
	case "windows":
		return BuildTargetOsWindows
	case "darwin", "mac":
		return BuildTargetOsMac
	case "linux":
		return BuildTargetOsLinux
	case "ios":
		return BuildTargetOsiOS
	case "arduino":
		return BuildTargetOsArduino
	}
	return BuildTargetOsEmpty
}

func GetBuildTargetFromOsArch(os string, arch string) BuildTarget {
	bos := GetBuildTargetOsFromString(os)
	barch := GetBuildTargetArchFromString(arch)
	b := EmptyBuildTarget
	b.Targets[bos] = barch
	return b
}

func GetBuildTargetTargettingHost() BuildTarget {
	return GetBuildTargetFromOsArch(runtime.GOOS, runtime.GOARCH)
}

func (pt BuildTarget) IsEqual(other BuildTarget) bool {
	for i := 0; i < len(pt.Targets); i++ {
		if pt.Targets[i] != other.Targets[i] {
			return false
		}
	}
	return true
}

func (bt BuildTarget) HasOverlap(target BuildTarget) bool {
	for i := 0; i < len(bt.Targets); i++ {
		if (bt.Targets[i] & target.Targets[i]) != 0 {
			return true
		}
	}
	return false
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

// We know that only Arduino OS can have ESP8266 architecture
func (pt BuildTarget) Esp8266() bool {
	return pt.Targets[BuildTargetOsArduino]&BuildTargetArchEsp8266 == BuildTargetArchEsp8266
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
	case pt.Esp8266():
		return "esp8266"
	default:
		return "unknown"
	}
}

func (pt BuildTarget) ArchAsUcString() string {
	arch := pt.ArchAsString()
	return strings.ToUpper(arch)
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

func BuildTargetFromString(s string) BuildTarget {
	b := EmptyBuildTarget
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			openIdx := strings.Index(part, "(")
			closeIdx := strings.Index(part, ")")
			if openIdx > 0 && closeIdx > openIdx {
				os := strings.TrimSpace(part[0:openIdx])
				archStr := strings.TrimSpace(part[openIdx+1 : closeIdx])
				var arch BuildTargetArch
				archParts := strings.Split(archStr, "|")
				for _, archPart := range archParts {
					archPart = strings.TrimSpace(archPart)
					arch |= GetBuildTargetArchFromString(archPart)
				}
				b.Targets[GetBuildTargetOsFromString(os)] = arch
			}
		}
	}

	return b
}
