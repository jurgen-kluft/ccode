# CCODE - Package Manager + Project Generator

This is a project generator that uses Go and its package management for C++ packages. 
The structure of packages are defined in Go and files can be generated for Visual Studio (.sln, .vcxproj and .filters), Make, CMake, Xcode as well as Tundra. Work in progress for also supporting Zig as a build system.

Any C++ external dependency like Boost, DirectX or whatnot should be wrapped in a package (github or other git server).
There are a couple of notable features that can be triggered when generating the project files:

* generating `.clang-format` and/or `.gitignore`
* generating `source/test/cpp/test_main.cpp`
* converting any file in `embedded` (following the directory structure) to C style array's so as to embed those files into your library/app
* code generation of C++ enum from `.json` file

Also this project is a personal project and thus is not perfect but it serves my needs, feel free to post issues/requests if you want to see additional features.

This allows me to write packages (C++ libraries) and use them in another C++ package by defining a dependency on them. Using the go package management solution you can 'get' these packages and then by running `go run %name%.go` you can generate solution and project files. The goal is to support the following IDE's and/or build-systems:

* [Visual Studio](https://visualstudio.microsoft.com) (supported, Windows)
* [Xcode](https://developer.apple.com/xcode/) (supported, Mac)
* [Tundra](https://github.com/deplinenoise/tundra) (supported, Mac, Linux and Windows)
* [CMake](https://cmake.org/) (supported, Mac, Linux and Windows)
* [Make](https://www.gnu.org/software/make/manual/make.html) (supported, Mac and Linux, untested on Windows)
* [Zig](https://ziglang.org/learn/build-system/) (WIP)

Currently the design is quite set and the goal is to keep creating and maintaining packages to a minimum.

If you have repository/package that uses ccode, you can do the following to generate the CMake build files, this example uses the `cbase` repository:

1. `go run cbase.go --dev=cmake`
2. cd into `target/cmake`
3. `cmake -DCMAKE_BUILD_TYPE=DEBUG`
4. `make`

For Visual Studio build files (on Windows, Visual Studio is the default generator):

1. `go run cbase.go --dev=vs2022`
2. cd into `target/msdev`
3. You now should have Visual Studio solution and project files

For Tundra build files (on Mac and Linux, Tundra is the default generator):

1. `go run cbase.go --dev=tundra`
2. cd into `target/tundra`
3. run `tundra2 debug` or `tundra2 release`

These are the steps to make a new package:

1. Create a new Github repository like `mylibrary`
2. In the root create a `mylibrary.go` file
3. In the root create a folder called `package` with a file in it called `package.go`
4. Once you have specified everything in package.go:
   * In the root 'go get' (this will get all your specified dependencies in GO_PATH)
   * To generate the VS solution (default on Windows) and projects just run: `go run mylibrary.go`  
   * To generate the Tundra build file (default on MacOS) run: `go run mylibrary.go`

Example:
The content of the `mylibrary.go` file:

```go
package main

import (
    "github.com/jurgen-kluft/mylibrary/package"
    "github.com/jurgen-kluft/ccode"
)

func main() {
    ccode.Init()

    // This will generate
    // - ./.gitignore
    // - ./.clang-format
    // - ./source/test/cpp/test_main.cpp    
    ccode.GenerateFiles()
    
    // This will generate the Visual Studio solution and projects, CMakeLists.txt, 
    // Tundra build files or Zig build files
    ccode.Generate(mylibrary.GetPackage())

    // You can also insert generated C++ enums with ToString functions, the my_enums.h
    // file should already exist and have 2 delimiter lines that you can configure as 
    // 'between'. 
    ccode.GenerateCppEnums("embedded/my_enums.h.json", "main/include/cbase/my_enums.h")

    // Or if you are up to it, even generating structs is possible
    ccode.GenerateCppStructs("embedded/my_structs.h.json", "main/include/cbase/my_structs.h")
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

1. `source\main\cpp`: the cpp files of your library. Header files should be included as ```#include "mylibrary/header.h"```
2. `source\main\include\mylibrary`: the header files of your library
3. `source\test\cpp`: the cpp files of your unittest app
4. `source\test\include`: the header files of your unittest app
5. `embedded\**`: all the files that need to be auto embedded or are used for code generation 
   - binary file to .cpp `C array`
   - C++ enum code generation (from `.json` file)
