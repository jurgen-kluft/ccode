package denv

import "runtime"

func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}
