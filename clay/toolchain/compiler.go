package toolchain

// Compiler is an interface that defines the methods required for compiling source files
type Compiler interface {
	SetupArgs(defines []string, includes []string)
	Compile(sourceAbsFilepath []string, objRelFilepath []string) error
}

// Note: Compiler dependency management.
// A compiler compiles a source file + header files into an object file.
// The compiler before compiling the source file, should query the depTrackr
// to check if the source file + header files are up to date.
// After compiling, the compiler should add the object file, source file + header
// files as an item with dependencies to the depTrackr.

// // Before we compile, verify the state of the object file.
// if cl.toolChain.depTrackr.QueryItem(objFilepath) {
// 	// object file is up-to-date
// 	cl.toolChain.depTrackr.CopyItem(objFilepath)
// 	return objFilepath, nil
// }
// ....
// ... compile the source file
// ....
// // Source file has been compiled, add the .d file to the dependency tracker
// dotdFilepath := sourceRelFilepath + ".d"
// cl.toolChain.depTrackr.AddItem(dotdFilepath, []deptrackr.StringItem{})
