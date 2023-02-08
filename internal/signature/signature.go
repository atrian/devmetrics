// Package signature модуль для подписи и проверки подлинности подписи в передаваемых метриках
package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type Sha256Hasher struct{}

func NewSha256Hasher() *Sha256Hasher {
	return &Sha256Hasher{}
}

// Hash формирует подпись для метрики
func (s Sha256Hasher) Hash(metric string, key string) string {
	if key == "" {
		return ""
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(metric))

	return hex.EncodeToString(h.Sum(nil))
}

// Compare проверяет равен ли переданный хеш расчитанному
func (s Sha256Hasher) Compare(hash string, metric string, key string) bool {
	return hash == s.Hash(metric, key)
}
