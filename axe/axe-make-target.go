package axe

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

func NewMakeTargetMacOS() MakeTarget {
	arch := runtime.GOARCH
	return NewMakeTarget(OS_MAC, COMPILER_CLANG, arch)
}

func NewMakeTargetLinux() MakeTarget {
	arch := runtime.GOARCH
	return NewMakeTarget(OS_LINUX, COMPILER_CLANG, arch)
}

func NewMakeTargetWindows() MakeTarget {
	return NewMakeTarget(OS_WINDOWS, COMPILER_VC, ARCH_X64)
}

func NewDefaultMakeTarget() MakeTarget {
	// use golang internal 'runtime' variables to determine the default target
	if strings.Contains(runtime.GOOS, "windows") {
		return NewMakeTargetWindows()
	} else if strings.Contains(runtime.GOOS, "darwin") {
		return NewMakeTargetMacOS()
	}
	return NewMakeTargetLinux()
}

const (
	OS_LINUX   = "linux"
	OS_WINDOWS = "windows"
	OS_MAC     = "mac"
	OS_IOS     = "ios"
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
