package axe

type RuntimeLibraryType int

const (
	RuntimeLibraryNone RuntimeLibraryType = iota
	RuntimeLibraryMultiThreaded
	RuntimeLibraryMultiThreadedDLL
)

type VisualStudioConfig struct {
	Version                      EnumVisualStudio
	SlnHeader                    []string
	ProjectTools                 string // e.g. 14.0
	PlatformToolset              string // e.g. v140
	WindowsTargetPlatformVersion string // e.g. 10.0
	RuntimeLibrary               RuntimeLibraryType
}

func (r RuntimeLibraryType) String(debug bool) string {
	if debug {
		switch r {
		case RuntimeLibraryMultiThreaded:
			return "MultiThreadedDebug"
		case RuntimeLibraryMultiThreadedDLL:
			return "MultiThreadedDebugDLL"
		}
	} else {
		switch r {
		case RuntimeLibraryMultiThreaded:
			return "MultiThreadedDebug"
		case RuntimeLibraryMultiThreadedDLL:
			return "MultiThreadedDebugDLL"
		}
	}
	return ""
}

// ----------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------
var slnFileHeader2015 = []string{
	"Microsoft Visual Studio Solution File, Format Version 12.00",
	"# Visual Studio 14",
	"VisualStudioVersion = 14.0.25420.1",
	"MinimumVisualStudioVersion = 10.0.40219.1",
	"",
}

var slnFileHeader2017 = []string{
	"Microsoft Visual Studio Solution File, Format Version 12.00",
	"# Visual Studio 15",
	"VisualStudioVersion = 15.0.26403.0",
	"MinimumVisualStudioVersion = 10.0.40219.1",
	"",
}

var slnFileHeader2019 = []string{
	"Microsoft Visual Studio Solution File, Format Version 12.00",
	"# Visual Studio 16",
	"VisualStudioVersion = 16.0.28803.352",
	"MinimumVisualStudioVersion = 10.0.40219.1",
	"",
}

var slnFileHeader2022 = []string{
	"Microsoft Visual Studio Solution File, Format Version 12.00",
	"# Visual Studio Version 17",
	"VisualStudioVersion = 17.0.31314.256",
	"MinimumVisualStudioVersion = 10.0.40219.1",
	"",
}

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

var VisualStudioSlnHeaderMap = map[EnumVisualStudio][]string{
	VisualStudio2015: slnFileHeader2015,
	VisualStudio2017: slnFileHeader2017,
	VisualStudio2019: slnFileHeader2019,
	VisualStudio2022: slnFileHeader2022,
}

// ----------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------

func NewVisualStudioConfig(version EnumVisualStudio) *VisualStudioConfig {
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

	return &VisualStudioConfig{
		Version:                      version,
		SlnHeader:                    VisualStudioSlnHeaderMap[version],
		ProjectTools:                 tools,
		PlatformToolset:              platform,
		WindowsTargetPlatformVersion: target,
		RuntimeLibrary:               RuntimeLibraryMultiThreaded, // Static or Dynamic Linking of the Runtime
	}
}
