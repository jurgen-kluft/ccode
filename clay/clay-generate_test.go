package clay

import (
	"os"
	"path/filepath"
	"testing"

	corepkg "github.com/jurgen-kluft/go-core"
	"github.com/jurgen-kluft/go-ide/denv"
)

func TestGenerateVs2022(t *testing.T) {
	root := t.TempDir()
	pkg := denv.NewPackage("", "example")
	pkg.RootPath = root
	pkg.RepoPath = ""
	if err := os.MkdirAll(filepath.Join(root, "example", "source", "main", "cpp"), 0o755); err != nil {
		t.Fatal(err)
	}
	project := denv.SetupCppAppProjectForDesktop(pkg, "example", "main")
	pkg.AddMainApp(project)
	pkg.SetGetVarsFunc(func(_ denv.BuildTarget, _ denv.BuildConfig, _ string, vars *corepkg.Vars) {
		vars.Set("build.cpp_standard", "/std:c++20")
		vars.Set("compiler.c.link.extra_flags", "/SUBSYSTEM:CONSOLE", "/MACHINE:X64")
	})

	outputDir := filepath.Join(root, "generated")
	app := NewApp(pkg)
	if err := app.Generate([]string{"--dev=vs2022", "--build=debug", "--output=" + outputDir}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "example.sln")); err != nil {
		t.Fatalf("generated solution: %v", err)
	}
}

func TestParseGenerateOptions(t *testing.T) {
	options, err := ParseGenerateOptions([]string{"--dev=vs2022", "--arch=arm64", "-p", "app_example"})
	if err != nil {
		t.Fatalf("ParseGenerateOptions() error = %v", err)
	}
	if options.Dev != "vs2022" || options.Arch != "arm64" || options.StartupProject != "app_example" {
		t.Fatalf("ParseGenerateOptions() = %#v", options)
	}
}

func TestParseGenerateOptionsRejectsUnsupportedGenerator(t *testing.T) {
	if _, err := ParseGenerateOptions([]string{"--dev=xcode"}); err == nil {
		t.Fatal("ParseGenerateOptions() accepted unsupported generator")
	}
	if _, err := ParseGenerateOptions([]string{"--dev=vs2022", "--arch=esp32"}); err == nil {
		t.Fatal("ParseGenerateOptions() accepted unsupported Visual Studio architecture")
	}
}
