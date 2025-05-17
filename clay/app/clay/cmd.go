package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jurgen-kluft/ccode/clay"
	"github.com/jurgen-kluft/ccode/clay/app/boards"
)

// Clay cli app
//    Commands:
//    - build --board (esp32, esp32s3)
//    - clean
//    - flash
//    - list libraries
//    - list boards (fuzzy searc)
//    - list flash sizes

var EspSdkPath = "/Users/obnosis5/sdk/arduino/esp32"
var BuildPath = "build"

func GetArgv(i int, _default string) string {
	if i < len(os.Args) {
		return os.Args[i]
	}
	return _default
}

func main() {
	// Parse command line arguments
	command := GetArgv(1, "build")
	switch command {
	case "build":
		boardName := GetArgv(2, "esp32")
		Build(boardName)
	case "clean":
		Clean()
	case "flash":
		Flash()
	case "list-libraries":
		ListLibraries()
	case "list-boards":
		fuzzyBoardName := GetArgv(2, "esp32")
		if max, err := strconv.Atoi(GetArgv(3, "0")); err == nil {
			ListBoards(fuzzyBoardName, max)
		}
	case "list-flash-sizes":
		cpuName := GetArgv(2, "esp32")
		boardName := GetArgv(3, "esp32")
		ListFlashSizes(cpuName, boardName)
	default:
		PrintUsage()
	}
}

func PrintUsage() {
	fmt.Println("Usage: clay [command] [options]")
	fmt.Println("Commands:")
	fmt.Println("  build <boardName>")
	fmt.Println("  clean")
	fmt.Println("  flash")
	fmt.Println("  list-libraries")
	fmt.Println("  list-boards [fuzzy] [max]")
	fmt.Println("  list-flash-sizes <cpuName> <boardName>")
	fmt.Println("Options:")
	fmt.Println("  fuzzy             Fuzzy search string for listing boards")
	fmt.Println("  max               Maximum number of boards to list")
	fmt.Println("  cpuName           CPU name for listing flash sizes")
	fmt.Println("  boardName         Board name (e.g. esp32, esp32s3) ")
	fmt.Println("  --help            Show this help message")
	fmt.Println("  --version         Show version information")

	fmt.Println("Examples:")
	fmt.Println("  clay build esp32")
	fmt.Println("  clay clean")
	fmt.Println("  clay flash")
	fmt.Println("  clay list-libraries")
	fmt.Println("  clay list-boards esp32 5")
}

func Build(Board string) error {
	// Note: We should be running this from the "target/esp" directory
	// Create the build directory
	os.MkdirAll("/"+BuildPath+"/", os.ModePerm)

	var buildEnv *clay.BuildEnvironment
	switch Board {
	case "esp32":
		buildEnv = clay.NewBuildEnvironmentEsp32(BuildPath)
	case "esp32s3":
		//buildEnv = clay.NewBuildEnvironmentEsp32S3(BuildPath)
	}

	if buildEnv == nil {
		return fmt.Errorf("Unsupported board: " + Board)
	}

	prj := CreateProject()
	return prj.Build(buildEnv)
}

func Clean() error {
	// Note: We should be running this from the "target/esp" directory
	// Remove all folders and files from "build/"
	if err := os.RemoveAll(BuildPath + "/"); err != nil {
		return fmt.Errorf("Failed to remove build directory: %v", err)
	}

	if err := os.MkdirAll(BuildPath+"/", os.ModePerm); err != nil {
		return fmt.Errorf("Failed to create build directory: %v", err)
	}

	return nil
}

func Flash() error {

	return nil
}

func ListLibraries() error {
	prj := CreateProject()
	fmt.Printf("Project: %s\n", prj.Name)
	for _, lib := range prj.Executable.Libraries {
		fmt.Printf("Library: %s\n", lib.Name)
		fmt.Printf("  Version: %s\n", lib.Version)
	}
	return nil
}

func ListBoards(fuzzy string, max int) error {
	if max == 0 {
		max = 10
	}
	boardsFilePath := filepath.Join(EspSdkPath, "boards.txt")
	return boards.PrintAllMatchingBoards(boardsFilePath, fuzzy, max)
}

func ListFlashSizes(cpuName string, boardName string) error {
	boardsFilePath := filepath.Join(EspSdkPath, "boards.txt")
	return boards.PrintAllFlashSizes(boardsFilePath, cpuName, boardName)
}

// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------
// Note: Everything below here is generated from the package definition

// !!Project!!

func CreateProject() *clay.Project {
	prjName := "axe"
	prjVersion := "0.1.0"
	prj := &clay.Project{
		Name:       prjName,
		Version:    prjVersion,
		BuildPath:  filepath.Join(BuildPath, prjName),
		Executable: clay.NewExecutable(prjName, prjVersion, BuildPath),
	}
	AddLibraries(prj.Executable)
	return prj
}

func AddLibraries(exe *clay.Executable) {
	{ // chash library
		name := "chash"
		library := clay.NewCppLibrary(name, "0.1.0", name, name+".a")

		// Include directories
		library.IncludeDirs.Add("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/chash/source/main/include", false)
		library.IncludeDirs.Add("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/cbase/source/main/include", false)
		library.IncludeDirs.Add("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/ccore/source/main/include", false)

		// Define macros
		library.Defines.Add("TARGET_DEBUG")
		library.Defines.Add("TARGET_ESP32")

		// Source files of chash
		library.AddSourceFile("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/chash/source/main/cpp/c_crc.cpp", "c_crc.cpp", false)
		library.AddSourceFile("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/chash/source/main/cpp/c_hash.cpp", "c_hash.cpp", false)
		// etc..

		exe.AddLibrary(library)
	}

	{ // cbase library
		name := "cbase"
		library := clay.NewCppLibrary(name, "0.1.0", name, name+".a")

		// Include directories
		library.IncludeDirs.Add("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/cbase/source/main/include", false)
		library.IncludeDirs.Add("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/ccore/source/main/include", false)

		// Define macros
		library.Defines.Add("TARGET_DEBUG")
		library.Defines.Add("TARGET_ESP32")

		// Source files of cbase
		library.AddSourceFile("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/cbase/source/main/cpp/c_allocator_system_esp32.cpp", "c_allocator_system_esp32.cpp", false)
		library.AddSourceFile("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/cbase/source/main/cpp/c_base.cpp", "c_base.cpp", false)
		// etc..

		exe.AddLibrary(library)
	}
}
