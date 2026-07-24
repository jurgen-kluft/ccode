package embedded

import (
	_ "embed"
	"io"
	"log"
	"os"
)

//go:embed test_main_cbase.txt
var testMainCBase string

//go:embed test_main_ccore.txt
var testMainCCore string

//go:embed test_main_basic.txt
var testMain string

var testMainFilename = "source/test/cpp/test_main.cpp"

func WriteTestMainCpp(ccore bool, cbase bool, overwrite bool) {
	// check if the file exists, if it does not, create it
	if _, err := os.Stat(testMainFilename); os.IsNotExist(err) || overwrite {
		// even if the file exists, we want to overwrite it
		f, err := os.Create(testMainFilename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if cbase {
			_, err = io.WriteString(f, testMainCBase)
		} else if ccore {
			_, err = io.WriteString(f, testMainCCore)
		} else {
			_, err = io.WriteString(f, testMain)
		}
		if err != nil {
			log.Fatal(err)
		}
	}
}
