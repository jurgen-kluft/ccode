package msvc

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

// How does this work?
//
// Visual Studio is installed in one of two locations depending on version and product
//   2017-2019 : C:\Program Files (x86)\Microsoft Visual Studio\<Version>\<Product>
//   2022-     : C:\Program Files\Microsoft Visual Studio\<Version>\<Product>
//
// This will be the value for the VSINSTALLDIR environment variable.
//
// Since it is possible to install any combination of Visual Studio products
// we have to check all of them. The first product with a VC tools version
// will be used unless you ask for a specific product and/or VC tools version.
//
// The VsDevCmd.bat script is used to initialize the Developer Command Prompt for VS.
// It will unconditionally call the bat files inside "%VSINSTALLDIR%\Common7\Tools\vsdevcmd\core"
// followed by "%VSINSTALLDIR%\Common7\Tools\vsdevcmd\ext" unless run with -no_ext.
//
// The only two bat files that we care about are:
//   "%VSINSTALLDIR%\Common7\Tools\vsdevcmd\core\winsdk.bat"
//   "%VSINSTALLDIR%\Common7\Tools\vsdevcmd\ext\vcvars.bat"
//
// And the rest of this is just reverse engineered from these bat scripts.

type VsVersion string

const (
	VsVersion2013 VsVersion = "2013"
	VsVersion2015 VsVersion = "2015"
	VsVersion2017 VsVersion = "2017"
	VsVersion2019 VsVersion = "2019"
	VsVersion2022 VsVersion = "2022"
)

func (v VsVersion) String() string {
	return string(v)
}

// Note that while Community, Professional and Enterprise products are installed
// in C:\Program Files while BuildTools are always installed in C:\Program Files (x86)
var vsDefaultPath = "C:\\Program Files (x86)\\Microsoft Visual Studio"

// Add new Visual Studio versions here and update vsDefaultVersion
var vsDefaultPaths = map[VsVersion]string{
	VsVersion2017: vsDefaultPath,
	VsVersion2019: vsDefaultPath,
	VsVersion2022: "C:\\Program Files\\Microsoft Visual Studio",
}

var vsDefaultVersion = VsVersion2022

type VsProduct string

const (
	vsProductUnknown      VsProduct = "Unknown"
	vsProductBuildTools   VsProduct = "BuildTools" // default
	vsProductCommunity    VsProduct = "Community"
	vsProductProfessional VsProduct = "Professional"
	vsProductEnterprise   VsProduct = "Enterprise"
)

func (p VsProduct) String() string {
	return string(p)
}

var vsProducts = []VsProduct{
	vsProductBuildTools,
	vsProductCommunity,
	vsProductProfessional,
	vsProductEnterprise,
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

func getArch(arch WinSupportedArch) WinSupportedArch {
	if arch2, ok := supportedArchMappings[strings.ToLower(arch.String())]; ok {
		return arch2
	}
	return WinArchx64
}

// getHostArch Gets the host architecture from the options, default to x64
func getHostArch(hostArch WinSupportedArch) WinSupportedArch {
	if hostArch == "" {
		if runtime.GOOS == "windows" {
			switch runtime.GOARCH {
			case "amd64", "x86_64":
				return WinArchx64
			case "arm64":
				return WinArchArm64
			case "386", "i386":
				return WinArchx86
			case "arm":
				return WinArchArm
			}
		}
		return WinArchx64 // If not specified, default to x64
	}

	return getArch(hostArch)

}

func getTargetArch(targetArch WinSupportedArch) WinSupportedArch {
	if targetArch == "" {
		return WinArchx64 // If not specified, default to x64
	}
	return getArch(targetArch)
}

type VcTools struct {
	vsInstallDir   string
	vcInstallDir   string
	vcToolsVersion string
}

type InstalledVcTools struct {
	vcTools []*VcTools
}

func NewInstalledVcTools() *InstalledVcTools {
	return &InstalledVcTools{
		vcTools: []*VcTools{},
	}
}

func (v *InstalledVcTools) find(vsPath string, VsVersion VsVersion, VsProduct VsProduct, targetVcToolsVersion string) error {
	if vsPath == "" {
		if VsProduct == vsProductBuildTools {
			vsPath = vsDefaultPath
		} else {
			vsPath = vsDefaultPaths[VsVersion]
			if vsPath == "" {
				vsPath = vsDefaultPaths[vsDefaultVersion]
			}
		}
	}

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L729

	// we ignore Microsoft.VCToolsVersion.v143.default.txt and use Microsoft.VCToolsVersion.default.txt
	// unless a specific VC tools version was requested

	vsInstallDir := filepath.Join(vsPath, VsVersion.String(), VsProduct.String())
	vcInstallDir := filepath.Join(vsInstallDir, "VC")

	var vcToolsVersion string
	if targetVcToolsVersion == "" {
		versionFile := filepath.Join(vcInstallDir, "Auxiliary", "Build", "Microsoft.VCToolsVersion.default.txt")
		data, err := corepkg.FileOpenReadClose(versionFile)
		if err == nil {
			lines := strings.Split(string(data), "\n")
			if len(lines) > 0 {
				vcToolsVersion = strings.TrimSpace(lines[0])
			}
		}
	} else {
		vcToolsVersion = targetVcToolsVersion
	}

	vcTools := &VcTools{
		vsInstallDir:   vsInstallDir,
		vcInstallDir:   vcInstallDir,
		vcToolsVersion: vcToolsVersion,
	}
	v.vcTools = append(v.vcTools, vcTools)

	if vcToolsVersion != "" {
		testPath := filepath.Join(vcInstallDir, "Tools", "MSVC", vcToolsVersion, "include", "vcruntime.h")
		if corepkg.FileExists(testPath) {
			v.vcTools = append(v.vcTools, vcTools)
			return nil
		}
	}

	return nil
}

type MsvcVersion struct {
	vsPath               string
	vsVersion            VsVersion        // default is 2022
	vsProduct            VsProduct        // default is BuildTools
	hostArch             WinSupportedArch // default is x64
	targetArch           WinSupportedArch // default is x64
	winAppPlatform       WinAppPlatform   // default is Desktop
	targetWinSDKVersion  string           // Windows SDK version
	targetVcToolsVersion string           // Visual C++ tools version
	atlMfc               string
}

func NewMsvcVersion() *MsvcVersion {
	return &MsvcVersion{
		vsPath:               "",
		vsVersion:            vsDefaultVersion,
		vsProduct:            vsProductUnknown,
		hostArch:             getHostArch(""),
		targetArch:           getTargetArch(""),
		winAppPlatform:       WinAppDesktop,
		targetWinSDKVersion:  "",
		targetVcToolsVersion: "",
		atlMfc:               "false", // default is false
	}
}

func setupMsvcVersion(msvcVersion *MsvcVersion, useClang bool) (msdev *MsvcEnvironment, err error) {

	if vsDefaultPath, ok := vsDefaultPaths[msvcVersion.vsVersion]; !ok {
		corepkg.LogWarnf("Visual Studio %s has not been tested and might not work out of the box", msvcVersion.vsVersion.String())
	} else if msvcVersion.vsPath == "" {
		msvcVersion.vsPath = vsDefaultPath
	}

	// // ------------------
	// // Windows SDK
	// // ------------------

	// // file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/core/winsdk.bat#L513

	// winSdk, err := FindWindowsSDK(msvcVersion.WinAppPlatform)
	// if err != nil {
	// 	return nil, err
	// }
	// winsdkDir := winSdk.Dir
	// winsdkVersion := msvcVersion.targetWinSDKVersion
	// if len(winsdkVersion) > 0 && !winSdk.HasVersion(winsdkVersion) {
	// 	return nil, fmt.Errorf("Windows SDK version %s not found in %s", winsdkVersion, winSdk.Dir)
	// } else {
	// 	winsdkVersion = winSdk.GetLatestVersion()
	// }

	// msdev = NewMsvcEnvironment()

	// msdev.Path = append(msdev.Path, filepath.Join(winsdkDir, "bin", winsdkVersion, msvcVersion.hostArch.String()))

	// msdev.IncludePaths = append(msdev.IncludePaths, filepath.Join(winsdkDir, "Include", winsdkVersion, "shared"))
	// msdev.IncludePaths = append(msdev.IncludePaths, filepath.Join(winsdkDir, "Include", winsdkVersion, "um"))
	// msdev.IncludePaths = append(msdev.IncludePaths, filepath.Join(winsdkDir, "Include", winsdkVersion, "winrt")) // WinRT (used by DirectX 12 headers)

	// // We assume that the Universal CRT isn't loaded from a different directory
	// ucrtSdkDir := winsdkDir
	// ucrtVersion := winsdkVersion

	// msdev.IncludePaths = append(msdev.IncludePaths, filepath.Join(ucrtSdkDir, "Include", ucrtVersion, "ucrt"))

	// msdev.Libs = append(msdev.Libs, filepath.Join(winsdkDir, "Lib", winsdkVersion, "um", msvcVersion.targetArch.String()))
	// msdev.Libs = append(msdev.Libs, filepath.Join(ucrtSdkDir, "Lib", ucrtVersion, "ucrt", msvcVersion.targetArch.String()))

	// // Skip if the Universal CRT is loaded from the same path as the Windows SDK
	// if ucrtSdkDir != winsdkDir || ucrtVersion != winsdkVersion {
	// 	msdev.Libs = append(msdev.Libs, filepath.Join(ucrtSdkDir, "Lib", ucrtVersion, "um", msvcVersion.targetArch.String()))
	// }

	// -------------------
	// Visual C++
	// -------------------

	installedVcTools := NewInstalledVcTools()

	if installedVcTools.find(msvcVersion.vsPath, msvcVersion.vsVersion, msvcVersion.vsProduct, msvcVersion.targetVcToolsVersion) != nil {
		for _, product := range vsProducts {
			if product == msvcVersion.vsProduct {
				continue // Skip the product we already have done
			}

			if installedVcTools.find(msvcVersion.vsPath, msvcVersion.vsVersion, product, msvcVersion.targetVcToolsVersion) == nil {
				msvcVersion.vsProduct = product
				break
			}
		}
	}

	if len(installedVcTools.vcTools) == 0 {
		vcProduct := "Visual C++ tools"
		vcProductVersionDisclaimer := ""
		if msvcVersion.targetVcToolsVersion != "" {
			vcProduct = fmt.Sprintf("%s [Version %s]", vcProduct, msvcVersion.targetVcToolsVersion)
			vcProductVersionDisclaimer = "Note that a specific version of the Visual C++ tools has been requested. Remove the setting VcToolsVersion if this was undesirable\n"
		}

		searchSet := []string{}
		for _, tools := range installedVcTools.vcTools {
			searchSet = append(searchSet, tools.vsInstallDir)
		}

		vsDefaultPath := strings.ReplaceAll(vsDefaultPaths[vsDefaultVersion], "\\", "\\\\")
		corepkg.LogFatalf("%s not found\n\n  Cannot find %s in any of the following locations:\n    %s\n\n  Check that 'Desktop development with C++' is installed together with the product version in Visual Studio Installer\n\n  If you want to use a specific version of Visual Studio you can try setting Path, Version and Product like this:\n\n  Tools = {\n    { \"msvc-vs-latest\", Path = \"%s\", Version = \"%s\", Product = \"%s\" }\n  }\n\n  %s",
			vcProduct, vcProduct, strings.Join(searchSet, "\n    "), vsDefaultPath, vsDefaultVersion, vsProducts[0], vcProductVersionDisclaimer)
		return nil, fmt.Errorf("%s not found", vcProduct)
	}

	msdev = NewMsvcEnvironment(msvcVersion.vsVersion, msvcVersion.vsProduct)

	// to do: extension SDKs?
	vcTools := installedVcTools.vcTools[0]

	// VCToolsInstallDir
	vcToolsDir := filepath.Join(vcTools.vcInstallDir, "Tools", "MSVC", vcTools.vcToolsVersion)

	// VCToolsRedistDir
	// Ignored for now. Don't have a use case for this
	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L707

	msdev.Installed = corepkg.DirExists(vcTools.vsInstallDir)

	msdev.Path = append(msdev.Path, filepath.Join(vcTools.vsInstallDir, "Common7", "IDE", "VC", "VCPackages"))

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L761

	msdev.Path = append(msdev.Path, filepath.Join(vcToolsDir, "bin", "Host"+msvcVersion.hostArch.String(), msvcVersion.targetArch.String()))

	// to do: IFCPATH? C++ header/units and/or modules?
	// to do: LIBPATH? Fuse with #using C++/CLI
	// to do: https://learn.microsoft.com/en-us/windows/uwp/cpp-and-winrt-apis/intro-to-using-cpp-with-winrt#sdk-support-for-cwinrt

	msdev.IncludePaths = append(msdev.IncludePaths, filepath.Join(vcTools.vsInstallDir, "VC", "Auxiliary", "VS", "include"))
	msdev.IncludePaths = append(msdev.IncludePaths, filepath.Join(vcToolsDir, "include"))

	switch msvcVersion.winAppPlatform {
	case "Desktop":
		msdev.Libs = append(msdev.Libs, filepath.Join(vcToolsDir, "lib", msvcVersion.targetArch.String()))
		if msvcVersion.atlMfc == "true" {
			msdev.IncludePaths = append(msdev.IncludePaths, filepath.Join(vcToolsDir, "atlmfc", "include"))
			msdev.Libs = append(msdev.Libs, filepath.Join(vcToolsDir, "atlmfc", "lib", msvcVersion.targetArch.String()))
		}
	case "UWP":
		// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#825
		msdev.Libs = append(msdev.Libs, filepath.Join(vcToolsDir, "store", msvcVersion.targetArch.String()))
	case "OneCore":
		// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#830
		msdev.Libs = append(msdev.Libs, filepath.Join(vcToolsDir, "lib", "onecore", msvcVersion.targetArch.String()))
	}

	vcBin := ""
	if useClang {
		// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/VC/Tools/Llvm
		switch msvcVersion.targetArch {
		case WinArchx64:
			vcBin = filepath.Join(vcTools.vsInstallDir, "VC", "Tools", "llvm", "x64", "bin")
		case WinArchArm64:
			vcBin = filepath.Join(vcTools.vsInstallDir, "VC", "Tools", "llvm", "ARM64", "bin")
		case WinArchx86:
			vcBin = filepath.Join(vcTools.vsInstallDir, "VC", "Tools", "llvm", "bin")
		default:
			return nil, fmt.Errorf("msvc-clang: target architecture '%s' not supported", msvcVersion.targetArch.String())
		}
	} else {
		switch msvcVersion.targetArch {
		case WinArchx64:
			vcBin = filepath.Join(vcToolsDir, "bin", "Host"+msvcVersion.hostArch.String(), "x64")
		case WinArchArm64:
			vcBin = filepath.Join(vcToolsDir, "bin", "Host"+msvcVersion.hostArch.String(), "arm")
		case WinArchx86:
			vcBin = filepath.Join(vcToolsDir, "bin", "Host"+msvcVersion.hostArch.String(), "x86")
		default:
			return nil, fmt.Errorf("msvc-clang: target architecture '%s' not supported", msvcVersion.targetArch.String())
		}
	}

	// Force MSPDBSRV.EXE (fix for issue with cl.exe running in parallel and otherwise corrupting PDB files)
	// These options were added to Visual C++ in Visual Studio 2013. They do not exist in older versions.
	msdev.CcOptions = []string{"/FS"} // This is the C compiler option
	msdev.CxxOptions = []string{"/FS"}

	msdev.VcToolsVersion = vcTools.vcToolsVersion
	msdev.VsInstallDir = vcTools.vsInstallDir
	msdev.VcInstallDir = vcTools.vcInstallDir

	//
	// Tools
	//
	msdev.CompilerPath = vcBin
	msdev.CompilerBin = "cl.exe"
	msdev.ArchiverPath = vcBin
	msdev.ArchiverBin = "lib.exe"
	msdev.LinkerPath = vcBin
	msdev.LinkerBin = "link.exe"
	//msdev.RcPath = filepath.Join(winsdkDir, "bin", winsdkVersion, msvcVersion.hostArch.String())
	msdev.RcBin = "rc.exe"

	//
	// DevEnv
	//
	msdev.DevEnvDir = filepath.Join(vcTools.vsInstallDir, "Common7", "IDE")

	return msdev, nil
}
