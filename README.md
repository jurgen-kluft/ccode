# CCODE - Package Manager + Project Generator

This is a project generator that uses Go and its package management for C++ packages. 
The structure of packages are defined in Go and files can be generated for Visual Studio (.sln, .vcxproj and .filters) as well as [Tundra](https://github.com/deplinenoise/tundra). Work in progress for also enabling CMake and Zig as build systems.

Any C++ external dependency like Boost, DirectX or whatnot should be wrapped in a package (github or other git server).
There are a couple of notable features that are triggered when generating the project files:

* generating `.clang-format` and `.gitignore`
* generating `source/test/cpp/test_main.cpp`
* converting any file in `embedded` (following the directory structure) to C style array's so as to embed those files into your library/app

Also this project is a personal project and so is not perfect but it serves my needs, feel free to post issues/requests if you want to see additional features.

This allows you to write packages (C++ libraries) and use them in another package by defining a dependency on them. Using the go package management solution you can 'get' these packages and then by running 'go run $name.go' you can generate projects files . The goal is to support these IDE's and/or build-systems:

* [Visual Studio](https://visualstudio.microsoft.com) (supported)
* [Tundra](https://github.com/deplinenoise/tundra) (supported)
* [CMake](https://cmake.org/) (supported)
* [Zig](https://ziglang.org/learn/build-system/) (WIP)

Currently the design is quite set and the goal is to keep creating and maintaining packages to a minimum.

If you have repository/package that uses ccode, you can do the following to generate the CMake build files:

1. `go run cbase.go --DEV=cmake`
2. cd into `target/cmake`
3. `cmake -DCMAKE_BUILD_TYPE=DEBUG`
4. `make`

For Visual Studio build files (on Windows):

1. `go run cbase.go`
2. In the root of your package you now should have a `cbase_test.sln` solution file

For Tundra:

1. `go run cbase.go --DEV=tundra`
2. cd into `target/tundra`
3. run `tundra`

These are the steps to make a new package:

1. Create a new Github repository like ``mylibrary``
2. In the root create a ``mylibrary.go`` file
3. In the root create a folder called ``package`` with a file in it called ``package.go``
4. Once you have specified everything in package.go:
   * In the root 'go get' (this will get all your specified dependencies in GO_PATH)
   * To generate the VS solution (default on Windows) and projects just run: ``go run mylibrary.go``  
   * To generate the Tundra build file (default on MacOS) run: ``go run mylibrary.go``

Example:
The content of the ```mylibrary.go``` file:

```go
package main

import (
    "github.com/jurgen-kluft/mylibrary/package"
    "github.com/jurgen-kluft/ccode"
)

func main() {
    ccode.Init()
    ccode.Generate(mylibrary.GetPackage())
}
```

The content of the ```/package/package.go``` file with one dependency on 'myunittest':

```go
package mylibrary

import (
    "github.com/jurgen-kluft/ccode/denv"
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
    mainlib := denv.SetupDefaultCppLibProject("mylibrary", "github.com/githubusername/mylibrary")
    mainlib.Dependencies = append(mainlib.Dependencies, unittestpkg.GetMainLib())

    // 'mylibrary' unittest project
    maintest := denv.SetupDefaultCppTestProject("mylibrary_test", "github.com/githubusername/mylibrary")

    mainpkg.AddMainLib(mainlib)
    mainpkg.AddUnittest(maintest)
    return mainpkg
}
```

There are some requirements for the layout of folders inside of your repository to hold the library and unittest files, this is the layout:

1. source\main\cpp: all the cpp files of your library. Source files should include header files like  
   ```#include "mylibrary/header.h"```
2. source\main\include\mylibrary: all the header files of your library
3. source\test\cpp: all the cpp files of your unittest app
4. source\test\include: all the header files of your unittest app
5. embedded\**: all the files that need to be auto embedded (file to .cpp 'C array') 
