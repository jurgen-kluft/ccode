package clay

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
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

	if a.Config.ProjectName == "" || a.Config.ProjectName == "*" {
		a.Config.ProjectName = ""
		if len(projectNames) == 1 {
			a.Config.ProjectName = projectNames[0]
			corepkg.LogInff("Selecting project: %s", a.Config.ProjectName)
		}
	}

	if a.Config.ProjectName == "" {
		return corepkg.LogErrorf(nil, "please specify a project name to flash using -p <project>")
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

type BoardInfo struct {
	Manufacturer      string
	Device            string
	MAC               string
	ChipType          string
	CrystalFrequency  int // in MHz
	Revision          string
	WiFi              bool
	Bluetooth         bool
	Zigbee            bool
	DualCore          bool
	LowPowerCore      bool
	EmbeddedSRamSize  int // in KB
	EmbeddedPSRamSize int // in MB
	FlashSize         int // in MB
	FlashType         string
	FlashVoltage      string
}

func (info *BoardInfo) Print() {
	fmt.Println("ESP Tool Information:")
	fmt.Printf("  Manufacturer: %s\n", info.Manufacturer)
	fmt.Printf("  Device: %s\n", info.Device)
	fmt.Printf("  MAC Address: %s\n", info.MAC)
	fmt.Printf("  Chip Type: %s\n", info.ChipType)
	fmt.Printf("  Crystal Frequency: %dMHz\n", info.CrystalFrequency)
	fmt.Printf("  Revision: %s\n", info.Revision)
	if info.WiFi {
		fmt.Printf("  Feature: WiFi\n")
	}
	if info.Bluetooth {
		fmt.Printf("  Feature: Bluetooth\n")
	}
	if info.Zigbee {
		fmt.Printf("  Feature: Zigbee\n")
	}
	if info.DualCore {
		fmt.Printf("  Feature: Dual-Core\n")
	} else {
		fmt.Printf("  Feature: Single-Core\n")
	}
	if info.LowPowerCore {
		fmt.Printf("  Feature: Low-Power Core\n")
	}
	if info.EmbeddedSRamSize > 0 {
		fmt.Printf("  Embedded SRAM Size: %dKB\n", info.EmbeddedSRamSize)
	}
	if info.EmbeddedPSRamSize > 0 {
		fmt.Printf("  Embedded PSRAM Size: %dMB\n", info.EmbeddedPSRamSize)
	}
	fmt.Printf("  Flash Size: %dMB\n", info.FlashSize)
	fmt.Printf("  Flash Type: %s\n", info.FlashType)
	fmt.Printf("  Flash Voltage: %s\n", info.FlashVoltage)
}

func (a *App) IdentifyBoard() (info *BoardInfo, err error) {
	// This will write a board.json file to the clay directory that contains
	// the identified board information:
	// - Manufacturer
	// - Device
	// - Chip type
	// - Crystal frequency
	// - Flash mode
	// - Flash speed
	// - Detected flash size
	// - Features: Bluetooth, WiFi, Dual-Core, LP-Core, Embedded PSRAM size
	// - MAC address
	var espressifToolchain *cespressif.Toolchain
	if espressifToolchain, err = cespressif.ParseToolchain(a.Config.TargetArch); err != nil {
		return nil, err
	}

	espressifToolchain.ResolveVars(a.Config.TargetArch)

	espToolPath, ok := espressifToolchain.GetToolPath("esptool_py")
	if !ok {
		return nil, fmt.Errorf("esptool_py not found in toolchain for arch %s", a.Config.TargetArch)
	}

	info = &BoardInfo{}

	// Launch esptool flash-id
	fmt.Println("Collecting ESP info using esptool...")

	cmd := exec.Command(espToolPath, "flash-id")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("failed to run esptool" + string(out) + err.Error())
		return nil, fmt.Errorf("failed to run esptool: %w", err)
	}

	// Normalize line endings
	output := strings.ReplaceAll(string(out), "\r\n", "\n")

	// Split output into lines
	lines := strings.Split(output, "\n")

	// Delete empty lines
	lines = slices.DeleteFunc(lines, func(s string) bool { return len(strings.TrimSpace(s)) == 0 })

	// Regex patterns
	chipRegex := regexp.MustCompile(`Chip type:\s+(.+)\(revision (.+)\)`)
	featuresRegex := regexp.MustCompile(`Features:\s+(.+)`)
	crystalRegex := regexp.MustCompile(`Crystal frequency:\s+(.+)`)
	macRegex := regexp.MustCompile(`MAC:\s+(.+)`)
	manufacturerRegex := regexp.MustCompile(`Manufacturer:\s+(.+)`)
	deviceRegex := regexp.MustCompile(`Device:\s+(.+)`)
	flashSizeRegex := regexp.MustCompile(`Detected flash size:\s+(.+)`)
	flashTypeRegex := regexp.MustCompile(`Flash type set in eFuse:\s+(.+)`)
	flashVoltageRegex := regexp.MustCompile(`Flash voltage set by eFuse:\s+(.+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case chipRegex.MatchString(line):
			matches := chipRegex.FindStringSubmatch(line)
			info.ChipType = strings.TrimSpace(matches[1])
			info.Revision = matches[2]
		case featuresRegex.MatchString(line):
			features := featuresRegex.FindStringSubmatch(line)[1]
			fmt.Printf("Detected features: %s\n", features)
			for feature := range strings.SplitSeq(features, ", ") {
				switch feature {
				case "Wi-Fi":
					info.WiFi = true
				case "BT 5 (LE)":
					info.Bluetooth = true
				case "Zigbee":
					info.Zigbee = true
				case "Dual Core":
					info.DualCore = true
				case "Dual Core + LP Core":
					info.DualCore = true
					info.LowPowerCore = true
				default:
					if strings.HasPrefix(feature, "Embedded PSRAM") {
						var sizeMB int
						fmt.Sscanf(feature, "Embedded PSRAM %dMB", &sizeMB)
						info.EmbeddedPSRamSize = sizeMB
					}
				}
			}
		case crystalRegex.MatchString(line):
			freq := crystalRegex.FindStringSubmatch(line)[1]
			fmt.Sscanf(freq, "%dMHz", &info.CrystalFrequency)
		case macRegex.MatchString(line):
			info.MAC = macRegex.FindStringSubmatch(line)[1]
		case manufacturerRegex.MatchString(line):
			info.Manufacturer = manufacturerRegex.FindStringSubmatch(line)[1]
			if info.Manufacturer == "ef" {
				info.Manufacturer = "Espressif"
			}
		case deviceRegex.MatchString(line):
			info.Device = deviceRegex.FindStringSubmatch(line)[1]
		case flashSizeRegex.MatchString(line):
			flashSize := flashSizeRegex.FindStringSubmatch(line)[1]
			var size int
			fmt.Sscanf(flashSize, "%dMB", &size)
			info.FlashSize = size
		case flashTypeRegex.MatchString(line):
			info.FlashType = flashTypeRegex.FindStringSubmatch(line)[1]
		case flashVoltageRegex.MatchString(line):
			info.FlashVoltage = flashVoltageRegex.FindStringSubmatch(line)[1]
		}
	}

	sramSizesInKB := map[string]int{
		"esp32":   520,
		"esp32s2": 320,
		"esp32s3": 512,
		"esp32c3": 400,
		"esp32c6": 400,
		"esp32h2": 320,
		"esp32c2": 272,
		"esp32c5": 400,
	}

	// Using the chip type, determine the internal SRAM size. Chip Type might not
	// be formatted correctly, example: "ESP32-S3 (QFN56)", so we need to clean it up.
	cleanChipType := strings.ToLower(info.ChipType)
	cleanChipType = strings.SplitN(cleanChipType, " ", 2)[0]
	cleanChipType = strings.ReplaceAll(cleanChipType, "-", "")
	if sramSize, ok := sramSizesInKB[cleanChipType]; ok {
		fmt.Printf("Detected internal SRAM size: %dKB\n", sramSize/1024)
		info.EmbeddedSRamSize = sramSize
	} else {
		fmt.Printf("Unknown internal SRAM size for chip type: %s\n", info.ChipType)
		info.EmbeddedSRamSize = 320 * 1024 // default to 320KB
	}

	// Write to board_info.json in the clay directory
	stringBuilder := corepkg.NewStringBuilder()
	stringBuilder.WriteLn("{")
	stringBuilder.WriteLn(`  "manufacturer": "`, info.Manufacturer, `",`)
	stringBuilder.WriteLn(`  "device": "`, info.Device, `",`)
	stringBuilder.WriteLn(`  "mac": "`, info.MAC, `",`)
	stringBuilder.WriteLn(`  "chip_type": "`, info.ChipType, `",`)
	stringBuilder.WriteLn(`  "crystal_frequency_mhz": `, strconv.Itoa(info.CrystalFrequency), `,`)
	stringBuilder.WriteLn(`  "revision": "`, info.Revision, `",`)
	stringBuilder.WriteLn(`  "wifi": "`, strconv.FormatBool(info.WiFi), `",`)
	stringBuilder.WriteLn(`  "bluetooth": "`, strconv.FormatBool(info.Bluetooth), `",`)
	stringBuilder.WriteLn(`  "zigbee": "`, strconv.FormatBool(info.Zigbee), `",`)
	stringBuilder.WriteLn(`  "dual_core": "`, strconv.FormatBool(info.DualCore), `",`)
	stringBuilder.WriteLn(`  "low_power_core": "`, strconv.FormatBool(info.LowPowerCore), `",`)
	stringBuilder.WriteLn(`  "embedded_sram_size_kb": `, strconv.Itoa(info.EmbeddedSRamSize), `,`)
	stringBuilder.WriteLn(`  "embedded_psram_size_mb": `, strconv.Itoa(info.EmbeddedPSRamSize), `,`)
	stringBuilder.WriteLn(`  "flash_size_mb": `, strconv.Itoa(info.FlashSize), `,`)
	stringBuilder.WriteLn(`  "flash_type": "`, info.FlashType, `",`)
	stringBuilder.WriteLn(`  "flash_voltage": "`, info.FlashVoltage, `"`)
	stringBuilder.WriteLn("}")

	boardInfoPath := "board_info.json"
	fmt.Println("Writing board information to " + boardInfoPath)
	if err := os.WriteFile(boardInfoPath, []byte(stringBuilder.String()), 0644); err != nil {
		return nil, fmt.Errorf("failed to write board info to %s: %w", boardInfoPath, err)
	}

	return info, nil

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
