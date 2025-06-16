package clay

import (
	"flag"
	"os"
	"time"

	"github.com/jurgen-kluft/ccode/foundation"
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
		version := foundation.NewVersionInfo()
		foundation.LogPrintf("Version: %s\n", version.Version)
	default:
		UsageAppArduino()
	}

	if err != nil {
		foundation.LogFatalf("Error: %v\n", err)
	}
}

func UsageAppArduino() {
	foundation.LogPrintln("Usage: clay [command] [options]")
	foundation.LogPrintln("Commands:")
	foundation.LogPrintln("  build-info -p <projectName> -build <projectConfig> -arch <arch>")
	foundation.LogPrintln("  build -p <projectName> -build <projectConfig> -arch <arch>")
	foundation.LogPrintln("  clean -p <projectName> -build <projectConfig> -arch <arch>")
	foundation.LogPrintln("  flash -p <projectName> -build <projectConfig> -arch <arch>")
	foundation.LogPrintln("  list-libraries")
	foundation.LogPrintln("  list-boards -b <boardName> -m <matches>")
	foundation.LogPrintln("  list-flash-sizes -c <cpuName> -b <boardName>")
	foundation.LogPrintln("Options:")
	foundation.LogPrintln("  projectName       Project name (if more than one) ")
	foundation.LogPrintln("  projectConfig     Config name (debug, release, final) ")
	foundation.LogPrintln("  boardName         Board name (e.g. esp32, esp32s3) ")
	foundation.LogPrintln("  matches           Maximum number of boards to list")
	foundation.LogPrintln("  cpuName           CPU name for listing flash sizes")
	foundation.LogPrintln("  --help            Show this help message")
	foundation.LogPrintln("  --version         Show version information")

	foundation.LogPrintln("Examples:")
	foundation.LogPrintln("  clay build-info (generates buildinfo.h and buildinfo.cpp)")
	foundation.LogPrintln("  clay build-info -build debug -arch esp32s3")
	foundation.LogPrintln("  clay build")
	foundation.LogPrintln("  clay build -build debug -arch esp32s3")
	foundation.LogPrintln("  clay clean -build debug -arch esp32s3")
	foundation.LogPrintln("  clay flash -build debug-dev -arch esp32s3")
	foundation.LogPrintln("  clay list-libraries")
	foundation.LogPrintln("  clay list-boards -b esp32 -m 5")
	foundation.LogPrintln("  clay list-flash-sizes -c esp32 -b esp32")
}

func BuildInfo(projectName string, buildConfig *Config) error {
	EspSdkPath := "/Users/obnosis5/sdk/arduino/esp32"
	if env := os.Getenv("ESP_SDK"); env != "" {
		EspSdkPath = env
	}

	prjs := ClayAppCreateProjectsFunc()
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
	foundation.LogPrintln("Ok, build info generated Ok")
	return nil
}

func Flash(projectName string, buildConfig *Config) error {

	prjs := ClayAppCreateProjectsFunc()
	for _, prj := range prjs {
		prj.SetToolchain(buildConfig)
	}

	buildPath := GetBuildPath(buildConfig.GetSubDir())
	for _, prj := range prjs {
		if projectName == prj.Name || projectName == "" {
			if prj.IsExecutable && prj.Config.Matches(buildConfig) {
				foundation.LogPrintf("Flashing project: %s, config: %s\n", prj.Name, prj.Config.ConfigString())
				startTime := time.Now()
				{
					if err := prj.Flash(buildConfig, buildPath); err != nil {
						return foundation.LogErrorf(err, "Build failed")
					}
				}
				foundation.LogPrintf("Flashing done ... (duration %s)\n", time.Since(startTime).Round(time.Second))
				foundation.LogPrintln()
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
