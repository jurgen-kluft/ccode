package foundation

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
	flags := NewBitFlags(&FeatureFlagDeclared)

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
	flags := NewBitFlags(&FeatureFlagDeclared)

	flags.Set(FeatureFlagMergeBoolIntoBitset)
	jsonData, err := flags.MarshalJSON()
	if err != nil {
		t.Errorf("Error marshalling to JSON: %v", err)
	}
	expectedJson := `"merge_bool_into_bitset"`
	if string(jsonData) != expectedJson {
		t.Errorf("Expected JSON to be %s, got %s", expectedJson, string(jsonData))
	}

	flags.Set(FeatureFlagOptimizeMemberLayout)
	jsonData, err = flags.MarshalJSON()
	if err != nil {
		t.Errorf("Error marshalling to JSON: %v", err)
	}

	expectedJson1 := `"merge_bool_into_bitset|optimize_member_layout"`
	expectedJson2 := `"optimize_member_layout|merge_bool_into_bitset"`
	if string(jsonData) != expectedJson1 && string(jsonData) != expectedJson2 {
		t.Errorf("Expected JSON to be %s or %s, got %s", expectedJson1, expectedJson2, string(jsonData))
	}
}

func TestBitFlagsUnmarshallJson(t *testing.T) {
	flags := NewBitFlags(&FeatureFlagDeclared)

	jsonData := `"merge_bool_into_bitset"`
	err := flags.UnmarshalJSON([]byte(jsonData))
	if err != nil {
		t.Errorf("Error unmarshalling from JSON: %v", err)
	}
	if !flags.Only(FeatureFlagMergeBoolIntoBitset) {
		t.Errorf("Expected value to be %d, got %d", FeatureFlagMergeBoolIntoBitset, flags.Bits())
	}

	jsonData = `"merge_bool_into_bitset|optimize_member_layout"`
	err = flags.UnmarshalJSON([]byte(jsonData))
	if err != nil {
		t.Errorf("Error unmarshalling from JSON: %v", err)
	}

	if !flags.Only(FeatureFlagMergeBoolIntoBitset | FeatureFlagOptimizeMemberLayout) {
		t.Errorf("Expected value to be %d, got %d", FeatureFlagMergeBoolIntoBitset|FeatureFlagOptimizeMemberLayout, flags.Bits())
	}
}
