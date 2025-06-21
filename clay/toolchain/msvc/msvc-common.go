package msvc

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/foundation"
)

// Visual Studio tooling layout

var vcBinMap = map[WinSupportedArch]map[WinSupportedArch]string{
	WinArchx86: {WinArchx86: "", WinArchx64: "x86_amd64", WinArchArm: "x86_arm"},
	WinArchx64: {WinArchx86: "", WinArchx64: "amd64", WinArchArm: "x86_arm"},
}

func getVcBin(hostArch, targetArch WinSupportedArch) (string, error) {
	if hostBinMap, exists := vcBinMap[hostArch]; !exists {
		return "", fmt.Errorf("unknown host architecture: %s", hostArch.String())
	} else if binPath, exists := hostBinMap[targetArch]; !exists {
		return "", fmt.Errorf("unknown target architecture: %s", targetArch.String())
	} else {
		return binPath, nil
	}
}

var vcLibMap = map[WinSupportedArch]map[WinSupportedArch]string{
	WinArchx86: {WinArchx86: "", WinArchx64: "amd64", WinArchArm: "arm"},
	WinArchx64: {WinArchx86: "", WinArchx64: "amd64", WinArchArm: "arm"},
}

func getVcLib(hostArch, targetArch WinSupportedArch) (string, error) {
	if hostLibMap, exists := vcLibMap[hostArch]; !exists {
		return "", fmt.Errorf("unknown host architecture: %s", hostArch.String())
	} else if libPath, exists := hostLibMap[targetArch]; !exists {
		return "", fmt.Errorf("unknown target architecture: %s", targetArch.String())
	} else {
		return libPath, nil
	}
}

type winSdkDirs struct {
	bin      string
	includes []string
	libs     []string
}

// Windows SDK layout
var preWin8SdkDirs = winSdkDirs{
	bin:      "bin",
	includes: []string{"include"},
	libs:     []string{"lib"},
}

// Windows 8 SDK layout
var win8SdkDirs = winSdkDirs{
	bin:      "bin",
	includes: []string{"include"},
	libs:     []string{"lib\\win8\\um"},
}

// Windows 8.1 SDK layout
var win81SdkDirs = winSdkDirs{
	bin:      "bin",
	includes: []string{"include"},
	libs:     []string{"lib\\win6.3\\um"},
}

var preWin8SdkDirsx86 = winSdkDirs{
	bin:      "",
	includes: []string{},
	libs:     []string{},
}

var preWin8SdkDirsx64 = winSdkDirs{
	bin:      "x64",
	includes: []string{},
	libs:     []string{"x64"},
}

var postWin8Sdkx86 = winSdkDirs{
	bin:      "x86",
	includes: []string{"shared", "um"},
	libs:     []string{"x86"},
}

var postWin8Sdkx64 = winSdkDirs{
	bin:      "x64",
	includes: []string{"shared", "um"},
	libs:     []string{"x64"},
}

var postWin8SdkArm = winSdkDirs{
	bin:      "arm",
	includes: []string{"shared", "um"},
	libs:     []string{"arm"},
}

func getPostWin8Sdk(arch WinSupportedArch) *winSdkDirs {
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

var vsSdkMap = map[string]string{
	"9.0":  "v6.0A",
	"10.0": "v7.0A",
	"10.1": "v7.1A",
	"11.0": "v8.0",
	"12.0": "v8.1",
	// The current Visual Studio 2015 download does not include the full Windows
	// 10 SDK, and new Win32 apps created in VS2015 default to using the 8.1 SDK
	"14.0": "v8.1",
}

// Each quadruplet specifies a registry key value that gets us the SDK location,
// followed by a folder structure (for each supported target architecture)
// and finally the corresponding bin, include and lib folder's relative location
type winSdkInfo struct {
	regKey     string
	regValue   string
	sdkDirBase *winSdkDirs
	x86SdkDirs *winSdkDirs
	x64SdkDirs *winSdkDirs
	armSdkDirs *winSdkDirs
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

/*
local function get_host_arch()
  local snative = native.getenv("PROCESSOR_ARCHITECTURE")
  local swow = native.getenv("PROCESSOR_ARCHITEW6432", "")
  if snative == "AMD64" or swow == "AMD64" then
    return "x64"
  elseif snative == "IA64" or swow == "IA64" then
    return "itanium";
  else
    return "x86"
  end
end

function path_combine(path, path_to_append)
  if path == nil then
    return path_to_append
  end
  if path:find("\\$") then
    return path .. path_to_append
  end
  return path .. "\\" .. path_to_append
end

function path_it(maybe_list)
  if type(maybe_list) == "table" then
    return ipairs(maybe_list)
  end
  return ipairs({maybe_list})
end

*/

func getPreWin10Sdk(sdkVersion string, vsVersion VsVersion, targetArch WinSupportedArch) (string, winSdkDirs) {
	result := winSdkDirs{}

	sdk, exists := preWin10SdkMap[sdkVersion]
	if !exists {
		panic("The requested version of Visual Studio isn't supported")
	}

	sdkRoot, err := foundation.QueryRegistryForStringValue(foundation.RegistryKeyLocalMachine, sdk.regKey, sdk.regValue)
	if sdkRoot == "" || err != nil {
		panic("The requested version of the SDK isn't installed")
	}
	sdkRoot = strings.ReplaceAll(sdkRoot, "\\+$", "\\")

	sdkDirBase := sdk.sdkDirBase

	sdkDir := getPostWin8Sdk(targetArch)
	result.bin = sdkRoot + sdkDirBase.bin + "\\" + sdkDir.bin

	result.includes = make([]string, 0, len(sdkDirBase.includes)+len(sdkDir.includes))
	for _, includeDir := range sdkDir.includes {
		result.includes = append(result.includes, sdkRoot+sdkDirBase.includes[0]+"\\"+includeDir)
	}

	result.libs = make([]string, 0, len(sdkDirBase.libs)+len(sdkDir.libs))
	result.libs = append(result.libs, sdkRoot+sdkDirBase.libs[0]+"\\"+sdkDir.libs[0])

	// Windows 10 changed CRT to be split between Windows SDK and VC.
	// It appears that when targeting pre-win10 with VS2015 you should always use
	// use 10.0.10150.0, according to Microsoft.Cpp.Common.props in MSBuild.
	if vsVersion == "14.0" {
		win10SdkRoot, err := foundation.QueryRegistryForStringValue(foundation.RegistryKeyLocalMachine, win10Sdk[0], win10Sdk[1])
		if win10SdkRoot == "" || err != nil {
			panic("The windows 10 UCRT is required when building using Visual studio 2015")
		}
		result.includes = append(result.includes, "include", win10SdkRoot+"Include\\10.0.10150.0\\ucrt")
		result.libs = append(result.libs, win10SdkRoot+"Lib\\10.0.10150.0\\ucrt\\"+sdkDir.libs[0])
	}

	return sdkRoot, result
}

func getWin10Sdk(sdkVersion string, targetArch WinSupportedArch) (string, winSdkDirs) {
	sdkVersion = sdkVersion[2:] // Remove v prefix

	// This only checks if the windows 10 SDK specifically is installed. A
	// 'dir exists' method would be needed here to check if a specific SDK
	// target folder exists.
	sdkRoot, err := foundation.QueryRegistryForStringValue(foundation.RegistryKeyLocalMachine, win10Sdk[0], win10Sdk[1])
	if sdkRoot == "" || err != nil {
		panic("The requested version of the SDK isn't installed")
	}

	result := winSdkDirs{}

	postWin8Sdk := getPostWin8Sdk(targetArch)
	result.bin = sdkRoot + "bin\\" + postWin8Sdk.bin

	result.includes = make([]string, 0, len(postWin8Sdk.includes))
	sdkDirBaseInclude := sdkRoot + "include\\" + sdkVersion + "\\"
	result.includes = append(result.includes, sdkDirBaseInclude+"shared")
	result.includes = append(result.includes, sdkDirBaseInclude+"ucrt")
	result.includes = append(result.includes, sdkDirBaseInclude+"um")

	result.libs = make([]string, 0, len(postWin8Sdk.libs))
	sdkDirBaseLib := sdkRoot + "Lib\\" + sdkVersion + "\\"
	result.libs = append(result.libs, sdkDirBaseLib+"ucrt\\"+postWin8Sdk.libs[0])
	result.libs = append(result.libs, sdkDirBaseLib+"um\\"+postWin8Sdk.libs[0])

	return sdkRoot, result
}

func getSdk(sdkVersion string, vsVersion VsVersion, targetArch WinSupportedArch) (string, winSdkDirs) {
	// All versions using v10.0.xxxxx.x use specific releases of the
	// Win10 SDK. Other versions are assumed to be pre-win10
	if sdkVersion[:6] == "v10.0." {
		return getWin10Sdk(sdkVersion, targetArch)
	}
	return getPreWin10Sdk(sdkVersion, vsVersion, targetArch)
}

// MsDevSetup represents the installation of Microsoft Visual Studio that was found.
// type MsDevSetup struct {
// 	RootPath     string   // The root path of the installation, e.g., "C:\\Program Files (x86)\\Microsoft Visual Studio\\2019\\Community"
// 	Version      string   // The version of the installation, e.g., "16.0"
// 	Arch         string   // The architecture of the installation, e.g., "x86", "x64", "arm64"
// 	BinPath      string   // The path to the bin directory, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64"
// 	CCPath       string   // The path to the cl.exe compiler, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64\\cl.exe"
// 	CXXPath      string   // The path to the cl.exe compiler, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64\\cl.exe"
// 	LIBPath      string   // The path to the lib directory, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\lib\\x64"
// 	LDPath       string   // The path to the link.exe linker, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64\\link.exe"
// 	RCPath       string   // The path to the rc.exe resource compiler, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\bin\\Hostx64\\x64\\rc.exe"
// 	IncludePaths []string //
// 	LibraryPaths []string // The paths to the library directories, e.g., RootPath + "\\VC\\Tools\\MSVC\\14.29.30133\\lib\\x64"
// 	CCOpts       []string // Compiler options, e.g., "/nologo /W3 /O2 /DWIN32 /D_WINDOWS /D_USRDLL /D_MBCS"
// 	CXXOpts      []string // C++ compiler options, e.g., "/nologo /W3 /O2 /DWIN32 /D_WINDOWS /D_USRDLL /D_MBCS"
// }

type MsDevSetup struct {
	CompilerBin   string
	CcOptions     []string
	CxxOptions    []string
	IncludePaths  []string
	ArchiverBin   string
	LinkerBin     string
	Libs          []string
	LibPath       []string
	RcBin         string
	RcOptions     []string
	VsInstallDir  string
	VcInstallDir  string
	DevEnvDir     string
	WindowsSdkDir string
}

func InitMsvcVisualStudio(version VsVersion, _sdkVersion string, _hostArch WinSupportedArch, _targetArch WinSupportedArch) (*MsDevSetup, error) {
	targetArch := getTargetArch(_targetArch)

	sdkVersion := _sdkVersion
	if sdkVersion == "" {
		sdkVersion = string(version)
	}
	if sdkVersion, ok := vsSdkMap[sdkVersion]; ok {
		sdkVersion = string(sdkVersion)
	}

	// We will find any edition of VS (including Express) here
	vsRoot, err := foundation.QueryRegistryForStringValue(foundation.RegistryKeyLocalMachine, "SOFTWARE\\Microsoft\\VisualStudio\\SxS\\VS7", string(version))
	if vsRoot == "" || err != nil {
		// This is necessary for supporting the "Visual C++ Build Tools", which includes only the Compiler & SDK (not Visual Studio)
		vcRoot, err := foundation.QueryRegistryForStringValue(foundation.RegistryKeyLocalMachine, "SOFTWARE\\Microsoft\\VisualStudio\\SxS\\VC7", string(version))
		if vcRoot != "" && err == nil {
			vsRoot = strings.ReplaceAll(vcRoot, "\\VC\\$", "\\")
		}
	}
	if vsRoot == "" {
		return nil, fmt.Errorf("Visual Studio [Version %s] isn't installed. Please use a different Visual Studio version", string(version))
	}
	vsRoot = strings.TrimSuffix(vsRoot, "\\")

	vcBin, err := getVcBin(_hostArch, targetArch)
	if err != nil {
		return nil, err
	}
	vcBin = vsRoot + "\\vc\\bin\\" + vcBin

	vcLib, err := getVcLib(_hostArch, targetArch)
	if err != nil {
		return nil, err
	}
	vcLib = vsRoot + "\\vc\\lib\\" + vcLib

	//
	// Now fix up the SDK
	//
	sdkRoot, sdkDirs := getSdk(sdkVersion, version, targetArch)

	env := &MsDevSetup{}
	env.WindowsSdkDir = sdkRoot

	//
	// Tools
	//
	env.CompilerBin = filepath.Join(vcBin, "cl.exe")
	env.ArchiverBin = filepath.Join(vcBin, "lib.exe")
	env.LinkerBin = filepath.Join(vcBin, "link.exe")
	env.RcBin = filepath.Join(sdkDirs.bin, "rc.exe")

	if sdkVersion == "9.0" {
		// clear the "/nologo" option (it was first added in VS2010)
		env.RcOptions = []string{}
	}

	if version == "12.0" || version == "14.0" {
		// Force MSPDBSRV.EXE
		env.CcOptions = []string{"/FS"}
		env.CxxOptions = []string{"/FS"}
	}

	// Wire-up the external environment
	env.VsInstallDir = vsRoot
	env.VcInstallDir = vsRoot + "\\VC"
	env.DevEnvDir = vsRoot + "\\Common7\\IDE"

	include := make([]string, 0, len(sdkDirs.includes)+2)

	for _, v := range sdkDirs.includes {
		include = append(include, v)
	}
	include = append(include, vsRoot+"\\VC\\ATLMFC\\INCLUDE")
	include = append(include, vsRoot+"\\VC\\INCLUDE")

	// if MFC isn't installed with VS
	// the linker will throw an error when looking for libs
	mfcLibPath := vsRoot + "\\VC\\ATLMFC\\lib\\" + vcLib
	if !foundation.DirExists(mfcLibPath) {
		return nil, fmt.Errorf("MFC libraries not found in %s", mfcLibPath)
	}

	libPaths := make([]string, 0, len(sdkDirs.libs)+2)
	for _, libDir := range sdkDirs.libs {
		libPaths = append(libPaths, libDir)
	}
	libPaths = append(libPaths, mfcLibPath)
	libPaths = append(libPaths, vcLib)

	env.IncludePaths = include

	env.Libs = libPaths
	env.LibPath = libPaths

	// Extend PATH with the necessary directories
	path := make([]string, 0, 5)
	path = append(path, sdkRoot)
	path = append(path, vsRoot+"\\Common7\\IDE")

	if _hostArch == WinArchx86 {
		path = append(path, vsRoot+"\\VC\\Bin")
	} else if _hostArch == WinArchx64 {
		path = append(path, vsRoot+"\\VC\\Bin\\amd64")
	} else if _hostArch == WinArchArm {
		path = append(path, vsRoot+"\\VC\\Bin\\arm")
	}

	// env.Append("PATH", path...)

	return env, nil
}
