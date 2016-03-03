package denv

import (
	"os"
	"strings"
)

type ProjectWriter interface {
	WriteLn(string) error
	WriteLns([]string) error
}

type ProjectTextWriter struct {
	fhnd *os.File
}

func (writer *ProjectTextWriter) Open(filepath string) (err error) {
	writer.fhnd, err = os.OpenFile(filepath, os.O_CREATE|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	return nil
}
func (writer *ProjectTextWriter) Close() (err error) {
	err = writer.fhnd.Close()
	return err
}

const (
	cTabChar     = "\t"
	cNewLineChar = "\n"
)

func (writer *ProjectTextWriter) WriteLn(line string) (err error) {
	offset := 0
	for offset < len(line) && line[offset] == '+' {
		_, err = writer.fhnd.WriteString(cTabChar)
		if err != nil {
			return err
		}
		offset++
	}
	if offset < len(line) {
		_, err = writer.fhnd.WriteString(line[offset:])
		if err != nil {
			return err
		}
		_, err = writer.fhnd.WriteString(cNewLineChar)
	}
	return err
}

func (writer *ProjectTextWriter) WriteLns(lines []string) (err error) {
	for _, line := range lines {
		line = strings.Trim(line, " ")
		// Skip empty lines
		if len(line) > 0 {
			writer.WriteLn(line)
		}
	}
	return err
}
