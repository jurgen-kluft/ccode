package msvc

import (
	"fmt"
	"testing"
)

func TestSetupMsvcVersion(t *testing.T) {
	msvcVersion := NewMsvcVersion()
	msdevSetup, err := setupMsvcVersion(msvcVersion, false)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Visual Studio Tools version: %s\n", msdevSetup.VcToolsVersion)
	fmt.Printf("Visual Studio Install directory: %s\n", msdevSetup.VsInstallDir)
	fmt.Printf("Visual C/C++ Install directory: %s\n", msdevSetup.VcInstallDir)
}
