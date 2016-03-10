package uid

import (
	"crypto/sha256"
	"fmt"
)

func GetGUID(str string) string {
	b := []byte(str)
	hashfunc := sha256.New()
	hashfunc.Write(b)
	hash := hashfunc.Sum(nil)
	return fmt.Sprintf("%08X-%04X-%04X-%04X-%12X", hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:16])
}
