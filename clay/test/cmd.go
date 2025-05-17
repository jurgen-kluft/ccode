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
	project.SetBuildEnvironment(esp32)

	// Add a test library
	testLib := clay.NewCLibrary("Test_Lib", "1.0.0", "test_lib", "libtest_lib.a")
	testLib.IncludeDirs.Add("test_lib/include", false)
	testLib.AddSourceFilesFrom("test_lib/src", clay.OptionAddCppFiles)
	testLib.PrepareOutput(project.BuildPath)

	project.AddUserLibrary(testLib)

	project.Build(esp32)
}
