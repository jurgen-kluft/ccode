package axe

type VisualStudioConfig struct {
	Version                      EnumVisualStudio
	SlnHeader                    string
	ProjectTools                 string // e.g. 14.0
	PlatformToolset              string // e.g. v140
	WindowsTargetPlatformVersion string // e.g. 10.0
}

// ----------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------

const slnFileHeader2015 = `Microsoft Visual Studio Solution File, Format Version 12.00\r\n"
					"# Visual Studio 14\r\n"
					"VisualStudioVersion = 14.0.25420.1\r\n"
					"MinimumVisualStudioVersion = 10.0.40219.1\r\n"
					"\r\n`

const slnFileHeader2017 = `Microsoft Visual Studio Solution File, Format Version 12.00\r\n"
                    "# Visual Studio 15\r\n"
                    "VisualStudioVersion = 15.0.26403.0\r\n"
                    "MinimumVisualStudioVersion = 10.0.40219.1\r\n"
                    "\r\n`

const slnFileHeader2019 = `Microsoft Visual Studio Solution File, Format Version 12.00\r\n"
                    "# Visual Studio 16\r\n"
                    "VisualStudioVersion = 16.0.28803.352\r\n"
                    "MinimumVisualStudioVersion = 10.0.40219.1\r\n"
                    "\r\n`

const slnFileHeader2022 = `Microsoft Visual Studio Solution File, Format Version 12.00\r\n"
                    "# Visual Studio Version 17\r\n"
                    "VisualStudioVersion = 17.0.31314.256\r\n"
                    "MinimumVisualStudioVersion = 10.0.40219.1\r\n"
                    "\r\n`

// ----------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------

type EnumVisualStudio int

const (
	VisualStudio2015 EnumVisualStudio = iota
	VisualStudio2017
	VisualStudio2019
	VisualStudio2022
)

// ----------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------

var VisualStudioSlnHeaderMap = map[EnumVisualStudio]string{
	VisualStudio2015: slnFileHeader2015,
	VisualStudio2017: slnFileHeader2017,
	VisualStudio2019: slnFileHeader2019,
	VisualStudio2022: slnFileHeader2022,
}

// ----------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------

func NewVisualStudioConfig(version EnumVisualStudio) VisualStudioConfig {
	tools := "14.0"
	platform := "v140"
	target := "10.0"

	switch version {
	case VisualStudio2017:
		tools = "15.0"
		platform = "v141"
	case VisualStudio2019:
		tools = "16.0"
		platform = "v142"
	case VisualStudio2022:
		tools = "17.0"
		platform = "v143"
	}

	return VisualStudioConfig{
		Version:                      version,
		SlnHeader:                    VisualStudioSlnHeaderMap[version],
		ProjectTools:                 tools,
		PlatformToolset:              platform,
		WindowsTargetPlatformVersion: target,
	}
}
