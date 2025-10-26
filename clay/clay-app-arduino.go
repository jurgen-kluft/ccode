package clay

import (
	"flag"
	"os"
	"time"

	corepkg "github.com/jurgen-kluft/ccode/core"
	cespressif "github.com/jurgen-kluft/ccode/espressif"
)

// Clay App
//
//	<project>: name of a project (if you have more than one project)
//	<config>: debug-dev-none, release-dev-none, release-final-none
//	<arch>: esp32 (default), esp2866
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

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
func ArduinoEspSdkPath(arch string) string {
	EspSdkPath := "$HOME/sdk/arduino/esp32"
	switch arch {
	case "esp32":
		EspSdkPath = "$HOME/sdk/arduino/esp32"
		if env := os.Getenv("ESP32_SDK"); env != "" {
			EspSdkPath = env
		}
	case "esp8266":
		EspSdkPath = "$HOME/sdk/arduino/esp8266"
		if env := os.Getenv("ESP8266_SDK"); env != "" {
			EspSdkPath = env
		}
	}
	EspSdkPath = os.ExpandEnv(EspSdkPath)
	return EspSdkPath
}

func (a *App) BuildInfo() error {
	espSdkPath := ArduinoEspSdkPath(a.BuildTarget.Arch().String())
	prjs, err := a.CreateProjects(a.BuildTarget, a.BuildConfig)
	if err != nil {
		return err
	}

	for _, prj := range prjs {
		if a.Config.ProjectName == "*" || a.Config.ProjectName == prj.DevProject.Name {
			for _, cfg := range prj.Config {
				if cfg.BuildConfig.IsEqual(a.BuildConfig) {
					buildPath := prj.GetBuildPath(a.GetBuildPath(GetBuildDirname(a.BuildConfig, a.BuildTarget)))
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

func (a *App) Flash() error {
	if a.Config.ProjectName == "*" {
		return corepkg.LogErrorf(nil, "please specify a project name to flash using -p <project>")
	}

	buildPath := a.GetBuildPath(GetBuildDirname(a.BuildConfig, a.BuildTarget))

	prjs, err := a.CreateProjects(a.BuildTarget, a.BuildConfig)
	if err != nil {
		return err
	}

	for _, prj := range prjs {
		a.SetToolchain(prj, buildPath)
	}

	projectNames := []string{}
	projectMap := map[string]*Project{}
	for _, prj := range prjs {
		if prj.IsExecutable() && prj.CanBuildFor(a.BuildConfig, a.BuildTarget) {
			projectNames = append(projectNames, prj.DevProject.Name)
			projectMap[prj.DevProject.Name] = prj
		}
	}

	cm := corepkg.NewClosestMatch(projectNames, []int{2})
	closest := cm.ClosestN(a.Config.ProjectName, 1)

	for _, prjName := range closest {
		prj := projectMap[prjName]
		if prj.IsExecutable() && prj.CanBuildFor(a.BuildConfig, a.BuildTarget) && prj.DevProject.Name == closest[0] {

			corepkg.LogInff("Flashing project: %s, config: %s", prj.DevProject.Name, a.BuildConfig.String())
			startTime := time.Now()
			{
				if err := prj.Flash(a.BuildConfig, a.BuildTarget, buildPath); err != nil {
					return corepkg.LogErrorf(err, "Build failed")
				}
			}
			corepkg.LogInff("Flashing done ... (duration %s)", time.Since(startTime).Round(time.Second))
			corepkg.LogInfo()
		}
	}
	return nil
}

func (a *App) SerialMonitor(port string, baud int) error {

	return nil
}

func (a *App) ListBoards(arch string, boardName string, matches int) error {
	if matches <= 0 {
		matches = 10
	}
	if espressifToolchain, err := cespressif.ParseToolchain(arch); err != nil {
		return err
	} else {
		return cespressif.PrintAllMatchingBoards(espressifToolchain, boardName, matches)
	}
}

func (a *App) PrintBoardInfo(arch string, boardName string, matches int) error {
	if matches <= 0 {
		matches = 10
	}
	if espressifToolchain, err := cespressif.ParseToolchain(arch); err != nil {
		return err
	} else {
		return cespressif.PrintAllBoardInfos(espressifToolchain, boardName, matches)
	}
}

func (a *App) ListFlashSizes(arch string, boardName string) error {
	if espressifToolchain, err := cespressif.ParseToolchain(arch); err != nil {
		return err
	} else {
		return cespressif.PrintAllFlashSizes(espressifToolchain, arch, boardName)
	}
}
