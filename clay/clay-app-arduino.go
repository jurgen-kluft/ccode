package clay

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jurgen-kluft/ccode/clay/toolchain"
	corepkg "github.com/jurgen-kluft/ccode/core"
	"github.com/jurgen-kluft/ccode/denv"
)

// Clay App
//
//	<project>: name of a project (if you have more than one project)
//	<config>: debug, release (default), final
//	<arch>: esp32 (default), esp32s3
//	<board>: esp32 (default), esp32s3
//
//	Commands:
//	- build -arch <arch> -p <project> -build <config>
//	- build-info -arch <arch> -p <project> -build <config>
//	- clean -arch <arch> -p <project> -build <config>
//	- flash -arch <arch> -p <project> -build <config>
//	- list-libraries
//	- list-boards <arch>
//	- list-flash-sizes -c <cpuName> -arch <arch>

func ParseArch() string {
	var arch string
	flag.StringVar(&arch, "arch", "", "Architecture (x64, amd64, arm64, esp32, esp8266)")
	flag.Parse()
	return arch
}

func ParseArchBoardNameAndMax() (string, string, int) {
	var arch string
	var boardName string
	var matches int
	flag.StringVar(&arch, "arch", "", "Architecture (x64, amd64, arm64, esp32, esp8266)")
	flag.StringVar(&boardName, "board", "esp32", "Board name (esp32, esp32s3)")
	flag.IntVar(&matches, "max", 10, "Maximum number of boards to list")
	flag.Parse()
	return arch, boardName, matches
}

func ParsePortAndBaud() (string, int) {
	var port string
	var baud int
	flag.StringVar(&port, "p", "/dev/ttyUSB0", "Serial port (e.g. /dev/ttyUSB0)")
	flag.IntVar(&baud, "b", 115200, "Baud rate (e.g. 115200)")
	flag.Parse()
	return port, baud
}

func ParseArchAndBoardName() (string, string) {
	var arch string
	var boardName string
	flag.StringVar(&arch, "arch", "", "architecture (x64, amd64, arm64, esp32, esp8266)")
	flag.StringVar(&boardName, "board", "esp32", "Board name (esp32, esp32s3, generic)")
	flag.Parse()
	return arch, boardName
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
		err = ListLibraries(denv.BuildTargetFromString(fmt.Sprintf("%s(%s)", clayConfig.TargetOs, clayConfig.TargetArch)))
	case "list-boards":
		err = ListBoards(ParseArchBoardNameAndMax())
	case "list-flash-sizes":
		err = ListFlashSizes(ParseArchAndBoardName())
	case "board-info":
		err = PrintBoardInfo(ParseArchBoardNameAndMax())
	case "generate-boards":
		err = GenerateBoards(ParseArch())
	case "version":
		version := corepkg.NewVersionInfo()
		corepkg.LogInff("Version: %s", version.Version)
	default:
		UsageApp()
	}

	if err != nil {
		corepkg.LogFatalf("Error: %v", err)
	}
}

func UsageApp() {
	corepkg.LogInfo("Usage: clay [command] [options]")
	corepkg.LogInfo("Commands:")
	corepkg.LogInfo("  build-info -p <name> --build <config> --arch <arch>")
	corepkg.LogInfo("  build -p <name> --arch <arch> --build <config> --board <board>")
	corepkg.LogInfo("  clean -p <name> --arch <arch> --build <config> --board <board>")
	corepkg.LogInfo("  flash -p <name> --arch <arch> --build <config> --board <board>")
	corepkg.LogInfo("  list-libraries")
	corepkg.LogInfo("  list-boards --arch <arch> --board <name of board> --max <matches>")
	corepkg.LogInfo("  list-flash-sizes --arch <arch> --board <name of board>")
	corepkg.LogInfo("Options:")
	corepkg.LogInfo("  name              Project name (if more than one) ")
	corepkg.LogInfo("  config            Config name (debug, release, final) ")
	corepkg.LogInfo("  board             Board name for Arduino (e.g. esp32, c3, s3, xiao_esp32c3) ")
	corepkg.LogInfo("  matches           Maximum number of boards to list")
	corepkg.LogInfo("  arch              Architecture for listing flash sizes (esp32 or esp8266)")
	corepkg.LogInfo("  --help            Show this help message")
	corepkg.LogInfo("  --version         Show version information")

	corepkg.LogInfo("Examples:")
	corepkg.LogInfo("  clay build-info (generates buildinfo.h and buildinfo.cpp)")
	corepkg.LogInfo("  clay build-info --build debug --arch esp32 --board esp32s3")
	corepkg.LogInfo("  clay build")
	corepkg.LogInfo("  clay build --build debug --arch esp32 --board esp32s3")
	corepkg.LogInfo("  clay clean --build debug --arch esp32 --board esp32s3")
	corepkg.LogInfo("  clay flash --build debug-dev --arch esp32 --board esp32s3")
	corepkg.LogInfo("  clay list-libraries")
	corepkg.LogInfo("  clay list-boards --arch <arch> --board esp32 --max 5")
	corepkg.LogInfo("  clay board-info --arch <arch> --board xiao --max 2")
	corepkg.LogInfo("  clay list-flash-sizes --arch <arch> --board esp32")
}

func BuildInfo(projectName string, buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) error {
	espSdkPath := toolchain.ArduinoEspSdkPath(buildTarget.Arch().String())

	prjs, err := CreateProjects(buildTarget, buildConfig)
	if err != nil {
		return err
	}

	for _, prj := range prjs {
		if projectName == "" || projectName == prj.DevProject.Name {
			for _, cfg := range prj.Config {
				if cfg.BuildConfig.IsEqual(buildConfig) {
					buildPath := prj.GetBuildPath(GetBuildPath(GetBuildDirname(buildConfig, buildTarget)))
					appPath, _ := os.Getwd()
					if err := GenerateBuildInfo(buildPath, appPath, espSdkPath, BuildInfoFilenameWithoutExt); err != nil {
						return err
					}
				}
			}
		}
	}
	corepkg.LogInfo("Ok, build info generated Ok")
	return nil
}

func Flash(projectName string, buildConfig denv.BuildConfig, buildTarget denv.BuildTarget) error {
	buildPath := GetBuildPath(GetBuildDirname(buildConfig, buildTarget))

	prjs, err := CreateProjects(buildTarget, buildConfig)
	if err != nil {
		return err
	}

	for _, prj := range prjs {
		prj.SetToolchain(buildConfig, buildTarget, buildPath)
	}

	projectNames := []string{}
	for _, prj := range prjs {
		if prj.IsExecutable() && prj.CanBuildFor(buildConfig, buildTarget) {
			projectNames = append(projectNames, prj.DevProject.Name)
		}
	}

	cm := corepkg.NewClosestMatch(projectNames, []int{2})
	closest := cm.ClosestN(projectName, 1)

	for _, prj := range prjs {
		if prj.IsExecutable() && prj.CanBuildFor(buildConfig, buildTarget) && prj.DevProject.Name == closest[0] {

			corepkg.LogInff("Flashing project: %s, config: %s", prj.DevProject.Name, buildConfig.String())
			startTime := time.Now()
			{
				if err := prj.Flash(buildConfig, buildTarget, buildPath); err != nil {
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

func ListBoards(arch string, boardName string, matches int) error {
	if matches <= 0 {
		matches = 10
	}
	espressifToolchain := NewEspressifToolchain(arch)
	if espressifToolchain == nil {
		return corepkg.LogErrorf(nil, "Unsupported architecture: %s", arch)
	}
	if err := ParseToolchainFiles(espressifToolchain); err != nil {
		return err
	}
	return PrintAllMatchingBoards(espressifToolchain, boardName, matches)
}

func GenerateBoards(arch string) error {
	espressifToolchain := NewEspressifToolchain(arch)
	if espressifToolchain == nil {
		return corepkg.LogErrorf(nil, "Unsupported architecture: %s", arch)
	}
	if err := ParseToolchainFiles(espressifToolchain); err != nil {
		return err
	}
	if err := GenerateToolchainJson(espressifToolchain, arch+".json"); err != nil {
		return err
	}
	return corepkg.LogErrorf(nil, "Unsupported architecture: %s", clayConfig.TargetArch)
}

func PrintBoardInfo(arch string, boardName string, matches int) error {
	if matches <= 0 {
		matches = 10
	}
	espressifToolchain := NewEspressifToolchain(arch)
	if espressifToolchain == nil {
		return corepkg.LogErrorf(nil, "Unsupported architecture: %s", arch)
	}
	if err := ParseToolchainFiles(espressifToolchain); err != nil {
		return err
	}
	return PrintAllBoardInfos(espressifToolchain, boardName, matches)
}

func ListFlashSizes(arch string, boardName string) error {
	espressifToolchain := NewEspressifToolchain(arch)
	if espressifToolchain == nil {
		return corepkg.LogErrorf(nil, "Unsupported architecture: %s", arch)
	}
	if err := ParseToolchainFiles(espressifToolchain); err != nil {
		return err
	}
	return PrintAllFlashSizes(espressifToolchain, arch, boardName)
}
