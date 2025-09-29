package embedded

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
)

// This is the public function that will generate the C++ code
func GenerateCppCode(inputFile string, outputFile string) error {
	// Read the JSON file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Parse the JSON file
	decoder := corepkg.NewJsonDecoder()
	if !decoder.Begin(string(data)) {
		return fmt.Errorf("error decoding JSON from file %s", inputFile)
	}
	r := newCppCodeGenerator()
	if err = r.decodeJSON(decoder); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	inputFileLines, err := r.readFileLines(outputFile)
	if err != nil {
		return fmt.Errorf("error reading input file '%s': %v", outputFile, err)
	}

	// generate the C++ code
	generatedCode := r.generateCppCode()

	// insert the generated code into the input file
	if outputFileLines, err := r.insertGeneratedCode(inputFileLines, generatedCode); err != nil {
		return fmt.Errorf("error inserting generated C++ code into file '%s' with error %v", outputFile, err)
	} else {
		if err := r.writeFileLines(outputFile, outputFileLines); err != nil {
			return fmt.Errorf("error writing output file '%s': %v", outputFile, err)
		}
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

func newCppCode(indentType indentType, indentSize int) *cppCode {
	cpp := &cppCode{
		lines:      make([]string, 0, 1024),
		indent:     "",
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

func decodeFeatureFlags(decoder *corepkg.JsonDecoder, ff *featureFlags) *featureFlags {
	ff.decodeJSON(decoder)
	return ff
}

func decodeCppEnum(decoder *corepkg.JsonDecoder, ce *cppEnum) {
	fields := map[string]corepkg.JsonDecode{
		"name":      func(decoder *corepkg.JsonDecoder) { ce.name = decoder.DecodeString() },
		"namespace": func(decoder *corepkg.JsonDecoder) { ce.namespace = decoder.DecodeBool() },
		"enumtype":  func(decoder *corepkg.JsonDecoder) { ce.enumType = decoder.DecodeString() },
		"members":   func(decoder *corepkg.JsonDecoder) { ce.members = decoder.DecodeStringArray() },
		"generate": func(decoder *corepkg.JsonDecoder) {
			ce.generate = make([]enumGenerate, 0, 4)
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				ce.generate = append(ce.generate, enumGenerate(decoder.DecodeString()))
			})
		},
	}
	decoder.Decode(fields)
}

// type cppStructMember struct {
// 	Name        string `json:"name"`
// 	Type        string `json:"type,omitempty"`
// 	Initializer string `json:"initializer,omitempty"`
// }

func decodeCppStructMember(decoder *corepkg.JsonDecoder, csm *cppStructMember) {
	decodeInitializer := func(decoder *corepkg.JsonDecoder) { csm.initializer = decoder.DecodeString() }
	decodeType := func(decoder *corepkg.JsonDecoder) { csm.memberType = decoder.DecodeString() }
	fields := map[string]corepkg.JsonDecode{
		"name":        func(decoder *corepkg.JsonDecoder) { csm.name = decoder.DecodeString() },
		"type":        decodeType,
		"membertype":  decodeType,
		"init":        decodeInitializer,
		"initializer": decodeInitializer,
	}
	decoder.Decode(fields)
}

func decodeCppStruct(decoder *corepkg.JsonDecoder, cs *cppStruct) {
	memberPrefix := func(decoder *corepkg.JsonDecoder) { cs.memberPrefix = decoder.DecodeString() }
	fields := map[string]corepkg.JsonDecode{
		"name":         func(decoder *corepkg.JsonDecoder) { cs.name = decoder.DecodeString() },
		"prefix":       memberPrefix,
		"memberprefix": memberPrefix,
		"features":     func(decoder *corepkg.JsonDecoder) { cs.features = decodeFeatureFlags(decoder, newFeatureFlags()) },
		"members": func(decoder *corepkg.JsonDecoder) {
			cs.members = make([]cppStructMember, 0, 4)
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				cs.members = append(cs.members, newCppStructMember())
				decodeCppStructMember(decoder, &cs.members[len(cs.members)-1])
			})
		},
	}
	decoder.Decode(fields)
}

func (r *cppCodeGenerator) decodeJSON(decoder *corepkg.JsonDecoder) error {
	indentType := func(decoder *corepkg.JsonDecoder) { r.indentType = indentType(decoder.DecodeString()) }
	indentSize := func(decoder *corepkg.JsonDecoder) { r.indentSize = int(decoder.DecodeInt32()) }
	memberPrefix := func(decoder *corepkg.JsonDecoder) { r.memberPrefix = decoder.DecodeString() }
	bitsetType := func(decoder *corepkg.JsonDecoder) { r.bitsetType = decoder.DecodeString() }
	bitsetSize := func(decoder *corepkg.JsonDecoder) { r.bitsetSize = int(decoder.DecodeInt32()) }
	fields := map[string]corepkg.JsonDecode{
		"between":      func(decoder *corepkg.JsonDecoder) { r.between = decoder.DecodeString() },
		"indenttype":   indentType,
		"indent_type":  indentType,
		"indentsize":   indentSize,
		"indent_size":  indentSize,
		"prefix":       memberPrefix,
		"memberprefix": memberPrefix,
		"features":     func(decoder *corepkg.JsonDecoder) { r.features = decodeFeatureFlags(decoder, newFeatureFlags()) },
		"bitsettype":   bitsetType,
		"bitset_type":  bitsetType,
		"bitsetsize":   bitsetSize,
		"bitset_size":  bitsetSize,
		"enums": func(decoder *corepkg.JsonDecoder) {
			r.cppEnum = make([]cppEnum, 0, 4)
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				r.cppEnum = append(r.cppEnum, newCppEnum())
				decodeCppEnum(decoder, &r.cppEnum[len(r.cppEnum)-1])
			})
		},
		"structs": func(decoder *corepkg.JsonDecoder) {
			r.cppStruct = make([]cppStruct, 0, 4)
			decoder.DecodeArray(func(decoder *corepkg.JsonDecoder) {
				r.cppStruct = append(r.cppStruct, newCppStruct())
				decodeCppStruct(decoder, &r.cppStruct[len(r.cppStruct)-1])
			})
		},
	}
	return decoder.Decode(fields)
}

func (r *cppCodeGenerator) generateCppCode() []string {
	generatedCode := make([]string, 0, 1024)
	for _, e := range r.cppEnum {
		lines := r.generateEnum(e)
		generatedCode = append(generatedCode, lines...)
	}

	for _, s := range r.cppStruct {
		lines := r.generateStruct(&s)
		generatedCode = append(generatedCode, lines...)
	}

	return generatedCode
}

func (r *cppCodeGenerator) readFileLines(filepath string) ([]string, error) {
	inStream, err := os.Open(filepath)
	if err != nil {
		return []string{}, err
	}
	defer inStream.Close()

	// Read the full file into memory as and array of lines
	lines := make([]string, 0, 8192)
	scanner := bufio.NewScanner(inStream)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func (r *cppCodeGenerator) writeFileLines(filepath string, lines []string) error {
	outStream, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer outStream.Close()

	// Turn this into a buffered writer
	writer := bufio.NewWriter(outStream)

	// Write the lines to the file
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("error writing to file %s: %v", filepath, err)
		}
	}

	if err := writer.Flush(); err != nil { // Ensure all data is written to the file
		return fmt.Errorf("error flushing file %s: %v", filepath, err)
	}

	if err := outStream.Sync(); err != nil {
		return fmt.Errorf("error syncing file %s: %v", filepath, err)
	}

	return nil
}

// This will generate the C++ code to the output file
func (r *cppCodeGenerator) insertGeneratedCode(inputFileLines []string, generatedCode []string) (outputFileLines []string, err error) {

	// Find the first between string, that is the begin line
	// Do keep track of what the indent is for the begin line
	// Find the second between string, that is the end line
	betweenLine := ""
	initialIndentation := 0
	beginLine := -1
	endLine := -1
	for i, line := range inputFileLines {
		if len(line) == 0 {
			continue
		}

		// Look for the between string using 'contains'
		pos := strings.Index(line, r.between)
		if pos >= 0 {
			if beginLine == -1 {
				betweenLine = line
				beginLine = i
				initialIndentation = pos
			} else {
				if endLine == -1 {
					endLine = i
					break
				}
			}
		}
	}

	if beginLine >= endLine {
		err = fmt.Errorf("could not find the first and second 'between' line")
		return
	}

	// Extract the initial indentation from the first between line
	initialIndentationStr := betweenLine[:initialIndentation]

	// Reserve enough space for the output file lines
	outputFileLines = make([]string, 0, len(inputFileLines)+len(generatedCode)+4)

	// =====================================================================
	// Write the lines from the input file before the first between line
	// =====================================================================
	inputFileHeaderLines := inputFileLines[:beginLine]
	outputFileLines = append(outputFileLines, inputFileHeaderLines...)

	// =====================================================================
	// Insert the generated code
	// =====================================================================
	outputFileLines = append(outputFileLines, betweenLine)
	for _, line := range generatedCode {
		if len(line) > 0 {
			outputFileLines = append(outputFileLines, initialIndentationStr+line)
		} else {
			outputFileLines = append(outputFileLines, "")
		}
	}
	outputFileLines = append(outputFileLines, betweenLine)

	// =====================================================================
	// Write the lines of the input file after the second between line
	// =====================================================================
	inputFileFooterLines := inputFileLines[endLine+1:]
	outputFileLines = append(outputFileLines, inputFileFooterLines...)

	return
}
