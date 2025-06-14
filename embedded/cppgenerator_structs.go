package embedded

import (
	"strconv"

	"github.com/jurgen-kluft/ccode/foundation"
)

type featureFlags struct {
	foundation.BitFlags
}
type featureFlag = foundation.Flag

const (
	featureFlagNone              featureFlag = 0
	featureFlagBoolsAsBitset     featureFlag = 1 << iota // 1
	featureFlagOptimizeLayout                            // 2
	featureFlagInitializeMembers                         // 4
	featureFlagGenerateClass                             // 8
)

var featureFlagDecl = map[string]featureFlag{
	"bools_as_bitset":    featureFlagBoolsAsBitset,
	"optimize_layout":    featureFlagOptimizeLayout,
	"initialize_members": featureFlagInitializeMembers,
	"generate_class":     featureFlagGenerateClass,
}

func newFeatureFlags() *featureFlags {
	flags := &featureFlags{
		BitFlags: foundation.NewBitFlags(featureFlagBoolsAsBitset, &featureFlagDecl),
	}
	return flags
}

func (f *featureFlags) decodeJSON(decoder *foundation.JsonDecoder) error {
	return f.BitFlags.DecodeJSON(decoder)
}

var defaultMemberAlignmentMap = map[string]int{
	"bool":      1, // 1 byte
	"int":       4, // 4 bytes
	"float":     4, // 4 bytes
	"double":    8, // 8 bytes
	"char":      1, // 1 byte
	"short":     2, // 2 bytes
	"long":      4, // 4 bytes
	"long long": 8, // 8 bytes
	"uint8_t":   1, // 1 byte
	"uint16_t":  2, // 2 bytes
	"uint32_t":  4, // 4 bytes
	"uint64_t":  8, // 8 bytes
	"int8_t":    1, // 1 byte
	"int16_t":   2, // 2 bytes
	"int32_t":   4, // 4 bytes
	"int64_t":   8, // 8 bytes
	"i8":        1, // 1 byte
	"i16":       2, // 2 bytes
	"i32":       4, // 4 bytes
	"i64":       8, // 8 bytes
	"u8":        1, // 1 byte
	"u16":       2, // 2 bytes
	"u32":       4, // 4 bytes
	"u64":       8, // 8 bytes
	"s8":        1, // 1 byte
	"s16":       2, // 2 bytes
	"s32":       4, // 4 bytes
	"s64":       8, // 8 bytes
	"size_t":    8, // 8 bytes on 64-bit systems
}

var defaultMemberInitializerMap = map[string]string{
	"bool":      "false", // default value for bool
	"int":       "0",     // default value for int
	"float":     "0.0f",  // default value for float
	"double":    "0.0",   // default value for double
	"char":      "0",     // default value for char (null character)
	"short":     "0",     // default value for short
	"long":      "0",     // default value for long
	"long long": "0",     // default value for long long
	"uint8_t":   "0",     // default value for uint8_t
	"uint16_t":  "0",     // default value for uint16_t
	"uint32_t":  "0",     // default value for uint32_t
	"uint64_t":  "0",     // default value for uint64_t
	"int8_t":    "0",     // default value for int8_t
	"int16_t":   "0",     // default value for int16_t
	"int32_t":   "0",     // default value for int32_t
	"int64_t":   "0",     // default value for int64_t
	"i8":        "0",     // default value for i8
	"i16":       "0",     // default value for i16
	"i32":       "0",     // default value for i32
	"i64":       "0",     // default value for i64
	"u8":        "0",     // default value for u8
	"u16":       "0",     // default value for u16
	"u32":       "0",     // default value for u32
	"u64":       "0",     // default value for u64
	"s8":        "0",     // default value for s8
	"s16":       "0",     // default value for s16
	"s32":       "0",     // default value for s32
	"s64":       "0",     // default value for s64
	"size_t":    "0",     // default value for size_t
}

type cppStructMember struct {
	name        string // name of the member, e.g. "width", "height", etc.
	memberType  string // type of the member, e.g. "int", "float", "bool", etc.
	bits        int    // number of bits, used for sorting members
	initializer string // initializer for the member, e.g. "", "0", "1.0f", "false", etc.
}

func newCppStructMember() cppStructMember {
	return cppStructMember{
		name:        "",
		memberType:  "",
		initializer: "",
	}
}

type cppStruct struct {
	name         string
	memberPrefix string
	features     *featureFlags
	members      []cppStructMember
}

func newCppStruct() cppStruct {
	return cppStruct{
		name:         "",
		memberPrefix: "m_",
		features:     newFeatureFlags(),
		members:      make([]cppStructMember, 0),
	}
}

func (r *cppCodeGenerator) generateStruct(initialIndentation string, cs *cppStruct) []string {
	cppCode := newCppCode(initialIndentation, r.indentType, r.indentSize)

	// Generate the class or struct declarations
	if cs.features.Has(featureFlagGenerateClass) {
		cppCode.appendTextLine("class ", cs.name)
	} else {
		cppCode.appendTextLine("struct ", cs.name)
	}
	cppCode.appendTextLine("{")
	cppCode.increaseIndentation()

	boolset := cs.features.Has(featureFlagBoolsAsBitset)
	bitsetSizePo2 := 3 // 2^3 = 8 bits
	switch r.bitsetSize {
	case 16:
		bitsetSizePo2 = 4 // 2^4 = 16 bits
	case 32:
		bitsetSizePo2 = 5 // 2^5 = 32 bits
	case 64:
		bitsetSizePo2 = 6 // 2^6 = 64 bits
	}
	bitsetSize := 1 << bitsetSizePo2

	bitsetMembers := make([]string, 0, 32)
	bitsetMemberInits := make([]uint64, 0, 32)
	if boolset {
		// Find all the bool members and convert them to a bitset
		bitsetCount := 0
		for _, member := range cs.members {
			if member.memberType == "bool" && (bitsetCount&(bitsetSize-1)) == 0 {
				bitsetIdx := bitsetCount >> bitsetSizePo2 // bitset index (0 for first 8 bools, 1 for next 8, etc.)
				bitsetMemberName := cs.memberPrefix + "bitset" + strconv.Itoa(bitsetIdx)
				bitsetMembers = append(bitsetMembers, bitsetMemberName)
				bitsetMemberInits = append(bitsetMemberInits, 0) // Initialize to 0
				bitsetCount++
			}
		}

		// < 8 bools can be stored in a single byte, for each one we emit a get/set function
		bitsetCount = 0
		for _, member := range cs.members {
			if member.memberType == "bool" {
				bitset := bitsetCount >> bitsetSizePo2   // bitset index (0 for first 8 bools, 1 for next 8, etc.)
				bitidx := bitsetCount & (bitsetSize - 1) // bit index within the byte
				shift := "(1 << " + strconv.Itoa(bitidx) + ")"
				bitsetMemberName := bitsetMembers[bitset]
				if member.initializer == "true" {
					bitsetMemberInits[bitset] |= 1 << bitidx // Set the bit in the bitset initializer
				}
				cppCode.appendTextLine("inline bool get_", member.name, "() const { return (", bitsetMemberName, " & ", shift, ") != 0; }")
				cppCode.appendTextLine("inline void set_", member.name, "(bool value) { ", bitsetMemberName, " = (", bitsetMemberName, " & ~", shift, ") | (value ? ", shift, " : 0); }")
				bitsetCount++
			}
		}
	}

	initializer := r.features.Has(featureFlagInitializeMembers) || cs.features.Has(featureFlagInitializeMembers)

	for mi, member := range cs.members {
		if initializer && member.initializer == "" {
			if defaultInitializer, ok := defaultMemberInitializerMap[member.memberType]; ok {
				cs.members[mi].initializer = defaultInitializer
			}
		}
	}

	// Add the member variables
	for _, member := range cs.members {
		if boolset && member.memberType == "bool" {
			continue
		}
		if len(member.initializer) > 0 {
			cppCode.appendTextLine(member.memberType, " ", cs.memberPrefix, member.name, " = ", member.initializer, ";")
		} else {
			cppCode.appendTextLine(member.memberType, " ", cs.memberPrefix, member.name, ";")
		}
	}

	// Write the bitset members if bools are stored as bitset
	for bsmi, bitsetMember := range bitsetMembers {
		initializeValue := bitsetMemberInits[bsmi]
		initializeStr := "0b" + strconv.FormatUint(initializeValue, 2)
		cppCode.appendTextLine(r.bitsetType, " ", bitsetMember, " = ", initializeStr, ";")
	}

	cppCode.decreaseIndentation()
	cppCode.appendTextLine("};")
	cppCode.appendTextLine("")

	return cppCode.lines
}
