package corepkg

import (
	"golang.org/x/sys/windows/registry"
)

func queryRegistryForStringValue(rootKey RegistryKey, keyPath string, item string) (string, error) {
	var windowsRootKey registry.Key
	if rootKey == RegistryKeyCurrentUser {
		windowsRootKey = registry.CURRENT_USER
	} else if rootKey == RegistryKeyLocalMachine {
		windowsRootKey = registry.LOCAL_MACHINE
	}

	// Attempt to open the registry key with read access
	k, err := registry.OpenKey(windowsRootKey, keyPath, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	// Attempt to read the value for the item from the registry path
	value, _, err := k.GetStringValue(item)
	if err != nil {
		// err == registry.ErrNotExist {
		return "", err // Other error occurred
	}
	return value, nil // Successfully retrieved the value
}
