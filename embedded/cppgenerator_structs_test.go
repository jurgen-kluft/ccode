package embedded

import (
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

/*
Example C++ structs generated from the above JSON:

// == Generated Structs ==
//
// struct Rect2DInt {
//     i16 mX;
//     i16 mY;
//     u16 mWidth;
//     u16 mHeight;
// };
// struct Viewport {
//     Rect2DInt mRect;
//     float mMinDepth = 0.0f;
//     float mMaxDepth = 0.0f;
// };

*/

func TestCppGeneratorStructs(t *testing.T) {
	jsonData := `{
		"between": "== Generated Structs ==",
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

	linesOfCode := r.generateCppCode("")
	if len(linesOfCode) == 0 {
		t.Fatal("No C++ code generated")
	}

}
