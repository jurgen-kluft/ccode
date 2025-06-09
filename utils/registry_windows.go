package ccode_utils

import (
	"golang.org/x/sys/windows/registry"
)

func queryRegistryForStringValue(rootKey RegistryKey, keyPath string, item string) (string, bool) {
	var windowsRootKey registry.Key
	if rootKey == RegistryKeyCurrentUser {
		windowsRootKey = registry.CURRENT_USER
	} else if rootKey == RegistryKeyLocalMachine {
		windowsRootKey = registry.LOCAL_MACHINE
	}

	// Attempt to open the registry key with read access
	k, err := registry.OpenKey(windowsRootKey, keyPath+"\\"+item, registry.READ)
	if err != nil {
		return "", false
	}
	defer k.Close()

	// Attempt to read the value from the registry key
	value, _, err := k.GetStringValue("")
	if err != nil {
		// err == registry.ErrNotExist {
		return "", false // Other error occurred
	}
	return value, true // Successfully retrieved the value
}
