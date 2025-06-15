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

type vsVersion string

const (
	vsVersion2017 vsVersion = "2017"
	vsVersion2019 vsVersion = "2019"
	vsVersion2022 vsVersion = "2022"
)

func (v vsVersion) String() string {
	return string(v)
}

// Note that while Community, Professional and Enterprise products are installed
// in C:\Program Files while BuildTools are always installed in C:\Program Files (x86)
var vs_default_path = "C:\\Program Files (x86)\\Microsoft Visual Studio"

// Add new Visual Studio versions here and update vs_default_version
var vs_default_paths = map[vsVersion]string{
	vsVersion2017: vs_default_path,
	vsVersion2019: vs_default_path,
	vsVersion2022: "C:\\Program Files\\Microsoft Visual Studio",
}

var vs_default_version = vsVersion2022

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

type winSupportedArch string

const (
	winArchx86   winSupportedArch = "x86"
	winArchx64   winSupportedArch = "x64"
	winArchArm   winSupportedArch = "arm"
	winArchArm64 winSupportedArch = "arm64"
)

func (a winSupportedArch) String() string {
	return string(a)
}

var supported_arch_mappings = map[string]winSupportedArch{
	"x86":   winArchx86,
	"x64":   winArchx64,
	"arm":   winArchArm,
	"arm64": winArchArm64,
	"amd64": winArchx64,
}

type winAppPlatform string

const (
	Desktop winAppPlatform = "Desktop" // default
	UWP     winAppPlatform = "UWP"     // Universal Windows Platform
	OneCore winAppPlatform = "OneCore" // OneCore (Windows 10, Windows 11, Xbox, HoloLens)
)

var supported_app_platforms = map[string]winAppPlatform{
	"desktop": Desktop,
	"uwp":     UWP,
	"onecore": OneCore,
}

func getProduct(options *foundation.Vars) vsProduct {
	vsProduct := options.GetFirstOrEmpty("Product")
	switch strings.ToLower(vsProduct) {
	case "buildtools":
		return vsProductBuildTools
	case "community":
		return vsProductCommunity
	case "professional":
		return vsProductProfessional
	case "enterprise":
		return vsProductEnterprise
	default:
		return vsProductBuildTools // fallback, should never reach here
	}
}

func getVsVersion(options *foundation.Vars) vsVersion {
	vsVersion := options.GetFirstOrEmpty("Version")
	switch strings.ToLower(vsVersion) {
	case "2017":
		return vsVersion2017
	case "2019":
		return vsVersion2019
	case "2022":
		return vsVersion2022
	default:
		return vs_default_version // fallback, should never reach here
	}
}

func getArch(arch string) winSupportedArch {
	if arch2, ok := supported_arch_mappings[strings.ToLower(arch)]; ok {
		return arch2
	}
	return winArchx64
}

// getHostArch Gets the host architecture from the options, default to x64
func getHostArch(options *foundation.Vars) winSupportedArch {
	if hostArch, ok := options.GetFirst("HostArch"); ok {
		return getArch(hostArch)
	}
	if runtime.GOOS == "windows" {
		if runtime.GOARCH == "amd64" || runtime.GOARCH == "x86_64" {
			return winArchx64
		} else if runtime.GOARCH == "arm64" {
			return winArchArm64
		} else if runtime.GOARCH == "386" || runtime.GOARCH == "i386" {
			return winArchx86
		} else if runtime.GOARCH == "arm" {
			return winArchArm
		}
	}
	return winArchx64 // If not specified, default to x64
}

func getTargetArch(options *foundation.Vars) winSupportedArch {
	if targetArch, ok := options.GetFirst("TargetArch"); ok {
		return getArch(targetArch)
	}
	return winArchx64 // If not specified, default to x64
}

func getAppPlatform(options *foundation.Vars) winAppPlatform {
	winAppPlatform := options.GetFirstOrEmpty("AppPlatform")
	if platform, ok := supported_app_platforms[strings.ToLower(winAppPlatform)]; ok {
		return platform
	}
	return Desktop // default
}

func findVcTools(vsPath string, vsVersion vsVersion, vsProduct vsProduct, targetVcToolsVersion string, searchSet []string) (string, string, string, error) {
	if vsPath == "" {
		if vsProduct == vsProductBuildTools {
			vsPath = vs_default_path
		} else {
			vsPath = vs_default_paths[vsVersion]
			if vsPath == "" {
				vsPath = vs_default_paths[vs_default_version]
			}
		}
	}

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L729

	// we ignore Microsoft.VCToolsVersion.v143.default.txt and use Microsoft.VCToolsVersion.default.txt
	// unless a specific VC tools version was requested

	vsInstallDir := filepath.Join(vsPath, vsVersion.String(), vsProduct.String())
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

func SetupMsvcVersion(env *foundation.Vars, options *foundation.Vars, extra *foundation.Vars) (externalEnv *foundation.Vars, err error) {

	// These control the environment
	vsPath := options.GetFirstOrEmpty("Path")
	vsVersion := getVsVersion(options)                                  // default is 2022
	vsProduct := getProduct(options)                                    // default is BuildTools
	hostArch := getHostArch(options)                                    // default is x64
	targetArch := getTargetArch(options)                                // default is x64
	winAppPlatform := getAppPlatform(options)                           // default is Desktop
	targetWinsdkVersion := options.GetFirstOrEmpty("WindowsSdkVersion") // Windows SDK version
	targetVcToolsVersion := options.GetFirstOrEmpty("VcToolsVersion")   // Visual C++ tools version
	atlMfc := options.GetFirstOrEmpty("AtlMfc")

	if vsDefaultPath, ok := vs_default_paths[vsVersion]; !ok {
		foundation.LogWarnf("Visual Studio %s has not been tested and might not work out of the box", vsVersion)
	} else if vsPath == "" {
		vsPath = vsDefaultPath
	}

	envPath, _ := env.Get("PATH")
	envInclude, _ := env.Get("INCLUDE")
	envLib, _ := env.Get("LIB")
	envLibPath, _ := env.Get("LIBPATH")

	// ------------------
	// Windows SDK
	// ------------------

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/core/winsdk.bat#L513

	winsdkDir, winsdkVersion, err := findWindowsSDK(targetWinsdkVersion, winAppPlatform)
	if err != nil {
		return nil, err
	}

	envPath = append(envPath, filepath.Join(winsdkDir, "bin", winsdkVersion, hostArch.String()))

	envInclude = append(envInclude, filepath.Join(winsdkDir, "Include", winsdkVersion, "shared"))
	envInclude = append(envInclude, filepath.Join(winsdkDir, "Include", winsdkVersion, "um"))
	envInclude = append(envInclude, filepath.Join(winsdkDir, "Include", winsdkVersion, "winrt")) // WinRT (used by DirectX 12 headers)

	envLib = append(envLib, filepath.Join(winsdkDir, "Lib", winsdkVersion, "um", targetArch.String()))

	// We assume that the Universal CRT isn't loaded from a different directory
	ucrtSdkDir := winsdkDir
	ucrtVersion := winsdkVersion

	envInclude = append(envInclude, filepath.Join(ucrtSdkDir, "Include", ucrtVersion, "ucrt"))

	envLib = append(envLib, filepath.Join(ucrtSdkDir, "Lib", ucrtVersion, "ucrt", targetArch.String()))

	// Skip if the Universal CRT is loaded from the same path as the Windows SDK
	if ucrtSdkDir != winsdkDir || ucrtVersion != winsdkVersion {
		envLib = append(envLib, filepath.Join(ucrtSdkDir, "Lib", ucrtVersion, "um", targetArch.String()))
	}

	// -------------------
	// Visual C++
	// -------------------

	searchSet := []string{}

	vsInstallDir, vcInstallDir, vcToolsVersion, err := findVcTools(vsPath, vsVersion, vsProduct, targetVcToolsVersion, searchSet)
	if err != nil {
		for _, product := range vsProducts {
			if product == vsProduct {
				continue // Skip the product we already have done
			}
			vsInstallDir, vcInstallDir, vcToolsVersion, err = findVcTools(vsPath, vsVersion, product, targetVcToolsVersion, searchSet)
			if err == nil {
				vsProduct = product
				break
			}
		}
	}

	if vcToolsVersion == "" {
		vcProduct := "Visual C++ tools"
		vcProductVersionDisclaimer := ""
		if targetVcToolsVersion != "" {
			vcProduct = fmt.Sprintf("%s [Version %s]", vcProduct, targetVcToolsVersion)
			vcProductVersionDisclaimer = "Note that a specific version of the Visual C++ tools has been requested. Remove the setting VcToolsVersion if this was undesirable\n"
		}
		vsDefaultPath := strings.ReplaceAll(vs_default_paths[vs_default_version], "\\", "\\\\")
		foundation.LogFatalf("%s not found\n\n  Cannot find %s in any of the following locations:\n    %s\n\n  Check that 'Desktop development with C++' is installed together with the product version in Visual Studio Installer\n\n  If you want to use a specific version of Visual Studio you can try setting Path, Version and Product like this:\n\n  Tools = {\n    { \"msvc-vs-latest\", Path = \"%s\", Version = \"%s\", Product = \"%s\" }\n  }\n\n  %s",
			vcProduct, vcProduct, strings.Join(searchSet, "\n    "), vsDefaultPath, vs_default_version, vsProducts[0], vcProductVersionDisclaimer)
	}

	// to do: extension SDKs?

	// VCToolsInstallDir
	vcToolsDir := filepath.Join(vcInstallDir, "Tools", "MSVC", vcToolsVersion)

	// VCToolsRedistDir
	// Ignored for now. Don't have a use case for this

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L707

	envPath = append(envPath, filepath.Join(vsInstallDir, "Common7", "IDE", "VC", "VCPackages"))

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L761

	envPath = append(envPath, filepath.Join(vcToolsDir, "bin", "Host"+hostArch.String(), targetArch.String()))

	// to do: IFCPATH? C++ header/units and/or modules?
	// to do: LIBPATH? Fuse with #using C++/CLI
	// to do: https://learn.microsoft.com/en-us/windows/uwp/cpp-and-winrt-apis/intro-to-using-cpp-with-winrt#sdk-support-for-cwinrt

	envInclude = append(envInclude, filepath.Join(vsInstallDir, "VC", "Auxiliary", "VS", "include"))
	envInclude = append(envInclude, filepath.Join(vcToolsDir, "include"))

	if winAppPlatform == "Desktop" {
		envLib = append(envLib, filepath.Join(vcToolsDir, "lib", targetArch.String()))
		if atlMfc == "true" {
			envInclude = append(envInclude, filepath.Join(vcToolsDir, "atlmfc", "include"))
			envLib = append(envLib, filepath.Join(vcToolsDir, "atlmfc", "lib", targetArch.String()))
		}
	} else if winAppPlatform == "UWP" {
		// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#825
		envLib = append(envLib, filepath.Join(vcToolsDir, "store", targetArch.String()))
	} else if winAppPlatform == "OneCore" {
		// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#830
		envLib = append(envLib, filepath.Join(vcToolsDir, "lib", "onecore", targetArch.String()))
	}

	if extra.GetFirstOrEmpty("Clang") == "true" {
		// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/VC/Tools/Llvm
		if targetArch == winArchx64 {
			envPath = append(envPath, filepath.Join(vsInstallDir, "VC", "Tools", "Llvm", "x64", "bin"))
		} else if targetArch == winArchArm64 {
			envPath = append(envPath, filepath.Join(vsInstallDir, "VC", "Tools", "Llvm", "ARM64", "bin"))
		} else if targetArch == winArchx86 {
			envPath = append(envPath, filepath.Join(vsInstallDir, "VC", "Tools", "Llvm", "bin"))
		} else {
			return nil, fmt.Errorf("msvc-clang: target architecture '%s' not supported", targetArch.String())
		}
	}

	if paths, ok := env.Get("PATH"); ok {
		envPath = append(envPath, paths...)
	}

	// Force MSPDBSRV.EXE (fix for issue with cl.exe running in parallel and otherwise corrupting PDB files)
	// These options were added to Visual C++ in Visual Studio 2013. They do not exist in older versions.
	env.Set("CCOPTS", "/FS")
	env.Set("CXXOPTS", "/FS")

	externalEnv = foundation.NewVars()
	externalEnv.Set("VSTOOLSVERSION", vcToolsVersion)
	externalEnv.Set("VSINSTALLDIR", vsInstallDir)
	externalEnv.Set("VCINSTALLDIR", vcInstallDir)
	externalEnv.Set("INCLUDE", envInclude...)
	externalEnv.Set("LIB", envLib...)
	externalEnv.Set("LIBPATH", envLibPath...)
	externalEnv.Set("PATH", envPath...)

	// Since there's a bit of magic involved in finding these we log them once, at the end.
	// This also makes it easy to lock the SDK and C++ tools version if you want to do that.
	if targetWinsdkVersion == "" {
		foundation.LogInfof("  WindowsSdkVersion : %s", winsdkVersion) // verbose?
	}
	if targetVcToolsVersion == "" {
		foundation.LogInfof("  VcToolsVersion    : %s", vcToolsVersion) // verbose?
	}

	return externalEnv, nil
}
