package msvc

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jurgen-kluft/ccode/foundation"
)

// From Visual Studio 2017 onwards, this is the recommended way to find the Windows SDK.
func findWindowsSDK(targetWinsdkVersion string, winAppPlatform winAppPlatform) (string, string, error) {
	// file:///C:/Program%20Files%20(x86)/Microsoft%20Visual%20Studio/2022/BuildTools/Common7/Tools/vsdevcmd/core/winsdk.bat#L63
	//   HKLM\SOFTWARE\Wow6432Node
	//   HKCU\SOFTWARE\Wow6432Node (ignored)
	//   HKLM\SOFTWARE             (ignored)
	//   HKCU\SOFTWARE             (ignored)
	winsdkKey := `SOFTWARE\Wow6432Node\Microsoft\Microsoft SDKs\Windows\\v10.0`
	winsdkDir, err := foundation.QueryRegistryForStringValue(foundation.RegistryKeyLocalMachine, winsdkKey, "InstallationFolder")
	if err != nil {
		return "", "", err
	}

	winsdkVersions := []string{}

	// Due to the SDK installer changes beginning with the 10.0.15063.0
	checkFile := "winsdkver.h"
	if winAppPlatform == winAppUWP {
		checkFile = "Windows.h"
	}

	dirs, err := foundation.DirList(filepath.Join(winsdkDir, "Include"))
	if err != nil {
		return "", "", fmt.Errorf("failed to list Windows SDK include directory: %v", err)
	}

	for _, winsdkVersion := range dirs {
		if strings.HasPrefix(winsdkVersion, "10.") {
			testPath := filepath.Join(winsdkDir, "Include", winsdkVersion, "um", checkFile)
			if foundation.FileExists(testPath) {
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
