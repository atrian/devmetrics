package crypto

type Hasher interface {
	Hash(metric string, key string) string
	Compare(hash string, metric string, key string) bool
}
