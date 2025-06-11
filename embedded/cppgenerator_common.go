package embedded

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type CppCode struct {
	Lines       []string
	Indent      string
	Indentation string
}

type IndentType string

const (
	IndentTypeTab     IndentType = "tab"
	IndentTypeSpace   IndentType = "space"
	IndentTypeDefault IndentType = "default"
)

func (cpp *CppCode) increaseIndentation() {
	cpp.Indentation += cpp.Indent
}

func (cpp *CppCode) decreaseIndentation() {
	cpp.Indentation = cpp.Indentation[:len(cpp.Indentation)-len(cpp.Indent)]
}

func (cpp *CppCode) appendTextLine(text ...string) {
	for _, t := range text {
		cpp.Lines = append(cpp.Lines, cpp.Indentation+t)
	}
}

func (cpp *CppCode) addNamespaceOpen(enum CppEnum) {
	cpp.appendTextLine("namespace " + enum.Name + "{")
	cpp.increaseIndentation()
}

func (cpp *CppCode) addNamespaceClose(enum CppEnum) {
	cpp.decreaseIndentation()
	cpp.appendTextLine("} // namespace " + enum.Name)
}

// ---------------------------------------------------------------------------------

type CppCodeGenerator struct {
	Between      string        `json:"between,omitempty"`
	IndentType   IndentType    `json:"indent_type,omitempty"`
	IndentSize   int           `json:"indent_size,omitempty"`
	MemberPrefix string        `json:"member_prefix,omitempty"`
	Features     *FeatureFlags `json:"features,omitempty"`
	CppEnum      []CppEnum     `json:"enums,omitempty"`
	CppStruct    []CppStruct   `json:"structs"`
}

func NewCppCodeGenerator() *CppCodeGenerator {
	g := &CppCodeGenerator{}
	g.Between = "== Generated Structs =="
	g.IndentType = IndentTypeSpace
	g.IndentSize = 4
	g.MemberPrefix = "m_"
	g.Features = NewFeatureFlags()
	g.CppStruct = make([]CppStruct, 0)
	g.CppEnum = make([]CppEnum, 0)
	return g
}

func NewCppCodeGeneratorFromJSON(data []byte) (*CppCodeGenerator, error) {
	r := &CppCodeGenerator{}
	err := json.Unmarshal(data, r)
	return r, err
}

func (g *CppCodeGenerator) SetDefaults() {
	if g.Between == "" {
		g.Between = "== Generated Structs =="
	}
	if g.IndentType == "" {
		g.IndentType = IndentTypeSpace
	}
	if g.IndentSize <= 0 {
		g.IndentSize = 4
	}
	if g.MemberPrefix == "" {
		g.MemberPrefix = "m_"
	}
	if g.Features == nil {
		g.Features = NewFeatureFlags()
	}
	if g.CppStruct == nil {
		g.CppStruct = make([]CppStruct, 0)
	}
}

func (g *CppCodeGenerator) UnmarshalJSON(text []byte) error {
	type Alias CppCodeGenerator
	aux := Alias{}
	if err := json.Unmarshal(text, &aux); err != nil {
		return err
	}
	*g = CppCodeGenerator(aux)
	g.SetDefaults()
	return nil
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
