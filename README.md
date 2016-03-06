# XCODE - Project Generator

This is a project generator that uses Go and its package management (glide) to work with C++ or C# packages.
The structure of packages are defined in Go and files can be generated for Visual Studio, .sln, .vcxproj and .filters.
Any C++ external dependency like Boost, DirectX or whatnot should be wrapped in a package (github or other git server).
