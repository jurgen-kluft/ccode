package toolchain

type ArchiverType int

const (
	// ArchiverTypeStatic represents a static archive (e.g., .a or .lib).
	ArchiverTypeStatic ArchiverType = iota
	// ArchiverTypeDynamic represents a shared/dynamic archive (e.g., .so or .dll).
	ArchiverTypeDynamic
)

// Archiver (also known as Lib) is an interface that defines the methods required for
// creating an archive (.a, .lib) file.
type Archiver interface {

	// Returns the filepath for the archive
	// e.g. "path/to/library/name.ext", your should provide the filepath without extension:
	//      "path/to/library/name", and it will return "path/to/library/libname.a" or "path/to/library/name.lib"
	LibFilepath(filepath string) string

	// SetupArgs prepares the arguments for the archiver based on user-defined variables.
	// It should be called before using the Archive method.
	SetupArgs()

	// Archive takes a list of input object file paths and an output archive file path.
	// Both paths are relative to the build path.
	Archive(inputObjAbsFilepaths []string, outputArchiveRelFilepath string) error
}

// Note: Archiver dependency management.
// An archive is a collection of object files that together form a library.
// The archiver before archiving the list of object files, should query the depTrackr
// to check if the object files are up to date.
// After archiving, the archiver should add the archive file + the object files as
// an item with dependencies to the depTrackr.

// Example case:
// Let's say we have previously archived 'test.a' with 'test1.o', 'test2.o' and 'test3.o'.
// Now we are asking it to archive 'test1.o', 'test2.o', 'test3.o' but also 'test4.o'.
// Querying the dependency tracker to see if the archive is already up-to-date will tell us it is
// up-to-date, because 'test1.o', 'test2.o' and 'test3.o' are up-to-date, 'test.a' is up-to-date as
// well as 'test.a.d' (the dependency file for the archive).
// Q: So how do we deal with a change in the list of object files belonging to an archive ?
