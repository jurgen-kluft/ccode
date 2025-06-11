package embedded

import (
	"encoding/json"

	"github.com/jurgen-kluft/ccode/foundation"
)

// Feature is a BitFlags representing various features that can be enabled/disabled
// It is similar to a "[Flags] enum{}" in C# or a bitmask in C/C++.

type FeatureFlags struct {
	foundation.BitFlags
}
type FeatureFlag = foundation.Flag

const (
	FeatureFlagNone              FeatureFlag = 0
	FeatureFlagBoolsAsBitset     FeatureFlag = 1 << iota // 1
	FeatureFlagOptimizeLayout                            // 2
	FeatureFlagInitializeMembers                         // 4
	FeatureFlagGenerateClass                             // 8
)

var FeatureFlagDecl = map[string]FeatureFlag{
	"bools_as_bitset":    FeatureFlagBoolsAsBitset,
	"optimize_layout":    FeatureFlagOptimizeLayout,
	"initialize_members": FeatureFlagInitializeMembers,
	"generate_class":     FeatureFlagGenerateClass,
}

func NewFeatureFlags() *FeatureFlags {
	flags := &FeatureFlags{
		BitFlags: foundation.NewBitFlags(FeatureFlagBoolsAsBitset, &FeatureFlagDecl),
	}
	return flags
}

func (f *FeatureFlags) UnmarshalJSON(data []byte) error {
	if value, err := foundation.UnmarshalBitFlagsFromJSON(data, FeatureFlagDecl); err != nil {
		return err
	} else {
		f.BitFlags = foundation.NewBitFlags(value, &FeatureFlagDecl)
		return nil
	}
}

type CppStructMember struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Initializer string `json:"initializer,omitempty"`
}

type CppStruct struct {
	Name         string            `json:"name"`
	Type         string            `json:"type,omitempty"`
	MemberPrefix string            `json:"prefix,omitempty"`
	Features     *FeatureFlags     `json:"features,omitempty"`
	Members      []CppStructMember `json:"members"`
}

func NewCppStruct() *CppStruct {
	g := &CppStruct{}
	g.Name = ""
	g.Type = "struct"
	g.MemberPrefix = "m_"
	g.Features = NewFeatureFlags()
	g.Members = make([]CppStructMember, 0)
	return g
}

func (g *CppStruct) SetDefaults() {
	if g.Type == "" {
		g.Type = "struct"
	}
	if g.MemberPrefix == "" {
		g.MemberPrefix = "m_"
	}
	if g.Features == nil {
		g.Features = NewFeatureFlags()
	}
	if g.Members == nil {
		g.Members = make([]CppStructMember, 0)
	}
}

func (g *CppStruct) UnmarshalJSON(text []byte) error {
	type Alias CppStruct
	aux := Alias{}
	if err := json.Unmarshal(text, &aux); err != nil {
		return err
	}
	*g = CppStruct(aux)
	g.SetDefaults()
	return nil
}
