package corepkg

import "unicode/utf8"

// --------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------

type LineWriter struct {
	line           *StringBuilder
	lineLen        int
	lines          []string
	tabstops       []int
	indent         map[string]string
	indentNumChars int
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
		l.indentNumChars = 4
		l.indent = map[string]string{
			"":         "    ",
			"+":        "        ",
			"++":       "            ",
			"+++":      "                ",
			"++++":     "                    ",
			"+++++":    "                        ",
			"++++++":   "                            ",
			"+++++++":  "                                ",
			"++++++++": "                                    ",
		}
	} else {
		l.indentNumChars = 4
		l.indent = map[string]string{
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
	}
	l.line.Grow(8192)
	return l
}

func (w *LineWriter) Clear() {
	w.line.Reset()
	w.lineLen = 0
	w.lines = make([]string, 0, 8192)
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
		w.lineLen += utf8.RuneCountInString(str)
		w.line.WriteString(str)
	}
}

func (w *LineWriter) Append(other *LineWriter) {
	other.finalize()
	w.lines = append(w.lines, other.lines...)
}

func (w *LineWriter) WriteILine(indent string, strs ...string) {
	w.line.WriteString(w.indent[indent])
	for _, str := range strs {
		w.lineLen += utf8.RuneCountInString(str)
		w.line.WriteString(str)
	}
	w.NewLine()
}

type TabStop int
type EndOfLine int

func (w *LineWriter) SetTabStops(stops ...int) {
	tabstop := 0
	for _, stop := range stops {
		if stop > tabstop {
			tabstop = stop
			w.tabstops = append(w.tabstops, stop)
		}
	}
}

func (w *LineWriter) WriteAligned(strs ...any) {
	// Example:
	//           linewriter.SetTabStops(32, 64, 96)
	//           linewriter.WriteAligned("PRODUCT_FRAMEWORK", 0, `:= `, project.Name)
	// Output: "PRODUCT_FRAMEWORK               := project.Name"

	tabstop := 0
	for _, s := range strs {
		switch val := s.(type) {
		case EndOfLine:
			w.NewLine()
			continue
		case TabStop:
			i := int(val)
			if i < len(w.tabstops) {
				if w.tabstops[i] > tabstop {
					tabstop = w.tabstops[i]
				}
			}
		case string:
			// Move the carrot to the tabstop position if it's not already there
			for w.lineLen < tabstop {
				w.line.WriteString(" ")
				w.lineLen += 1 // Currently every indentation is 4 'characters'
			}
			w.lineLen += utf8.RuneCountInString(val)
			w.line.WriteString(val)
		}
	}
}

func (w *LineWriter) WriteAlignedLine(strs ...interface{}) {
	w.WriteAligned(strs...)
	w.NewLine()
}

func (w *LineWriter) WriteLine(strs ...string) {
	for _, str := range strs {
		w.lineLen += utf8.RuneCountInString(str)
		w.line.WriteString(str)
	}
	w.lines = append(w.lines, w.line.String())
	w.lineLen = 0
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
	w.lineLen = 0
	w.line.Reset()
}

func (w *LineWriter) WriteToFile(filename string) error {
	w.finalize()
	return WriteLinesToFile(filename, w.lines)
}
