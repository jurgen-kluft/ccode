package embedded

import (
	_ "embed"
	"io"
	"log"
	"os"
)

//go:embed test_main.txt
var testMain string
var testMainFilename = "source/test/cpp/test_main.cxx"

func WriteTestMainCxx(overwrite bool) {
	// check if the file exists, if it does not, create it
	_, err := os.Stat(testMainFilename)
	if err == nil && !overwrite {
		return
	}

	// even if the file exists, we want to overwrite it
	f, err := os.Create(testMainFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = io.WriteString(f, testMain)
	if err != nil {
		log.Fatal(err)
	}
}
