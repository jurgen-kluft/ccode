package embedded

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jurgen-kluft/ccode/foundation"
)

// This is the public function that will generate the C++ code
func GenerateCppCode(inputFile string, outputFile string) error {
	// Read the JSON file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Parse the JSON file
	decoder := foundation.NewJsonDecoder()
	if !decoder.Begin(string(data)) {
		return fmt.Errorf("error decoding JSON from file %s", inputFile)
	}
	r := newCppCodeGenerator()
	if err = r.decodeJSON(decoder); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	// insert the generated code into the output file
	err = r.insertGeneratedCode(outputFile)
	if err != nil {
		return fmt.Errorf("error inserting generated C++ code into file '%s' with error %v", outputFile, err)
	}

	return nil
}

type indentType string

const (
	indentTypeTab     indentType = "tab"
	indentTypeSpace   indentType = "space"
	indentTypeDefault indentType = "default"
)

type cppCode struct {
	lines      []string
	indent     string
	indentType indentType
	indentSize int
}

func newCppCode(initialIndentation string, indentType indentType, indentSize int) *cppCode {
	cpp := &cppCode{
		lines:      make([]string, 0, 1024),
		indent:     initialIndentation,
		indentType: indentType,
		indentSize: indentSize,
	}
	return cpp
}

func (cpp *cppCode) increaseIndentation() {
	if cpp.indentType == indentTypeSpace {
		cpp.indent += strings.Repeat(" ", cpp.indentSize)
	} else if cpp.indentType == indentTypeTab {
		cpp.indent += strings.Repeat("\t", cpp.indentSize)
	}
}

func (cpp *cppCode) decreaseIndentation() {
	l := len(cpp.indent)
	if cpp.indentType == indentTypeSpace {
		if l >= cpp.indentSize {
			cpp.indent = cpp.indent[:l-cpp.indentSize]
		} else {
			cpp.indent = ""
		}
	} else if cpp.indentType == indentTypeTab {
		if l >= cpp.indentSize {
			cpp.indent = cpp.indent[:l-cpp.indentSize]
		} else {
			cpp.indent = ""
		}
	}
}

func (cpp *cppCode) appendTextLine(text ...string) {
	line := cpp.indent
	for _, t := range text {
		line += t
	}
	cpp.lines = append(cpp.lines, line)
}

func (cpp *cppCode) addNamespaceOpen(name string) {
	cpp.appendTextLine("namespace ", name, "{")
	cpp.increaseIndentation()
}

func (cpp *cppCode) addNamespaceClose(name string) {
	cpp.decreaseIndentation()
	cpp.appendTextLine("} // namespace ", name)
}

// ---------------------------------------------------------------------------------

type cppCodeGenerator struct {
	between      string
	indentType   indentType
	indentSize   int
	memberPrefix string
	features     *featureFlags
	bitsetType   string
	bitsetSize   int
	cppEnum      []cppEnum
	cppStruct    []cppStruct
}

func newCppCodeGenerator() *cppCodeGenerator {
	g := &cppCodeGenerator{}
	g.between = "== Generated Structs =="
	g.indentType = indentTypeSpace
	g.indentSize = 4
	g.memberPrefix = "m_"
	g.features = newFeatureFlags()
	g.bitsetType = "u8" // Default bitset type
	g.bitsetSize = 8    // Default bitset size (8 bits)
	g.cppStruct = make([]cppStruct, 0)
	g.cppEnum = make([]cppEnum, 0)
	return g
}

func decodeFeatureFlags(decoder *foundation.JsonDecoder, ff *featureFlags) *featureFlags {
	ff.decodeJSON(decoder)
	return ff
}

func decodeCppEnum(decoder *foundation.JsonDecoder, ce *cppEnum) {
	fields := map[string]foundation.JsonDecode{
		"name":      func(decoder *foundation.JsonDecoder) { ce.name = decoder.DecodeString() },
		"namespace": func(decoder *foundation.JsonDecoder) { ce.namespace = decoder.DecodeBool() },
		"enumtype":  func(decoder *foundation.JsonDecoder) { ce.enumType = decoder.DecodeString() },
		"members":   func(decoder *foundation.JsonDecoder) { ce.members = decoder.DecodeStringArray() },
		"generate": func(decoder *foundation.JsonDecoder) {
			ce.generate = make([]enumGenerate, 0, 4)
			i := 0
			for !decoder.ReadUntilArrayEnd() {
				ce.generate = append(ce.generate, enumGenerate(decoder.DecodeString()))
				i++
			}
		},
	}
	decoder.Decode(fields)
}

// type cppStructMember struct {
// 	Name        string `json:"name"`
// 	Type        string `json:"type,omitempty"`
// 	Initializer string `json:"initializer,omitempty"`
// }

func decodeCppStructMember(decoder *foundation.JsonDecoder, csm *cppStructMember) {
	decodeInitializer := func(decoder *foundation.JsonDecoder) { csm.initializer = decoder.DecodeString() }
	decodeType := func(decoder *foundation.JsonDecoder) { csm.memberType = decoder.DecodeString() }
	fields := map[string]foundation.JsonDecode{
		"name":        func(decoder *foundation.JsonDecoder) { csm.name = decoder.DecodeString() },
		"type":        decodeType,
		"membertype":  decodeType,
		"init":        decodeInitializer,
		"initializer": decodeInitializer,
	}
	decoder.Decode(fields)
}

func decodeCppStruct(decoder *foundation.JsonDecoder, cs *cppStruct) {
	memberPrefix := func(decoder *foundation.JsonDecoder) { cs.memberPrefix = decoder.DecodeString() }
	fields := map[string]foundation.JsonDecode{
		"name":         func(decoder *foundation.JsonDecoder) { cs.name = decoder.DecodeString() },
		"prefix":       memberPrefix,
		"memberprefix": memberPrefix,
		"features":     func(decoder *foundation.JsonDecoder) { cs.features = decodeFeatureFlags(decoder, newFeatureFlags()) },
		"members": func(decoder *foundation.JsonDecoder) {
			cs.members = make([]cppStructMember, 0, 4)
			for !decoder.ReadUntilArrayEnd() {
				cs.members = append(cs.members, newCppStructMember())
				decodeCppStructMember(decoder, &cs.members[len(cs.members)-1])
			}
		},
	}
	decoder.Decode(fields)
}

func (r *cppCodeGenerator) decodeJSON(decoder *foundation.JsonDecoder) error {
	indentType := func(decoder *foundation.JsonDecoder) { r.indentType = indentType(decoder.DecodeString()) }
	indentSize := func(decoder *foundation.JsonDecoder) { r.indentSize = int(decoder.DecodeInt32()) }
	memberPrefix := func(decoder *foundation.JsonDecoder) { r.memberPrefix = decoder.DecodeString() }
	bitsetType := func(decoder *foundation.JsonDecoder) { r.bitsetType = decoder.DecodeString() }
	bitsetSize := func(decoder *foundation.JsonDecoder) { r.bitsetSize = int(decoder.DecodeInt32()) }
	fields := map[string]foundation.JsonDecode{
		"between":      func(decoder *foundation.JsonDecoder) { r.between = decoder.DecodeString() },
		"indenttype":   indentType,
		"indent_type":  indentType,
		"indentsize":   indentSize,
		"indent_size":  indentSize,
		"prefix":       memberPrefix,
		"memberprefix": memberPrefix,
		"features":     func(decoder *foundation.JsonDecoder) { r.features = decodeFeatureFlags(decoder, newFeatureFlags()) },
		"bitsettype":   bitsetType,
		"bitset_type":  bitsetType,
		"bitsetsize":   bitsetSize,
		"bitset_size":  bitsetSize,
		"enums": func(decoder *foundation.JsonDecoder) {
			r.cppEnum = make([]cppEnum, 0, 4)
			for !decoder.ReadUntilArrayEnd() {
				r.cppEnum = append(r.cppEnum, newCppEnum())
				decodeCppEnum(decoder, &r.cppEnum[len(r.cppEnum)-1])
			}
		},
		"structs": func(decoder *foundation.JsonDecoder) {
			r.cppStruct = make([]cppStruct, 0, 4)
			for !decoder.ReadUntilArrayEnd() {
				r.cppStruct = append(r.cppStruct, newCppStruct())
				decodeCppStruct(decoder, &r.cppStruct[len(r.cppStruct)-1])
			}
		},
	}
	return decoder.Decode(fields)
}

func (r *cppCodeGenerator) generateCppCode(initialIndentation string) []string {
	generatedCode := make([]string, 0, 1024)
	for _, e := range r.cppEnum {
		lines := r.generateEnum(initialIndentation, e)
		generatedCode = append(generatedCode, lines...)
	}

	for _, s := range r.cppStruct {
		lines := r.generateStruct(initialIndentation, &s)
		generatedCode = append(generatedCode, lines...)
	}

	return generatedCode
}

// This will generate the C++ code to the output file
func (r *cppCodeGenerator) insertGeneratedCode(outputFile string) error {

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
	// Do keep track of what the indent is for the begin line
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
		if strings.Contains(line, r.between) {
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
				// Detected initial indent
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

	// Write the lines up to the begin line with the correct indent
	for i := 0; i < beginLine; i++ {
		outStream.WriteString(lines[i] + "\n")
	}

	// =====================================================================
	// Write the generated code
	// Do write it using the correct indent
	// =====================================================================
	outStream.WriteString(betweenLine + "\n")
	generatedCode := r.generateCppCode(initialIndentation)
	for _, line := range generatedCode {
		if len(line) > 0 {
			// Write the line with the initial indent
			outStream.WriteString(initialIndentation + line + "\n")
		} else {
			outStream.WriteString("\n")
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
