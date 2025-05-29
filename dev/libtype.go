package dev

// LibType defines the type of lib, there are 2 types of libraries, system and user
type LibType int

const (
	SystemLibrary LibType = 1 //
	UserLibrary   LibType = 2 //
	Framework     LibType = 4 //
)
