package embedded

import (
	_ "embed"
	"io"
	"log"
	"os"
)

//go:embed .clang-format
var clangFormat string
var clangFormatFilename = ".clang-format"

func WriteClangFormat(overwrite bool) {
	// check if the file exists, if it does not, create it
	_, err := os.Stat(clangFormatFilename)
	if err == nil && !overwrite {
		return
	}

	// even if the file exists, we want to overwrite it
	f, err := os.Create(clangFormatFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = io.WriteString(f, clangFormat)
	if err != nil {
		log.Fatal(err)
	}
}
