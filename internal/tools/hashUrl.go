package tools

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashUrl(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))
	return hex.EncodeToString(sum[:])
}
