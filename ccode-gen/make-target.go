package ccode_gen

import (
	"runtime"
	"strings"
)

type MakeTarget interface {
	OSIsLinux() bool
	OSIsWindows() bool
	OSIsMac() bool
	OSIsIos() bool

	CompilerIsGcc() bool
	CompilerIsClang() bool
	CompilerIsVc() bool

	ArchIsX64() bool
	ArchIsAmd64() bool
	ArchIsArm64() bool

	OSAsString() string
	CompilerAsString() string
	ArchAsString() string
}

func NewMakeTarget(os string, compiler string, cpu string) MakeTarget {
	target := &MakeTargetInstance{
		Os:       os,
		Compiler: compiler,
		Cpu:      cpu,
	}
	target.MakeTarget = target
	return target
}

func NewMakeTargetWindows() MakeTarget {
	return NewMakeTarget(OS_WINDOWS, COMPILER_VC, ARCH_X64)
}

func NewMakeTargetMacOS() MakeTarget {
	arch := runtime.GOARCH
	return NewMakeTarget(OS_MAC, COMPILER_CLANG, arch)
}

func NewMakeTargetLinux() MakeTarget {
	arch := runtime.GOARCH
	return NewMakeTarget(OS_LINUX, COMPILER_CLANG, arch)
}

func NewMakeTargetEsp32() MakeTarget {
	return NewMakeTarget(OS_ARDUINO, COMPILER_GCC, ARCH_ESP32)
}

func NewDefaultMakeTarget(_dev DevEnum, _os string, _arch string) MakeTarget {
	if _os == "windows" {
		return NewMakeTargetWindows()
	} else if _os == "darwin" {
		return NewMakeTargetMacOS()
	} else if _os == "linux" {
		return NewMakeTargetLinux()
	} else if _os == "arduino" {
		return NewMakeTargetEsp32()
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
	ARCH_X64   = "x86_64"
	ARCH_AMD64 = "amd64"
	ARCH_ARM64 = "arm64"
	ARCH_ESP32 = "esp32"
)

type MakeTargetInstance struct {
	MakeTarget
	Os       string
	Compiler string
	Cpu      string
}

func (t *MakeTargetInstance) OSIsLinux() bool {
	return strings.EqualFold(t.Os, OS_LINUX)
}
func (t *MakeTargetInstance) OSIsWindows() bool {
	return strings.EqualFold(t.Os, OS_WINDOWS)
}
func (t *MakeTargetInstance) OSIsMac() bool {
	return strings.EqualFold(t.Os, OS_MAC)
}
func (t *MakeTargetInstance) OSIsIos() bool {
	return strings.EqualFold(t.Os, OS_IOS)
}
func (t *MakeTargetInstance) CompilerIsGcc() bool {
	return strings.EqualFold(t.Compiler, COMPILER_GCC)
}
func (t *MakeTargetInstance) CompilerIsClang() bool {
	return strings.EqualFold(t.Compiler, COMPILER_CLANG)
}
func (t *MakeTargetInstance) CompilerIsVc() bool {
	return strings.EqualFold(t.Compiler, COMPILER_VC)
}
func (t *MakeTargetInstance) ArchIsX64() bool {
	return strings.EqualFold(t.Cpu, ARCH_X64)
}
func (t *MakeTargetInstance) ArchIsAmd64() bool {
	return strings.EqualFold(t.Cpu, ARCH_AMD64)
}
func (t *MakeTargetInstance) ArchIsArm64() bool {
	return strings.EqualFold(t.Cpu, ARCH_ARM64)
}
func (t *MakeTargetInstance) OSAsString() string {
	return t.Os
}
func (t *MakeTargetInstance) CompilerAsString() string {
	return t.Compiler
}
func (t *MakeTargetInstance) ArchAsString() string {
	return t.Cpu
}
