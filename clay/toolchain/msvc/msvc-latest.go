package msvc

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jurgen-kluft/ccode/foundation"
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
	VsVersion2017 VsVersion = "2017"
	VsVersion2019 VsVersion = "2019"
	VsVersion2022 VsVersion = "2022"
)

func (v VsVersion) String() string {
	return string(v)
}

// Note that while Community, Professional and Enterprise products are installed
// in C:\Program Files while BuildTools are always installed in C:\Program Files (x86)
var vs_default_path = "C:\\Program Files (x86)\\Microsoft Visual Studio"

// Add new Visual Studio versions here and update vsDefaultVersion
var vs_default_paths = map[VsVersion]string{
	VsVersion2017: vs_default_path,
	VsVersion2019: vs_default_path,
	VsVersion2022: "C:\\Program Files\\Microsoft Visual Studio",
}

var vsDefaultVersion = VsVersion2022

type vsProduct string

const (
	vsProductBuildTools   vsProduct = "BuildTools" // default
	vsProductCommunity    vsProduct = "Community"
	vsProductProfessional vsProduct = "Professional"
	vsProductEnterprise   vsProduct = "Enterprise"
)

func (p vsProduct) String() string {
	return string(p)
}

var vsProducts = []vsProduct{
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

var supported_arch_mappings = map[string]WinSupportedArch{
	"x86":   WinArchx86,
	"x64":   WinArchx64,
	"arm":   WinArchArm,
	"arm64": WinArchArm64,
	"amd64": WinArchx64,
}

type winAppPlatform string

const (
	winAppDesktop winAppPlatform = "Desktop" // default
	winAppUWP     winAppPlatform = "UWP"     // Universal Windows Platform
	winAppOneCore winAppPlatform = "OneCore" // OneCore (Windows 10, Windows 11, Xbox, HoloLens)
)

var supported_app_platforms = map[string]winAppPlatform{
	"desktop": winAppDesktop,
	"uwp":     winAppUWP,
	"onecore": winAppOneCore,
}

func getArch(arch WinSupportedArch) WinSupportedArch {
	if arch2, ok := supported_arch_mappings[strings.ToLower(arch.String())]; ok {
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

func findVcTools(vsPath string, VsVersion VsVersion, vsProduct vsProduct, targetVcToolsVersion string, searchSet []string) (string, string, string, error) {
	if vsPath == "" {
		if vsProduct == vsProductBuildTools {
			vsPath = vs_default_path
		} else {
			vsPath = vs_default_paths[VsVersion]
			if vsPath == "" {
				vsPath = vs_default_paths[vsDefaultVersion]
			}
		}
	}

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L729

	// we ignore Microsoft.VCToolsVersion.v143.default.txt and use Microsoft.VCToolsVersion.default.txt
	// unless a specific VC tools version was requested

	vsInstallDir := filepath.Join(vsPath, VsVersion.String(), vsProduct.String())
	vcInstallDir := filepath.Join(vsInstallDir, "VC")

	var vcToolsVersion string
	if targetVcToolsVersion == "" {
		versionFile := filepath.Join(vcInstallDir, "Auxiliary", "Build", "Microsoft.VCToolsVersion.default.txt")
		data, err := foundation.FileOpenReadClose(versionFile)
		if err == nil {
			lines := strings.Split(string(data), "\n")
			if len(lines) > 0 {
				vcToolsVersion = strings.TrimSpace(lines[0])
			}
		}
	} else {
		vcToolsVersion = targetVcToolsVersion
	}

	if vcToolsVersion != "" {
		testPath := filepath.Join(vcInstallDir, "Tools", "MSVC", vcToolsVersion, "include", "vcruntime.h")
		if foundation.FileExists(testPath) {
			return vsInstallDir, vcInstallDir, vcToolsVersion, nil
		}
	}

	searchSet = append(searchSet, vsInstallDir)
	return "", "", "", fmt.Errorf("VC tools version '%s' not found in %s", vcToolsVersion, vsInstallDir)
}

type MsvcVersion struct {
	vsPath               string
	VsVersion            VsVersion        // default is 2022
	vsProduct            vsProduct        // default is BuildTools
	hostArch             WinSupportedArch // default is x64
	targetArch           WinSupportedArch // default is x64
	winAppPlatform       winAppPlatform   // default is Desktop
	targetWinsdkVersion  string           // Windows SDK version
	targetVcToolsVersion string           // Visual C++ tools version
	atlMfc               string
}

func NewMsvcVersion() *MsvcVersion {
	return &MsvcVersion{
		vsPath:               "",
		VsVersion:            vsDefaultVersion,
		vsProduct:            vsProductBuildTools,
		hostArch:             getHostArch(""),
		targetArch:           getTargetArch(""),
		winAppPlatform:       winAppDesktop,
		targetWinsdkVersion:  "",
		targetVcToolsVersion: "",
		atlMfc:               "false", // default is false
	}
}

func setupMsvcVersion(msdev *MsDevSetup, msvcVersion *MsvcVersion, useClang bool) error {

	// These control the environment

	if vsDefaultPath, ok := vs_default_paths[msvcVersion.VsVersion]; !ok {
		foundation.LogWarnf("Visual Studio %s has not been tested and might not work out of the box", msvcVersion.VsVersion.String())
	} else if msvcVersion.vsPath == "" {
		msvcVersion.vsPath = vsDefaultPath
	}

	envPath := msdev.Path
	envInclude := msdev.IncludePaths
	envLib := msdev.Libs
	envLibPaths := msdev.LibPaths

	// ------------------
	// Windows SDK
	// ------------------

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/core/winsdk.bat#L513

	winsdkDir, winsdkVersion, err := findWindowsSDK(msvcVersion.targetWinsdkVersion, msvcVersion.winAppPlatform)
	if err != nil {
		return err
	}

	envPath = append(envPath, filepath.Join(winsdkDir, "bin", winsdkVersion, msvcVersion.hostArch.String()))

	envInclude = append(envInclude, filepath.Join(winsdkDir, "Include", winsdkVersion, "shared"))
	envInclude = append(envInclude, filepath.Join(winsdkDir, "Include", winsdkVersion, "um"))
	envInclude = append(envInclude, filepath.Join(winsdkDir, "Include", winsdkVersion, "winrt")) // WinRT (used by DirectX 12 headers)

	envLib = append(envLib, filepath.Join(winsdkDir, "Lib", winsdkVersion, "um", msvcVersion.targetArch.String()))

	// We assume that the Universal CRT isn't loaded from a different directory
	ucrtSdkDir := winsdkDir
	ucrtVersion := winsdkVersion

	envInclude = append(envInclude, filepath.Join(ucrtSdkDir, "Include", ucrtVersion, "ucrt"))

	envLib = append(envLib, filepath.Join(ucrtSdkDir, "Lib", ucrtVersion, "ucrt", msvcVersion.targetArch.String()))

	// Skip if the Universal CRT is loaded from the same path as the Windows SDK
	if ucrtSdkDir != winsdkDir || ucrtVersion != winsdkVersion {
		envLib = append(envLib, filepath.Join(ucrtSdkDir, "Lib", ucrtVersion, "um", msvcVersion.targetArch.String()))
	}

	// -------------------
	// Visual C++
	// -------------------

	searchSet := []string{}

	vsInstallDir, vcInstallDir, vcToolsVersion, err := findVcTools(msvcVersion.vsPath, msvcVersion.VsVersion, msvcVersion.vsProduct, msvcVersion.targetVcToolsVersion, searchSet)
	if err != nil {
		for _, product := range vsProducts {
			if product == msvcVersion.vsProduct {
				continue // Skip the product we already have done
			}
			vsInstallDir, vcInstallDir, vcToolsVersion, err = findVcTools(msvcVersion.vsPath, msvcVersion.VsVersion, product, msvcVersion.targetVcToolsVersion, searchSet)
			if err == nil {
				msvcVersion.vsProduct = product
				break
			}
		}
	}

	if vcToolsVersion == "" {
		vcProduct := "Visual C++ tools"
		vcProductVersionDisclaimer := ""
		if msvcVersion.targetVcToolsVersion != "" {
			vcProduct = fmt.Sprintf("%s [Version %s]", vcProduct, msvcVersion.targetVcToolsVersion)
			vcProductVersionDisclaimer = "Note that a specific version of the Visual C++ tools has been requested. Remove the setting VcToolsVersion if this was undesirable\n"
		}
		vsDefaultPath := strings.ReplaceAll(vs_default_paths[vsDefaultVersion], "\\", "\\\\")
		foundation.LogFatalf("%s not found\n\n  Cannot find %s in any of the following locations:\n    %s\n\n  Check that 'Desktop development with C++' is installed together with the product version in Visual Studio Installer\n\n  If you want to use a specific version of Visual Studio you can try setting Path, Version and Product like this:\n\n  Tools = {\n    { \"msvc-vs-latest\", Path = \"%s\", Version = \"%s\", Product = \"%s\" }\n  }\n\n  %s",
			vcProduct, vcProduct, strings.Join(searchSet, "\n    "), vsDefaultPath, vsDefaultVersion, vsProducts[0], vcProductVersionDisclaimer)
	}

	// to do: extension SDKs?

	// VCToolsInstallDir
	vcToolsDir := filepath.Join(vcInstallDir, "Tools", "MSVC", vcToolsVersion)

	// VCToolsRedistDir
	// Ignored for now. Don't have a use case for this

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L707

	envPath = append(envPath, filepath.Join(vsInstallDir, "Common7", "IDE", "VC", "VCPackages"))

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L761

	envPath = append(envPath, filepath.Join(vcToolsDir, "bin", "Host"+msvcVersion.hostArch.String(), msvcVersion.targetArch.String()))

	// to do: IFCPATH? C++ header/units and/or modules?
	// to do: LIBPATH? Fuse with #using C++/CLI
	// to do: https://learn.microsoft.com/en-us/windows/uwp/cpp-and-winrt-apis/intro-to-using-cpp-with-winrt#sdk-support-for-cwinrt

	envInclude = append(envInclude, filepath.Join(vsInstallDir, "VC", "Auxiliary", "VS", "include"))
	envInclude = append(envInclude, filepath.Join(vcToolsDir, "include"))

	if msvcVersion.winAppPlatform == "Desktop" {
		envLib = append(envLib, filepath.Join(vcToolsDir, "lib", msvcVersion.targetArch.String()))
		if msvcVersion.atlMfc == "true" {
			envInclude = append(envInclude, filepath.Join(vcToolsDir, "atlmfc", "include"))
			envLib = append(envLib, filepath.Join(vcToolsDir, "atlmfc", "lib", msvcVersion.targetArch.String()))
		}
	} else if msvcVersion.winAppPlatform == "UWP" {
		// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#825
		envLib = append(envLib, filepath.Join(vcToolsDir, "store", msvcVersion.targetArch.String()))
	} else if msvcVersion.winAppPlatform == "OneCore" {
		// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#830
		envLib = append(envLib, filepath.Join(vcToolsDir, "lib", "onecore", msvcVersion.targetArch.String()))
	}

	if useClang {
		// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/VC/Tools/Llvm
		if msvcVersion.targetArch == WinArchx64 {
			envPath = append(envPath, filepath.Join(vsInstallDir, "VC", "Tools", "Llvm", "x64", "bin"))
		} else if msvcVersion.targetArch == WinArchArm64 {
			envPath = append(envPath, filepath.Join(vsInstallDir, "VC", "Tools", "Llvm", "ARM64", "bin"))
		} else if msvcVersion.targetArch == WinArchx86 {
			envPath = append(envPath, filepath.Join(vsInstallDir, "VC", "Tools", "Llvm", "bin"))
		} else {
			return fmt.Errorf("msvc-clang: target architecture '%s' not supported", msvcVersion.targetArch.String())
		}
	}

	// Force MSPDBSRV.EXE (fix for issue with cl.exe running in parallel and otherwise corrupting PDB files)
	// These options were added to Visual C++ in Visual Studio 2013. They do not exist in older versions.
	msdev.CcOptions = []string{"/FS"} // This is the C compiler option
	msdev.CxxOptions = []string{"/FS"}

	msdev.VcToolsVersion = vcToolsVersion
	msdev.VsInstallDir = vsInstallDir
	msdev.VcInstallDir = vcInstallDir
	msdev.IncludePaths = envInclude
	msdev.Libs = envLib
	msdev.LibPaths = envLibPaths
	msdev.Path = envPath

	// Since there's a bit of magic involved in finding these we log them once, at the end.
	// This also makes it easy to lock the SDK and C++ tools version if you want to do that.
	if msvcVersion.targetWinsdkVersion == "" {
		foundation.LogInfof("  WindowsSdkVersion : %s", winsdkVersion) // verbose?
	}
	if msvcVersion.targetVcToolsVersion == "" {
		foundation.LogInfof("  VcToolsVersion    : %s", vcToolsVersion) // verbose?
	}

	return nil
}
