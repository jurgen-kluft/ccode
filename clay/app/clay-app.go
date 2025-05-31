package main

import (
	clay "github.com/jurgen-kluft/ccode/clay"
)

func main() {
	clay.ClayAppCreateProjectsFunc = CreateProjects
	clay.ClayAppMainArduino()
}
