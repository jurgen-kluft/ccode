package dev

// LibraryType defines the type of lib, there are 2 types of libraries, system and user
type LibraryType int

const (
	LibraryTypeUnknown   LibraryType = 0 // Unknown type, used for error handling
	LibraryTypeSystem    LibraryType = 1 //
	LibraryTypeUser      LibraryType = 2 //
	LibraryTypeFramework LibraryType = 4 //
)
