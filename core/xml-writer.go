package corepkg

// --------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------
type XmlWriter struct {
	writer     *LineWriter
	Tags       []string
	DoBeginTag bool
	NoNewLine  bool
}

func NewXmlWriter() *XmlWriter {
	x := &XmlWriter{
		writer:     NewLineWriter(IndentModeSpaces),
		Tags:       make([]string, 0, 16),
		DoBeginTag: false,
		NoNewLine:  false,
	}
	return x
}

func (xml *XmlWriter) WriteToFile(filename string) error {
	return xml.writer.WriteToFile(filename)
}

// --------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------

type XmlTagScope struct {
	XmlWriter *XmlWriter
}

func NewXmlTagScope(xml *XmlWriter) *XmlTagScope {
	return &XmlTagScope{xml}
}

func (s XmlTagScope) NewXmlTagScope() XmlTagScope {
	newscope := XmlTagScope{s.XmlWriter}
	s.XmlWriter = nil
	return newscope
}

func (xml XmlTagScope) Close() {
	if xml.XmlWriter != nil {
		xml.XmlWriter.EndTag()
		xml.XmlWriter = nil
	}
}

// --------------------------------------------------------------------------------------------
// XmlWriter implementation
// --------------------------------------------------------------------------------------------

func (xml *XmlWriter) WriteHeader() {
	xml.writer.Write("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
}

func (xml *XmlWriter) WriteDocType(name, publicId, systemId string) {
	xml.newLine(0)
	xml.writer.Write("<!DOCTYPE ")
	xml.writer.Write(name)
	xml.writer.Write(" PUBLIC ")
	xml.QuoteString(publicId)
	xml.writer.Write(" ")
	xml.QuoteString(systemId)
	xml.writer.Write(">")
}

func (xml *XmlWriter) TagScope(name string) *XmlTagScope {
	xml.BeginTag(name)
	return &XmlTagScope{xml}
}

func (xml *XmlWriter) BeginTag(name string) {
	xml.CloseBeginTag()
	xml.newLine(0)
	xml.writer.Write("<")
	xml.writer.Write(name)
	xml.DoBeginTag = true
	xml.Tags = append(xml.Tags, name)
}

func (xml *XmlWriter) EndTag() {
	if len(xml.Tags) == 0 {
		panic("Error")
	}

	if xml.DoBeginTag {
		xml.writer.Write("/>")
		xml.DoBeginTag = false
	} else {
		xml.CloseBeginTag()
		xml.newLine(-1)
		xml.writer.Write("</")
		xml.writer.Write(xml.Tags[len(xml.Tags)-1])
		xml.writer.Write(">")
	}

	xml.Tags = xml.Tags[:len(xml.Tags)-1]
}

func (xml *XmlWriter) Attr(name, value string) {
	xml.writer.Write(" ")
	xml.writer.Write(name)
	xml.writer.Write("=")
	xml.QuoteString(value)
}

func (xml *XmlWriter) Body(text string) {
	xml.CloseBeginTag()
	xml.newLine(0)
	xml.writer.Write(text)
}

func (xml *XmlWriter) newLine(offset int) {
	if xml.NoNewLine {
		return
	}

	xml.writer.NewLine()
	n := len(xml.Tags) + offset
	for i := 0; i < n; i++ {
		xml.writer.Write("  ")
	}
}

func (xml *XmlWriter) TagWithBody(tagName, bodyText string) {
	xml.BeginTag(tagName)
	xml.NoNewLine = true
	xml.Body(bodyText)
	xml.EndTag()
	xml.NoNewLine = false
}

func (xml *XmlWriter) TagWithBodyBool(tagName string, b bool) {
	text := "true"
	if !b {
		text = "false"
	}
	xml.TagWithBody(tagName, text)
}

func (xml *XmlWriter) CloseBeginTag() {
	if !xml.DoBeginTag {
		return
	}
	xml.writer.Write(">")
	xml.DoBeginTag = false
}

func (xml *XmlWriter) QuoteString(v string) {
	xml.writer.Write("\"")

	for _, ch := range v {
		switch ch {
		case '"':
			xml.writer.Write("&quot;")
		case '\'':
			xml.writer.Write("&apos;")
		case '<':
			xml.writer.Write("&lt;")
		case '>':
			xml.writer.Write("&gt;")
		case '&':
			xml.writer.Write("&amp;")
		default:
			xml.writer.Write(string(ch))
		}
	}

	xml.writer.Write("\"")
}
