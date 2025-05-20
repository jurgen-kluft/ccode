package clay

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	cutils "github.com/jurgen-kluft/ccode/cutils"
)

// Clay App
//    <project>: name of a project (if you have more than one project)
//    <config>: debug, release (default), final
//    <board>: esp32 (default), esp32s3
//    <cpuName>: esp32 (default), esp32s3
//
//    Commands:
//    - build -b <board> -p <project> -x <config>
//    - build-info -b <board> -p <project> -x <config>
//    - clean -b <board> -p <project> -x <config>
//    - flash -b <board> -p <project> -x <config>
//    - list-libraries
//    - list-boards <board>
//    - list-flash-sizes -c <cpuName> -b <board>

const (
	BuildInfoFilenameWithoutExt = "buildinfo"
)

var ClayAppCreateProjectsFunc func(buildPath string) []*Project

func GetBuildPath(board string) string {
	// /build/esp32
	// /build/esp32s3
	buildPath := filepath.Join("build", board)
	return buildPath
}

func ParseBoardNameAndMax() (string, int) {
	var boardName string
	var matches int
	flag.StringVar(&boardName, "b", "esp32", "Board name (esp32, esp32s3)")
	flag.IntVar(&matches, "m", 10, "Maximum number of boards to list")
	flag.Parse()
	return boardName, matches
}

func ParseCpuAndBoardName() (string, string) {
	var cpu string
	var boardName string
	flag.StringVar(&cpu, "c", "esp32", "CPU name (esp32, esp32s3)")
	flag.StringVar(&boardName, "b", "esp32", "Board name (esp32, esp32s3)")
	flag.Parse()
	return cpu, boardName
}

func ParseProjectNameBoardNameAndConfig() (string, string, string) {
	var projectName string
	var projectConfig string
	var boardName string
	flag.StringVar(&projectName, "p", "", "Name of the project")
	flag.StringVar(&projectConfig, "x", "", "Build configuration (debug, release, final)")
	flag.StringVar(&boardName, "b", "esp32", "Board name (esp32, esp32s3)")
	flag.Parse()
	return projectName, projectConfig, boardName
}

func ClayAppMain() {
	// Consume the first argument as the command
	command := os.Args[1]
	os.Args = os.Args[1:]

	// Parse command line arguments
	var err error
	switch command {
	case "build":
		err = Build(ParseProjectNameBoardNameAndConfig())
	case "build-info":
		err = BuildInfo(ParseProjectNameBoardNameAndConfig())
	case "clean":
		err = Clean(ParseProjectNameBoardNameAndConfig())
	case "flash":
		err = Flash(ParseProjectNameBoardNameAndConfig())
	case "list-libraries":
		err = ListLibraries()
	case "list-boards":
		err = ListBoards(ParseBoardNameAndMax())
	case "list-flash-sizes":
		err = ListFlashSizes(ParseCpuAndBoardName())
	case "version":
		version := cutils.NewVersionInfo()
		fmt.Printf("Version: %s\n", version.Version)
	default:
		Usage()
	}

	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func Usage() {
	fmt.Println("Usage: clay [command] [options]")
	fmt.Println("Commands:")
	fmt.Println("  build-info -p <projectName> -x <projectConfig> -b <boardName>")
	fmt.Println("  build -p <projectName> -x <projectConfig> -b <boardName>")
	fmt.Println("  clean -p <projectName> -x <projectConfig> -b <boardName>")
	fmt.Println("  flash -p <projectName> -x <projectConfig> -b <boardName>")
	fmt.Println("  list-libraries")
	fmt.Println("  list-boards -b <boardName> -m <matches>")
	fmt.Println("  list-flash-sizes -c <cpuName> -b <boardName>")
	fmt.Println("Options:")
	fmt.Println("  projectName       Project name (if more than one) ")
	fmt.Println("  configName        Config name (debug, release, final) ")
	fmt.Println("  boardName         Board name (e.g. esp32, esp32s3) ")
	fmt.Println("  matches           Maximum number of boards to list")
	fmt.Println("  cpuName           CPU name for listing flash sizes")
	fmt.Println("  --help            Show this help message")
	fmt.Println("  --version         Show version information")

	fmt.Println("Examples:")
	fmt.Println("  clay build-info (generates buildinfo.h and buildinfo.cpp)")
	fmt.Println("  clay build-info -x debug -b esp32s3")
	fmt.Println("  clay build")
	fmt.Println("  clay build -x debug -b esp32s3")
	fmt.Println("  clay clean -x debug -b esp32s3")
	fmt.Println("  clay flash -x debug -b esp32s3")
	fmt.Println("  clay list-libraries")
	fmt.Println("  clay list-boards -b esp32 -m 5")
	fmt.Println("  clay list-flash-sizes -c esp32 -b esp32")
}

func Build(projectName string, projectConfig string, board string) error {
	// Note: We should be running this from the "target/esp" directory
	// Create the build directory
	buildPath := GetBuildPath(board)
	os.MkdirAll(buildPath+"/", os.ModePerm)

	var buildEnv *BuildEnvironment
	switch board {
	case "esp32":
		buildEnv = NewBuildEnvironmentEsp32(buildPath)
	case "esp32s3":
		//buildEnv = clay.NewBuildEnvironmentEsp32S3(BuildPath)
	}

	if buildEnv == nil {
		return fmt.Errorf("Unsupported board: " + board)
	}

	prjs := ClayAppCreateProjectsFunc(buildPath)
	for _, prj := range prjs {
		if projectName == "" || projectName == prj.Name {
			if projectConfig == "" || strings.EqualFold(prj.Config, projectConfig) {
				log.Printf("Building project: %s, config: %s\n", prj.Name, prj.Config)
				startTime := time.Now()
				{
					prj.SetBuildEnvironment(buildEnv)
					AddBuildInfoAsCppLibrary(prj)
					if err := prj.Build(); err != nil {
						return fmt.Errorf("Build failed on project %s with config %s: %v", prj.Name, prj.Config, err)
					}
				}
				log.Printf("Building done ... (duration %s s)\n", time.Since(startTime).Round(time.Second))
				log.Println()
			}
		}
	}
	return nil
}

func BuildInfo(projectName string, projectConfig string, board string) error {
	EspSdkPath := "/Users/obnosis5/sdk/arduino/esp32"
	if env := os.Getenv("ESP_SDK"); env != "" {
		EspSdkPath = env
	}

	prjs := ClayAppCreateProjectsFunc(GetBuildPath(board))
	for _, prj := range prjs {
		if projectName == "" || projectName == prj.Name {
			if projectConfig == "" || projectConfig == prj.Config {
				buildPath := prj.BuildPath
				appPath, _ := os.Getwd()
				if err := GenerateBuildInfo(buildPath, appPath, EspSdkPath, BuildInfoFilenameWithoutExt); err != nil {
					return err
				}
			}
		}
	}
	log.Println("Ok, build info generated Ok")
	return nil
}

func Clean(projectName string, projectConfig string, board string) error {
	buildPath := GetBuildPath(board)

	prjs := ClayAppCreateProjectsFunc(buildPath)
	for _, prj := range prjs {
		if projectName == "" || projectName == prj.Name {
			if projectConfig == "" || projectConfig == prj.Config {
				buildPath = filepath.Join(buildPath, prj.Name, projectConfig)

				// Note: We should be running this from the "target/esp" directory
				// Remove all folders and files from "build/"
				if err := os.RemoveAll(buildPath + "/"); err != nil {
					return fmt.Errorf("Failed to remove build directory: %v", err)
				}

				if err := os.MkdirAll(buildPath+"/", os.ModePerm); err != nil {
					return fmt.Errorf("Failed to create build directory: %v", err)
				}
			}
		}
	}

	return nil
}

func Flash(project string, board string, config string) error {
	buildPath := GetBuildPath(board)

	var buildEnv *BuildEnvironment
	switch board {
	case "esp32":
		buildEnv = NewBuildEnvironmentEsp32(buildPath)
	case "esp32s3":
		//buildEnv = clay.NewBuildEnvironmentEsp32S3(buildPath)
	}

	if buildEnv == nil {
		return fmt.Errorf("Unsupported board: " + board)
	}

	prjs := ClayAppCreateProjectsFunc(buildPath)
	for _, prj := range prjs {
		if project == prj.Name || project == "" {
			prj.SetBuildEnvironment(buildEnv)
			if err := prj.Flash(); err != nil {
				return fmt.Errorf("Build failed: %v", err)
			}
		}
	}
	return nil
}

func ListLibraries() error {
	buildPath := ""
	prjs := ClayAppCreateProjectsFunc(buildPath)

	configs := make([]string, 0, 16)
	nameToIndex := make(map[string]int)
	for _, prj := range prjs {
		if idx, ok := nameToIndex[prj.Name]; !ok {
			idx = len(configs)
			nameToIndex[prj.Name] = idx
			configs = append(configs, prj.Config)
		} else {
			configs[idx] += ", " + prj.Config
		}
	}

	for _, prj := range prjs {
		if i, ok := nameToIndex[prj.Name]; ok {
			fmt.Printf("Project: %s\n", prj.Name)
			fmt.Printf("Configs: %s\n", configs[i])
			fmt.Printf("  Libraries:\n")
			for _, lib := range prj.Executable.Libraries {
				fmt.Printf("  - %s\n", lib.Name)
			}

			// Remove the entry from the map to avoid duplicates
			delete(nameToIndex, prj.Name)
		}
	}

	return nil
}

func ListBoards(boardName string, matches int) error {
	if matches <= 0 {
		matches = 10
	}
	EspSdkPath := "/Users/obnosis5/sdk/arduino/esp32"
	if env := os.Getenv("ESP_SDK"); env != "" {
		EspSdkPath = env
	}

	boardsFilePath := filepath.Join(EspSdkPath, "boards.txt")
	return PrintAllMatchingBoards(boardsFilePath, boardName, matches)
}

func ListFlashSizes(cpuName string, boardName string) error {
	EspSdkPath := "/Users/obnosis5/sdk/arduino/esp32"
	if env := os.Getenv("ESP_SDK"); env != "" {
		EspSdkPath = env
	}

	boardsFilePath := filepath.Join(EspSdkPath, "boards.txt")
	return PrintAllFlashSizes(boardsFilePath, cpuName, boardName)
}

// AddBuildInfoAsCppLibrary checks if 'buildinfo.h' and 'buildinfo.cpp' exist,
// if so it creates a C++ library and adds it to the project
func AddBuildInfoAsCppLibrary(prj *Project) {
	name := BuildInfoFilenameWithoutExt
	hdrFilepath := filepath.Join(prj.BuildPath, name+".h")
	srcFilepath := filepath.Join(prj.BuildPath, name+".cpp")
	if cutils.FileExists(hdrFilepath) && cutils.FileExists(srcFilepath) {
		library := NewCppLibrary(name, "0.1.0", name, name+".a")
		library.IncludeDirs.Add(filepath.Dir(hdrFilepath), false)
		library.AddSourceFile(srcFilepath, filepath.Base(srcFilepath), true)
		prj.Executable.AddLibrary(library)
	}
}
