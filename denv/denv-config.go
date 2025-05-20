package denv

type DevConfigType int

const (
	DevConfigTypeStaticLibrary  DevConfigType = 1
	DevConfigTypeDynamicLibrary DevConfigType = 2
	DevConfigTypeExecutable     DevConfigType = 4
	DevConfigTypeDebug          DevConfigType = 8
	DevConfigTypeRelease        DevConfigType = 16
	DevConfigTypeFinal          DevConfigType = 64
	DevConfigTypeDevelopment    DevConfigType = 128
	DevConfigTypeTest           DevConfigType = 256
	DevConfigTypeProfile        DevConfigType = 512
	DevConfigTypeProduction     DevConfigType = 1024
	DevConfigTypeAll            DevConfigType = 0xFFFF
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
	Defines             *ValueSet
	LinkFlags           *ValueSet
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
		Defines:             NewValueSet(),
		LinkFlags:           NewValueSet(),
		Libs:                []*DevLib{},
	}

	return config
}

// -----------------------------------------------------------------------------------------------------
// DevConfigType to Tundra string
// -----------------------------------------------------------------------------------------------------

func (t DevConfigType) Tundra() string {
	switch t {
	case DevConfigTypeDebug:
		return "*-*-debug-*"
	case DevConfigTypeRelease:
		return "*-*-release-*"
	case DevConfigTypeFinal:
		return "*-*-final-*"
	case DevConfigTypeDebug | DevConfigTypeTest:
		return "*-*-debug-test"
	case DevConfigTypeRelease | DevConfigTypeTest:
		return "*-*-release-test"
	case DevConfigTypeFinal | DevConfigTypeTest:
		return "*-*-final-test"
	case DevConfigTypeDebug | DevConfigTypeProfile:
		return "*-*-debug-profile"
	case DevConfigTypeRelease | DevConfigTypeProfile:
		return "*-*-release-profile"
	case DevConfigTypeFinal | DevConfigTypeProfile:
		return "*-*-final-profile"
	}
	return "*-*-debug-*"
}
