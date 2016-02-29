package xcode

import (
	"strings"
)

func SplitTargets(targets string) []string {
	return strings.Split(targets, ";")
}

func IsValidTarget(target string) bool {
	if target == "Win32" {
		return true
	} else if target == "Win64" {
		return true
	}
	return false
}

func DefinesPerTarget(target string) string {
	if target == "Win32" {
		return "TARGET_PC;TARGET_32BIT;WIN32"
	} else if target == "Win64" {
		return "TARGET_PC;TARGET_64BIT;WIN32"
	}
	return ""
}
