package main

import (
	_ "embed"
	"fmt"
	"os"
	"strconv"

	axe "github.com/jurgen-kluft/ccode/axe"
)

func main() {

	if len(os.Args) < 5 {
		fmt.Println("Usage: espmake-tools-object-info <elfSizePath> <formType> <sortIndex> <objFiles>")
		return
	}

	// Get the arguments from the command line
	elfSizePath := os.Args[1]
	formType, err := strconv.Atoi(os.Args[2])
	sortIndex, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println(err)
		return
	}
	objFiles := os.Args[4:]

	err = axe.ObjectInfo(elfSizePath, formType, sortIndex, objFiles)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Memory usage calculated successfully.")
	}
}
