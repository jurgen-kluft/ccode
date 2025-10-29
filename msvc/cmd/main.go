package main

import (
	"fmt"

	"github.com/jurgen-kluft/ccode/msvc"
)

func main() {

	// Detect the Microsoft Visual Studio versions and determine all necessary paths to libraries, include and executables

	versions := []msvc.VsVersion{
		msvc.VsVersion2013,
		msvc.VsVersion2015,
		msvc.VsVersion2017,
		msvc.VsVersion2019,
		msvc.VsVersion2022,
	}

	for _, version := range versions {
		// func func InitMsvcVisualStudio(_vsVersion VsVersion, _sdkVersion string, _hostArch WinSupportedArch, _targetArch WinSupportedArch) (*MsvcEnvironment, error) {(_vsVersion VsVersion, _sdkVersion string, _hostArch WinSupportedArch, _targetArch WinSupportedArch) (*MsvcEnvironment, error) {
		msvc, err := msvc.InitMsvcVisualStudio(version, "", msvc.WinArchx64, msvc.WinArchx64)
		if err != nil {
			fmt.Println(err)
			continue
		}
		msvc.Print()
	}
}
