package clay

import (
	"flag"
	"os"
	"time"

	corepkg "github.com/jurgen-kluft/ccode/core"
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
	case "board-info":
		err = PrintBoardInfo(ParseBoardNameAndMax())
	case "generate-boards":
		err = GenerateBoards()
	case "version":
		version := corepkg.NewVersionInfo()
		corepkg.LogInff("Version: %s", version.Version)
	default:
		UsageAppArduino()
	}

	if err != nil {
		corepkg.LogFatalf("Error: %v", err)
	}
}

func UsageAppArduino() {
	corepkg.LogInfo("Usage: clay [command] [options]")
	corepkg.LogInfo("Commands:")
	corepkg.LogInfo("  build-info -p <name> -build <config> -arch <arch>")
	corepkg.LogInfo("  build -p <name> -build <config> -board <board>")
	corepkg.LogInfo("  clean -p <name> -build <config> -board <board>")
	corepkg.LogInfo("  flash -p <name> -build <config> -board <board>")
	corepkg.LogInfo("  list-libraries")
	corepkg.LogInfo("  list-boards -b <name of board> -m <matches>")
	corepkg.LogInfo("  list-flash-sizes -c <cpu> -b <name of board>")
	corepkg.LogInfo("Options:")
	corepkg.LogInfo("  name              Project name (if more than one) ")
	corepkg.LogInfo("  config            Config name (debug, release, final) ")
	corepkg.LogInfo("  board             Board name (e.g. esp32, c3, s3, xiao_esp32c3) ")
	corepkg.LogInfo("  matches           Maximum number of boards to list")
	corepkg.LogInfo("  cpu               CPU name for listing flash sizes")
	corepkg.LogInfo("  --help            Show this help message")
	corepkg.LogInfo("  --version         Show version information")

	corepkg.LogInfo("Examples:")
	corepkg.LogInfo("  clay build-info (generates buildinfo.h and buildinfo.cpp)")
	corepkg.LogInfo("  clay build-info -build debug -arch esp32 -board esp32s3")
	corepkg.LogInfo("  clay build")
	corepkg.LogInfo("  clay build -build debug -arch esp32 -board esp32s3")
	corepkg.LogInfo("  clay clean -build debug -arch esp32 -board esp32s3")
	corepkg.LogInfo("  clay flash -build debug-dev -arch esp32 -board esp32s3")
	corepkg.LogInfo("  clay list-libraries")
	corepkg.LogInfo("  clay list-boards -b esp32 -m 5")
	corepkg.LogInfo("  clay board-info -b xiao -m 2")
	corepkg.LogInfo("  clay list-flash-sizes -c esp32 -b esp32")
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
	corepkg.LogInfo("Ok, build info generated Ok")
	return nil
}

func Flash(projectName string, buildConfig *Config) error {

	esp32Toolchain := NewEsp32Toolchain()
	err := LoadBoards(esp32Toolchain)
	if err != nil {
		return err
	}

	board := esp32Toolchain.GetBoardByName(clayConfig.TargetBoard)

	prjs := ClayAppCreateProjectsFunc()
	for _, prj := range prjs {
		prj.SetToolchain(buildConfig, board)
	}

	projectNames := []string{}
	for _, prj := range prjs {
		if prj.IsExecutable && prj.Config.Matches(buildConfig) {
			projectNames = append(projectNames, prj.Name)
		}
	}

	cm := corepkg.NewClosestMatch(projectNames, []int{2})
	closest := cm.ClosestN(projectName, 1)

	buildPath := GetBuildPath(buildConfig.GetSubDir())
	for _, prj := range prjs {
		if prj.IsExecutable && prj.Config.Matches(buildConfig) && prj.Name == closest[0] {

			corepkg.LogInff("Flashing project: %s, config: %s", prj.Name, prj.Config.ConfigString())
			startTime := time.Now()
			{
				if err := prj.Flash(buildConfig, buildPath); err != nil {
					return corepkg.LogErrorf(err, "Build failed")
				}
			}
			corepkg.LogInff("Flashing done ... (duration %s)", time.Since(startTime).Round(time.Second))
			corepkg.LogInfo()
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
	esp32Toolchain := NewEsp32Toolchain()
	if err := ParseEsp32Toolchain(esp32Toolchain); err != nil {
		return err
	}
	return PrintAllMatchingBoards(esp32Toolchain, boardName, matches)
}

func GenerateBoards() error {
	esp32Toolchain := NewEsp32Toolchain()
	if err := ParseEsp32Toolchain(esp32Toolchain); err != nil {
		return err
	}
	return GenerateAllBoards(esp32Toolchain)
}

func PrintBoardInfo(boardName string, matches int) error {
	if matches <= 0 {
		matches = 10
	}
	esp32Toolchain := NewEsp32Toolchain()
	if err := ParseEsp32Toolchain(esp32Toolchain); err != nil {
		return err
	}
	return PrintAllBoardInfos(esp32Toolchain, boardName, matches)
}

func ListFlashSizes(cpuName string, boardName string) error {
	esp32Toolchain := NewEsp32Toolchain()
	if err := ParseEsp32Toolchain(esp32Toolchain); err != nil {
		return err
	}
	return PrintAllFlashSizes(esp32Toolchain, cpuName, boardName)
}
