package xcode

import "strings"

type XmlWriter struct {
	Buffer     strings.Builder
	Tags       []string
	DoBeginTag bool
	NoNewLine  bool
}

func NewXmlWriter() *XmlWriter {
	return &XmlWriter{}
}

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

func (xml *XmlWriter) writeHeader() {
	xml.Buffer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
}

func (xml *XmlWriter) writeDocType(name, publicId, systemId string) {
	xml.NewLine(0)
	xml.Buffer.WriteString("<!DOCTYPE ")
	xml.Buffer.WriteString(name)
	xml.Buffer.WriteString(" PUBLIC ")
	xml.quoteString(publicId)
	xml.Buffer.WriteString(" ")
	xml.quoteString(systemId)
	xml.Buffer.WriteString(">")
}

func (xml *XmlWriter) TagScope(name string) *XmlTagScope {
	xml.BeginTag(name)
	return &XmlTagScope{xml}
}

func (xml *XmlWriter) BeginTag(name string) {
	xml.CloseBeginTag()
	xml.NewLine(0)
	xml.Buffer.WriteString("<")
	xml.Buffer.WriteString(name)
	xml.DoBeginTag = true
	xml.Tags = append(xml.Tags, name)
}

func (xml *XmlWriter) EndTag() {
	if len(xml.Tags) == 0 {
		panic("Error")
	}

	if xml.DoBeginTag {
		xml.Buffer.WriteString("/>")
		xml.DoBeginTag = false
	} else {
		xml.CloseBeginTag()
		xml.NewLine(-1)
		xml.Buffer.WriteString("</")
		xml.Buffer.WriteString(xml.Tags[len(xml.Tags)-1])
		xml.Buffer.WriteString(">")
	}

	xml.Tags = xml.Tags[:len(xml.Tags)-1]
}

func (xml *XmlWriter) Attr(name, value string) {
	xml.Buffer.WriteString(" ")
	xml.Buffer.WriteString(name)
	xml.Buffer.WriteString("=")
	xml.quoteString(value)
}

func (xml *XmlWriter) Body(text string) {
	xml.CloseBeginTag()
	xml.NewLine(0)
	xml.Buffer.WriteString(text)
}

func (xml *XmlWriter) NewLine(offset int) {
	if xml.NoNewLine {
		return
	}

	xml.Buffer.WriteString("\n")
	n := len(xml.Tags) + offset
	for i := 0; i < n; i++ {
		xml.Buffer.WriteString("  ")
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
	xml.Buffer.WriteString(">")
	xml.DoBeginTag = false
}

func (xml *XmlWriter) quoteString(v string) {
	xml.Buffer.WriteString("\"")

	for _, ch := range v {
		switch ch {
		case '"':
			xml.Buffer.WriteString("&quot;")
		case '\'':
			xml.Buffer.WriteString("&apos;")
		case '<':
			xml.Buffer.WriteString("&lt;")
		case '>':
			xml.Buffer.WriteString("&gt;")
		case '&':
			xml.Buffer.WriteString("&amp;")
		default:
			xml.Buffer.WriteRune(ch)
		}
	}

	xml.Buffer.WriteString("\"")
}
