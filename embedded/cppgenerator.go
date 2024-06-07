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
func GenerateCppCode(inputFile string, outputFile string) error {
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
	beginLine := -1
	endLine := -1
	indentation := ""
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
				indentation = strings.Repeat(" ", numSpaces)
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
		lines := r.generateEnum(indentation, e)
		for _, line := range lines {
			outStream.WriteString(line + "\n")
		}
	}

	outStream.WriteString(betweenLine + "\n")

	// =====================================================================
	// =====================================================================

	// Write the lines after the end line
	for i := endLine + 1; i < len(lines); i++ {
		outStream.WriteString(lines[i] + "\n")
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

func (r *CppCodeGenerator) increaseIndentation(indentation string) string {
	return r.Indentation + indentation
}

func (r *CppCodeGenerator) decreaseIndentation(indentation string) string {
	return indentation[len(r.Indentation):]
}

func (r *CppCodeGenerator) generateEnum(indentation string, e CppEnum) []string {
	cppCode := &CppCode{}

	// Set the enum header
	if e.Namespace {
		cppCode.addNamespaceOpen(indentation, e)
		indentation = r.increaseIndentation(indentation) // Set the namespace indentation
	}

	cppCode.addEnumEnum(indentation, e)

	if e.whichCodeToGenerate(Mask) {
		cppCode.addEnumMask(indentation, e)

		if e.whichCodeToGenerate(EnumToMask) {
			cppCode.addEnumToMaskFunction(indentation, e)
		}
	}

	if e.whichCodeToGenerate(ToString) {
		cppCode.addEnumToString(indentation, e)
	}

	// Set the enum footer
	if e.Namespace {
		indentation = r.decreaseIndentation(indentation) // Set the namespace indentation
		cppCode.addNamespaceClose(indentation, e)
	}

	return cppCode.Lines
}

type CppCode struct {
	Lines []string
}

// CppCodeGenerator contains information for generating C++ code
type CppCodeGenerator struct {
	Indentation string    `json:"indentation"`
	Between     string    `json:"between"`
	CppEnum     []CppEnum `json:"enums"`
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
	Enums     []string       `json:"enums"`
	Generate  []EnumGenerate `json:"generate"`
}

// ---------------------------------------------------------------------------------------------------------
// C++ code generation functions for a CppEnum
// ---------------------------------------------------------------------------------------------------------

func (cpp *CppCode) appendTextLine(text string) {
	cpp.Lines = append(cpp.Lines, text)
}

func (cpp *CppCode) appendTextLines(text []string) {
	cpp.Lines = append(cpp.Lines, text...)
}

func (cpp *CppCode) addNamespaceOpen(indentation string, enum CppEnum) {
	cpp.appendTextLine(indentation + "namespace " + enum.Name + "{")
}

func (cpp *CppCode) addEnumEnum(indentation string, enum CppEnum) {
	if enum.Namespace {
		cpp.appendTextLine(indentation + "enum Enum {")
	} else {
		cpp.appendTextLine(indentation + "enum " + enum.Name + " {")
	}
	for _, v := range enum.Enums {
		cpp.appendTextLine(indentation + "    " + v + ",")
	}
	cpp.appendTextLine(indentation + "    Count")
	cpp.appendTextLine(indentation + "};")
}

func (cpp *CppCode) addEnumToString(indentation string, enum CppEnum) {
	cpp.appendTextLine("")
	cpp.appendTextLine(indentation + "static const char* s_value_" + strings.ToLower(enum.Name) + "_names[] = {")
	for _, v := range enum.Enums {
		cpp.appendTextLine(indentation + "    \"" + v + "\",")
	}
	cpp.appendTextLine(indentation + "};")
	cpp.appendTextLine("")
	if enum.Namespace {
		cpp.appendTextLine(indentation + "static const char* ToString(Enum e) {")
		cpp.appendTextLine(indentation + "    return (e < Enum::Count ? s_value_names[(int)e] : \"unsupported\" );")
	} else {
		cpp.appendTextLine(indentation + "static const char* ToString(" + enum.Name + " e ) {")
		cpp.appendTextLine(indentation + "    return (e < " + enum.Name + "::Count ? s_value_" + strings.ToLower(enum.Name) + "_names[(int)e] : \"unsupported\" );")
	}
	cpp.appendTextLine(indentation + "}")
}

func (cpp *CppCode) addEnumMask(indentation string, enum CppEnum) {
	cpp.appendTextLine("")
	if enum.Namespace {
		cpp.appendTextLine(indentation + "enum Mask {")
	} else {
		cpp.appendTextLine(indentation + "enum " + enum.Name + "_mask {")
	}
	for i, v := range enum.Enums {
		cpp.appendTextLine(indentation + "    " + v + "_mask = " + v + " << " + fmt.Sprint(i) + ",")
	}
	cpp.appendTextLine(indentation + "};")
}

func (cpp *CppCode) addEnumToMaskFunction(indentation string, enum CppEnum) {
	cpp.appendTextLine("")
	if enum.Namespace {
		cpp.appendTextLine(indentation + "static Mask EnumToMask(Enum e) {")
		cpp.appendTextLine(indentation + "    return (Mask)(1 << (int)e);")
	} else {
		cpp.appendTextLine(indentation + "static " + enum.Name + "_mask EnumToMask(" + enum.Name + " e) {")
		cpp.appendTextLine(indentation + "    return (" + enum.Name + "_mask)(1 << (int)e);")
	}
	cpp.appendTextLine(indentation + "}")
}

func (cpp *CppCode) addNamespaceClose(indentation string, enum CppEnum) {
	cpp.appendTextLine(indentation + "} // namespace " + enum.Name)
}
