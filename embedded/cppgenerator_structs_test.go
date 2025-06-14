package embedded

import (
	"strings"
	"testing"

	"github.com/jurgen-kluft/ccode/foundation"
)

/*
Example configuration for generating C++ structs:

JSON
{
	"features": "bools_as_bitset|optimize_layout|initialize_members|generate_class",
    "structs": [
        {
            "name": "Rect2DInt",
			"features": "bools_as_bitset|optimize_layout|initialize_members|generate_class",
            "members": [
                {
                    "name": "x",
                    "type": "i16",
					"initializer": "0"
                },
                {
                    "name": "y",
                    "type": "i16",
					"initializer": "0"
                },
                {
                    "name": "width",
                    "type": "u16"
                },
                {
                    "name": "height",
                    "type": "u16"
                }
            ]
        },
        {
            "name": "Viewport",
			"features": "bools_as_bitset|optimize_layout|initialize_members|generate_class",
            "members": [
                {
                    "name": "rect",
                    "type": "Rect2DInt",
                    "initializer": ""
                },
                {
                    "name": "min_depth",
                    "type": "float",
                    "initializer": "0.0f"
                },
                {
                    "name": "max_depth",
                    "type": "float",
                    "initializer": "0.0f"
                }
            ]
        }
    ]
}
*/

func TestCppGeneratorStructs(t *testing.T) {
	jsonData := `{
		"between": "// == Generated Structs ==",
		"indentType": "space",
		"indentSize": 4,
		"memberPrefix": "m_",
		"features": "bools_as_bitset|initialize_members",
		"structs": [
			{
				"name": "Rect2DInt",
    		    "features": "generate_class",
				"members": [
					{"name": "x", "type": "i16", "initializer": "0"},
					{"name": "y", "type": "i16", "initializer": "0"},
					{"name": "width", "type": "u16"},
					{"name": "height", "type": "u16"}
				]
			},
			{
				"name": "Viewport",
    		    "features": "optimize_layout",
				"members": [
					{"name": "rect", "type": "Rect2DInt"},
					{"name": "min_depth", "type": "float", "initializer": "0.0f"},
					{"name": "max_depth", "type": "float", "initializer": "0.0f"}
				]
			},
			{
				"name": "LotsOfBools",
				"members": [
					{
						"name": "nice",
						"type": "bool",
						"init": "false"
					},
					{
						"name": "cool",
						"type": "bool",
						"init": "true"
					},
					{
						"name": "depth_clamp",
						"type": "bool",
						"init": "true"
					},
					{
						"name": "stencil_test",
						"type": "bool",
						"init": "false"
					},
					{
						"name": "depth_test",
						"type": "bool",
						"init": "true"
					}
				]
			}
		]
	}`

	decoder := foundation.NewJsonDecoder()
	if !decoder.Begin(jsonData) {
		t.Fatal("Failed to begin decoding JSON")
	}

	r := newCppCodeGenerator()
	if err := r.decodeJSON(decoder); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(r.cppStruct) == 0 {
		t.Fatal("No structs found in the generator")
	}

	linesOfCode := r.generateCppCode()
	if len(linesOfCode) == 0 {
		t.Fatal("No C++ code generated")
	}

	testCppFile := []string{
		`#include <cstdint>`,
		`#include <cstddef>`,
		``,
		`namespace embedded`,
		`{`,
		`	class Test;`,
		``,
		`	// == Generated Structs ==`,
		`	// == Generated Structs ==`,
		``,
		`	class Test`,
		`	{`,
		`	public:`,
		`		bool test() { return m_test; }`,
		`	private:`,
		`		bool m_test = true;`,
		`	}`,
		``,
		`}`,
	}

	updatedTestCppFile, err := r.insertGeneratedCode(testCppFile, linesOfCode)
	if err != nil {
		t.Fatalf("Failed to insert generated code: %v", err)
	}

	if len(updatedTestCppFile) != (len(testCppFile) + len(linesOfCode)) {
		t.Fatalf("Inserted code is longer than generated code: %d vs %d", len(updatedTestCppFile), len(linesOfCode))
	}

	generated := false
	codeLineIndex := 0
	inputLineIndex := 0
	for _, line := range updatedTestCppFile {
		if generated {
			if strings.Contains(line, r.between) {
				inputLineIndex++
				generated = false
				continue
			}

			if codeLineIndex >= len(linesOfCode) {
				t.Errorf("More lines in generated code than expected: %d vs %d", codeLineIndex, len(linesOfCode))
				break
			}

			if !strings.HasSuffix(line, linesOfCode[codeLineIndex]) {
				t.Errorf("Line %d mismatch: expected '%s', got '%s'", codeLineIndex+1, linesOfCode[codeLineIndex], line)
			}
			codeLineIndex++
			continue
		} else if strings.Contains(line, r.between) {
			inputLineIndex++
			generated = true
			continue
		}

		if testCppFile[inputLineIndex] != line {
			t.Errorf("Line %d mismatch: expected '%s', got '%s'", inputLineIndex+1, testCppFile[inputLineIndex], line)
		}
		inputLineIndex++
	}
}
