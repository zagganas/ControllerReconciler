package util

import (
	"crypto/md5"
	"encoding/base32"
	"encoding/hex"
	"hash/fnv"
	"strings"
)

func CalculateHash(toHash []byte) string {
	hash := fnv.New64a()
	hash.Write(toHash)
	hashedBytes := hash.Sum(nil)
	encoded := base32.StdEncoding.EncodeToString(hashedBytes)
	encoded = strings.Trim(encoded, "=")

	return strings.ToLower(encoded)
}

func CalculateMd5(toHash []byte) string {
	sumBytes := md5.Sum(toHash)
	encoded := hex.EncodeToString(sumBytes[:])

	return strings.ToLower(encoded)
}
