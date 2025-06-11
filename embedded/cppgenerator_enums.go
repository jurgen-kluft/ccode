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

type CppEnumGenerator struct {
	IndentType IndentType `json:"indent_type,default=space"`
	IndentSize int        `json:"indent_size,default=4"`
	Between    string     `json:"between"`
	CppEnum    []CppEnum  `json:"enums"`
}

type EnumGenerate string

const (
	ToString   EnumGenerate = "ToString"
	Mask       EnumGenerate = "Mask"
	EnumToMask EnumGenerate = "EnumToMask"
)

type CppEnum struct {
	Name      string         `json:"name"`
	Namespace bool           `json:"namespace,omitempty"`
	Type      string         `json:"type,omitempty"`
	Members   []string       `json:"members"`
	Generate  []EnumGenerate `json:"generate"`
}

// This is the public function that will generate the C++ code
func GenerateCppEnums(inputFile string, outputFile string) error {
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

// C++ code generator
func cppCodeGeneratorFromJSON(data []byte) (*CppEnumGenerator, error) {
	r := &CppEnumGenerator{}
	err := json.Unmarshal(data, r)
	return r, err
}

// This will generate the C++ code to the output file
func (r *CppEnumGenerator) generate(outputFile string) error {

	// We need to open the file and find the between string, and replace everything between the begin and end line

	// Open the file
	// Read the full file into memory as and array of lines
	// Find the first between string, that is the begin line
	// Find the second between string, that is the end line
	// Replace everything between the begin and end line with the generated code
	// Write the file back to disk

	inStream, err := os.Open(outputFile)
	if err != nil {
		return err
	}
	defer inStream.Close()

	// Read the full file into memory as and array of lines
	lines := []string{}
	scanner := bufio.NewScanner(inStream)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Find the first between string, that is the begin line
	// Do keep track of what the indentation is for the begin line
	// Find the second between string, that is the end line
	betweenLine := ""
	initialIndentation := ""
	beginLine := -1
	endLine := -1
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}

		// Look for the between string using 'contains'
		if strings.Contains(line, r.Between) {
			if beginLine == -1 {
				betweenLine = line
				beginLine = i
				numSpaces := 0
				for _, c := range line {
					if c == ' ' {
						numSpaces++
					} else if c == '\t' {
						numSpaces += 4
					} else {
						break
					}
				}
				// Detected initial indentation
				initialIndentation = strings.Repeat(" ", numSpaces)
			} else {
				if endLine == -1 {
					endLine = i
					break
				}
			}
		}
	}

	if beginLine >= endLine {
		return fmt.Errorf("could not find the between string in the file")
	}

	// Replace everything between the begin and end line with the generated code
	// Write the file back to disk

	// Open the file for writing
	outStream, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outStream.Close()

	// Write the lines up to the begin line with the correct indentation
	for i := 0; i < beginLine; i++ {
		outStream.WriteString(lines[i] + "\n")
	}

	// =====================================================================
	// Write the generated code
	// Do write it using the correct indentation
	// =====================================================================

	outStream.WriteString(betweenLine + "\n")

	// generate the enums
	for _, e := range r.CppEnum {
		lines := r.generateEnum(initialIndentation, e)
		for _, line := range lines {
			outStream.WriteString(line + "\n")
		}
	}

	outStream.WriteString(betweenLine + "\n")

	// =====================================================================
	// =====================================================================

	for i := endLine + 1; i < len(lines); i++ {
		outStream.WriteString(lines[i] + "\n")
	}

	return nil
}

func (e CppEnum) whichCodeToGenerate(g EnumGenerate) bool {
	for _, v := range e.Generate {
		if strings.EqualFold(string(v), string(g)) {
			return true
		}
	}
	return false
}

func (r *CppEnumGenerator) generateEnum(initialIndentation string, e CppEnum) []string {
	indent := ""
	if r.IndentType == IndentTypeSpace || r.IndentType == IndentTypeDefault {
		indent = strings.Repeat(" ", r.IndentSize)
	} else if r.IndentType == IndentTypeTab {
		indent = "\t"
		indent = strings.Repeat(indent, r.IndentSize)
	}

	cppCode := &CppCode{
		Lines:       []string{},
		Indent:      indent,
		Indentation: initialIndentation,
	}

	// Set the enum header
	if e.Namespace {
		cppCode.addNamespaceOpen(e)
	}

	cppCode.addEnumEnum(e)

	if e.whichCodeToGenerate(Mask) {
		cppCode.addEnumMask(e)

		if e.whichCodeToGenerate(EnumToMask) {
			cppCode.addEnumToMaskFunction(e)
		}
	}

	if e.whichCodeToGenerate(ToString) {
		cppCode.addEnumToString(e)
	}

	// Set the enum footer
	if e.Namespace {
		cppCode.addNamespaceClose(e)
	}

	return cppCode.Lines
}

// =============================================================================
// CppCode enum specific functions
// =============================================================================

func (cpp *CppCode) addEnumEnum(enum CppEnum) {
	if enum.Namespace {
		cpp.appendTextLine("enum Enum {")
	} else {
		cpp.appendTextLine("enum " + enum.Name + " {")
	}
	cpp.increaseIndentation()
	for _, v := range enum.Members {
		cpp.appendTextLine(v + ",")
	}
	cpp.appendTextLine("Count")

	cpp.decreaseIndentation()
	cpp.appendTextLine("};")
}

func (cpp *CppCode) addEnumToString(enum CppEnum) {
	cpp.appendTextLine("")
	cpp.appendTextLine("static const char* s_value_" + strings.ToLower(enum.Name) + "_names[] = {")
	cpp.increaseIndentation()
	for _, v := range enum.Members {
		cpp.appendTextLine("\"" + v + "\",")
	}
	cpp.decreaseIndentation()
	cpp.appendTextLine("};")

	cpp.appendTextLine("")
	if enum.Namespace {
		cpp.appendTextLine("static const char* ToString(Enum e) {")
		cpp.increaseIndentation()
		cpp.appendTextLine("return (e < Enum::Count ? s_value_names[(int)e] : \"unsupported\" );")
		cpp.decreaseIndentation()
		cpp.appendTextLine("}")
	} else {
		cpp.appendTextLine("static const char* ToString(" + enum.Name + " e ) {")
		cpp.increaseIndentation()
		cpp.appendTextLine("return (e < " + enum.Name + "::Count ? s_value_" + strings.ToLower(enum.Name) + "_names[(int)e] : \"unsupported\" );")
		cpp.decreaseIndentation()
		cpp.appendTextLine("}")
	}
}

func (cpp *CppCode) addEnumMask(enum CppEnum) {
	cpp.appendTextLine("")
	if enum.Namespace {
		cpp.appendTextLine("enum Mask {")
	} else {
		cpp.appendTextLine("enum " + enum.Name + "_mask {")
	}
	cpp.increaseIndentation()
	for i, v := range enum.Members {
		cpp.appendTextLine(v + "_mask = " + v + " << " + fmt.Sprint(i) + ",")
	}
	cpp.decreaseIndentation()
	cpp.appendTextLine("};")
}

func (cpp *CppCode) addEnumToMaskFunction(enum CppEnum) {
	cpp.appendTextLine("")
	if enum.Namespace {
		cpp.appendTextLine("static Mask EnumToMask(Enum e) {")
		cpp.increaseIndentation()
		cpp.appendTextLine("return (Mask)(1 << (int)e);")
		cpp.decreaseIndentation()
		cpp.appendTextLine("}")
	} else {
		cpp.appendTextLine("static " + enum.Name + "_mask EnumToMask(" + enum.Name + " e) {")
		cpp.increaseIndentation()
		cpp.appendTextLine("return (" + enum.Name + "_mask)(1 << (int)e);")
		cpp.decreaseIndentation()
		cpp.appendTextLine("}")
	}
}
