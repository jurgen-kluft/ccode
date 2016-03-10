# XCODE - Project Generator

This is a project generator that uses Go and its package management (glide) to work with C++ or C# packages.
The structure of packages are defined in Go and files can be generated for Visual Studio, .sln, .vcxproj and .filters.
Any C++ external dependency like Boost, DirectX or whatnot should be wrapped in a package (github or other git server).

This allows you to write packages (C++ libraries) and use them in another package by defining a dependency on them. Using the package management solution (Glide) you can 'get' these packages and then by running 'go run $name.go' you can generate projects files (currently only Visual Studio 2015). The goal is to support other IDE's and/or build-systems.

Currently the design is still in flux but the goal is to keep creating and maintaining packages to a minimum. These are the steps to make a new package:

1. Create a new Github repository like 'mylibrary'
2. In the root create a mylibrary.go file
3. In the root create a folder called 'package' with a file in it called 'package.go'

Example:
The content of the ```mylibrary.go``` file:
```
package main

import (
	"github.com/jurgen-kluft/mylibrary/package"
	"github.com/jurgen-kluft/xcode"
)

func main() {
	xcode.Generate(mylibrary.GetPackage())
}
```

The content of the ```/package/package.go``` file with one dependency on 'myunittest':
```
package mylibrary

import (
	"github.com/jurgen-kluft/xcode/denv"
	"github.com/githubusername/myunittest/package"
)

// GetPackage returns the package object of 'mylibrary'
func GetPackage() *denv.Package {
	// Dependencies
	unittestpkg := myunittest.GetPackage()

	// The main (mylibrary) package
	mainpkg := denv.NewPackage("mylibrary")
	mainpkg.AddPackage(unittestpkg)

	// 'mylibrary' library
	mainlib := denv.SetupDefaultCppLibProject("mylibrary", "github.com\\githubusername\\mylibrary")
	mainlib.Dependencies = append(mainlib.Dependencies, unittestpkg.GetMainLib())

	// 'mylibrary' unittest project
	maintest := denv.SetupDefaultCppTestProject("mylibrary_test", "github.com\\githubusername\\mylibrary")

	mainpkg.AddMainLib(mainlib)
	mainpkg.AddUnittest(maintest)
	return mainpkg
}
```

There are some requirements for the layout of folders inside of your repository to hold the library and unittest files, this is the layout:

1. source\main\cpp: all the cpp files of your library. Source files should include header files like ```#include "mylibrary/header.h"```
2. source\main\include\mylibrary: all the header files of your library
3. source\test\cpp: all the cpp files of your unittest app
4. source\test\include: all the header files of your unittest app
 
