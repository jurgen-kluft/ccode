package embedded

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// This is the public function that will generate the C++ code
func GenerateCppCode(inputFile string, outputFile string) error {
	// Read the JSON file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Parse the JSON file
	r, err := newCppCodeGeneratorFromJSON(data)
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

type cppCode struct {
	lines       []string
	indent      string
	indentation string
}

type indentType string

const (
	indentTypeTab     indentType = "tab"
	indentTypeSpace   indentType = "space"
	indentTypeDefault indentType = "default"
)

func (cpp *cppCode) increaseIndentation() {
	cpp.indentation += cpp.indent
}

func (cpp *cppCode) decreaseIndentation() {
	cpp.indentation = cpp.indentation[:len(cpp.indentation)-len(cpp.indent)]
}

func (cpp *cppCode) appendTextLine(text ...string) {
	for _, t := range text {
		cpp.lines = append(cpp.lines, cpp.indentation+t)
	}
}

func (cpp *cppCode) addNamespaceOpen(name string) {
	cpp.appendTextLine("namespace " + name + "{")
	cpp.increaseIndentation()
}

func (cpp *cppCode) addNamespaceClose(name string) {
	cpp.decreaseIndentation()
	cpp.appendTextLine("} // namespace " + name)
}

// ---------------------------------------------------------------------------------

type cppCodeGenerator struct {
	between      string
	indentType   indentType
	indentSize   int
	memberPrefix string
	features     *FeatureFlags
	cppEnum      []cppEnum
	cppStruct    []cppStruct
}

func newCppCodeGenerator() *cppCodeGenerator {
	g := &cppCodeGenerator{}
	g.between = "== Generated Structs =="
	g.indentType = indentTypeSpace
	g.indentSize = 4
	g.memberPrefix = "m_"
	g.features = NewFeatureFlags()
	g.cppStruct = make([]cppStruct, 0)
	g.cppEnum = make([]cppEnum, 0)
	return g
}

func newCppCodeGeneratorFromJSON(data []byte) (*cppCodeGenerator, error) {
	r := &cppCodeGenerator{}
	err := json.Unmarshal(data, r)
	return r, err
}

func (g *cppCodeGenerator) setDefaults() {
	if g.between == "" {
		g.between = "== Generated Structs =="
	}
	if g.indentType == "" {
		g.indentType = indentTypeSpace
	}
	if g.indentSize <= 0 {
		g.indentSize = 4
	}
	if g.memberPrefix == "" {
		g.memberPrefix = "m_"
	}
	if g.features == nil {
		g.features = NewFeatureFlags()
	}
	if g.cppStruct == nil {
		g.cppStruct = make([]cppStruct, 0)
	}
}

// This will generate the C++ code to the output file
func (r *cppCodeGenerator) generate(outputFile string) error {

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
	for _, e := range r.cppEnum {
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
