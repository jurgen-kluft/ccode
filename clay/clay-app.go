package clay

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	cutils "github.com/jurgen-kluft/ccode/utils"
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

func GetBuildPath(subdir string) string {
	buildPath := filepath.Join("build", subdir)
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

func ParsePortAndBaud() (string, int) {
	var port string
	var baud int
	flag.StringVar(&port, "p", "/dev/ttyUSB0", "Serial port (e.g. /dev/ttyUSB0)")
	flag.IntVar(&baud, "b", 115200, "Baud rate (e.g. 115200)")
	flag.Parse()
	return port, baud
}

func ParseCpuAndBoardName() (string, string) {
	var cpu string
	var boardName string
	flag.StringVar(&cpu, "c", "esp32", "CPU name (esp32, esp32s3)")
	flag.StringVar(&boardName, "b", "esp32", "Board name (esp32, esp32s3)")
	flag.Parse()
	return cpu, boardName
}

func ParseProjectNameAndConfig() (string, *Config) {
	var projectName string
	var targetOs string
	var targetArch string
	var targetBuild string
	var targetVariant string
	flag.StringVar(&projectName, "p", "", "Name of the project")
	flag.StringVar(&targetOs, "os", "", "Target OS (windows, darwin, linux, arduino)")
	flag.StringVar(&targetBuild, "build", "release-dev", "Build configuration (debug-dev, release-test, final-prod)")
	flag.StringVar(&targetArch, "arch", "", "Cpu Architecture (amd64, x64, arm64, esp32, esp32s3)")
	flag.Parse()

	configParts := strings.Split(targetBuild, "-")

	if len(configParts) != 2 {
		targetBuild = "release"
		targetVariant = "dev"
	} else {
		targetBuild = configParts[0]
		targetVariant = configParts[1]
	}

	if runtime.GOOS == "windows" {
		targetOs = "windows"
	} else if runtime.GOOS == "darwin" {
		targetOs = "darwin"
	} else {
		targetOs = "linux"
	}

	if strings.HasPrefix(targetArch, "esp32") {
		targetOs = "arduino"
	}

	if targetArch == "" {
		targetArch = runtime.GOARCH
		if targetOs == "arduino" {
			targetArch = "esp32"
		} else if targetOs == "darwin" {
			targetArch = "arm64"
		} else if targetOs == "windows" {
			targetArch = "x64"
		} else if targetOs == "linux" {
			targetArch = "amd64"
		}
	}

	config := NewConfig(targetOs, targetArch, targetBuild, targetVariant)
	return projectName, config
}

func ClayAppMain() {
	// Consume the first argument as the command
	command := os.Args[1]
	os.Args = os.Args[1:]

	// Parse command line arguments
	var err error
	switch command {
	case "build":
		err = Build(ParseProjectNameAndConfig())
	case "build-info":
		err = BuildInfo(ParseProjectNameAndConfig())
	case "clean":
		err = Clean(ParseProjectNameAndConfig())
	case "flash":
		err = Flash(ParseProjectNameAndConfig())
	case "monitor":
		err = SerialMonitor(ParsePortAndBaud())
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
	fmt.Println("  clay build-info -c debug -b esp32s3")
	fmt.Println("  clay build")
	fmt.Println("  clay build -c debug -b esp32s3")
	fmt.Println("  clay clean -c debug -b esp32s3")
	fmt.Println("  clay flash -c debug -b esp32s3")
	fmt.Println("  clay list-libraries")
	fmt.Println("  clay list-boards -b esp32 -m 5")
	fmt.Println("  clay list-flash-sizes -c esp32 -b esp32")
}

func Build(projectName string, targetConfig *Config) error {
	// Note: We should be running this from the "target/{build target}" directory
	// Create the build directory
	buildPath := GetBuildPath(targetConfig.GetSubDir())
	os.MkdirAll(buildPath+"/", os.ModePerm)

	prjs := ClayAppCreateProjectsFunc(buildPath)
	for _, prj := range prjs {
		prj.SetToolchain(targetConfig)
	}

	for _, prj := range prjs {
		if projectName == "" || projectName == prj.Name {
			if prj.Config.Matches(targetConfig) {
				log.Printf("Building project: %s, config: %s-%s, arch: %s\n", prj.Name, prj.Config.Config.Build(), prj.Config.Config.Variant(), prj.Config.Target.ArchAsString())
				startTime := time.Now()
				{
					AddBuildInfoAsCppLibrary(prj)
					if err := prj.Build(); err != nil {
						return fmt.Errorf("Build failed on project %s with config %s: %v", prj.Name, prj.Config, err)
					}
				}
				log.Printf("Building done ... (duration %s)\n", time.Since(startTime).Round(time.Second))
				log.Println()
			}
		}
	}
	return nil
}

func BuildInfo(projectName string, buildConfig *Config) error {
	EspSdkPath := "/Users/obnosis5/sdk/arduino/esp32"
	if env := os.Getenv("ESP_SDK"); env != "" {
		EspSdkPath = env
	}

	prjs := ClayAppCreateProjectsFunc("build")
	for _, prj := range prjs {
		if projectName == "" || projectName == prj.Name {
			if prj.Config.Matches(buildConfig) {
				buildPath := prj.GetBuildPath()
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

func Clean(projectName string, buildConfig *Config) error {
	prjs := ClayAppCreateProjectsFunc("build")
	for _, prj := range prjs {
		if projectName == "" || projectName == prj.Name {
			if prj.Config.Matches(buildConfig) {

				buildPath := prj.GetBuildPath()

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

func Flash(projectName string, buildConfig *Config) error {

	prjs := ClayAppCreateProjectsFunc("build")
	for _, prj := range prjs {
		prj.SetToolchain(buildConfig)
	}

	for _, prj := range prjs {
		if projectName == prj.Name || projectName == "" {
			if prj.Config.Matches(buildConfig) {
				log.Printf("Flashing project: %s, config: %s\n", prj.Name, prj.Config)
				startTime := time.Now()
				{
					if err := prj.Flash(); err != nil {
						return fmt.Errorf("Build failed: %v", err)
					}
				}
				log.Printf("Flashing done ... (duration %s)\n", time.Since(startTime).Round(time.Second))
				log.Println()
			}
		}
	}
	return nil
}

func SerialMonitor(port string, baud int) error {

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
			configs = append(configs, prj.Config.Config.Build()+"-"+prj.Config.Config.Variant())
		} else {
			configs[idx] += ", " + prj.Config.Config.Build() + "-" + prj.Config.Config.Variant()
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

	return PrintAllMatchingBoards(EspSdkPath, boardName, matches)
}

func ListFlashSizes(cpuName string, boardName string) error {
	EspSdkPath := "/Users/obnosis5/sdk/arduino/esp32"
	if env := os.Getenv("ESP_SDK"); env != "" {
		EspSdkPath = env
	}

	return PrintAllFlashSizes(EspSdkPath, cpuName, boardName)
}

// AddBuildInfoAsCppLibrary checks if 'buildinfo.h' and 'buildinfo.cpp' exist,
// if so it creates a C++ library and adds it to the project
func AddBuildInfoAsCppLibrary(prj *Project) {
	name := BuildInfoFilenameWithoutExt
	hdrFilepath := filepath.Join(prj.GetBuildPath(), name+".h")
	srcFilepath := filepath.Join(prj.GetBuildPath(), name+".cpp")
	if cutils.FileExists(hdrFilepath) && cutils.FileExists(srcFilepath) {
		library := NewLibrary(name, prj.Config)
		library.IncludeDirs.Add(filepath.Dir(hdrFilepath))
		library.AddSourceFile(srcFilepath, filepath.Base(srcFilepath))
		prj.Executable.AddLibrary(library)
	}
}
