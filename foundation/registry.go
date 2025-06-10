package foundation

type RegistryKey int

const (
	RegistryKeyCurrentUser RegistryKey = iota
	RegistryKeyLocalMachine
)

func QueryRegistryForStringValue(root string, keyPath string, item string) (string, bool) {
	var rootKey RegistryKey
	if root == "HKCU" {
		rootKey = RegistryKeyCurrentUser
	} else if root == "HKLM" {
		rootKey = RegistryKeyLocalMachine
	} else {
		return "", false // Invalid root key
	}
	return queryRegistryForStringValue(rootKey, keyPath, item)
}
