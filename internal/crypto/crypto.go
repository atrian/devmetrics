package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type Sha256Hasher struct{}

func (s Sha256Hasher) Hash(metric string, key string) string {
	if "" == key {
		return ""
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(metric))

	return hex.EncodeToString(h.Sum(nil))
}

func NewSha256Hasher() *Sha256Hasher {
	return &Sha256Hasher{}
}
