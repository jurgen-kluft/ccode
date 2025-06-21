package msvc

import (
	"fmt"
	"testing"

	"github.com/jurgen-kluft/ccode/foundation"
)

func TestSetupMsvcVersion(t *testing.T) {

	env := foundation.NewVars()

	msvcVersion := NewMsvcVersion()

	externalEnv, err := SetupMsvcVersion(env, msvcVersion, false)
	if err != nil {
		t.Fatal(err)
	}
	if externalEnv == nil {
		t.Fatal("Expected non-nil external environment")
	}
	if len(externalEnv.Keys) == 0 {
		t.Fatal("Expected external environment to have keys")
	}

	fmt.Printf("Visual Studio Tools version: %s\n", externalEnv.GetFirstOrEmpty("VSTOOLSVERSION"))
	fmt.Printf("Visual Studio Install directory: %s\n", externalEnv.GetFirstOrEmpty("VSINSTALLDIR"))
	fmt.Printf("Visual C/C++ Install directory: %s\n", externalEnv.GetFirstOrEmpty("VCINSTALLDIR"))
}
