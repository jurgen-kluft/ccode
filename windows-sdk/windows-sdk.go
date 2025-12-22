package winsdk

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

type WindowsSdkLayout struct {
	Bin      string
	Includes []string
	LibPaths []string
}

var preWin8SdkDirs = WindowsSdkLayout{Bin: "bin", Includes: []string{"include"}, LibPaths: []string{"lib"}}
var win8SdkDirs = WindowsSdkLayout{Bin: "bin", Includes: []string{"include"}, LibPaths: []string{"lib\\win8\\um"}}
var win81SdkDirs = WindowsSdkLayout{Bin: "bin", Includes: []string{"include"}, LibPaths: []string{"lib\\win6.3\\um"}}
var preWin8SdkDirsx86 = WindowsSdkLayout{Bin: "", Includes: []string{}, LibPaths: []string{}}
var preWin8SdkDirsx64 = WindowsSdkLayout{Bin: "x64", Includes: []string{}, LibPaths: []string{"x64"}}
var postWin8Sdkx86 = WindowsSdkLayout{Bin: "x86", Includes: []string{"shared", "um"}, LibPaths: []string{"x86"}}
var postWin8Sdkx64 = WindowsSdkLayout{Bin: "x64", Includes: []string{"shared", "um"}, LibPaths: []string{"x64"}}
var postWin8SdkArm = WindowsSdkLayout{Bin: "arm", Includes: []string{"shared", "um"}, LibPaths: []string{"arm"}}

func getPostWin8Sdk(arch WinSupportedArch) *WindowsSdkLayout {
	switch arch {
	case WinArchx86:
		return &postWin8Sdkx86
	case WinArchx64:
		return &postWin8Sdkx64
	case WinArchArm:
		return &postWin8SdkArm
	case WinArchArm64:
		return &postWin8SdkArm
	}
	return &postWin8SdkArm
}

// Each quadruplet specifies a registry key value that gets us the SDK location,
// followed by a folder structure (for each supported target architecture)
// and finally the corresponding bin, include and lib folder's relative location
type winSdkInfo struct {
	regKey     string
	regValue   string
	sdkDirBase *WindowsSdkLayout
	x86SdkDirs *WindowsSdkLayout
	x64SdkDirs *WindowsSdkLayout
	armSdkDirs *WindowsSdkLayout
}

var preWin10SdkMap = map[string]winSdkInfo{
	"v6.0A": {"SOFTWARE\\Microsoft\\Microsoft SDKs\\Windows\\v6.0A", "InstallationFolder", &preWin8SdkDirs, &preWin8SdkDirsx86, &preWin8SdkDirsx64, nil},
	"v7.0A": {"SOFTWARE\\Microsoft\\Microsoft SDKs\\Windows\\v7.0A", "InstallationFolder", &preWin8SdkDirs, &preWin8SdkDirsx86, &preWin8SdkDirsx64, nil},
	"v7.1A": {"SOFTWARE\\Microsoft\\Microsoft SDKs\\Windows\\v7.1A", "InstallationFolder", &preWin8SdkDirs, &preWin8SdkDirsx86, &preWin8SdkDirsx64, nil},
	"v8.0":  {"SOFTWARE\\Microsoft\\Windows Kits\\Installed Roots", "KitsRoot", &win8SdkDirs, &postWin8Sdkx86, &postWin8Sdkx64, &postWin8SdkArm},
	"v8.1":  {"SOFTWARE\\Microsoft\\Windows Kits\\Installed Roots", "KitsRoot81", &win81SdkDirs, &postWin8Sdkx86, &postWin8Sdkx64, &postWin8SdkArm},
}

var win10Sdk = []string{
	"SOFTWARE\\Microsoft\\Windows Kits\\Installed Roots", "KitsRoot10",
}

func getPreWin10Sdk(sdkVersion string, targetArch WinSupportedArch) (string, WindowsSdkLayout, error) {
	result := WindowsSdkLayout{}

	sdk, exists := preWin10SdkMap[sdkVersion]
	if !exists {
		return "", result, fmt.Errorf("the requested version of Windows isn't supported: %s", sdkVersion)
	}

	sdkRoot, err := corepkg.QueryRegistryForStringValue(corepkg.RegistryKeyLocalMachine, sdk.regKey, sdk.regValue)
	if sdkRoot == "" || err != nil {
		return "", result, fmt.Errorf("the requested version of the SDK isn't installed: %s", sdkVersion)
	}
	sdkRoot = strings.ReplaceAll(sdkRoot, "\\+$", "\\")

	sdkDirBase := sdk.sdkDirBase

	sdkDir := getPostWin8Sdk(targetArch)
	result.Bin = sdkRoot + sdkDirBase.Bin + "\\" + sdkDir.Bin

	result.Includes = make([]string, 0, len(sdkDirBase.Includes)+len(sdkDir.Includes))
	for _, includeDir := range sdkDir.Includes {
		result.Includes = append(result.Includes, sdkRoot+sdkDirBase.Includes[0]+"\\"+includeDir)
	}

	result.LibPaths = make([]string, 0, len(sdkDirBase.LibPaths)+len(sdkDir.LibPaths))
	result.LibPaths = append(result.LibPaths, sdkRoot+sdkDirBase.LibPaths[0]+"\\"+sdkDir.LibPaths[0])

	// Windows 10 changed CRT to be split between Windows SDK and VC.
	// It appears that when targeting pre-win10 with VS2015 you should always use
	// use 10.0.10150.0, according to Microsoft.Cpp.Common.props in MSBuild.
	// if vsVersion == "14.0" {
	// 	win10SdkRoot, err := corepkg.QueryRegistryForStringValue(corepkg.RegistryKeyLocalMachine, win10Sdk[0], win10Sdk[1])
	// 	if win10SdkRoot == "" || err != nil {
	// 		panic("The windows 10 UCRT is required when building using Visual studio 2015")
	// 	}
	// 	result.Includes = append(result.includes, "include", win10SdkRoot+"Include\\10.0.10150.0\\ucrt")
	// 	result.libs = append(result.libs, win10SdkRoot+"Lib\\10.0.10150.0\\ucrt\\"+sdkDir.libs[0])
	// }

	return sdkRoot, result, nil
}

func getWin10Sdk(sdkVersion string, targetArch WinSupportedArch) (string, WindowsSdkLayout, error) {
	result := WindowsSdkLayout{}

	// This only checks if the windows 10 SDK specifically is installed. A
	// 'dir exists' method would be needed here to check if a specific SDK
	// target folder exists.
	sdkRoot, err := corepkg.QueryRegistryForStringValue(corepkg.RegistryKeyLocalMachine, win10Sdk[0], win10Sdk[1])
	if sdkRoot == "" || err != nil {
		return "", result, fmt.Errorf("The requested version of the SDK isn't installed")
	}

	postWin8Sdk := getPostWin8Sdk(targetArch)
	result.Bin = sdkRoot + "bin\\" + postWin8Sdk.Bin

	result.Includes = make([]string, 0, len(postWin8Sdk.Includes))
	sdkDirBaseInclude := sdkRoot + "include\\" + sdkVersion + "\\"
	result.Includes = append(result.Includes, sdkDirBaseInclude+"shared")
	result.Includes = append(result.Includes, sdkDirBaseInclude+"ucrt")
	result.Includes = append(result.Includes, sdkDirBaseInclude+"um")

	result.LibPaths = make([]string, 0, len(postWin8Sdk.LibPaths))
	sdkDirBaseLib := sdkRoot + "Lib\\" + sdkVersion + "\\"
	result.LibPaths = append(result.LibPaths, sdkDirBaseLib+"ucrt\\"+postWin8Sdk.LibPaths[0])
	result.LibPaths = append(result.LibPaths, sdkDirBaseLib+"um\\"+postWin8Sdk.LibPaths[0])

	return sdkRoot, result, nil
}

func getSdk(sdkVersion string, targetArch WinSupportedArch) (string, WindowsSdkLayout, error) {
	// All versions using v10.0.xxxxx.x use specific releases of the
	// Win10 SDK. Other versions are assumed to be pre-win10
	if strings.HasPrefix(sdkVersion, "10.0.") || strings.HasPrefix(sdkVersion, "v10.0.") {
		return getWin10Sdk(sdkVersion, targetArch)
	}
	return getPreWin10Sdk(sdkVersion, targetArch)
}

type WinSupportedArch string

const (
	WinArchx86   WinSupportedArch = "x86"
	WinArchx64   WinSupportedArch = "x64"
	WinArchArm   WinSupportedArch = "arm"
	WinArchArm64 WinSupportedArch = "arm64"
)

func (a WinSupportedArch) String() string {
	return string(a)
}

var supportedArchMappings = map[string]WinSupportedArch{
	"x86":   WinArchx86,
	"x64":   WinArchx64,
	"arm":   WinArchArm,
	"arm64": WinArchArm64,
	"amd64": WinArchx64,
}

type WinAppPlatform string

const (
	WinAppDesktop WinAppPlatform = "Desktop" // default
	WinAppUWP     WinAppPlatform = "UWP"     // Universal Windows Platform
	WinAppOneCore WinAppPlatform = "OneCore" // OneCore (Windows 10, Windows 11, Xbox, HoloLens)
)

var supportedAppPlatforms = map[string]WinAppPlatform{
	"desktop": WinAppDesktop,
	"uwp":     WinAppUWP,
	"onecore": WinAppOneCore,
}

type WindowsSDK struct {
	Version string
	Dir     string // Directory of the Windows SDK
	Layout  *WindowsSdkLayout
}

func newWindowsSDK(version string, dir string, layout WindowsSdkLayout) *WindowsSDK {
	sdk := &WindowsSDK{
		Version: version,
		Dir:     dir,
		Layout:  &layout,
	}
	return sdk
}

func (w *WindowsSDK) Print() {
	fmt.Printf("----------------------------------------\n")
	fmt.Printf("Windows SDK Version: %s\n", w.Version)
	fmt.Printf("Windows SDK Directory: %s\n", w.Dir)

	fmt.Printf("Binary Directory: %s\n", w.Layout.Bin)

	fmt.Println("Include Directories:")
	for _, includeDir := range w.Layout.Includes {
		fmt.Printf("  %s\n", includeDir)
	}

	fmt.Println("Library Directories:")
	for _, libDir := range w.Layout.LibPaths {
		fmt.Printf("  %s\n", libDir)
	}

	fmt.Println()
}

type WindowsSDKs []*WindowsSDK

func (w WindowsSDKs) HasVersion(version string) bool {
	for _, v := range w {
		if v.Version == version {
			return true
		}
	}
	return false
}

func (w WindowsSDKs) GetLatestVersion() string {
	if len(w) == 0 {
		return ""
	}
	return w[len(w)-1].Version
}

func SelectLatestWindowsSDK(sdks WindowsSDKs) *WindowsSDK {
	if len(sdks) == 0 {
		return nil
	}

	versionIndex := 0
	versionCurrent := 0
	for i, sdk := range sdks {
		version, _ := strconv.Atoi(strings.ReplaceAll(sdk.Version, ".", ""))
		if version > versionCurrent {
			versionCurrent = version
			versionIndex = i
		}
	}

	return sdks[versionIndex]
}

// From Visual Studio 2017 onwards, this is the recommended way to find the Windows SDK.
func Find(winAppPlatform WinAppPlatform) (WindowsSDKs, error) {
	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/core/winsdk.bat#L63
	//   HKLM\SOFTWARE\Wow6432Node
	//   HKCU\SOFTWARE\Wow6432Node (ignored)
	//   HKLM\SOFTWARE             (ignored)
	//   HKCU\SOFTWARE             (ignored)
	winsdkKey := `SOFTWARE\Wow6432Node\Microsoft\Microsoft SDKs\Windows\\v10.0`
	winsdkDir, err := corepkg.QueryRegistryForStringValue(corepkg.RegistryKeyLocalMachine, winsdkKey, "InstallationFolder")
	if err != nil {
		return nil, err
	}
	// fmt.Printf("Detected Windows SDK installation folder: %s\n", winsdkDir)

	winsdkVersions := []string{}

	// Due to the SDK installer changes beginning with the 10.0.xxxxx.x versions,
	checkFile := "winsdkver.h"
	if winAppPlatform == WinAppUWP {
		checkFile = "Windows.h"
	}

	dirs, err := corepkg.DirList(filepath.Join(winsdkDir, "Include"))
	if err != nil {
		return nil, fmt.Errorf("failed to list Windows SDK include directory: %v", err)
	}

	for _, winsdkVersion := range dirs {
		if strings.HasPrefix(winsdkVersion, "10.") {
			testPath := filepath.Join(winsdkDir, "Include", winsdkVersion, "um", checkFile)
			if corepkg.FileExists(testPath) {
				winsdkVersions = append(winsdkVersions, winsdkVersion)
				//fmt.Printf("Detected Windows SDK version: %s\n", winsdkVersion)
			}
		}
	}

	if len(winsdkVersions) == 0 {
		return nil, fmt.Errorf("no Windows SDK versions found in %s", winsdkDir)
	}

	windowsSDKs := []*WindowsSDK{}
	for _, version := range winsdkVersions {
		dir, layout, err := getSdk(version, WinArchx64)
		if err != nil {
			return nil, fmt.Errorf("failed to get Windows SDK layout for version %s: %v", version, err)
		}
		winsdk := newWindowsSDK(version, dir, layout)
		windowsSDKs = append(windowsSDKs, winsdk)
	}

	return windowsSDKs, nil
}
