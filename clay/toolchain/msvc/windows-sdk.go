package msvc

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

type WindowsSDK struct {
	Dir      string   // Directory of the Windows SDK
	Versions []string // List of available Windows SDK versions
}

func newWindowsSDK(dir string, versions []string) *WindowsSDK {
	sdk := &WindowsSDK{
		Dir:      dir,
		Versions: versions,
	}
	slices.Sort(sdk.Versions)
	return sdk
}

func (sdk *WindowsSDK) HasVersion(version string) bool {
	return slices.Contains(sdk.Versions, version)
}

func (sdk *WindowsSDK) GetLatestVersion() string {
	if len(sdk.Versions) == 0 {
		return ""
	}
	return sdk.Versions[len(sdk.Versions)-1]
}

// From Visual Studio 2017 onwards, this is the recommended way to find the Windows SDK.
func FindWindowsSDK(winAppPlatform WinAppPlatform) (*WindowsSDK, error) {
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

	winsdkVersions := []string{}

	// Due to the SDK installer changes beginning with the 10.0.15063.0
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
			}
		}
	}

	if len(winsdkVersions) == 0 {
		return nil, fmt.Errorf("no Windows SDK versions found in %s", winsdkDir)
	}

	return newWindowsSDK(winsdkDir, slices.Clip(winsdkVersions)), nil
}
