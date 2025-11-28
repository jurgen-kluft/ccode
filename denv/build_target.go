package denv

import (
	"runtime"
	"strings"
)

// BuildTarget defines:
// - The target OS (Windows, Mac, Linux, iOS, Arduino)
// - The target architecture (x86, x64, arm32, arm64, esp32)
// - The bitness (32-bit, 64-bit)

// BuildTargetOs indicates the target OS to build for
type BuildTargetOs uint64

const (
	BuildTargetOsEmpty   BuildTargetOs = 0
	BuildTargetOsWindows BuildTargetOs = 1 << (32 + 0)
	BuildTargetOsMac     BuildTargetOs = 1 << (32 + 1)
	BuildTargetOsLinux   BuildTargetOs = 1 << (32 + 2)
	BuildTargetOsiOS     BuildTargetOs = 1 << (32 + 3)
	BuildTargetOsArduino BuildTargetOs = 1 << (32 + 4)
	BuildTargetOsCount   BuildTargetOs = 1 << (32 + 5)
)

func (bto BuildTargetOs) String() string {
	switch {
	case bto.Windows():
		return "windows"
	case bto.Mac():
		return "darwin"
	case bto.Linux():
		return "linux"
	case bto.iOS():
		return "ios"
	case bto.Arduino():
		return "arduino"
	}
	return "unknown"
}

func (os BuildTargetOs) Windows() bool {
	return os&BuildTargetOsWindows != 0
}
func (os BuildTargetOs) Mac() bool {
	return os&BuildTargetOsMac != 0
}
func (os BuildTargetOs) Linux() bool {
	return os&BuildTargetOsLinux != 0
}
func (os BuildTargetOs) Arduino() bool {
	return os&BuildTargetOsArduino != 0
}
func (os BuildTargetOs) iOS() bool {
	return os&BuildTargetOsiOS != 0
}
func (os BuildTargetOs) Settings() BuildTargetOsSettings {
	if s, ok := BuildTargetOsSettingsDB[os]; ok {
		return s
	}
	return BuildTargetOsSettings{}
}

type BuildTargetOsSettings struct {
	ExeTargetPrefix string
	ExeTargetSuffix string
	DllTargetPrefix string
	DllTargetSuffix string
	LibTargetPrefix string
	LibTargetSuffix string
}

var BuildTargetOsSettingsDB = map[BuildTargetOs]BuildTargetOsSettings{
	BuildTargetOsWindows: {
		ExeTargetPrefix: "",
		ExeTargetSuffix: ".exe",
		DllTargetPrefix: "",
		DllTargetSuffix: ".dll",
		LibTargetPrefix: "",
		LibTargetSuffix: ".lib",
	},
	BuildTargetOsMac: {
		ExeTargetPrefix: "",
		ExeTargetSuffix: "",
		DllTargetPrefix: "lib",
		DllTargetSuffix: ".dylib",
		LibTargetPrefix: "lib",
		LibTargetSuffix: ".a",
	},
	BuildTargetOsLinux: {
		ExeTargetPrefix: "",
		ExeTargetSuffix: "",
		DllTargetPrefix: "lib",
		DllTargetSuffix: ".so",
		LibTargetPrefix: "lib",
		LibTargetSuffix: ".a",
	},
	BuildTargetOsiOS: {
		ExeTargetPrefix: "",
		ExeTargetSuffix: "",
		DllTargetPrefix: "lib",
		DllTargetSuffix: ".dylib",
		LibTargetPrefix: "lib",
		LibTargetSuffix: ".a",
	},
	BuildTargetOsArduino: {
		ExeTargetPrefix: "",
		ExeTargetSuffix: ".elf",
		DllTargetPrefix: "",
		DllTargetSuffix: ".dll",
		LibTargetPrefix: "lib",
		LibTargetSuffix: ".a",
	},
}

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
type BuildTarget map[BuildTargetOs]BuildTargetArch

var BuildTargetWindowsX86 = BuildTarget{BuildTargetOsWindows: BuildTargetArchX86}
var BuildTargetWindowsX64 = BuildTarget{BuildTargetOsWindows: BuildTargetArchX64}
var BuildTargetMacX64 = BuildTarget{BuildTargetOsMac: BuildTargetArchX64}
var BuildTargetMacArm64 = BuildTarget{BuildTargetOsMac: BuildTargetArchArm64}
var BuildTargetMacUniversal = BuildTarget{BuildTargetOsMac: BuildTargetArchX64 | BuildTargetArchArm64}
var BuildTargetLinuxX86 = BuildTarget{BuildTargetOsLinux: BuildTargetArchX86}
var BuildTargetLinuxX64 = BuildTarget{BuildTargetOsLinux: BuildTargetArchX64}
var BuildTargetLinuxArm32 = BuildTarget{BuildTargetOsLinux: BuildTargetArchArm32}
var BuildTargetLinuxArm64 = BuildTarget{BuildTargetOsLinux: BuildTargetArchArm64}
var BuildTargetLinuxUniversal = BuildTarget{BuildTargetOsLinux: BuildTargetArchX86 | BuildTargetArchX64 | BuildTargetArchArm32 | BuildTargetArchArm64}
var BuildTargetAppleiOS = BuildTarget{BuildTargetOsiOS: BuildTargetArchArm64}
var BuildTargetArduinoEsp32 = BuildTarget{BuildTargetOsArduino: BuildTargetArchEsp32}
var BuildTargetArduinoEsp8266 = BuildTarget{BuildTargetOsArduino: BuildTargetArchEsp8266}
var BuildTargetsAll = BuildTarget{
	BuildTargetOsWindows: BuildTargetArchX86 | BuildTargetArchX64,
	BuildTargetOsMac:     BuildTargetArchX64 | BuildTargetArchArm64,
	BuildTargetOsLinux:   BuildTargetArchX86 | BuildTargetArchX64 | BuildTargetArchArm32 | BuildTargetArchArm64,
	BuildTargetOsiOS:     BuildTargetArchArm64,
	BuildTargetOsArduino: BuildTargetArchEsp32 | BuildTargetArchEsp8266,
}

var BuildTargetsDesktop = BuildTarget{
	BuildTargetOsWindows: BuildTargetArchX86 | BuildTargetArchX64,
	BuildTargetOsMac:     BuildTargetArchX64 | BuildTargetArchArm64,
	BuildTargetOsLinux:   BuildTargetArchX86 | BuildTargetArchX64 | BuildTargetArchArm32 | BuildTargetArchArm64,
}

var BuildTargetsArduino = BuildTarget{BuildTargetOsArduino: BuildTargetArchEsp32 | BuildTargetArchEsp8266}
var EmptyBuildTarget = BuildTarget{}
var CurrentBuildTarget = BuildTarget{}

func SetBuildTarget(os string, arch string) BuildTarget {
	// Set the build target based on the provided dev, os, and arch
	CurrentBuildTarget = EmptyBuildTarget

	if os == "arduino" {
		if arch == "esp32" {
			CurrentBuildTarget = BuildTargetArduinoEsp32
		} else {
			CurrentBuildTarget = BuildTargetArduinoEsp8266
		}
	} else if os == "windows" {
		if arch == "x86" {
			CurrentBuildTarget = BuildTargetWindowsX86
		} else {
			CurrentBuildTarget = BuildTargetWindowsX64
		}
	} else if os == "darwin" {
		if arch == "x64" {
			CurrentBuildTarget = BuildTargetMacX64
		} else {
			CurrentBuildTarget = BuildTargetMacArm64
		}
	} else if os == "linux" {
		if arch == "x64" {
			CurrentBuildTarget = BuildTargetLinuxX64
		} else {
			CurrentBuildTarget = BuildTargetLinuxArm64
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

	if bos != BuildTargetOsEmpty && barch != BuildTargetArchNone {
		return BuildTarget{bos: barch}
	}
	return EmptyBuildTarget
}

func GetBuildTargetTargettingHost() BuildTarget {
	return GetBuildTargetFromOsArch(runtime.GOOS, runtime.GOARCH)
}

func (pt BuildTarget) IsEqual(other BuildTarget) bool {
	if len(pt) != len(other) {
		return false
	}
	for os, arch := range pt {
		if otherArch, ok := other[os]; !ok || otherArch != arch {
			return false
		}
	}
	return true
}

func (bt BuildTarget) Union(target BuildTarget) BuildTarget {
	result := BuildTarget{}
	for os, arch := range bt {
		result[os] = arch
	}
	for os, arch := range target {
		if existingArch, ok := result[os]; ok {
			result[os] = existingArch | arch
		} else {
			result[os] = arch
		}
	}
	return result
}

func (bt BuildTarget) HasOverlap(target BuildTarget) bool {
	for os, arch := range bt {
		if targetArch, ok := target[os]; ok && (arch&targetArch) != 0 {
			return true
		}
	}
	return false
}

func (bt BuildTarget) Arch() BuildTargetArch {
	if len(bt) == 1 {
		for _, arch := range bt {
			return arch
		}
	}
	return BuildTargetArchNone
}

func (bt BuildTarget) Os() BuildTargetOs {
	if len(bt) == 1 {
		for os := range bt {
			return os
		}
	}
	return BuildTargetOsEmpty
}

func (pt BuildTarget) Windows() bool {
	arch, ok := (pt)[BuildTargetOsWindows]
	return ok && arch != BuildTargetArchNone
}

func (pt BuildTarget) Mac() bool {
	arch, ok := (pt)[BuildTargetOsMac]
	return ok && arch != BuildTargetArchNone
}

func (pt BuildTarget) AppleiOS() bool {
	arch, ok := (pt)[BuildTargetOsiOS]
	return ok && arch != BuildTargetArchNone
}

func (pt BuildTarget) Linux() bool {
	arch, ok := (pt)[BuildTargetOsLinux]
	return ok && arch != BuildTargetArchNone
}

func (pt BuildTarget) Arduino() bool {
	arch, ok := (pt)[BuildTargetOsArduino]
	return ok && arch != BuildTargetArchNone
}

func (pt BuildTarget) OsHasArch(_os BuildTargetOs, _arch BuildTargetArch) bool {
	if arch, ok := (pt)[_os]; ok {
		return arch&_arch != 0
	}
	return false
}

func (pt BuildTarget) HasArch(_arch BuildTargetArch) bool {
	for _, arch := range pt {
		return arch&_arch != 0
	}
	return false
}

func (pt BuildTarget) X64() bool {
	return pt.HasArch(BuildTargetArchX64)
}

func (pt BuildTarget) X86() bool {
	return pt.HasArch(BuildTargetArchX86)
}

func (pt BuildTarget) Arm64() bool {
	return pt.HasArch(BuildTargetArchArm64)
}

func (pt BuildTarget) Arm32() bool {
	return pt.HasArch(BuildTargetArchArm32)
}

func (pt BuildTarget) Esp32() bool {
	return pt.HasArch(BuildTargetArchEsp32)
}

func (pt BuildTarget) Esp8266() bool {
	return pt.HasArch(BuildTargetArchEsp8266)
}

func (pt BuildTarget) ArchAsUcString() string {
	if len(pt) == 1 {
		for _, arch := range pt {
			return strings.ToUpper(arch.String())
		}
	}
	return "UNKNOWN"
}

func (pt BuildTarget) String() string {
	var full string
	for os, arch := range pt {
		if arch != BuildTargetArchNone {
			os := os.String()
			if full == "" {
				full = os + "(" + arch.String() + ")"
			} else {
				full += "," + os + "(" + arch.String() + ")"
			}
		}
	}
	return full
}

func BuildTargetFromString(s string) BuildTarget {
	b := BuildTarget{}
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			openIdx := strings.Index(part, "(")
			closeIdx := strings.Index(part, ")")
			if openIdx > 0 && closeIdx > openIdx {
				osStr := strings.TrimSpace(part[0:openIdx])
				os := GetBuildTargetOsFromString(osStr)
				archStr := strings.TrimSpace(part[openIdx+1 : closeIdx])
				var arch BuildTargetArch
				archParts := strings.Split(archStr, "|")
				for _, archPart := range archParts {
					archPart = strings.TrimSpace(archPart)
					arch |= GetBuildTargetArchFromString(archPart)
				}
				b[os] = arch
			}
		}
	}
	return b
}
