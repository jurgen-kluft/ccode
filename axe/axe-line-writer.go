package axe

// --------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------
type LineWriter struct {
	line  *stringBuilder
	lines []string
}

func NewLineWriter() *LineWriter {
	l := &LineWriter{
		line:  NewStringBuilder(),
		lines: make([]string, 0, 8192),
	}
	l.line.Grow(8192)
	return l
}

func (w *LineWriter) Write(str string) {
	w.line.WriteString(str)
}

func (w *LineWriter) WriteILine(str ...string) {
	for _, s := range str {
		w.line.WriteString(s)
	}
	w.NewLine()
}

func (w *LineWriter) WriteLine(str string) {
	w.line.WriteString(str)
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

func (w *LineWriter) finalize() {
	if w.line.Len() > 0 {
		w.lines = append(w.lines, w.line.String())
	}
}

func (w *LineWriter) WriteToFile(filename string) error {
	w.finalize()
	return WriteLinesToFile(filename, w.lines)
}
