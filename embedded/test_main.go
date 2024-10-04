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

	if cbase {
		_, err = io.WriteString(f, testMainCBase)
		if err != nil {
			log.Fatal(err)
		}
	} else if ccore {
		_, err = io.WriteString(f, testMainCCore)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_, err = io.WriteString(f, testMain)
		if err != nil {
			log.Fatal(err)
		}
	}
}
