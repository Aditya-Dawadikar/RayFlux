package hash

import (
	"crypto/sha1"
	"math/big"
)

// GetConsistentIndex returns a stable index based on userID + topic.
func GetConsistentIndex(userID, topic string, targets []string) int {
	if len(targets) == 0 {
		return -1
	}

	key := userID + "|" + topic

	hasher := sha1.New()
	hasher.Write([]byte(key))
	hashBytes := hasher.Sum(nil)

	hashInt := big.NewInt(0).SetBytes(hashBytes)
	index := int(hashInt.Mod(hashInt, big.NewInt(int64(len(targets)))).Int64())

	return index
}
