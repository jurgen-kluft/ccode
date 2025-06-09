package tctools

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"slices"

	utils "github.com/jurgen-kluft/ccode/utils"
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

// Note that while Community, Professional and Enterprise products are installed
// in C:\Program Files while BuildTools are always installed in C:\Program Files (x86)
var vs_default_path = "C:\\Program Files (x86)\\Microsoft Visual Studio"

// Add new Visual Studio versions here and update vs_default_version
var vs_default_paths = map[string]string{
	"2017": vs_default_path,
	"2019": vs_default_path,
	"2022": "C:\\Program Files\\Microsoft Visual Studio",
}

var vs_default_version = "2022"

var vs_products = []string{
	"BuildTools", // default
	"Community",
	"Professional",
	"Enterprise",
}

var supported_arch_mappings = map[string]string{
	"x86":   "x86",
	"x64":   "x64",
	"arm":   "arm",
	"arm64": "arm64",
	"amd64": "x64", // alias
}

var supported_app_platforms = map[string]string{
	"desktop": "Desktop", // default
	"uwp":     "UWP",
	"onecore": "OneCore",
}

func getArch(arch string) string {
	arch2, ok := supported_arch_mappings[strings.ToLower(arch)]
	if !ok {
		return ""
	}
	return arch2
}

func getArchTuple(hostArch, targetArch string) (string, string, error) {
	hostArch2 := getArch(hostArch)
	targetArch2 := getArch(targetArch)
	if hostArch2 == "" || targetArch2 == "" {
		return "", "", fmt.Errorf("unknown host/target architecture '%s' expected one of %s", hostArch, utils.MapToString(supported_arch_mappings, "{key}={value}", ","))
	}
	return hostArch2, targetArch2, nil
}

func getAppPlatform(appPlatform string) (string, error) {
	appPlatform2, ok := supported_app_platforms[strings.ToLower(appPlatform)]
	if !ok {
		return "", fmt.Errorf("unknown app platform '%s' expected one of %s", appPlatform, utils.MapToString(supported_app_platforms, "{key}={value}", ","))
	}
	return appPlatform2, nil
}

func findWindowsSDK(targetWinsdkVersion, appPlatform string) (string, string, error) {
	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/core/winsdk.bat#L63
	//   HKLM\SOFTWARE\Wow6432Node
	//   HKCU\SOFTWARE\Wow6432Node (ignored)
	//   HKLM\SOFTWARE             (ignored)
	//   HKCU\SOFTWARE             (ignored)
	winsdkKey := "SOFTWARE\\Wow6432Node\\Microsoft\\Microsoft SDKs\\Windows\\v10.0"
	winsdkDir, ok := utils.QueryRegistryForStringValue("HKLM", winsdkKey, "InstallationFolder")
	if !ok {
		return "", "", fmt.Errorf("failed to query Windows SDK installation folder")
	}

	winsdkVersions := []string{}

	// Due to the SDK installer changes beginning with the 10.0.15063.0
	checkFile := "winsdkver.h"
	if strings.ToLower(appPlatform) == "uwp" {
		checkFile = "Windows.h"
	}

	dirs, err := utils.ListDirectory(filepath.Join(winsdkDir, "Include"))
	if err != nil {
		return "", "", fmt.Errorf("failed to list Windows SDK include directory: %v", err)
	}

	for _, winsdkVersion := range dirs {
		if strings.HasPrefix(winsdkVersion, "10.") {
			testPath := filepath.Join(winsdkDir, "Include", winsdkVersion, "um", checkFile)
			if utils.FileExists(testPath) {
				winsdkVersions = append(winsdkVersions, winsdkVersion)
			}
		}
	}

	if len(winsdkVersions) == 0 {
		return "", "", fmt.Errorf("no Windows SDK versions found in %s", winsdkDir)
	}

	if targetWinsdkVersion != "" {
		if slices.Contains(winsdkVersions, targetWinsdkVersion) {
			return winsdkDir, targetWinsdkVersion, nil
		}
		return "", "", fmt.Errorf("Windows SDK version '%s' not found. Available versions: %s", targetWinsdkVersion, strings.Join(winsdkVersions, ", "))
	}
	return winsdkDir, winsdkVersions[len(winsdkVersions)-1], nil // latest
}

func findVcTools(vsPath, vsVersion, vsProduct, targetVcToolsVersion string, searchSet []string) (string, string, string, error) {
	if vsPath == "" {
		if vsProduct == "BuildTools" {
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

	vsInstallDir := filepath.Join(vsPath, vsVersion, vsProduct)
	vcInstallDir := filepath.Join(vsInstallDir, "VC")
	var vcToolsVersion string

	if targetVcToolsVersion == "" {
		versionFile := filepath.Join(vcInstallDir, "Auxiliary", "Build", "Microsoft.VCToolsVersion.default.txt")
		data, err := utils.FileRead(versionFile)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to read VC tools version file: %v", err)
		}
		lines := strings.Split(string(data), "\n")
		if len(lines) > 0 {
			vcToolsVersion = strings.TrimSpace(lines[0])
		}
	} else {
		vcToolsVersion = targetVcToolsVersion
	}

	if vcToolsVersion != "" {
		testPath := filepath.Join(vcInstallDir, "Tools", "MSVC", vcToolsVersion, "include", "vcruntime.h")
		if utils.FileExists(testPath) {
			return vsInstallDir, vcInstallDir, vcToolsVersion, nil
		}
	}

	searchSet = append(searchSet, vsInstallDir)
	return "", "", "", fmt.Errorf("VC tools version '%s' not found in %s", vcToolsVersion, vsInstallDir)
}

/*
function apply(env, options, extra)
  if native.host_platform ~= "windows" then
    error("the msvc toolset only works on windows hosts")
  end

  tundra.unitgen.load_toolset('msvc', env)

  options = options or {}
  extra = extra or {}

  -- these control the environment
  local vs_path = options.Path or options.InstallationPath
  local vs_version = options.Version or extra.Version or vs_default_version
  local vs_product = options.Product
  local host_arch = options.HostArch or "x64"
  local target_arch = options.TargetArch or "x64"
  local app_platform = options.AppPlatform or options.PlatformType or "Desktop" -- Desktop, UWP or OneCore
  local target_winsdk_version = options.WindowsSdkVersion or options.SdkVersion -- Windows SDK version
  local target_vc_tools_version = options.VcToolsVersion -- Visual C++ tools version
  local atl_mfc = options.AtlMfc or false

  if vs_default_paths[vs_version] == nil then
    print("Warning: Visual Studio " .. vs_version .. " has not been tested and might not work out of the box")
  end

  host_arch, target_arch = get_arch_tuple(host_arch, target_arch)
  app_platform = get_app_platform(app_platform)

  local env_path = {}
  local env_include = {}
  local env_lib = {}
  local env_lib_path = {} -- WinRT

  -----------------
  -- Windows SDK --
  -----------------

  -- file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/core/winsdk.bat#L513

  local winsdk_dir, winsdk_version = find_winsdk(target_winsdk_version, app_platform)

  env_path[#env_path + 1] = native_path.join(winsdk_dir, "bin\\" .. winsdk_version .. "\\" .. host_arch)

  env_include[#env_include + 1] = native_path.join(winsdk_dir, "Include\\" .. winsdk_version .. "\\shared")
  env_include[#env_include + 1] = native_path.join(winsdk_dir, "Include\\" .. winsdk_version .. "\\um")
  env_include[#env_include + 1] = native_path.join(winsdk_dir, "Include\\" .. winsdk_version .. "\\winrt") -- WinRT (used by DirectX 12 headers)
  -- env_include[#env_include + 1] = native_path.join(winsdk_dir, "Include\\" .. winsdk_version .. "\\cppwinrt") -- WinRT

  env_lib[#env_lib + 1] = native_path.join(winsdk_dir, "Lib\\" .. winsdk_version .. "\\um\\" .. target_arch)

  -- We assume that the Universal CRT isn't loaded from a different directory
  local ucrt_sdk_dir = winsdk_dir
  local ucrt_version = winsdk_version

  env_include[#env_include + 1] = native_path.join(ucrt_sdk_dir, "Include\\" .. ucrt_version .. "\\ucrt")

  env_lib[#env_lib + 1] = native_path.join(ucrt_sdk_dir, "Lib\\" .. ucrt_version .. "\\ucrt\\" .. target_arch)

  -- Skip if the Universal CRT is loaded from the same path as the Windows SDK
  if ucrt_sdk_dir ~= winsdk_dir and ucrt_version ~= winsdk_version then
    env_lib[#env_lib + 1] = native_path.join(ucrt_sdk_dir, "Lib\\" .. ucrt_version .. "\\um\\" .. target_arch)
  end

  ----------------
  -- Visual C++ --
  ----------------

  local search_set = {}

  local vs_install_dir = nil
  local vc_install_dir = nil
  local vc_tools_version = nil

  -- If product is unspecified search for a suitable product
  if vs_product == nil then
    for _, product in pairs(vs_products) do
      vs_install_dir, vc_install_dir, vc_tools_version = find_vc_tools(vs_path, vs_version, product, target_vc_tools_version, search_set)
      if vc_tools_version ~= nil then
        vs_product = product
        break
      end
    end
  else
    vs_install_dir, vc_install_dir, vc_tools_version = find_vc_tools(vs_path, vs_version, vs_product, target_vc_tools_version, search_set)
  end

  if vc_tools_version == nil then
    local vc_product = "Visual C++ tools"
    local vc_product_version_disclaimer = ""
    if target_vc_tools_version ~= nil then
      vc_product = vc_product .. " [Version " .. target_vc_tools_version .. "]"
      vc_product_version_disclaimer = "Note that a specific version of the Visual C++ tools has been requested. Remove the setting VcToolsVersion if this was undesirable\n"
    end
    local vs_default_path = vs_default_paths[vs_default_version]:gsub("\\", "\\\\")
    error(vc_product .. " not found\n\n" .. "  Cannot find " .. vc_product ..
              " in any of the following locations:\n    " .. table.concat(search_set, "\n    ") ..
              "\n\n  Check that 'Desktop development with C++' is installed together with the product version in Visual Studio Installer\n\n" ..
              "  If you want to use a specific version of Visual Studio you can try setting Path, Version and Product like this:\n\n" ..
              "  Tools = {\n    { \"msvc-vs-latest\", Path = \"" .. vs_default_path .. "\", Version = \"" ..
              vs_default_version .. "\", Product = \"" .. vs_products[1] .. "\" }\n  }\n\n  " ..
              vc_product_version_disclaimer)
  end

  -- to do: extension SDKs?

  -- VCToolsInstallDir
  local vc_tools_dir = native_path.join(vc_install_dir, "Tools\\MSVC\\" .. vc_tools_version)

  -- VCToolsRedistDir
  -- Ignored for now. Don't have a use case for this

  -- file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L707

  env_path[#env_path + 1] = native_path.join(vs_install_dir, "Common7\\IDE\\VC\\VCPackages")

  -- file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L761

  env_path[#env_path + 1] = native_path.join(vc_tools_dir, "bin\\Host" .. host_arch .. "\\" .. target_arch)

  -- to do: IFCPATH? C++ header/units and/or modules?
  -- to do: LIBPATH? Fuse with #using C++/CLI
  -- to do: https://learn.microsoft.com/en-us/windows/uwp/cpp-and-winrt-apis/intro-to-using-cpp-with-winrt#sdk-support-for-cwinrt

  env_include[#env_include + 1] = native_path.join(vs_install_dir, "VC\\Auxiliary\\VS\\include")
  env_include[#env_include + 1] = native_path.join(vc_tools_dir, "include")

  if app_platform == "Desktop" then
    env_lib[#env_lib + 1] = native_path.join(vc_tools_dir, "lib\\" .. target_arch)
    if atl_mfc then
      env_include[#env_include + 1] = native_path.join(vc_tools_dir, "atlmfc\\include")
      env_lib[#env_lib + 1] = native_path.join(vc_tools_dir, "atlmfc\\lib\\" .. target_arch)
    end
  elseif app_platform == "UWP" then
    -- file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#825
    env_lib[#env_lib + 1] = native_path.join(vc_tools_dir, "store\\" .. target_arch)
  elseif app_platform == "OneCore" then
    -- file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#830
    env_lib[#env_lib + 1] = native_path.join(vc_tools_dir, "lib\\onecore\\" .. target_arch)
  end

  if extra.Clang then
    -- file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/VC/Tools/Llvm
    if target_arch == "x64" then
      env_path[#env_path + 1] = native_path.join(vs_install_dir, "VC\\Tools\\Llvm\\x64\\bin")
    elseif target_arch == "arm64" then
      env_path[#env_path + 1] = native_path.join(vs_install_dir, "VC\\Tools\\Llvm\\ARM64\\bin")
    elseif target_arch == "x86" then
      env_path[#env_path + 1] = native_path.join(vs_install_dir, "VC\\Tools\\Llvm\\bin")
    else
      error("msvc-clang: target architecutre '" .. target_arch .. "' not supported")
    end
  end

  env_path[#env_path + 1] = env:get_external_env_var('PATH')

  -- Force MSPDBSRV.EXE (fix for issue with cl.exe running in parallel and otherwise corrupting PDB files)
  -- These options where added to Visual C++ in Visual Studio 2013. They do not exist in older versions.
  env:set("CCOPTS", "/FS")
  env:set("CXXOPTS", "/FS")

  env:set_external_env_var("VSINSTALLDIR", vs_install_dir)
  env:set_external_env_var("VCINSTALLDIR", vc_install_dir)
  env:set_external_env_var("INCLUDE", table.concat(env_include, ";"))
  env:set_external_env_var("LIB", table.concat(env_lib, ";"))
  env:set_external_env_var("LIBPATH", table.concat(env_lib_path, ";"))
  env:set_external_env_var("PATH", table.concat(env_path, ";"))

  -- Since there's a bit of magic involved in finding these we log them once, at the end.
  -- This also makes it easy to lock the SDK and C++ tools version if you want to do that.
  if target_winsdk_version == nil then
    print("  WindowsSdkVersion : " .. winsdk_version) -- verbose?
  end
  if target_vc_tools_version == nil then
    print("  VcToolsVersion    : " .. vc_tools_version) -- verbose?
  end
end
*/

func ApplyMsvcVersion(env *utils.Vars, options *utils.Vars, extra *utils.Vars) {

	// Load the msvc toolset
	ApplyMsvc(env, options, extra)

	// These control the environment
	vsPath := options.GetOneDefault("Path", "")
	vsVersion := options.GetOneDefault("Version", vs_default_version)
	vsProduct := options.GetOneDefault("Product", "")
	hostArch := options.GetOneDefault("HostArch", "x64")
	targetArch := options.GetOneDefault("TargetArch", "x64")
	appPlatform := options.GetOneDefault("AppPlatform", "Desktop")        // Desktop, UWP or OneCore
	targetWinsdkVersion := options.GetOneDefault("WindowsSdkVersion", "") // Windows SDK version
	targetVcToolsVersion := options.GetOneDefault("VcToolsVersion", "")   // Visual C++ tools version
	atlMfc := options.GetOneDefault("AtlMfc", "false")

	if vsDefaultPath, ok := vs_default_paths[vsVersion]; !ok {
		utils.LogWarnf("Visual Studio %s has not been tested and might not work out of the box", vsVersion)
	} else if vsPath == "" {
		vsPath = vsDefaultPath
	}

	hostArch, targetArch, err := getArchTuple(hostArch, targetArch)
	if err != nil {
		utils.LogFatal(err)
	}

	appPlatform, err = getAppPlatform(appPlatform)
	if err != nil {
		utils.LogFatal(err)
	}

	envPath := []string{}
	envInclude := []string{}
	envLib := []string{}
	envLibPath := []string{} // WinRT

	// ------------------
	// Windows SDK
	// ------------------

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/core/winsdk.bat#L513

	winsdkDir, winsdkVersion, err := findWindowsSDK(targetWinsdkVersion, appPlatform)
	if err != nil {
		utils.LogFatal(err)
	}

	envPath = append(envPath, filepath.Join(winsdkDir, "bin", winsdkVersion, hostArch))

	envInclude = append(envInclude, filepath.Join(winsdkDir, "Include", winsdkVersion, "shared"))
	envInclude = append(envInclude, filepath.Join(winsdkDir, "Include", winsdkVersion, "um"))
	envInclude = append(envInclude, filepath.Join(winsdkDir, "Include", winsdkVersion, "winrt")) // WinRT (used by DirectX 12 headers)

	envLib = append(envLib, filepath.Join(winsdkDir, "Lib", winsdkVersion, "um", targetArch))

	// We assume that the Universal CRT isn't loaded from a different directory
	ucrtSdkDir := winsdkDir
	ucrtVersion := winsdkVersion

	envInclude = append(envInclude, filepath.Join(ucrtSdkDir, "Include", ucrtVersion, "ucrt"))

	envLib = append(envLib, filepath.Join(ucrtSdkDir, "Lib", ucrtVersion, "ucrt", targetArch))

	// Skip if the Universal CRT is loaded from the same path as the Windows SDK
	if ucrtSdkDir != winsdkDir || ucrtVersion != winsdkVersion {
		envLib = append(envLib, filepath.Join(ucrtSdkDir, "Lib", ucrtVersion, "um", targetArch))
	}

	// -------------------
	// Visual C++
	// -------------------

	searchSet := []string{}

	var vsInstallDir, vcInstallDir, vcToolsVersion string

	// If product is unspecified search for a suitable product
	if vsProduct == "" {
		for _, product := range vs_products {
			var err error
			vsInstallDir, vcInstallDir, vcToolsVersion, err = findVcTools(vsPath, vsVersion, product, targetVcToolsVersion, searchSet)
			if err == nil && vcToolsVersion != "" {
				vsProduct = product
				break
			}
		}
	} else {
		var err error
		vsInstallDir, vcInstallDir, vcToolsVersion, err = findVcTools(vsPath, vsVersion, vsProduct, targetVcToolsVersion, searchSet)
		if err != nil {
			utils.LogFatal(err)
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
		utils.LogFatalf("%s not found\n\n  Cannot find %s in any of the following locations:\n    %s\n\n  Check that 'Desktop development with C++' is installed together with the product version in Visual Studio Installer\n\n  If you want to use a specific version of Visual Studio you can try setting Path, Version and Product like this:\n\n  Tools = {\n    { \"msvc-vs-latest\", Path = \"%s\", Version = \"%s\", Product = \"%s\" }\n  }\n\n  %s",
			vcProduct, vcProduct, strings.Join(searchSet, "\n    "), vsDefaultPath, vs_default_version, vs_products[0], vcProductVersionDisclaimer)
	}

	// to do: extension SDKs?

	// VCToolsInstallDir
	vcToolsDir := filepath.Join(vcInstallDir, "Tools", "MSVC", vcToolsVersion)

	// VCToolsRedistDir
	// Ignored for now. Don't have a use case for this

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L707

	envPath = append(envPath, filepath.Join(vsInstallDir, "Common7", "IDE", "VC", "VCPackages"))

	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#L761

	envPath = append(envPath, filepath.Join(vcToolsDir, "bin", "Host"+hostArch, targetArch))

	// to do: IFCPATH? C++ header/units and/or modules?
	// to do: LIBPATH? Fuse with #using C++/CLI
	// to do: https://learn.microsoft.com/en-us/windows/uwp/cpp-and-winrt-apis/intro-to-using-cpp-with-winrt#sdk-support-for-cwinrt

	envInclude = append(envInclude, filepath.Join(vsInstallDir, "VC", "Auxiliary", "VS", "include"))
	envInclude = append(envInclude, filepath.Join(vcToolsDir, "include"))

	if appPlatform == "Desktop" {
		envLib = append(envLib, filepath.Join(vcToolsDir, "lib", targetArch))
		if atlMfc == "true" {
			envInclude = append(envInclude, filepath.Join(vcToolsDir, "atlmfc", "include"))
			envLib = append(envLib, filepath.Join(vcToolsDir, "atlmfc", "lib", targetArch))
		}
	} else if appPlatform == "UWP" {
		// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#825
		envLib = append(envLib, filepath.Join(vcToolsDir, "store", targetArch))
	} else if appPlatform == "OneCore" {
		// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/ext/vcvars.bat#830
		envLib = append(envLib, filepath.Join(vcToolsDir, "lib", "onecore", targetArch))
	}

	if extra.GetOneDefault("Clang", "false") == "true" {
		// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/VC/Tools/Llvm
		if targetArch == "x64" {
			envPath = append(envPath, filepath.Join(vsInstallDir, "VC", "Tools", "Llvm", "x64", "bin"))
		} else if targetArch == "arm64" {
			envPath = append(envPath, filepath.Join(vsInstallDir, "VC", "Tools", "Llvm", "ARM64", "bin"))
		} else if targetArch == "x86" {
			envPath = append(envPath, filepath.Join(vsInstallDir, "VC", "Tools", "Llvm", "bin"))
		} else {
			utils.LogFatalf("msvc-clang: target architecture '%s' not supported", targetArch)
		}
	}

	envPath = append(envPath, env.GetAll("PATH")...)

	// Force MSPDBSRV.EXE (fix for issue with cl.exe running in parallel and otherwise corrupting PDB files)
	// These options were added to Visual C++ in Visual Studio 2013. They do not exist in older versions.
	env.Set("CCOPTS", "/FS")
	env.Set("CXXOPTS", "/FS")

	env.Set("VSINSTALLDIR", vsInstallDir)
	env.Set("VCINSTALLDIR", vcInstallDir)
	env.Set("INCLUDE", strings.Join(envInclude, ";"))
	env.Set("LIB", strings.Join(envLib, ";"))
	env.Set("LIBPATH", strings.Join(envLibPath, ";"))
	env.Set("PATH", strings.Join(envPath, ";"))

	// Since there's a bit of magic involved in finding these we log them once, at the end.
	// This also makes it easy to lock the SDK and C++ tools version if you want to do that.
	if targetWinsdkVersion == "" {
		log.Printf("  WindowsSdkVersion : %s", winsdkVersion) // verbose?
	}
	if targetVcToolsVersion == "" {
		log.Printf("  VcToolsVersion    : %s", vcToolsVersion) // verbose?
	}
}
