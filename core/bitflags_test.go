package corepkg

import (
	"testing"
)

type FeatureFlag = Flag

const (
	FeatureFlagMergeBoolIntoBitset  FeatureFlag = 1 << iota // 1
	FeatureFlagOptimizeMemberLayout                         // 2
)

var FeatureFlagDeclared = map[string]Flag{
	"merge_bool_into_bitset": FeatureFlagMergeBoolIntoBitset,
	"optimize_member_layout": FeatureFlagOptimizeMemberLayout,
}

func TestBitFlags(t *testing.T) {
	flags := NewBitFlags(0, &FeatureFlagDeclared)

	if flags.Bits() != 0 {
		t.Errorf("Expected initial value to be 0, got %d", flags.Bits())
	}

	flags.Set(FeatureFlagMergeBoolIntoBitset)
	if !flags.Only(FeatureFlagMergeBoolIntoBitset) {
		t.Errorf("Expected initial value to be %d, got %d", FeatureFlagMergeBoolIntoBitset, flags.Bits())
	}

	flags.Set(FeatureFlagOptimizeMemberLayout)
	if !flags.Only(FeatureFlagMergeBoolIntoBitset | FeatureFlagOptimizeMemberLayout) {
		t.Errorf("Expected value to be %d, got %d", FeatureFlagMergeBoolIntoBitset|FeatureFlagOptimizeMemberLayout, flags.Bits())
	}

	flags.SetAll()

	if !flags.Only(FeatureFlagMergeBoolIntoBitset | FeatureFlagOptimizeMemberLayout) {
		t.Errorf("Expected all bits to be set, got %d", flags.Bits())
	}
}

func TestBitFlagsMarshallJson(t *testing.T) {
	flags := NewBitFlags(0, &FeatureFlagDeclared)

	encoder := NewJsonEncoder("    ")
	encoder.Begin()
	flags.Set(FeatureFlagMergeBoolIntoBitset)
	err := flags.EncodeJSON(encoder)
	if err != nil {
		t.Errorf("Error marshalling to JSON: %v", err)
	}
	resultingJson := encoder.End()
	expectedJson := `"merge_bool_into_bitset"`
	if resultingJson != expectedJson {
		t.Errorf("Expected JSON to be %s, got %s", expectedJson, resultingJson)
	}

	encoder.Begin()
	flags.Set(FeatureFlagOptimizeMemberLayout)
	err = flags.EncodeJSON(encoder)
	if err != nil {
		t.Errorf("Error marshalling to JSON: %v", err)
	}
	resultingJson = encoder.End()

	expectedJson1 := `"merge_bool_into_bitset|optimize_member_layout"`
	expectedJson2 := `"optimize_member_layout|merge_bool_into_bitset"`
	if resultingJson != expectedJson1 && resultingJson != expectedJson2 {
		t.Errorf("Expected JSON to be %s or %s, got %s", expectedJson1, expectedJson2, resultingJson)
	}
}

func TestBitFlagsUnmarshallJson(t *testing.T) {
	flags := NewBitFlags(0, &FeatureFlagDeclared)

	decoder := NewJsonDecoder()

	jsonData := `{ "flags": "merge_bool_into_bitset" }`
	if !decoder.Begin(jsonData) {
		t.Errorf("Failed to start JSON decoder on: %s", jsonData)
	}

	fields := map[string]JsonDecode{
		"flags": func(decoder *JsonDecoder) {
			flags.DecodeJSON(decoder)
		},
	}

	if err := decoder.Decode(fields); err != nil {
		t.Errorf("Failed to decode JSON: %v", decoder.Error)
	}

	if !flags.Only(FeatureFlagMergeBoolIntoBitset) {
		t.Errorf("Expected value to be %d, got %d", FeatureFlagMergeBoolIntoBitset, flags.Bits())
	}

	jsonData = `{ "flags": "merge_bool_into_bitset|optimize_member_layout" }`
	decoder.Begin(jsonData)
	if err := decoder.Decode(fields); err != nil {
		t.Errorf("Error unmarshalling from JSON: %v", err)
	}
	if !flags.Only(FeatureFlagMergeBoolIntoBitset | FeatureFlagOptimizeMemberLayout) {
		t.Errorf("Expected value to be %d, got %d", FeatureFlagMergeBoolIntoBitset|FeatureFlagOptimizeMemberLayout, flags.Bits())
	}
}
