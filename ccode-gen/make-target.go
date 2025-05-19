package ccode_gen

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
	Os       string
	Compiler string
	Arch     string
}

var gMakeTargets map[string]*MakeTarget

func NewMakeTarget(os string, compiler string, cpu string) *MakeTarget {
	if gMakeTargets == nil {
		gMakeTargets = make(map[string]*MakeTarget)
	}
	// Keep make targets unique by os:compiler:arch
	key := strings.ToLower(os + ":" + compiler + ":" + cpu)
	if target, ok := gMakeTargets[key]; ok {
		return target
	}
	target := &MakeTarget{
		Os:       os,
		Compiler: compiler,
		Arch:     cpu,
	}
	gMakeTargets[key] = target
	return target
}

func NewMakeTargetWindows() *MakeTarget {
	return NewMakeTarget(OS_WINDOWS, COMPILER_VC, ARCH_X64)
}

func NewMakeTargetMacOS() *MakeTarget {
	arch := runtime.GOARCH
	return NewMakeTarget(OS_MAC, COMPILER_CLANG, arch)
}

func NewMakeTargetLinux() *MakeTarget {
	arch := runtime.GOARCH
	return NewMakeTarget(OS_LINUX, COMPILER_CLANG, arch)
}

func NewMakeTargetEsp32(_arch string) *MakeTarget {
	if strings.EqualFold(_arch, ARCH_ESP32S3) {
		return NewMakeTarget(OS_ARDUINO, COMPILER_GCC, ARCH_ESP32S3)
	}
	return NewMakeTarget(OS_ARDUINO, COMPILER_GCC, ARCH_ESP32)
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
	return strings.EqualFold(t.Os, OS_LINUX)
}
func (t *MakeTarget) OSIsWindows() bool {
	return strings.EqualFold(t.Os, OS_WINDOWS)
}
func (t *MakeTarget) OSIsMac() bool {
	return strings.EqualFold(t.Os, OS_MAC)
}
func (t *MakeTarget) OSIsIos() bool {
	return strings.EqualFold(t.Os, OS_IOS)
}
func (t *MakeTarget) OSIsArduino() bool {
	return strings.EqualFold(t.Os, OS_ARDUINO)
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
	return strings.EqualFold(t.Arch, ARCH_X64)
}
func (t *MakeTarget) ArchIsAmd64() bool {
	return strings.EqualFold(t.Arch, ARCH_AMD64)
}
func (t *MakeTarget) ArchIsArm64() bool {
	return strings.EqualFold(t.Arch, ARCH_ARM64)
}
func (t *MakeTarget) ArchIsEsp32() bool {
	return strings.HasPrefix(t.Arch, ARCH_ESP32)
}
func (t *MakeTarget) ArchIsEsp32Generic() bool {
	return strings.EqualFold(t.Arch, ARCH_ESP32)
}
func (t *MakeTarget) ArchIsEsp32S3() bool {
	return strings.HasPrefix(t.Arch, ARCH_ESP32S3)
}
func (t *MakeTarget) OSAsString() string {
	return t.Os
}
func (t *MakeTarget) CompilerAsString() string {
	return t.Compiler
}
func (t *MakeTarget) ArchAsString() string {
	return t.Arch
}
