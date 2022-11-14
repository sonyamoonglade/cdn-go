package hash

import (
	"crypto/sha1"
	"encoding/hex"
)

func SHA1Name(v string) string {
	sha := sha1.Sum([]byte(v))
	return hex.EncodeToString(sha[:])
}
