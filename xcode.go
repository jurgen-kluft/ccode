package xcode

type Config struct {
	name      string   // Debug, Release
	os        string   // Windows, Darwin
	arch      string   // x86, x64, ARM
	defines   []string //
	includes  []string
	libraries []string
	linking   []string
}

// Version (based on semver)
type Version struct {
	Major byte
	Minor byte
	Patch byte
}

type Package struct {
	name     string
	guid     string
	author   string
	version  Version
	os       string // Windows, Darwin
	arch     string // x86, x64, ARM
	language string
	configs  []Config
}

type Dependency struct {
	packageName string
	version     Version
}

func StaticLibrary() {

}

func Unittest() {

}
