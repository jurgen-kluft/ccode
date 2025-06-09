package clay

import (
	"flag"
	"os"
	"time"

	utils "github.com/jurgen-kluft/ccode/utils"
)

// Clay App Arduino
//    <project>: name of a project (if you have more than one project)
//    <config>: debug, release (default), final
//    <arch>: esp32 (default), esp32s3
//    <cpuName>: esp32 (default), esp32s3
//
//    Commands:
//    - build -arch <arch> -p <project> -build <config>
//    - build-info -arch <arch> -p <project> -build <config>
//    - clean -arch <arch> -p <project> -build <config>
//    - flash -arch <arch> -p <project> -build <config>
//    - list-libraries
//    - list-boards <arch>
//    - list-flash-sizes -c <cpuName> -arch <arch>

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
		utils.LogPrintf("Version: %s\n", version.Version)
	default:
		UsageAppArduino()
	}

	if err != nil {
		utils.LogFatalf("Error: %v\n", err)
	}
}

func UsageAppArduino() {
	utils.LogPrintln("Usage: clay [command] [options]")
	utils.LogPrintln("Commands:")
	utils.LogPrintln("  build-info -p <projectName> -build <projectConfig> -arch <arch>")
	utils.LogPrintln("  build -p <projectName> -build <projectConfig> -arch <arch>")
	utils.LogPrintln("  clean -p <projectName> -build <projectConfig> -arch <arch>")
	utils.LogPrintln("  flash -p <projectName> -build <projectConfig> -arch <arch>")
	utils.LogPrintln("  list-libraries")
	utils.LogPrintln("  list-boards -b <boardName> -m <matches>")
	utils.LogPrintln("  list-flash-sizes -c <cpuName> -b <boardName>")
	utils.LogPrintln("Options:")
	utils.LogPrintln("  projectName       Project name (if more than one) ")
	utils.LogPrintln("  projectConfig     Config name (debug, release, final) ")
	utils.LogPrintln("  boardName         Board name (e.g. esp32, esp32s3) ")
	utils.LogPrintln("  matches           Maximum number of boards to list")
	utils.LogPrintln("  cpuName           CPU name for listing flash sizes")
	utils.LogPrintln("  --help            Show this help message")
	utils.LogPrintln("  --version         Show version information")

	utils.LogPrintln("Examples:")
	utils.LogPrintln("  clay build-info (generates buildinfo.h and buildinfo.cpp)")
	utils.LogPrintln("  clay build-info -build debug -arch esp32s3")
	utils.LogPrintln("  clay build")
	utils.LogPrintln("  clay build -build debug -arch esp32s3")
	utils.LogPrintln("  clay clean -build debug -arch esp32s3")
	utils.LogPrintln("  clay flash -build debug-dev -arch esp32s3")
	utils.LogPrintln("  clay list-libraries")
	utils.LogPrintln("  clay list-boards -b esp32 -m 5")
	utils.LogPrintln("  clay list-flash-sizes -c esp32 -b esp32")
}

func BuildInfo(projectName string, buildConfig *Config) error {
	EspSdkPath := "/Users/obnosis5/sdk/arduino/esp32"
	if env := os.Getenv("ESP_SDK"); env != "" {
		EspSdkPath = env
	}

	prjs := ClayAppCreateProjectsFunc(buildConfig.Target.ArchAsString())
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
	utils.LogPrintln("Ok, build info generated Ok")
	return nil
}

func Flash(projectName string, buildConfig *Config) error {

	prjs := ClayAppCreateProjectsFunc(buildConfig.Target.ArchAsString())
	for _, prj := range prjs {
		prj.SetToolchain(buildConfig)
	}

	buildPath := GetBuildPath(buildConfig.GetSubDir())
	for _, prj := range prjs {
		if projectName == prj.Name || projectName == "" {
			if prj.IsExecutable && prj.Config.Matches(buildConfig) {
				utils.LogPrintf("Flashing project: %s, config: %s\n", prj.Name, prj.Config.ConfigString())
				startTime := time.Now()
				{
					if err := prj.Flash(buildConfig, buildPath); err != nil {
						return utils.LogErrorf(err, "Build failed")
					}
				}
				utils.LogPrintf("Flashing done ... (duration %s)\n", time.Since(startTime).Round(time.Second))
				utils.LogPrintln()
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
