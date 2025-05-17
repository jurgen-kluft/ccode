package main

import "github.com/jurgen-kluft/ccode/clay"

// Test clay package

func main() {
	esp32 := clay.NewBuildEnvironmentEsp32("build")
	if esp32 == nil {
		panic("Failed to create ESP32 target")
	}

	// Create a new test project
	project := clay.NewProject("TestProject", "1.0.0", "build")

	// Initialize the project core
	project.InitCore(esp32)

	// Add a test library
	testLib := clay.NewCLibrary("TestLib", "1.0.0", "testlib", "libtestlib.a")
	testLib.IncludeDirs.Add("testlib/include", false)
	testLib.AddSourceFilesFrom("testlib/src", clay.OptionAddCppFiles)
	project.AddUserLibrary(testLib)

	project.Build(esp32)
}
