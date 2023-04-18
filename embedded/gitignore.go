package embedded

import (
	_ "embed"
	"io"
	"log"
	"os"
)

//go:embed .gitignore
var gitIgnore string
var gitIgnoreFilename = ".gitignore"

func WriteGitIgnore(overwrite bool) {
	// check if the file exists, if it does not, create it
	_, err := os.Stat(gitIgnoreFilename)
	if err == nil && !overwrite {
		return
	}

	// even if the file exists, we want to overwrite it
	f, err := os.Create(gitIgnoreFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = io.WriteString(f, gitIgnore)
	if err != nil {
		log.Fatal(err)
	}
}
