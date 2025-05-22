package denv

type DevConfigType int

const (
	DevConfigTypeStaticLibrary  DevConfigType = 1
	DevConfigTypeDynamicLibrary DevConfigType = 2
	DevConfigTypeExecutable     DevConfigType = 4
    DevConfigTypeOutputMask = DevConfigTypeStaticLibrary | DevConfigTypeDynamicLibrary | DevConfigTypeExecutable
	DevConfigTypeDebug          DevConfigType = 8
	DevConfigTypeRelease        DevConfigType = 16
	DevConfigTypeFinal          DevConfigType = 64
    DevConfigTypeConfigMask = DevConfigTypeDebug | DevConfigTypeRelease | DevConfigTypeFinal
    DevConfigTypeDevelopment    DevConfigType = 128
	DevConfigTypeTest           DevConfigType = 256
	DevConfigTypeProfile        DevConfigType = 512
	DevConfigTypeProduction     DevConfigType = 1024
    DevConfigTypeVariantMask = DevConfigTypeDevelopment | DevConfigTypeTest | DevConfigTypeProfile | DevConfigTypeProduction
	DevConfigTypeAll            DevConfigType = DevConfigTypeOutputMask | DevConfigTypeConfigMask | DevConfigTypeVariantMask
)

func (t DevConfigType) Contains(o DevConfigType) bool {
	return t&o == o
}

func (t DevConfigType) GetProjectType() DevConfigType {
	return t & (DevConfigTypeStaticLibrary | DevConfigTypeDynamicLibrary | DevConfigTypeExecutable)
}

func (t DevConfigType) GetConfigType() DevConfigType {
	return t & (DevConfigTypeDebug | DevConfigTypeRelease | DevConfigTypeFinal)
}

func (t DevConfigType) GetConfigTypeVariant() DevConfigType {
	return t & (DevConfigTypeDevelopment | DevConfigTypeTest | DevConfigTypeProfile | DevConfigTypeProduction)
}

func (t DevConfigType) IsStaticLibrary() bool {
	return t&DevConfigTypeStaticLibrary != 0
}

func (t DevConfigType) IsDynamicLibrary() bool {
	return t&DevConfigTypeDynamicLibrary != 0
}

func (t DevConfigType) IsLibrary() bool {
	return t&(DevConfigTypeStaticLibrary|DevConfigTypeDynamicLibrary) != 0
}

func (t DevConfigType) IsApplication() bool {
	return t&DevConfigTypeExecutable != 0
}
func (t DevConfigType) IsExecutable() bool {
	return t&DevConfigTypeExecutable != 0
}

func (t DevConfigType) IsDebug() bool {
	return t&DevConfigTypeDebug != 0
}

func (t DevConfigType) IsRelease() bool {
	return t&DevConfigTypeRelease != 0
}

func (t DevConfigType) IsFinal() bool {
	return t&DevConfigTypeFinal != 0
}

func (t DevConfigType) IsDevelopment() bool {
	return t&DevConfigTypeDevelopment != 0
}

func (t DevConfigType) IsTest() bool {
	return t&DevConfigTypeTest != 0
}

func (t DevConfigType) IsProfile() bool {
	return t&DevConfigTypeProfile != 0
}

func (t DevConfigType) IsProduction() bool {
	return t&DevConfigTypeProduction != 0
}

func (t DevConfigType) ProjectString() string {
	switch t.GetProjectType() {
	case DevConfigTypeExecutable:
		return "c_exe"
	case DevConfigTypeDynamicLibrary:
		return "c_dll"
	case DevConfigTypeStaticLibrary:
		return "c_lib"
	}
	return "error"
}

func (t DevConfigType) ConfigString() string {
	str := "Debug"

	if t.IsDebug() {
		str = "Debug"
	} else if t.IsRelease() {
		str = "Release"
	} else if t.IsFinal() {
		str = "Final"
	}

	if t.IsTest() {
		str += "Test"
	} else if t.IsProfile() {
		str += "Profile"
	} else if t.IsProduction() {
		str += "Prod"
	} else if t.IsDevelopment() {
		str += "Dev"
	}

	return str
}

type DevConfig struct {
	ConfigType          DevConfigType
	LocalIncludeDirs    []string // Relative paths
	ExternalIncludeDirs []string // Absolute paths
	SourceDirs          []string
	Defines             *DevValueSet
	LinkFlags           *DevValueSet
	Libs                []*DevLib
}

func NewDevConfig(configType DevConfigType) *DevConfig {
	var config = &DevConfig{
		// Type:    "Static", // Static, Dynamic, Executable
		// Config:  "Debug",  // Debug, Release, Final
		// Build:   "Dev",    // Development(dev), Unittest(test), Profile(prof), Production(prod)
		ConfigType:          configType,
		LocalIncludeDirs:    []string{},
		ExternalIncludeDirs: []string{},
		SourceDirs:          []string{},
		Defines:             NewDevValueSet(),
		LinkFlags:           NewDevValueSet(),
		Libs:                []*DevLib{},
	}

	return config
}

// -----------------------------------------------------------------------------------------------------
// DevConfigType to Tundra string
// -----------------------------------------------------------------------------------------------------

func (t DevConfigType) Tundra() string {
    config := "*-*-"
	switch t & DevConfigTypeConfigMask {
	case DevConfigTypeDebug:
		config+="debug"
	case DevConfigTypeRelease:
		config+= "release"
	case DevConfigTypeFinal:
		config += "final"
	}

    switch t & DevConfigTypeVariantMask {
    case DevConfigTypeDevelopment: config += "-dev"
    case DevConfigTypeTest: config += "-test"
    case DevConfigTypeProfile: config += "-profile"
    case DevConfigTypeProduction: config += "-prod"
    default: config += "-*"
    }

	return config
}
