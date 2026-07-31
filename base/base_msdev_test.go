package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jurgen-kluft/gide/denv"
)

func TestGenerateForDevUsesMsdev(t *testing.T) {
	root := t.TempDir()
	pkg := denv.NewPackage("", "example")
	pkg.RootPath = root
	pkg.RepoPath = ""
	pkg.AddMainLib(denv.SetupCppHeaderProject(pkg, "example"))

	if err := generateForDev(pkg, "vs2022", denv.BuildTargetWindowsX64, false); err != nil {
		t.Fatalf("generateForDev() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "example", "target", "vs2022", "example.sln")); err != nil {
		t.Fatalf("generated solution: %v", err)
	}
}
