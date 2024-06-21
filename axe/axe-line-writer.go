package axe

// --------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------
var spaceIndentationMap = map[string]string{
	"":         "",
	"+":        "    ",
	"++":       "        ",
	"+++":      "            ",
	"++++":     "                ",
	"+++++":    "                    ",
	"++++++":   "                        ",
	"+++++++":  "                            ",
	"++++++++": "                                ",
}

var tabsIndentationMap = map[string]string{
	"":         "",
	"+":        "\t",
	"++":       "\t\t",
	"+++":      "\t\t\t",
	"++++":     "\t\t\t\t",
	"+++++":    "\t\t\t\t\t",
	"++++++":   "\t\t\t\t\t\t",
	"+++++++":  "\t\t\t\t\t\t\t",
	"++++++++": "\t\t\t\t\t\t\t\t",
}

type LineWriter struct {
	line   *stringBuilder
	lines  []string
	indent map[string]string
}

type IndentMode int

const (
	IndentModeSpaces IndentMode = iota
	IndentModeTabs
)

func NewLineWriter(mode IndentMode) *LineWriter {
	l := &LineWriter{
		line:  NewStringBuilder(),
		lines: make([]string, 0, 8192),
	}
	if mode == IndentModeSpaces {
		l.indent = spaceIndentationMap
	} else {
		l.indent = tabsIndentationMap
	}
	l.line.Grow(8192)
	return l
}

func (w *LineWriter) finalize() {
	if w.line.Len() > 0 {
		w.lines = append(w.lines, w.line.String())
	}
}

// Enumerate will call the enumerator function for each key-value pair in the dictionary.
//
//	'last' will be 0 for all but the last key-value pair, and 1 for the last key-value pair.
func (w *LineWriter) Enumerate(enumerator func(i int, line string, last int)) {
	n := len(w.lines) - 1
	for i, line := range w.lines {
		last := 0
		if i == n {
			last = 1
		}
		enumerator(i, line, last)
	}
}

func (w *LineWriter) IsEmpty() bool {
	return len(w.lines) == 0 && w.line.Len() == 0
}

func (w *LineWriter) Write(strs ...string) {
	for _, str := range strs {
		w.line.WriteString(str)
	}
}

func (w *LineWriter) Append(other *LineWriter) {
	other.finalize()
	w.lines = append(w.lines, other.lines...)
}

func (w *LineWriter) WriteILine(indent string, str ...string) {
	w.line.WriteString(w.indent[indent])
	for _, s := range str {
		w.line.WriteString(s)
	}
	w.NewLine()
}

func (w *LineWriter) WriteLine(strs ...string) {
	for _, str := range strs {
		w.line.WriteString(str)
	}
	w.lines = append(w.lines, w.line.String())
	w.line.Reset()
}

func (w *LineWriter) WriteLines(strs ...string) {
	for _, str := range strs {
		w.WriteLine(str)
	}
}

func (w *LineWriter) WriteManyLines(str []string) {
	for _, line := range str {
		w.WriteLine(line)
	}
}

func (w *LineWriter) NewLine() {
	w.lines = append(w.lines, w.line.String())
	w.line.Reset()
}

func (w *LineWriter) WriteToFile(filename string) error {
	w.finalize()
	return WriteLinesToFile(filename, w.lines)
}
