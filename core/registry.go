package corepkg

type RegistryKey int

const (
	RegistryKeyCurrentUser RegistryKey = iota
	RegistryKeyLocalMachine
)

func QueryRegistryForStringValue(rootKey RegistryKey, keyPath string, item string) (string, error) {
	return queryRegistryForStringValue(rootKey, keyPath, item)
}
