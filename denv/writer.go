package denv

import (
	"fmt"
	"os"
	"strings"
)

type ProjectWriterr interface {
	WriteLn(string) error
	WriteLns([]string) error
}

type ProjectTextWriterr struct {
	fhnd *os.File
}

func (writer *ProjectTextWriterr) Open(filepath string) (err error) {
	writer.fhnd, err = os.OpenFile(filepath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Error opening file: '%s' with error ''%s'\n", filepath, err.Error())
		return err
	}
	return nil
}
func (writer *ProjectTextWriterr) Close() (err error) {
	err = writer.fhnd.Close()
	return err
}

const (
	cTabChar     = "\t"
	cNewLineChar = "\n"
)

func (writer *ProjectTextWriterr) WriteLn(line string) (err error) {
	if len(line) == 0 {
		_, err = writer.fhnd.WriteString(cNewLineChar)
	} else {
		offset := 0
		for offset < len(line) && line[offset] == '+' {
			_, err = writer.fhnd.WriteString(cTabChar)
			if err != nil {
				fmt.Printf("Error writing to file with error '%s'", err.Error())
				return err
			}
			offset++
		}
		if offset < len(line) {
			_, err = writer.fhnd.WriteString(line[offset:])
			if err != nil {
				fmt.Printf("Error writing to file with error '%s'", err.Error())
				return err
			}
			_, err = writer.fhnd.WriteString(cNewLineChar)
		}
	}
	return err
}

func (writer *ProjectTextWriterr) WriteLns(lines []string) (err error) {
	for _, line := range lines {
		line = strings.Trim(line, " ")
		err = writer.WriteLn(line)
	}
	return err
}
