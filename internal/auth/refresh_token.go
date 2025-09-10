package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32) // 32 bytes = 256 bits
	rand.Read(key)

	encodedString := hex.EncodeToString(key)

	return encodedString, nil

}