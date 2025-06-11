package embedded

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jurgen-kluft/ccode/foundation"
)

func GenerateCppStructs(inputFile string, outputFile string) error {
	// Read the JSON file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Parse the JSON file
	r, err := cppCodeGeneratorFromJSON(data)
	if err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	// generate the C++ code
	err = r.generate(outputFile)
	if err != nil {
		return fmt.Errorf("error generating C++ code: %v", err)
	}

	return nil
}

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

type CppStructGenerator struct {
	Between      string        `json:"between,omitempty"`
	IndentType   IndentType    `json:"indent_type,omitempty"`
	IndentSize   int           `json:"indent_size,omitempty"`
	MemberPrefix string        `json:"member_prefix,omitempty"`
	Features     *FeatureFlags `json:"features,omitempty"`
	CppStruct    []CppStruct   `json:"structs"`
}

func NewCppStructGenerator() *CppStructGenerator {
	g := &CppStructGenerator{}
	g.Between = "== Generated Structs =="
	g.IndentType = IndentTypeSpace
	g.IndentSize = 4
	g.MemberPrefix = "m_"
	g.Features = NewFeatureFlags()
	g.CppStruct = make([]CppStruct, 0)
	return g
}

func (g *CppStructGenerator) SetDefaults() {
	if g.Between == "" {
		g.Between = "== Generated Structs =="
	}
	if g.IndentType == "" {
		g.IndentType = IndentTypeSpace
	}
	if g.IndentSize <= 0 {
		g.IndentSize = 4
	}
	if g.MemberPrefix == "" {
		g.MemberPrefix = "m_"
	}
	if g.Features == nil {
		g.Features = NewFeatureFlags()
	}
	if g.CppStruct == nil {
		g.CppStruct = make([]CppStruct, 0)
	}
}

func (g *CppStructGenerator) UnmarshalJSON(text []byte) error {
	type Alias CppStructGenerator
	aux := Alias{}
	if err := json.Unmarshal(text, &aux); err != nil {
		return err
	}
	*g = CppStructGenerator(aux)
	g.SetDefaults()
	return nil
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
