package clay

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	utils "github.com/jurgen-kluft/ccode/utils"
)

// Clay App Arduino
//    <project>: name of a project (if you have more than one project)
//    <config>: debug, release (default), final
//    <board>: esp32 (default), esp32s3
//    <cpuName>: esp32 (default), esp32s3
//
//    Commands:
//    - build -b <board> -p <project> -c <config>
//    - build-info -b <board> -p <project> -c <config>
//    - clean -b <board> -p <project> -c <config>
//    - flash -b <board> -p <project> -c <config>
//    - list-libraries
//    - list-boards <board>
//    - list-flash-sizes -c <cpuName> -b <board>

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

func ClayAppMainArduino() {
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
		version := utils.NewVersionInfo()
		fmt.Printf("Version: %s\n", version.Version)
	default:
		UsageAppArduino()
	}

	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func UsageAppArduino() {
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

func BuildInfo(projectName string, buildConfig *Config) error {
	EspSdkPath := "/Users/obnosis5/sdk/arduino/esp32"
	if env := os.Getenv("ESP_SDK"); env != "" {
		EspSdkPath = env
	}

	prjs := ClayAppCreateProjectsFunc("build")
	for _, prj := range prjs {
		if projectName == "" || projectName == prj.Name {
			if prj.Config.Matches(buildConfig) {
				buildPath := prj.GetBuildPath(GetBuildPath(buildConfig.GetSubDir()))
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
					buildPath := GetBuildPath(buildConfig.GetSubDir())
					if err := prj.Flash(buildPath); err != nil {
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
