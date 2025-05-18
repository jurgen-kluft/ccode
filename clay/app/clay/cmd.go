package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	ccode_utils "github.com/jurgen-kluft/ccode/ccode-utils"
	"github.com/jurgen-kluft/ccode/clay"
	"github.com/jurgen-kluft/ccode/clay/app/boards"
)

// Clay cli app
//    Commands:
//    - build board (esp32, esp32s3)
//    - build-info board (esp32, esp32s3)
//    - clean board (esp32, esp32s3)
//    - flash board (esp32, esp32s3)
//    - list libraries
//    - list boards (fuzzy search)
//    - list flash sizes

var BuildInfoFilenameWithoutExt = "buildinfo"
var EspSdkPath = "/Users/obnosis5/sdk/arduino/esp32"

func GetArgv(i int, _default string) string {
	if i < len(os.Args) {
		return os.Args[i]
	}
	return _default
}

func GetBuildPath(board string) string {
	// /build/esp32
	// /build/esp32s3
	buildPath := filepath.Join("build", board)
	return buildPath
}

func main() {
	// Parse command line arguments
	command := GetArgv(1, "build")
	switch command {
	case "build":
		boardName := GetArgv(2, "esp32")
		Build(boardName)
	case "build-info":
		boardName := GetArgv(2, "esp32")
		GenerateBuildInfo(boardName)
	case "clean":
		boardName := GetArgv(2, "esp32")
		Clean(boardName)
	case "flash":
		boardName := GetArgv(2, "esp32")
		Flash(boardName)
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
	fmt.Println("  build-info <boardName> (generates buildinfo.h and buildinfo.cpp)")
	fmt.Println("  build <boardName>")
	fmt.Println("  clean <boardName>")
	fmt.Println("  flash <boardName>")
	fmt.Println("  list-libraries")
	fmt.Println("  list-boards [boardName] [matches]")
	fmt.Println("  list-flash-sizes <cpuName> <boardName>")
	fmt.Println("Options:")
	fmt.Println("  matches           Maximum number of boards to list")
	fmt.Println("  cpuName           CPU name for listing flash sizes")
	fmt.Println("  boardName         Board name (e.g. esp32, esp32s3) ")
	fmt.Println("  --help            Show this help message")
	fmt.Println("  --version         Show version information")

	fmt.Println("Examples:")
	fmt.Println("  clay build-info esp32 (generates buildinfo.h and buildinfo.cpp)")
	fmt.Println("  clay build esp32")
	fmt.Println("  clay clean esp32")
	fmt.Println("  clay flash esp32")
	fmt.Println("  clay list-libraries")
	fmt.Println("  clay list-boards esp32 5")
}

func Build(board string) error {
	// Note: We should be running this from the "target/esp" directory
	// Create the build directory
	buildPath := GetBuildPath(board)
	os.MkdirAll(buildPath+"/", os.ModePerm)

	var buildEnv *clay.BuildEnvironment
	switch board {
	case "esp32":
		buildEnv = clay.NewBuildEnvironmentEsp32(buildPath)
	case "esp32s3":
		//buildEnv = clay.NewBuildEnvironmentEsp32S3(BuildPath)
	}

	if buildEnv == nil {
		return fmt.Errorf("Unsupported board: " + board)
	}

	prj := CreateProject(buildPath)
	prj.SetBuildEnvironment(buildEnv)
	AddBuildInfoAsCppLibrary(prj)
	return prj.Build()
}

func GenerateBuildInfo(board string) error {
	buildPath := GetBuildPath(board)

	appPath, _ := os.Getwd()
	if err := clay.GenerateBuildInfo(buildPath, appPath, EspSdkPath, BuildInfoFilenameWithoutExt); err == nil {
		log.Println("Ok, build info generated Ok")
	} else {
		log.Printf("Error, build info failed: %v\n", err)
	}
	return nil
}

func Clean(board string) error {
	buildPath := GetBuildPath(board)

	// Note: We should be running this from the "target/esp" directory
	// Remove all folders and files from "build/"
	if err := os.RemoveAll(buildPath + "/"); err != nil {
		return fmt.Errorf("Failed to remove build directory: %v", err)
	}

	if err := os.MkdirAll(buildPath+"/", os.ModePerm); err != nil {
		return fmt.Errorf("Failed to create build directory: %v", err)
	}

	return nil
}

func Flash(board string) error {
	buildPath := GetBuildPath(board)

	var buildEnv *clay.BuildEnvironment
	switch board {
	case "esp32":
		buildEnv = clay.NewBuildEnvironmentEsp32(buildPath)
	case "esp32s3":
		//buildEnv = clay.NewBuildEnvironmentEsp32S3(buildPath)
	}

	if buildEnv == nil {
		return fmt.Errorf("Unsupported board: " + board)
	}

	prj := CreateProject(buildPath)
	prj.SetBuildEnvironment(buildEnv)
	AddBuildInfoAsCppLibrary(prj)
	return prj.Flash()
}

func ListLibraries() error {
	prj := CreateProject("")
	fmt.Printf("Project: %s\n", prj.Name)
	for _, lib := range prj.Executable.Libraries {
		fmt.Printf("Library: %s\n", lib.Name)
		fmt.Printf("  Version: %s\n", lib.Version)
	}
	return nil
}

func ListBoards(boardName string, matches int) error {
	if matches <= 0 {
		matches = 10
	}
	boardsFilePath := filepath.Join(EspSdkPath, "boards.txt")
	return boards.PrintAllMatchingBoards(boardsFilePath, boardName, matches)
}

func ListFlashSizes(cpuName string, boardName string) error {
	boardsFilePath := filepath.Join(EspSdkPath, "boards.txt")
	return boards.PrintAllFlashSizes(boardsFilePath, cpuName, boardName)
}

// AddBuildInfoAsCppLibrary checks if 'buildinfo.h' and 'buildinfo.cpp' exist,
// if so it creates a C++ library and adds it to the project
func AddBuildInfoAsCppLibrary(prj *clay.Project) {
	name := BuildInfoFilenameWithoutExt
	hdrFilepath := filepath.Join(prj.BuildPath, name, BuildInfoFilenameWithoutExt+".h")
	srcFilepath := filepath.Join(prj.BuildPath, name, BuildInfoFilenameWithoutExt+".cpp")
	if ccode_utils.FileExists(hdrFilepath) && ccode_utils.FileExists(srcFilepath) {
		library := clay.NewCppLibrary(name, "0.1.0", name, name+".a")
		library.IncludeDirs.Add(filepath.Dir(hdrFilepath), false)
		library.AddSourceFile(srcFilepath, srcFilepath, true)
		prj.Executable.AddLibrary(library)
	}
}

// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------
// Note: Everything below here is generated from the package definition

// !!Project!!

func CreateProject(buildPath string) *clay.Project {
	prjName := "test_project"
	prjVersion := "0.1.0"
	prj := clay.NewProject(prjName, prjVersion, buildPath)
	AddLibraries(prj)
	return prj
}

func AddLibraries(prj *clay.Project) {
	{
		name := "test_lib"
		library := clay.NewCppLibrary(name, "0.1.0", name, name+".a")

		// Include directories
		library.IncludeDirs.Add("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/ccode/clay/app/clay/test_lib/include", false)

		// Define macros
		library.Defines.Add("TARGET_DEBUG")
		library.Defines.Add("TARGET_ESP32")

		// Source files of chash
		library.AddSourceFile("/Users/obnosis5/dev.go/src/github.com/jurgen-kluft/ccode/clay/app/clay/test_lib/src/test.cpp", "test.cpp", true)
		// etc..

		prj.Executable.AddLibrary(library)
	}

}
