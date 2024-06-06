package embedded

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

/* Example JSON
{
    "between": "=> Gpu Enumerations - GENERATED <=",
    "enums": [
        {
            "name": "BlendOperation",
            "type": "byte",
            "namespace": true,
            "enums": [
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

// This is the public function that will generate the C++ code
func CppGenerateCode(inputFile string, outputFile string) error {
	// Read the JSON file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	// Parse the JSON file
	r, err := cppCodeGeneratorFromJSON(data)
	if err != nil {
		return err
	}

	// generate the C++ code
	err = r.generate(outputFile)
	if err != nil {
		return err
	}

	return nil
}

// C++ code generator
func cppCodeGeneratorFromJSON(data []byte) (*CppCodeGenerator, error) {
	r := &CppCodeGenerator{}
	err := json.Unmarshal(data, r)
	return r, err
}

// This will generate the C++ code to the output file
func (r *CppCodeGenerator) generate(outputFile string) error {

	// We need to open the file and find the between string, and replace everything between the begin and end line

	// Open the file
	// Read the full file into memory as and array of lines
	// Find the first between string, that is the begin line
	// Find the second between string, that is the end line
	// Replace everything between the begin and end line with the generated code
	// Write the file back to disk

	f, err := os.Open(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read the full file into memory as and array of lines
	lines := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Find the first between string, that is the begin line
	// Do keep track of what the indentation is for the begin line
	// Find the second between string, that is the end line
	beginLine := -1
	endLine := -1
	indentation := ""
	for i, line := range lines {
		// Look for the between string using 'contains'
		if strings.Contains(line, r.Between) {
			if beginLine == -1 {
				beginLine = i
				indentation = ""
				for _, c := range line {
					if c == ' ' {
						indentation += " "
					} else if c == '\t' {
						indentation += "\t"
					} else {
						break
					}
				}
				break
			}

			if endLine == -1 {
				endLine = i
				break
			}
		}
	}

	// Replace everything between the begin and end line with the generated code
	// Write the file back to disk
	if beginLine != -1 && endLine != -1 {

		// Open the file for writing
		f, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		defer f.Close()

		// Write the lines up to the begin line with the correct indentation
		for i := 0; i < beginLine; i++ {
			f.WriteString(lines[i] + "\n")
		}

		// =====================================================================
		// Write the generated code
		// Do write it using the correct indentation
		// =====================================================================

		// generate the enums
		for _, e := range r.CppEnum {
			if e.Namespace {

				lines := r.generateNamespaceEnum(e)
				for _, line := range lines {
					f.WriteString(indentation + line + "\n")
				}
			}
		}

		// =====================================================================
		// =====================================================================

		// Write the lines after the end line
		for i := endLine + 1; i < len(lines); i++ {
			f.WriteString(lines[i] + "\n")
		}

		return nil
	}
	return nil
}

func (e CppEnum) whichCodeToGenerate(g EnumGenerate) bool {
	for _, v := range e.Generate {
		// Do a case insensitive compare
		if strings.EqualFold(string(v), string(g)) {
			return true
		}
	}
	return false
}

func (r *CppCodeGenerator) generateEnum(e CppEnum) []string {
	lines := []string{}

	// generate the enum
	lines = append(lines, "enum "+e.Name+" {\n")
	for _, v := range e.Enums {
		lines = append(lines, v+", ")
	}
	lines = append(lines, "Count\n};\n")

	// generate the mask
	if e.whichCodeToGenerate(Mask) {
		lines = append(lines, "enum "+e.Name+"Mask {\n")
		for i, v := range e.Enums {
			lines = append(lines, v+"_mask = 1 << "+fmt.Sprint(i)+", ")
		}
		lines = append(lines, "\n};\n")
	}

	// generate the MaskToEnum function
	if e.whichCodeToGenerate(MaskToEnum) {
		lines = append(lines, "static "+e.Name+" MaskToEnum( u32 mask ) {\n")
		lines = append(lines, "s8 const bit = Math::Log2( mask );\n")
		lines = append(lines, "return ( bit >= 0 && bit < Count ) ? ("+e.Name+")bit : "+e.Name+"::Count;\n}\n")
	}

	// generate the EnumToMask function
	if e.whichCodeToGenerate(EnumToMask) {
		lines = append(lines, "static "+e.Name+"Mask EnumToMask( "+e.Name+" e ) {\n")
		lines = append(lines, "return ( e < "+e.Name+"::Count ) ? ("+e.Name+"Mask)( 1 << (u32)e ) : "+e.Name+"Mask::Count_mask;\n}\n")
	}

	// generate the ToString function
	if e.whichCodeToGenerate(ToString) {
		lines = append(lines, "static const char* s_value_names[] = {\n")
		for _, v := range e.Enums {
			lines = append(lines, "\""+v+"\", ")
		}
		lines = append(lines, "\"Count\"\n};\n")

		lines = append(lines, "static const char* ToString( "+e.Name+" e ) {\n")
		lines = append(lines, "return ((u32)e < "+e.Name+"::Count ? s_value_names[(int)e] : \"unsupported\" );\n}\n")
	}
	return lines
}

func (r *CppCodeGenerator) generateNamespaceEnum(e CppEnum) []string {
	lines := []string{}

	// generate the namespace
	lines = append(lines, "namespace "+e.Name+" {\n")

	// generate the enum
	lines = append(lines, "enum Enum {\n")
	for _, v := range e.Enums {
		lines = append(lines, v+", ")
	}
	lines = append(lines, "Count\n};\n")

	// generate the mask
	lines = append(lines, "enum Mask {\n")
	for i, v := range e.Enums {
		lines = append(lines, v+"_mask = 1 << "+fmt.Sprint(i)+", ")
	}
	lines = append(lines, "};\n")

	// generate the ToString function
	lines = append(lines, "static const char* s_value_names[] = {\n")
	for _, v := range e.Enums {
		lines = append(lines, "\""+v+"\", ")
	}
	lines = append(lines, "};\n")

	lines = append(lines, "static const char* ToString( Enum e ) {\n")
	lines = append(lines, "return ((u32)e < Enum::Count ? s_value_names[(int)e] : \"unsupported\" );\n}\n")

	// Close the namespace
	lines = append(lines, "} // namespace "+e.Name+"\n")

	return lines
}

// CppCodeGenerator contains information for generating C++ code
type CppCodeGenerator struct {
	Between string    `json:"between"`
	CppEnum []CppEnum `json:"enums"`
}

type EnumGenerate string

const (
	ToString   EnumGenerate = "ToGenerate"
	Mask       EnumGenerate = "Mask"
	MaskToEnum EnumGenerate = "MaskToEnum"
	EnumToMask EnumGenerate = "EnumToMask"
)

type CppEnum struct {
	Name      string         `json:"name"`
	Type      string         `json:"type,omitempty"`
	Namespace bool           `json:"namespace,omitempty"`
	Enums     []string       `json:"enums"`
	Generate  []EnumGenerate `json:"generate"`
}
