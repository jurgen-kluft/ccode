package denv

import (
	"runtime"
	"strings"
)

type MakeHost struct {
	Os       string
	Compiler string
	Arch     string
}

// MakeHost is a singleton instance of the current host environment
func NewMakeHost(compiler string) MakeHost {
	return MakeHost{
		Os:       runtime.GOOS,
		Compiler: compiler,
		Arch:     runtime.GOARCH,
	}
}

func (h *MakeHost) HostOSIsLinux() bool {
	return strings.EqualFold(h.Os, OS_LINUX)
}
func (h *MakeHost) HostOSIsWindows() bool {
	return strings.EqualFold(h.Os, OS_WINDOWS)
}
func (h *MakeHost) HostOSIsMac() bool {
	return strings.EqualFold(h.Os, OS_MAC)
}
func (h *MakeHost) HostOSIsIos() bool {
	return strings.EqualFold(h.Os, OS_IOS)
}
func (h *MakeHost) HostArchIsX64() bool {
	return strings.EqualFold(h.Arch, ARCH_X64)
}
func (h *MakeHost) HostArchIsAmd64() bool {
	return strings.EqualFold(h.Arch, ARCH_AMD64)
}
func (h *MakeHost) HostArchIsArm64() bool {
	return strings.EqualFold(h.Arch, ARCH_ARM64)
}
func (h *MakeHost) HostOSAsString() string {
	return h.Os
}
func (h *MakeHost) HostArchAsString() string {
	return h.Arch
}

type MakeTarget struct {
	Compiler    string
	BuildTarget BuildTarget
}

var gMakeTargets map[BuildTarget]*MakeTarget

func NewMakeTarget(buildTarget BuildTarget, compiler string) *MakeTarget {
	if gMakeTargets == nil {
		gMakeTargets = make(map[BuildTarget]*MakeTarget)
	}
	// Keep make targets unique by os:compiler:arch
	if target, ok := gMakeTargets[buildTarget]; ok {
		return target
	}
	target := &MakeTarget{
		BuildTarget: buildTarget,
		Compiler:    compiler,
	}
	gMakeTargets[buildTarget] = target
	return target
}

func NewMakeTargetWindows() *MakeTarget {
	return NewMakeTarget(BuildTargetWindowsX64, COMPILER_VC)
}

func NewMakeTargetMacOS() *MakeTarget {
	arch := runtime.GOARCH
	if arch == ARCH_ARM64 {
		return NewMakeTarget(BuildTargetMacArm64, COMPILER_CLANG)
	}
	return NewMakeTarget(BuildTargetMacX64, COMPILER_CLANG)
}

func NewMakeTargetLinux() *MakeTarget {
	arch := runtime.GOARCH
	if arch == ARCH_ARM64 {
		return NewMakeTarget(BuildTargetLinuxArm64, COMPILER_CLANG)
	}
	return NewMakeTarget(BuildTargetLinuxX64, COMPILER_GCC)
}

func NewMakeTargetEsp32(_arch string) *MakeTarget {
	return NewMakeTarget(BuildTargetArduinoEsp32, COMPILER_GCC)
}

func NewDefaultMakeTarget(_dev DevEnum, _os string, _arch string) *MakeTarget {
	if _os == "windows" {
		return NewMakeTargetWindows()
	} else if _os == "darwin" {
		return NewMakeTargetMacOS()
	} else if _os == "linux" {
		return NewMakeTargetLinux()
	} else if _os == "arduino" {
		return NewMakeTargetEsp32(_arch)
	}
	return NewMakeTargetLinux()
}

const (
	OS_LINUX   = "linux"
	OS_WINDOWS = "windows"
	OS_MAC     = "mac"
	OS_IOS     = "ios"
	OS_ARDUINO = "arduino"
)

const (
	COMPILER_GCC   = "gcc"
	COMPILER_CLANG = "clang"
	COMPILER_VC    = "vc"
)

const (
	ARCH_X64     = "x86_64"
	ARCH_AMD64   = "amd64"
	ARCH_ARM64   = "arm64"
	ARCH_ESP32   = "esp32"
	ARCH_ESP32S3 = "esp32s3"
)

func (t *MakeTarget) OSIsLinux() bool {
	return t.BuildTarget.Linux()
}
func (t *MakeTarget) OSIsWindows() bool {
	return t.BuildTarget.Windows()
}
func (t *MakeTarget) OSIsMac() bool {
	return t.BuildTarget.Mac()
}
func (t *MakeTarget) OSIsiOS() bool {
	return t.BuildTarget.AppleiOS()
}
func (t *MakeTarget) OSIsArduino() bool {
	return t.BuildTarget.Arduino()
}
func (t *MakeTarget) CompilerIsGcc() bool {
	return strings.EqualFold(t.Compiler, COMPILER_GCC)
}
func (t *MakeTarget) CompilerIsClang() bool {
	return strings.EqualFold(t.Compiler, COMPILER_CLANG)
}
func (t *MakeTarget) CompilerIsVc() bool {
	return strings.EqualFold(t.Compiler, COMPILER_VC)
}
func (t *MakeTarget) ArchIsX64() bool {
	return t.BuildTarget.X64()
}
func (t *MakeTarget) ArchIsAmd64() bool {
	return t.BuildTarget.X64()
}
func (t *MakeTarget) ArchIsArm64() bool {
	return t.BuildTarget.Arm64()
}
func (t *MakeTarget) ArchIsEsp32() bool {
	return t.BuildTarget.Esp32()
}
func (t *MakeTarget) ArchIsEsp32Generic() bool {
	return t.BuildTarget.Esp32()
}
func (t *MakeTarget) ArchIsEsp32S3() bool {
	return t.BuildTarget.Esp32()
}
func (t *MakeTarget) OSAsString() string {
	return t.BuildTarget.OSAsString()
}
func (t *MakeTarget) CompilerAsString() string {
	return t.Compiler
}
func (t *MakeTarget) ArchAsString() string {
	return t.BuildTarget.ArchAsString()
}
