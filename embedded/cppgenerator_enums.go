package embedded

import (
	"fmt"
	"strings"
)

/* Example JSON
{
    "between": "=== GPU Enumerations ===",
    "enums": [
        {
            "name": "BlendOperation",
            "type": "byte",
            "namespace": true,
            "members": [
                "Add",
                "Subtract",
                "RevSubtract",
                "Min",
                "Max"
            ],
            "generate": [
                "ToString",
                "Mask",
                "MaskToEnum",
                "EnumToMask"
            ]
        }
    ]
}
*/

/* Example C++
namespace BlendOperation {
    enum Enum {
        Add, Subtract, RevSubtract, Min, Max, Count
    };

    enum Mask {
        Add_mask = 1 << 0, Subtract_mask = 1 << 1, RevSubtract_mask = 1 << 2, Min_mask = 1 << 3, Max_mask = 1 << 4, Count_mask = 1 << 5
    };

    static Enum MaskToEnum( u32 mask ) {
        s8 const bit = Math::Log2( mask );
        return ( bit >= 0 && bit < Count ) ? (Enum)bit : Enum::Count;
    }

    static Mask EnumToMask( Enum e ) {
        return ( e < Enum::Count ) ? (Mask)( 1 << (u32)e ) : Mask::Count_mask;
    }

    static const char* s_value_names[] = {
        "Add", "Subtract", "RevSubtract", "Min", "Max", "Count"
    };

    static const char* ToString( Enum e ) {
        return ((u32)e < Enum::Count ? s_value_names[(int)e] : "unsupported" );
    }
} // namespace BlendOperation
*/

type enumGenerate string

const (
	enumGenerateToString   enumGenerate = "ToString"
	enumGenerateMask       enumGenerate = "Mask"
	enumGenerateEnumToMask enumGenerate = "EnumToMask"
)

type cppEnum struct {
	name      string
	namespace bool
	enumType  string
	members   []string
	generate  []enumGenerate
}

func newCppEnum() cppEnum {
	return cppEnum{
		name:      "",
		namespace: false,
		enumType:  "int",
		members:   []string{},
		generate:  []enumGenerate{},
	}
}

func (e cppEnum) whichCodeToGenerate(g enumGenerate) bool {
	for _, v := range e.generate {
		if strings.EqualFold(string(v), string(g)) {
			return true
		}
	}
	return false
}

func (r *cppCodeGenerator) generateEnum(e cppEnum) []string {
	cppCode := newCppCode(r.indentType, r.indentSize)

	if e.namespace {
		cppCode.addNamespaceOpen(e.name)
	}

	cppCode.addEnumEnum(e)

	if e.whichCodeToGenerate(enumGenerateMask) {
		cppCode.addEnumMask(e)

		if e.whichCodeToGenerate(enumGenerateEnumToMask) {
			cppCode.addEnumToMaskFunction(e)
		}
	}

	if e.whichCodeToGenerate(enumGenerateToString) {
		cppCode.addEnumToString(e)
	}

	if e.namespace {
		cppCode.addNamespaceClose(e.name)
	}

	return cppCode.lines
}

// =============================================================================
// cppCode enum specific functions
// =============================================================================

func (cpp *cppCode) addEnumEnum(enum cppEnum) {
	if enum.namespace {
		cpp.appendTextLine("enum Enum {")
	} else {
		cpp.appendTextLine("enum " + enum.name + " {")
	}
	cpp.increaseIndentation()
	for _, v := range enum.members {
		cpp.appendTextLine(v + ",")
	}
	cpp.appendTextLine("Count")

	cpp.decreaseIndentation()
	cpp.appendTextLine("};")
}

func (cpp *cppCode) addEnumToString(enum cppEnum) {
	cpp.appendTextLine("")
	cpp.appendTextLine("static const char* s_value_" + strings.ToLower(enum.name) + "_names[] = {")
	cpp.increaseIndentation()
	for _, v := range enum.members {
		cpp.appendTextLine("\"" + v + "\",")
	}
	cpp.decreaseIndentation()
	cpp.appendTextLine("};")

	cpp.appendTextLine("")
	if enum.namespace {
		cpp.appendTextLine("static const char* ToString(Enum e) {")
		cpp.increaseIndentation()
		cpp.appendTextLine("return (e < Enum::Count ? s_value_names[(int)e] : \"unsupported\" );")
		cpp.decreaseIndentation()
		cpp.appendTextLine("}")
	} else {
		cpp.appendTextLine("static const char* ToString(" + enum.name + " e ) {")
		cpp.increaseIndentation()
		cpp.appendTextLine("return (e < " + enum.name + "::Count ? s_value_" + strings.ToLower(enum.name) + "_names[(int)e] : \"unsupported\" );")
		cpp.decreaseIndentation()
		cpp.appendTextLine("}")
	}
}

func (cpp *cppCode) addEnumMask(enum cppEnum) {
	cpp.appendTextLine("")
	if enum.namespace {
		cpp.appendTextLine("enum Mask {")
	} else {
		cpp.appendTextLine("enum " + enum.name + "_mask {")
	}
	cpp.increaseIndentation()
	for i, v := range enum.members {
		cpp.appendTextLine(v + "_mask = " + v + " << " + fmt.Sprint(i) + ",")
	}
	cpp.decreaseIndentation()
	cpp.appendTextLine("};")
}

func (cpp *cppCode) addEnumToMaskFunction(enum cppEnum) {
	cpp.appendTextLine("")
	if enum.namespace {
		cpp.appendTextLine("static Mask EnumToMask(Enum e) {")
		cpp.increaseIndentation()
		cpp.appendTextLine("return (Mask)(1 << (int)e);")
		cpp.decreaseIndentation()
		cpp.appendTextLine("}")
	} else {
		cpp.appendTextLine("static " + enum.name + "_mask EnumToMask(" + enum.name + " e) {")
		cpp.increaseIndentation()
		cpp.appendTextLine("return (" + enum.name + "_mask)(1 << (int)e);")
		cpp.decreaseIndentation()
		cpp.appendTextLine("}")
	}
}
