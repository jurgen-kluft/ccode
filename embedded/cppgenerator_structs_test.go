package embedded

import "testing"

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
		"indent_type": "space",
		"indent_size": 4,
		"member_prefix": "m_",
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
			}
		]
	}`

	generator, err := newCppCodeGeneratorFromJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(generator.cppStruct) == 0 {
		t.Fatal("No structs found in the generator")
	}

}
