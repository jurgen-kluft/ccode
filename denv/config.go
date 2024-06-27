package denv

type Config struct {
	Type         string
	Config       string
	Build        string
	Defines      []string
	LibraryFiles []string
}

func NewDebugConfig() *Config {
	var DebugConfig = &Config{
		Type:         "Static", // Static, Dynamic, Executable
		Config:       "Debug",  // Debug, Release, Final
		Build:        "Dev",    // Development, Unittest, Profile, Shipping
		Defines:      []string{},
		LibraryFiles: []string{},
	}

	return DebugConfig
}

func NewReleaseConfig() *Config {
	var ReleaseConfig = &Config{
		Type:         "Static",  // Static, Dynamic, Executable
		Config:       "Release", // Debug, Release, Final
		Build:        "Dev",     // Development, Unittest, Profile, Shipping
		Defines:      []string{},
		LibraryFiles: []string{},
	}
	return ReleaseConfig
}
