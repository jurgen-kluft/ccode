package embedded

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
