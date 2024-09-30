package denv

type ConfigType int

const (
	ConfigTypeStaticLibrary  ConfigType = 1
	ConfigTypeDynamicLibrary ConfigType = 2
	ConfigTypeExecutable     ConfigType = 4
	ConfigTypeDebug          ConfigType = 8
	ConfigTypeRelease        ConfigType = 16
	ConfigTypeLibrary        ConfigType = 32
	ConfigTypeFinal          ConfigType = 64
	ConfigTypeDevelopment    ConfigType = 128
	ConfigTypeUnittest       ConfigType = 256
	ConfigTypeProfile        ConfigType = 512
	ConfigTypeProduction     ConfigType = 1024
	ConfigTypeAll            ConfigType = 0xFFFF
)

func (t ConfigType) Contains(o ConfigType) bool {
	return t&o == o
}

func (t ConfigType) IsStaticLibrary() bool {
	return t&ConfigTypeStaticLibrary != 0
}

func (t ConfigType) IsDynamicLibrary() bool {
	return t&ConfigTypeDynamicLibrary != 0
}

func (t ConfigType) IsLibrary() bool {
	return t&ConfigTypeLibrary != 0
}

func (t ConfigType) IsExecutable() bool {
	return t&ConfigTypeExecutable != 0
}

func (t ConfigType) IsDebug() bool {
	return t&ConfigTypeDebug != 0
}

func (t ConfigType) IsRelease() bool {
	return t&ConfigTypeRelease != 0
}

func (t ConfigType) IsFinal() bool {
	return t&ConfigTypeFinal != 0
}

func (t ConfigType) IsDevelopment() bool {
	return t&ConfigTypeDevelopment != 0
}

func (t ConfigType) IsUnittest() bool {
	return t&ConfigTypeUnittest != 0
}

func (t ConfigType) IsProfile() bool {
	return t&ConfigTypeProfile != 0
}

type Config struct {
	ConfigType  ConfigType
	IncludeDirs []string
	SourceDirs  []string
	Defines     *ValueSet
	LinkFlags   *ValueSet
	Libs        []*Lib
}

func NewConfig(configType ConfigType) *Config {
	var config = &Config{
		// Type:    "Static", // Static, Dynamic, Executable
		// Config:  "Debug",  // Debug, Release, Final
		// Build:   "Dev",    // Development(dev), Unittest(test), Profile(prof), Production(prod)
		ConfigType: configType,
		Defines:    NewValueSet(),
		LinkFlags:  NewValueSet(),
		Libs:       []*Lib{},
	}

	return config
}
