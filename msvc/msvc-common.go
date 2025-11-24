package msvc

import (
	"fmt"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
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

// Visual Studio 2008      9.0
// Visual Studio 2008      9.0

// Visual Studio 2010      10.0
// Visual Studio 2010      10.0

// Visual Studio 2012      11.0
// Visual Studio 2012      11.0

// Visual Studio 2013      12.0
// Visual Studio 2013      12.0

// Visual Studio 2015      14.0
// Visual Studio 2015      14.0

// Visual Studio 2017      15.0
// Visual Studio 2017      15.9.11

// Visual Studio 2019      16.0.0
// Visual Studio 2019      16.11.35

// Visual Studio 2022      17.0.1
// Visual Studio 2022      17.13.6

var vsSdkMap = map[string]string{
	"9.0":  "v6.0A",
	"2008": "v6.0A",

	"10.0": "v7.0A",
	"10.1": "v7.1A",
	"2010": "v7.0A",

	"11.0": "v8.0",
	"2012": "v8.0",

	"12.0": "v8.1",
	"2013": "v8.1",

	// The current Visual Studio 2015 download does not include the full Windows
	// 10 SDK, and new Win32 apps created in VS2015 default to using the 8.1 SDK
	"14.0": "v8.1",
	"2015": "v8.1",
}

// MsvcEnvironment represents the installation of Microsoft Visual Studio that was found.
type MsvcEnvironment struct {
	Installed      bool
	Version        VsVersion
	Product        VsProduct
	CompilerPath   string
	CompilerBin    string
	CcOptions      []string
	CxxOptions     []string
	IncludePaths   []string
	ArchiverPath   string
	ArchiverBin    string
	LinkerPath     string
	LinkerBin      string
	Libs           []string
	LibPaths       []string
	RcPath         string
	RcBin          string
	RcOptions      []string
	VcToolsVersion string
	VsInstallDir   string
	VcInstallDir   string
	DevEnvDir      string
	Path           []string // Additional PATH entries to be set
}

func NewMsvcEnvironment(version VsVersion, product VsProduct) *MsvcEnvironment {
	return &MsvcEnvironment{
		Installed:    false,
		Version:      version,
		Product:      product,
		CompilerPath: "",
		CompilerBin:  "",
		CcOptions:    []string{},
		CxxOptions:   []string{},
		IncludePaths: []string{},
		ArchiverPath: "",
		ArchiverBin:  "",
		LinkerPath:   "",
		LinkerBin:    "",
		Libs:         []string{},
		LibPaths:     []string{},
		RcPath:       "",
		RcBin:        "",
		RcOptions:    []string{},
		VsInstallDir: "",
		VcInstallDir: "",
		DevEnvDir:    "",
		Path:         []string{},
	}
}

func Find(_vsVersion VsVersion, _sdkVersion string, _hostArch WinSupportedArch, _targetArch WinSupportedArch) (*MsvcEnvironment, error) {
	targetArch := getTargetArch(_targetArch)

	if _vsVersion >= VsVersion2017 {
		msvcVersion := NewMsvcVersion()
		msvcVersion.vsVersion = _vsVersion
		msvcVersion.vsProduct = vsProductProfessional
		msvcVersion.hostArch = _hostArch
		msvcVersion.targetArch = targetArch
		return SetupMsvcVersion(msvcVersion, false)
	}

	ok := false
	sdkVersion := _sdkVersion
	if sdkVersion, ok = vsSdkMap[sdkVersion]; ok {
		sdkVersion = string(sdkVersion)
	} else if sdkVersion, ok = vsSdkMap[_vsVersion.String()]; ok {
		sdkVersion = string(sdkVersion)
	} else {
		return nil, fmt.Errorf("the requested version of the SDK isn't supported: %s", _sdkVersion)
	}

	// We will find any edition of VS (including Express) here
	vsRoot, err := corepkg.QueryRegistryForStringValue(corepkg.RegistryKeyLocalMachine, "SOFTWARE\\Microsoft\\VisualStudio\\SxS\\VS7", string(_vsVersion))
	if vsRoot == "" || err != nil {
		// This is necessary for supporting the "Visual C++ Build Tools", which includes only the Compiler & SDK (not Visual Studio)
		vcRoot, err := corepkg.QueryRegistryForStringValue(corepkg.RegistryKeyLocalMachine, "SOFTWARE\\Microsoft\\VisualStudio\\SxS\\VC7", string(_vsVersion))
		if vcRoot != "" && err == nil {
			vsRoot = strings.ReplaceAll(vcRoot, "\\VC\\$", "\\")
		}
	}

	// TODO, find out the exact edition from the registry (Express, Community, Professional, Enterprise)
	product := vsProductUnknown

	msdev := NewMsvcEnvironment(_vsVersion, product)

	if vsRoot == "" {
		return msdev, nil
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
	// Installed
	//
	msdev.Installed = corepkg.DirExists(vcBin) && corepkg.DirExists(vcLib)

	//
	// Tools
	//
	msdev.CompilerPath = vcBin
	msdev.CompilerBin = "cl.exe"
	msdev.ArchiverPath = vcBin
	msdev.ArchiverBin = "lib.exe"
	msdev.LinkerPath = vcBin
	msdev.LinkerBin = "link.exe"
	msdev.RcBin = "rc.exe"

	if sdkVersion == "9.0" {
		// clear the "/nologo" option (it was first added in VS2010)
		msdev.RcOptions = []string{}
	}

	if _vsVersion == "12.0" || _vsVersion == "14.0" {
		// Force MSPDBSRV.EXE
		msdev.CcOptions = []string{"/FS"}
		msdev.CxxOptions = []string{"/FS"}
	}

	// Wire-up the external environment
	msdev.VsInstallDir = vsRoot
	msdev.VcInstallDir = vsRoot + "\\VC"
	msdev.DevEnvDir = vsRoot + "\\Common7\\IDE"

	include := make([]string, 0)

	include = append(include, vsRoot+"\\VC\\ATLMFC\\INCLUDE")
	include = append(include, vsRoot+"\\VC\\INCLUDE")

	// if MFC isn't installed with VS
	// the linker will throw an error when looking for libs
	mfcLibPath := vsRoot + "\\VC\\ATLMFC\\lib\\" + vcLib
	if !corepkg.DirExists(mfcLibPath) {
		return nil, fmt.Errorf("MFC libraries not found in %s", mfcLibPath)
	}

	libPaths := make([]string, 0)
	libPaths = append(libPaths, mfcLibPath)
	libPaths = append(libPaths, vcLib)

	msdev.IncludePaths = include

	msdev.Libs = libPaths
	msdev.LibPaths = libPaths

	// Extend PATH with the necessary directories
	path := make([]string, 0, 5)
	path = append(path, vsRoot+"\\Common7\\IDE")

	switch _hostArch {
	case WinArchx86:
		path = append(path, vsRoot+"\\VC\\Bin")
	case WinArchx64:
		path = append(path, vsRoot+"\\VC\\Bin\\amd64")
	case WinArchArm:
		path = append(path, vsRoot+"\\VC\\Bin\\arm")
	}

	msdev.Path = path

	return msdev, nil
}

func (msvc *MsvcEnvironment) Print() {
	fmt.Println("========================================")
	fmt.Println("Microsoft Visual Studio Environment:")
	fmt.Printf("  Installed: %v\n", msvc.Installed)
	fmt.Printf("  Version: %s\n", msvc.Version.String())
	fmt.Printf("  Product: %s\n", msvc.Product.String())
	fmt.Printf("  Compiler Path: %s\n", msvc.CompilerPath)
	fmt.Printf("  Compiler Bin: %s\n", msvc.CompilerBin)
	fmt.Printf("  Cc Options: %v\n", msvc.CcOptions)
	fmt.Printf("  Cxx Options: %v\n", msvc.CxxOptions)
	fmt.Printf("  Include Paths: %v\n", msvc.IncludePaths)
	fmt.Printf("  Archiver Path: %s\n", msvc.ArchiverPath)
	fmt.Printf("  Archiver Bin: %s\n", msvc.ArchiverBin)
	fmt.Printf("  Linker Path: %s\n", msvc.LinkerPath)
	fmt.Printf("  Linker Bin: %s\n", msvc.LinkerBin)
	fmt.Printf("  Libs: %v\n", msvc.Libs)
	fmt.Printf("  Lib Paths: %v\n", msvc.LibPaths)
	fmt.Printf("  Rc Path: %s\n", msvc.RcPath)
	fmt.Printf("  Rc Bin: %s\n", msvc.RcBin)
	fmt.Printf("  Rc Options: %v\n", msvc.RcOptions)
	fmt.Printf("  Vc Tools Version: %s\n", msvc.VcToolsVersion)
	fmt.Printf("  Vs Install Dir: %s\n", msvc.VsInstallDir)
	fmt.Printf("  Vc Install Dir: %s\n", msvc.VcInstallDir)
	fmt.Printf("  DevEnv Dir: %s\n", msvc.DevEnvDir)
	fmt.Printf("  Path: %v\n", msvc.Path)

	fmt.Println()
}
